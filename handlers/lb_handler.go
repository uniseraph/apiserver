package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/zanecloud/apiserver/types"
	"github.com/zanecloud/apiserver/utils"
	"github.com/zanecloud/zlb/api/daemon"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"strings"
)

type VDomainRequestHead struct {
	PoolId string
	LbType string
}

type VDomainSetCookieFilterRequest struct {
	VDomainRequestHead
	daemon.CookieFilter
}

const COOKIE_FILTER_ON = 1
const COOKIE_FILTER_OFF = 0

//curl -i -X POST -d '{"PoolId":"59c07d76421aa92b96679283"}' http://localhost:8080/api/lb/domains/b.com/set-cookie-filter
func setVDomainCookieFilter(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &VDomainSetCookieFilterRequest{}
	req.Lifecycle = COOKIE_FILTER_ON

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	VDomain := mux.Vars(r)["domain"]

	if req.Name == "" {
		req.Name = "ZANE_GRAY_TAG"
	}

	if req.Value == "" {
		req.Value = "coupon"
	}

	if req.LbType == "" {
		req.LbType = "zlb"
	}

	utils.GetMgoCollections(ctx, w, []string{"pool"}, func(cs map[string]*mgo.Collection) {
		colPool, _ := cs["pool"]
		pool := types.PoolInfo{}

		if err := colPool.FindId(bson.ObjectIdHex(req.PoolId)).One(&pool); err != nil {
			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
				HttpError(w, "没有这样的pool "+err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, "查询pool失败:"+err.Error(), http.StatusInternalServerError)
			return
		}

		if req.LbType != "zlb" {
			HttpError(w, "no such lbtype:"+req.LbType, http.StatusBadRequest)
			return
		}

		postURL := fmt.Sprintf("http://%s:6300/zlb/domains/%s/setCookieFilter", pool.TunneldAddr, VDomain)

		logrus.Debugf("cookiefilter is %#v", req.CookieFilter)
		buf, _ := json.Marshal(req.CookieFilter)

		response, err := http.Post(postURL, "application/json", strings.NewReader(string(buf)))
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			msg, _ := ioutil.ReadAll(response.Body)
			http.Error(w, fmt.Sprintf("后端调用失败:%s, statuscode is %d", string(msg), response.StatusCode), http.StatusInternalServerError)
			return
		} else {
			HttpOK(w, nil)
		}

	})
}

type VDomainCreateRequest struct {
	VDomainRequestHead
	daemon.HealthCheckCfg
}

//curl -i -X POST -d '{"PoolId":"59c07d76421aa92b96679283"}' http://localhost:8080/api/lb/domains/b.com/create
func createVDomain(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	req := &VDomainCreateRequest{}

	req.Timeout = 1000
	req.Fall = 3
	req.Rise = 2
	req.Concurrency = 10
	req.Interval = 2000

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.LbType == "" {
		req.LbType = "zlb"
	}

	if req.Valid_statuses == "" {
		req.Valid_statuses = "200,302"
	}

	if req.Uri == "" {
		req.Uri = "/"
	}

	if req.Type == "" {
		req.Type = "http"
	}

	VDomain := mux.Vars(r)["domain"]

	utils.GetMgoCollections(ctx, w, []string{"pool"}, func(cs map[string]*mgo.Collection) {
		colPool, _ := cs["pool"]
		pool := types.PoolInfo{}

		if err := colPool.Find(bson.M{"_id": bson.ObjectIdHex(req.PoolId)}).One(&pool); err != nil {
			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if req.LbType != "zlb" {
			HttpError(w, "no such lbtype:"+req.LbType, http.StatusBadRequest)
			return
		}

		postURL := fmt.Sprintf("http://%s:6300/zlb/domains/%s/create", pool.TunneldAddr, VDomain)
		cfg := req.HealthCheckCfg
		buf, _ := json.Marshal(cfg)

		response, err := http.Post(postURL, "application/json", strings.NewReader(string(buf)))
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			msg, _ := ioutil.ReadAll(response.Body)
			http.Error(w, fmt.Sprintf("后端调用失败:%s, statuscode is %d", string(msg), response.StatusCode), http.StatusInternalServerError)
			return
		} else {
			HttpOK(w, nil)
		}
	})

}

type VDomainListRequest struct {
	VDomainRequestHead
}
type VDomainListResponse struct {
	VDomains []string
}

// curl -i -X POST -d '{ "PoolId":"59c07d76421aa92b96679283" , "LbType":"zlb"  }'  http://localhost:8080/api/lb/domains/list
func getVDomainList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := &VDomainListRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.LbType == "" {
		req.LbType = "zlb"
	}

	utils.GetMgoCollections(ctx, w, []string{"pool"}, func(cs map[string]*mgo.Collection) {
		colPool, _ := cs["pool"]
		pool := types.PoolInfo{}

		if err := colPool.FindId(bson.ObjectIdHex(req.PoolId)).One(&pool); err != nil {
			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
				HttpError(w, "没有这样的pool "+err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, "查询pool失败:"+err.Error(), http.StatusInternalServerError)
			return
		}

		if req.LbType != "zlb" {
			HttpError(w, "no such lbtype:"+req.LbType, http.StatusBadRequest)
			return
		}

		postURL := fmt.Sprintf("http://%s:6300/zlb/domains/list", pool.TunneldAddr)

		response, err := http.Post(postURL, "application/json", nil)
		if err != nil {
			HttpError(w, "调用zlb-api查询失败"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			msg, _ := ioutil.ReadAll(response.Body)
			http.Error(w, "后端查询失败"+string(msg), http.StatusInternalServerError)
			return
		} else {

			vdomains := make([]string, 0, 10)
			_ = json.NewDecoder(response.Body).Decode(&vdomains)

			if len(vdomains) != 0 {
				HttpOK(w, VDomainListResponse{VDomains: vdomains})
			} else {
				HttpOK(w, VDomainListResponse{VDomains: []string{}})
			}

		}
	})

}

type VDomainInspectRequest struct {
	//PoolId string
	//VDomain string
	//LbType string
	VDomainRequestHead
}

type VDomainInspectResponse struct {
	PoolId  string
	VDomain string
	LbType  string
	daemon.HealthCheckCfg
}

//curl -i -X POST -d '{"PoolId":"59c07d76421aa92b96679283"}' http://localhost:8080/api/lb/domains/b.com/inspect
func inspectVDomain(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := &VDomainInspectRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.LbType == "" {
		req.LbType = "zlb"
	}

	VDomain := mux.Vars(r)["domain"]

	utils.GetMgoCollections(ctx, w, []string{"pool"}, func(cs map[string]*mgo.Collection) {
		colPool, _ := cs["pool"]
		pool := types.PoolInfo{}

		if err := colPool.Find(bson.M{"_id": bson.ObjectIdHex(req.PoolId)}).One(&pool); err != nil {
			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if req.LbType != "zlb" {
			HttpError(w, "no such lbtype:"+req.LbType, http.StatusBadRequest)
			return
		}

		postURL := fmt.Sprintf("http://%s:6300/zlb/domains/%s/inspect", pool.TunneldAddr, VDomain)

		response, err := http.Post(postURL, "application/json", nil)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			msg, _ := ioutil.ReadAll(response.Body)
			http.Error(w, string(msg), http.StatusInternalServerError)
			return
		} else {

			cfg := daemon.HealthCheckCfg{}
			_ = json.NewDecoder(response.Body).Decode(&cfg)

			HttpOK(w, VDomainInspectResponse{HealthCheckCfg: cfg, PoolId: req.PoolId, LbType: req.LbType, VDomain: VDomain})

		}
	})

}

type VDomainDeleteRequest struct {
	VDomainInspectRequest
}

//curl -i -X POST -d '{"PoolId":"59c07d76421aa92b96679283"}' http://localhost:8080/api/lb/domains/b.com/remove

func removeVDomain(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := &VDomainDeleteRequest{}

	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.LbType == "" {
		req.LbType = "zlb"
	}

	VDomain := mux.Vars(r)["domain"]

	utils.GetMgoCollections(ctx, w, []string{"pool"}, func(cs map[string]*mgo.Collection) {
		colPool, _ := cs["pool"]
		pool := types.PoolInfo{}

		if err := colPool.Find(bson.M{"_id": bson.ObjectIdHex(req.PoolId)}).One(&pool); err != nil {
			if err == mgo.ErrNotFound {
				// 对错误类型进行区分，有可能只是没有这个pool，不应该用500错误
				HttpError(w, err.Error(), http.StatusNotFound)
				return
			}
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if req.LbType != "zlb" {
			HttpError(w, "no such lbtype:"+req.LbType, http.StatusBadRequest)
			return
		}

		postURL := fmt.Sprintf("http://%s:6300/zlb/domains/%s/remove", pool.TunneldAddr, VDomain)

		response, err := http.Post(postURL, "application/json", nil)
		if err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			msg, _ := ioutil.ReadAll(response.Body)
			http.Error(w, string(msg), http.StatusInternalServerError)
			return
		} else {

			HttpOK(w, nil)
		}
	})
}
