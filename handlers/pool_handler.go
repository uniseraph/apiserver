package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/proxy"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	regexp "regexp"
	"strings"
	"time"
)

type PoolInspectResponseUser struct {
	Id   string
	Name string
}

type PoolInspectResponseTeam struct {
	Id   string
	Name string
}

type PoolInspectResponse struct {
	Pool  types.PoolInfo
	Users []PoolInspectResponseUser
	Teams []PoolInspectResponseTeam
}

func getPoolJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	utils.GetMgoCollections(ctx, w, []string{"pool", "team", "user"}, func(cs map[string]*mgo.Collection) {
		pool := types.PoolInfo{}
		if err := cs["pool"].Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&pool); err != nil {

			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var selector bson.M

		//查找该pool所在的Team
		teams := make([]types.Team, 0, 10)
		selector = bson.M{
			"poolids": bson.ObjectIdHex(id),
		}
		if err := cs["team"].Find(selector).All(&teams); err != nil {

			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//查找该pool所在的User
		users := make([]types.User, 0, 10)
		selector = bson.M{
			"poolids": bson.ObjectIdHex(id),
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
		rlt := PoolInspectResponse{}
		rlt.Pool = pool
		rlt.Teams = make([]PoolInspectResponseTeam, 0, len(teams))
		for _, t := range teams {
			rt := PoolInspectResponseTeam{
				Id:   t.Id.Hex(),
				Name: t.Name,
			}
			rlt.Teams = append(rlt.Teams, rt)
		}
		rlt.Users = make([]PoolInspectResponseUser, 0, len(users))
		for _, u := range users {
			ru := PoolInspectResponseUser{
				Id:   u.Id.Hex(),
				Name: u.Name,
			}
			rlt.Users = append(rlt.Users, ru)
		}

		HttpOK(w, rlt)
	})
}

type PoolsRegisterRequest struct {
	Name       string
	Driver     string
	EnvTreeId  string
	DriverOpts types.DriverOpts
	Provider   string
	Labels     []string `json:",omitempty"`
}

type PoolsRegisterResponse struct {
	Id    string
	Name  string
	Proxy string
}

func getPoolsJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"pool", "team"}, func(cs map[string]*mgo.Collection) {
		poolSelector := bson.M{}
		poolIds := make([]bson.ObjectId, 0, 20)

		user, err := getCurrentUser(ctx)
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
		poolIds = append(poolIds, user.PoolIds...)

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
				poolIds = append(poolIds, team.PoolIds...)
			}
		}

		poolSelector["_id"] = bson.M{
			"$in": poolIds,
		}

	AUTHORIZED:

		pools := make([]*types.PoolInfo, 0, 20)
		if err := cs["pool"].Find(poolSelector).All(&pools); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, pools)

	})

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

func refreshPool(ctx context.Context, id string) (*PoolsFlushResponse, error) {

	result := &PoolsFlushResponse{}
	err := utils.ExecMgoCollections(ctx, []string{"pool", "node"}, func(cs map[string]*mgo.Collection) error {

		colPool, _ := cs["pool"]
		colNode, _ := cs["node"]

		if err := colPool.FindId(bson.ObjectIdHex(id)).One(&result.PoolInfo); err != nil {
			return err
		}

		if result.PoolInfo.Driver != "swarm" {
			//HttpError(w, "目前只支持swarm集群", http.StatusInternalServerError)
			return errors.New("目前只支持swarm集群")
		}

		if len(result.ProxyEndpoint) == 0 {
			return errors.New("没有集群代理")
		}

		clusterInfo, err := utils.GetClusterInfo(ctx, result.ProxyEndpoint)
		if err != nil {
			return errors.New("获取集群信息错误" + err.Error())
		}

		//同步集群的静态信息，需要写到mongodb的pool表
		result.PoolInfo.Labels = clusterInfo.Labels
		result.PoolInfo.CPUs = clusterInfo.NCPU
		result.PoolInfo.Memory = clusterInfo.MemTotal
		result.PoolInfo.ClusterStore = clusterInfo.ClusterStore
		result.PoolInfo.ClusterAdvertise = clusterInfo.ClusterAdvertise
		result.PoolInfo.Containers = clusterInfo.Containers

		result.PoolInfo.Provider = "native"

		strategy, filters, nodes, err := utils.ParseNodes(clusterInfo.SystemStatus, &result.PoolInfo)
		if err != nil {
			return errors.New("解析集群节点信息错误" + err.Error())
		}

		result.PoolInfo.Strategy = strategy
		result.PoolInfo.Filters = filters
		result.PoolInfo.NodeCount = len(nodes)
		result.Nodes = nodes

		if err := getTunneldInfo(ctx, &result.PoolInfo); err != nil {
			logrus.Errorf("flush tunneld info err : %s", err.Error())
			// 这个错误不要紧，下次在刷
		}

		logrus.Debugf("postPoolsFlush::result is %#v", result)

		//httpJsonResponse(w,result)

		if err := colPool.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": bson.M{"labels": result.PoolInfo.Labels,
			"cpus":              clusterInfo.NCPU,
			"memory":            clusterInfo.MemTotal,
			"clusterstore":      clusterInfo.ClusterStore,
			"clusteradvertise":  clusterInfo.ClusterAdvertise,
			"containers":        clusterInfo.Containers,
			"containerspaused":  clusterInfo.ContainersPaused,
			"containersrunning": clusterInfo.ContainersRunning,
			"containersstopped": clusterInfo.ContainersStopped,
			"updatedtime":       time.Now().Unix(),
			"strategy":          strategy,
			"filters":           filters,
			"tunneldaddr":       result.PoolInfo.TunneldAddr,
			"tunneldport":       result.PoolInfo.TunneldPort,
			"nodecount":         len(nodes),
			"provider":          result.PoolInfo.Provider,
			"lb":                result.PoolInfo.LB,
		}}); err != nil {
			return err
		}

		logrus.Debugf("postPoolsFlush:update pool success!")

		//同步集群的动态信息，不需要写到mongodb
		result.Runtime.Containers = clusterInfo.Containers
		result.Runtime.ContainersPaused = clusterInfo.ContainersPaused
		result.Runtime.ContainersRunning = clusterInfo.ContainersRunning
		result.Runtime.ContainersStopped = clusterInfo.ContainersStopped

		if err := colNode.Update(bson.M{"poolid": id}, bson.M{"$set": bson.M{"status": "offline"}}); err != nil {
			if err != mgo.ErrNotFound {
				return err
			}
		}

		for _, node := range result.Nodes {
			if _, err := colNode.Upsert(bson.M{"poolid": id, "endpoint": node.Endpoint}, &node); err != nil {
				logrus.Infof("upsert the node %#v error:%s", node, err.Error())
				continue
			}
		}

		return nil

	})

	if err != nil {

		return nil, err
	}

	return result, nil
}

func postPoolsFlush(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	result, err := refreshPool(ctx, id)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func getTunneldInfo(ctx context.Context, pool *types.PoolInfo) error {

	re := regexp.MustCompile(`^tcp://(.+):`)

	// tcp://1.1.1.1:
	str := re.FindString(pool.DriverOpts.EndPoint)
	str = strings.TrimLeft(str, "tcp://")
	str = strings.TrimRight(str, ":")

	metadUrl := fmt.Sprintf("http://%s:6400/services/tunneld/inspect", str)

	logrus.Debugf("the pool1's metad url is %s", metadUrl)

	cli := &http.Client{}

	resp, err := cli.Get(metadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf(string(buf))
	}

	tunneld := &types.AgentService{}
	if err := json.NewDecoder(resp.Body).Decode(tunneld); err != nil {
		return err
	}

	pool.TunneldAddr = tunneld.Address
	pool.TunneldPort = tunneld.Port
	return nil

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

	if req.DriverOpts.APIVersion == "" {
		HttpError(w, "APIVersion 不能是空", http.StatusBadRequest)
		return
	}

	poolInfo := &types.PoolInfo{
		Id:          bson.NewObjectId(),
		Driver:      req.Driver,
		DriverOpts:  req.DriverOpts,
		Labels:      req.Labels,
		Name:        req.Name,
		EnvTreeId:   req.EnvTreeId,
		CreatedTime: time.Now().Unix(),
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

	logrus.Debugf("register pool : no conflict pool")

	colEnvTree := mgoSession.DB(mgoDB).C("env_tree_meta")
	envTree := &types.EnvTreeMeta{}
	if err := colEnvTree.FindId(bson.ObjectIdHex(req.EnvTreeId)).One(envTree); err != nil {
		HttpError(w, "没有这样的env_tree:"+req.EnvTreeId, http.StatusNotFound)
		return
	}

	poolInfo.EnvTreeName = envTree.Name

	logrus.Debugf("register pool: set env tree success")

	p, err := proxy.NewProxyInstanceAndStart(utils.GetAPIServerConfig(ctx), poolInfo)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	poolInfo.ProxyEndpoint = p.Endpoint()
	poolInfo.Status = "running"

	if err = c.Insert(poolInfo); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := PoolsRegisterResponse{
		Name:  poolInfo.Name,
		Id:    poolInfo.Id.Hex(),
		Proxy: poolInfo.ProxyEndpoint,
	}

	refreshPool(ctx, poolInfo.Id.Hex())

	//httpJsonResponse(w, result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	/*
		系统审计
	*/

	utils.CreateSystemAuditLogWithCtx(ctx, r, types.SystemAuditModuleTypePool, types.SystemAuditModuleOperationTypeCreate, poolInfo.Id.Hex(), "", map[string]interface{}{"Pool": poolInfo})
}

/*
/pools/:id/add-team

请求参数：
	TeamId
返回：无
权限控制：系统管理员。
*/
func addPoolTeam(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var poolId string
	var teamId string

	//检查参数合法性
	if poolId = mux.Vars(r)["id"]; len(poolId) <= 0 {
		HttpError(w, "PoolId is empty", http.StatusBadRequest)
		return
	}
	if teamId = r.FormValue("TeamId"); len(teamId) <= 0 {
		HttpError(w, "TeamId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"team", "pool"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["pool"].FindId(bson.ObjectIdHex(poolId)).Count(); err != nil {
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

		if err := cs["team"].Update(bson.M{"_id": bson.ObjectIdHex(teamId)}, bson.M{"$addToSet": bson.M{"poolids": bson.ObjectIdHex(poolId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})

}

/*
/pools/:id/remove-team

请求参数：
	TeamId
返回：无
权限控制：系统管理员。
*/
func removePoolTeam(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var poolId string
	var teamId string

	//检查参数合法性
	if poolId = mux.Vars(r)["id"]; len(poolId) <= 0 {
		HttpError(w, "PoolId is empty", http.StatusBadRequest)
		return
	}
	if teamId = r.FormValue("TeamId"); len(teamId) <= 0 {
		HttpError(w, "TeamId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"team", "pool"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["pool"].FindId(bson.ObjectIdHex(poolId)).Count(); err != nil {
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

		if err := cs["team"].Update(bson.M{"_id": bson.ObjectIdHex(teamId)}, bson.M{"$pull": bson.M{"poolids": bson.ObjectIdHex(poolId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

/*

请求参数：
	UserId
返回：无
权限控制：系统管理员。
/pools/:id/add-user
*/
func addPoolMember(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var poolId string
	var userId string

	//检查参数合法性
	if poolId = mux.Vars(r)["id"]; len(poolId) <= 0 {
		HttpError(w, "PoolId is empty", http.StatusBadRequest)
		return
	}
	if userId = r.FormValue("UserId"); len(userId) <= 0 {
		HttpError(w, "UserId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"user", "pool"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["pool"].FindId(bson.ObjectIdHex(poolId)).Count(); err != nil {
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

		if err := cs["user"].Update(bson.M{"_id": bson.ObjectIdHex(userId)}, bson.M{"$addToSet": bson.M{"poolids": bson.ObjectIdHex(poolId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

/*

请求参数：
	UserId
返回：无
权限控制：系统管理员。
/pools/:id/remove-user
*/
func removePoolMember(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var poolId string
	var userId string

	//检查参数合法性
	if poolId = mux.Vars(r)["id"]; len(poolId) <= 0 {
		HttpError(w, "PoolId is empty", http.StatusBadRequest)
		return
	}
	if userId = r.FormValue("UserId"); len(userId) <= 0 {
		HttpError(w, "UserId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"user", "pool"}, func(cs map[string]*mgo.Collection) {
		//检查PoolId合法性
		if c, err := cs["pool"].FindId(bson.ObjectIdHex(poolId)).Count(); err != nil {
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

		if err := cs["user"].Update(bson.M{"_id": bson.ObjectIdHex(userId)}, bson.M{"$pull": bson.M{"poolids": bson.ObjectIdHex(poolId)}}); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

/*
	更新Pool信息
*/

type PoolUpdateRequest struct {
	Name string
}

func updatePool(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if len(id) <= 0 {
		HttpError(w, "Application Id could not be empty", http.StatusBadRequest)
		return
	}

	req := &PoolUpdateRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	/*
		校验入参合法性
		如果一个需要更新的入参都没有
		则不必执行
	*/

	//校验条件是
	//各个属性至少要有一个是有值的
	//!(len(req.Name) > 0 || len(req.Name) > 0)
	if !(len(req.Name) > 0) {
		HttpError(w, "至少有一个参数需要更新，不能没有需要更新的参数", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"pool"}, func(cs map[string]*mgo.Collection) {
		/*
			系统审计
		*/
		oldPool := &types.PoolInfo{}

		if err := cs["pool"].FindId(bson.ObjectIdHex(id)).One(oldPool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//需要更新的属性一条条加进去
		data := bson.M{}
		data["name"] = req.Name

		if err := cs["pool"].UpdateId(bson.ObjectIdHex(id), bson.M{"$set": data}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*
			系统审计
		*/

		newPool := &types.PoolInfo{}
		if err := cs["pool"].FindId(bson.ObjectIdHex(id)).One(newPool); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*
			系统审计
		*/

		logData := map[string]interface{}{"NewPool": newPool, "OldPool": oldPool}
		utils.CreateSystemAuditLogWithCtx(ctx, r, types.SystemAuditModuleTypePool, types.SystemAuditModuleOperationTypeUpdate, newPool.Id.Hex(), "", logData)
	})
}

/*
	删除Pool
	删除时先检查本Pool是否还有应用，如果有应用应拒绝删除。
	删除Pool时，应同步删除本Pool相关的Pool EnvValue。
*/

func deletePool(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if len(id) <= 0 {
		HttpError(w, "Application Id could not be empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"pool", "application", "env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		/*
			系统审计
		*/
		pool := &types.PoolInfo{}
		if err := cs["pool"].FindId(bson.ObjectIdHex(id)).One(pool); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var selector bson.M

		selector = bson.M{
			"poolid": id,
		}

		//检查是否还有应用
		if c, err := cs["application"].Find(selector).Count(); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else if c > 0 {
			//说明还有应用
			HttpError(w, "该集群还有应用存在，请先删除应用再删除集群", http.StatusInternalServerError)
			return
		}

		if err := proxy.Stop(pool.Id.Hex()); err != nil {
			HttpError(w, "关闭该集群代理失败"+err.Error(), http.StatusInternalServerError)
			return
		}

		//删除集群相应的EnvValue
		selector = bson.M{
			"pool": bson.ObjectIdHex(id),
		}

		if info, err := cs["env_tree_node_param_value"].RemoveAll(selector); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			logrus.Infof("Delete Pool[%v] related envs: %v", id, info)
		}

		//删除集群

		if err := cs["pool"].RemoveId(bson.ObjectIdHex(id)); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, "删除集群成功")

		/*
			系统审计
		*/

		logData := map[string]interface{}{"Pool": pool}
		utils.CreateSystemAuditLogWithCtx(ctx, r, types.SystemAuditModuleTypePool, types.SystemAuditModuleOperationTypeDelete, pool.Id.Hex(), "", logData)

	})
}
