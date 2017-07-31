package proxy

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/types"
	"sync"
	"fmt"
)

var driver2FactoryFunc = make(map[string]FactoryFunc)

var id2Proxy = make(map[string]Proxy)

var mux sync.Mutex

type StartProxyOpts struct {
}

type Proxy interface {
	Start(opts *StartProxyOpts) error
	Stop() error
	Pool() *types.PoolInfo
	Endpoint() string
}

type FactoryFunc func( config *types.APIServerConfig  , p *types.PoolInfo) (Proxy, error)

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

	if err := proxy.Start(&StartProxyOpts{}); err != nil {
		return nil, err
	}


	mux.Lock()
	id2Proxy[proxy.Pool().Id.Hex()] = proxy
	mux.Unlock()

	return proxy, err
}


func Close(id string) error {

	mux.Lock()


	p , ok := id2Proxy[id]
	if !ok {
		mux.Unlock()
		return errors.New( fmt.Sprintf("no such a proxy:%s" ,id ))
	}
	mux.Unlock()


	return p.Stop()



}
