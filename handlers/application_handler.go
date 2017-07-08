package handlers

import (
	"context"
	"encoding/json"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
	"gopkg.in/mgo.v2"
)

type ApplicationCreateRequest struct {
	TemplateId, PoolId, Title, Description string
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
	if err := colPool.FindId(bson.ObjectIdHex(req.PoolId)).One(pool) ; err !=nil {
		if err==mgo.ErrNotFound{
			HttpErrorAndPanic(w, err.Error(),http.StatusNotFound)
		}
		HttpErrorAndPanic(w,err.Error(),http.StatusInternalServerError)
	}


	colTemplate := mgoSession.DB(config.MgoDB).C("template")
	template := &types.Template{}
	if err := colTemplate.FindId(bson.ObjectIdHex(req.PoolId)).One(template) ; err !=nil {
		if err==mgo.ErrNotFound{
			HttpErrorAndPanic(w, err.Error(),http.StatusNotFound)
		}
		HttpErrorAndPanic(w,err.Error(),http.StatusInternalServerError)
	}




	c := mgoSession.DB(config.MgoDB).C("application")

	app := &types.Application{}
	app.Id = bson.NewObjectId()
	app.PoolId = req.PoolId
	app.TemplateId = req.TemplateId
	app.Title = req.Title
	app.Description = req.Description
	app.PoolId = req.PoolId

	app.CreatorId = currentuser.Id.Hex()
	app.UpdaterId = currentuser.Id.Hex()
	app.UpdaterName = currentuser.Name
	app.CreatedTime = time.Now().Unix()
	app.UpdatedTime = time.Now().Unix()

	app.Services = mergeServices(template.Services,pool)


	if err := c.Insert(app); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}



	httpJsonResponse(w,app)
}



func mergeServices(services []types.Service, info *types.PoolInfo) []types.Service {

	//TODO
	return []types.Service{}
}

func getApplicationList(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func getContainerSSHInfo(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func scaleApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func upgradeApplication(ctx context.Context, w http.ResponseWriter, r *http.Request) {}

var stopApplication = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}

var restartApplication = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}

var startApplication = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}

var getApplication = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
var rollbackApplication = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
