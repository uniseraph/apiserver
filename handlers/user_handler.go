package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	 "github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

func getUserLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		pass = r.Form.Get("pass")
		name = mux.Vars(r)["name"]
	)

	if pass == "" {
		HttpError(w, "pass can't be empty", http.StatusBadRequest)
		return
	}

	mgoSession, err := utils.GetMgoSession(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	logrus.Debugf("getUserLogoin::name is %s , pass is %s", name, pass)
	result := types.User{}
	if err := mgoSession.DB(mgoDB).C("user").Find(bson.M{"name": name}).One(&result); err != nil {
		HttpError(w, err.Error(), http.StatusNotFound)
		return
	}

	logrus.Debugf("getUserLogin::get the user %#v", result)
	if result.Pass != pass {
		HttpError(w, "pass is error", http.StatusUnauthorized)
		return
	}

	client, err := utils.GetRedisClient(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := client.Set(utils.KEY_REDIS_UID, result.Id.String(), time.Minute*10).Err(); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uid_cookie := &http.Cookie{
		Name:     "uid",
		Value:    result.Id.String(),
		Path:     "/",
		HttpOnly: false,
		MaxAge:   600,
	}
	http.SetCookie(w, uid_cookie)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

type UsersCreateRequest struct {
	types.User
}

func postUsersCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		name = r.Form.Get("Name")
		pass = r.Form.Get("pass")
	)

	req := UsersCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if name != "" {
		req.Name = name
	}

	if pass != "" {
		req.Pass = pass
	}

	if req.Name == "" || req.Pass == "" {
		HttpError(w, "name and pass cant be empty", http.StatusBadRequest)
		return
	}

	mgoSession, err := utils.GetMgoSession(ctx)
	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("user")

	// mongodb需要在user.name有唯一性索引
	n, err := c.Find(bson.M{"name": req.Name}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n != 0 {
		HttpError(w, "the user name is conflict", http.StatusConflict)
		return
	}

	//创建用户时候，可以分配角色
	if err := c.Insert(&types.User{Name: req.Name,
		Id:                              bson.NewObjectId(),
		Pass:                            req.Pass,
		Email:                           req.Email,
		Comments:                        req.Comments,
		RoleSet:                         req.RoleSet,
		CreatedTime:  					 time.Now().Unix(),
	}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Name", req.Name)
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
	json.NewEncoder(w).Encode(result)

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
		HttpError(w, err.Error(), http.StatusConflict)
		return
	}

	if err := c.Insert(&types.Team{
		Name:        req.Name,
		Id:          bson.NewObjectId(),
		Description: req.Description,
		Leader : types.Leader {
			Id : req.Leader.Id,
			Name: req.Leader.Name,
		} ,
	}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Name", req.Name)
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

func postUserRoleSet(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func postTeamAppoint(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}
func postTeamRemove(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func getTeamsJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func getUserInspect(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func getUsersJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}


func postUserRemove(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}


