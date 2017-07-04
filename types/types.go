package types

import (
	"crypto/tls"
	"github.com/docker/go-connections/tlsconfig"
	"gopkg.in/mgo.v2/bson"
	"math"
)

const (
	ROLESET_DEFAULT = math.MaxUint64

	ROLESET_NORMAL   = 1      //普通员工
	ROLESET_APPADMIN = 1 << 1 //应用管理员
	ROLESET_SYSADMIN = 1 << 2 //系统管理员
)

type APIServerConfig struct {
	MgoDB     string
	MgoURLs   string
	RedisAddr string
	Addr      string
	Port      int
	tlsConfig *tls.Config
	RootDir   string
}

type DriverOpts struct {
	Version    string
	EndPoint   string
	APIVersion string
	Labels     map[string]interface{} `json:",omitempty"`
	TlsConfig  *tlsconfig.Options     `json:",omitempty"`
	Opts       map[string]interface{} `json:",omitempty"`
}

type PoolInfo struct {
	Id bson.ObjectId "_id"

	Name   string
	Status string

	Driver         string
	DriverOpts     DriverOpts
	Labels         map[string]interface{} `json:",omitempty"`
	ProxyEndpoints []string               `json:",omitempty"`
}

type Roleset uint64

type User struct {
	Id   bson.ObjectId "_id"
	Name string
	Pass string `json:",omitempty"`
	Salt string `json:"-"`
	RoleSet     Roleset
	Email       string
	TeamIds     []bson.ObjectId
	Tel         string `json:tel",omitempty"`
	CreatedTime int64  `json:",omitempty"`
	Comments    string `json:",omitempty"`
}

type Leader struct {
	Id   string
	Name string
}

type Team struct {
	Id          bson.ObjectId "_id"
	Name        string
	Description string
	Leader      Leader
	//UserIds     []bson.ObjectId
	Users       []User
	CreatedTime int64 `json:",omitempty"`
}

//type TeamUser struct {
//	UserId string
//	TeamId string
//}

/*
	参数目录
	zheng.cui
*/

//参数目录树元数据

//EnvTreeMeta has one EnvTreeNodeDir entry point
type EnvTreeMeta struct {
	Id 	    bson.ObjectId "_id"
	Name        string
	Description string
	CreatedTime int64  `json:",omitempty"`
	UpdatedTime int64  `json:",omitempty"`
}

//EnvTreeNodeDir has many sub EnvTreeNodeDirs and EnvTreeNodeParamKeys} pairs
//EnvTreeNodeDir belongs to EnvTreeMeta
type EnvTreeNodeDir struct {
	Id 	    bson.ObjectId "_id"
	Name        string
	//一个父目录
	//最顶级的父目录为空，用于结合EnvTreeMeta查询该树的起点
	//EnvTreeNodeDir
	Parent      bson.ObjectId
	//多个子目录
	//EnvTreeNodeDir
	Children    []bson.ObjectId
	//多个值
	//EnvTreeNodeParamKey
	Keys        []bson.ObjectId
	//EnvTreeMeta
	Tree 	    bson.ObjectId
	CreatedTime int64  `json:",omitempty"`
	UpdatedTime int64  `json:",omitempty"`
}

//参数目录树节点的参数名称
//EnvTreeNodeParamKey has many EnvTreeNodeParamValue
type EnvTreeNodeParamKey struct {
	Id 	    bson.ObjectId "_id"
	Name        string
	//默认值
	Default     string
	//EnvTreeMeta
	Tree 	    bson.ObjectId
	CreatedTime int64  `json:",omitempty"`
	UpdatedTime int64  `json:",omitempty"`
}

//参数目录树节点的参数值
//EnvTreeNodeParamValue belongs to EnvTreeNodeParamKey
//EnvTreeNodeParamValue belongs to Pool
//这其实是一个Key和Pool的关联关系表
//用来查询一个Key被哪些Pool所用，并且每个值都是什么
type EnvTreeNodeParamValue struct {
	Id 	    bson.ObjectId "_id"
	Value  	    string
	//对应的参数名称
	//EnvTreeNodeParamKey
	Key         bson.ObjectId
	//EnvTreeMeta
	Tree 	    bson.ObjectId
	//PoolInfo
	Pool        bson.ObjectId
	CreatedTime int64  `json:",omitempty"`
	UpdatedTime int64  `json:",omitempty"`
}