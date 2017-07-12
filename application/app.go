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
func CreateApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo) error {

	p, err := buildProject(app, pool)
	if err != nil {
		return nil
	}
	err = p.Up(ctx, options.Up{
		options.Create{ForceRecreate: false,
			NoBuild:    true,
			ForceBuild: false},
	})

	if err != nil {
		return err
	}
	return nil

}

func buildProject(app *types.Application, pool *types.PoolInfo) (p project.APIProject, err error) {
	buf, err := buildComposeFileBinary(app, pool)
	if err != nil {
		return nil, err
	}

	factory, err := client.NewDefaultFactory(client.Options{
		TLS:        false,
		TLSVerify:  false,
		Host:       pool.ProxyEndpoint,
		APIVersion: pool.DriverOpts.APIVersion,
	})

	if err != nil {
		return nil, err
	}

	p, err = docker.NewProject(&ctx.Context{
		Context: project.Context{
			//ComposeFiles: []string{"docker-compose.yml"},
			ComposeBytes: [][]byte{buf},
			ProjectName:  app.Name,
		},
		ClientFactory: factory,
	}, nil)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func buildComposeFileBinary(app *types.Application, pool *types.PoolInfo) (buf []byte, err error) {

	ec := &project.ExportedConfig{
		Version:  "2",
		Services: map[string]*config.ServiceConfig{},
	}

	for _, s := range app.Services {

		composeService := &config.ServiceConfig{
			Image:       s.ImageName + ":" + s.ImageTag,
			Restart:     s.Restart,
			NetworkMode: "bridge",
			Ports:       s.Ports,
		}

		ec.Services[s.Name] = composeService

	}

	buf, err = yaml.Marshal(ec)

	if err != nil {
		return
	}

	logrus.Debugf("application %#v encode to bytes is %s", app, string(buf))

	return
}

func StartApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo, services []string) error {
	p, err := buildProject(app, pool)
	if err != nil {
		return nil
	}
	if err := p.Start(ctx, services...); err != nil {
		return err
	}
	return nil
}

func ScaleApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo, services map[string]int) error {
	p, err := buildProject(app, pool)
	if err != nil {
		return nil
	}

	if err := p.Scale(ctx, 30, services); err != nil {
		return err
	}
	return nil
}

func ListContainers(ctx context.Context, app *types.Application, pool *types.PoolInfo, services []string) ([]string, error) {
	p, err := buildProject(app, pool)
	if err != nil {
		return nil, err
	}

	result, err := p.Containers(ctx, project.Filter{project.AnyState}, services...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func StopApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo, services []string) error {

	p, err := buildProject(app, pool)
	if err != nil {
		return err
	}

	if err := p.Stop(ctx, 30, services...); err != nil {
		return err
	}

	return nil
}

func DeleteApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo) error {

	p, err := buildProject(app, pool)
	if err != nil {
		return err
	}

	if err := p.Delete(ctx, options.Delete{false, true}); err != nil {
		return err
	}

	return nil
}
