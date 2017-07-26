package handlers

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"time"
)

type SessionCreateResp struct {
	Id   string
	Role uint64
}

//当前用户登录接口
func postSessionCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {

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
	user := types.User{}
	if err := mgoSession.DB(mgoDB).C("user").Find(bson.M{"name": name}).One(&user); err != nil {
		HttpError(w, err.Error(), http.StatusNotFound)
		return
	}

	logrus.Debugf("getUserLogin::get the user %#v", user)
	//校验用户输入的密码，与该ID的用户模型中Pass是否匹配
	if ok, err := utils.ValidatePassword(user, pass); ok != true || err != nil {
		/*
			系统审计
		*/
		opUser := user
		_ = utils.CreateSystemAuditLog(mgoSession.DB(mgoDB), r, opUser.Id.Hex(), types.SystemAuditModuleTypeUser, types.SystemAuditModuleOperationTypeLoginFailed, "", "", map[string]interface{}{"Name": name})

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
		"uid":     user.Id.Hex(),
		"roleSet": strconv.FormatUint(uint64(user.RoleSet), 10), //fmt.Sprintf("%d", result.RoleSet),
	}
	err = client.HMSet(utils.RedisSessionKey(sessionKey), sessionContents).Err()
	if err != nil {
		logrus.Fatalf("Redis hmset error: %#v", err)
		panic(err)
	}
	age := time.Minute * 15
	//设置session5分钟超时
	//如果5分钟之内没有操作
	//会找不到redis中的key，导致认证不再可以通过，需要重新登录
	client.Expire(utils.RedisSessionKey(sessionKey), age)

	sessionIDCookie := &http.Cookie{
		Name:     "sessionID",
		Value:    sessionKey,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   int(age),
	}

	logrus.Debugf("getUserLogin::get the cookie %#v", sessionIDCookie)

	http.SetCookie(w, sessionIDCookie)

	result := SessionCreateResp{
		Id:   user.Id.Hex(),
		Role: uint64(user.RoleSet),
	}
	HttpOK(w, result)
	/*
		系统审计
	*/
	opUser := user
	_ = utils.CreateSystemAuditLog(mgoSession.DB(mgoDB), r, opUser.Id.Hex(), types.SystemAuditModuleTypeUser, types.SystemAuditModuleOperationTypeLogined, "", "", map[string]interface{}{"Name": name})
}

//当前用户登出接口
func postSessionDestroy(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	err := utils.DestroySession(ctx, r)
	if err != nil {
		logrus.Fatalf("logout failed with error: %#v", err)
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	age := time.Hour * 24 * 7

	//清空sessionID
	sessionIDCookie := &http.Cookie{
		Name:     "sessionID",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		MaxAge:   int(age),
	}
	http.SetCookie(w, sessionIDCookie)

	HttpOK(w, nil)
}
