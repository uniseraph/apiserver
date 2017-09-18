package swarm

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/zanecloud/apiserver/proxy"
	store "github.com/zanecloud/apiserver/types"
	"gopkg.in/mgo.v2"
	"net"
	"net/http"
	"strings"
	"gopkg.in/mgo.v2/bson"
)




type Proxy struct {
	//PoolInfo        *store.PoolInfo
	PoolId          string
	APIServerConfig *store.APIServerConfig
	session *mgo.Session
	dockerClient *dockerclient.Client
	endpoint        string

	server          *http.Server
	ctx             context.Context
	cancel          context.CancelFunc
}

func init() {
	proxy.Register("swarm", NewProxy)
}

func NewProxy(config *store.APIServerConfig, pool *store.PoolInfo) (proxy.Proxy, error) {

	p := &Proxy{
		PoolId: pool.Id.Hex(),
		APIServerConfig: config,
		endpoint: pool.DriverOpts.EndPoint,
	}


	session, err := mgo.Dial(config.MgoURLs)
	if err != nil {
		return nil , err
	}
	session.SetMode(mgo.Monotonic, true)
	p.session = session



	var client *http.Client
	if pool.DriverOpts.TlsConfig != nil {
		tlsc, err := tlsconfig.Client(*pool.DriverOpts.TlsConfig)
		if err != nil {
			return nil, err
		}
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
			CheckRedirect: client.CheckRedirect,
		}
	}
	cli, err := dockerclient.NewClient(pool.DriverOpts.EndPoint, pool.DriverOpts.APIVersion, client, nil)
	if err != nil {
		session.Close()
		return nil ,err
	}

	p.dockerClient = cli

	ctx := context.WithValue(context.Background(), proxy.KEY_PROXY_SELF, p)
	ctx, cancel := context.WithCancel(ctx)

	p.ctx = ctx
	p.cancel = cancel

	return p, nil

}


func (p *Proxy) Start(opts *proxy.StartProxyOpts) error {


	h, err := NewPoolHandler(p.ctx, opts.PoolInfo)
	if err != nil {
		return err
	}
	p.server = &http.Server{
		Handler: h,
	}

	var paddr string



	if opts.PoolInfo.ProxyEndpoint != "" {
		//TODO 有可能apiserver换了一台机器重启，所以proxy的ip会发送变化,这种情况下也没必要保存端口不变
		//目前保持端口不变
		if parts := strings.SplitN(opts.PoolInfo.ProxyEndpoint, "://", 2); len(parts) == 2 {
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

			if err == http.ErrServerClosed {
				logrus.Infof("close the pool proxy server")
			} else {
				logrus.Warnf("proxy error:" + err.Error())
			}
		}
	}()

	go func() {
		//负责回收ctx中的资源
		select {
		case <-p.ctx.Done():
			if err := p.dockerClient.Close(); err != nil {
				logrus.Debugf("close the pool:%s  docker client , err:%s", p.PoolId, err.Error())
			} else {
				logrus.Debugf("close the pool:%s docker client success", p.PoolId)
			}

			logrus.Debugf("close the pool:%s mgo session", p.PoolId)
			p.session.Close()
			return
		}

	}()

	return nil
}

func (p *Proxy) Stop() error {

	defer p.cancel()

	if err := p.server.Close(); err != nil {
		logrus.WithFields(logrus.Fields{"err": err.Error()}).Errorf("swarm proxy::Stop close the proxy server error")
		return err
	}

	return nil

}

func (p *Proxy) Endpoint() string {
	return p.endpoint
}

func (p *Proxy) Pool() (*store.PoolInfo , error ) {


	session := p.session.Clone()
	defer session.Close()

	pool := &store.PoolInfo{}


	if err := session.DB(p.APIServerConfig.MgoDB).C("pool").FindId(bson.ObjectIdHex(p.PoolId)).One(pool) ; err !=nil {
		return nil , err
	}


	return pool,nil
}
