package utils

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/store"
	"gopkg.in/mgo.v2"
)

const (
	KEY_REDIS_ADDR   = "redis.addr"
	KEY_REDIS_CLIENT = "redis.client"
	KEY_REDIS_UID    = "redis.uid"
)
const KEY_MGO_URLS = "mgo.urls"
const KEY_MGO_SESSION = "mgo.session"
const KEY_MGO_DB = "mgo.db"

//const KEY_POOL_NAME = "pool.name"
const KEY_PROXY_SELF = "proxy.self"
const KEY_POOL_CLIENT = "pool.client"
const KEY_LISTENER_ADDR = "addr"
const KEY_LISTENER_PORT = "port"
const KEY_APISERVER_CONFIG = "apiserver.config"

func GetAPIServerConfig(ctx context.Context) *store.APIServerConfig {
	config, ok := ctx.Value(KEY_APISERVER_CONFIG).(*store.APIServerConfig)
	if !ok {
		logrus.Errorf("can't get APIServerConfig by %s", KEY_APISERVER_CONFIG)
		panic("can't get APIServerConfig")
	}

	return config
}

func PutAPIServerConfig(ctx context.Context, config *store.APIServerConfig) context.Context {
	return context.WithValue(ctx, KEY_APISERVER_CONFIG, config)
}

func GetMgoSession(ctx context.Context) (*mgo.Session, error) {
	session, ok := ctx.Value(KEY_MGO_SESSION).(*mgo.Session)
	if !ok {
		logrus.Errorf("can't get mgoSession by %s", KEY_MGO_SESSION)
		return nil, errors.New("can't get mgoSession")
	}

	return session, nil
}

// Clone一个mgoSession ， 需要使用者自己close
func GetMgoSessionClone(ctx context.Context) (*mgo.Session, error) {

	session , err := GetMgoSession(ctx)
	if err !=nil {
		return nil ,err
	}


	return session.Clone(),nil
}


func PutMgoSession(ctx context.Context, mgoSession *mgo.Session) context.Context {
	return context.WithValue(ctx, KEY_MGO_SESSION, mgoSession)
}
func PutRedisClient(ctx context.Context, redisClient *redis.Client) context.Context {
	return context.WithValue(ctx, KEY_REDIS_CLIENT, redisClient)
}

func GetRedisClient(ctx context.Context) (*redis.Client, error) {
	client, ok := ctx.Value(KEY_REDIS_CLIENT).(*redis.Client)
	if !ok {
		logrus.Errorf("can't get redisClient by %s", KEY_REDIS_CLIENT)
		return nil, errors.New("can't get redisClient")
	}

	return client, nil
}
