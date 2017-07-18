package cli

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"context"
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/proxy"
	_ "github.com/zanecloud/apiserver/proxy/swarm"
	store "github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const startCommandName = "start"

var (
	clientCipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}

	clientDefault = tls.Config{
		MinVersion:         tls.VersionTLS12,
		CipherSuites:       clientCipherSuites,
		InsecureSkipVerify: true,
	}
)

func getTlsConfig(c *cli.Context) (*tls.Config, error) {
	if !c.Bool("tls") {
		return nil, nil
	}

	keyFile := c.String("tlskey")
	certFile := c.String("tlscert")

	tlsConfig := clientDefault
	if certFile != "" && keyFile != "" {
		tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("Could not load X509 key pair: %v. Make sure the key is not encrypted", err)
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
	}
	return &tlsConfig, nil
}

func startCommand(c *cli.Context) {

	config := parserAPIServerConfig(c)

	ctx := utils.PutAPIServerConfig(context.Background(), config)
	h, err := handlers.NewMainHandler(ctx)
	if err != nil {
		logrus.Fatal(err)
		return
	}
	server := http.Server{
		Handler: h,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Addr, config.Port))
	if err != nil {
		logrus.Fatal(err)
		return
	}

	go startProxys(ctx)

	if err := server.Serve(listener); err != nil {
		logrus.Fatal(err)
	}
}

func parserAPIServerConfig(c *cli.Context) *store.APIServerConfig {

	return &store.APIServerConfig{
		MgoDB:     c.String(utils.KEY_MGO_DB),
		MgoURLs:   c.String(utils.KEY_MGO_URLS),
		RedisAddr: c.String(utils.KEY_REDIS_ADDR),
		Addr:      c.String(utils.KEY_LISTENER_ADDR),
		Port:      c.Int(utils.KEY_LISTENER_PORT),
		RootDir:   c.String(utils.KEY_ROOT_DIR),
	}

}

func startProxys(ctx context.Context) {

	config := utils.GetAPIServerConfig(ctx)
	session, err := mgo.Dial(config.MgoURLs)
	if err != nil {
		logrus.Errorf("startProxys::dial mongodb %s  error: %s", config.MgoURLs, err.Error())
		return
	}

	logrus.Debug("startProxys::start a mgosession")

	defer func() {
		logrus.Debug("startProxys::close mgo session")
		session.Close()
	}()

	session.SetMode(mgo.Monotonic, true)

	var pools []store.PoolInfo
	if err := session.DB(config.MgoDB).C("pool").Find(bson.M{}).All(&pools); err != nil {
		logrus.Errorf("startProxys::get all pool error : %", err.Error())
		return
	}

	for i, _ := range pools {
		logrus.Debugf("startProxys:: start the pool:%s", pools[i].Name)

		proxy, err := proxy.NewProxyInstanceAndStart(ctx, &pools[i])
		if err != nil {
			logrus.Errorf("startProxys:: startProxy error:%s", err.Error())
		}

		pools[i].ProxyEndpoint = proxy.Endpoint()

	}

}
