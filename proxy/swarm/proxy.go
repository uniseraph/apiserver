package swarm

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/proxy"
	store "github.com/zanecloud/apiserver/types"
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
	server   *http.Server
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

	h, err := NewHandler(p)
	if err != nil {
		return err
	}
	p.server = &http.Server{
		Handler: h,
	}

	var paddr string

	logrus.Debugf("proxy.Start:: PoolInfo is %#v", p.PoolInfo)

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
