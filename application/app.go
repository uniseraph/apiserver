package application

import (
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/zanecloud/apiserver/types"

	"context"
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	"gopkg.in/yaml.v2"
)

//需要根据pool的驱动不同，调用不同的接口创建容器／应用，暂时只管swarm/compose
func CreateApplication(app *types.Application, pool *types.PoolInfo) error {

	factory, err := client.NewDefaultFactory(client.Options{
		TLS:        false,
		TLSVerify:  false,
		Host:       pool.ProxyEndpoint,
		APIVersion: pool.DriverOpts.APIVersion,
	})

	if err != nil {
		return err
	}

	ec := &project.ExportedConfig{
		Version:  "2",
		Services: map[string]*config.ServiceConfig{},
	}

	for _, s := range app.Services {

		composeService := &config.ServiceConfig{
			Image:       s.ImageName + ":" + s.ImageTag,
			Restart:     s.Restart,
			NetworkMode: "bridge",
		}

		ec.Services[s.Name] = composeService

	}

	buf, err := yaml.Marshal(ec)

	if err != nil {
		return err
	}

	logrus.Debugf("application %#v encode to bytes is %s", app, string(buf))

	project, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			//ComposeFiles: []string{"docker-compose.yml"},
			ComposeBytes: [][]byte{buf},
			ProjectName:  app.Name,
		},
		ClientFactory: factory,
	}, nil)

	if err != nil {
		return err
	}

	err = project.Up(context.Background(), options.Up{})

	if err != nil {
		return err
	}
	return nil

}
