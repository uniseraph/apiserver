package handlers

import (
	"context"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"github.com/zanecloud/apiserver/types"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"github.com/Sirupsen/logrus"
)

type ResponseBody struct {
	Code    int
	Message string
}

type MyHandler struct {
	h Handler
	roleset types.Roleset
}


var routes = map[string]map[string]Handler{
	"HEAD": {},
	"GET":  {
		"/users/{name:.*}/login":   getUserLogin,
		"/users/{id:.*}/inspect":   checkUserPermission(getUserInspect,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN),
		"/users/{id:.*}/detail":    checkUserPermission(getUserInspect,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/users/ps":                checkUserPermission(getUsersJSON,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN),
		"/users/list":              checkUserPermission(getUsersJSON,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/users/current":           getUserCurrent,

		"/teams/{name:.*}/inspect": checkUserPermission(getTeamJSON,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/teams/ps":                checkUserPermission(getTeamsJSON,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN),
		"/teams/list":              checkUserPermission(getTeamsJSON,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN),

	},
	"POST": {
		"/pools/{id:.*}/inspect": checkUserPermission(getPoolJSON,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN ),
		"/pools/register":          checkUserPermission(postPoolsRegister,types.ROLESET_SYSADMIN),
		"/pools/ps":                checkUserPermission(getPoolsJSON,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/pools/json":              checkUserPermission(getPoolsJSON,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),

		"/users/{name:.*}/login":   getUserLogin,
		"/users/current":           getUserCurrent,
		"/users/create":            checkUserPermission(postUsersCreate,types.ROLESET_SYSADMIN),
		"/users/{id:.*}/inspect":   checkUserPermission(getUserInspect,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/users/{id:.*}/detail":    checkUserPermission(getUserInspect,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/users/ps":                checkUserPermission(getUsersJSON,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/users/list":              checkUserPermission(getUsersJSON,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN),
		"/users/{id:.*}/resetpass": checkUserPermission(postUserResetPass,types.ROLESET_SYSADMIN),
		"/users/{id:.*}/remove":    checkUserPermission(postUserRemove,types.ROLESET_SYSADMIN),
		"/users/{id:.*}/update":     checkUserPermission(postUserUpdate,types.ROLESET_SYSADMIN |  types.ROLESET_NORMAL),
		"/users/{id:.*}/join":    checkUserPermission(postUserJoin, types.ROLESET_SYSADMIN  ),
		"/users/{id:.*}/quit":    checkUserPermission(postUserQuit, types.ROLESET_SYSADMIN  ),

		"/teams/{name:.*}/inspect": checkUserPermission(getTeamJSON,types.ROLESET_NORMAL |  types.ROLESET_SYSADMIN),
		"/teams/ps":                checkUserPermission(getTeamsJSON,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN),
		"/teams/list":              checkUserPermission(getTeamsJSON,types.ROLESET_NORMAL | types.ROLESET_SYSADMIN),
		"/teams/create":            checkUserPermission(postTeamsCreate,types.ROLESET_SYSADMIN ),
		"/teams/{id:.*}/update":     checkUserPermission(postTeamUpdate,types.ROLESET_SYSADMIN ) ,
		"/teams/{id:.*}/appoint": checkUserPermission(postTeamAppoint,types.ROLESET_SYSADMIN),
		"/teams/{id:.*}/revoke": checkUserPermission(postTeamRevoke,types.ROLESET_SYSADMIN),
		"/teams/{id:.*}/remove":  checkUserPermission(postTeamRemove,types.ROLESET_SYSADMIN),
	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": OptionsHandler,
	},
}


func checkUserPermission1(h Handler , roleset types.Roleset) Handler {

	wrap := func(ctx context.Context, w http.ResponseWriter, r *http.Request){

		h(ctx,w,r)
	}
	return wrap
}


func checkUserPermission(h Handler , roleset types.Roleset) Handler {


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

		c1 := utils.PutCurrentUser(ctx,&result)

		h(c1,w,r)


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

	return NewHandler(c1, routes), nil
}
