package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/types"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {

	for i := 0; i < 51; i++ {
		t.Log("\ncreate a template... ")
		templateResponse, err := createTemplate(&handlers.TemplateCreateRequest{
			Template: types.Template{
				Title:       fmt.Sprintf("template%d", i),
				Name:        fmt.Sprintf("template%d", i),
				Version:     "v1.0",
				Description: "test template",
			},
		})
		if err != nil {
			t.Error(err)
		} else {
			t.Log("create template success ")
			t.Log(templateResponse)
		}

		t.Log("insepect the   template... ")
		template, err := inspectTemplate(templateResponse.Id)
		if err != nil {
			t.Error(err)
		} else {
			t.Log(template)
		}

		t.Log("copy the template")
		newT, err := copyTemplate(templateResponse.Id, &handlers.TemplateCopyRequest{
			Title: template.Title + "clone",
		})

		if err != nil {
			t.Error(err)
		} else {
			t.Log("clone template success...")
			t.Log(newT)
		}

	}

	t.Log("\nlist the first page")
	list, err := listTemplate(&handlers.TemplateListRequest{
		PageSize: 10,
		Page:     1,
		Keyword:  "temp",
	})
	if err != nil {
		t.Error(err)
	} else {
		t.Log(list)
	}

	t.Log("\nlist the 2nd page")
	list1, err := listTemplate(&handlers.TemplateListRequest{
		PageSize: 10,
		Page:     2,
		Keyword:  "temp",
	})
	if err != nil {
		t.Error(err)
	} else {
		t.Log(list1)
	}

}

func inspectTemplate(id string) (*types.Template, error) {

	url := fmt.Sprintf("http://localhost:8080/api/templates/%s/inspect", id)
	req, err := http.NewRequest(http.MethodPost, url, nil)
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
		logrus.Debugf("login read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	result := types.Template{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func copyTemplate(id string, copyReq *handlers.TemplateCopyRequest) (*handlers.TemplateCopyResponse, error) {

	buf, _ := json.Marshal(copyReq)

	url := fmt.Sprintf("http://localhost:8080/api/templates/%s/copy", id)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))
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
		logrus.Debugf("login read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	result := handlers.TemplateCopyResponse{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func createTemplate(r *handlers.TemplateCreateRequest) (*handlers.TemplateCreateResponse, error) {

	buf, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/templates/create", strings.NewReader(string(buf)))
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
		logrus.Debugf("login read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	result := handlers.TemplateCreateResponse{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil

}

func listTemplate(r *handlers.TemplateListRequest) (*handlers.TemplateListResponse, error) {

	buf, _ := json.Marshal(r)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/templates/list", strings.NewReader(string(buf)))
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
		logrus.Debugf("login read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var result handlers.TemplateListResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil

}
