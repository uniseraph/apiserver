package utils

import (
	"context"
	"github.com/Sirupsen/logrus"
	"net/http"
)

func GetSessionContent(ctx context.Context, w http.ResponseWriter, r *http.Request) (sessionContent map[string]string, err error) {
	cookie, err := r.Cookie("sessionID")
	if err != nil {
		return
	}

	sessionID := cookie.Value
	redisClient, err := GetRedisClient(ctx)
	if err != nil {
		return
	}

	//通过保存在session中的内容
	// - uid
	// - roleSet
	//来判断当前登录用户是否有权限
	content := redisClient.HGetAll(RedisSessionKey(sessionID))
	logrus.Debugf("HGETALL content: %#v", content)
	sessionContent, err = redisClient.HGetAll(RedisSessionKey(sessionID)).Result()
	logrus.Infof("SessionContent: %#v", sessionContent)
	//如果没有找到或者redis出错
	//则认证失败
	if err != nil {
		HttpError(w, err.Error(), http.StatusUnauthorized)
		return
	}
	return
}

//清楚redis中的session记录
//实现当前用户登出
func DestroySession(ctx context.Context, r *http.Request) (err error) {
	cookie, err := r.Cookie("sessionID")
	if err != nil {
		return
	}

	sessionID := cookie.Value
	redisClient, err := GetRedisClient(ctx)
	if err != nil {
		return
	}

	cmd := redisClient.Del(RedisSessionKey(sessionID))
	logrus.Debugf("del session for id: %s, response: %#v", sessionID, cmd)
	return
}
