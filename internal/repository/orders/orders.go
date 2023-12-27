package orders

import (
	"context"
	"reflect"
	"strings"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/orders"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrdersRepository interface {
	GetAll(paginationParams *types.PaginationParams, searchParams *models.OrderSearch) (*database.PaginatedResult, error)
	ExceptExchangeBotUser(db *gorm.DB) *gorm.DB
	JoinWithTradeOrders(db *gorm.DB) *gorm.DB
	GroupByOrderID(db *gorm.DB) *gorm.DB
	SelectFieldsWithCommission(fields []string, targetStruct interface{}) func(db *gorm.DB) *gorm.DB
}

type RepoConfig struct {
	ServicePrefix    string
	OrdersTable      string
	TradeOrdersTable string
}

type ordersRepository struct {
	ctx              context.Context
	cancel           context.CancelFunc
	database         *gorm.DB
	repoConfig       *RepoConfig
	logger           *zap.Logger
	builders         builders.BuilderService
	dslSearchEnabled bool
}

func NewGORMOrdersRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (OrdersRepository, error) {
	//database.AutoMigrate(&models.Order{})
	repoConfig := &RepoConfig{
		ServicePrefix:    servicePrefix,
		OrdersTable:      servicePrefix + "_orders",
		TradeOrdersTable: servicePrefix + "_trade_orders"}

	err := config.PrefixService.RegisterServiceTables(servicePrefix, []string{repoConfig.OrdersTable, repoConfig.TradeOrdersTable})
	if err != nil {
		return nil, err
	}

	repository := &ordersRepository{database: database, repoConfig: repoConfig, logger: config.Logger, builders: builders, dslSearchEnabled: true}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (z *ordersRepository) GetAll(paginationParams *types.PaginationParams, searchParams *models.OrderSearch) (*database.PaginatedResult, error) {
	var data []*models.Order
	var count int64

	db := z.database.Debug().Table(z.repoConfig.OrdersTable)

	paginationParams.SortOrder = strings.Replace(paginationParams.SortOrder, "commission", "calculated_commission", -1)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	// dateFields := []string{"created_at", "updated_at", "deleted_at"}
	// z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.OrdersTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		z.ExceptExchangeBotUser,
		z.JoinWithTradeOrders,
		z.GroupByOrderID,
		z.SelectFieldsWithCommission(nil, reflect.TypeOf(models.Order{})),
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.OrdersTable).Scopes(
		whereScope,
		z.ExceptExchangeBotUser,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		z.logger.Error("error counting orders:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}

func (z *ordersRepository) GetUserOrders(paginationParams *types.PaginationParams, searchParams *models.OrderSearch) (*database.PaginatedResult, error) {
	var data []*models.Order
	var count int64

	db := z.database.Debug().Table(z.repoConfig.OrdersTable)

	paginationParams.SortOrder = strings.Replace(paginationParams.SortOrder, "commission", "calculated_commission", -1)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	// dateFields := []string{"created_at", "updated_at", "deleted_at"}
	// z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.OrdersTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		z.ExceptExchangeBotUser,
		z.JoinWithTradeOrders,
		z.GroupByOrderID,
		z.SelectFieldsWithCommission(nil, reflect.TypeOf(models.Order{})),
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.OrdersTable).Scopes(
		whereScope,
		z.ExceptExchangeBotUser,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		z.logger.Error("error counting orders:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}
