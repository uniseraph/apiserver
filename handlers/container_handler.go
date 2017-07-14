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

	"github.com/Sirupsen/logrus"
	"fmt"
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


	utils.GetMgoCollections(ctx,w,[]string{"application", "container"}, func(cs map[string]*mgo.Collection) {
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