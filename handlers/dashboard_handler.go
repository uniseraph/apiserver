package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"sort"
	"time"
)

type PoolDashboardRequest struct {
	id        string
	StartTime string
}

type PoolDashboardResponse struct {
	summary *PoolDashboardSummary
	trend   *PoolDashboardTrend
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
	Id, Title, Name, Version string
	Count                    int
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
		summary: &PoolDashboardSummary{},
		trend:   &PoolDashboardTrend{},
	}

	utils.GetMgoCollections(ctx, w, []string{"pool", "application", "deployment"}, func(cs map[string]*mgo.Collection) {

		poolCol, _ := cs["pool"]
		applicationCol, _ := cs["application"]
		deploymentCol, _ := cs["deployment"]

		pool := &types.PoolInfo{}
		if err := poolCol.FindId(bson.ObjectId(req.id)).One(pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, fmt.Sprintf("no such a pool:%s", req.id), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rsp.summary.Memory = pool.Memory
		rsp.summary.Containers = pool.Containers
		rsp.summary.Nodes = pool.NodeCount
		rsp.summary.CPUs = pool.CPUs

		//TODO 需要swarm提供
		rsp.summary.MemoryUsed = 0
		rsp.summary.CPUsUsed = 0
		rsp.summary.Disk = 0
		rsp.summary.DiskUsed = 0
		rsp.summary.DataDisk = make(map[string]interface{})

		apps, err := applicationCol.Find(bson.M{"poolid": req.id}).Count()
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rsp.summary.Applications = apps

		deployments := make([]types.Deployment, 0, 200)
		if err := applicationCol.Find(bson.M{"poolid": req.id, "createtime": bson.M{"$gte": from}}).Sort("createtime").All(deployments); err != nil {
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

		rsp.trend.Creates = sortResult(creates)
		rsp.trend.Upgrades = sortResult(upgrades)
		rsp.trend.Rollbacks = sortResult(rollbacks)

		if as, err := getMostApplication(deploymentCol, req.id, types.DEPLOYMENT_OPERATION_UPGRADE); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			rsp.trend.MostUpgradeApplications = as
		}

		if as, err := getMostApplication(deploymentCol, req.id, types.DEPLOYMENT_OPERATION_ROLLBACK); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			rsp.trend.MostUpgradeApplications = as
		}

	})

	HttpOK(w, rsp)
}

// db.deployment.aggregate([
// 	{  $group : { _id : { operationtype: "upgrade" , applicationid:"$applicationid"    } ,"count":{ $sum : 1  }    }  } ,
//  { $sort:{count:-1} } ,
//  { $limit:10 } ])

func getMostApplication(deploymentCol *mgo.Collection, poolid, operationtype string) ([]*Application, error) {
	groupOp := bson.M{
		"$group": bson.M{
			"_id":     bson.M{"operationtype": operationtype, "poolid": poolid, "applicationid": "$applicationid"},
			"count":   bson.M{"$sum": 1},
			"title":   bson.M{"title": "$title"},
			"version": bson.M{"version": "$version"},
			"name":    bson.M{"name": "$name"},
		},
	}
	sortOp := bson.M{"$sort": bson.M{"count": -1}}
	limitOp := bson.M{"$limit": 10}

	ops := []bson.M{groupOp, sortOp, limitOp}

	result := make([]*Application, 0, 10)

	if err := deploymentCol.Pipe(ops).All(result); err != nil {
		return nil, err
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
