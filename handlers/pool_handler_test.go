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

func TestPool(t *testing.T) {
	result, err := registerPool("pool1", &handlers.PoolsRegisterRequest{
		Name:   "pool1",
		Driver: "swarm",
		DriverOpts: types.DriverOpts{
			Version:    "v1.0",
			EndPoint:   "tcp://47.92.49.245:2375",
			APIVersion: "v1.23",
		},
	})

	if err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}

	r, _ := result.(handlers.PoolsRegisterResponse)
	dockerclient, err := dockerclient.NewClient(r.Proxy, "v1.23", nil, map[string]string{})
	if err != nil {
		t.Error(err)
	}

	info, err := dockerclient.Info(context.Background())
	if err != nil {
		t.Error(err)
	} else {
		t.Log(info)
	}
}

func registerPool(name string, request *handlers.PoolsRegisterRequest) (interface{}, error) {

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
		return err.Error(), err
	}

	if resp.StatusCode != http.StatusOK {
		return string(body), errors.New(string(body))
	}

	//fmt.Println(string(body))
	result := handlers.PoolsRegisterResponse{}
	json.Unmarshal(body, &result)

	return result, nil
}
