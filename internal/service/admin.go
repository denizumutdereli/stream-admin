package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/repository"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type AdminServiceStatuses struct {
	redisConnected    bool
	databaseConnected bool
}

type AdminService interface {
	MonitorServices(ctx context.Context)
	Live(ctx context.Context) (int, bool, appErrors.Error)
	Read(ctx context.Context) (int, bool, appErrors.Error)
	SetIsLeader(ctx context.Context, isLeader bool)
}

type adminService struct {
	config   *config.Config
	logger   *zap.Logger
	status   *AdminServiceStatuses
	redis    *transport.RedisManager
	repo     repository.AdminRepository
	isLeader bool
}

func NewAdminService(appContext *types.ExchangeConfig, repo *repository.AdminRepository) AdminService {
	service := &adminService{
		config:   appContext.Config,
		redis:    appContext.Redis,
		logger:   appContext.Logger,
		repo:     *repo,
		status:   &AdminServiceStatuses{redisConnected: true, databaseConnected: true},
		isLeader: false,
	}

	//ctx := context.Background()

	//go service.MonitorServices(ctx)

	return service
}

func (es *adminService) SetIsLeader(ctx context.Context, isLeader bool) {
	es.isLeader = isLeader
}

func (s *adminService) MonitorServices(ctx context.Context) {
	s.logger.Debug("## Service monitoring started")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			s.status.redisConnected = s.redis.IsConnected()
			s.status.databaseConnected = s.repo.IsConnected()

			if !(s.status.redisConnected && s.status.databaseConnected) {
				s.logger.Debug("service status:", zap.Bool("Redis", s.status.redisConnected), zap.Bool("Database", s.status.databaseConnected))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *adminService) Live(ctx context.Context) (int, bool, appErrors.Error) {
	serviceStatus := fmt.Sprintf("Redis:%v Database: %v", s.status.redisConnected, s.status.databaseConnected)
	if s.status.redisConnected && s.status.databaseConnected {
		return http.StatusOK, true, nil
	}
	return http.StatusServiceUnavailable, false, appErrors.AppError(http.StatusServiceUnavailable, "", fmt.Sprintf("services are not fully operational %s", serviceStatus), nil)
}

func (s *adminService) Read(ctx context.Context) (int, bool, appErrors.Error) {
	return s.Live(ctx)
}
