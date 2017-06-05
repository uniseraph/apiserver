package cli

import (
	"github.com/codegangsta/cli"

	"github.com/zanecloud/apiserver/utils"
)

var (
	flMgoUrls = cli.StringFlag{
		Name:   utils.KEY_MGO_URLS,
		Value:  "localhost",
		EnvVar: "MGO_URLS",
		Usage:  "mongodb urls",
	}

	flMgoDB = cli.StringFlag{
		Name:   utils.KEY_MGO_DB,
		Value:  "zanecloud",
		EnvVar: "MGO_DB",
		Usage:  "mongodb database",
	}

	//flClusterEndpoint = cli.StringFlag{
	//	Name:   "clusterEndpoint",
	//	Value:  "localhost:2375",
	//	EnvVar: utils.KEY_ACS_CLUSTER_ENDPOINT,
	//	Usage:  "cluster endpoint",
	//}
	//
	//flClusterScheme = cli.StringFlag{
	//	Name:   "clusterScheme",
	//	Value:  "http",
	//	EnvVar: utils.KEY_ACS_CLUSTER_SCHEME,
	//	Usage:  "cluster scheme",
	//}
	//
	//flClusterTls = cli.BoolFlag{
	//	Name:   "tls",
	//	EnvVar: utils.KEY_ACS_CLUSTER_ENDPOINT_TLS_CERT,
	//	Usage:  "use TLS to connect to swarm/docker",
	//}
	//
	//flClusterTlsKeyFile = cli.StringFlag{
	//	Name:   "tlskey",
	//	EnvVar: utils.KEY_ACS_CLUSTER_ENDPOINT_TLS_KEY,
	//	Usage:  "path to TLS key file",
	//}
	//
	//flClusterTlsCertFile = cli.StringFlag{
	//	Name:   "tlscert",
	//	EnvVar: utils.KEY_ACS_CLUSTER_ENDPOINT_TLS_CERT,
	//	Usage:  "path to TLS cert file",
	//}

	//	flApiUrl = cli.StringFlag{
	//		Name: "apiUrl",
	//		Value: "http://192.168.99.100:80",
	//		EnvVar: utils.KEY_ACS_API_URL,
	//		Usage: "api url",
	//	}
	//
	//
	//	flBackendUrl = cli.StringFlag{
	//		Name: "backendUrl",
	//		Value: "http://10.210.182.53:2375",
	//		EnvVar: utils.KEY_ACS_BACKEND_URL,
	//		Usage: "backend server url",
	//	}
)
