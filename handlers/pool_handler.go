package handlers

import (
	"context"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/proxy"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
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

type PoolsRegisterRequest struct {
	Name       string
	Driver     string
	DriverOpts types.DriverOpts
	Labels     []string `json:",omitempty"`
}

type PoolsRegisterResponse struct {
	Id    string
	Name  string
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

	result := make([]types.PoolInfo, 0, 20)
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
	Containers        int
	ContainersRunning int
	ContainersPaused  int
	ContainersStopped int
}
type PoolsFlushResponse struct {
	types.PoolInfo
	Nodes   []types.Node
	Runtime PoolRuntimeInfo
}

func postPoolsFlush(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	var result PoolsFlushResponse

	id := mux.Vars(r)["id"]

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)

	c := mgoSession.DB(config.MgoDB).C("pool")

	if err := c.FindId(bson.ObjectIdHex(id)).One(&result.PoolInfo); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//logrus.Debugf("postPoolsFlush::the pool info is %#v",result.PoolInfo)
	if result.PoolInfo.Driver != "swarm" {
		HttpError(w, "目前只支持swarm集群", http.StatusInternalServerError)
		return
	}

	if len(result.ProxyEndpoints) == 0 {
		HttpError(w, "没有集群代理", http.StatusInternalServerError)
		return
	}

	clusterInfo, err := utils.GetClusterInfo(ctx, result.ProxyEndpoints[0])
	if err != nil {
		HttpError(w, "获取集群信息错误"+err.Error(), http.StatusInternalServerError)
		return
	}

	//同步集群的静态信息，需要写到mongodb的pool表
	result.PoolInfo.Labels = clusterInfo.Labels
	result.PoolInfo.NCPU = clusterInfo.NCPU
	result.PoolInfo.MemTotal = clusterInfo.MemTotal
	result.PoolInfo.ClusterStore = clusterInfo.ClusterStore
	result.PoolInfo.ClusterAdvertise = clusterInfo.ClusterAdvertise

	strategy, filters, nodes, err := utils.ParseNodes(clusterInfo.SystemStatus, result.PoolInfo.Id.Hex(), result.PoolInfo.Name)
	if err != nil {
		HttpError(w, "解析集群节点信息错误"+err.Error(), http.StatusInternalServerError)
		return
	}

	result.PoolInfo.Strategy = strategy
	result.PoolInfo.Filters = filters
	result.Nodes = nodes

	logrus.Debugf("postPoolsFlush::result is %#v", result)

	//httpJsonResponse(w,result)

	if err := c.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{"labels": clusterInfo.Labels,
		"ncpu":              clusterInfo.NCPU,
		"memtotal":          clusterInfo.MemTotal,
		"clusterstore":      clusterInfo.ClusterStore,
		"clusteradvertise":  clusterInfo.ClusterAdvertise,
		"containers":        clusterInfo.Containers,
		"containerspaused":  clusterInfo.ContainersPaused,
		"containersrunning": clusterInfo.ContainersRunning,
		"containersstopped": clusterInfo.ContainersStopped,
		"updatedtime":       time.Now().Unix(),
		"strategy":          strategy,
		"filters":           filters,
	}}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.Debugf("postPoolsFlush:update pool success!")

	//同步集群的动态信息，不需要写到mongodb
	result.Runtime.Containers = clusterInfo.Containers
	result.Runtime.ContainersPaused = clusterInfo.ContainersPaused
	result.Runtime.ContainersRunning = clusterInfo.ContainersRunning
	result.Runtime.ContainersStopped = clusterInfo.ContainersStopped

	//TODO

	cNode := mgoSession.DB(config.MgoDB).C("node")

	if err := cNode.Update(bson.M{"poolid": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{"status": "offline"}}); err != nil {
		if err != mgo.ErrNotFound {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for _, node := range result.Nodes {
		if _, err := cNode.Upsert(bson.M{"poolid": bson.ObjectIdHex(id), "endpoint": node.Endpoint}, &node); err != nil {
			logrus.Infof("upsert the node %#v error:%s", node, err.Error())
			continue
		}
	}

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

	if name != "" {
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
		Name:  poolInfo.Name,
		Id:    poolInfo.Id.Hex(),
		Proxy: poolInfo.ProxyEndpoints[0],
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
