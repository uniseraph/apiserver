package handlers

import (
	"context"
	"encoding/json"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type VDomainCreateRequest struct {
	poolId         string
	vdomain        string
	lb             string
	proto          string `json:type",string"`
	uri            string
	valid_statuses string
	interval       int
	timeout        int
	fall           int
	rise           int
	concurrency    int
}

var client = &http.Client{}

//Req：/zlb/domains/{domainName}
// curl -X POST --data '{"type":"http","uri":"/check","valid_statuses":"200,302","interval":2000,
// "timeout":1000,"fall":3,"rise":2,"concurrency":10}'  "http://127.0.0.1:6300/zlb/domains/b.com" -v
func createVDomain(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &VDomainCreateRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.GetMgoCollections(ctx, w, []string{"pool"}, func(cs map[string]*mgo.Collection) {

		colPool, _ := cs["pool"]

		pool := types.PoolInfo{}
		if err := colPool.Find(bson.M{"_id": bson.ObjectIdHex(req.poolId)}).One(&pool); err != nil {

			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})

}

func inspectVDomain(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func getVDomainList(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}

func deleteVDomain(ctx context.Context, w http.ResponseWriter, r *http.Request) {

}
