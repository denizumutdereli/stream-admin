package factory

import (
	"github.com/denizumutdereli/stream-admin/internal/service/administrator/stream"
	"github.com/denizumutdereli/stream-admin/internal/wsserver"
)

func (f *serviceFactory) NewWsServer() *wsserver.Server {
	wsServer := wsserver.NewServer(f.appContext)
	f.wsserver = wsServer
	return wsServer
}

func (f *serviceFactory) NewStreamAssetsService() (*stream.AssetsService, error) {
	return stream.NewAssetsService(f.config, f.logger, f.redis)
}
