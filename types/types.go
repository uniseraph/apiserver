package types

import (
	"crypto/tls"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/tlsconfig"
	"gopkg.in/mgo.v2/bson"
	"math"
)

const (
	ROLESET_DEFAULT = math.MaxUint64

	ROLESET_NORMAL   = 1      //普通员工
	ROLESET_APPADMIN = 1 << 1 //应用管理员
	ROLESET_SYSADMIN = 1 << 2 //系统管理员

	DEPLOYMENT_OPERATION_CREATE   = "create"
	DEPLOYMENT_OPERATION_UPGRADE  = "upgrade"
	DEPLOYMENT_OPERATION_ROLLBACK = "rollback"

	LABEL_CONTAINER_CPUS      = "com.zanecloud.omega.container.cpus"
	LABEL_CONTAINER_EXCLUSIVE = "com.zanecloud.omega.container.cpu.exclusive"

	LABEL_VOLUME_PREFIX     = "com.zanecloud.omega.disk"
	LABEL_VOLUME_MOUNTPOINT = "mountPoint"
	LABEL_VOLUME_MEDIATYPE  = "mediaType"
	LABEL_VOLUME_SIZE       = "size"
	LABEL_VOLUME_IOCLASS    = "ioClass"
	LABEL_VOLUME_EXCLUSIVE  = "exclusive"
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

	Name             string
	Status           string
	Provider         string
	CPUs             int
	Memory           int64
	Disk             int64
	ClusterStore     string
	ClusterAdvertise string
	Strategy         string
	Filters          string
	Driver           string
	DriverOpts       DriverOpts
	EnvTreeId        string
	EnvTreeName      string
	NodeCount        int
	TunneldAddr      string `json:",omitempty"`
	TunneldPort      int
	Labels           []string `json:",omitempty"`
	ProxyEndpoint    string   `json:",omitempty"`
	Containers       int
	UpdatedTime      int64
	CreatedTime      int64
}

type Roleset uint64

type User struct {
	Id             bson.ObjectId "_id"
	Name           string
	Pass           string `json:",omitempty"`
	Salt           string `json:"-"`
	RoleSet        Roleset
	Email          string
	TeamIds        []bson.ObjectId
	Tel            string          `json:",omitempty"`
	CreatedTime    int64           `json:",omitempty"`
	Comments       string          `json:",omitempty"`
	PoolIds        []bson.ObjectId //一个用户has many pool
	ApplicationIds []bson.ObjectId //一个用户has many application
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
	Users          []User
	CreatedTime    int64           `json:",omitempty"`
	PoolIds        []bson.ObjectId //一个Team has many pool
	ApplicationIds []bson.ObjectId //一个Team has many application
}

//type TeamUser struct {
//	UserId string
//	TeamId string
//}

//调用docker info，获取swarm集群的信息
type ClusterInfo struct {
	types.Info
	SystemStatus [][]string
}

type Node struct {
	//Id             bson.ObjectId "_id"
	PoolId         string
	Hostname       string
	Endpoint       string
	NodeId         string
	Status         string
	Containers     string
	ReservedCPUs   string
	ReservedMemory string

	//ContainersRunning int
	//ContainersPaused  int
	//ContainersStopped int
	Labels        map[string]string
	ServerVersion string
}

type Service struct {
	Title          string
	Name           string
	ImageName      string
	ImageTag       string
	CPU            string
	ExclusiveCPU   bool
	Memory         string
	NetworkMode    string
	ReplicaCount   int `json:",string"`
	ServiceTimeout int `json:",string"`
	Description    string
	Restart        string
	Command        string
	Envs           []Env
	Volumns        []Volumne
	Labels         []Label
	//Ports        []string
	Ports      []Port
	Privileged bool
	CapAdd     []string
	CapDrop    []string
	Mutex      string
}

type Port struct {
	SourcePort int `json:",string"`
	//LoadBalancerId string
	TargetGroupArn string //aliyun slb vservergroupId or aws elbv2 targetGroupArn
	//LoadBalancerId string //aliyun slb lbid
}
type Env struct {
	Label
}
type Label struct {
	Name  string
	Value string
}
type Volumne struct {
	//Name          string
	Driver        string
	ContainerPath string
	HostPath      string
	MountType     string
	MediaType     string
	IopsClass     int
	Size          int `json:",string"`
}

type Template struct {
	Id          bson.ObjectId "_id"
	Title       string
	Name        string
	Version     string
	Description string
	Services    []Service

	CreatorId   string `json:",omitempty"`
	CreatedTime int64  `json:",omitempty"`
	UpdaterId   string `json:",omitempty"`
	UpdaterName string `json:",omitempty"`
	UpdatedTime int64  `json:",omitempty"`
}

type Application struct {
	Id          bson.ObjectId "_id"
	TemplateId  string        `json:ApplicationTemplateId",omitempty"`
	PoolId      string        `json:",omitempty"`
	Title       string
	Name        string
	Version     string
	Description string
	Status      string

	Services []Service

	CreatorId   string `json:",omitempty"`
	CreatedTime int64  `json:",omitempty"`
	UpdaterId   string `json:",omitempty"`
	UpdaterName string `json:",omitempty"`
	UpdatedTime int64  `json:",omitempty"`
}

type DeploymentOpts map[string]interface{}

type Deployment struct {
	Id                 bson.ObjectId "_id"
	ApplicationId      string
	ApplicationVersion string
	OperationType      string
	Operator           string
	PoolId             string
	CreatorId          string
	CreatedTime        int64
	App                *Application
	Opts               DeploymentOpts
}

//copy自 consul/api/agent.go 避免引入consul/api及其依赖的库
type AgentService struct {
	ID                string
	Service           string
	Tags              []string
	Port              int
	Address           string
	EnableTagOverride bool
	CreateIndex       uint64
	ModifyIndex       uint64
}
