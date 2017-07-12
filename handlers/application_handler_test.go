package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/handlers"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func createApplication(name string, request *handlers.ApplicationCreateRequest) (interface{}, error) {

	url := fmt.Sprintf("http://localhost:8080/api/applications/create")

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
		logrus.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
		return err.Error(), err
	}

	if resp.StatusCode != http.StatusOK {
		return string(body), errors.New(string(body))
	}

	//fmt.Println(string(body))
	result := handlers.ApplicationCreateResponse{}
	json.Unmarshal(body, &result)

	return result, nil
}

func TestApplication(t *testing.T) {

}
