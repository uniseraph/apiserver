package store

import (
	"github.com/docker/go-connections/tlsconfig"
	"gopkg.in/mgo.v2/bson"
	"crypto/tls"
)

type APIServerConfig struct {
	MgoDB   string
	MgoURLs string
	Addr    string
	Port    int
	tlsConfig *tls.Config
}

type DriverOpts struct {
	Version    string
	EndPoint   string
	APIVersion string
	Labels     map[string]interface{}               `json:",omitempty"`
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
	ProxyEndpoints []string `json:",omitempty"`
}
