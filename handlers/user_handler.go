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

func getUserCurrent(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("uid")
	if err != nil {
		HttpError(w, "please login", http.StatusForbidden)
		return
	}

	uid := cookie.Value

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("user")

	result := types.User{}

	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(uid)}).One(&result); err != nil {

		if err == mgo.ErrNotFound {
			HttpError(w, fmt.Sprintf("no such a user id is %s", uid), http.StatusNotFound)
			return
		}

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//需要过滤
	result.Pass = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func getUserLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		pass = r.Form.Get("Pass")
		name = mux.Vars(r)["name"]
	)

	if pass == "" {
		HttpError(w, "pass can't be empty", http.StatusBadRequest)
		return
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	logrus.Debugf("getUserLogoin::name is %s ", name)
	result := types.User{}
	if err := mgoSession.DB(mgoDB).C("user").Find(bson.M{"name": name}).One(&result); err != nil {
		HttpError(w, err.Error(), http.StatusNotFound)
		return
	}

	logrus.Debugf("getUserLogin::get the user %#v", result)
	if result.Pass != utils.Md5(pass) {
		HttpError(w, "pass is error", http.StatusForbidden)
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
		Value:    result.Id.Hex(),
		Path:     "/",
		HttpOnly: false,
		MaxAge:   600,
	}

	logrus.Debugf("getUserLogin::get the cookie %#v", uid_cookie)

	http.SetCookie(w, uid_cookie)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ok")
}

type UsersCreateRequest struct {
	types.User
}

type UsersCreateResponse struct {
	Id string
}

func postUsersCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		name = r.Form.Get("Name")
		pass = r.Form.Get("Pass")
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

	logrus.Debugf("the new user is %#v", req)

	if req.Name == "" || req.Pass == "" {
		HttpError(w, "name or pass cant be empty", http.StatusBadRequest)
		return
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

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
	user := &types.User{Name: req.Name,
		Id:          bson.NewObjectId(),
		Pass:        utils.Md5(req.Pass),
		Email:       req.Email,
		Comments:    req.Comments,
		RoleSet:     req.RoleSet,
		CreatedTime: time.Now().Unix(),
		Tel:         req.Tel,
	}
	if err := c.Insert(user); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//fmt.Fprintf(w, "{%q:%q}", "Id", user.Id.Hex())

	resp := &UsersCreateResponse{Id:user.Id.Hex()}
	json.NewEncoder(w).Encode(resp)
}

type UserResetPassRequest struct {
	Id      string
	NewPass string
}

//"/users/{id:.*}/reset"
func postUserResetPass(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	req := UserResetPassRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB
	c := mgoSession.DB(mgoDB).C("user")

	if err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{"pass": utils.Md5(req.NewPass)}}); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func getUserInspect(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("user")

	result := types.User{}

	if err := c.Find(bson.M{"$or": []bson.M{bson.M{"_id": bson.ObjectIdHex(id)}}}).One(&result); err != nil {

		if err == mgo.ErrNotFound {
			HttpError(w, fmt.Sprintf("no such a user name or id is %s", id), http.StatusNotFound)
			return
		}

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//pass不要输出
	result.Pass = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func getUsersJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("user")

	var results []types.User
	if err := c.Find(bson.M{}).All(&results); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, _ := range results {
		results[i].Pass = ""
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)

}

///users/{name:.*}/remove
func postUserRemove(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("user")

	if err := c.Remove(bson.M{"_id": bson.ObjectIdHex(id)}); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, "no such a user", http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{%q:%q}", "Id", id)

}

// /users/{id:.*}/join?TeamId=xxx"
func postUserJoin(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		teamId = r.Form.Get("TeamId")
		userId = mux.Vars(r)["id"]
	)

	//需要判断当前用户是否为团队主管
	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
		return
	}

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()
	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB
	c_team := mgoSession.DB(mgoDB).C("team")
	team := &types.Team{}
	if err := c_team.FindId(bson.M{"_id": bson.ObjectIdHex(teamId)}).One(team); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, fmt.Sprintf("no such a team :%s", teamId), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if team.Leader.Id != currentUser.Id.Hex() {
		HttpError(w, fmt.Sprintf("current user:%s isn't the team:%s  leader", currentUser.Id.Hex(), teamId), http.StatusForbidden)
		return
	}

	c := mgoSession.DB(mgoDB).C("userteam")

	n, err := c.Find(bson.M{"teamid": teamId, "userid": userId}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n != 0 {
		HttpError(w, fmt.Sprintf("the user:%s has been in  the team %s", userId, teamId), http.StatusInternalServerError)
		return
	}

	if err := c.Insert(&types.TeamUser{
		UserId: userId,
		TeamId: teamId,
	}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// /users/{id:.*}/quit?TeamId=xxx
func postUserQuit(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		teamId = r.Form.Get("TeamId")
		userId = mux.Vars(r)["id"]
	)

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB
	c := mgoSession.DB(mgoDB).C("userteam")

	n, err := c.Find(bson.M{"teamid": teamId, "userid": userId}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n == 0 {
		HttpError(w, fmt.Sprintf("the user:%s hasn't  been in  the team %s", userId, teamId), http.StatusInternalServerError)
		return
	}

	if err := c.Remove(bson.M{
		"userid": userId,
		"teamId": teamId,
	}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

type UserUpdateRequest struct {
	Name     string
	Roleset  types.Roleset
	Pass     string
	Tel      string
	Email    string
	Comments string
}

func postUserUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	req := UserUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	logrus.Debugf("postUserUpdate::the request is %#v", req)

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer mgoSession.Close()

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB
	c := mgoSession.DB(mgoDB).C("user")

	selector := bson.M{"_id": bson.ObjectIdHex(id)}

	data := bson.M{}
	if req.Name != "" {
		data = bson.M{"name": req.Name}
	}

	//TODO 这里要去roleset必须传
	data["Roleset"] = req.Roleset

	if req.Pass != "" {
		data["Pass"] = utils.Md5(req.Pass)
	}

	if req.Tel != "" {
		data["Tel"] = req.Tel
	}

	if req.Email != "" {
		data["Email"] = req.Email
	}

	if req.Comments != "" {
		data["Comments"] = req.Comments
	}

	if err := c.Update(selector, bson.M{"$set": data}); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
