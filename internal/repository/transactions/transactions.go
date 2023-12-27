package transactions

import (
	"context"
	"reflect"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/transactions"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TransactionRepository interface {
	GetSearchParameters() ([]types.SearchParameters, error)
	GetFiatTransactions(paginationParams *types.PaginationParams, searchParams *models.FiatTransactionsSearch) (*database.PaginatedResult, error)
	GetCryptoTransactions(paginationParams *types.PaginationParams, searchParams *models.CryptoTransactionsSearch) (*database.PaginatedResult, error)
	GetCryptoWallets(paginationParams *types.PaginationParams, searchParams *models.CryptoWalletsSearch) (*database.PaginatedResult, error)
}

type RepoConfig struct {
	ServicePrefix           string
	FiatTransactionsTable   string
	CryptoTransactionsTable string
	CryptoWalletsTable      string
}

type transactionRepository struct {
	ctx              context.Context
	cancel           context.CancelFunc
	database         *gorm.DB
	repoConfig       *RepoConfig
	logger           *zap.Logger
	builders         builders.BuilderService
	dslSearchEnabled bool
}

func NewGORMTransactionsRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (TransactionRepository, error) {
	database.AutoMigrate(&models.CryptoTransactions{}, &models.FiatTransactions{})
	repoConfig := &RepoConfig{
		ServicePrefix:           servicePrefix,
		FiatTransactionsTable:   servicePrefix + "_fiat_transactions",
		CryptoTransactionsTable: servicePrefix + "_crypto_transactions",
		CryptoWalletsTable:      servicePrefix + "_addresses",
	}

	err := config.PrefixService.RegisterServiceTables(servicePrefix,
		[]string{
			repoConfig.FiatTransactionsTable,
			repoConfig.CryptoTransactionsTable,
		})
	if err != nil {
		return nil, err
	}

	repository := &transactionRepository{database: database, repoConfig: repoConfig, logger: config.Logger, builders: builders, dslSearchEnabled: true}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (z *transactionRepository) GetSearchParameters() ([]types.SearchParameters, error) {

	var data []types.SearchParameters

	modelsAndTables := []struct {
		modelType reflect.Type
		tableName string
	}{
		{reflect.TypeOf(models.FiatTransactionsSearch{}), z.repoConfig.FiatTransactionsTable},
		{reflect.TypeOf(models.CryptoTransactionsSearch{}), z.repoConfig.CryptoTransactionsTable},
		{reflect.TypeOf(models.CryptoWalletsSearch{}), z.repoConfig.CryptoWalletsTable},
	}

	for _, mt := range modelsAndTables {
		searchParams, err := z.builders.ConstructSearchParameters(mt.modelType, mt.tableName)
		if err != nil {
			return nil, err
		}
		data = append(data, searchParams)
	}

	return data, nil
}

func (z *transactionRepository) GetFiatTransactions(paginationParams *types.PaginationParams, searchParams *models.FiatTransactionsSearch) (*database.PaginatedResult, error) {
	var data []models.FiatTransactions
	var count int64

	db := z.database.Debug().Table(z.repoConfig.FiatTransactionsTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.FiatTransactionsTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.FiatTransactionsTable).Scopes(
		whereScope,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		z.logger.Error("error counting data:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}

func (z *transactionRepository) GetCryptoTransactions(paginationParams *types.PaginationParams, searchParams *models.CryptoTransactionsSearch) (*database.PaginatedResult, error) {
	var data []models.CryptoTransactions
	var count int64

	db := z.database.Debug().Table(z.repoConfig.CryptoTransactionsTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.CryptoTransactionsTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.CryptoTransactionsTable).Scopes(
		whereScope,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		z.logger.Error("error counting data:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}

func (z *transactionRepository) GetCryptoWallets(paginationParams *types.PaginationParams, searchParams *models.CryptoWalletsSearch) (*database.PaginatedResult, error) {
	var data []models.CryptoWallets
	var count int64

	db := z.database.Debug().Table(z.repoConfig.CryptoWalletsTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.CryptoWalletsTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.CryptoTransactionsTable).Scopes(
		whereScope,
	)

	if err := countQuery.Count(&count).Error; err != nil {
		z.logger.Error("error counting data:", zap.Error(err))
	}

	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query = query.Offset(offset).Limit(paginationParams.Limit)

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	paginatedResults := database.PaginateTheResults(data, count, offset, paginationParams.Page, paginationParams.Limit)

	return paginatedResults, nil
}
