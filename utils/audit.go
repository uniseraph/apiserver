package utils

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/types"
	"strings"
	"time"
)

/*
	生成登录SSH用的临时唯一Token
*/
func CreateSSHSession(ctx context.Context, cname string, cid string, scid string, aid string, sname string, user *types.User, pool *types.PoolInfo) (token string, err error) {
	redis, err := GetRedisClient(ctx)
	if err != nil {
		return "", err
	}

	sessionUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	//去掉uuid的横线
	r := strings.NewReplacer("-", "")
	rdsKey := r.Replace(sessionUUID.String())

	rdsContent := map[string]interface{}{
		"cname": cname,
		"cid":   cid,
		"scid":  scid,
		"uid":   user.Id.Hex(),
		"uname": user.Name,
		"pname": pool.Name,
		"pid":   pool.Id.Hex(),
		"aid":   aid,
		"sname": sname,
	}
	err = redis.HMSet(ContainerAuditSessionKey(rdsKey), rdsContent).Err()
	if err != nil {
		logrus.Fatalf("Redis hmset error: %#v", err)
		return "", err
	}
	//五分钟失效
	age := time.Minute * 5
	//设置key超时
	redis.Expire(ContainerAuditSessionKey(rdsKey), age)

	return rdsKey, nil
}

func RemoveSSHSession(ctx context.Context, token string) (err error) {
	redis, err := GetRedisClient(ctx)
	if err != nil {
		return err
	}

	redis.Del(ContainerAuditSessionKey(token))

	return nil
}

//格式化返回的SSH连接字符串
func GenerateSSHToken(token string, pool *types.PoolInfo) (ssh string) {
	return fmt.Sprintf("ssh -p %d %s@%s", pool.TunneldPort, token, pool.TunneldAddr)
}

//从SSH字符串中，解析出Redis中的KEY
func ParseSSHToken(token string) (key string, err error) {
	if len(token) <= 0 {
		return "", errors.New("Token is empty string")
	}
	arr := strings.Split(token, "@")

	k := arr[len(arr)-1]
	return k, nil
}

func FetchContainerFromSSHCache(ctx context.Context, key string) (info map[string]string, err error) {
	redis, err := GetRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	rdsContent, err := redis.HGetAll(ContainerAuditSessionKey(key)).Result()
	if err != nil {
		return nil, err
	}

	return rdsContent, nil
}

//解析一下格式的时间字符串
//2017-7-7 00:00
func parseTime(timeStr string) (t time.Time) {
	return time.Now()
}
