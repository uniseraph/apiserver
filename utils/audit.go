package utils

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/proxy/swarm"
	"github.com/zanecloud/apiserver/types"
	"strings"
	"time"
)

/*
	生成登录SSH用的临时唯一Token
*/
func CreateSSHSession(ctx context.Context, container swarm.Container, user *types.User) (token string, err error) {
	redis, err := GetRedisClient(ctx)
	if err != nil {
		return nil, err
	}

	sessionUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	//去掉uuid的横线
	r := strings.NewReplacer("-", "_")
	rdsKey := r.Replace(sessionUUID.String())

	rdsContent := map[string]interface{}{
		"cname": container.Name,
		"cid":   container.Id,
		"uid":   user.Id.Hex(),
		"uname": user.Name,
		"pname": container.PoolName,
		//存入tunnel的url
		//"url": container.url,
	}
	err = redis.HMSet(ContainerAuditSessionKey(rdsKey), rdsContent).Err()
	if err != nil {
		logrus.Fatalf("Redis hmset error: %#v", err)
		return nil, err
	}
	//五分钟失效
	age := time.Minute * 5
	//设置session一周超时
	//一周后再登录，会找不到redis中的key，导致认证不再可以通过，需要重新登录
	redis.Expire(ContainerAuditSessionKey(rdsKey), age)

	return rdsKey, nil
}

//格式化返回的SSH连接字符串
func GenerateSSHToken(user string, token string) (ssh string) {
	return fmt.Sprintf("ssh://%s@%s", user, token)
}

//从SSH字符串中，解析出Redis中的KEY
func ParseSSHToken(token string) (key string, err error) {
	if len(token) <= 0 {
		return nil, errors.New("Token is empty string")
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
