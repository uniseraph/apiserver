package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"sort"
	"time"
)

type PoolDashboardRequest struct {
	PoolId    string
	StartTime string
}

type PoolDashboardResponse struct {
	Summary *PoolDashboardSummary
	Trend   *PoolDashboardTrend
}

type PoolDashboardSummary struct {
	Nodes        int
	CPUs         int
	CPUsUsed     int
	Memory       int64
	MemoryUsed   int64
	Disk         int64
	DiskUsed     int64
	DataDisk     map[string]interface{}
	Applications int
	Containers   int
}

type PoolDashboardTrend struct {
	Creates                  []*Record
	Upgrades                 []*Record
	Rollbacks                []*Record
	MostUpgradeApplications  []*Application
	MostRollbackApplications []*Application
}

type Application struct {
	Id      string
	Title   string
	Name    string
	Version string
	Count   int
}

type Record struct {
	day   string
	count int
}

func poolDashboard(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &PoolDashboardRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	from, err := time.Parse(
		"2006-01-02 15:04:05",
		req.StartTime+" 00:00:00")
	if err != nil {
		HttpError(w, "StartTime格式错误"+err.Error(), http.StatusBadRequest)
		return
	}

	rsp := &PoolDashboardResponse{
		Summary: &PoolDashboardSummary{},
		Trend:   &PoolDashboardTrend{},
	}

	utils.GetMgoCollections(ctx, w, []string{"pool", "application", "deployment"}, func(cs map[string]*mgo.Collection) {

		poolCol, _ := cs["pool"]
		applicationCol, _ := cs["application"]
		deploymentCol, _ := cs["deployment"]

		pool := &types.PoolInfo{}
		if err := poolCol.FindId(bson.ObjectIdHex(req.PoolId)).One(pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, fmt.Sprintf("no such a pool:%s", req.PoolId), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//logrus.Debugf("poolDashboard get the pool %#v", pool)

		rsp.Summary.Memory = pool.Memory
		rsp.Summary.Containers = pool.Containers
		rsp.Summary.Nodes = pool.NodeCount
		rsp.Summary.CPUs = pool.CPUs

		//TODO 需要swarm提供
		rsp.Summary.MemoryUsed = 0
		rsp.Summary.CPUsUsed = 0
		rsp.Summary.Disk = 0
		rsp.Summary.DiskUsed = 0
		rsp.Summary.DataDisk = make(map[string]interface{})

		apps, err := applicationCol.Find(bson.M{"poolid": req.PoolId}).Count()
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rsp.Summary.Applications = apps

		deployments := make([]types.Deployment, 0, 200)
		if err := deploymentCol.Find(bson.M{"poolid": req.PoolId, "createdtime": bson.M{"$gte": from}}).Sort("-createdtime").All(&deployments); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		creates := make(map[string]int)
		rollbacks := make(map[string]int)
		upgrades := make(map[string]int)

		for i, _ := range deployments {
			year, month, day := time.Unix(deployments[i].CreatedTime, 0).Date()

			daystr := buildDayStr(year, month, day)

			var target map[string]int
			if deployments[i].OperationType == types.DEPLOYMENT_OPERATION_CREATE {
				target = creates
			} else if deployments[i].OperationType == types.DEPLOYMENT_OPERATION_UPGRADE {
				target = upgrades
			} else {
				target = rollbacks
			}

			count, ok := target[daystr]

			if !ok {
				target[daystr] = 1
			} else {
				target[daystr] = count + 1
			}
		}

		rsp.Trend.Creates = sortResult(creates)
		rsp.Trend.Upgrades = sortResult(upgrades)
		rsp.Trend.Rollbacks = sortResult(rollbacks)

		if as, err := getMostApplication(deploymentCol, req.PoolId, types.DEPLOYMENT_OPERATION_UPGRADE); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {

			if bs, err := appendDetail(as, applicationCol); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			} else {
				rsp.Trend.MostUpgradeApplications = bs
			}
		}

		if as, err := getMostApplication(deploymentCol, req.PoolId, types.DEPLOYMENT_OPERATION_ROLLBACK); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			if bs, err := appendDetail(as, applicationCol); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			} else {
				rsp.Trend.MostRollbackApplications = bs
			}
		}

		HttpOK(w, rsp)
	})

}
func appendDetail(applications []*Application, applicationCol *mgo.Collection) ([]*Application, error) {

	result := make([]*Application, 0, len(applications))

	for i, _ := range applications {

		application := &types.Application{}
		if err := applicationCol.FindId(bson.ObjectIdHex(applications[i].Id)).One(application); err != nil {
			if err == mgo.ErrNotFound {
				//ignore the application
				continue
			}

			return nil, err
		}

		applications[i].Name = application.Name
		applications[i].Title = application.Title
		applications[i].Version = application.Version

		result = append(result, &Application{
			Id:      applications[i].Id,
			Count:   applications[i].Count,
			Name:    application.Name,
			Version: application.Version,
			Title:   application.Title,
		})

	}

	return result, nil

}

// db.deployment.aggregate([
// 	{  $group : { _id : { operationtype: "upgrade" , applicationid:"$applicationid"    } ,"count":{ $sum : 1  }    }  } ,
//  { $sort:{count:-1} } ,
//  { $limit:10 } ])

//db.deployment.aggregate([
// { $match:{ operationtype:"upgrade"  } },
//  {  $group : { _id : {  applicationid:"$applicationid" } ,count:{ $sum : 1  }    }  }  ,
//  {  $project : { _id:0, count:1,  id:"$_id.applicationid"   }   }     ]    )
func getMostApplication(deploymentCol *mgo.Collection, poolid, operationtype string) ([]*Application, error) {

	matchOp := bson.M{
		"$match": bson.M{
			"operationtype": operationtype,
			"poolid":        poolid,
		},
	}

	groupOp := bson.M{
		"$group": bson.M{
			//	"_id": bson.M{
			//		"applicationid": "$applicationid",
			//		"title":         "$title",
			//		"version":       "$version",
			//		"name":          "$name",
			//	},
			"_id": bson.M{
				"applicationid": "$applicationid",
			},
			"count": bson.M{"$sum": 1},
		},
	}

	projectOp := bson.M{
		"$project": bson.M{
			//	"_id": bson.NewObjectId(),
			"id": "$_id.applicationid",
			//	"version" :  "$_id.version",
			//	"title" :  "$_id.title",
			//	"name" :  "$_id.name",
			"count": 1,
		},
	}

	sortOp := bson.M{"$sort": bson.M{"count": -1}}
	limitOp := bson.M{"$limit": 10}

	ops := []bson.M{matchOp, groupOp, projectOp, sortOp, limitOp}

	result := make([]*Application, 0, 10)

	if err := deploymentCol.Pipe(ops).All(&result); err != nil {
		return nil, err
	}

	for i, _ := range result {
		logrus.Debugf("operation is %s,i is %d, result is %#v", operationtype, i, result[i])
	}

	return result, nil

}

func buildDayStr(year int, month time.Month, day int) string {

	daystr := fmt.Sprintf("%s-", year)

	if month <= 9 {
		daystr += fmt.Sprintf("0%d-", month)
	} else {
		daystr += fmt.Sprintf("%d-", month)
	}

	if day <= 9 {
		daystr += fmt.Sprintf("0%d", day)
	} else {
		daystr += fmt.Sprintf("%d", day)

	}

	return daystr

}

func sortResult(day2count map[string]int) []*Record {
	var keys []string
	for key, _ := range day2count {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	result := make([]*Record, 0, len(keys))

	for _, key := range keys {

		result = append(result, &Record{
			day:   key,
			count: day2count[key],
		})
	}

	return result
}
