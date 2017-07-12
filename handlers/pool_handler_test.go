package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	dockerclient "github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/types"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var pool *handlers.PoolsRegisterResponse

func TestPool(t *testing.T) {

	t.Run("Pool=1", func(t *testing.T) {
		var metaId string
		if meta, err := createEnvTreeMeta(); err != nil {
			t.Error(err)
		} else {
			t.Log(meta)
			metaId = meta.Id
		}

		data, err := registerPool("pool1", &handlers.PoolsRegisterRequest{
			Name:      "pool1",
			Driver:    "swarm",
			EnvTreeId: metaId,
			DriverOpts: types.DriverOpts{
				Version:    "v1.0",
				EndPoint:   "tcp://47.92.49.245:2375",
				APIVersion: "v1.23",
			},
		})

		pool = data

		log.Infof("PPPPPPPPPPPPPPPPPPPP: %s", pool.Name)

		if err != nil {
			t.Error("registerPool Error:", err)
			return
		} else {
			t.Log(pool)
		}

		dockerclient, err := dockerclient.NewClient(pool.Proxy, "v1.23", nil, map[string]string{})
		if err != nil {
			t.Error(err)
		}

		info, err := dockerclient.Info(context.Background())
		if err != nil {
			t.Error(err)
			return
		} else {
			t.Log(info)
		}

		poolFlushResponse, err := flushPool(pool.Id)
		if err != nil {
			t.Error(err)
			return
		} else {
			t.Log("flush the pool success !")
			t.Log(poolFlushResponse)
		}
	})

	t.Run("Pool=2", func(t *testing.T) {
		log.Infof("=============\nPool is %s, %s", pool.Name, user.Id)
		if err := addUserPoolId(pool.Id, user.Id); err != nil {
			t.Error("addUserPoolId Error: ", err)
		}
	})

	t.Run("Pool=3", func(t *testing.T) {
		if rlts, err := getUserPools(); err != nil {
			t.Error("getUserPools Error:", err)
		} else if len(rlts) != 1 {
			t.Error("Pools size is not correct:", len(rlts))
		} else if rlts[0].Name != "pool1" {
			t.Error("Pool Name is not correct:", rlts)
		} else {
			t.Log(rlts)
		}
	})
}

func registerPool(name string, request *handlers.PoolsRegisterRequest) (*handlers.PoolsRegisterResponse, error) {

	url := fmt.Sprintf("http://localhost:8080/api/pools/register")

	buf, _ := json.Marshal(request)

	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	fmt.Println(string(body))
	result := &handlers.PoolsRegisterResponse{}
	json.Unmarshal(body, result)

	return result, nil
}

func flushPool(id string) (*handlers.PoolsFlushResponse, error) {

	url := fmt.Sprintf("http://localhost:8080/api/pools/%s/flush", id)

	//buf, _ := json.Marshal(request)

	req, _ := http.NewRequest(http.MethodPost, url, nil)

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	//fmt.Println(string(body))
	result := &handlers.PoolsFlushResponse{}
	json.Unmarshal(body, result)

	return result, nil
}

func addUserPoolId(id string, uid string) error {
	url := fmt.Sprintf("http://localhost:8080/api/pools/%s/add-user?UserId=%s", id, uid)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(""))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	return nil
}

func getUserPools() (pools []*handlers.UserPoolsResponse, err error) {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/user/pools", strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	t := []*handlers.UserPoolsResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil
}
