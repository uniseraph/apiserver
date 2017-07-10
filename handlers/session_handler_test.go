package handlers_test

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestSession(t *testing.T) {
	for _, ck := range cookies {
		log.Infof("cookie: %s => %s", ck.Name, ck.Value)
	}

	err := sessionLogout()
	if err != nil {
		t.Error(err)
	} else {
		t.Log("Logout success")
	}

	//登录root用户
	//获得测试使用的登录态cookie
	resp, err := sessionCreate(client, nil)
	if err != nil {
		t.Error(err)
	} else {
		cookies = resp.Cookies()
		t.Log(cookies)
	}
}

//root用户登录
//获取登陆后的cookie
func sessionCreate(client *http.Client, user *handlers.SessionCreateResp) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", "http://localhost:8080/api/users/root/login?Pass=hell05a", nil)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		log.Errorf("login post err:%s", err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("login statuscode:%d err:%s", resp.StatusCode, string(body))
	}

	//fmt.Println(string(body))
	json.Unmarshal(body, user)
	log.Infof("Resp body: %s", string(body))

	return resp, nil
}

func sessionLogout() error {

	//先调用登出接口
	//是的cookie失效
	url := fmt.Sprintf("http://localhost:8080/api/session/logout")

	req, err := http.NewRequest(http.MethodPost, url, nil)
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

	if resp.StatusCode != http.StatusOK {

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("login read body statuscode:%d err:%s", resp.StatusCode, err.Error())
			return err
		}

		return errors.New(string(body))
	}
	resp.Body.Close()
	//log.Infof("logout success.")

	//再调用current user接口
	//应该看到返回状态是401

	url = fmt.Sprintf("http://localhost:8080/api/users/current")

	req, err = http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//如果是未授权，则退出成功
	if resp.StatusCode != http.StatusForbidden {
		return errors.New(string("After logout, current user api is not 401"))
	} else {
		return nil
	}

	return nil

}
