package swarm

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/proxy"
	store "github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"net"
	"net/http"
	"strings"
	"gopkg.in/mgo.v2"
)

type Proxy struct {
	PoolInfo        *store.PoolInfo
	APIServerConfig *store.APIServerConfig
	endpoint string
	server   *http.Server
	ctx      context.Context
}

func init() {
	proxy.Register("swarm", NewProxy)
}

func NewProxy(config *store.APIServerConfig, pool *store.PoolInfo) (proxy.Proxy, error) {

	proxy := &Proxy{
		PoolInfo: pool ,
		APIServerConfig: config,
	}

	proxy.ctx =  context.WithValue(context.Background(), utils.KEY_PROXY_SELF , proxy )


	return proxy , nil

}


func (p *Proxy) preparePoolContext(key2value map[string]interface{}) {

	for k , v := range key2value {
		p.ctx = context.WithValue(p.ctx,k ,v)
	}


}
func (p *Proxy) Start(opts *proxy.StartProxyOpts) error {

	var client *http.Client
	if p.PoolInfo.DriverOpts.TlsConfig != nil {
		tlsc, err := tlsconfig.Client(*p.PoolInfo.DriverOpts.TlsConfig)
		if err != nil {
			return  err
		}
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
			CheckRedirect: client.CheckRedirect,
		}
	}
	cli, err := dockerclient.NewClient(p.PoolInfo.DriverOpts.EndPoint, p.PoolInfo.DriverOpts.APIVersion, client, nil)
	if err != nil {
		return  err
	}


	session, err := mgo.Dial(p.APIServerConfig.MgoURLs)
	if err != nil {
		return  err
	}
	session.SetMode(mgo.Monotonic, true)


	p.preparePoolContext(  map[string]interface{} { utils.KEY_POOL_CLIENT : cli , utils.KEY_MGO_SESSION :session } )



	h, err := NewPoolHandler(p.ctx, p.PoolInfo)
	if err != nil {
		return err
	}
	p.server = &http.Server{
		Handler: h,
	}

	var paddr string


	if p.PoolInfo.ProxyEndpoint != "" {
		//TODO 有可能apiserver换了一台机器重启，所以proxy的ip会发送变化,这种情况下也没必要保存端口不变
		//目前保持端口不变
		if parts := strings.SplitN(p.PoolInfo.ProxyEndpoint, "://", 2); len(parts) == 2 {
			paddr = parts[1]
		} else {
			paddr = parts[0]
		}
	} else {
		paddr = fmt.Sprintf("%s:%d", p.APIServerConfig.Addr, 0)
	}

	logrus.Debugf("proxy.Start:: paddr is %s", paddr)

	listener, err := net.Listen("tcp4", paddr)
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	p.endpoint = listener.Addr().Network() + "://" + listener.Addr().String()

	go func() {
		if err := p.server.Serve(listener); err != nil {
			logrus.Fatal(err)
		}
	}()
	return nil
}

func (p *Proxy) Stop() error {

	if err := p.server.Close(); err != nil {
		logrus.WithFields(logrus.Fields{"err": err.Error()}).Errorf("close the proxy server error")
		return err
	}

	return nil

}

func (p *Proxy) Endpoint() string {
	return p.endpoint
}

func (p *Proxy) Pool() *store.PoolInfo {
	return p.PoolInfo
}
