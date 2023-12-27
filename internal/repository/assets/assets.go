package assets

import (
	"context"
	"reflect"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/assets"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AssetsRepository interface {
	GetSearchParameters() ([]types.SearchParameters, error)
	GetCoins(paginationParams *types.PaginationParams, searchParams *models.AssetsCoinsSearch) (*database.PaginatedResult, error)
	GetAssets(paginationParams *types.PaginationParams, searchParams *models.AssetsSearch) (*database.PaginatedResult, error)
	GetNetworks(paginationParams *types.PaginationParams, searchParams *models.AssetsNetworksSearch) (*database.PaginatedResult, error)
}

type RepoConfig struct {
	ServicePrefix string
	CoinsTable    string
	AssetsTable   string
	NetworkTable  string
}

type assetsRepository struct {
	ctx              context.Context
	cancel           context.CancelFunc
	database         *gorm.DB
	repoConfig       *RepoConfig
	logger           *zap.Logger
	builders         builders.BuilderService
	dslSearchEnabled bool
}

func NewGORMAssetsRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (AssetsRepository, error) {
	database.AutoMigrate(&models.AssetsCoins{}, &models.AssetsNetworks{}, &models.Assets{})
	repoConfig := &RepoConfig{
		ServicePrefix: servicePrefix,
		CoinsTable:    servicePrefix + "_coins",
		AssetsTable:   servicePrefix + "_assets",
		NetworkTable:  servicePrefix + "_networks",
	}

	err := config.PrefixService.RegisterServiceTables(servicePrefix, []string{repoConfig.CoinsTable, repoConfig.AssetsTable, repoConfig.NetworkTable})
	if err != nil {
		return nil, err
	}

	repository := &assetsRepository{database: database, repoConfig: repoConfig, logger: config.Logger, builders: builders, dslSearchEnabled: true}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	return repository, nil
}

func (z *assetsRepository) GetSearchParameters() ([]types.SearchParameters, error) {

	var data []types.SearchParameters

	modelsAndTables := []struct {
		modelType reflect.Type
		tableName string
	}{
		{reflect.TypeOf(models.AssetsSearch{}), z.repoConfig.AssetsTable},
		{reflect.TypeOf(models.AssetsCoinsSearch{}), z.repoConfig.CoinsTable},
		{reflect.TypeOf(models.AssetsNetworksSearch{}), z.repoConfig.NetworkTable},
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

func (z *assetsRepository) GetCoins(paginationParams *types.PaginationParams, searchParams *models.AssetsCoinsSearch) (*database.PaginatedResult, error) {
	var data []*models.AssetsCoins
	var count int64

	db := z.database.Debug().Table(z.repoConfig.CoinsTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.CoinsTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.CoinsTable).Scopes(
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

func (z *assetsRepository) GetAssets(paginationParams *types.PaginationParams, searchParams *models.AssetsSearch) (*database.PaginatedResult, error) {
	var data []*models.Assets
	var count int64

	db := z.database.Debug().Table(z.repoConfig.AssetsTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.AssetsTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.AssetsTable).Scopes(
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

func (z *assetsRepository) GetNetworks(paginationParams *types.PaginationParams, searchParams *models.AssetsNetworksSearch) (*database.PaginatedResult, error) {
	var data []*models.AssetsNetworks
	var count int64

	db := z.database.Debug().Table(z.repoConfig.NetworkTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.NetworkTable, z.dslSearchEnabled)

	query := db.Scopes(
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.NetworkTable).Scopes(
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
