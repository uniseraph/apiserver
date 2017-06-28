package proxy

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/types"
)

var driver2FactoryFunc = make(map[string]FactoryFunc)

type StartProxyOpts struct {
}

type Proxy interface {
	Start(opts *StartProxyOpts) error
	Stop() error
	Pool() *types.PoolInfo
	Endpoint() string
}

type FactoryFunc func(c context.Context, p *types.PoolInfo) (Proxy, error)

func Register(driver string, ff FactoryFunc) {

	if _, ok := driver2FactoryFunc[driver]; ok {
		logrus.Warnf("ignore dup proxy driver %s , ", driver)
		return
	}
	driver2FactoryFunc[driver] = ff

}

func NewProxyInstaces(ctx context.Context, poolInfo *types.PoolInfo, n int) ([]Proxy, error) {

	ff, ok := driver2FactoryFunc[poolInfo.Driver]
	if !ok {
		logrus.Warnf("no such pool proxy driver %s , ", poolInfo.Driver)
		return nil, errors.Errorf("no such pool proxy driver %s", poolInfo.Driver)
	}

	//ff := driver2FactoryFunc[poolInfo.Driver]

	result := make([]Proxy, n)

	for i := 0; i < n; i++ {
		proxy, err := ff(ctx, poolInfo)
		if err != nil {
			logrus.Warnf("new proxy instance error :%s", err.Error())

			//for j:=0 ; j<i ; j++ {
			//	if errStop := result[j].Stop(&StopProxyOpts{}) ; errStop!=nil {
			//		logrus.Errorf("stop error the proxy %#v, error:%s" , result[j],errStop)
			//		result[j]=nil
			//	}
			//}
			return nil, err
		}
		result[i] = proxy
	}
	return result, nil

}

func NewProxyInstanceAndStart(ctx context.Context, poolInfo *types.PoolInfo) (Proxy, error) {

	ff, ok := driver2FactoryFunc[poolInfo.Driver]
	if !ok {
		logrus.Warnf("no such pool proxy driver %s , ", poolInfo.Driver)
		return nil, errors.Errorf("no such pool proxy driver %s", poolInfo.Driver)
	}
	//ff := driver2FactoryFunc[poolInfo.Driver]

	proxy, err := ff(ctx, poolInfo)
	if err != nil {
		return nil, err
	}

	if err := proxy.Start(&StartProxyOpts{}); err != nil {
		return nil, err
	}

	return proxy, err
}
