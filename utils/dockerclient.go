package utils

//
//import (
//	"crypto/tls"
//	"sync"
//
//	"gitlab.alipay-inc.com/acs/dockerclient"
//)
//
//const (
//	KEY_ACS_CLUSTER_SCHEME            = "ACS_CLUSTER_SCHEME"
//	KEY_ACS_CLUSTER_ENDPOINT          = "ACS_CLUSTER_ENDPOINT"
//	KEY_ACS_CLUSTER_ENDPOINT_TLS      = "ACS_CLUSTER_ENDPOINT_TLS"
//	KEY_ACS_CLUSTER_ENDPOINT_TLS_KEY  = "ACS_CLUSTER_ENDPOINT_TLS_KEYFILE"
//	KEY_ACS_CLUSTER_ENDPOINT_TLS_CERT = "ACS_CLUSTER_ENDPOINT_TLS_CERTFILE"
//	KEY_ACS_CLUSTER_DISCOVERY         = "ACS_CLUSTER_DISCOVERY"
//)
//
//var client *dockerclient.DockerClient
//var once *sync.Once
//var initError error
//
//func init() {
//	once = &sync.Once{}
//}
//
//func InitDockerClient(scheme string, endpoint string, tlsConfig *tls.Config) (*dockerclient.DockerClient, error) {
//
//	once.Do(func() {
//		client, initError = dockerclient.NewDockerClient(
//			scheme+"://"+endpoint,
//			tlsConfig)
//	})
//
//	return client, initError
//}
