package handlers

import (
	"context"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
)

type  ResponseBody struct {
	Code    int
	Message string
}


var routes = map[string]map[string]Handler{
	"HEAD": {},
	"GET":  {},
	"POST": {
		"/pools/{name:.*}/inspect": getPoolJSON,
		"/pools/register":          postPoolsRegister,
		"/pools/ps":                getPoolsJSON,
		"/pools/json":              getPoolsJSON,

		"/users/{name:.*}/login":    getUserLogin,
		"/users/create":             postUsersCreate,
		"/users/{id:.*}/inspect":  getUserInspect,
		"/users/{id:.*}/detail":  getUserInspect,
		"/users/ps":		     getUsersJSON,
		"/users/list":		     getUsersJSON,
		"/users/{name:.*}/roles":    postUserRoleSet,
		"/users/{name:.*}/remove":   postUserRemove,


		"/teams/{name:.*}/inspect": getTeamJSON,
		"/teams/ps":                getTeamsJSON,
		"/teams/list":              getTeamsJSON,
		"/teams/create":            postTeamsCreate,
		"/teams/{name:.*}/join":    postTeamJoin,
		"/teams/{name:.*}/appoint":  postTeamAppoint,
		"/teams/{name:.*}/remove" :  postTeamRemove ,
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
