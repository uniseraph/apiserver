package swarm

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/proxy"
	"github.com/zanecloud/apiserver/store"
	"github.com/zanecloud/apiserver/utils"
	"net"
	"net/http"
	"strings"
)

type Proxy struct {
	PoolInfo        *store.PoolInfo
	APIServerConfig *store.APIServerConfig
	//mgoDB    string
	//mgoURLs  string
	endpoint string
}

func init() {
	proxy.Register("swarm", NewProxy)
}

func NewProxy(ctx context.Context, pool *store.PoolInfo) (proxy.Proxy, error) {
	//
	//mgoDB, nil := getMgoDB(ctx)
	//mgoURLs, nil := getMgoURLs(ctx)

	return &Proxy{
		PoolInfo:        pool,
		APIServerConfig: utils.GetAPIServerConfig(ctx),
	}, nil
}

func (p *Proxy) Start(opts *proxy.StartProxyOpts) error {

	ctx := setContext(p)

	h, err := NewHandler(ctx)
	if err != nil {
		return err
	}
	server := http.Server{
		Handler: h,
	}

	var paddr string

	logrus.Debugf("proxy.Start:: PoolInfo is %#v", p.PoolInfo)

	if p.PoolInfo.ProxyEndpoints[0] != "" {
		//有可能apiserver换了一台机器重启，所以proxy的ip会发送变化,这种情况下也没必要保存端口不变
		//TODO
		if parts := strings.SplitN(p.PoolInfo.ProxyEndpoints[0], "://", 2); len(parts) == 2 {
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
		if err := server.Serve(listener); err != nil {
			logrus.Fatal(err)
		}
	}()
	return nil
}
func setContext(p *Proxy) context.Context {
	ctx := context.WithValue(context.Background(), utils.KEY_PROXY_SELF, p)
	logrus.Debugf("proxy %s's context is %#v", p.Pool().Name, ctx)
	//c1 := context.WithValue(ctx, utils.KEY_APISERVER_CONFIG, p.APIServerConfig)
	c1 := utils.PutAPIServerConfig(ctx, p.APIServerConfig)
	logrus.Debugf("proxy %s's context is %#v", p.Pool().Name, c1)

	return c1
}

func (p *Proxy) Stop() error {

	//TODO
	return nil
}

func (p *Proxy) Endpoint() string {
	return p.endpoint
}

func (p *Proxy) Pool() *store.PoolInfo {
	return p.PoolInfo
}
