package handlers

import (
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"context"
	"net/http"
	"github.com/zanecloud/apiserver/utils"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/types"
	"encoding/json"
	"fmt"
)

func getTeamsJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {


	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()
	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("team")

	var results []types.Team
	if err := c.Find(bson.M{}).All(&results); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}


func getTeamJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	name := mux.Vars(r)["name"]

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()
	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("team")

	result := types.Team{}
	if err := c.Find(bson.M{"name": name}).One(&result); err != nil {

		if err == mgo.ErrNotFound {
			// 对错误类型进行区分，有可能只是没有这个team，不应该用500错误
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func postTeamAppoint(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}
func postTeamRemove(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}


type TeamsCreateRequest struct {
	types.Team
}

func postTeamsCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		name = r.Form.Get("name")
	)

	req := TeamsCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if name != "" {
		req.Name = name
	}

	if req.Name == "" {
		HttpError(w, "The team's name cant be empty", http.StatusBadRequest)
		return
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()
	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("team")

	n, err := c.Find(bson.M{"Name": req.Name}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n != 0 {
		HttpError(w, "the team's name is dup", http.StatusConflict)
		return
	}

	team := &types.Team{
		Name:        req.Name,
		Id:          bson.NewObjectId(),
		Description: req.Description,
		Leader: types.Leader{
			Id:   req.Leader.Id,
			Name: req.Leader.Name,
		},
	}
	if err := c.Insert(team); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{%q:%q}", "Id", team.Id.Hex())
}

type TeamJoinRequest struct {
}

// 一批用户加入某个team
func postTeamJoin(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	//不用teamId，
	//name := mux.Vars("name")

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

}
