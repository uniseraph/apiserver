package handlers

import (
	"context"
	"net/http"

	"encoding/json"
	dockerclient "github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/proxy/swarm"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"github.com/Sirupsen/logrus"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/zanecloud/apiserver/types"
	"io"
	"strings"
	"time"
)

type ContainerListRequest struct {
	PageRequest
	ApplicationId string
	ServiceName   string
	PoolId        string
}

type ContainerListResponse struct {
	PageResponse
	Data []*swarm.Container
}

func getContainerJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	utils.GetMgoCollections(ctx, w, []string{"container", "pool"}, func(cs map[string]*mgo.Collection) {
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
	Tail string
	//Details    bool

}

func getContainerLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	req := &LogsContainerRequest{
		ShowStdout: false,
		ShowStderr: false,
		Since:      "0",
		Timestamps: false,
		Tail:       "100",
	}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"container", "pool"}, func(cs map[string]*mgo.Collection) {
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

		ioreader, err := dockerclient.ContainerLogs(r.Context(), container.ContainerId, dockertypes.ContainerLogsOptions{
			Follow:     false,
			ShowStderr: req.ShowStderr,
			ShowStdout: req.ShowStdout,
			Since:      req.Since,
			Timestamps: req.Timestamps,
			Tail:       req.Tail,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer ioreader.Close()

		c, err := dockerclient.ContainerInspect(r.Context(), container.ContainerId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		if c.Config.Tty {
			io.Copy(utils.NewWriteFlusher(w), ioreader)
		} else {
			_, err = stdcopy.StdCopy(utils.NewWriteFlusher(w), utils.NewWriteFlusher(w), ioreader)
		}

	})
}

func getContainerList(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &ContainerListRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		HttpError(w, err.Error(), http.StatusBadRequest)
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

	utils.GetMgoCollections(ctx, w, []string{"container", "pool"}, func(cs map[string]*mgo.Collection) {
		colContainer := cs["container"]
		colPool := cs["pool"]
		result := ContainerListResponse{
			Data: make([]*swarm.Container, 200),
		}

		selector := bson.M{}

		if req.ServiceName != "" {
			selector["service"] = req.ServiceName
		}
		if req.ApplicationId != "" {
			selector["applicationid"] = req.ApplicationId
		}

		if req.PoolId != "" {
			selector["poolid"] = req.PoolId
		}

		logrus.WithFields(logrus.Fields{"selector": selector}).Debug("getContainerList build a selector")

		n, err := colContainer.Find(selector).Count()

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

		if len(result.Data) != 0 {

			//不会有跨pool的container查询
			poolId := result.Data[0].PoolId

			poolInfo := &types.PoolInfo{}
			if err := colPool.FindId(bson.ObjectIdHex(poolId)).One(poolInfo); err != nil {
				if err == mgo.ErrNotFound {
					http.Error(w, "No such a pool:"+poolId, http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var client *http.Client
			if poolInfo.DriverOpts.TlsConfig != nil {
				tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
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
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cli.Close()

			for _, c := range result.Data {

				containerJSON, err := cli.ContainerInspect(ctx, c.ContainerId)

				if err != nil {
					logrus.Debugf("getContainerList inspect the container:%s error:%s", c.ContainerId[0:6], err.Error())
					continue
				}

				c.IP = containerJSON.NetworkSettings.IPAddress
				c.Status = containerJSON.State.Status
				c.State = containerJSON.State

				c.Ports = []*swarm.PortMapping{}
				for port, bindings := range containerJSON.NetworkSettings.Ports {

					if len(bindings) == 0 {
						continue
					}
					s := strings.Split(string(port), "/")
					if len(s) != 2 {
						continue
					}

					for i, _ := range bindings {
						pm := &swarm.PortMapping{
							ContainerPort: s[0],
							Proto:         s[1],
							HostPort:      bindings[i].HostPort,
							HostIp:        bindings[i].HostIP,
						}
						c.Ports = append(c.Ports, pm)

					}

				}
			}

		}
		result.Keyword = req.Keyword
		result.Page = req.Page
		result.PageSize = req.PageSize
		result.PageCount = result.Total / result.PageSize

		HttpOK(w, result)
	})

}

func restartContainer(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	id := mux.Vars(r)["id"]

	//config:=utils.GetAPIServerConfig(ctx)
	//
	//poolInfo := config.
	//cli, err := dockerclient.NewClient(poolInfo.DriverOpts.EndPoint, poolInfo.DriverOpts.APIVersion, client, nil)
	//if err != nil {
	//	return nil, err
	//}
	//defer cli.Close()

	utils.GetMgoCollections(ctx, w, []string{"container", "pool"}, func(cs map[string]*mgo.Collection) {

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

		var client *http.Client
		if poolInfo.DriverOpts.TlsConfig != nil {
			tlsc, err := tlsconfig.Client(*poolInfo.DriverOpts.TlsConfig)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cli.Close()

		timeout := time.Duration(30) * time.Second
		if err := cli.ContainerRestart(ctx, container.ContainerId, &timeout); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		HttpOK(w, "")

	})

}
