package swarm

import (
	"github.com/docker/docker/api/types"
	"gopkg.in/mgo.v2/bson"
)

// mongodb中Container表，只记录容器创建时间和状态，具体信息需要从集群中获取，避免同步
type Container struct {
	Id              bson.ObjectId "_id"
	ContainerId     string        //这是docker／swarm生成的id
	Name            string
	//PoolName        string
	PoolId          string
	Service         string
	Project         string
	IP              string
	ApplicationId   string
	ApplicationName string
	Status          string
	Memory          int64
	CPU             int64
	CPUExclusive    bool
	IsDeleted       bool
	GmtDeleted      int64
	GmtCreated      int64
	Node            *types.ContainerNode  `json:",omitempty"`
	State           *types.ContainerState `json:",omitempty"`
	StartedTime     int64
}
