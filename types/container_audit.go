package types

import (
	"gopkg.in/mgo.v2/bson"
)

type ContainerAuditUser struct {
	Id   bson.ObjectId "_id"
	Name string
}

type ContainerAuditPool struct {
	Id   bson.ObjectId "_id"
	Name string
}

type ContainerAuditApplication struct {
	Id      bson.ObjectId "_id"
	Name    string
	Title   string
	Version string
}

type ContainerAuditService struct {
	Id    bson.ObjectId "_id"
	Name  string
	Title string
}

type ContainerAuditContainer struct {
	Id   bson.ObjectId "_id"
	Name string
}

//容器审计的跟踪模型
//
type ContainerAuditTrace struct {
	Id    bson.ObjectId "_id"
	Token string        //临时有效的token

	//当前用户
	UserId bson.ObjectId
	User   ContainerAuditUser

	//被操作资源
	PoolId        bson.ObjectId
	Pool          ContainerAuditPool
	ApplicationId bson.ObjectId
	Application   ContainerAuditApplication
	ServiceId     bson.ObjectId
	Service       ContainerAuditService
	ContainerId   bson.ObjectId
	Container     ContainerAuditContainer

	CreatedTime int64 `json:",omitempty"`
}

//容器审计
type ContainerAuditLog struct {
	Id bson.ObjectId "_id"
	//客户端IP
	Ip string
	//跟踪ID，用于某次会话的统计，就是TOKEN
	TraceId string

	//用户操作行为
	Cmd       string
	Arguments []string
	Stderr    string
	Stdout    string
	Stdin     string
	ExitCode  int8

	//本次审计操作是否被允许
	Permission bool

	CreatedTime int64 `json:",omitempty"`
}
