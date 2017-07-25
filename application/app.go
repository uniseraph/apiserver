package application

import (
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/zanecloud/apiserver/types"

	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	composeyml "github.com/docker/libcompose/yaml"
	"github.com/zanecloud/apiserver/proxy/swarm"
	"gopkg.in/yaml.v2"
	"strconv"
)

//需要根据pool的驱动不同，调用不同的接口创建容器／应用，暂时只管swarm/compose
func UpApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo) error {

	p, err := buildProject(app, pool)
	if err != nil {
		return err
	}
	err = p.Up(ctx, options.Up{
		options.Create{ForceRecreate: false,
			NoBuild:    true,
			ForceBuild: false},
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Debug("up application err")
		return err
	}

	return nil

}

func UpgradeApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo) error {
	p, err := buildProject(app, pool)
	if err != nil {
		return err
	}
	err = p.Upgrade(ctx, options.Up{
		options.Create{ForceRecreate: false,
			NoBuild:    true,
			ForceBuild: false},
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Debug("up application err")
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
func buildDefaultNetwork() map[string]*config.NetworkConfig {
	result := make(map[string]*config.NetworkConfig)

	//result["default"]=&config.NetworkConfig{
	//	Driver: "bridge",
	//}

	return result
}

func buildDefaultVolumes() map[string]*config.VolumeConfig {
	return make(map[string]*config.VolumeConfig)
}
func buildComposeFileBinary(app *types.Application, pool *types.PoolInfo) (buf []byte, err error) {

	ec := &project.ExportedConfig{
		Version:  "2",
		Services: map[string]*config.ServiceConfig{},
		Networks: buildDefaultNetwork(),
		Volumes:  buildDefaultVolumes(),
	}

	for _, appService := range app.Services {

		composeService := &config.ServiceConfig{
			Image:       appService.ImageName + ":" + appService.ImageTag,
			Restart:     appService.Restart,
			NetworkMode: "bridge",
			CPUSet:      appService.CPU,
			//Ports:       s.Ports,
		}

		if appService.Memory != "" {
			mem, err := strconv.ParseInt(appService.Memory, 10, 64)

			if err == nil {
				composeService.MemLimit = composeyml.MemStringorInt(mem << 20)
			}
		}

		composeService.Ports = make([]string, len(appService.Ports))
		for i, _ := range appService.Ports {
			composeService.Ports[i] = strconv.Itoa(appService.Ports[i].SourcePort)
		}

		composeService.Labels = map[string]string{}
		for i, _ := range appService.Labels {
			composeService.Labels[appService.Labels[i].Name] = appService.Labels[i].Value
		}
		composeService.Labels[swarm.LABEL_APPLICATION_ID] = app.Id.Hex()

		//TODO 加上cpuset的label
		if appService.CPU != "" && appService.ExclusiveCPU == true {
			composeService.Labels[types.LABEL_CONTAINER_CPUS] = appService.CPU
			//	composeService.Labels[swarm.LABEL_CONTAINER_EXCLUSIVE] = appService.ExclusiveCPU
		}

		composeService.Environment = make([]string, 0, len(appService.Envs))
		for i, _ := range appService.Envs {

			composeService.Environment = append(composeService.Environment, fmt.Sprintf("%s=%s", appService.Envs[i].Name, appService.Envs[i].Value))
		}


		composeService.Volumes = &composeyml.Volumes{
			Volumes: make([]*composeyml.Volume,0,len(appService.Volumns)),
		}


		for i,_ := range appService.Volumns {

			composeService.Volumes.Volumes = append(composeService.Volumes.Volumes , &composeyml.Volume{
				Destination : appService.Volumns[i].ContainerPath ,
//				AccessMode: "rw",

			}  )
		}

		ec.Services[appService.Name] = composeService

	}

	buf, err = yaml.Marshal(ec)

	if err != nil {
		return
	}

	logrus.Debugf("application %#v encode to bytes is \n%s", app, string(buf))

	return
}

func StartApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo, services []string) error {
	p, err := buildProject(app, pool)
	if err != nil {
		return err
	}
	if err := p.Start(ctx, services...); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Debug("start application err")
		return err
	}
	return nil
}

func ScaleApplication(ctx context.Context, app *types.Application, pool *types.PoolInfo, services map[string]int) error {
	p, err := buildProject(app, pool)
	if err != nil {
		return err
	}

	if err := p.Scale(ctx, 30, services); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Debug("scale application err")
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
		logrus.WithFields(logrus.Fields{"err": err}).Debug("list application err")
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
		logrus.WithFields(logrus.Fields{"err": err}).Debug("stop application err")
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
		logrus.WithFields(logrus.Fields{"err": err}).Debug("delete application err")
		return err
	}

	return nil
}
