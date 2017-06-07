package handlers

import (
	"context"
	"github.com/zanecloud/apiserver/store"
	"net/http"

	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	 "github.com/zanecloud/apiserver/proxy"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func getPoolJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	mgoSession , err := utils.GetMgoSession(ctx)

	if err!=nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w,err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	result := store.PoolInfo{}
	if err := c.Find(bson.M{"name": name}).One(&result); err != nil {

		if err == mgo.ErrNotFound {
			// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

const POOL_LABEL = "com.zanecloud.omega.pool"

var routes = map[string]map[string]Handler{
	"HEAD": {},
	"GET": {
		"/pools/{name:.*}/inspect": MgoSessionInject(getPoolJSON),
		"/pools/ps": MgoSessionInject(getPoolsJSON),
		"/pools/json": MgoSessionInject(getPoolsJSON),


	},
	"POST": {

		"/pools/register": MgoSessionInject(postPoolsRegister),
	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": OptionsHandler,
	},
}

func NewMainHandler(ctx context.Context) http.Handler {
	return NewHandler(ctx, routes)
}

type PoolsRegisterRequest struct {
	store.PoolInfo
}

func getPoolsJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	mgoSession , err := utils.GetMgoSession(ctx)

	if err!=nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w,err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	var result []*store.PoolInfo
	if err := c.Find(bson.M{}).All(&result); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

func postPoolsRegister(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		name = r.Form.Get("name")
	)

	req := PoolsRegisterRequest{
		store.PoolInfo{	},
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.PoolInfo.Name = name
	req.PoolInfo.Id = bson.NewObjectId()

	mgoSession, err := utils.GetMgoSession(ctx)

	if err!=nil  {
		//走不到这里的
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	n, err := c.Find(bson.M{"name": name}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n >= 1 {
		HttpError(w, "the pool is exist", http.StatusConflict)
		return
	}

	p, err := proxy.NewProxyInstanceAndStart(ctx, &req.PoolInfo)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.PoolInfo.ProxyEndpoints = make([]string, 1)
	req.PoolInfo.ProxyEndpoints[0] = p.Endpoint()
	req.PoolInfo.Status = "running"

	if err = c.Insert(&req.PoolInfo); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Name", name)

}
