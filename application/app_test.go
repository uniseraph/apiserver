package application_test

import (
	"github.com/zanecloud/apiserver/application"
	"github.com/zanecloud/apiserver/types"
	"testing"
)

func TestCreateApplication(t *testing.T) {

	err := application.CreateApplication(&types.Application{
		Name: "nginx",
		Services: []types.Service{
			types.Service{
				ImageName: "docker.io/nginx",
				ImageTag:  "1.8",
				Restart:   "always",
			},
		},
	}, &types.PoolInfo{

		ProxyEndpoint: "tcp://127.0.0.1:50369",
		DriverOpts: types.DriverOpts{
			APIVersion: "v1.23",
		},
	})

	if err != nil {
		t.Log("create application success!")
	} else {
		t.Error(err)
	}

}
