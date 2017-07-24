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
	apiserver "github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
	"strconv"

	"gopkg.in/mgo.v2"
	//"io/ioutil"
)

var eventshandler = newEventsHandler()

const LABEL_CPUCOUNT = "com.zanecloud.omgea.container.cpus"
const LABEL_CPUEXCLUSIVE = "com.zanecloud.omega.container.exclusive"
const LABEL_COMPOSE_PROJECT = "com.docker.compose.project"
const LABEL_COMPOSE_SERVICE = "com.docker.compose.service"
const LABEL_APPLICATION_ID = "com.zanecloud.compose.application.id"

var routers = map[string]map[string]Handler{
	"HEAD": {},
	"GET": {
		"/containers/{name:.*}/attach/ws": proxyHijack,
		"/events":                         getEvents, //docker-1.11.1不需要，之后版本需要
	},
	"POST": {
		"/containers/create":                postContainersCreate,
		"/containers/{idorname:.*}/restart": proxyAsyncWithCallBack(restartContainer),
		//"/containers/{name:.*}/kill":   handlers.MgoSessionInject(proxyAsyncWithCallBack(updateContainer)),

		//	"/containers/{name:.*}/start":  handlers.MgoSessionInject(dockerClientInject(postContainersStart)),
		"/exec/{execid:.*}/start":      postExecStart,
		"/containers/{name:.*}/attach": proxyHijack,
	},
	"PUT": {},
	"DELETE": {
		"/containers/{name:.*}": proxyAsyncWithCallBack(deleteContainer),
	},
	"OPTIONS": {
		"": OptionsHandler,
	},
}

func deleteContainer(ctx context.Context, req *http.Request, resp *http.Response) {

	if err := req.ParseForm(); err != nil {
		logrus.Errorf("parse the request error:%s", err.Error())
		return
	}

	nameOrId := mux.Vars(req)["name"]

	logrus.Debugf("deleteContainer::status code is %d", resp.StatusCode)
	logrus.Debugf("deleteContainer::delete the container %s", nameOrId)
	//	logrus.Debugf("deleteContainer::req is %#v", req)

	//删除容器失败，则不需要做拦截
	if resp.StatusCode != http.StatusNoContent {
		return
	}
	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		logrus.Errorf("cant get mgo session")
		return
	}
	defer mgoSession.Close()

	poolInfo, err := getPoolInfo(ctx)
	if err != nil {
		logrus.Errorf("cant get pool info")
		return
	}

	mgoDB, err := getMgoDB(ctx)
	if err != nil {
		return
	}

	c := mgoSession.DB(mgoDB).C("container")

	//TODO or 删除
	if err := c.Remove(bson.M{"poolid": poolInfo.Id.Hex(), "containerid": nameOrId}); err != nil {
		if err == mgo.ErrNotFound {
			err = c.Remove(bson.M{"poolid": poolInfo.Id.Hex(), "name": nameOrId})
			if err != nil {
				logrus.Errorf("deleteContainer:: delete container from mgodb error:%s", err.Error())
			}
		}
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
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		tlsConfig = nil
	}

	if err := hijack(tlsConfig, poolInfo.DriverOpts.EndPoint, w, r); err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
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

type Handler func(c context.Context, w http.ResponseWriter, r *http.Request)

func NewPoolHandler(ctx context.Context, poolInfo * apiserver.PoolInfo ) (http.Handler, error) {


	r := mux.NewRouter()
	for method, mappings := range routers {
		for route, fct := range mappings {
			logrus.WithFields(logrus.Fields{"method": method, "route": route}).Info("Registering HTTP route in pool proxy")

			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				logrus.WithFields(logrus.Fields{"method": r.Method,
					"uri":                      r.RequestURI,
					"pool.Name":                poolInfo.Name,
					"pool.DriverOpts.Endpoint": poolInfo.DriverOpts.EndPoint}).Debug("HTTP request received in proxy")
				localFct(ctx, w, r)
			}
			localMethod := method

			r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			r.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

	// 作为swarm的代理，默认逻辑是所有请求都是转发给后端的swarm集群
	rootfunc := func(w http.ResponseWriter, req *http.Request) {
		logrus.WithFields(logrus.Fields{"method": req.Method,
			"uri":                      req.RequestURI,
			"pool.Name":                poolInfo.Name,
			"pool.DriverOpts.Endpoint": poolInfo.DriverOpts.EndPoint}).Debug("HTTP request received in proxy rootfunc")

		if err := proxyAsync(ctx, w, req, nil); err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
		}
	}
	r.PathPrefix("/v{version:[0-9.]+}" + "/").HandlerFunc(rootfunc)
	r.PathPrefix("/").HandlerFunc(rootfunc)

	return r, nil
}
//func preparePoolContext(p *Proxy, session *mgo.Session, cli *dockerclient.Client)  {
//	p.ctx = context.WithValue(p.ctx, utils.KEY_PROXY_SELF, p)
//	p.ctx = context.WithValue(p.ctx, utils.KEY_APISERVER_CONFIG, p.APIServerConfig)
//	p.ctx = context.WithValue(p.ctx, utils.KEY_MGO_SESSION, session)
//	p.ctx = context.WithValue(p.ctx, utils.KEY_POOL_CLIENT, cli)
//
//}

func proxyAsyncWithCallBack(callback func(context.Context, *http.Request, *http.Response)) Handler {

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

		f := func(resp *http.Response) {
			callback(ctx, r, resp)
		}

		if err := proxyAsync(ctx, w, r, f); err != nil {
			httpError(w, err.Error(), http.StatusInternalServerError)
		}

	}

}

func getMgoDB(ctx context.Context) (string, error) {
	config := utils.GetAPIServerConfig(ctx)
	return config.MgoDB, nil
}

func getPoolInfo(ctx context.Context) (*apiserver.PoolInfo, error) {
	p, ok := ctx.Value(utils.KEY_PROXY_SELF).(*Proxy)

	if !ok {
		logrus.Errorf("can't get proxy.self form ctx:%#v", ctx)
		return nil, errors.Errorf("can't get proxy.self form ctx:%#v", ctx)
	}

	return p.PoolInfo, nil
}

//
//func getDockerClient(ctx context.Context) (dockerclient.APIClient, error) {
//	client, ok := ctx.Value(utils.KEY_POOL_CLIENT).(dockerclient.APIClient)
//
//	if !ok {
//		logrus.Errorf("can't get pool.client from ctx:%#v", ctx)
//		return nil, errors.Errorf("can't get pool.client from ctx:%#v", ctx)
//	}
//
//	return client, nil
//}

func httpError(w http.ResponseWriter, err string, status int) {
	utils.HttpError(w, err, status)
}

type ContainerCreateConfig struct {
	container.Config
	HostConfig       container.HostConfig
	NetworkingConfig network.NetworkingConfig
}

// POST /exec/{execid:.*}/start
func postExecStart(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Connection") == "" {
		proxyAsync(ctx, w, r, nil)
	}
	proxyHijack(ctx, w, r)
}

//TODO
func validImage(_imageName string) error {
	return nil
}
func postContainersCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		config ContainerCreateConfig
		name   = r.Form.Get("name")
	)

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	logrus.Debug("postContainersCreate::check image valid")

	if err := validImage(config.Image); err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
	}

	poolInfo, err := getPoolInfo(ctx)
	if err != nil {
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//logrus.Debugf("postContainersCreate::before create a container , poolInfo is %#v  ", poolInfo)

	cli, _ := utils.GetPoolClient(ctx)

	resp, err := cli.ContainerCreate(ctx, &config.Config, &config.HostConfig, &config.NetworkingConfig, name)
	if err != nil {
		logrus.WithFields(logrus.Fields{"resp": resp, "err": err}).Debug("postContainersCreate:create container err")

		resp.ID = ""
		resp.Warnings = []string{}
		respBody, _ := json.Marshal(resp)

		//TODO imageNotFoundError 需要处理
		if strings.HasPrefix(err.Error(), "Conflict") {

			//httpError(w, "postContainersCreate:create container name conflict"+err.Error(), http.StatusConflict)
			httpError(w, string(respBody), http.StatusConflict)
			return
		} else {
			httpError(w, string(respBody), http.StatusInternalServerError)
			return
		}
	}

	logrus.WithFields(logrus.Fields{"resp": resp}).Debugf("create container success!")
	//TODO save to mongodb

	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mgoSession.Close()

	mgoDB, err := getMgoDB(ctx)
	if err != nil {
		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//这些信息是通过创建上下文的相对静态信息
	container := buildContainerInfoForSave(name, resp.ID, poolInfo, &config)

	//docker inspect 查容器的具体信息
	flushContainerInfo(ctx, container)

	if err := mgoSession.DB(mgoDB).C("container").Insert(container); err != nil {

		logrus.WithFields(logrus.Fields{"container": container, "error": err}).Debug("postContainersCreate::after insert container table error")

		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Id", resp.ID)

}

type NoOpResponseWriter struct {
}

func (w *NoOpResponseWriter) Header() http.Header {
	panic("no impl")
}

func (w *NoOpResponseWriter) Write([]byte) (int, error) {
	panic("no impl")
}

func (w *NoOpResponseWriter) WriteHeader(int) {
	panic("no impl")
}

func flushContainerInfo(ctx context.Context, container *Container) {

	//utils.GetMgoCollections(ctx, &NoOpResponseWriter{}, []string{"container"}, func(cs map[string]*mgo.Collection) {

	dockerclient, _ := utils.GetPoolClient(ctx)

	poolInfo, _ := getPoolInfo(ctx)

	containerJSON, err := dockerclient.ContainerInspect(ctx, container.ContainerId)
	if err != nil {
		logrus.Errorf("inspect the container:%d in the pool:(%s,%s) error:%s", container.ContainerId, poolInfo.Id, poolInfo.Name, err.Error())
		return
	}

	container.Name = containerJSON.Name
	container.Node = containerJSON.Node
	container.State = containerJSON.State

	if networkSettings, ok := containerJSON.NetworkSettings.Networks["bridge"]; ok {
		container.IP = networkSettings.IPAddress
	}

	logrus.WithFields(logrus.Fields{"container": containerJSON}).Debug("flush the container info")

	if applicationId, ok := containerJSON.Config.Labels[LABEL_APPLICATION_ID]; ok {
		container.ApplicationId = applicationId
	}

	//	colApplication, _ := cs["container"]
	//	if err := colApplication.UpdateId(container.Id, container); err != nil {
	//		logrus.Errorf("flushContainerInfo::save  container:%#v into db  error:%s", container, err.Error())
	//		return

	//	}

	//})

}
func OptionsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func buildContainerInfoForSave(name string, containerId string, poolInfo *apiserver.PoolInfo, config *ContainerCreateConfig) *Container {

	var cpuCount int64
	var exclusive bool
	var err error

	if lCpuCount, ok := config.Config.Labels[LABEL_CPUCOUNT]; ok {
		cpuCount, err = strconv.ParseInt(lCpuCount, 10, 64)
		if err != nil {
			cpuCount = 0
		}
	} else {
		cpuCount = 0
	}

	if lexclusive, ok := config.Config.Labels[LABEL_CPUEXCLUSIVE]; ok {
		exclusive, err = strconv.ParseBool(lexclusive)
		if err != nil {
			exclusive = false
		}
	} else {
		exclusive = false
	}

	c := &Container{
		Id:          bson.NewObjectId(),
		ContainerId: containerId,
		Name:        name,
		PoolId:      poolInfo.Id.Hex(),
		//PoolName:     poolInfo.Name,
		IsDeleted:    false,
		GmtCreated:   time.Now().Unix(),
		GmtDeleted:   0,
		Memory:       config.HostConfig.Memory,
		CPU:          cpuCount,
		CPUExclusive: exclusive,
		Status:       "running", // TODO 需要create/start多个钩子然后设置不同的状态

	}

	if service, ok := config.Config.Labels[LABEL_COMPOSE_SERVICE]; ok {
		c.Service = service
	}

	if project, ok := config.Config.Labels[LABEL_COMPOSE_PROJECT]; ok {
		c.Project = project

		// 我们的application对应compose的project
		c.ApplicationName = project
	}

	if applicationId, ok := config.Config.Labels[LABEL_APPLICATION_ID]; ok {
		c.ApplicationId = applicationId
	}

	return c
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

func newClientAndSchemeOR(poolInfo *apiserver.PoolInfo) (*http.Client, string, string, error) {
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

	//logrus.WithFields(logrus.Fields{"client": client, "scheme": scheme, "addr": addr}).Debug("proxyAsync: get the backend pool client info ")

	// RequestURI may not be sent to client
	r.RequestURI = ""

	r.URL.Scheme = scheme
	r.URL.Host = addr

	//logrus.WithFields(logrus.Fields{"method": r.Method, "url": r.URL, "uri": r.RequestURI}).Debug("Proxy request")
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
		httpError(w, err.Error(), 400)
		return
	}

	var until int64 = -1
	if r.Form.Get("until") != "" {
		u, err := strconv.ParseInt(r.Form.Get("until"), 10, 64)
		if err != nil {
			httpError(w, err.Error(), 400)
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
func restartContainer(ctx context.Context, req *http.Request, resp *http.Response) {

	idorname := mux.Vars(req)["idorname"]

	logrus.Debugf("restartContainer::status code is %d", resp.StatusCode)
	logrus.Debugf("restartContainer::restart the container %s", idorname)

	//restart容器失败，则不需要做拦截
	if resp.StatusCode != http.StatusNoContent {
		return
	}
	mgoSession, err := utils.GetMgoSessionClone(ctx)
	if err != nil {
		logrus.Errorf("cant get mgo session")
		return
	}
	defer mgoSession.Close()

	poolInfo, err := getPoolInfo(ctx)
	if err != nil {
		logrus.Errorf("cant get pool info")
		return
	}

	mgoDB, err := getMgoDB(ctx)
	if err != nil {
		return
	}

	c := mgoSession.DB(mgoDB).C("container")

	container := &Container{}

	selector1 := bson.M{"poolid": poolInfo.Id.Hex(), "containerid": idorname}
	selector2 := bson.M{"poolid": poolInfo.Id.Hex(), "name": idorname}

	if err := c.Find(bson.M{"$or": []bson.M{selector1, selector2}}).One(container); err != nil {

		logrus.WithFields(logrus.Fields{"idorname": idorname, "poolid": poolInfo.Id}).Error("nosuch a container")
		return
	}

	container.StartedTime = time.Now().Unix()

	//TODO or 删除
	if err := c.UpdateId(container.Id, container); err != nil {
		logrus.WithFields(logrus.Fields{"idorname": idorname, "poolid": poolInfo.Id}).Error("restart  the  container error")
		return
	}
}
