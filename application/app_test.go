package application_test

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/zanecloud/apiserver/application"
	"github.com/zanecloud/apiserver/types"
	"os"
	"testing"
)

func TestCreateApplication(t *testing.T) {

	logrus.SetOutput(os.Stderr)
	level, err := logrus.ParseLevel("debug")
	if err != nil {
		logrus.Fatalf(err.Error())
	}
	logrus.SetLevel(level)

	app1 := &types.Application{
		Name: "testx",
		Id:   "applicationidxxxxx000",
		Services: []types.Service{
			types.Service{
				ImageName: "nginx",
				ImageTag:  "1.8",
				Restart:   "always",
				Name:      "nginx",
				Volumns: []types.Volumne{
					types.Volumne{
						ContainerPath: "/var/log/nginx",
						HostPath:      "aaa",
					},
				},
				Ports: []types.Port{{80, "lbidxxxx"}},
				Envs: []types.Env{{types.Label{Name: "env1", Value: "env1"}},
					{types.Label{Name: "env2", Value: "env2"}}},
				Labels: []types.Label{{Name: "key1", Value: "value1"},
					{Name: "key2", Value: "value2"}},
			},
		},
	}
	pool1 := &types.PoolInfo{

		ProxyEndpoint: "tcp://47.92.142.65:2375",
		DriverOpts: types.DriverOpts{
			APIVersion: "v1.23",
		},
	}

	ctx := context.Background()

	err = application.UpgradeApplication(ctx, app1, pool1)

	if err != nil {
		t.Error(err)
		return

	} else {
		t.Log("create application success!")

	}

	err = application.StartApplication(ctx, app1, pool1, []string{"nginx"})
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("start the app success!")
	}

	err = application.ScaleApplication(ctx, app1, pool1, map[string]int{"nginx": 3})
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("scale the app to 3 success!")
	}

	err = application.ScaleApplication(ctx, app1, pool1, map[string]int{"nginx": 4})
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("scale the app to 4 uccess!")
	}

	err = application.ScaleApplication(ctx, app1, pool1, map[string]int{"nginx": 6})
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("scale the app to 6success!")
	}

	result, err := application.ListContainers(ctx, app1, pool1, []string{"nginx"})
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log(result)
	}

	err = application.StopApplication(ctx, app1, pool1, []string{"nginx"})
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log(result)
	}

	err = application.DeleteApplication(ctx, app1, pool1)
	if err != nil {
		t.Error(err.Error())
		return
	} else {
		t.Log(result)
	}
}
