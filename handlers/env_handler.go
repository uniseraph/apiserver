package handlers

import (
	"context"
	"encoding/json"
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
/envs/dirs/list
/envs/dirs/create
/envs/dirs/:id/update
/envs/dirs/:id/remove
/envs/values/list
/envs/values/:id/detail
/envs/values/create
/envs/values/:id/update
/envs/values/:id/remove
/envs/values/:id/update-values
*/

/*
	Request
*/
type EnvTreeMetaRequest struct {
	types.EnvTreeMeta
}

type EnvTreeNodeDirRequest struct {
	types.EnvTreeNodeDir
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
	types.EnvTreeNodeDir
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

func getTreeDirs(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func createDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func updateDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func deleteDir(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

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
