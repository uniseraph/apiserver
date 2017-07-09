package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gpmgo/gopm/modules/log"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
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
	Children    []*EnvTreeNodeDirsResponse
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

	utils.GetMgoCollections(ctx, w, []string{"env_tree_meta"}, func(cs map[string]*mgo.Collection) {
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

		resp := &EnvTreeMetaResponse{
			Id:          tree.Id.Hex(),
			Name:        tree.Name,
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

		var results []types.EnvTreeNodeDir

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
			//p_dir := &types.EnvTreeNodeDir{}
			//if err := cs["env_tree_node_dir"].FindId(bson.ObjectIdHex(req.ParentId)).One(&p_dir); err != nil {
			//	HttpError(w, "ParentId is invalide", http.StatusNotFound)
			//	return
			//}
			//p_dir.Children = append(p_dir.Children, dir.Id)
			////更新父节点
			//if err := cs["env_tree_node_dir"].Insert(dir); err != nil {
			//	HttpError(w, err.Error(), http.StatusInternalServerError)
			//	return
			//}
			data := bson.M{"children": dir.Id}
			selector := bson.M{"_id": bson.ObjectIdHex(req.ParentId)}

			if err := cs["env_tree_node_dir"].Update(selector, bson.M{"$push": data}); err != nil {
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
	Data      []EnvTreeNodeParamKVResponse
}

func getTreeValues(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvValuesListRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key", "env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		var keys []types.EnvTreeNodeParamKey
		var results []EnvTreeNodeParamKVResponse

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
			data["name"] = bson.RegEx{fmt.Sprintf("%s*", req.Name), ""}
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
		//找到参数目录树中的全部匹配的KEY
		if err := cs["env_tree_node_param_key"].Find(data).Skip(req.Page * req.PageSize).Limit(req.PageSize).All(&keys); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//整理成客户端需要的数据结构
		for _, k := range keys {
			kv_rlt := EnvTreeNodeParamKVResponse{}
			kv_rlt.Id = k.Id.Hex()
			kv_rlt.Value = k.Default
			kv_rlt.Name = k.Name
			kv_rlt.Description = k.Description

			results = append(results, kv_rlt)
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
	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key", "env_tree_node_param_value", "pool"}, func(cs map[string]*mgo.Collection) {
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

		var values []*types.EnvTreeNodeParamValue

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

		//TODO
		//过滤器
		selector = bson.M{}
		//要根据当前用户有权限的pool查找该用户所有pool
		//用户所有pool中查找跟该dir对应的tree建立关系的poll
		//建立关系的pool中如果存在没有创建实际VALUE的情况
		//则使用KEY中的default代替

		var pools []*types.PoolInfo

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

		var results []*EnvValuesDetailsValueResponse

		//整理成每个KEY对应的每个集群信息
		for _, pool := range pools {
			value, ok := m_pid[pool.Id.Hex()]
			//如果找的到对应关系
			//说明这个VALUE跟某个具体的POOL是绑定的
			//该POOL使用了这个VALUE的值
			if ok {
				//返回每个集群的当前值
				result := &EnvValuesDetailsValueResponse{
					PoolId:   pool.Id.Hex(),
					PoolName: pool.Name,
					Value:    value.Value,
				}

				results = append(results, result)
			} else {
				//说明在此KEY下
				//这个POOL并没有VALUE实例
				//那么该POOL将使用KEY的默认值
				result := &EnvValuesDetailsValueResponse{
					PoolId:   pool.Id.Hex(),
					PoolName: pool.Name,
					Value:    key.Default,
				}

				results = append(results, result)
			}
		}

		rlt := EnvValuesDetailsResponse{
			Id:          id,
			Name:        key.Name,
			Value:       key.Default,
			Description: key.Description,
			Values:      results,
		}
		HttpOK(w, rlt)
	})
}

//创建一个参数名称
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
		err := cs["env_tree_node_param_key"].Find(query).One(&key)
		if err == mgo.ErrNotFound {
			//创建KEY实例
			key = &types.EnvTreeNodeParamKey{
				Id:          bson.NewObjectId(),
				Name:        req.Name,
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
			if err := cs["env_tree_node_dir"].Update(selector, bson.M{"$push": data}); err != nil {
				if err == mgo.ErrNotFound {
					HttpError(w, err.Error(), http.StatusNotFound)
					return
				}

				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
		HttpError(w, "Params error!", http.StatusBadRequest)
		return
	}
	req.Id = id

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key"}, func(cs map[string]*mgo.Collection) {
		data := bson.M{}

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

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		bulk := cs["env_tree_node_param_value"].Bulk()

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
			//更新目标实例的value值
			data := bson.M{
				"value": req.Value,
			}

			//https://docs.mongodb.com/manual/reference/method/Bulk.find.update/#Bulk.find.update
			bulk.Upsert(selector, bson.M{"$set": data})
		}

		if rlts, err := bulk.Run(); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree dir", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			log.Info("bulk upsert results: %#v", *rlts)
		}

		HttpOK(w, nil)
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
		rsp := EnvValuesDetailsValueResponse{}
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
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
		}

		rsp.PoolId = value.Pool.Hex()
		//如果找不到VALUE
		//则需要使用KEY的默认值
		if err == mgo.ErrNotFound {
			key := types.EnvTreeNodeParamKey{}
			if err = cs["env_tree_node_param_key"].FindId(bson.ObjectIdHex(keyId)).One(&key); err != nil {
				if err == mgo.ErrNotFound {
					HttpError(w, fmt.Sprintf("no such key for id: %s", keyId), http.StatusNotFound)
					return
				}
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}

			rsp.Value = key.Default
		} else {
			rsp.Value = value.Value
		}

		HttpOK(w, rsp)
	})

}

/*
	辅助方法
*/

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
	}
}
