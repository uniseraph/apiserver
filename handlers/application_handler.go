package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/application"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

type ApplicationCreateRequest struct {
	TemplateId                 string `json:ApplicationTemplateId",omitempty"`
	PoolId, Title, Description string
}

type ApplicationCreateResponse struct {
	types.Application
}

func createApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &ApplicationCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	//肯定有的，不用处理
	mgoSession, _ := utils.GetMgoSessionClone(ctx)
	config := utils.GetAPIServerConfig(ctx)
	currentuser, _ := utils.GetCurrentUser(ctx)

	colPool := mgoSession.DB(config.MgoDB).C("pool")
	pool := &types.PoolInfo{}
	if err := colPool.FindId(bson.ObjectIdHex(req.PoolId)).One(pool); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	colTemplate := mgoSession.DB(config.MgoDB).C("template")
	template := &types.Template{}
	if err := colTemplate.FindId(bson.ObjectIdHex(req.TemplateId)).One(template); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	colApplication := mgoSession.DB(config.MgoDB).C("application")

	n, err := colApplication.Find(bson.M{"poolid": req.PoolId, "name": template.Name}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusNotFound)
		return
	}

	if n >= 1 {
		HttpError(w, "该集群中存在同名应用", http.StatusInternalServerError)
		return
	}

	app := &types.Application{}
	app.Id = bson.NewObjectId()
	app.PoolId = req.PoolId
	app.TemplateId = req.TemplateId
	app.Title = req.Title
	app.Description = req.Description
	app.PoolId = req.PoolId
	app.Name = template.Name
	app.Version = template.Version

	app.CreatorId = currentuser.Id.Hex()
	app.UpdaterId = currentuser.Id.Hex()
	app.UpdaterName = currentuser.Name
	app.CreatedTime = time.Now().Unix()
	app.UpdatedTime = time.Now().Unix()
	app.Status = "running"

	app.Services = mergeServices(template.Services, pool)

	if err := colApplication.Insert(app); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//if err := application.CreateApplication(app, pool); err != nil {
	//	HttpError(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	//TODO 马上start ？
	//if err := application.StartApplication(app, pool); err != nil {
	//	//TODO 需要删除所有已创建成功的容器？？？
	//	HttpError(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	m := make(map[string]int)
	for _, service := range app.Services {
		m[service.Name] = service.ReplicaCount
	}

	if err := application.ScaleApplication(app, pool, m); err != nil {
		//	//TODO 需要删除所有已创建成功的容器？？？
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	HttpOK(w, app)
}

//用参数目录填充service定义中的环境变量
func mergeServices(services []types.Service, info *types.PoolInfo) []types.Service {

	//TODO
	return services

}

//PoolId -- 集群ID
//Keyword -- Title或Name前缀搜索，可以为空
//PageSize -- 每页显示多少条
//Page -- 当前页

type ApplicationListRequest struct {
	PageRequest
	PoolId string
	Name   string
}

type ApplicationListResponse struct {
	PageResponse
	Data []*types.Application
}

func getApplicationList(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &ApplicationListRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.PoolId == "" {
		HttpError(w, "PoolId 不能为空", http.StatusBadRequest)
		return
	}

	if req.Page == 0 {
		HttpError(w, "从第一页开始", http.StatusBadRequest)
		return
	}

	if req.PageSize == 0 {
		req.PageSize = 20
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("application")

	result := ApplicationListResponse{
		Data: make([]*types.Application, 100),
	}

	pattern := fmt.Sprintf("^%s", req.Keyword)

	selector := bson.M{"poolid": req.PoolId}

	if req.Name != "" {
		selector["name"] = req.Name
	}

	if req.Keyword != "" {
		regex1 := bson.M{"name": bson.M{"$regex": bson.RegEx{Pattern: pattern, Options: "i"}}}
		regex2 := bson.M{"title": bson.M{"$regex": bson.RegEx{Pattern: pattern, Options: "i"}}}
		selector = bson.M{"$and": []bson.M{bson.M{"$or": []bson.M{regex1, regex2}}, selector}}
	}

	logrus.Debugf("getApplication::过滤条件为%#v", selector)

	if result.Total, err = c.Find(selector).Count(); err != nil {
		HttpError(w, fmt.Sprintf("查询记录数出错，%s", err.Error()), http.StatusInternalServerError)
		return
	}

	logrus.Debugf("getApplication::符合条件的application有%d个", result.Total)

	if err := c.Find(selector).Sort("title").Limit(req.PageSize).Skip(req.PageSize * (req.Page - 1)).All(&result.Data); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result.Keyword = req.Keyword
	result.Page = req.Page
	result.PageSize = req.PageSize
	result.PageCount = result.Total / result.PageSize

	HttpOK(w, &result)

}

func getContainerSSHInfo(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func scaleApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func upgradeApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func stopApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

//TODO 这个接口非常危险
func restartApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func startApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func getApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	utils.GetMgoCollections(ctx, w, []string{"application"}, func(cs map[string]*mgo.Collection) {

		resp := &types.Application{}

		if err := cs["application"].FindId(bson.ObjectIdHex(id)).One(resp); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的应用", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		HttpOK(w, resp)
	})

}

func rollbackApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
