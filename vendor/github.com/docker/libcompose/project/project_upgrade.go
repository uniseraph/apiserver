package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
)

// Up creates and starts the specified services (kinda like docker run).
func (p *Project) Upgrade(ctx context.Context, options options.Upgrade, services ...string) error {
	if err := p.initialize(ctx); err != nil {
		return err
	}
	return p.perform(events.ProjectUpStart, events.ProjectUpDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceUpStart, events.ServiceUp, func(service Service) error {
			return service.Upgrade(ctx, options)
		})
	}), func(service Service) error {
		return service.UpgradeCreate(ctx, options.Create)
	})
}
