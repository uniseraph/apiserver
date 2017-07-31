package handlers

import (
	"context"
	"net/http"
	"strconv"

	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ResponseBody struct {
	Code    int
	Message string
}
type Handler func(c context.Context, w http.ResponseWriter, r *http.Request)
type OpPermissionCheckHandler func(h Handler, roleset types.Roleset) Handler

type MyHandler struct {
	h         Handler                  // 业务逻辑
	opChecker OpPermissionCheckHandler // 检查当前用户的角色是否满足需求， 只判断行为权限，不判断数据权限
	roleset   types.Roleset            // 只有拥有这些角色的用户才有权限
}

var routers = map[string]map[string]*MyHandler{
	"HEAD": {},
	"GET": {

		"/containers/{id:.*}/inspect": &MyHandler{h: getContainerJSON, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/containers/{id:.*}/logs":    &MyHandler{h: getContainerLogs, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},

		"/users/{name:.*}/login": &MyHandler{h: postSessionCreate},
		"/users/current":         &MyHandler{h: getUserCurrent, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/inspect": &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/detail":  &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/ps":              &MyHandler{h: getUsersJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/list":            &MyHandler{h: getUsersJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/user/pools":            &MyHandler{h: getUserPools, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/teams/{id:.*}/inspect": &MyHandler{h: getTeamJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/ps":              &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/list":            &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

		/*
			参数目录树
		*/

		"/envs/trees/list":                   &MyHandler{h: getTrees, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/trees/create":                 &MyHandler{h: createTree, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/trees/{id:.*}/update":         &MyHandler{h: updateTree, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/trees/{id:.*}/remove":         &MyHandler{h: deleteTree, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/dirs/list":                    &MyHandler{h: getTreeDirs, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/dirs/create":                  &MyHandler{h: createDir, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/dirs/{id:.*}/update":          &MyHandler{h: updateDir, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/dirs/{id:.*}/remove":          &MyHandler{h: deleteDir, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/list":                  &MyHandler{h: getTreeValues, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/detail":        &MyHandler{h: getTreeValueDetails, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/values/create":                &MyHandler{h: createValue, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/update":        &MyHandler{h: updateValue, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/remove":        &MyHandler{h: deleteValue, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/update-values": &MyHandler{h: updateValueAttributes, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/value/get":                    &MyHandler{h: getValue, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/values/search":                &MyHandler{h: getEnvKeyNameWithPrefix, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

		/*
			容器日志审计
		*/

		"/audit/ssh":        &MyHandler{h: createSSHSession, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/audit/login":      &MyHandler{h: validateSSHSession},
		"/audit/log":        &MyHandler{h: createAuditLog},
		"/audit/log/update": &MyHandler{h: updateAuditLog},
		"/audit/list":       &MyHandler{h: getContainerAuditList, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

		/*
			系统审计日志
		*/

		"/logs/list": &MyHandler{h: getSystemAuditList, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},

		/*
			应用授权
		*/
		"/applications/{id:.*}/add-team":    &MyHandler{h: addApplicationTeam, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/remove-team": &MyHandler{h: removeApplicationTeam, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/add-user":    &MyHandler{h: addApplicationMember, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/remove-user": &MyHandler{h: removeApplicationMember, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
	},
	"POST": {
		"/pools/{id:.*}/refresh":     &MyHandler{h: postPoolsFlush, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/{id:.*}/inspect":     &MyHandler{h: getPoolJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/register":            &MyHandler{h: postPoolsRegister, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/ps":                  &MyHandler{h: getPoolsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/json":                &MyHandler{h: getPoolsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/pools/{id:.*}/add-team":    &MyHandler{h: addPoolTeam, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/{id:.*}/remove-team": &MyHandler{h: removePoolTeam, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/{id:.*}/add-user":    &MyHandler{h: addPoolMember, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/{id:.*}/remove-user": &MyHandler{h: removePoolMember, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/{id:.*}/update":      &MyHandler{h: updatePool, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/pools/{id:.*}/remove":      &MyHandler{h: deletePool, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},

		"/users/{name:.*}/login":   &MyHandler{h: postSessionCreate},
		"/users/current":           &MyHandler{h: getUserCurrent, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/create":            &MyHandler{h: postUsersCreate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/inspect":   &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/detail":    &MyHandler{h: getUserInspect, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/ps":                &MyHandler{h: getUsersJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/list":              &MyHandler{h: getUsersJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/users/{id:.*}/resetpass": &MyHandler{h: postUserResetPass, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/remove":    &MyHandler{h: postUserRemove, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/users/{id:.*}/update":    &MyHandler{h: postUserUpdate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/join":      &MyHandler{h: postUserJoin, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/users/{id:.*}/quit":      &MyHandler{h: postUserQuit, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},
		"/user/pools":              &MyHandler{h: getUserPools, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN | types.ROLESET_NORMAL},

		"/teams/create":          &MyHandler{h: postTeamsCreate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/inspect": &MyHandler{h: getTeamJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/ps":              &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/list":            &MyHandler{h: getTeamsJSON, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/update":  &MyHandler{h: postTeamUpdate, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/appoint": &MyHandler{h: postTeamAppoint, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/revoke":  &MyHandler{h: postTeamRevoke, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/teams/{id:.*}/remove":  &MyHandler{h: postTeamRemove, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},

		"/session/{name:.*}/login": &MyHandler{h: postSessionCreate},
		"/session/logout":          &MyHandler{h: postSessionDestroy, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},

		//"/actions/check" : &MyHandler{h: postActionsCheck } ,

		/*
			参数目录树
		*/

		"/envs/trees/list":                   &MyHandler{h: getTrees, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/trees/create":                 &MyHandler{h: createTree, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/trees/{id:.*}/update":         &MyHandler{h: updateTree, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/trees/{id:.*}/remove":         &MyHandler{h: deleteTree, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/dirs/list":                    &MyHandler{h: getTreeDirs, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/dirs/create":                  &MyHandler{h: createDir, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/dirs/{id:.*}/update":          &MyHandler{h: updateDir, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/dirs/{id:.*}/remove":          &MyHandler{h: deleteDir, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/list":                  &MyHandler{h: getTreeValues, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/detail":        &MyHandler{h: getTreeValueDetails, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/values/create":                &MyHandler{h: createValue, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/update":        &MyHandler{h: updateValue, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/remove":        &MyHandler{h: deleteValue, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/values/{id:.*}/update-values": &MyHandler{h: updateValueAttributes, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/envs/value/get":                    &MyHandler{h: getValue, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/envs/values/search":                &MyHandler{h: getEnvKeyNameWithPrefix, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},

		/*
			容器日志审计
		*/

		"/audit/ssh":        &MyHandler{h: createSSHSession, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_SYSADMIN},
		"/audit/login":      &MyHandler{h: validateSSHSession},
		"/audit/log":        &MyHandler{h: createAuditLog},
		"/audit/log/update": &MyHandler{h: updateAuditLog},
		"/audit/list":       &MyHandler{h: getContainerAuditList, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},

		/*
			系统审计日志
		*/

		"/logs/list": &MyHandler{h: getSystemAuditList, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},

		"/containers/{id:.*}/inspect": &MyHandler{h: getContainerJSON, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/containers/{id:.*}/logs":    &MyHandler{h: getContainerLogs, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/containers/list":            &MyHandler{h: getContainerList, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},

		"/applications/list":                       &MyHandler{h: getApplicationList, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/history":            &MyHandler{h: getApplicationHistory, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/inspect":            &MyHandler{h: getApplication, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/detail":             &MyHandler{h: getApplication, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/start":              &MyHandler{h: startApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/restart":            &MyHandler{h: restartApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/stop":               &MyHandler{h: stopApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/rollback":           &MyHandler{h: rollbackApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/upgrade":            &MyHandler{h: upgradeApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/scale":              &MyHandler{h: scaleApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/containers/:id/ssh-info":    &MyHandler{h: getContainerSSHInfo, opChecker: checkUserPermission, roleset: types.ROLESET_NORMAL | types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/create":                     &MyHandler{h: createApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/add-team":           &MyHandler{h: addApplicationTeam, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/remove-team":        &MyHandler{h: removeApplicationTeam, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/add-user":           &MyHandler{h: addApplicationMember, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/remove-user":        &MyHandler{h: removeApplicationMember, opChecker: checkUserPermission, roleset: types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/containers/list":    &MyHandler{h: getContainerList, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/containers/{id:.*}/restart": &MyHandler{h: restartContainer, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/applications/{id:.*}/remove":             &MyHandler{h: removeApplication, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},

		"/templates/list":            &MyHandler{h: getTemplateList},
		"/templates/{id:.*}/inspect": &MyHandler{h: getTemplate},
		"/templates/{id:.*}/detail":  &MyHandler{h: getTemplate},
		"/templates/create":          &MyHandler{h: createTemplate, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/templates/{id:.*}/copy":    &MyHandler{h: copyTemplate, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/templates/{id:.*}/update":  &MyHandler{h: updateTemplate, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
		"/templates/{id:.*}/remove":  &MyHandler{h: removeTemplate, opChecker: checkUserPermission, roleset: types.ROLESET_APPADMIN | types.ROLESET_SYSADMIN},
	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": &MyHandler{h: OptionsHandler},
	},
}

func checkUserPermission(h Handler, rs types.Roleset) Handler {

	wrap := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sessionID")
		//如果cookie中不存在sessionID
		//则err不为空
		//则认为禁止登陆
		if err != nil {
			HttpError(w, "please login", http.StatusUnauthorized)
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
		//content := redisClient.HGetAll(utils.RedisSessionKey(sessionID))
		//logrus.Debugf("HGETALL content: %#v", content)
		sessionContent, err := redisClient.HGetAll(utils.RedisSessionKey(sessionID)).Result()
		//logrus.Infof("SessionContent: %#v", sessionContent)
		//如果没有找到或者redis出错
		//则认证失败
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//如果session在redis中内容为空
		//则认证失败
		if len(sessionContent) == 0 {
			HttpError(w, "sessionContent is empty", http.StatusUnauthorized)
			return
		}
		//如果session中uid字段为空
		//则认证失败
		var uid = string(sessionContent["uid"])
		if len(uid) == 0 {
			HttpError(w, "session data error for uid field.", http.StatusInternalServerError)
			return
		}
		//校验权限是否满足要求
		var roleSet types.Roleset
		//如果权限字段为空，则给用户默认权限
		//否则使用redis中session缓存写入的权限
		if len(sessionContent["roleSet"]) == 0 {
			roleSet = types.ROLESET_DEFAULT
		} else {
			value, err := strconv.ParseInt(sessionContent["roleSet"], 10, 64)
			if err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			roleSet = types.Roleset(value)
		}

		if rs&roleSet == 0 {
			logrus.Infof("current roleset  is %d ,current user id is %s , so it no permission", roleSet, uid)
			HttpError(w, "no permission", http.StatusForbidden)
			return
		}

		//如果鉴权成功
		//根据uid把当前用户信息load到context中
		//以便request的剩余生命周期里，可以通过context直接得到用户信息
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
				HttpError(w, fmt.Sprintf("no such a user id is %s", uid), http.StatusForbidden)
				return
			}

			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c1 := utils.PutCurrentUser(ctx, &result)

		//校验身份成功后
		//每次操作，都会使得
		//当前session的超时时间，更新为未来15分钟内有效
		age := time.Minute * 15
		//设置session5分钟超时
		//如果15分钟之内没有操作
		//会找不到redis中的key，导致认证不再可以通过，需要重新登录
		redisClient.Expire(utils.RedisSessionKey(sessionID), age)

		h(c1, w, r)

	}

	return wrap
}

func NewMainHandler(ctx context.Context, config *types.APIServerConfig) (http.Handler, error) {
	r := mux.NewRouter()

	for method, mappings := range routers {
		for route, myHandler := range mappings {
			logrus.WithFields(logrus.Fields{"method": method, "route": route}).Debug("Registering HTTP route")

			localRoute := route
			localHandler := myHandler
			wrap := func(w http.ResponseWriter, req *http.Request) {
				logrus.WithFields(logrus.Fields{"method": req.Method, "uri": req.RequestURI, "localHandler": localHandler}).Debug("HTTP request received")

				if localHandler.opChecker != nil {
					localHandler.opChecker(localHandler.h, localHandler.roleset)(ctx, w, req)
				} else {
					localHandler.h(ctx, w, req)
				}
			}
			localMethod := method
			//r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			r.Path("/api" + localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

	r.Path("/api/actions/check").Methods(http.MethodPost).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkUserPermission(postActionsCheck, types.ROLESET_NORMAL|types.ROLESET_SYSADMIN)(ctx, w, r)
	})

	fsh := http.StripPrefix("/", http.FileServer(http.Dir(config.RootDir)))
	//r.Path("/").Methods(http.MethodGet).Handler(http.StripPrefix("/",fsh))

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{"method": r.Method, "uri": r.RequestURI}).Debug("HTTP request received")
		fsh.ServeHTTP(w, r)
	})

	return r, nil
}

func OptionsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HttpError(w http.ResponseWriter, err string, status int) {
	utils.HttpError(w, err, status)

}

func HttpOK(w http.ResponseWriter, result interface{}) {
	utils.HttpOK(w, result)
}

//"/actions/check" : &MyHandler{h: postActionsCheck } ,
func postActionsCheck(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	currentUser, err := getCurrentUser(ctx)
	if err != nil {
		HttpError(w, err.Error(), http.StatusForbidden)
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
	action2MyHandler, _ := routers["POST"]

	for _, action := range req.Actions {

		if myHandler, ok := action2MyHandler[action]; ok {

			//所有角色都有权限
			if myHandler.opChecker == nil {

				result.Action2Result[action] = true

			} else {
				if myHandler.roleset&currentUser.RoleSet != 0 {
					result.Action2Result[action] = true
				} else {
					result.Action2Result[action] = false
				}

			}
		}

	}

	HttpOK(w, result)
}
