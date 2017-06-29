package types

import (
	"crypto/tls"
	"github.com/docker/go-connections/tlsconfig"
	"gopkg.in/mgo.v2/bson"

)


const (
	ROLESET_DEFAULT = 0

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
	DriverOpts     *DriverOpts
	Labels         map[string]interface{} `json:",omitempty"`
	ProxyEndpoints []string               `json:",omitempty"`
}

type Roleset uint64

type User struct {
	Id          bson.ObjectId "_id"
	Name        string
	Pass        string `json:",omitempty"`
	//Pass        string
	RoleSet     Roleset
	Email       string
	Tel         string
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
}


type TeamUser struct {
	UserId string
	TeamId string
}