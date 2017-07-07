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
	utils.GetMgoCollections(ctx, w, []string{"env_tree_meta"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]

		if err := cs["env_tree_meta"].Remove(bson.M{"_id": bson.ObjectIdHex(id)}); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
		data := bson.M{
			"$pull": bson.M{
				"children": id,
			},
		}

		//从父级节点的children数组中
		//删除自己的记录，避免污染父节点
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

func getTreeValues(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := EnvValuesListRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key", "env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		var keys []types.EnvTreeNodeParamKey
		var values []types.EnvTreeNodeParamValue
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
			data["name"] = bson.M{
				"$regex": bson.RegEx{
					Pattern: fmt.Sprintf("/%s/", req.Name),
					Options: "",
				},
			}
		}

		//找到参数目录树中的全部匹配的KEY
		if err := cs["env_tree_node_param_key"].Find(data).All(&keys); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//如果找不到匹配的KEY
		//则没必要再次查找KEY对应的VALUE，直接返回空内容即可
		if len(keys) == 0 {
			HttpOK(w, []EnvTreeNodeParamKVResponse{})
			return
		}

		//找到KEY的Id数组，用于批量查询
		v_ids := make([]bson.ObjectId, len(keys))
		//构造KEY的ID和实例对应的MAP
		//用于VALUES匹配查询
		k_map := make(map[string]types.EnvTreeNodeParamKey)

		for _, key := range keys {
			v_ids = append(v_ids, key.Id)
			k_map[key.Id.Hex()] = key
		}

		data = bson.M{
			"tree": bson.ObjectIdHex(req.TreeId),
			"key": bson.M{
				"$in": v_ids,
			},
		}

		//查询所有条件匹配参数值的总数
		//不是参数KEY，是参数值VALUE
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
		//根据KEY和TREE查找对应的VALUE
		if err := cs["env_tree_node_param_value"].Find(data).Skip(req.Page * req.PageSize).Limit(req.PageSize).All(&values); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//整理成客户端需要的数据结构
		results = make([]EnvTreeNodeParamKVResponse, len(values))
		for _, value := range values {
			kv_rlt := EnvTreeNodeParamKVResponse{}
			kv_rlt.Id = value.Id.Hex()
			kv_rlt.Value = value.Value
			//从对应的KEY实例中找到Name
			if k, ok := k_map[value.Id.Hex()]; ok {
				kv_rlt.Name = k.Name
			}
			kv_rlt.Description = value.Description

			results = append(results, kv_rlt)
		}

		//计算一共有多少页
		pc := c / req.PageSize
		if c%req.PageSize > 0 {
			pc += 1
		}
		rsp := map[string]interface{}{
			"Total":     c,
			"PageCount": pc,
			"PageSize":  req.PageSize,
			"Page":      req.Page,
			"Data":      results,
		}

		HttpOK(w, rsp)
	})
}

func getTreeValueDetails(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]
		

		cs["env_tree_node_param_value"].FindId(bson.ObjectIdHex(id))
	})
}

//创建一个参数值
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

		query := bson.M{
			"Dir":  dir.Id,
			"Tree": tree.Id,
			"Name": req.Name,
		}
		//查找KEY是否存在
		//如果不存在则创建一条新的参数KEY
		err := cs["env_tree_node_param_key"].Find(query).One(&key)
		if err == mgo.ErrNotFound {
			//创建KEY实例
			key = &types.EnvTreeNodeParamKey{
				Id:          bson.NewObjectId(),
				Name:        req.Name,
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

		//创建KEY的VALUE
		value := &types.EnvTreeNodeParamValue{
			Id:          bson.NewObjectId(),
			Value:       req.Value,
			Description: req.Description,
			Key:         key.Id,
			Tree:        tree.Id,
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
		}
		if err := cs["env_tree_node_param_value"].Insert(value); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//封装成客户端的返回对象
		kv := EnvTreeNodeParamKVResponse{
			Id:          value.Id.Hex(),
			Name:        key.Name,
			Value:       value.Value,
			Description: value.Description,
			DirId:       dir.Id.Hex(),
			TreeId:      tree.Id.Hex(),
		}

		HttpOK(w, kv)
	})
}

//更新某个VALUE的值
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

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_key", "env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		data := bson.M{}

		//因为Name是保存在KEY的结构中
		//如果需要更新KEY
		if req.Name != "" {
			//根据VALUE的ID找到VALUE实例
			v := &types.EnvTreeNodeParamValue{}
			if err := cs["env_tree_node_param_value"].FindId(bson.ObjectIdHex(id)).One(&v); err != nil {
				HttpError(w, "param id is invalide", http.StatusNotFound)
				return
			}
			data := bson.M{
				"name":        req.Name,
				"updatedtime": time.Now().Unix(),
			}
			selector := bson.M{"_id": v.Key}

			//根据VALUE实例的KEY ID更新KEY
			if err := cs["env_tree_node_param_key"].Update(selector, bson.M{"$set": data}); err != nil {
				if err == mgo.ErrNotFound {
					HttpError(w, err.Error(), http.StatusNotFound)
					return
				}

				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		//如果这两个参数都为空
		//则不需要更改VALUE实例
		if req.Description == "" && req.Value == "" {
			HttpOK(w, req)
			return
		}

		//否则更新两个或者其中某个属性

		if req.Description != "" {
			data["description"] = req.Description
		}
		if req.Value != "" {
			data["value"] = req.Value
		}
		data["updatedTime"] = time.Now().Unix()

		selector := bson.M{"_id": bson.ObjectIdHex(id)}

		if err := cs["env_tree_node_param_value"].Update(selector, bson.M{"$set": data}); err != nil {
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

//删除参数必须同步各集群的当前参数值
//其实是删除KEY
func deleteValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		id := mux.Vars(r)["id"]

		//通过VALUE找到KEY的id，以用于批量删除VALUE
		v := &types.EnvTreeNodeParamValue{}
		if err := cs["env_tree_node_param_value"].FindId(bson.ObjectIdHex(id)).One(&v); err != nil {
			HttpError(w, "param id is invalide", http.StatusNotFound)
			return
		}

		//删除条件是KEY的ID以及TREE的ID
		selector := bson.M{
			"key":  v.Key,
			"tree": v.Tree,
		}

		//删除所有同一个TREE中，相同KEY的VALUE
		if err := cs["env_tree_node_param_value"].Remove(selector); err != nil {
			if err == mgo.ErrNotFound {
				HttpError(w, "no such a tree", http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, nil)
	})
}

type EnvValuesUpdateValues struct {
	Id     string
	PoolId string
	Value  string
}

type EnvUpdateValueAttributesResponse struct {
	Id     string
	PoolId string
	Value  string
	Status int
}

//批量更新某个VALUE
func updateValueAttributes(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	reqs := make([]*EnvValuesUpdateValues, 10)

	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	//返回数据
	//因为无法事物的方式批量更新
	//所以返回结果里面要告知调用者，哪些插入失败，失败原因如何
	var rsps []*EnvUpdateValueAttributesResponse

	if len(reqs) <= 0 {
		HttpError(w, "Need valid request", http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"env_tree_node_param_value"}, func(cs map[string]*mgo.Collection) {
		for _, req := range reqs {
			//如果这个请求参数结构中
			//value为空，则没有必要更新
			if len(req.Value) <= 0 {
				continue
			}
			//根据id以及pool找到VALUE实例
			//其实只需要id即可
			selector := bson.M{
				"_id":  req.Id,
				"pool": req.PoolId,
			}
			//更新目标实例的value值
			data := bson.M{
				"value": req.Value,
			}

			rsp := &EnvUpdateValueAttributesResponse{
				Id:     req.Id,
				PoolId: req.PoolId,
				Value:  req.Value,
			}

			//如果某次插入失败，因为不是事物，导致部分成功部分失败怎么办？
			//告知前端每一个插入的结果，是否重试由前端处理
			if err := cs["env_tree_node_param_value"].Update(selector, bson.M{"$set": data}); err != nil {
				if err == mgo.ErrNotFound {
					rsp.Status = http.StatusNotFound
				}
				rsp.Status = http.StatusInternalServerError
			} else {
				rsp.Status = http.StatusOK
			}
			rsps = append(rsps, rsp)
		}

		HttpOK(w, rsps)
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
