package handlers

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/proxy"
	 "github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"

	"fmt"
	"github.com/docker/swarm/swarmclient"
)

func getPoolJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	mgoSession, err := utils.GetMgoSessionClone(ctx)

	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	result := types.PoolInfo{}
	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result); err != nil {

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

type PoolsRegisterRequest struct{
	Name       string
	Driver     string
	DriverOpts types.DriverOpts
	Labels     []string `json:",omitempty"`
}


type PoolsRegisterResponse struct {
	Id   string
	Name string
	Proxy string
}
func getPoolsJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	mgoSession, err := utils.GetMgoSessionClone(ctx)

	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	result :=  make( []types.PoolInfo,20)
	if err := c.Find(bson.M{}).All(&result); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

//   /pools/{id:.*}/flush
type PoolsFlushRequest struct {
	Id string
}
type PoolRuntimeInfo struct {
	Containers         int
	ContainersRunning  int
	ContainersPaused   int
	ContainersStopped  int
}
type PoolsFlushResponse struct {
	types.PoolInfo
	//dockertypes.Info
	Runtime  PoolRuntimeInfo
}
func postPoolsFlush(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	var result PoolsFlushResponse

	id := mux.Vars(r)["id"]

	mgoSession , err := utils.GetMgoSessionClone(ctx)
	if err!=nil {
		HttpError(w, err.Error(),http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c:= mgoSession.DB(config.MgoDB).C("pool")


	if  err := c.FindId(bson.ObjectIdHex(id)).One(&result.PoolInfo) ; err !=nil {
		HttpError(w, err.Error(),http.StatusInternalServerError)
		return
	}

	//logrus.Debugf("postPoolsFlush::the pool info is %#v",result.PoolInfo)
	if result.PoolInfo.Driver != "swarm" {
		HttpError(w, "目前只支持swarm pool",http.StatusInternalServerError)
		return
	}


	if len(result.ProxyEndpoints)==0 {
		HttpError(w, "没有本地Pool 代理", http.StatusInternalServerError)
		return
	}

	client , err := client.NewClient(result.ProxyEndpoints[0],result.DriverOpts.Version,nil,nil)

	if err != nil {
		HttpError(w, fmt.Sprintf("连接后端集群%s失败,原因是:%s",result.ProxyEndpoints[0],err.Error()), http.StatusInternalServerError)
		return
	}
	defer  client.Close()

	info , err := client.Info(ctx)

	if err != nil {
		HttpError(w, fmt.Sprintf("同步后端集群%s失败,原因是:%s",result.ProxyEndpoints[0],err.Error()), http.StatusInternalServerError)
		return
	}

        //同步集群的静态信息，需要写到mongodb的pool表
	result.PoolInfo.Labels =  info.Labels
	result.PoolInfo.NCPU = info.NCPU
	result.PoolInfo.MemTotal = info.MemTotal
	result.PoolInfo.ClusterStore = info.ClusterStore
	result.PoolInfo.ClusterAdvertise = info.ClusterAdvertise

	//同步集群的动态信息，不需要写到mongodb
	result.Runtime.Containers = info.Containers
	result.Runtime.ContainersPaused  = info.ContainersPaused
	result.Runtime.ContainersRunning = info.ContainersRunning
	result.Runtime.ContainersStopped = info.ContainersStopped


	logrus.Debugf("postPoolsFlush::result is %#v",result)
	
	//httpJsonResponse(w,result)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}
func postPoolsRegister(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		name = r.Form.Get("Name")
	)

	req := PoolsRegisterRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}


	poolInfo := &types.PoolInfo{
		Id:             bson.NewObjectId(),
		Driver:         req.Driver,
		DriverOpts:     req.DriverOpts,
		Labels:         req.Labels,
		Name:           req.Name,
		ProxyEndpoints: make([]string, 1),
	}


	if name!= "" {
		poolInfo.Name = name
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)

	if err != nil {
		//走不到这里的
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	n, err := c.Find(bson.M{"name": poolInfo.Name}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n >= 1 {
		HttpError(w, "the pool is exist", http.StatusConflict)
		return
	}

	p, err := proxy.NewProxyInstanceAndStart(ctx, poolInfo)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	poolInfo.ProxyEndpoints[0] = p.Endpoint()
	poolInfo.Status = "running"

	if err = c.Insert(poolInfo); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := PoolsRegisterResponse{
		Name: poolInfo.Name,
		Id : poolInfo.Id.Hex(),
		Proxy : poolInfo.ProxyEndpoints[0],
	}
	//httpJsonResponse(w, result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
	//fmt.Fprintf(w, "{%q:%q}", "Name", name)

}
func httpJsonResponse(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
