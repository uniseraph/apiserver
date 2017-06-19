package swarm

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/store"
	"github.com/zanecloud/apiserver/utils"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"strconv"
	"gopkg.in/mgo.v2/bson"

	"gopkg.in/mgo.v2"
	//"io/ioutil"
)

var eventshandler = newEventsHandler()

const LABEL_CPUCOUNT     ="com.zanecloud.omgea.container.cpus"
const LABEL_CPUEXCLUSIVE ="com.zanecloud.omega.container.exclusive"
var routers = map[string]map[string]handlers.Handler{
	"HEAD": {},
	"GET": {
		"/containers/{name:.*}/attach/ws": proxyHijack,
		"/events":                         getEvents,  //docker-1.11.1不需要，之后版本需要
	},
	"POST": {
		"/containers/create":           handlers.MgoSessionInject(postContainersCreate),
		//"/containers/{name:.*}/kill":   handlers.MgoSessionInject(proxyAsyncWithCallBack(updateContainer)),

		//	"/containers/{name:.*}/start":  handlers.MgoSessionInject(dockerClientInject(postContainersStart)),
		"/exec/{execid:.*}/start":      postExecStart,
		"/containers/{name:.*}/attach": proxyHijack,
	},
	"PUT":    {},
	"DELETE": {
		"/containers/{name:.*}":    handlers.MgoSessionInject(proxyAsyncWithCallBack(deleteContainer)),
	},
	"OPTIONS": {
		"": handlers.OptionsHandler,
	},
}

func deleteContainer(ctx context.Context, req *http.Request, resp *http.Response) {

	if err := req.ParseForm(); err != nil {
		logrus.Errorf("parse the request error:%s",err.Error())
		return
	}


	nameOrId := mux.Vars(req)["name"]


	logrus.Debugf("deleteContainer::status code is %d" , resp.StatusCode)
	logrus.Debugf("deleteContainer::update the container %s",nameOrId)
	logrus.Debugf("deleteContainer::req is %#v" , req)

	//删除容器失败，则不需要做拦截
	if(resp.StatusCode != http.StatusNoContent){
		return
	}
	mgoSession, err :=utils.GetMgoSession(ctx)
	if err!=nil{
		logrus.Errorf("cant get mgo session")
		return
	}


	poolInfo , err := getPoolInfo(ctx)
	if err!=nil{
		logrus.Errorf("cant get pool info")
		return
	}

	mgoDB, err := getMgoDB(ctx)
	if err != nil {
		return
	}

	c := mgoSession.DB(mgoDB).C("container")

	if err := c.Remove(bson.M{ "poolname": poolInfo.Name , "id":nameOrId }) ; err!=nil{
		if err == mgo.ErrNotFound {
			err = c.Remove(bson.M{ "poolname": poolInfo.Name , "name":nameOrId })
			if err != nil {
				logrus.Errorf("deleteContainer:: delete container from mgodb error:%s",err.Error())
			}
		}
	}

}


func dockerClientInject(h handlers.Handler) handlers.Handler {

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		poolInfo, err := getPoolInfo(ctx)
		if err != nil {
			handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logrus.Debugf("before inject a dockerclient , poolInfo is %#v  ", poolInfo)

		var client *http.Client
		if poolInfo.DriverOpts.TlsConfig != nil {

			tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
			if err != nil {
				handlers.HttpError(w, err.Error(), http.StatusBadRequest)
				return
			}

			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: tlsc,
				},
				CheckRedirect: client.CheckRedirect,
			}
		}

		cli, err := dockerclient.NewClient(poolInfo.DriverOpts.EndPoint, poolInfo.DriverOpts.APIVersion, client, nil)
		defer cli.Close()
		if err != nil {
			handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logrus.Debugf("ctx is %#v", ctx)
		h(context.WithValue(ctx,utils.KEY_POOL_CLIENT , cli) , w , r)
	}
}


// Proxy a hijack request to the right node
func proxyHijack(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	logrus.WithFields(logrus.Fields{"url": r.URL}).Debug("enter proxyHijack")

	poolInfo, _ := getPoolInfo(ctx)

	var tlsConfig *tls.Config
	var err error

	if poolInfo.DriverOpts.TlsConfig != nil {
		tlsConfig, err = tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
		if err != nil {
			handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		tlsConfig = nil
	}

	if err := hijack(tlsConfig, poolInfo.DriverOpts.EndPoint, w, r); err != nil {
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// endpoint :  tcp://127.0.0.1:2375
//             localhost:2375
//             unix:///var/run/docker.sock
func hijack(tlsConfig *tls.Config, endpoint string, w http.ResponseWriter, r *http.Request) error {

	var proto, addr string
	if parts := strings.SplitN(endpoint, "://", 2); len(parts) == 2 {
		proto, addr = parts[0], parts[1]
	} else if len(parts) == 1 {
		proto, addr = "tcp", parts[0]
	}

	logrus.WithFields(logrus.Fields{"proto": proto, "addr": addr}).Debug("Proxy hijack request")

	var (
		d   net.Conn
		err error
	)

	if tlsConfig != nil {
		d, err = tls.Dial("tcp", addr, tlsConfig)
	} else {
		if proto == "unix" {
			d, err = net.Dial("unix", addr)
		} else {
			d, err = net.Dial("tcp", addr)
		}
	}
	if err != nil {
		return err
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		return err
	}
	nc, _, err := hj.Hijack()
	if err != nil {
		return err
	}
	defer nc.Close()
	defer d.Close()

	err = r.Write(d)
	if err != nil {
		return err
	}

	cp := func(dst io.Writer, src io.Reader, chDone chan struct{}) {
		io.Copy(dst, src)
		if conn, ok := dst.(interface {
			CloseWrite() error
		}); ok {
			conn.CloseWrite()
		}
		close(chDone)
	}
	inDone := make(chan struct{})
	outDone := make(chan struct{})
	go cp(d, nc, inDone)
	go cp(nc, d, outDone)

	// 1. When stdin is done, wait for stdout always
	// 2. When stdout is done, close the stream and wait for stdin to finish
	//
	// On 2, stdin copy should return immediately now since the out stream is closed.
	// Note that we probably don't actually even need to wait here.
	//
	// If we don't close the stream when stdout is done, in some cases stdin will hange
	select {
	case <-inDone:
		// wait for out to be done
		<-outDone
	case <-outDone:
		// close the conn and wait for stdin
		nc.Close()
		<-inDone
	}
	return nil
}

func NewHandler(ctx context.Context) http.Handler {

	r := mux.NewRouter()

	handlers.SetupPrimaryRouter(r, ctx, routers)

	poolInfo, _ := getPoolInfo(ctx)

	// 作为swarm的代理，默认逻辑是所有请求都是转发给后端的swarm集群
	rootwrap := func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{"method": r.Method, "uri": r.RequestURI, "pool backend endpoint": poolInfo.DriverOpts.EndPoint}).Debug("HTTP request received in rootwrap")
		if err := proxyAsync(ctx, w, r, nil); err != nil {
			handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		}
	}
	r.Path("/v{version:[0-9.]+}" + "/").HandlerFunc(rootwrap)

	r.PathPrefix("/").HandlerFunc(rootwrap)

	return r
}


func proxyAsyncWithCallBack(callback func(context.Context,  *http.Request , *http.Response)) handlers.Handler {


	return func(ctx context.Context, w http.ResponseWriter, r *http.Request){

			f := func(resp *http.Response) {
				callback(ctx,r , resp)
			}


			if err:=proxyAsync(ctx,w,r,f); err!=nil {
				handlers.HttpError(w,err.Error(),http.StatusInternalServerError)
			}

	}


}

func getMgoDB(ctx context.Context) (string, error) {
	config := utils.GetAPIServerConfig(ctx)
	return config.MgoDB, nil
}


func getPoolInfo(ctx context.Context) (*store.PoolInfo, error) {
	p, ok := ctx.Value(utils.KEY_PROXY_SELF).(*Proxy)

	if !ok {
		logrus.Errorf("can't get proxy.self form ctx:%#v", ctx)
		return nil, errors.Errorf("can't get proxy.self form ctx:%#v", ctx)
	}

	return p.PoolInfo, nil
}


func getDockerClient (ctx context.Context) (dockerclient.APIClient, error) {
	client, ok := ctx.Value(utils.KEY_POOL_CLIENT).(dockerclient.APIClient)

	if !ok {
		logrus.Errorf("can't get pool.client from ctx:%#v", ctx)
		return nil, errors.Errorf("can't get pool.client from ctx:%#v", ctx)
	}

	return client, nil
}


type ContainerCreateConfig struct {
	container.Config
	HostConfig       container.HostConfig
	NetworkingConfig network.NetworkingConfig
}

// POST /exec/{execid:.*}/start
func postExecStart(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Connection") == "" {
		proxyAsync(ctx, w, r ,nil)
	}
	proxyHijack(ctx, w, r)
}

//TODO
func validImage(_imageName string) error {
	return nil
}
func postContainersCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		handlers.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		config ContainerCreateConfig
		name   = r.Form.Get("name")
	)

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		handlers.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	logrus.Debug("check image valid")

	if err:= validImage(config.Image) ; err!=nil {
		handlers.HttpError(w, err.Error(),http.StatusInternalServerError)
	}


	poolInfo, err := getPoolInfo(ctx)
	if err != nil {
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.Debugf("before create a container , poolInfo is %#v  ", poolInfo)

	var client *http.Client
	if poolInfo.DriverOpts.TlsConfig != nil {

		tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
		if err != nil {
			handlers.HttpError(w, err.Error(), http.StatusBadRequest)
			return
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
			CheckRedirect: client.CheckRedirect,
		}
	}

	cli, err := dockerclient.NewClient(poolInfo.DriverOpts.EndPoint, poolInfo.DriverOpts.APIVersion, client, nil)
	defer cli.Close()
	if err != nil {
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := cli.ContainerCreate(ctx, &config.Config, &config.HostConfig, &config.NetworkingConfig, name)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Conflict") {
			handlers.HttpError(w, err.Error(), http.StatusConflict)
			return
		} else {
			handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//TODO save to mongodb

	mgoSession, err := utils.GetMgoSession(ctx)
	if err != nil {

		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB, err := getMgoDB(ctx)
	if err != nil {
		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := mgoSession.DB(mgoDB).C("container").Insert(buildContainerInfoForSave(name,resp.ID,poolInfo,&config)); err != nil {

		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Id", resp.ID)

}

func  buildContainerInfoForSave(name string, id string , poolInfo *store.PoolInfo,config *ContainerCreateConfig ) (*Container) {

	var cpuCount int64
	var exclusive bool
	var err error
	if lCpuCount, ok := config.Config.Labels[LABEL_CPUCOUNT]; ok {
		cpuCount  , err =strconv.ParseInt( lCpuCount , 10 , 64)
		if err!=nil {
			cpuCount = 0
		}
	}else{
		cpuCount =0
	}

	if lexclusive, ok := config.Config.Labels[LABEL_CPUEXCLUSIVE] ; ok{
		exclusive , err = strconv.ParseBool(lexclusive)
		if err !=nil {
			exclusive =false
		}
	} else{
		exclusive = false
	}

	return &Container{
		Id : id ,
		Name : name,
		PoolName: poolInfo.Name,
		IsDeleted:  false,
		GmtCreated: time.Now().Unix(),
		GmtDeleted: 0 ,
		Memory : config.HostConfig.Memory ,
		CPU: cpuCount ,
		CPUExclusive: exclusive ,
	}
}


// POST /containers/{name:.*}/start
//func postContainersStart(ctx context.Context, w http.ResponseWriter, r *http.Request) {
//
//	cli, err := getDockerClient(ctx)
//	if err!=nil {
//		handlers.HttpError(w, err.Error(),http.StatusInternalServerError)
//		return
//	}
//
//	name := mux.Vars(r)["name"]
//
//	if err := cli.ContainerStart(ctx, name, types.ContainerStartOptions{}) ; err!=nil{
//		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	w.WriteHeader(http.StatusNoContent)
//}

//func newClientAndScheme(tlsConfig *tls.Config) (*http.Client, string) {
//	if tlsConfig != nil {
//		return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}, "https"
//	}
//	return &http.Client{}, "http"
//}

func newClientAndSchemeOR(poolInfo *store.PoolInfo) (*http.Client, string, string, error) {
	protoAddrParts := strings.SplitN(poolInfo.DriverOpts.EndPoint, "://", 2)

	var proto, addr string

	if len(protoAddrParts) == 2 {
		proto = protoAddrParts[0]
		addr = protoAddrParts[1]
	} else if len(protoAddrParts) == 1 {
		proto = "tcp"
		addr = protoAddrParts[0]
	}
	if proto == "tcp" {
		parsed, err := url.Parse("tcp://" + addr)
		if err != nil {
			return nil, "", "", err
		}
		addr = parsed.Host
		//basePath = parsed.Path
	}

	transport := new(http.Transport)
	sockets.ConfigureTransport(transport, proto, addr)

	if poolInfo.DriverOpts.TlsConfig != nil {
		tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
		if err != nil {
			return nil, "", "", err
		}

		transport.TLSClientConfig = tlsc

		return &http.Client{
			Transport:     transport,
			CheckRedirect: dockerclient.CheckRedirect,
		}, "https", addr, nil
	} else {

		return &http.Client{
			Transport:     transport,
			CheckRedirect: dockerclient.CheckRedirect,
		}, "http", addr, nil
	}
}

func proxyAsync(ctx context.Context, w http.ResponseWriter, r *http.Request, callback func(*http.Response)) error {
	// Use a new client for each request

	poolInfo, _ := getPoolInfo(ctx)

	//TODO using backend tlsconfig
	client, scheme, addr, err := newClientAndSchemeOR(poolInfo)
	if err != nil {
		//handlers.HttpError(w,err.Error(),http.StatusInternalServerError)
		return err
	}

	logrus.WithFields(logrus.Fields{"client": client, "scheme": scheme, "addr": addr}).Debug("proxyAsync: get the backend pool client info ")

	// RequestURI may not be sent to client
	r.RequestURI = ""

	r.URL.Scheme = scheme
	r.URL.Host = addr

	logrus.WithFields(logrus.Fields{"method": r.Method, "url": r.URL, "uri":r.RequestURI}).Debug("Proxy request")
	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	//data, err := ioutil.ReadAll(resp.Body)

	//logrus.WithFields(logrus.Fields{"resp.body":string(data)}).Debug("proxyAysnc : receive a response")


	utils.CopyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(utils.NewWriteFlusher(w), resp.Body)




	if callback != nil {
		callback(resp)
	}

	// cleanup
	resp.Body.Close()
	closeIdleConnections(client)

	return nil
}

// prevents leak with https
func closeIdleConnections(client *http.Client) {
	if tr, ok := client.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}


func getEvents(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		handlers.HttpError(w, err.Error(), 400)
		return
	}

	var until int64 = -1
	if r.Form.Get("until") != "" {
		u, err := strconv.ParseInt(r.Form.Get("until"), 10, 64)
		if err != nil {
			handlers.HttpError(w, err.Error(), 400)
			return
		}
		until = u
	}

	eventshandler.Add(r.RemoteAddr, w)

	w.Header().Set("Content-Type", "application/json")

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	eventshandler.Wait(r.RemoteAddr, until)
}
