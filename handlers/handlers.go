package handlers

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"

	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/utils"

	"gopkg.in/mgo.v2"
)

type Handler func(c context.Context, w http.ResponseWriter, r *http.Request)

func NewHandler(ctx context.Context, rs map[string]map[string]Handler) http.Handler {

	r := mux.NewRouter()
	SetupPrimaryRouter(r, ctx, rs)
	return r

}

func SetupPrimaryRouter(r *mux.Router, ctx context.Context, rs map[string]map[string]Handler) {
	for method, mappings := range rs {
		for route, fct := range mappings {
			log.WithFields(log.Fields{"method": method, "route": route}).Debug("Registering HTTP route")

			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				log.WithFields(log.Fields{"method": r.Method, "uri": r.RequestURI}).Debug("HTTP request received")
				localFct(ctx, w, r)
			}
			localMethod := method

			r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			r.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}
}

//func getContainersJSON(c *Context, w http.ResponseWriter, r *http.Request) {
//
//	all := BoolValue(r, "all")
//	size := BoolValue(r, "size")
//	filters := r.Form.Get("filters")
//
//	log.WithField("filters", filters).Debug("container ps")
//
//	//var result []dockerclient.Container
//
//	result := []dockerclient.Container{}
//	client, err := utils.InitDockerClient(
//		c.ClusterScheme,
//		c.ClusterEndpoint,
//		c.TlsConfig,
//	)
//
//	if err != nil {
//		httpError(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	containers, err := client.ListContainers(all, size, url.QueryEscape(filters))
//
//	if err != nil {
//		httpError(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	for _, c := range containers {
//		_, exists := c.Labels[SYSTEM_LABEL]
//		if exists {
//			log.Info("filter the container " + c.Id + "  with acs label")
//		} else {
//			result = append(result, c)
//		}
//	}
//
//	json.NewEncoder(w).Encode(result)
//}

//func selectTarget(scheme , endpoint string) (string){
//	if scheme=="swarm" {
//
//		addrs := strings.Split(endpoint, ",")
//
//		index := rand.Int31n((int32(len(addrs))))
//
//		return addrs[index]
//
//	}else{
//		return endpoint
//	}
//}

// inspect container采用swarm的实现方式，避免dockerclient.APIVersion对版本出现影响
// 也防止返回数据在反序列化和序列化之后丢失信息
//func getContainerJSON(c *Context, w http.ResponseWriter, r *http.Request) {
//	name := mux.Vars(r)["name"]
//
//	client, scheme := newClientAndScheme(c.TlsConfig)
//
//	var resp *http.Response
//	var err error
//
//	index , addrs := selectTargetIndex(c.ClusterScheme,c.ClusterEndpoint)
//
//	for i:=index ; i<index+len(addrs) ;i++ {
//
//		resp, err = client.Get(scheme + "://" +
//			addrs[i%len(addrs)] + "/containers/" + name + "/json")
//
////		log.Debug("connect the node " + addrs[i%len(addrs)])
//
//		if err != nil {
//		//	resp.Body.Close()
//		//	closeIdleConnections(client)
//			continue
//		}
//		break
//	}
//
//	if err!=nil {
//		httpError(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//
//	defer resp.Body.Close()
//	defer closeIdleConnections(client)
//
//	data, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		httpError(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	if resp.StatusCode != http.StatusOK {
//		httpError(w, string(data), resp.StatusCode)
//		return
//	}
//
//	var info *dockerclient.ContainerInfo
//	if err := json.Unmarshal(data, &info); err != nil {
//		httpError(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	if _, exists := info.Config.Labels[SYSTEM_LABEL]; exists {
//		log.Infof("inspect a container with ACS System Label, ignore this request")
//		httpError(w, fmt.Sprintf("No such container %s", name), http.StatusNotFound)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	w.Write(data)
//}

func HttpError(w http.ResponseWriter, err string, status int) {
	log.WithField("status", status).Errorf("HTTP error: %v", err)
	http.Error(w, err, status)
}

func BoolValue(r *http.Request, k string) bool {
	s := strings.ToLower(strings.TrimSpace(r.FormValue(k)))
	return !(s == "" || s == "0" || s == "no" || s == "false" || s == "none")
}

// prevents leak with https
func closeIdleConnections(client *http.Client) {
	if tr, ok := client.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}

// POST /exec/{execid:.*}/start
//func postExecStart(c *Context, w http.ResponseWriter, r *http.Request) {
//	if r.Header.Get("Connection") == "" {
//		proxyContainer(c, w, r)
//	}
//	proxyHijack(c, w, r)
//}

// Proxy a hijack request to the right node
//func proxyHijack(c *Context, w http.ResponseWriter, r *http.Request) {
//
//	log.WithFields(log.Fields{ "url":r.URL}).Debug("enter proxyHijack")
//
//	var scheme string
//	if c.TlsConfig==nil{
//		scheme = "http"
//	}else{
//		scheme = "https"
//	}
//	//
//	//if c.ClusterScheme =="swarm" {
//	//	index , addrs := selectTargetIndex(c.ClusterScheme, c.ClusterEndpoint)
//	//	for i:=index ; i<index+len(addrs) ;i++ {
//	//		node :=  addrs[ i%len(addrs) ]
//	//		if err :=  hijack(c.TlsConfig, scheme +"://"+ node , w, r); err != nil {
//	//			log.WithFields(log.Fields{"node":node}).Info("conect the swarm node fail")
//	//			continue
//	//		}
//	//		log.WithFields(log.Fields{"node":node}).Debug("conect the swarm node success")
//	//
//	//		return
//	//	}
//	//	httpError(w,  "ALL nodes cann't be connected : " + c.ClusterEndpoint,http.StatusInternalServerError)
//	//}else {
//
//		if err := hijack(c.TlsConfig, scheme + "://" + c.ClusterEndpoint, w, r); err != nil {
//			httpError(w, err.Error(), http.StatusInternalServerError)
//		}
//
//	//}
//
//
//
//}

func hijack(tlsConfig *tls.Config, addr string, w http.ResponseWriter, r *http.Request) error {
	if parts := strings.SplitN(addr, "://", 2); len(parts) == 2 {
		addr = parts[1]
	}

	log.WithField("addr", addr).Debug("Proxy hijack request")

	var (
		d   net.Conn
		err error
	)

	if tlsConfig != nil {
		d, err = tls.Dial("tcp", addr, tlsConfig)
	} else {
		d, err = net.Dial("tcp", addr)
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

// GET /_ping
//func ping(c *Context, w http.ResponseWriter, r *http.Request) {
//	w.Write([]byte{'O', 'K'})
//}

// Proxy a request to the right node
//func proxyContainer(c *Context, w http.ResponseWriter, r *http.Request) {
//
//	//log.WithFields(log.Fields{"scheme": c.ClusterScheme, "endpoint": c.ClusterEndpoint}).Debug("enter proxyContainer")
//
//	if err := proxy(c.TlsConfig,c.ClusterScheme, c.ClusterEndpoint, w, r); err != nil {
//		httpError(w, err.Error(), http.StatusInternalServerError)
//	}
//}

//func proxy(tlsConfig *tls.Config, scheme string ,  endpoint string, w http.ResponseWriter, r *http.Request) error {
//	log.WithFields(log.Fields{"endpoint": endpoint , "url":r.URL}).Debug("enter proxy")
//
//	body :=  r.Body
//	defer body.Close()
//
//	r.Body = ioutil.NopCloser(r.Body)
//
//	//if scheme=="swarm" {
//	//	index , addrs := selectTargetIndex(scheme, endpoint)
//	//	for i:=index ; i<index+len(addrs) ;i++ {
//	//		//log.WithFields(log.Fields{"endpoint":  addrs[i%len(addrs)]}).Debug("connect the swarm node")
//	//		node := addrs[i%len(addrs)]
//	//		if err := proxyAsync(tlsConfig , node ,w,r,nil) ; err==nil{
//	//			log.WithFields(log.Fields{"node":node , "url":r.URL}).Info("exec swarm api success")
//	//			return nil;
//	//		}else{
//	//			log.WithFields(log.Fields{"node":node ,
//	//										"url":r.URL ,
//	//											"err":err.Error()}).Info("exec swarm api fail")
//	//	}}
//	//	return errors.New( "ALL nodes cann't be connected : " + endpoint)
//	//}else{
//		return proxyAsync(tlsConfig , endpoint ,w,r,nil)
//	//}
//}

//func newClientAndScheme(tlsConfig *tls.Config) (*http.Client, string) {
//	if tlsConfig != nil {
//		return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}, "https"
//	}
//	return &http.Client{}, "http"
//}
//func proxyAsync(tlsConfig *tls.Config, endpoint string, w http.ResponseWriter, r *http.Request, callback func(*http.Response)) error {
//	// Use a new client for each request
//	client, scheme := newClientAndScheme(tlsConfig)
//
//	// RequestURI may not be sent to client
//	r.RequestURI = ""
//
//	r.URL.Scheme = scheme
//	r.URL.Host = endpoint
//
//	//log.WithFields(log.Fields{"method": r.Method, "url": r.URL}).Debug("Proxy request")
//	resp, err := client.Do(r)
//	if err != nil {
//		return err
//	}
//
//	utils.CopyHeader(w.Header(), resp.Header)
//	w.WriteHeader(resp.StatusCode)
//	io.Copy(utils.NewWriteFlusher(w), resp.Body)
//
//	if callback != nil {
//		callback(resp)
//	}
//
//	// cleanup
//	resp.Body.Close()
//	closeIdleConnections(client)
//
//	return nil
//}

// Default handler for methods not supported by clustering.
func notImplementedHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	HttpError(w, "Not supported in clustering mode.", http.StatusNotImplemented)
}

func OptionsHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

//func postContainersExecWrap(c *Context , w http.ResponseWriter  , r *http.Request){
//	defer r.Body.Close()
//
//	var  err error
//	var  retry bool
//        if err,retry = postContainersExec(c , w , mux.Vars(r)["name"], ioutil.NopCloser(  r.Body)) ; err!=nil && retry==true{
//			 httpError(w, err.Error(),http.StatusInternalServerError)
//		 }
//
//
//
//}

// POST /containers/{name:.*}/exec
//func postContainersExec(c *Context, w http.ResponseWriter, name string , body io.ReadCloser) ( error , bool) {
//	client, scheme:= newClientAndScheme(c.TlsConfig)
//
//	resp, err := client.Post(scheme +"://"+ c.ClusterEndpoint +"/containers/"+name+"/exec",
//		"application/json", body)
//	if err != nil {
//		log.WithFields( log.Fields{ "node":c.ClusterEndpoint,
//			"err":err.Error()}).Debug("post swarm fail")
//		return err  , true
//	}
//
//	// cleanup
//	defer resp.Body.Close()
//	defer closeIdleConnections(client)
//
//
//	if resp.StatusCode == http.StatusNotFound {
//		body, err := ioutil.ReadAll(resp.Body)
//		if err != nil {
//			log.WithFields(log.Fields{}).Info("read resp body error")
//			return err  , false
//		}
//		http.Error(w, string(body) , http.StatusNotFound)
//		return errors.New(string(body)) , false
//
//	}
//
//
//	// check status code
//	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
//		body, err := ioutil.ReadAll(resp.Body)
//		if err != nil {
//			log.WithFields(log.Fields{}).Info("read resp body error")
//			http.Error(w, "read resp body error" , resp.StatusCode)
//
//			return err , false
//		}
//		http.Error(w, string(body) , resp.StatusCode)
//
//		return errors.New(string(body)) , false
//	}
//
//	data, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		http.Error(w, "read resp body error" , resp.StatusCode)
//
//		return err , false
//	}
//
//	id := struct{ ID string }{}
//
//	if err := json.Unmarshal(data, &id); err != nil {
//		http.Error(w, "Unmarshal data error" , resp.StatusCode)
//
//		return err ,false
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	w.WriteHeader(resp.StatusCode)
//	w.Write(data)
//
//	return nil , false
//}

func MgoSessionAware(h Handler) Handler {

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

		mgoURLS, ok := ctx.Value(utils.KEY_MGO_URLS).(string)
		if !ok {
			// context 里面没有 mongourl，这是不应该的
			log.Errorf("no mogodburl in ctx , ctx is #%v", ctx)
			HttpError(w, "no mogodburl in ctx", http.StatusInternalServerError)
			return
		}

		session, err := mgo.Dial(mgoURLS)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() {
			log.Debug("close mgo session")
			session.Close()
		}()

		session.SetMode(mgo.Monotonic, true)

		c := context.WithValue(ctx, utils.KEY_MGO_SESSION, session)

		log.Debugf("ctx is %#v", c)

		h(c, w, r)

	}
}
