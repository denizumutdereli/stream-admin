package factory

import (
	"context"

	contextMessage "github.com/denizumutdereli/stream-admin/internal/comm/message"
	"go.uber.org/zap"
)

func (f *serviceFactory) NewAdminContextMessageService(ctx context.Context) (contextMessage.ContextMessages, error) {

	service, err := f.registry.services.RegisterAdminContextMessageService(f.config, f.redis, f.nats)
	if err != nil {
		f.logger.Fatal("service registry error:", zap.Error(err))
		return nil, err
	}

	f.contextMessages = service
	return service, nil
}
