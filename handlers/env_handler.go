package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gpmgo/gopm/modules/log"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"sort"
	"strconv"
	"time"
)

/*
/envs/trees/list
/envs/trees/create
/envs/trees/:id/update
/envs/trees/:id/remove
*/

/*
	Request
*/
type EnvTreeMetaRequest struct {
	types.EnvTreeMeta
}

type EnvTreeNodeDirRequest struct {
	Name     string
	ParentId string
	TreeId   string
}

type EnvTreeNodeParamKeyRequest struct {
	types.EnvTreeNodeParamKey
}

type EnvTreeNodeParamValueRequest struct {
	types.EnvTreeNodeParamValue
}

//参数键值对
type EnvTreeNodeParamKVRequest struct {
	Id          string `json:",omitempty"`
	Name        string
	Mask        bool
	Value       string
	Description string
	DirId       string `json:",omitempty"`
	TreeId      string `json:",omitempty"`
}

/*
	Response
*/
type EnvTreeMetaResponse struct {
	Id          string
	Name        string
	Description string
	Root        string
	CreatedTime int64
}

type EnvTreeNodeDirResponse struct {
	Id       string
	Name     string
	ParentId string
	TreeId   string
}

//用于返回树形结构
type EnvTreeNodeDirsResponse struct {
	Id   bson.ObjectId "_id"
	Name string
	//一个父目录
	//最顶级的父目录为空，用于结合EnvTreeMeta查询该树的起点
	//EnvTreeNodeDir
	ParentId string
	//多个子目录
	//EnvTreeNodeDir
	//Children    []*EnvTreeNodeDirsResponse
	Children    EnvTreeNodeDirsResponseSlice
	CreatedTime int64 `json:",omitempty"`
	UpdatedTime int64 `json:",omitempty"`
}

type EnvTreeNodeParamKeyResponse struct {
	types.EnvTreeNodeParamKey
}

type EnvTreeNodeParamValueResponse struct {
	types.EnvTreeNodeParamValue
}

//参数键值对
type EnvTreeNodeParamKVResponse struct {
	Id          string
	Name        string
	Value       string
	Description string
	DirId       string `json:",omitempty"`
	TreeId      string `json:",omitempty"`
}

func getTrees(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"env_tree_meta"}, func(cs map[string]*mgo.Collection) {
		results := make([]types.EnvTreeMeta, 30)
		if err := cs["env_tree_meta"].Find(bson.M{}).All(&results); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, results)
	})
}

//创建参数目录树元数据
func createTree(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvTreeMetaRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_meta", "env_tree_node_dir"}, func(cs map[string]*mgo.Collection) {
		//创建树的元数据
		tree := &types.EnvTreeMeta{
			Id:          bson.NewObjectId(),
			Name:        req.Name,
			Description: req.Description,
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
		}
		if err := cs["env_tree_meta"].Insert(tree); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//创建一个节点目录
		dir := &types.EnvTreeNodeDir{
			Id:          bson.NewObjectId(),
			Name:        "全部",
			Tree:        tree.Id,
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
		}

		//创建根节点
		if err := cs["env_tree_node_dir"].Insert(dir); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := &EnvTreeMetaResponse{
			Id:          tree.Id.Hex(),
			Name:        tree.Name,
			Root:        dir.Id.Hex(),
			Description: tree.Description,
			CreatedTime: tree.CreatedTime,
		}
		HttpOK(w, resp)
	})
}

func updateTree(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvTreeMetaRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_meta"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]

		data := bson.M{}
		if req.Name != "" {
			data["name"] = req.Name
		}

		if req.Description != "" {
			data["description"] = req.Description
		}
		data["updatedTime"] = time.Now().Unix()

		selector := bson.M{"_id": bson.ObjectIdHex(id)}

		if err := cs["env_tree_meta"].Update(selector, bson.M{"$set": data}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data["Id"] = id

		HttpOK(w, data)
	})

}

func deleteTree(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"env_tree_meta", "env_tree_node_dir", "env_tree_node_param_key", "env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]

		if err := cs["env_tree_meta"].Remove(bson.M{"_id": bson.ObjectIdHex(id)}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		selector := bson.M{
			"tree": bson.ObjectIdHex(id),
		}

		tables := []string{"env_tree_node_dir", "env_tree_node_param_key", "env_tree_node_param_value"}

		//如果删除树信息
		//则删除跟树相关的一系列信息
		//包括：目录节点，参数节点，以及参数值
		for _, t := range tables {
			if info, err := cs[t].RemoveAll(selector); err != nil {
				if err == mgo.ErrNotFound {
					HttpError(w, fmt.Sprintf("%#v", info), http.StatusNotFound)
					return
				}
				HttpError(w, fmt.Sprintf("%#v", info), http.StatusInternalServerError)
				return
			}
		}

		//TODO
		//还要清理跟POOL有关的信息

		HttpOK(w, nil)
	})
}

/*
/envs/dirs/list
/envs/dirs/create
/envs/dirs/:id/update
/envs/dirs/:id/remove
*/
func getTreeDirs(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_dir", "env_tree_meta"}, func(cs map[string]*mgo.Collection) {
		var id = r.Form.Get("TreeId")
		var idObject = bson.ObjectIdHex(id)

		results := make([]types.EnvTreeNodeDir, 0, 20)

		data := bson.M{
			"tree": idObject,
		}

		if err := cs["env_tree_node_dir"].Find(data).All(&results); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//如果数的节点是空的
		//则返回空对象
		if len(results) <= 0 {
			HttpOK(w, EnvTreeNodeDirsResponse{})
			return
		}

		//树形结构的结构体
		rsp := &EnvTreeNodeDirsResponse{}

		if err := TreeBuild(rsp, results); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		} else {
			//HttpError(w, "Good!!!!!!!", http.StatusInternalServerError)
			HttpOK(w, rsp)
		}
	})
}

//创建一个目录节点
func createDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvTreeNodeDirRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.TreeId) <= 0 {
		HttpError(w, "TreeId is empty", http.StatusNotFound)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_dir", "env_tree_meta"}, func(cs map[string]*mgo.Collection) {
		//先检查挂靠的树结构是否存在
		//避免传递一个错误的ID
		meta := &types.EnvTreeMeta{}
		if err := cs["env_tree_meta"].FindId(bson.ObjectIdHex(req.TreeId)).One(&meta); err != nil {
			HttpError(w, "TreeId is invalide", http.StatusNotFound)
			return
		}

		//创建一个节点目录
		dir := &types.EnvTreeNodeDir{
			Id:          bson.NewObjectId(),
			Name:        req.Name,
			Tree:        bson.ObjectIdHex(req.TreeId),
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
		}

		if len(req.ParentId) > 0 {
			dir.Parent = bson.ObjectIdHex(req.ParentId)
		} else {
			//父级为空，则为根节点
		}
		//创建子节点
		if err := cs["env_tree_node_dir"].Insert(dir); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//创建子节点如果成功
		//则父节点的子节点列表增加该节点
		if len(req.ParentId) > 0 {
			data := bson.M{"children": dir.Id}
			selector := bson.M{"_id": bson.ObjectIdHex(req.ParentId)}

			if err := cs["env_tree_node_dir"].Update(selector, bson.M{"$addToSet": data}); err != nil {
				if err == mgo.ErrNotFound {
					HttpError(w, err.Error(), http.StatusNotFound)
					return
				}

				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		resp := &EnvTreeNodeDirResponse{
			Id:       dir.Id.Hex(),
			Name:     dir.Name,
			ParentId: dir.Parent.Hex(),
			TreeId:   dir.Tree.Hex(),
		}
		HttpOK(w, resp)
	})
}

//更新一个目录节点
func updateDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvTreeNodeDirRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_dir"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]

		data := bson.M{}
		if req.Name != "" {
			data["name"] = req.Name
		}

		data["updatedtime"] = time.Now().Unix()

		selector := bson.M{"_id": bson.ObjectIdHex(id)}

		if err := cs["env_tree_node_dir"].Update(selector, bson.M{"$set": data}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var dir = EnvTreeNodeDirResponse{
			Id:   id,
			Name: req.Name,
		}

		HttpOK(w, dir)
	})
}

//删除目录节点
func deleteDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_dir"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]

		p_dir := &types.EnvTreeNodeDir{}
		//检查id是否存在记录
		if err := cs["env_tree_node_dir"].FindId(bson.ObjectIdHex(id)).One(&p_dir); err != nil {
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		//目录节点下面没有参数值的情况下
		//才允许删除该目录
		if len(p_dir.Keys) > 0 {
			HttpError(w, "this dir stil has some params, please remove params first.", http.StatusInternalServerError)
			return
		}

		//pull方法从父节点children数组中删除自己的id记录
		// > db.env_tree_node_dir.update( {"_id":ObjectId("59604ce52010e16432af0ddd")}, {"$pull": {"children": ObjectId("59604ce52010e16432af0dde")}} )
		data := bson.M{
			"$pull": bson.M{
				"children": bson.ObjectIdHex(id),
			},
		}
		log.Info("PID: %s", p_dir.Parent.Hex())

		//从父级节点的children数组中
		//删除自己的记录，避免污染父节点
		//ERROR
		//TODO
		if err := cs["env_tree_node_dir"].UpdateId(p_dir.Parent, data); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree dir", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//删除自己
		if err := cs["env_tree_node_dir"].Remove(bson.M{"_id": bson.ObjectIdHex(id)}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree dir", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

/*
/envs/values/list
/envs/values/:id/detail
/envs/values/create
/envs/values/:id/update
/envs/values/:id/remove
/envs/values/:id/update-values
*/

/*
TreeId -- 树Id
DirId -- 目录节点Id，为空时表示返回所有参数，不为空时返回该目录下的参数
Name -- 名称前缀搜索，可以为空
PageSize -- 每页显示多少条
Page -- 当前页
*/
type EnvValuesListRequest struct {
	TreeId   string
	DirId    string
	Name     string
	PageSize int
	Page     int
}

type EnvValuesListResponse struct {
	Total     int
	PageCount int
	PageSize  int
	Page      int
	Data      EnvTreeNodeParamKVResponseSlice
}

//用于/envs/values/list结果中Data数组，按照Name排序
type EnvTreeNodeParamKVResponseSlice []*EnvTreeNodeParamKVResponse

func (c EnvTreeNodeParamKVResponseSlice) Len() int {
	return len(c)
}
func (c EnvTreeNodeParamKVResponseSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c EnvTreeNodeParamKVResponseSlice) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

func getTreeValues(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvValuesListRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key", "env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		var keys []types.EnvTreeNodeParamKey
		//避免没有结果的时候返回nil
		//需要没有结果的时候返回空数组
		results := make(EnvTreeNodeParamKVResponseSlice, 0, 20)

		data := bson.M{}

		//TreeId不可为空
		if len(req.TreeId) > 0 {
			data["tree"] = bson.ObjectIdHex(req.TreeId)
		} else {
			HttpError(w, "Need TreeId", http.StatusInternalServerError)
			return
		}

		if len(req.DirId) > 0 {
			data["dir"] = bson.ObjectIdHex(req.DirId)
		}

		//如果有name参数
		//则按照前缀匹配进行正则查找
		if len(req.Name) > 0 {
			data["name"] = bson.M{"$regex": bson.RegEx{Pattern: fmt.Sprintf("^%s", req.Name), Options: "i"}}
		}

		//查询所有条件匹配参数值的总数
		c, err := cs["env_tree_node_param_key"].Find(data).Count()
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info("query param size: %d", c)

		//处理分页
		if req.PageSize == 0 {
			req.PageSize = 20
		}

		//页面上分页的page要从1开始
		//数据库中查询要从0开始
		page := req.Page - 1
		if page < 0 {
			page = 0
		}
		//找到参数目录树中的全部匹配的KEY
		if err := cs["env_tree_node_param_key"].Find(data).Skip(page * req.PageSize).Limit(req.PageSize).All(&keys); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//整理成客户端需要的数据结构
		for _, k := range keys {
			kv_rlt := &EnvTreeNodeParamKVResponse{}
			kv_rlt.Id = k.Id.Hex()

			//根据规则决定是否允许用户看到明文
			kv_rlt.Value = "********"
			if !k.Mask {
				//如果不是敏感数据
				kv_rlt.Value = k.Default
			} else {
				//如果是敏感数据
				//则根据用户是否是管理员，来决定返回内容是否是8个星号
				user, err := getCurrentUser(ctx)
				if err != nil {
					//如果找不到当前用户，则返回星号
					kv_rlt.Value = "********"
				}
				//检查当前用户是否有权限操作该容器
				if user.RoleSet&types.ROLESET_SYSADMIN != types.ROLESET_SYSADMIN {
					//如果不是管理员
					kv_rlt.Value = "********"
				} else {
					kv_rlt.Value = k.Default
				}
			}

			kv_rlt.Name = k.Name
			kv_rlt.Description = k.Description

			results = append(results, kv_rlt)
		}

		//按照名字给results排序
		if results.Len() > 0 {
			sort.Sort(results)
		}

		//计算一共有多少页
		pc := c / req.PageSize
		if c%req.PageSize > 0 {
			pc += 1
		}
		rsp := EnvValuesListResponse{
			Total:     c,
			Page:      req.Page,
			PageCount: pc,
			PageSize:  req.PageSize,
			Data:      results,
		}

		HttpOK(w, rsp)
	})
}

type EnvValuesDetailsResponse struct {
	Id          string
	Name        string
	Value       string
	Description string
	Mask        bool
	Values      []*EnvValuesDetailsValueResponse
}

type EnvValuesDetailsValueResponse struct {
	PoolId   string
	PoolName string
	Value    string
}

//根据参数KEY的id
//查询所有该KEY的使用情况
func getTreeValueDetails(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"team", "env_tree_node_param_key", "env_tree_node_param_value", "pool"}, func(cs map[string]*mgo.Collection) {
		//得到某个KEY的id
		id := mux.Vars(r)["id"]

		key := types.EnvTreeNodeParamKey{}

		//先根据key id查找该key是否存在
		if err := cs["env_tree_node_param_key"].FindId(bson.ObjectIdHex(id)).One(&key); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "not found key", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		values := make([]*types.EnvTreeNodeParamValue, 0, 20)

		selector := bson.M{
			"key": bson.ObjectIdHex(id),
		}

		//查找该KEY所拥有的全部VALUE
		if err := cs["env_tree_node_param_value"].Find(selector).All(&values); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "not found params", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		//过滤器
		//
		//要根据当前用户有权限的pool查找该用户所有pool
		//用户所有pool中查找跟该dir对应的tree建立关系的poll
		//建立关系的pool中如果存在没有创建实际VALUE的情况
		//则使用KEY中的default代替
		user, err := getCurrentUser(ctx)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//如果是系统管理员，选择器为空
		//可以看到任何pools
		if user.RoleSet&types.ROLESET_SYSADMIN == types.ROLESET_SYSADMIN {
			selector = bson.M{}
		} else {
			//如果不是系统管理员
			//则找到该用户能查看的所有pool id准备查询
			poolIds := make([]bson.ObjectId, 0, 10)
			//如果该用户加入过某些团队
			//则该团队能查看的pool
			//该用户也可以查看
			if len(user.TeamIds) > 0 {
				teams := make([]types.Team, 0, 10)
				selector = bson.M{
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

				for _, team := range teams {
					poolIds = append(poolIds, team.PoolIds...)
				}
			}
			//将授权给用户的pool id也加入查询条件
			poolIds = append(poolIds, user.PoolIds...)

			selector = bson.M{"_id": bson.M{
				"$in": poolIds,
			}}
		}

		pools := make([]*types.PoolInfo, 0, 20)

		//批量查找出Pool数据
		if err := cs["pool"].Find(selector).All(&pools); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "not found params", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		//做一个POOL的ID对应VALUE实例的关系模型
		//用于每个POOL根据POOL ID查询VALUE信息
		//避免查询数据库
		var m_pid = make(map[string]*types.EnvTreeNodeParamValue)

		//整理
		for _, v := range values {
			m_pid[v.Pool.Hex()] = v
		}

		results := make([]*EnvValuesDetailsValueResponse, 0, 20)

		//整理成每个KEY对应的每个集群信息
		for _, pool := range pools {
			value, ok := m_pid[pool.Id.Hex()]
			var result *EnvValuesDetailsValueResponse
			//如果找的到对应关系
			//说明这个VALUE跟某个具体的POOL是绑定的
			//该POOL使用了这个VALUE的值
			if ok {
				//返回每个集群的当前值
				result = &EnvValuesDetailsValueResponse{
					PoolId:   pool.Id.Hex(),
					PoolName: pool.Name,
					Value:    value.Value,
				}
			} else {
				//说明在此KEY下
				//这个POOL并没有VALUE实例
				//那么该POOL将使用KEY的默认值
				result = &EnvValuesDetailsValueResponse{
					PoolId:   pool.Id.Hex(),
					PoolName: pool.Name,
					Value:    key.Default,
				}
			}

			//根据规则决定是否允许用户看到明文
			if key.Mask {
				//如果是敏感数据
				//则根据用户是否是管理员，来决定返回内容是否是8个星号
				user, err := getCurrentUser(ctx)
				if err != nil {
					//如果找不到当前用户，则返回星号
					result.Value = "********"
				}
				//检查当前用户是否有权限操作该容器
				if user.RoleSet&types.ROLESET_SYSADMIN != types.ROLESET_SYSADMIN {
					//如果不是管理员
					result.Value = "********"
				}
			}

			results = append(results, result)
		}

		rlt := EnvValuesDetailsResponse{
			Id:          id,
			Name:        key.Name,
			Mask:        key.Mask,
			Value:       key.Default,
			Description: key.Description,
			Values:      results,
		}

		//根据规则决定是否允许用户看到明文
		if key.Mask {
			//如果是敏感数据
			//则根据用户是否是管理员，来决定返回内容是否是8个星号
			user, err := getCurrentUser(ctx)
			if err != nil {
				//如果找不到当前用户，则返回星号
				rlt.Value = "********"
			}
			//检查当前用户是否有权限操作该容器
			if user.RoleSet&types.ROLESET_SYSADMIN != types.ROLESET_SYSADMIN {
				//如果不是管理员
				rlt.Value = "********"
			}
		}

		HttpOK(w, rlt)
	})
}

//创建一个参数名称
//参数名称在一个树中唯一
//要对Mongo建唯一性索引
//要处理唯一性索引的错误，提示用户不该输入冲突的Name
func createValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvTreeNodeParamKVRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	//校验入参
	if len(req.TreeId) <= 0 || len(req.DirId) <= 0 {
		HttpError(w, "Params error!", http.StatusBadRequest)
		return
	}

	//校验入参
	if len(req.Name) <= 0 || len(req.Value) <= 0 {
		HttpError(w, "Params error!", http.StatusBadRequest)
		return
	}

	tables := []string{"env_tree_meta", "env_tree_node_dir", "env_tree_node_param_key", "env_tree_node_param_value"}
	utils.GetMgoCollections(ctx, w, tables, func(cs map[string]*mgo.Collection) {
		tree := &types.EnvTreeMeta{}
		dir := &types.EnvTreeNodeDir{}
		key := &types.EnvTreeNodeParamKey{}

		//检查tree id是否存在记录
		if err := cs["env_tree_meta"].FindId(bson.ObjectIdHex(req.TreeId)).One(&tree); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}
		//检查dir id是否存在记录
		if err := cs["env_tree_node_dir"].FindId(bson.ObjectIdHex(req.DirId)).One(&dir); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree dir", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		//根据树ID dir id 以及名字来确定唯一的KEY
		query := bson.M{
			"dir":  dir.Id,
			"tree": tree.Id,
			"name": req.Name,
			//"Description": req.Description,
		}
		//查找KEY是否存在
		//如果不存在则创建一条新的参数KEY
		c, err := cs["env_tree_node_param_key"].Find(query).Count()
		if err != nil {
			if err == mgo.ErrNotFound {
				goto CREATEENVKEY
			} else {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		//c>0
		//说明有冲突，该数下面有同名的KEY存在
		if c > 0 {
			HttpError(w, "同一颗参数目录树下面，不可以有同名的节点存在。", http.StatusInternalServerError)
			return
		}

	CREATEENVKEY:

		//创建KEY实例
		key = &types.EnvTreeNodeParamKey{
			Id:          bson.NewObjectId(),
			Name:        req.Name,
			Mask:        req.Mask,
			Default:     req.Value,
			Dir:         dir.Id,
			Tree:        tree.Id,
			Description: req.Description,
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
		}
		if err := cs["env_tree_node_param_key"].Insert(key); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := bson.M{"keys": key.Id}
		selector := bson.M{"_id": dir.Id}

		//将此KEY插入dir目录节点的参数数组中
		//$addToSet会将array当做set使用
		//去重keys，避免重复插入相同数据
		if err := cs["env_tree_node_dir"].Update(selector, bson.M{"$addToSet": data}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//KEY创建成功

		//封装成客户端的返回对象
		kv := EnvTreeNodeParamKVResponse{
			Id:          key.Id.Hex(),
			Name:        key.Name,
			Value:       key.Default,
			Description: key.Description,
			DirId:       dir.Id.Hex(),
			TreeId:      tree.Id.Hex(),
		}

		HttpOK(w, kv)
	})
}

//更新一个参数名称
func updateValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvTreeNodeParamKVRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := mux.Vars(r)["id"]
	//校验入参
	if len(id) <= 0 {
		HttpError(w, "请提供参数ID.", http.StatusBadRequest)
		return
	}
	req.Id = id

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key"}, func(cs map[string]*mgo.Collection) {
		data := bson.M{}

		/*
			系统审计
		*/
		oldEnvKey := &types.EnvTreeNodeParamKey{}

		if err := cs["env_tree_node_param_key"].FindId(bson.ObjectIdHex(id)).One(oldEnvKey); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//如果需要更新KEY
		if req.Name != "" {
			data["name"] = req.Name
		}

		//如果需要更新VALUE
		if req.Value != "" {
			data["default"] = req.Value
		}

		//如果需要更新Description
		if req.Description != "" {
			data["description"] = req.Description
		}

		if req.Mask {
			data["mask"] = req.Mask
		}

		data["updatedtime"] = time.Now().Unix()

		selector := bson.M{"_id": bson.ObjectIdHex(id)}

		//根据VALUE实例的KEY ID更新KEY
		if err := cs["env_tree_node_param_key"].Update(selector, bson.M{"$set": data}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, req)

		/*
			系统审计
		*/
		newEnvKey := &types.EnvTreeNodeParamKey{}

		if err := cs["env_tree_node_param_key"].FindId(bson.ObjectIdHex(id)).One(newEnvKey); err != nil {
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

		logData := map[string]interface{}{
			"OldEnvValue": oldEnvKey,
			"NewEnvValue": newEnvKey,
		}
		utils.CreateSystemAuditLogWithCtx(ctx, r, types.SystemAuditModuleTypeEnv, types.SystemAuditModuleOperationTypeUpdateEnvValue, "", "", logData)
	})
}

//删除一个参数名称
//必须同步删除各集群的当前参数值
//其实是删除KEY
func deleteValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key", "env_tree_node_param_value", "env_tree_node_dir"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]

		//删除KEY的所有VALUE
		selector := bson.M{
			"key": bson.ObjectIdHex(id),
		}

		if info, err := cs["env_tree_node_param_value"].RemoveAll(selector); err != nil {
			log.Info("Rmove values with the key: %s, result: %#v", id, info)
		}

		//删除该KEY
		if err := cs["env_tree_node_param_key"].RemoveId(bson.ObjectIdHex(id)); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a key", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//从DIR节点中清理掉自己
		data := bson.M{
			"$pull": bson.M{
				"keys": bson.ObjectIdHex(id),
			},
		}

		//根据KEY的ID，查找DIR的keys中含有该ID的DIR记录
		selector = bson.M{
			"keys": bson.M{
				"$in": []bson.ObjectId{bson.ObjectIdHex(id)},
			},
		}
		//从DIR节点的key数组中
		//删除自己的记录，避免污染
		if err := cs["env_tree_node_dir"].Update(selector, data); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree dir", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

type EnvValuesUpdateValues struct {
	PoolId string
	Value  string
}

//批量更新某个VALUE
//该方法，批量的将POOL和VALUE建立联系
func updateValueAttributes(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	reqs := make([]*EnvValuesUpdateValues, 10)

	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(reqs) <= 0 {
		HttpError(w, "Need valid request", http.StatusBadRequest)
		return
	}

	//KEY的ID
	id := mux.Vars(r)["id"]

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_value", "env_tree_node_param_key", "pool"}, func(cs map[string]*mgo.Collection) {
		bulk := cs["env_tree_node_param_value"].Bulk()

		/*
			系统审计
		*/
		auditData := make([]*types.SystemAuditModuleEnvUpdatePoolValueItem, 0, 20)

		//找到key
		key := &types.EnvTreeNodeParamKey{}

		if err := cs["env_tree_node_param_key"].FindId(bson.ObjectIdHex(id)).One(key); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, req := range reqs {
			//如果这个请求参数结构中
			//value为空，则没有必要更新
			if len(req.Value) <= 0 {
				continue
			}
			//根据id以及pool找到VALUE实例
			//其实只需要id即可
			selector := bson.M{
				"key":  bson.ObjectIdHex(id),
				"pool": bson.ObjectIdHex(req.PoolId),
			}

			/*
				系统审计
			*/

			//找到pool
			pool := &types.PoolInfo{}

			if err := cs["pool"].FindId(bson.ObjectIdHex(req.PoolId)).One(pool); err != nil {
				logrus.Errorf(err.Error())
			}

			//找到老版本的value
			v := &types.EnvTreeNodeParamValue{}
			if err := cs["env_tree_node_param_value"].Find(selector).One(v); err != nil {
				logrus.Errorf(err.Error())
			}

			auditItem := &types.SystemAuditModuleEnvUpdatePoolValueItem{
				EnvValue: map[string]string{
					"Id":    key.Id.Hex(),
					"Name":  key.Name,
					"Value": key.Default,
				},
				Pool: map[string]string{
					"Id":   pool.Id.Hex(),
					"Name": pool.Name,
				},
				ValueId:  v.Id,
				OldValue: v,
			}

			auditData = append(auditData, auditItem)

			//更新目标实例的value值
			data := bson.M{
				"value": req.Value,
			}

			//https://docs.mongodb.com/manual/reference/method/Bulk.find.update/#Bulk.find.update
			bulk.Upsert(selector, bson.M{"$set": data})
		}

		if rlts, err := bulk.Run(); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			log.Info("bulk upsert results: %#v", *rlts)
		}

		HttpOK(w, nil)

		/*
			系统审计
		*/

		//整理数据
		//找出每条Value的最新值
		for _, item := range auditData {
			if item.ValueId != "" {
				vId := item.ValueId
				v := &types.EnvTreeNodeParamValue{}
				//找到Value的最新值
				//存放到系统审计日志中
				if err := cs["env_tree_node_param_value"].FindId(vId).One(v); err != nil {
					logrus.Errorln(err.Error())
				} else {
					item.NewValue = v
					//每条变更都要单独入库
					utils.CreateSystemAuditLogWithCtx(ctx, r, types.SystemAuditModuleTypeEnv, types.SystemAuditModuleOperationTypeUpdateEnvValue, item.Pool["Id"], "", item)
				}
			}

		}
	})
}

//根据PoolId和KeyId获取Value
func getValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var poolId string
	var keyId string

	//检查参数合法性
	if poolId = r.FormValue("PoolId"); len(poolId) <= 0 {
		HttpError(w, "PoolId is empty", http.StatusBadRequest)
		return
	}
	if keyId = r.FormValue("KeyId"); len(keyId) <= 0 {
		HttpError(w, "KeyId is empty", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_value", "env_tree_node_param_key"}, func(cs map[string]*mgo.Collection) {
		rsp, err := GetValueHelper(cs, poolId, keyId)

		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, rsp)
	})

}

type EnvGetEnvKeyNameWithPrefixResponse struct {
	Total     int
	PageCount int
	PageSize  int
	Page      int
	Data      []types.EnvTreeNodeParamKey
}

/*
	前缀匹配，查询当前用户有权限访问的集群关联的Tree下面的参数名称（KEY）
*/
func getEnvKeyNameWithPrefix(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var prefix = r.Form.Get("Keyword")

	//允许prefix为空的情况
	//if prefix == "" {
	//	HttpError(w, "入参Keyword不可为空", http.StatusBadRequest)
	//	return
	//}

	var pageSize int
	s_pageSize := r.Form.Get("PageSize")
	if s_pageSize == "" {
		pageSize = 20
	} else {
		s, err := strconv.Atoi(s_pageSize)
		if err != nil {
			HttpError(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			pageSize = s
		}
	}

	var page int
	s_page := r.Form.Get("Page")
	if s_page == "" {
		page = 0
	} else {
		page, err := strconv.Atoi(s_page)
		if err != nil {
			HttpError(w, err.Error(), http.StatusBadRequest)
			return
		}
		page -= 1
	}
	if page < 0 {
		page = 0
	}

	user, err := getCurrentUser(ctx)

	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"pool", "team", "env_tree_meta", "env_tree_node_param_key"}, func(cs map[string]*mgo.Collection) {
		eids := make([]bson.ObjectId, 0, 20) // EnvTreeMeta的ID

		var treeId = r.Form.Get("TreeId")
		if treeId != "" {
			//如果调用方提供了某个参数目录树的ID
			//则只查找该目录树的参数名称
			//缩小了查找范围
			//不需要考虑权限问题
			eids = append(eids, bson.ObjectIdHex(treeId))
		} else {
			pids, err := utils.PoolIdsOfUser(cs["pool"], cs["team"], user)
			if err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//找到当前用户可访问的集群
			pools := make([]types.PoolInfo, 0, 20)

			if err := cs["pool"].Find(bson.M{"_id": bson.M{"$in": pids}}).All(&pools); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, pool := range pools {
				eids = append(eids, bson.ObjectIdHex(pool.EnvTreeId))
			}
		}

		keys := make([]types.EnvTreeNodeParamKey, 0, 20)

		selector := bson.M{
			"tree": bson.M{
				"$in": eids,
			},
		}
		if prefix != "" {
			selector["name"] = bson.M{
				"$regex": bson.RegEx{
					Pattern: fmt.Sprintf("^%s", prefix),
					Options: "i",
				},
			}
		}

		//按照name降序输出参数名称模型结果
		if err := cs["env_tree_node_param_key"].Find(selector).Sort("name").Skip(page * pageSize).Limit(pageSize).All(&keys); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var total int
		if t, err := cs["env_tree_node_param_key"].Find(selector).Count(); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			total = t
		}

		//返回分页后，根据前缀查询的KEYS

		var pageCount int
		if total%pageSize == 0 {
			pageCount = total / pageSize
		} else {
			pageCount = total/pageSize + 1
		}

		rsp := EnvGetEnvKeyNameWithPrefixResponse{
			Total:     total,
			PageSize:  len(keys),
			PageCount: pageCount,
			Page:      page + 1,
			Data:      keys,
		}

		HttpOK(w, rsp)
	})

}

func GetEnvValueByName(ctx context.Context, envTtreeId string, poolId string, keyName string) (*EnvValuesDetailsValueResponse, error) {

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		return nil, err
	}
	defer mgoSession.Close()

	config := utils.GetAPIServerConfig(ctx)
	return GetEnvValueByNameHelper(map[string]*mgo.Collection{
		"env_tree_node_param_key":   mgoSession.DB(config.MgoDB).C("env_tree_node_param_key"),
		"env_tree_node_param_value": mgoSession.DB(config.MgoDB).C("env_tree_node_param_value"),
	}, envTtreeId, poolId, keyName)
}

//根据TreeId以及参数名称keyName
//查找参数值Value
//找不到则提供Key的默认值
func GetEnvValueByNameHelper(cs map[string]*mgo.Collection, treeId string, poolId string, keyName string) (*EnvValuesDetailsValueResponse, error) {
	rsp := &EnvValuesDetailsValueResponse{}
	value := types.EnvTreeNodeParamValue{}
	key := types.EnvTreeNodeParamKey{}

	var selector bson.M

	//根据树ID和名字查找KEY
	//一个数下面，名字是唯一的
	selector = bson.M{
		"name": keyName,
		"tree": bson.ObjectIdHex(treeId),
	}
	//先查找是否有该Name的参数名称
	if err := cs["env_tree_node_param_key"].Find(selector).One(&key); err != nil {
		//如果找不到肯定业务出错了
		if err == mgo.ErrNotFound {
			return nil, errors.New(fmt.Sprintf("no such key for name: %s", keyName))
		}
		return nil, err
	}

	//根据
	selector = bson.M{
		"pool": bson.ObjectIdHex(poolId),
		"key":  key.Id,
	}

	var err error
	if err = cs["env_tree_node_param_value"].Find(selector).One(&value); err != nil {
		//如果服务端发生错误则退出
		//除非是找不到该VALUE
		//说明这个Pool没有对这个KEY设置自己的VALUE，要使用KEY的DEFAULT
		if err != mgo.ErrNotFound {
			return nil, err
		}
	}

	rsp.PoolId = value.Pool.Hex()
	//如果找不到VALUE
	//则需要使用KEY的默认值
	if err == mgo.ErrNotFound {
		rsp.Value = key.Default
	} else {
		rsp.Value = value.Value
	}

	return rsp, nil

}

func GetValueHelper(cs map[string]*mgo.Collection, poolId string, keyId string) (*EnvValuesDetailsValueResponse, error) {
	rsp := &EnvValuesDetailsValueResponse{}
	value := types.EnvTreeNodeParamValue{}

	selector := bson.M{
		"pool": bson.ObjectIdHex(poolId),
		"key":  bson.ObjectIdHex(keyId),
	}

	var err error
	if err = cs["env_tree_node_param_value"].Find(selector).One(&value); err != nil {
		//如果服务端发生错误则退出
		//除非是找不到该VALUE
		//说明要使用KEY的DEFAULT
		if err != mgo.ErrNotFound {
			return nil, err
		}
	}

	rsp.PoolId = value.Pool.Hex()
	//如果找不到VALUE
	//则需要使用KEY的默认值
	if err == mgo.ErrNotFound {
		key := types.EnvTreeNodeParamKey{}
		if err = cs["env_tree_node_param_key"].FindId(bson.ObjectIdHex(keyId)).One(&key); err != nil {
			if err == mgo.ErrNotFound {
				return nil, errors.New("no such key for id: %s")
			}
			return nil, err
		}

		rsp.Value = key.Default
	} else {
		rsp.Value = value.Value
	}

	return rsp, nil
}

/*
	辅助方法
*/

//用于/envs/values/list结果中Data数组，按照Name排序
type EnvTreeNodeDirsResponseSlice []*EnvTreeNodeDirsResponse

func (c EnvTreeNodeDirsResponseSlice) Len() int {
	return len(c)
}
func (c EnvTreeNodeDirsResponseSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c EnvTreeNodeDirsResponseSlice) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

//构建树结构
func TreeBuild(rsp *EnvTreeNodeDirsResponse, results []types.EnvTreeNodeDir) error {
	//根节点指针
	var root *types.EnvTreeNodeDir
	for _, r := range results {
		if r.Parent == "" {
			root = &r
			break
		}
	}

	if root != nil {
		//从根节点开始找起
		//将树状结构从数组中梳理出来
		rsp.Id = root.Id
		rsp.Name = root.Name
		//rsp.Children = make([]*EnvTreeNodeDirsResponse, len(root.Children))
		rsp.ParentId = ""
		rsp.CreatedTime = root.CreatedTime
		rsp.UpdatedTime = root.UpdatedTime
		TreeNodeBuild(root, rsp, results)
	} else {
		return errors.New("Could not found root node!")
	}
	return nil
}

//构建树结构所需的节点
//根据root节点构建余下的子节点
func TreeNodeBuild(node *types.EnvTreeNodeDir, node_rsp *EnvTreeNodeDirsResponse, results []types.EnvTreeNodeDir) {
	//使其初始化为数组
	node_rsp.Children = make([]*EnvTreeNodeDirsResponse, 0, 20)

	for _, child := range node.Children {
		for _, sub_node := range results {
			if child == sub_node.Id {
				sub_node_rsp := &EnvTreeNodeDirsResponse{
					Id:       sub_node.Id,
					Name:     sub_node.Name,
					ParentId: node_rsp.Id.Hex(),
					//Children:    make([]*EnvTreeNodeDirsResponse, len(sub_node.Children)),
					CreatedTime: sub_node.CreatedTime,
					UpdatedTime: sub_node.UpdatedTime,
				}
				node_rsp.Children = append(node_rsp.Children, sub_node_rsp)
				TreeNodeBuild(&sub_node, sub_node_rsp, results)
			}
		}
		//给子节点按照Name排序
		if node_rsp.Children.Len() > 0 {
			sort.Sort(node_rsp.Children)
		}
	}
}
