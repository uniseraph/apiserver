package handlers

import (
	"context"
	"net/http"

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

func checkUserPermission(h Handler, roleset types.Roleset) Handler {

	wrap := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

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

		if err := c.Find(bson.M{"$or": []bson.M{bson.M{"_id": bson.ObjectIdHex(uid)}}}).One(&result); err != nil {

			if err == mgo.ErrNotFound {
				HttpError(w, fmt.Sprintf("no such a user id is %s", uid), http.StatusNotFound)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if roleset&result.RoleSet == 0 {

			logrus.Infof("current roleset  is %d ,current user is %#v , so it no permission", roleset, result)

			HttpError(w, "no permission", http.StatusMethodNotAllowed)
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