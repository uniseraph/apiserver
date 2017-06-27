package handlers

import (
	"context"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
)


var routes = map[string]map[string]Handler{
	"HEAD": {},
	"GET": {
		"/pools/{name:.*}/inspect": getPoolJSON,
		"/teams/{name:.*}/inspect": getTeamJSON,

		"/pools/ps":                getPoolsJSON,
		"/pools/json":              getPoolsJSON,
		"/users/{name:.*}/login":   getUserLogin,
	},
	"POST": {
		"/pools/register":         postPoolsRegister,
		"/users/create":           postUsersCreate,
		"/users/{name:*}/roles":   postUserRoleSet,
		"/teams/create":           postTeamsCreate,
		"/teams/{name:.*}/join":   postTeamJoin,
		"/teams/{name:*}/appoint": postTeamAppoint,
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
