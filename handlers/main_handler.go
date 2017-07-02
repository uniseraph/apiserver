package handlers

import (
	"context"
	"net/http"
	"strconv"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	"strings"
	"encoding/json"
)

type ResponseBody struct {
	Code    int
	Message string
}
type Handler func(c context.Context, w http.ResponseWriter, r *http.Request)
type OpPermissionCheckHandler func (h Handler, roleset types.Roleset) Handler

type MyHandler struct {
	h         Handler                       // 业务逻辑
	opChecker OpPermissionCheckHandler      // 检查当前用户的角色是否满足需求， 只判断行为权限，不判断数据权限
	roleset   types.Roleset                 // 只有拥有这些角色的用户才有权限
}

var routes = map[string]map[string]*MyHandler{
	"HEAD": {},
	"GET":  {
		"/users/{name:.*}/login":   &MyHandler{h: getUserLogin},
		"/users/current":           &MyHandler{h: getUserCurrent },
		"/users/{id:.*}/inspect":   &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/detail":    &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/ps":                &MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/list":              &MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/inspect":   &MyHandler{h: getTeamJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/ps":                &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/list":              &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

	},
	"POST": {
		"/pools/{id:.*}/inspect": &MyHandler{h: getPoolJSON,  opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/register":        &MyHandler{h: postPoolsRegister, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/ps":              &MyHandler{h: getPoolsJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/json":            &MyHandler{h: getPoolsJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

		"/users/{name:.*}/login":   &MyHandler{h: getUserLogin},
		"/users/current":           &MyHandler{h: getUserCurrent },
		"/users/create":            &MyHandler{h: postUsersCreate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/inspect":   &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/detail":    &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/ps":                &MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/list":              &MyHandler{h: getUsersJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/resetpass": &MyHandler{h: postUserResetPass, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/remove":    &MyHandler{h: postUserRemove, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/update":    &MyHandler{h: postUserUpdate, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/join":      &MyHandler{h: postUserJoin, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/quit":      &MyHandler{h: postUserQuit, opChecker: checkUserPermission,roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

		"/teams/create":            &MyHandler{h: postTeamsCreate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/inspect":   &MyHandler{h: getTeamJSON, opChecker: checkUserPermission,roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/ps":                &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/list":              &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/update":    &MyHandler{h: postTeamUpdate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/appoint":   &MyHandler{h: postTeamAppoint, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/revoke":    &MyHandler{h: postTeamRevoke, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/remove":    &MyHandler{h: postTeamRemove, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},



		//"/actions/check" : &MyHandler{h: postActionsCheck } ,
	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"":&MyHandler{h: OptionsHandler} ,
	},
}

func checkUserPermission1(h Handler, roleset types.Roleset) Handler {

	wrap := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

		h(ctx, w, r)
	}
	return wrap
}

func checkUserPermission(h Handler, rs types.Roleset) Handler {

	wrap := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("sessionID")
		if err != nil {
			HttpError(w, "please login", http.StatusForbidden)
			return
		}

		sessionID := cookie.Value
		redisClient, err := utils.GetRedisClient(ctx)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//通过保存在session中的内容
		// - uid
		// - roleSet
		//来判断当前登录用户是否有权限
		content := redisClient.HGetAll(utils.RedisSessionKey(sessionID))
		logrus.Debugf("HGETALL content: %#v", content)
		sessionContent, err := redisClient.HGetAll(utils.RedisSessionKey(sessionID)).Result()
		logrus.Infof("SessionContent: %#v", sessionContent)
		//如果没有找到或者redis出错
		//则认证失败
		if err != nil {
			HttpError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		//如果session中uid字段为空
		//则认证失败
		var uid = string(sessionContent["uid"])
		if len(uid) == 0 {
			HttpError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		//校验权限是否满足要求
		var roleSet types.Roleset
		//如果权限字段为空，则给用户默认权限
		//否则使用redis中session缓存写入的权限
		if len(sessionContent["roleSet"]) == 0 {
			roleSet = types.ROLESET_DEFAULT
		}else {
			value, err := strconv.ParseInt(sessionContent["roleSet"], 10, 64)
			if err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			roleSet = types.Roleset(value)
		}

		if rs & roleSet == 0 {
			logrus.Infof("current roleset  is %d ,current user id is %s , so it no permission", roleSet, uid)
			HttpError(w, "no permission", http.StatusMethodNotAllowed)
			return
		}

		//如果鉴权成功
		//根据uid把当前用户信息load到context中
		//以便request的剩余生命周期里，可以通过context直接得到用户信息
		//TODO
		//留不留都行，不是每个API都需要拿到用户全部信息
		mgoSession, err := utils.GetMgoSessionClone(ctx)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer mgoSession.Close()

		mgoDB := utils.GetAPIServerConfig(ctx).MgoDB

		c := mgoSession.DB(mgoDB).C("user")

		result := types.User{}

		if err := c.Find(bson.M{"$or": []bson.M{bson.M{"_id": bson.ObjectIdHex(uid)}}}).One(&result); err != nil {

			if err == mgo.ErrNotFound {
				HttpError(w, fmt.Sprintf("no such a user id is %s", uid), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}


		c1 := utils.PutCurrentUser(ctx, &result)

		h(c1, w, r)

	}

	return wrap
}

func NewMainHandler(ctx context.Context) (http.Handler, error) {

	config := utils.GetAPIServerConfig(ctx)
	session, err := mgo.Dial(config.MgoURLs)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	c := utils.PutMgoSession(ctx, session)

	logrus.Infof("redis address is : %s", config.RedisAddr)
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if _, err := client.Ping().Result(); err != nil {
		return nil, err
	}

	c1 := utils.PutRedisClient(c, client)


	r := mux.NewRouter()

	SetupPrimaryRouter(r, c1, routes)

	r.Path("/api/actions/check").Methods(http.MethodPost).HandlerFunc(  func (w http.ResponseWriter , r * http.Request){

	//	logrus.WithFields(logrus.Fields{"ctx":c1}).Debugf("call /api/actions/check")

		checkUserPermission(postActionsCheck,types.ROLESET_NORMAL|types.ROLESET_SYSADMIN)(c1,w,r)
	})


	fsh := http.StripPrefix("/",http.FileServer(http.Dir(config.RootDir)))

	//r.Path("/").Methods(http.MethodGet).Handler(http.StripPrefix("/",fsh))

	r.PathPrefix("/").HandlerFunc( func (w http.ResponseWriter , r * http.Request){

		logrus.WithFields(logrus.Fields{"method": r.Method, "uri": r.RequestURI }).Debug("HTTP request received")


		fsh.ServeHTTP(w,r)
	})

	return r , nil

//	return NewHandler(c1, routes), nil
}

func SetupPrimaryRouter(r *mux.Router, ctx context.Context, rs map[string]map[string]*MyHandler) {
	for method, mappings := range rs {
		for route, myHandler := range mappings {
			logrus.WithFields(logrus.Fields{"method": method, "route": route}).Debug("Registering HTTP route")

			localRoute := route
			localHandler := myHandler
			wrap := func(w http.ResponseWriter, r *http.Request) {
				logrus.WithFields(logrus.Fields{"method": r.Method, "uri": r.RequestURI , "localHandler":localHandler}).Debug("HTTP request received")

				if localHandler.opChecker !=nil {
					localHandler.opChecker(localHandler.h,localHandler.roleset)(ctx,w,r)
				}else{
					localHandler.h(ctx,w,r)
				}
			}
			localMethod := method

			//r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			r.Path("/api"+localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}
}


func BoolValue(r *http.Request, k string) bool {
	s := strings.ToLower(strings.TrimSpace(r.FormValue(k)))
	return !(s == "" || s == "0" || s == "no" || s == "false" || s == "none")
}

// Default handler for methods not supported by clustering.
func notImplementedHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	utils.HttpError(w, "Not supported in clustering mode.", http.StatusNotImplemented)
}

func OptionsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HttpError(w http.ResponseWriter, err string, status int) {
	utils.HttpError(w, err, status)
}
//"/actions/check" : &MyHandler{h: postActionsCheck } ,
func postActionsCheck(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	currentUser, err := utils.GetCurrentUser(ctx)
	if err != nil {
		HttpError(w, err.Error(),http.StatusForbidden)
		return
	}

	req := ActionsCheckRequest{}

	result := ActionCheckResponse{
		Action2Result: map[string]bool{},
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	//这是系统初始化的变量，所以不需要判断是否存在
	action2MyHandler , _ := routes["POST"]

	for  _, action :=  range req.Actions {

		if myHandler , ok := action2MyHandler[action] ; ok {

			//所有角色都有权限
			if myHandler.opChecker == nil {

				result.Action2Result[action] =true

			}else {
				if myHandler.roleset & currentUser.RoleSet !=0 {
					result.Action2Result[action] =true
				}else{
					result.Action2Result[action] = false
				}

			}
		}

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}