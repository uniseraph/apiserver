package utils

import (
	"context"
	"github.com/zanecloud/apiserver/store"
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

const KEY_MGO_URLS = "mgo.urls"
const KEY_MGO_SESSION = "mgo.session"
const KEY_MGO_DB = "mgo.db"
const KEY_POOL_NAME = "pool.name"
const KEY_PROXY_SELF = "proxy.self"
const KEY_POOL_CLIENT = "pool.client"
const KEY_LISTENER_ADDR = "addr"
const KEY_LISTENER_PORT = "port"
const KEY_APISERVER_CONFIG= "apiserver.config"


func GetAPIServerConfig(ctx context.Context) *store.APIServerConfig {
	config , ok :=	ctx.Value(KEY_APISERVER_CONFIG).(*store.APIServerConfig)
	if !ok {
		logrus.Errorf("can't get APIServerConfig by %s" , KEY_APISERVER_CONFIG)
		panic("can't get APIServerConfig")
	}

	return config
}


func GetMgoSession(ctx context.Context) *mgo.Session {
	session , ok :=	ctx.Value(KEY_MGO_SESSION).(*mgo.Session)
	if !ok {
		logrus.Errorf("can't get mgoSession by %s" , KEY_MGO_SESSION)
		panic("can't get mgoSession")
	}

	return session
}