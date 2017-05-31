package cli

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/zanecloud/apiserver/handlers"
	"context"
	"github.com/zanecloud/apiserver/utils"
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


	//tlsConfig, err := getTlsConfig(c)
	//if err != nil {
	//	logrus.Fatal(err)
	//}

	c1 := setContext(c)

	server := http.Server{
		Handler: handlers.NewMainHandler(c1),
	}

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		logrus.Fatal(err)
		return
	}

	if err := server.Serve(listener); err != nil {
		logrus.Fatal(err)
	}
}
func setContext(c *cli.Context) context.Context {
	mgoURLS := c.String(utils.KEY_MGO_URLS)
	mgoDB := c.String(utils.KEY_MGO_DB)
	//ctx := context.WithValue(context.Background(),"tlsConfig",tlsConfig)
	ctx := context.WithValue(context.Background(), utils.KEY_MGO_URLS, mgoURLS)
	c1 := context.WithValue(ctx, utils.KEY_MGO_DB, mgoDB)
	return c1
}
