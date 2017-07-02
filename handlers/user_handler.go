package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
	"github.com/google/uuid"
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
	//校验用户输入的密码，与该ID的用户模型中Pass是否匹配
	if ok, err := utils.ValidatePassword(result, pass); ok != true || err != nil {
		HttpError(w, "pass is error", http.StatusForbidden)
		return
	}

	client, err := utils.GetRedisClient(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//生成每个用户唯一的一个session key
	//用于在缓存中保存登录状态
	sessionUUID, err := uuid.NewUUID()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessionKey := sessionUUID.String()
	//准备session的内容
	sessionContents := map[string]interface{}{
		"uid": result.Id.Hex(),
		"roleSet": strconv.FormatUint(uint64(result.RoleSet),10),  //fmt.Sprintf("%d", result.RoleSet),
	}
	err = client.HMSet(utils.RedisSessionKey(sessionKey), sessionContents).Err()
	if err != nil {
		logrus.Fatalf("Redis hmset error: %#v", err)
		panic(err)
	}
	age := time.Hour * 24 * 7
	//设置session一周超时
	//一周后再登录，会找不到redis中的key，导致认证不再可以通过，需要重新登录
	client.Expire(utils.RedisSessionKey(sessionKey), age)

	uidCookie := &http.Cookie{
		Name:     "sessionID",
		Value:    sessionKey,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   int(age),
	}

	logrus.Debugf("getUserLogin::get the cookie %#v", uidCookie)

	http.SetCookie(w, uidCookie)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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
	logrus.Debugf("User Create: name: %s, pass: %s", name, pass)

	req := UsersCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}


	//以下为遗留问题
	//为了兼容从url的参数字符串中读取参数，该参数优先于body的json
	//TODO
	//name要大于4个字符
	if name != "" {
		req.Name = name
	}

	//TODO
	//pass要大于8个字符
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

	//为用户密码加盐
	salt := utils.RandomStr(16)
	//生成加密后的密码，数据库中不保存明文密码
	encryptedPassword := utils.Md5(fmt.Sprint("%s:%s", req.Pass, salt))

	//创建用户时候，可以分配角色
	user := &types.User{Name: req.Name,
		Id:          bson.NewObjectId(),
		Pass:        encryptedPassword,
		Salt:	     salt,
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
	if err := c_team.Find(bson.M{"_id": bson.ObjectIdHex(teamId)}).One(team); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, fmt.Sprintf("no such a team :%s", teamId), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}


	c_user := mgoSession.DB(mgoDB).C("user")
	user := &types.User{}
	if err := c_user.Find(bson.M{"_id": bson.ObjectIdHex(userId)}).One(user); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, fmt.Sprintf("no such a user :%s", userId), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//如果当前用户不是改团队的主管并且当前用户不是系统管理员，则没有权限
	if team.Leader.Id != currentUser.Id.Hex()  && (currentUser.RoleSet & types.ROLESET_SYSADMIN == 0) {
		HttpError(w, fmt.Sprintf("current user:%s isn't the team:%s  leader ，and current user's roleset:%d dont include sysadmin", currentUser.Id.Hex(), teamId,currentUser.RoleSet), http.StatusForbidden)
		return
	}


	if err:= c_user.Update(bson.M{"_id":bson.ObjectIdHex(userId)} ,  bson.M{ "$addToSet" : bson.M{"teamids":bson.ObjectIdHex(teamId)}   } ); err!=nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//if err:= c_team.Update(bson.M{"_id":bson.ObjectIdHex(teamId)} ,  bson.M{ "$addToSet" : bson.M{"userids":bson.ObjectIdHex(userId)}   } ); err!=nil {
	if err:= c_team.Update(bson.M{"_id":bson.ObjectIdHex(teamId)} ,  bson.M{ "$addToSet" : bson.M{"users":user}   } ); err!=nil {

		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO没有事务保护





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
	if err := c_team.Find(bson.M{"_id": bson.ObjectIdHex(teamId)}).One(team); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, fmt.Sprintf("no such a team :%s", teamId), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}


	c_user := mgoSession.DB(mgoDB).C("user")
	user := &types.User{}
	if err := c_user.Find(bson.M{"_id": bson.ObjectIdHex(userId)}).One(user); err != nil {
		if err == mgo.ErrNotFound {
			HttpError(w, fmt.Sprintf("no such a user :%s", userId), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//如果当前用户不是改团队的主管并且当前用户不是系统管理员，则没有权限
	if team.Leader.Id != currentUser.Id.Hex()  && (currentUser.RoleSet & types.ROLESET_SYSADMIN == 0) {
		HttpError(w, fmt.Sprintf("current user:%s isn't the team:%s  leader ，and current user's roleset:%d dont include sysadmin", currentUser.Id.Hex(), teamId,currentUser.RoleSet), http.StatusForbidden)
		return
	}



	if err := c_team.UpdateId(bson.ObjectIdHex(teamId),
		bson.M{"$pull" : bson.M{ "users": bson.M{"_id": bson.ObjectIdHex(userId)}  } } ) ; err!=nil{
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := c_user.UpdateId(bson.ObjectIdHex(userId),
		bson.M{"$pull" : bson.M{"teamids": bson.ObjectIdHex(teamId)}}) ; err!=nil{
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}


	w.WriteHeader(http.StatusOK)

}


//"/users/{id:.*}/update":    &MyHandler{h: postUserUpdate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

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

	//TODO 这里要求roleset必须传
	data["roleset"] = req.Roleset

	if req.Pass != "" {
		data["pass"] = utils.Md5(req.Pass)
	}

	if req.Tel != "" {
		data["tel"] = req.Tel
	}

	if req.Email != "" {
		data["email"] = req.Email
	}

	if req.Comments != "" {
		data["comments"] = req.Comments
	}
	
	logrus.Debugf("postUserUpdate::the data is %#v", data)

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
