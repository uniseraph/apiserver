package proxy



type StartProxyOpts struct {

}

type StopProxyOpts struct {

}

type Proxy interface {
	Start(opts  *StartProxyOpts) error
	Stop(opts *StopProxyOpts) error
}