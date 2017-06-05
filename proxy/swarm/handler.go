package swarm

import (
	"context"
	"net/http"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"encoding/json"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/zanecloud/apiserver/store"
	"github.com/docker/docker/api/types/network"
	"fmt"
	"github.com/docker/docker/api/types"
	"strings"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/utils"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"time"
	"io"
	"net/url"
	"github.com/docker/go-connections/sockets"
)

var routers = map[string]map[string]handlers.Handler{
	"HEAD": {},
	"GET": {
	},
	"POST": {
		"/containers/create":           	handlers.MgoSessionAware(postContainersCreate),
		"/containers/{name:.*}/start":          handlers.MgoSessionAware(postContainersStart),

	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": handlers.OptionsHandler,
	},
}



func NewHandler(ctx context.Context) http.Handler {


	r := mux.NewRouter()

	handlers.SetupPrimaryRouter(r,ctx,routers)

	poolInfo , _ := getPoolInfo(ctx)

	// 作为swarm的代理，默认逻辑是所有请求都是转发给后端的swarm集群
	rootwrap := func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{"method": r.Method, "uri": r.RequestURI , "pool backend endpoint":poolInfo.DriverOpts.EndPoint  }).Debug("HTTP request received in proxy")
		if err := proxyAsync(ctx,w,r,nil) ; err!=nil {
			handlers.HttpError(w, err.Error(), http.StatusInternalServerError)

		}
	}
	r.Path("/v{version:[0-9.]+}" + "/").HandlerFunc(rootwrap)

	r.PathPrefix("/").HandlerFunc(rootwrap)


	return r
}



func getMgoSession(ctx context.Context) (*mgo.Session, error){
	mgoSession  , ok := ctx.Value(utils.KEY_MGO_SESSION).(*mgo.Session)

	if !ok {
		logrus.Errorf("can't get mgo.session  form ctx:%#v" , ctx)
		return nil , errors.Errorf("can't get mgo.session form ctx:%#v" , ctx)
	}

	return mgoSession,nil
}
func getMgoDB(ctx context.Context) (string, error){
	mgoDB , ok := ctx.Value(utils.KEY_MGO_DB).(string)

	if !ok {
		logrus.Errorf("can't get mgo.db form ctx:%#v" , ctx)
		return "" , errors.Errorf("can't get mgo.db form ctx:%#v" , ctx)
	}

	return mgoDB,nil
}

func getMgoURLs(ctx context.Context) (string, error){
	mgoURLs , ok := ctx.Value(utils.KEY_MGO_URLS).(string)

	if !ok {
		logrus.Errorf("can't get mgo.urls form ctx:%#v" , ctx)
		return "" , errors.Errorf("can't get mgo.urls form ctx:%#v" , ctx)
	}

	return mgoURLs,nil
}


func getPoolInfo(ctx context.Context) (*store.PoolInfo, error){
	p , ok := ctx.Value(utils.KEY_PROXY_SELF).(*Proxy)

	if !ok {
		logrus.Errorf("can't get proxy.self form ctx:%#v" , ctx)
		return nil , errors.Errorf("can't get proxy.self form ctx:%#v" , ctx)
	}

	return p.PoolInfo,nil
}


type ContainerCreateConfig struct {
	container.Config
	HostConfig container.HostConfig
	NetworkingConfig network.NetworkingConfig
}
func postContainersCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		handlers.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}


	var (
		config  ContainerCreateConfig
		name                    = r.Form.Get("name")
	)

	if  err := json.NewDecoder(r.Body).Decode(&config); err!=nil {
		handlers.HttpError(w , err.Error() , http.StatusBadRequest)
		return
	}




	poolInfo , err := getPoolInfo(ctx)
	if err!=nil {
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logrus.Debugf("before create a container , poolInfo is %#v  ", poolInfo)

	var client *http.Client
	if poolInfo.DriverOpts.TlsConfig!=nil  {

		tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
		if err != nil {
			handlers.HttpError(w , err.Error() , http.StatusBadRequest)
			return
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
			CheckRedirect: client.CheckRedirect,
		}
	}

	cli , err := dockerclient.NewClient(poolInfo.DriverOpts.EndPoint , poolInfo.DriverOpts.APIVersion , client , nil)
	defer cli.Close()
	if err!= nil {
		handlers.HttpError(w , err.Error() , http.StatusInternalServerError)
		return
	}


	resp , err := cli.ContainerCreate(ctx, &config.Config , &config.HostConfig , &config.NetworkingConfig, name)
	if err!= nil {
		if strings.HasPrefix(err.Error(), "Conflict") {
			handlers.HttpError(w, err.Error(), http.StatusConflict)
			return
		} else {
			handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}


	//TODO save to mongodb

	mgoSession , err := getMgoSession(ctx)
	if err!=nil {

		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx , resp.ID , types.ContainerRemoveOptions{Force:true})
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mgoDB , err := getMgoDB(ctx)
	if err !=nil {
		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx , resp.ID , types.ContainerRemoveOptions{Force:true})
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := mgoSession.DB(mgoDB).C("container").Insert(&Container{Id:resp.ID ,
		IsDeleted:false,
		GmtCreated: time.Now().Unix(),
		GmtDeleted: 0}) ; err!=nil{

		//TODO 如果清理容器失败，需要记录一下日志，便于人工干预
		cli.ContainerRemove(ctx , resp.ID , types.ContainerRemoveOptions{Force:true})
		handlers.HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}



	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Id", resp.ID)



}


// POST /containers/{name:.*}/start
func postContainersStart(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	cli ,ok := ctx.Value("dockerclient").(dockerclient.APIClient)
	if !ok {
		handlers.HttpError(w,"cant't find target pool", http.StatusInternalServerError)
		return
	}


	name := mux.Vars(r)["name"]

	err := cli.ContainerStart(ctx,name , types.ContainerStartOptions{})

	if err !=nil{
		handlers.HttpError(w, err.Error(),http.StatusInternalServerError)
	}


	w.WriteHeader(http.StatusNoContent)
}


//func newClientAndScheme(tlsConfig *tls.Config) (*http.Client, string) {
//	if tlsConfig != nil {
//		return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}, "https"
//	}
//	return &http.Client{}, "http"
//}

func newClinetAndSchemeOR(poolInfo *store.PoolInfo) (*http.Client , string ,string, error) {
	protoAddrParts := strings.SplitN(poolInfo.DriverOpts.EndPoint, "://", 2)

	var proto, addr  string

	if len(protoAddrParts) ==2 {
		proto = protoAddrParts[0]
		addr = protoAddrParts[1]
	}else if len(protoAddrParts) ==1 {
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

	if poolInfo.DriverOpts.TlsConfig!=nil{
		tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
		if err!=nil{
			return nil,"","",err
		}

		transport.TLSClientConfig = tlsc

		return &http.Client{
			Transport:     transport,
			CheckRedirect: dockerclient.CheckRedirect,
		}  , "https" , addr, nil
	}else {


		return &http.Client{
			Transport:     transport,
			CheckRedirect: dockerclient.CheckRedirect,
		}  , "http" ,addr , nil
	}
}


func proxyAsync( ctx context.Context , w http.ResponseWriter, r *http.Request, callback func(*http.Response)) error {
	// Use a new client for each request

	poolInfo , _ := getPoolInfo(ctx)



	//TODO using backend tlsconfig
	client, scheme ,addr , err := newClinetAndSchemeOR(poolInfo)
	if err!=nil {
		//handlers.HttpError(w,err.Error(),http.StatusInternalServerError)
		return err
	}

	logrus.WithFields(logrus.Fields{"client":client,"scheme":scheme,"addr":addr}).Debug("get the backend pool info ")

	// RequestURI may not be sent to client
	r.RequestURI = ""

	r.URL.Scheme = scheme
	r.URL.Host =   addr

	//log.WithFields(log.Fields{"method": r.Method, "url": r.URL}).Debug("Proxy request")
	resp, err := client.Do(r)
	if err != nil {
		return err
	}

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
