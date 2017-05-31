package store

import (
	"github.com/docker/go-connections/tlsconfig"
)


type DriverOpts struct {
	Name       string
	Version    string
	EndPoint   string
	APIVersion string
	Labels     []string            `json:",omitempty"`
	TlsConfig  *tlsconfig.Options  `json:",omitempty"`
	Opts       map[string]interface{}  `json:",omitempty"`
}


type PoolInfo struct {
	ID        string
	Name      string
	Status        string

	Driver        string
	DriverOpts   *DriverOpts
	Labels        []string            `json:",omitempty"`

}