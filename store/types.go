package store

import (
	"github.com/docker/go-connections/tlsconfig"
	"gopkg.in/mgo.v2/bson"
)

type DriverOpts struct {
	Name       string
	Version    string
	EndPoint   string
	APIVersion string
	Labels     []string               `json:",omitempty"`
	TlsConfig  *tlsconfig.Options     `json:",omitempty"`
	Opts       map[string]interface{} `json:",omitempty"`
}

type PoolInfo struct {
	Id bson.ObjectId "_id"

	Name   string
	Status string

	Driver     string
	DriverOpts *DriverOpts
	Labels     []string `json:",omitempty"`
	Endpoints  []string `json:",omitempty"`
}
