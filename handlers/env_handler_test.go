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

	//创建一棵树的全部节点
	//全部->mysql root->[master->mycat->[proxy01, proxy02], slave]
	//树的形状很奇怪，是为了测试，真实系统不会这么奇怪
	t.Run("Meta=6", func(t *testing.T) {
		meta, err := createEnvTreeMeta()
		id := meta.Id
		if err != nil {
			t.Error(err)
		} else {
			t.Log(meta)
			metaId = id
		}
		if dir, err := createEnvTreeNodeDirMySQLRoot(id, meta.Root); err != nil {
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
					myCatDirId = cat_dir.Id
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
			//全部->mysql root->[master->mycat->[proxy01, proxy02], slave]
			t.Log(tree)

			if tree.Name != "全部" {
				t.Error("Name Error", tree)
			}
			if tree.ParentId != "" {
				t.Error("ParentId Error", tree.ParentId)
			}
			if len(tree.Children) != 1 {
				t.Error("tree.Children size Error", tree.Children)
			}
			if len(tree.Children[0].Children) != 2 {
				t.Error("tree.Children[0].Children Error", tree.Children[0].Children)
			}
			if tree.Children[0].Children[0].Children[0].Name != "MyCAT Proxy" {
				t.Error("tree.Children[0].Children[0].Name Error", tree.Children[0].Children[0])
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
		if s_dir, err := createEnvTreeNodeDirMySQLRootSlaveNode(metaId, dirId); err != nil {
			t.Error(err)
		} else if s_dir.Name != "Slave" {
			t.Error("Name is not correct")
		} else {
			t.Log(s_dir)
			dirId = s_dir.Id
		}
	})

	//测试删除某个节点
	//检查，从父节点上找该节点应该找不到了
	t.Run("Meta=10", func(t *testing.T) {
		if err := deleteEnvDirNode(dirId); err != nil {
			t.Error(err)
		} else {
			t.Log(dirId)
		}

		//检查该树，是否还存在root->Slave-DirNodeModify->slave节点
		if tree, err := getEnvTreeDirList(metaId); err != nil {
			t.Error(err)
		} else {
			//得到一个这个样子的树
			//全部->mysql root->master->mycat->[proxy01, proxy02]
			if len(tree.Children[0].Children[1].Children) != 0 {
				t.Error(tree.Children[1])
			}
		}
	})

	//测试在my cat dir下面建立参数节点
	t.Run("Meta=11", func(t *testing.T) {
		if k, err := createParamsKey(metaId, myCatDirId); err != nil {
			t.Error(k)
		} else if k.Name != "balance" {
			t.Error(k)
		} else if k.Value != "round robin" {
			t.Error(k)
		} else {
			t.Log(k)
			kId = k.Id
		}
	})

	//测试更新该参数节点属性
	t.Run("Meta=12", func(t *testing.T) {
		if k, err := updateParamsKey(kId); err != nil {
			t.Error(err)
		} else if k.Value != "round robin modify" {
			t.Error(k)
		} else {
			t.Log(k)
		}
	})

	t.Run("Meta=13", func(t *testing.T) {
		if err := updateParamsKeyValues(kId); err != nil {
			t.Error(err)
		}
	})

	//测试删除一个参数节点
	t.Run("Meta=14", func(t *testing.T) {
		//先删除
		if err := deleteParamsKey(kId); err != nil {
			t.Error(err)
		}
		//再更新
		if k, err := updateParamsKey(kId); err != nil {
			t.Log(err)
		} else {
			t.Error(k)
		}
	})

	//测试分页展示参数KEY
	t.Run("Meta=15", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			if k, err := createParamsKeyWithName(metaId, myCatDirId, fmt.Sprintf("TestKey-%d", i)); err != nil {
				t.Error(k)
			} else if k.Value != "round robin" {
				t.Error("Value:", k.Value)
			} else {
				//t.Log(k)
			}

			if k, err := createParamsKeyWithName(metaId, myCatDirId, fmt.Sprintf("FakeKey-%d", i)); err != nil {
				t.Error(k)
			} else if k.Value != "round robin" {
				t.Error("Value", k.Value)
			} else {
				//t.Log(k)
			}
		}

		if rsp, err := getEnvValuesList(metaId, myCatDirId, "Test"); err != nil {
			t.Error(err)
		} else if len(rsp.Data) != 5 {
			t.Error("rsp data size is not correct!\n", rsp.Data)
		} else {
			for _, kv := range rsp.Data {
				if strings.Index(kv.Name, "Test") != 0 {
					t.Error("Key name is not matched: ", kv.Name)
				} else {
					t.Log("Key name is : ", kv.Name)
				}
			}
		}

	})

	//测试根据PoolID和KeyID查询Value
	t.Run("Meta=16", func(t *testing.T) {
		//建立Key
		if k, err := createParamsKey(metaId, myCatDirId); err != nil {
			t.Error(k)
		} else if k.Name != "balance" {
			t.Error(k)
		} else if k.Value != "round robin" {
			t.Error(k)
		} else {
			t.Log(k)
			kId = k.Id
		}

		pId := bson.NewObjectId().Hex()

		t.Log("Pid:", pId, "Kid:", kId)
		//当pool和key没有关系的时候
		//查询value其实是key的default值
		if rsp, err := getEnvValue(pId, kId); err != nil {
			t.Error(err)
		} else if rsp.Value == "VALUE1" {
			t.Error("EnvValue is not correct: ", rsp.Value)
		} else if rsp.Value != "round robin" {
			t.Error("EnvValue is not correct: ", rsp.Value)
		} else {
			t.Log(rsp)
		}

		//建立Key的Value和PoolId的对应关系
		if err := updateParamsKeyValuesWithPoolId(kId, pId); err != nil {
			t.Error(err)
		}

		//根据上面建立好的对应关系
		//查询有具体值的KEY
		if rsp, err := getEnvValue(pId, kId); err != nil {
			t.Error(err)
		} else if rsp.Value != "poolId-VALUE1" {
			t.Error("EnvValue is not correct: ", rsp.Value)
		} else {
			t.Log(rsp)
		}
	})
}

/*
	EnvTree测试
*/

func createEnvTreeMeta() (*handlers.EnvTreeMetaResponse, error) {
	tree := &types.EnvTreeMeta{
		Name:        "TestTreeMeta",
		Description: "TestDescriptionXXXXXXXXXXXX",
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/trees/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeMetaResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

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
	//=============
	log.Infoln("========================")
	log.Infoln(string(body))
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

func createEnvTreeNodeDirMySQLRoot(tree_id string, parent_id string) (*handlers.EnvTreeNodeDirResponse, error) {
	tree := &handlers.EnvTreeNodeDirRequest{
		Name:     "MySQL",
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

func createParamsKey(tree_id string, dir_id string) (*handlers.EnvTreeNodeParamKVResponse, error) {
	tree := &handlers.EnvTreeNodeParamKVRequest{
		Name:        "balance",
		Value:       "round robin",
		Description: "MyCat转发请求的负载均衡策略",
		DirId:       dir_id,
		TreeId:      tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/values/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeParamKVResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

func updateParamsKey(id string) (t *handlers.EnvTreeNodeParamKVResponse, err error) {
	v := &handlers.EnvTreeNodeParamKVRequest{
		Value: "round robin modify",
	}

	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://localhost:8080/api/envs/values/%s/update", id)
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
		log.Debugf("login read body err:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	t = &handlers.EnvTreeNodeParamKVResponse{}
	if err := json.Unmarshal(body, t); err != nil {
		return nil, err
	}

	return
}

func deleteParamsKey(id string) error {
	url := fmt.Sprintf("http://localhost:8080/api/envs/values/%s/remove", id)
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

func updateParamsKeyValues(id string) error {
	vs := []*handlers.EnvValuesUpdateValues{}
	vs = append(vs, &handlers.EnvValuesUpdateValues{
		PoolId: bson.NewObjectId().Hex(),
		Value:  "VALUE1",
	})
	vs = append(vs, &handlers.EnvValuesUpdateValues{
		PoolId: bson.NewObjectId().Hex(),
		Value:  "VALUE2",
	})
	vs = append(vs, &handlers.EnvValuesUpdateValues{
		PoolId: bson.NewObjectId().Hex(),
		Value:  "VALUE3",
	})

	buf, err := json.Marshal(vs)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://localhost:8080/api/envs/values/%s/update-values", id)
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

	return nil
}

func createParamsKeyWithName(tree_id string, dir_id string, name string) (*handlers.EnvTreeNodeParamKVResponse, error) {
	tree := &handlers.EnvTreeNodeParamKVRequest{
		Name:        name,
		Value:       "round robin",
		Description: "MyCat转发请求的负载均衡策略",
		DirId:       dir_id,
		TreeId:      tree_id,
	}

	buf, err := json.Marshal(tree)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/values/create", strings.NewReader(string(buf)))
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

	t := &handlers.EnvTreeNodeParamKVResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil

}

func getEnvValuesList(tree string, dir string, name string) (*handlers.EnvValuesListResponse, error) {
	data := &handlers.EnvValuesListRequest{
		Name:     name,
		PageSize: 5,
		Page:     3,
		DirId:    dir,
		TreeId:   tree,
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/envs/values/list", strings.NewReader(string(buf)))
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

	t := &handlers.EnvValuesListResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil
}

//获取某个PoolId和KeyId的Value
func getEnvValue(poolId string, keyId string) (*handlers.EnvValuesDetailsValueResponse, error) {
	url := fmt.Sprintf("http://localhost:8080/api/envs/value/get?PoolId=%s&KeyId=%s", poolId, keyId)

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

	t := &handlers.EnvValuesDetailsValueResponse{}
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, err
	}

	return t, nil
}

func updateParamsKeyValuesWithPoolId(id string, poolId string) error {
	vs := []*handlers.EnvValuesUpdateValues{}
	vs = append(vs, &handlers.EnvValuesUpdateValues{
		PoolId: poolId,
		Value:  "poolId-VALUE1",
	})

	buf, err := json.Marshal(vs)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://localhost:8080/api/envs/values/%s/update-values", id)
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

	return nil
}
