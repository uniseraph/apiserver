package handlers

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
)

var swarmProxyRoutes = map[string]map[string]Handler{
	"HEAD": {},
	"GET": {
	},
	"POST": {
		"/containers/create":           	postContainersCreate,
		"/containers/{name:.*}/start":          postContainersStart,

	},
	"PUT":    {},
	"DELETE": {},
	"OPTIONS": {
		"": optionsHandler,
	},
}



func NewSwarmProxyHandler(ctx context.Context) http.Handler {
	return newHandler(ctx , swarmProxyRoutes)
}


type ContainerCreateConfig struct {
	container.Config
	HostConfig container.HostConfig
	NetworkingConfig network.NetworkingConfig
}
func postContainersCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		httpError(w, err.Error(), http.StatusBadRequest)
		return
	}


	var (
		config  ContainerCreateConfig
		name                    = r.Form.Get("name")
	)

	if  err := json.NewDecoder(r.Body).Decode(&config); err!=nil {
		httpError(w , err.Error() , http.StatusBadRequest)
		return
	}


	poolID  , ok := config.Labels[POOL_LABEL];
	if !ok {
		httpError(w , "pool label is empty" , http.StatusBadRequest)
		return
	}


	if poolID == "localhost" {
		//TODO select pool from mongodb
	}

	pool := &store.PoolInfo{
		Driver: 	"swarm",
		DriverOpts: 	&store.DriverOpts{
			Name:       "swarm",
			Version:    "1.23",
			EndPoint:   "unix:///var/run/docker.sock",
			APIVersion: "1.0",
			Labels:     []string {},
			TlsConfig : nil,
			Opts:       make(map[string]interface{}) ,
		} ,
		Labels: 	[]string{},

	}


	var client *http.Client
	if pool.DriverOpts.TlsConfig!=nil  {

		tlsc, err := tlsconfig.Client(*pool.DriverOpts.TlsConfig)
		if err != nil {
			httpError(w , err.Error() , http.StatusBadRequest)
			return
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
			CheckRedirect: client.CheckRedirect,
		}
	}

	cli , err := dockerclient.NewClient(pool.DriverOpts.EndPoint , pool.DriverOpts.APIVersion , client , nil)
	defer cli.Close()
	if err!= nil {
		httpError(w , err.Error() , http.StatusInternalServerError)
		return
	}


	resp , err := cli.ContainerCreate(ctx, &config.Config , &config.HostConfig , &config.NetworkingConfig, name)
	if err!= nil {
		if strings.HasPrefix(err.Error(), "Conflict") {
			httpError(w, err.Error(), http.StatusConflict)
			return
		} else {
			httpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	//TODO save to mongodb



	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "{%q:%q}", "Id", resp.ID)


	cli.Close()
}


// POST /containers/{name:.*}/start
func postContainersStart(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	cli ,ok := ctx.Value("dockerclient").(dockerclient.APIClient)
	if !ok {
		httpError(w,"cant't find target pool", http.StatusInternalServerError)
		return
	}


	name := mux.Vars(r)["name"]

	err := cli.ContainerStart(ctx,name , types.ContainerStartOptions{})

	if err !=nil{
		httpError(w, err.Error(),http.StatusInternalServerError)
	}


	w.WriteHeader(http.StatusNoContent)
}
