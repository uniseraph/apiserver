package handlers_test

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/zanecloud/apiserver/handlers"
	"github.com/zanecloud/apiserver/types"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestEnvTree(t *testing.T) {
	var metaId string
	var dirId string

	t.Run("Meta=1", func(t *testing.T) {
		if id, err := createEnvTreeMeta(); err != nil {
			t.Error(err)
		} else {
			t.Log(id)
			metaId = id
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

	//创建一棵树的全部节点
	//mysql root->[master->mycat->[proxy01, proxy02], slave]
	//树的形状很奇怪，是为了测试，真实系统不会这么奇怪
	t.Run("Meta=6", func(t *testing.T) {
		id, err := createEnvTreeMeta()
		if err != nil {
			t.Error(err)
		} else {
			t.Log(id)
			metaId = id
		}
		if dir, err := createEnvTreeNodeDirMySQLRoot(id); err != nil {
			t.Error(err)
		} else if dir.Name != "MySQL" {
			t.Error("Name is not correct")
		} else {
			t.Log(dir)
			if m_dir, err := createEnvTreeNodeDirMySQLRootMasterNode(metaId, dir.Id); err != nil {
				t.Error(err)
			} else if m_dir.Name != "Master" {
				t.Error("Name is not correct")
			} else {
				t.Log(m_dir)
				if cat_dir, err := createEnvTreeNodeDirMySQLRootMycatproxyNode(metaId, m_dir.Id); err != nil {
					t.Error(err)
				} else if cat_dir.Name != "MyCAT Proxy" {
					t.Error("Name is not correct")
				} else {
					t.Log(cat_dir)
					if cat_dir, err := createEnvTreeNodeDirMySQLRootMycatproxy01Node(metaId, cat_dir.Id); err != nil {
						t.Error(err)
					} else if cat_dir.Name != "Proxy01" {
						t.Error("Name is not correct")
					} else {
						t.Log(cat_dir)
					}

					if cat_dir, err := createEnvTreeNodeDirMySQLRootMycatproxy02Node(metaId, cat_dir.Id); err != nil {
						t.Error(err)
					} else if cat_dir.Name != "Proxy02" {
						t.Error("Name is not correct")
					} else {
						t.Log(cat_dir)
					}
				}
			}

			if s_dir, err := createEnvTreeNodeDirMySQLRootSlaveNode(metaId, dir.Id); err != nil {
				t.Error(err)
			} else if s_dir.Name != "Slave" {
				t.Error("Name is not correct")
			} else {
				t.Log(s_dir)
				dirId = s_dir.Id
			}
		}
	})

	t.Run("Meta=7", func(t *testing.T) {
		if tree, err := getEnvTreeDirList(metaId); err != nil {
			t.Error(err)
		} else {
			//得到一个这个样子的树
			//mysql root->[master->mycat->[proxy01, proxy02], slave]
			t.Log(tree)

			if tree.Name != "MySQL" {
				t.Error(tree)
			}
			if tree.ParentId != "" {
				t.Error(tree.ParentId)
			}
			if len(tree.Children) != 2 {
				t.Error(tree.Children)
			}
			if tree.Children[0].Children[0].Name != "MyCAT Proxy" {
				t.Error(tree.Children[0].Children[0])
			}
		}
	})

	t.Run("Meta=8", func(t *testing.T) {
		dir := &handlers.EnvTreeNodeDirResponse{}
		if err := updateEnvDirNode(dirId, dir); err != nil {
			t.Error(err)
		} else if dir.Name != "Slave-DirNodeModify" {
			t.Error("Dir Node Name is not correct.")
		} else {
			t.Log(dir)
		}
	})

	t.Run("Meta=9", func(t *testing.T) {
		if err := deleteEnvDirNode(dirId); err != nil {
			t.Error(err)
		} else {
			t.Log(dirId)
		}

		//检查该树，是否还存在root->slave节点
		//期望只存在root->master一个节点
		if tree, err := getEnvTreeDirList(metaId); err != nil {
			t.Error(err)
		} else {
			//得到一个这个样子的树
			//mysql root->master->mycat->[proxy01, proxy02]
			if len(tree.Children) != 1 {
				t.Error(tree.Children)
			}
		}
	})
}

/*
	EnvTree测试
*/

func createEnvTreeMeta() (string, error) {
	tree := &types.EnvTreeMeta{
		Name:        "TestTreeMeta",
		Description: "TestDescriptionXXXXXXXXXXXX",
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/trees/create", strings.NewReader(string(buf)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("login read body err:%s", err.Error())
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	t := handlers.EnvTreeMetaResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return "", err
	}

	return t.Id, nil

}

func getEnvTreeList() ([]handlers.EnvTreeMetaResponse, error) {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/trees/list", strings.NewReader(""))
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

	t := []handlers.EnvTreeMetaResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil
}

///envs/trees/:id/update

func updateEnvTree(id string, t *handlers.EnvTreeMetaResponse) error {
	tree := &types.EnvTreeMeta{
		Name:        "TestTreeMetaModify",
		Description: "TestDescriptionXXXXXXXXXXXXModify",
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://localhost:8080/api/envs/trees/%s/update", id)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))
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

	if err := json.Unmarshal(body, t); err != nil {
		return err
	}

	return nil
}

///envs/trees/:id/remove
func deleteEnvTree(id string) error {
	url := fmt.Sprintf("http://localhost:8080/api/envs/trees/%s/remove", id)
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

	t := []handlers.EnvTreeMetaResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return err
	}

	return nil
}

func getEnvTreeDirList(id string) (*handlers.EnvTreeNodeDirsResponse, error) {
	url := fmt.Sprintf("http://localhost:8080/api/envs/dirs/list?TreeId=%s", id)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(""))
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

	t := &handlers.EnvTreeNodeDirsResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil
}

func createEnvTreeNodeDirMySQLRoot(tree_id string) (*handlers.EnvTreeNodeDirResponse, error) {
	tree := &handlers.EnvTreeNodeDirRequest{
		Name:   "MySQL",
		TreeId: tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/dirs/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeDirResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

func createEnvTreeNodeDirMySQLRootMasterNode(tree_id string, parent_id string) (*handlers.EnvTreeNodeDirResponse, error) {
	tree := &handlers.EnvTreeNodeDirRequest{
		Name:     "Master",
		ParentId: parent_id,
		TreeId:   tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/dirs/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeDirResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

func createEnvTreeNodeDirMySQLRootSlaveNode(tree_id string, parent_id string) (*handlers.EnvTreeNodeDirResponse, error) {
	tree := &handlers.EnvTreeNodeDirRequest{
		Name:     "Slave",
		ParentId: parent_id,
		TreeId:   tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/dirs/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeDirResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

func createEnvTreeNodeDirMySQLRootMycatproxyNode(tree_id string, parent_id string) (*handlers.EnvTreeNodeDirResponse, error) {
	tree := &handlers.EnvTreeNodeDirRequest{
		Name:     "MyCAT Proxy",
		ParentId: parent_id,
		TreeId:   tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/dirs/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeDirResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

func createEnvTreeNodeDirMySQLRootMycatproxy01Node(tree_id string, parent_id string) (*handlers.EnvTreeNodeDirResponse, error) {
	tree := &handlers.EnvTreeNodeDirRequest{
		Name:     "Proxy01",
		ParentId: parent_id,
		TreeId:   tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/dirs/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeDirResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

func createEnvTreeNodeDirMySQLRootMycatproxy02Node(tree_id string, parent_id string) (*handlers.EnvTreeNodeDirResponse, error) {
	tree := &handlers.EnvTreeNodeDirRequest{
		Name:     "Proxy02",
		ParentId: parent_id,
		TreeId:   tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/dirs/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeDirResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

//更新一个目录节点

func updateEnvDirNode(id string, t *handlers.EnvTreeNodeDirResponse) error {
	tree := &types.EnvTreeNodeDir{
		Name: "Slave-DirNodeModify",
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://localhost:8080/api/envs/dirs/%s/update", id)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(buf)))
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

	if err := json.Unmarshal(body, t); err != nil {
		return err
	}

	return nil
}

//删除一个目录节点

func deleteEnvDirNode(id string) error {
	url := fmt.Sprintf("http://localhost:8080/api/envs/dirs/%s/remove", id)
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
