package project

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project/options"
	"golang.org/x/net/context"
)

// Upgrade upgrade and create-start the specified services.
func (p *Project) Upgrade(ctx context.Context, options options.Up, upgradeServices map[string]int) error {
	if err := p.initialize(ctx); err != nil {
		return err
	}
	order := make([]string, 0, 0)
	services := make(map[string]Service)

	for name := range upgradeServices {
		if !p.ServiceConfigs.Has(name) {
			return fmt.Errorf("%s is not defined in the template", name)
		}

		service, err := p.CreateService(name)
		if err != nil {
			return fmt.Errorf("Failed to lookup service: %s: %v", service, err)
		}
		order = append(order, name)
		services[name] = service
	}

	for _, name := range order {
		log.Infof("Start inplace upgrade for service %s...", name)
		err := services[name].Upgrade(ctx, options)
		if err != nil {
			return fmt.Errorf("Failed to inplace upgrade for service %s: %v", name, err)
		}
	}
	return nil
}
