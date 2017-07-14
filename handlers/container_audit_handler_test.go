package handlers_test

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/types"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestEnvTree(t *testing.T) {
	var metaId string
	var dirId string
	var myCatDirId string
	var kId string //某个参数的名称KEY的ID

	t.Run("Meta=1", func(t *testing.T) {
		if meta, err := createEnvTreeMeta(); err != nil {
			t.Error(err)
		} else {
			t.Log(meta)
			metaId = meta.Id
		}
	})

	t.Run("Meta=2", func(t *testing.T) {
		trees, err := getEnvTreeList()
		if err != nil {
			t.Error(err)
		} else if len(trees) <= 0 {
			t.Error("Meta size is not correct")
		} else {
			t.Log(trees)
		}
	})

	t.Run("Meta=3", func(t *testing.T) {
		tree := handlers.EnvTreeMetaResponse{}
		err := updateEnvTree(metaId, &tree)
		if err != nil {
			t.Error(err)
		} else if tree.Name != "TestTreeMetaModify" {
			t.Error("Meta update failed")
		} else if tree.Description != "TestDescriptionXXXXXXXXXXXXModify" {
			t.Error("Meta update failed")
		} else {
			t.Log(tree)
		}
	})

	t.Run("Meta=4", func(t *testing.T) {
		trees_before, err := getEnvTreeList()
		if err != nil {
			t.Error(err)
		} else if len(trees_before) <= 0 {
			t.Error("Meta size is not correct")
		} else {
			t.Log(trees_before)
		}

		err = deleteEnvTree(metaId)
		if err != nil {
			t.Error(err)
		}

		trees_after, err := getEnvTreeList()
		if err != nil {
			t.Error(err)
		} else if len(trees_before) <= 0 {
			t.Error("Meta size is not correct")
		} else {
			t.Log(trees_after)
		}

		if len(trees_before) == (len(trees_after) + 1) {
			t.Log("Delete one Success")
		} else {
			t.Error("Deleted failed, before:%#v, after:%#v", trees_before, trees_after)
		}
	})

	t.Run("Meta=5", func(t *testing.T) {
		if tree, err := getEnvTreeDirList(metaId); err != nil {
			t.Error(err)
		} else if tree == nil {
			t.Error("Tree List is nil")
		} else {
			//得到一个空的结构体
			t.Log(tree)
		}
	})
}

/*
	测试辅助方法
*/

func postTestRequest(urlPath string, data interface{}, instance *interface{}) (err error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/%s", urlPath), strings.NewReader(string(buf)))
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

	if err := json.Unmarshal(body, instance); err != nil {
		return err
	}

	return nil

}

func getTestRequest(urlPath string, instance *interface{}) (err error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/api/%s", urlPath), strings.NewReader(""))
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

	t := []handlers.EnvTreeMetaResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return err
	}

	return nil
}
