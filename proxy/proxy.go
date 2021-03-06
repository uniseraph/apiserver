package proxy

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/types"
	"sync"
)


const KEY_PROXY_SELF = "proxy.self"

var driver2FactoryFunc = make(map[string]FactoryFunc)

var id2Proxy = make(map[string]Proxy)

var mux sync.Mutex

type StartProxyOpts struct {
	PoolInfo *types.PoolInfo
}

type Proxy interface {
	Start(opts *StartProxyOpts) error
	Stop() error
	Pool() (*types.PoolInfo,error)
	Endpoint() string
}

type FactoryFunc func(config *types.APIServerConfig, p *types.PoolInfo) (Proxy, error)

func Register(driver string, ff FactoryFunc) {

	if _, ok := driver2FactoryFunc[driver]; ok {
		logrus.Warnf("ignore dup proxy driver %s , ", driver)
		return
	}
	driver2FactoryFunc[driver] = ff

}

func NewProxyInstanceAndStart(config *types.APIServerConfig, poolInfo *types.PoolInfo) (Proxy, error) {

	ff, ok := driver2FactoryFunc[poolInfo.Driver]
	if !ok {
		logrus.Warnf("no such pool proxy driver %s  ", poolInfo.Driver)
		return nil, errors.Errorf("no such pool proxy driver %s", poolInfo.Driver)
	}
	//ff := driver2FactoryFunc[poolInfo.Driver]

	proxy, err := ff(config, poolInfo)
	if err != nil {
		return nil, err
	}

	if err := proxy.Start(&StartProxyOpts{PoolInfo:poolInfo}); err != nil {
		return nil, err
	}


	mux.Lock()
	id2Proxy[poolInfo.Id.Hex()] = proxy
	mux.Unlock()
	return proxy, err
}

func Stop(id string) error {

	mux.Lock()

	p, ok := id2Proxy[id]
	if !ok {
		mux.Unlock()
		return errors.Errorf("no such a pool id:%s", id)
	}
	mux.Unlock()

	if err := p.Stop(); err != nil {
		return err
	}

	return nil
}


func Close(id string) error {

	mux.Lock()


	p , ok := id2Proxy[id]
	if !ok {
		mux.Unlock()
		return errors.Errorf( "no such a proxy:%s" ,id )
	}
	mux.Unlock()


	return p.Stop()



}
