package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
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
			data = bson.M{"Name": req.Name}
		}

		if req.Description != "" {
			data["Description"] = req.Description
		}
		data["UpdatedTime"] = time.Now().Unix()

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
		fmt.Println("Results All: ", results)

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

func updateDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func deleteDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

/*
/envs/values/list
/envs/values/:id/detail
/envs/values/create
/envs/values/:id/update
/envs/values/:id/remove
/envs/values/:id/update-values
*/
func getTreeValues(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func getTreeValueDetails(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func createValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func updateValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func deleteValue(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func updateValueAttributes(ctx context.Context, w http.ResponseWriter, r *http.Request) {

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
