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

	"github.com/Sirupsen/logrus"
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


//"/teams/{id:.*}/appoint?UserId=xxx": checkUserPermission(postTeamAppoint,types.ROLESET_SYSADMIN),
func postTeamAppoint(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		userId = r.Form.Get("UserId")
		teamId = mux.Vars(r)["id"]
	)


	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c_user := mgoSession.DB(mgoDB).C("user")

	user := &types.User{}
	if err:= c_user.FindId(bson.ObjectIdHex(userId)).One(&user); err!=nil {

		if err==mgo.ErrNotFound{
			HttpError(w, fmt.Sprintf("no such a user:%s",userId) , http.StatusNotFound)
			return
		}

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := mgoSession.DB(mgoDB).C("team")

	selector := bson.ObjectIdHex(teamId)

	data := bson.M{"leader":&types.Leader{
		Id: userId,
		Name: user.Name ,
	}}

	if err := c.Update( selector , bson.M{"$set": data}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}


	w.WriteHeader(http.StatusOK)


}

//"/teams/{id:.*}/reovke?UserId=xxx": checkUserPermission(postTeamRevoke,types.ROLESET_SYSADMIN),
func postTeamRevoke(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		userId = r.Form.Get("UserId")
		teamId = mux.Vars(r)["id"]
	)


	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c_team := mgoSession.DB(mgoDB).C("team")

	team := &types.Team{}
	if err:= c_team.FindId(bson.ObjectIdHex(teamId)).One(&team); err!=nil {

		if err==mgo.ErrNotFound{
			HttpError(w, fmt.Sprintf("no such a team:%s",teamId) , http.StatusNotFound)
			return
		}

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if team.Leader.Id != userId {
		HttpError(w,  fmt.Sprintf("the user:%s isn't the team:%s leader" , userId,teamId),http.StatusForbidden)
		return
	}


	selector := bson.ObjectIdHex(teamId)

	data := bson.M{"leader":&types.Leader{
		Id: "",
		Name: "" ,
	}}

	if err := c_team.Update( selector , bson.M{"$set": data}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}


	w.WriteHeader(http.StatusOK)
}


//"/teams/{id:.*}/remove":  checkUserPermission(postTeamRemove,types.ROLESET_SYSADMIN),
func postTeamRemove(ctx context.Context, w http.ResponseWriter, r *http.Request) {


	id := mux.Vars(r)["id"]

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("team")

	if err := c.Remove(bson.M{"_id":bson.ObjectIdHex(id)}) ; err !=nil {
		if err == mgo.ErrNotFound {
			HttpError(w,"no such a team" , http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{%q:%q}", "Id",  id)

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
		name = r.Form.Get("Name")
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


type TeamUpdateRequest struct {
	Name string
	Description string
	Leader types.Leader
}
// /teams/{id:.*}/update
func postTeamUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {


	id := mux.Vars(r)["id"]

	req := TeamUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	logrus.Debugf("postTeamUpdate::the request is %#v", req)

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB
	c := mgoSession.DB(mgoDB).C("team")


	selector:=bson.M{"_id": bson.ObjectIdHex(id)}


	data := bson.M{}
	if req.Name != "" {
		data =  bson.M{"name":req.Name}
	}

	if req.Description != "" {
		data["Description"] = req.Description
	}

	if req.Leader.Name !="" && req.Leader.Name!=""  {
		data["Leader"] = req.Leader
	}


	if err := c.Update(selector, bson.M{"$set": data }); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
