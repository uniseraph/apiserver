package handlers

import (
	"context"
	"net/http"

	"github.com/zanecloud/apiserver/proxy/swarm"
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	dockerclient "github.com/docker/docker/client"

	"github.com/Sirupsen/logrus"
	"fmt"
	"github.com/zanecloud/apiserver/types"
	"github.com/docker/go-connections/tlsconfig"
	"time"
	dockertypes "github.com/docker/docker/api/types"
	"io"
)





type ContainerListRequest struct {
	PageRequest
	ApplicationId string
	ServiceName string
	PoolId string
}

type ContainerListResponse struct {
	PageResponse
	Data []*swarm.Container

}




func getContainerJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	utils.GetMgoCollections(ctx,w,[]string{"container","pool"}, func(cs map[string]*mgo.Collection) {
		container := &swarm.Container{}

		colContainer, _ := cs["container"]
		colPool, _ := cs["pool"]

		if err := colContainer.FindId(bson.ObjectIdHex(id)).One(container); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "No such a container", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		poolInfo := &types.PoolInfo{}
		if err := colPool.FindId(bson.ObjectIdHex(container.PoolId)).One(poolInfo); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "No such a pool:"+container.PoolId, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dockerclient, err := utils.CreateDockerClient(poolInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		containerJSON, err := dockerclient.ContainerInspect(r.Context(), container.ContainerId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, containerJSON)
	})

}



type LogsContainerRequest struct {

	ShowStdout bool
	ShowStderr bool
	Since      string
	Timestamps bool
	//Follow     bool
	Tail       string
	//Details    bool

}
func getContainerLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	req := &LogsContainerRequest{
		ShowStdout: false,
		ShowStderr: false,
		Since: "0",
		Timestamps: false,
		Tail : "100",
	}

	if err:= json.NewDecoder(r.Body).Decode(req) ; err != nil {
		HttpError(w , err.Error() , http.StatusBadRequest)
		return
	}


	utils.GetMgoCollections(ctx,w,[]string{"container","pool"}, func(cs map[string]*mgo.Collection) {
		container := &swarm.Container{}

		colContainer, _ := cs["container"]
		colPool, _ := cs["pool"]

		if err := colContainer.FindId(bson.ObjectIdHex(id)).One(container); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "No such a container", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		poolInfo := &types.PoolInfo{}
		if err := colPool.FindId(bson.ObjectIdHex(container.PoolId)).One(poolInfo); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "No such a pool:"+container.PoolId, http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dockerclient, err := utils.CreateDockerClient(poolInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dockerclient.Close()

		ioreader , err := dockerclient.ContainerLogs(r.Context(), container.ContainerId , dockertypes.ContainerLogsOptions{
			Follow: false ,
			ShowStderr: req.ShowStderr,
			ShowStdout: req.ShowStdout ,
			Since: req.Since ,
			Timestamps: req.Timestamps ,
			Tail: req.Tail,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer ioreader.Close()

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.Copy(utils.NewWriteFlusher(w),ioreader)



	})
}

func getContainerList(ctx context.Context, w http.ResponseWriter, r *http.Request) {


	req := &ContainerListRequest{}

	if err := json.NewDecoder(r.Body).Decode(req) ; err !=nil {
		HttpError(w, err.Error() , http.StatusBadRequest)
		return
	}

	applicationId := mux.Vars(r)["id"]

	if applicationId != "" {
		req.ApplicationId = applicationId
	}


	if req.ServiceName == "" {
		HttpError(w, "ServiceName 不能为空", http.StatusBadRequest)
		return
	}

	if req.Page == 0 {
		HttpError(w, "从第一页开始", http.StatusBadRequest)
		return
	}

	if req.PageSize == 0 {
		req.PageSize = 20
	}


	utils.GetMgoCollections(ctx,w,[]string{ "container"}, func(cs map[string]*mgo.Collection) {
		colContainer := cs["container"]

		result := ContainerListResponse{
			Data: make([]*swarm.Container,200),
		}

		selector:=bson.M{}

		if req.ServiceName!="" {
			selector["service"] = req.ServiceName
		}
		if req.ApplicationId!=""{
			selector["applicationid"] = req.ApplicationId
		}

		if req.PoolId !="" {
			selector["poolid"] = req.PoolId
		}


		logrus.WithFields(logrus.Fields{"selector":selector}).Debug("getContainerList build a selector")



		n, err :=  colContainer.Find(selector).Count()

		if err != nil {
			HttpError(w, fmt.Sprintf("查询记录数出错，%s", err.Error()), http.StatusInternalServerError)
			return
		}

		result.Total = n

		logrus.Debugf("getContainerList::符合条件的container有%d个", result.Total)

		if err := colContainer.Find(selector).Sort("title").Limit(req.PageSize).Skip(req.PageSize * (req.Page - 1)).All(&result.Data); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result.Keyword = req.Keyword
		result.Page = req.Page
		result.PageSize = req.PageSize
		result.PageCount = result.Total / result.PageSize

		HttpOK(w,result)
	})

}


func restartContainer(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id:= mux.Vars(r)["id"]

	//config:=utils.GetAPIServerConfig(ctx)
	//
	//poolInfo := config.
	//cli, err := dockerclient.NewClient(poolInfo.DriverOpts.EndPoint, poolInfo.DriverOpts.APIVersion, client, nil)
	//if err != nil {
	//	return nil, err
	//}
	//defer cli.Close()
	
	
	utils.GetMgoCollections(ctx,w,[]string{"container","pool"}, func(cs map[string]*mgo.Collection) {

		container := & swarm.Container{}

		colContainer , _ :=  cs["container"]
		colPool,_ := cs["pool"]

		if err := colContainer.FindId(bson.ObjectIdHex(id)).One(container); err != nil{
			if err == mgo.ErrNotFound {
				http.Error(w, "No such a container" , http.StatusNotFound)
				return
			}
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}

		poolInfo := &types.PoolInfo{}
		if err:= colPool.FindId(bson.ObjectIdHex(container.PoolId)).One(poolInfo) ; err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "No such a pool:"+container.PoolId , http.StatusNotFound)
				return
			}
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}

		var client *http.Client
		if poolInfo.DriverOpts.TlsConfig != nil {
			tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: tlsc,
				},
				CheckRedirect: client.CheckRedirect,
			}
		}

		cli , err := dockerclient.NewClient(poolInfo.DriverOpts.EndPoint,poolInfo.DriverOpts.APIVersion,client,nil)
		if err !=nil {
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}
		defer  cli.Close()

		timeout :=time.Duration(30)*time.Second
		if err := cli.ContainerRestart(ctx,container.ContainerId,&timeout) ; err!=nil{
			http.Error(w,err.Error(),http.StatusInternalServerError)
			return
		}

		HttpOK(w,"")

	})


}