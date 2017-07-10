package application_test

import (
	"flag"
	"github.com/zanecloud/apiserver/application"
	"github.com/zanecloud/apiserver/types"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", "/tmp")
	flag.Set("v", "3")
	flag.Parse()

	ret := m.Run()
	os.Exit(ret)
}
func TestCreateApplication(t *testing.T) {

	app1 := &types.Application{
		Name: "nginxtest",
		Services: []types.Service{
			types.Service{
				ImageName: "docker.io/nginx",
				ImageTag:  "1.8",
				Restart:   "always",
				Name:      "nginx",
				Ports:     []string{"80"},
			},
		},
	}
	pool1 := &types.PoolInfo{
		ProxyEndpoint: "tcp://127.0.0.1:57336",
		DriverOpts: types.DriverOpts{
			APIVersion: "v1.23",
		},
	}
	err := application.CreateApplication(app1, pool1)

	if err != nil {
		t.Error(err)
		return

	} else {
		t.Log("create application success!")

	}

	err = application.StartApplication(app1, pool1, []string{"nginx"})
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("start the app success!")
	}

	err = application.ScaleApplication(app1, pool1, map[string]int{"nginx": 3})
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("scale the app success!")
	}

	err = application.ScaleApplication(app1, pool1, map[string]int{"nginx": 6})
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("scale the app success!")
	}

	result, err := application.ListContainers(app1, pool1, []string{"nginx"})
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log(result)
	}
}
