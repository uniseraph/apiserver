package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/application"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"regexp"
	"time"
)

type ApplicationCreateRequest struct {
	ApplicationTemplateId      string
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
	logrus.Debugf("createApplication::recving a request %#v", req)

	//肯定有的，不用处理
	mgoSession, _ := utils.GetMgoSessionClone(ctx)
	defer mgoSession.Close()
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
	if err := colTemplate.FindId(bson.ObjectIdHex(req.ApplicationTemplateId)).One(template); err != nil {
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
	app.TemplateId = req.ApplicationTemplateId
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

	app.Services, err = mergeServices(ctx, template.Services, pool)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := colApplication.Insert(app); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m := make(map[string]int)
	for _, service := range app.Services {
		m[service.Name] = service.ReplicaCount
	}

	if err := application.ScaleApplication(ctx, app, pool, m); err != nil {
		//	//TODO 需要删除所有已创建成功的容器？？？

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	HttpOK(w, app)
}

func replaceEnv(ctx context.Context, l *types.Label, pool *types.PoolInfo) error {

	re := regexp.MustCompile(`\$\{(.+)\}`)

	loc := re.FindStringIndex(l.Value)

	if loc == nil {
		return nil
	}

	key := l.Value[loc[0]+2 : loc[1]-1]

	value, err := GetEnvValueByName(ctx, pool.EnvTreeId, pool.Id.Hex(), key)

	if err != nil {
		return err
	}

	l.Value = re.ReplaceAllString(l.Value, value.Value)

	return nil
}

//用参数目录填充service定义中的环境变量
func mergeServices(ctx context.Context, services []types.Service, pool *types.PoolInfo) ([]types.Service, error) {

	//TODO

	for _, service := range services {

		logrus.Debugf("before merge:: service is %#v", service)

		for i, _ := range service.Labels {

			if err := replaceEnv(ctx, &service.Labels[i], pool); err != nil {
				return nil, errors.New("替换Label的环境变量失败." + err.Error())
			}

		}

		for i, _ := range service.Envs {

			if err := replaceEnv(ctx, &service.Envs[i].Label, pool); err != nil {

				return nil, errors.New("替换Label的环境变量失败." + err.Error())
			}
		}
		logrus.Debugf("after merge:: service is %#v", service)

	}

	return services, nil

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
		selector["$or"] = []bson.M{regex1, regex2}
	}

	logrus.Debugf("getApplication::过滤条件为%#v", selector)

	utils.GetMgoCollections(ctx, w, []string{"application", "team"}, func(cs map[string]*mgo.Collection) {
		/*
			开始权限校验
		*/
		appIds := make([]bson.ObjectId, 0, 20)

		user, err := utils.GetCurrentUser(ctx)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*
			验证用户是否有权访问集群
		*/

		//检查当前用户是否有权限操作该容器
		if user.RoleSet&types.ROLESET_SYSADMIN == types.ROLESET_SYSADMIN {
			//如果用户是系统管理员
			//则不需要校验用户对该机器的权限
			goto AUTHORIZED
		}

		//已经给当前用户授权过的集群，可以查看
		appIds = append(appIds, user.ApplicationIds...)

		//如果该用户加入过某些团队
		//则该团队能查看的pool
		//该用户也可以查看
		//则验证通过
		if len(user.TeamIds) > 0 {
			teams := make([]types.Team, 0, 10)
			selector := bson.M{
				"_id": bson.M{
					"$in": user.TeamIds,
				},
			}
			//查找该用户所在Team
			if err := cs["team"].Find(selector).All(&teams); err != nil {
				if err == mgo.ErrNotFound {
					HttpError(w, "not found params", http.StatusNotFound)
					return
				}
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			//如果用户所在的某个TEAM
			//拥有对该集群的授权
			//则验证通过
			for _, team := range teams {
				appIds = append(appIds, team.ApplicationIds...)
			}
		}

		selector["_id"] = bson.M{
			"$in": appIds,
		}

	AUTHORIZED:

		if result.Total, err = cs["application"].Find(selector).Count(); err != nil {
			HttpError(w, fmt.Sprintf("查询记录数出错，%s", err.Error()), http.StatusInternalServerError)
			return
		}

		logrus.Debugf("getApplication::符合条件的application有%d个", result.Total)

		if err := cs["application"].Find(selector).Sort("title").Limit(req.PageSize).Skip(req.PageSize * (req.Page - 1)).All(&result.Data); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result.Keyword = req.Keyword
		result.Page = req.Page
		result.PageSize = req.PageSize
		result.PageCount = result.Total / result.PageSize

		HttpOK(w, &result)
	})
}

func getContainerSSHInfo(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

type ApplicationScaleRequest struct {
	ServiceName  string
	ReplicaCount int
}

func scaleApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &ApplicationScaleRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]

	app := &types.Application{}
	pool := &types.PoolInfo{}

	utils.GetMgoCollections(ctx, w, []string{"application", "pool"}, func(cs map[string]*mgo.Collection) {
		colApplication, _ := cs["application"]

		if err := colApplication.FindId(bson.ObjectIdHex(id)).One(app); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的应用Id:"+id, http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		found := false
		for i, _ := range app.Services {
			if app.Services[i].Name == req.ServiceName {
				found = true
				break
			}
		}

		if found == false {
			HttpError(w, "在应用中没有这样的服务:"+req.ServiceName, http.StatusBadRequest)
			return
		}

		colPool, _ := cs["pool"]
		if err := colPool.FindId(bson.ObjectIdHex(app.PoolId)).One(pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的集群Id:"+app.PoolId, http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := application.ScaleApplication(ctx, app, pool, map[string]int{
			req.ServiceName: req.ReplicaCount,
		}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, "")
	})

}

type ApplicationUpgradeRequest struct {
	ApplicationTemplateId string
}

type ApplicationUpgradeResponse struct {
	types.Application
}

func upgradeApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := &ApplicationUpgradeRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.Debugf("upgradeApplication::recving a request %#v", req)

	id := mux.Vars(r)["id"]

	app := &types.Application{}
	pool := &types.PoolInfo{}
	template := &types.Template{}

	utils.GetMgoCollections(ctx, w, []string{"application", "pool", "template"}, func(cs map[string]*mgo.Collection) {
		colApplication, _ := cs["app"]

		if err := colApplication.FindId(bson.ObjectIdHex(id)).One(app); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的应用Id:"+id, http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		colTemplate, _ := cs["template"]
		if err := colTemplate.FindId(bson.ObjectIdHex(req.ApplicationTemplateId)).One(template); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的集群Id:"+app.PoolId, http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if app.Name != template.Name {
			HttpError(w, "升级应用时，应用Id必须一致！", http.StatusBadRequest)
			return
		}

		colPool, _ := cs["pool"]
		if err := colPool.FindId(bson.ObjectIdHex(app.PoolId)).One(pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的集群Id:"+app.PoolId, http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		app.TemplateId = req.ApplicationTemplateId
		app.Version = template.Version
		currentUser, _ := utils.GetCurrentUser(ctx)
		app.UpdatedTime = time.Now().Unix()
		app.UpdaterId = currentUser.Id.Hex()
		app.UpdaterName = currentUser.Name
		app.Status = "running"

		services, err := mergeServices(ctx, template.Services, pool)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		app.Services = services

		if err := colApplication.UpdateId(bson.ObjectIdHex(id), app); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := application.UpApplication(ctx, app, pool); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, "")
	})

}

func stopApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	utils.GetMgoCollections(ctx, w, []string{"application"}, func(cs map[string]*mgo.Collection) {

		app := &types.Application{}

		colApplication, _ := cs["application"]
		if err := colApplication.FindId(bson.ObjectIdHex(id)).One(app); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的应用", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pool := &types.PoolInfo{}

		colPool, _ := cs["pool"]
		if err := colPool.FindId(bson.ObjectIdHex(app.PoolId)).One(pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的集群Id:"+app.PoolId, http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		services := make([]string, len(app.Services))

		for i, _ := range app.Services {
			services = append(services, app.Services[i].Name)
		}

		if err := application.StopApplication(ctx, app, pool, services); err != nil {
			HttpError(w, "停止应用失败："+err.Error(), http.StatusInternalServerError)
			return
		}

		currentUser, _ := utils.GetCurrentUser(ctx)

		app.UpdatedTime = time.Now().Unix()
		app.UpdaterId = currentUser.Id.Hex()
		app.UpdaterName = currentUser.Name
		app.Status = "stopped"

		if err := colApplication.UpdateId(bson.ObjectIdHex(id), app); err != nil {
			HttpError(w, "保存应用状态失败："+err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, app)

	})

}

//TODO 这个接口非常危险
func restartApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func startApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]
	utils.GetMgoCollections(ctx, w, []string{"application"}, func(cs map[string]*mgo.Collection) {

		app := &types.Application{}

		colApplication, _ := cs["application"]
		if err := colApplication.FindId(bson.ObjectIdHex(id)).One(app); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的应用", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pool := &types.PoolInfo{}

		colPool, _ := cs["pool"]
		if err := colPool.FindId(bson.ObjectIdHex(app.PoolId)).One(pool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "没有这样的集群Id:"+app.PoolId, http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		services := make([]string, len(app.Services))

		for i, _ := range app.Services {
			services = append(services, app.Services[i].Name)
		}

		if err := application.StartApplication(ctx, app, pool, services); err != nil {
			HttpError(w, "启动应用失败："+err.Error(), http.StatusInternalServerError)
			return
		}

		currentUser, _ := utils.GetCurrentUser(ctx)

		app.UpdatedTime = time.Now().Unix()
		app.UpdaterId = currentUser.Id.Hex()
		app.UpdaterName = currentUser.Name
		app.Status = "running"

		if err := colApplication.UpdateId(bson.ObjectIdHex(id), app); err != nil {
			HttpError(w, "保存应用状态失败："+err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, app)

	})

}

type ApplicationInspectResponseUser struct {
	Id   string
	Name string
}

type ApplicationInspectResponseTeam struct {
	Id   string
	Name string
}

type ApplicationInspectResponse struct {
	Application types.Application
	Users       []ApplicationInspectResponseUser
	Teams       []ApplicationInspectResponseTeam
}

func getApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	utils.GetMgoCollections(ctx, w, []string{"application", "team", "user"}, func(cs map[string]*mgo.Collection) {
		app := types.Application{}
		if err := cs["pool"].Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&app); err != nil {

			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个application，不应该用500错误
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var selector bson.M

		//查找该app所在的Team
		teams := make([]types.Team, 0, 10)
		selector = bson.M{
			"applicationids": bson.M{
				"$in": bson.ObjectIdHex(id),
			},
		}
		if err := cs["team"].Find(selector).All(&teams); err != nil {

			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//查找该app所在的User
		users := make([]types.User, 0, 10)
		selector = bson.M{
			"applicationids": bson.M{
				"$in": bson.ObjectIdHex(id),
			},
		}
		if err := cs["user"].Find(selector).All(&users); err != nil {

			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//整理数据格式
		rlt := ApplicationInspectResponse{}
		rlt.Application = app
		rlt.Teams = make([]ApplicationInspectResponseTeam, 0, len(teams))
		for _, t := range teams {
			rt := ApplicationInspectResponseTeam{
				Id:   t.Id.Hex(),
				Name: t.Name,
			}
			rlt.Teams = append(rlt.Teams, rt)
		}
		rlt.Users = make([]ApplicationInspectResponseUser, 0, len(users))
		for _, u := range users {
			ru := ApplicationInspectResponseUser{
				Id:   u.Id.Hex(),
				Name: u.Name,
			}
			rlt.Users = append(rlt.Users, ru)
		}

		HttpOK(w, rlt)
	})
}

func rollbackApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

/*
/applications/:id/add-team
	请求参数：
		TeamId
	返回：无
权限控制：应用管理员。
*/
func addApplicationTeam(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var appId string
	var teamId string

	//检查参数合法性
	if appId = mux.Vars(r)["id"]; len(appId) <= 0 {
		HttpError(w, "Application Id is empty", http.StatusBadRequest)
		return
	}
	if teamId = r.FormValue("TeamId"); len(teamId) <= 0 {
		HttpError(w, "TeamId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"team", "application"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["application"].FindId(bson.ObjectIdHex(appId)).Count(); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else if c <= 0 {
			HttpError(w, "", http.StatusNotFound)
			return
		}

		if err := cs["team"].Update(bson.M{"_id": bson.ObjectIdHex(teamId)}, bson.M{"$addToSet": bson.M{"applicationids": bson.ObjectIdHex(appId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})

}

/*
/applications/:id/remove-team
	请求参数：
		TeamId
	返回：无
权限控制：应用管理员。
*/
func removeApplicationTeam(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var appId string
	var teamId string

	//检查参数合法性
	if appId = mux.Vars(r)["id"]; len(appId) <= 0 {
		HttpError(w, "Application Id is empty", http.StatusBadRequest)
		return
	}
	if teamId = r.FormValue("TeamId"); len(teamId) <= 0 {
		HttpError(w, "TeamId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"team", "application"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["application"].FindId(bson.ObjectIdHex(appId)).Count(); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else if c <= 0 {
			HttpError(w, "", http.StatusNotFound)
			return
		}

		if err := cs["team"].Update(bson.M{"_id": bson.ObjectIdHex(teamId)}, bson.M{"$pull": bson.M{"applicationids": bson.ObjectIdHex(appId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

/*

/applications/:id/add-user
	请求参数：
		UserId
	返回：无
权限控制：应用管理员。
*/
func addApplicationMember(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var appId string
	var userId string

	//检查参数合法性
	if appId = mux.Vars(r)["id"]; len(appId) <= 0 {
		HttpError(w, "Application Id is empty", http.StatusBadRequest)
		return
	}
	if userId = r.FormValue("UserId"); len(userId) <= 0 {
		HttpError(w, "UserId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"user", "application"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["application"].FindId(bson.ObjectIdHex(appId)).Count(); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else if c <= 0 {
			HttpError(w, "", http.StatusNotFound)
			return
		}

		if err := cs["user"].Update(bson.M{"_id": bson.ObjectIdHex(userId)}, bson.M{"$addToSet": bson.M{"applicationids": bson.ObjectIdHex(appId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

/*

/applications/:id/remove-user
	请求参数：
		UserId
	返回：无
权限控制：应用管理员。
*/
func removeApplicationMember(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var appId string
	var userId string

	//检查参数合法性
	if appId = mux.Vars(r)["id"]; len(appId) <= 0 {
		HttpError(w, "Application Id is empty", http.StatusBadRequest)
		return
	}
	if userId = r.FormValue("UserId"); len(userId) <= 0 {
		HttpError(w, "UserId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"user", "application"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["application"].FindId(bson.ObjectIdHex(appId)).Count(); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else if c <= 0 {
			HttpError(w, "", http.StatusNotFound)
			return
		}

		if err := cs["user"].Update(bson.M{"_id": bson.ObjectIdHex(userId)}, bson.M{"$pull": bson.M{"applicationids": bson.ObjectIdHex(appId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}
