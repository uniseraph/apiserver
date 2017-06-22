package handlers

import (
	"context"
	"github.com/zanecloud/apiserver/store"
	"net/http"

	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/proxy"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func getPoolJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	mgoSession, err := utils.GetMgoSession(ctx)

	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	result := store.PoolInfo{}
	if err := c.Find(bson.M{"name": name}).One(&result); err != nil {

		if err == mgo.ErrNotFound {
			// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
			HttpError(w, err.Error(), http.StatusNotFound)
			return
		}
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

var routes = map[string]map[string]Handler{
	"HEAD": {},
	"GET": {
		"/pools/{name:.*}/inspect": getPoolJSON,
		"/pools/ps":                getPoolsJSON,
		"/pools/json":              getPoolsJSON,
		"/users/{name:.*}/login":   getUserLogin,
	},
	"POST": {
		"/pools/register": postPoolsRegister,
		"/users/register": postUsersRegister,
	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": OptionsHandler,
	},
}

func NewMainHandler(ctx context.Context) (http.Handler, error) {

	config := utils.GetAPIServerConfig(ctx)
	session, err := mgo.Dial(config.MgoURLs)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	c := utils.PutMgoSession(ctx, session)

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	c1 := utils.PutRedisClient(c, client)

	return NewHandler(c1, routes), nil
}

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
	result := store.User{}
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
	fmt.Printf(w,"ok")
}

type UsersRegisterRequest struct {
	store.User
}

func postUsersRegister(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		name = r.Form.Get("name")
		pass = r.Form.Get("pass")
	)

	req := UsersRegisterRequest{}

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

	//TODO mongodb需要在user.name有唯一性索引
	c.Find(bson.M{"Name": req.Name}).Count()

	//注册用户时候未分配权限
	if err := c.Insert(&store.User{Name: req.Name,
		Id:       bson.NewObjectId(),
		Pass:     req.Pass,
		Mail:     req.Mail,
		Comments: req.Comments,
	}); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Name", req.Name)
}

type PoolsRegisterRequest struct {
	Driver     string
	DriverOpts *store.DriverOpts
	Labels     map[string]interface{} `json:",omitempty"`
}

func getPoolsJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	mgoSession, err := utils.GetMgoSession(ctx)

	if err != nil {
		//走不到这里的,ctx中必然有mgoSesson
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	var result []*store.PoolInfo
	if err := c.Find(bson.M{}).All(&result); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

func postPoolsRegister(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		name = r.Form.Get("name")
	)

	req := PoolsRegisterRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	poolInfo := &store.PoolInfo{
		Id:             bson.NewObjectId(),
		Name:           name,
		Driver:         req.Driver,
		DriverOpts:     req.DriverOpts,
		Labels:         req.Labels,
		ProxyEndpoints: make([]string, 1),
	}

	mgoSession, err := utils.GetMgoSession(ctx)

	if err != nil {
		//走不到这里的
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

	c := mgoSession.DB(mgoDB).C("pool")

	n, err := c.Find(bson.M{"name": name}).Count()
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n >= 1 {
		HttpError(w, "the pool is exist", http.StatusConflict)
		return
	}

	p, err := proxy.NewProxyInstanceAndStart(ctx, poolInfo)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	poolInfo.ProxyEndpoints[0] = p.Endpoint()
	poolInfo.Status = "running"

	if err = c.Insert(poolInfo); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Name", name)

}
