package swarm

import (
	"github.com/zanecloud/apiserver/proxy"
	"net/http"
	"context"
	"net"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/store"
	"github.com/zanecloud/apiserver/utils"
)

type Proxy struct {
	PoolInfo *store.PoolInfo
	mgoDB    string
	mgoURLs  string
	endpoint string
}


func init() {
	proxy.Register("swarm" , NewProxy)
}


func NewProxy(ctx context.Context , pool *store.PoolInfo) (proxy.Proxy , error){

	mgoDB , nil := getMgoDB(ctx)
	mgoURLs, nil := getMgoURLs(ctx)

	return &Proxy{
		PoolInfo: pool,
		mgoDB: mgoDB,
		mgoURLs : mgoURLs,
	} , nil
}


func (p *Proxy) Start( opts *proxy.StartProxyOpts) error {


	ctx := setContext(p)



	server := http.Server{
		Handler: NewHandler(ctx),
	}

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		logrus.Fatal(err)
		return err
	}

	p.endpoint = listener.Addr().Network() + "://" + listener.Addr().String()

	go func () {
		if err := server.Serve(listener); err != nil {
			logrus.Fatal(err)
		}
	}()
	return nil
}
func setContext(p *Proxy) context.Context{
	ctx := context.WithValue(context.Background(), utils.KEY_PROXY_SELF, p)
	logrus.Debugf("proxy %s's context is %#v", p.Pool().Name, ctx)
	c1 := context.WithValue(ctx, utils.KEY_MGO_URLS, p.mgoURLs)
	logrus.Debugf("proxy %s's context is %#v", p.Pool().Name, c1)
	c2 := context.WithValue(c1, utils.KEY_MGO_DB, p.mgoDB)
	logrus.Debugf("proxy %s's context is %#v", p.Pool().Name, c2)
	return  c2
}




func (p *Proxy) Stop() error {
	return nil
}


func (p *Proxy) Endpoint() string {
	return p.endpoint
}


func (p *Proxy) Pool() *store.PoolInfo {
	return p.PoolInfo
}