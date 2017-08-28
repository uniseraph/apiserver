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
	"strings"
)

//from watchdog elbv2
const LABEL_ELBV2_ENABLE = "com.zanecloud.elbv2.enable"
const LABEL_ELBV2_TARGET_GROUP_ARN = "com.zanecloud.elbv2.target.grouparn"
const LABEL_ELBV2_TARGET_PORT = "com.zanecloud.elbv2.target.port"

//from watchdog slb

const LABEL_SLB_ENABLE = "com.zanecloud.slb.enable"
const LABEL_SLB_VSERVER_GROUP_ID = "com.zanecloud.slb.vservergroupid"
const LABEL_SLB_PORT = "com.zanecloud.slb.port"

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

	for _, as := range app.Services {

		sc := &config.ServiceConfig{
			Image:       as.ImageName + ":" + as.ImageTag,
			Restart:     as.Restart,
			NetworkMode: "bridge",
			//CPUSet:      as.CPU,
			Expose: []string{},
			Labels: make(map[string]string),
			//Ports:       s.Ports,
		}
		if as.NetworkMode == "host" {
			sc.NetworkMode = "host"
		}

		if as.Memory != "" {
			mem, err := strconv.ParseInt(as.Memory, 10, 64)
			if err == nil {
				sc.MemLimit = composeyml.MemStringorInt(mem << 20)
			}
		}
		for i, _ := range as.Labels {
			sc.Labels[as.Labels[i].Name] = strings.Replace(as.Labels[i].Value, "$", "$$", -1)
		}
		sc.Labels[swarm.LABEL_APPLICATION_ID] = app.Id.Hex()

		capNetAdmin := false

		sc.Ports = make([]string, len(as.Ports))
		for i, _ := range as.Ports {
			sc.Ports[i] = strconv.Itoa(as.Ports[i].SourcePort)
			if as.Ports[i].SourcePort < 1024 && as.NetworkMode == "host" && !capNetAdmin {
				sc.CapAdd = append(sc.CapAdd, "NET_ADMIN")
				capNetAdmin = true
			}

			if as.Ports[i].TargetGroupArn != "" && as.Ports[i].LoadBalancerId != "" {
				// aliyun slb
				sc.Labels[LABEL_SLB_ENABLE] = "true"
				//sc.Labels[LABEL_SLB_LBID] = as.Ports[i].LoadBalancerId
				sc.Labels[LABEL_SLB_VSERVER_GROUP_ID] = as.Ports[i].TargetGroupArn
				sc.Labels[LABEL_SLB_PORT] = strconv.Itoa(as.Ports[i].SourcePort)
			}

			if as.Ports[i].TargetGroupArn != "" && as.Ports[i].LoadBalancerId == "" {
				// aws elbv2
				sc.Labels[LABEL_ELBV2_ENABLE] = "true"
				sc.Labels[LABEL_ELBV2_TARGET_GROUP_ARN] = as.Ports[i].TargetGroupArn
				sc.Labels[LABEL_ELBV2_TARGET_PORT] = strconv.Itoa(as.Ports[i].SourcePort)
			}

			//expose
			//Expose ports without publishing them to the host machine - they’ll only be accessible to linked services.
			// Only the internal port can be specified.

			//if appService.NetworkMode == "host" {
			//	composeService.Expose = append(composeService.Expose, composeService.Ports[i])
			//}
		}

		if as.ExclusiveCPU == true {
			sc.Labels[types.LABEL_CONTAINER_EXCLUSIVE] = "true"
		}
		if as.CPU != "" {
			sc.Labels[types.LABEL_CONTAINER_CPUS] = as.CPU
		}

		sc.Environment = make([]string, 0, len(as.Envs))
		for i, _ := range as.Envs {
			parsedValue := strings.Replace(as.Envs[i].Value, "$", "$$", -1)
			sc.Environment = append(sc.Environment, fmt.Sprintf("%s=%s", as.Envs[i].Name, parsedValue))
		}

		sc.Volumes = &composeyml.Volumes{
			Volumes: make([]*composeyml.Volume, 0, len(as.Volumns)),
		}

		for i, _ := range as.Volumns {
			if as.Volumns[i].HostPath == "" { // 不指定宿主机目录，随便挂 ,匿名卷
				sc.Volumes.Volumes = append(sc.Volumes.Volumes, &composeyml.Volume{
					Destination: as.Volumns[i].ContainerPath,
				})
			} else if as.Volumns[i].HostPath[0:2] == "./" { //指定宿主机目录
				sc.Volumes.Volumes = append(sc.Volumes.Volumes, &composeyml.Volume{
					Destination: as.Volumns[i].ContainerPath,
					Source:      as.Volumns[i].HostPath,
				})

			} else { //有名卷方式，
				sc.Volumes.Volumes = append(sc.Volumes.Volumes, &composeyml.Volume{
					Destination: as.Volumns[i].ContainerPath,
					Source:      as.Volumns[i].HostPath,
				})

				ec.Volumes[as.Volumns[i].HostPath] = &config.VolumeConfig{}
			}
		}
		ec.Services[as.Name] = sc
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
