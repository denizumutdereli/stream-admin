package users

import (
	"context"
	"fmt"
	"reflect"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/users"
	ordersRepos "github.com/denizumutdereli/stream-admin/internal/repository/orders"
	"github.com/denizumutdereli/stream-admin/internal/repository/scopes"
	transactionsRepos "github.com/denizumutdereli/stream-admin/internal/repository/transactions"
	"github.com/denizumutdereli/stream-admin/internal/types"
	userTypes "github.com/denizumutdereli/stream-admin/internal/types/users"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UsersRepository interface {
	loadExternalTables(ctx context.Context) error
	GetSearchUserParameters() ([]types.SearchParameters, error)
	GetSearchKYCParameters() ([]types.SearchParameters, error)
	JoinWithUserSettings(db *gorm.DB) *gorm.DB
	JoinWithUserKYCData(db *gorm.DB) *gorm.DB
	SelectFieldsWithSettings(fields []string, targetStruct interface{}) func(db *gorm.DB) *gorm.DB
	GroupByUserIDWithSettings(db *gorm.DB) *gorm.DB
	GetUsers(paginationParams *types.PaginationParams, searchParams *models.UserSearch) (*database.PaginatedResult, error)
	GetUserDetailsBuilder(userId int, includeDetails *userTypes.UserDetailsIncluding, paginationParams *types.PaginationParams, orderRepo ordersRepos.OrdersRepository, transactionRepo transactionsRepos.TransactionRepository) (*database.DataResult, error)
	GetKYC(paginationParams *types.PaginationParams, searchParams *models.UserKYCSearch) (*database.PaginatedResult, error)
}

type RepoConfig struct {
	ServicePrefix         string
	UserTable             string
	UserFileTable         string
	UserSettingsTable     string
	MerchantTable         string
	KycTable              string
	FileServiceTable      string
	UserBankAccountsTable string
	TradeOrdersTable      string
}

type usersRepository struct {
	ctx              context.Context
	cancel           context.CancelFunc
	config           *config.Config
	database         *gorm.DB
	repoConfig       *RepoConfig
	logger           *zap.Logger
	builders         builders.BuilderService
	dslSearchEnabled bool
}

func NewGORMUsersRepository(database *gorm.DB, servicePrefix string, config *config.Config, builders builders.BuilderService) (UsersRepository, error) {
	//database.AutoMigrate(&models.User{})
	repoConfig := &RepoConfig{
		ServicePrefix:     servicePrefix,
		UserTable:         servicePrefix + "_user",
		UserFileTable:     servicePrefix + "_user_file",
		UserSettingsTable: servicePrefix + "_user_settings",
		KycTable:          servicePrefix + "_kyc",
	}

	prefixErr := config.PrefixService.RegisterServiceTables(servicePrefix,
		[]string{
			repoConfig.UserTable,
			repoConfig.UserFileTable,
			repoConfig.UserFileTable,
			repoConfig.UserSettingsTable,
			repoConfig.KycTable})
	if prefixErr != nil {
		return nil, prefixErr
	}

	repository := &usersRepository{database: database, config: config, repoConfig: repoConfig, logger: config.Logger, builders: builders, dslSearchEnabled: true}
	ctx, cancel := context.WithCancel(context.Background())
	repository.ctx = ctx
	repository.cancel = cancel

	err := repository.loadExternalTables(ctx)
	if err != nil {
		repository.logger.Fatal("Failed to load external tables")
	}

	return repository, nil
}

func (z *usersRepository) loadExternalTables(ctx context.Context) error {
	file_service, err := z.config.PrefixService.GetServicePrefix("file_service")
	if !err {
		errMsg := "error getting file_service prefix"
		z.logger.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	z.repoConfig.FileServiceTable = file_service + "_file"

	fiat_manager, err := z.config.PrefixService.GetServicePrefix("fiat_manager")
	if !err {
		errMsg := "error getting fiat_manager prefix"
		z.logger.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	z.repoConfig.UserBankAccountsTable = fiat_manager + "_user_bank_accounts"

	order_service, err := z.config.PrefixService.GetServicePrefix("orders")
	if !err {
		errMsg := "error getting orders prefix"
		z.logger.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	z.repoConfig.TradeOrdersTable = order_service + "_trade_orders"

	return nil
}

func (z *usersRepository) GetSearchUserParameters() ([]types.SearchParameters, error) {

	var data []types.SearchParameters

	modelsAndTables := []struct {
		modelType reflect.Type
		tableName string
	}{
		{reflect.TypeOf(models.UserSearch{}), z.repoConfig.UserTable},
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
func (z *usersRepository) GetSearchKYCParameters() ([]types.SearchParameters, error) {

	var data []types.SearchParameters

	modelsAndTables := []struct {
		modelType reflect.Type
		tableName string
	}{
		{reflect.TypeOf(models.UserKYCSearch{}), z.repoConfig.KycTable},
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

func (z *usersRepository) GetUsers(paginationParams *types.PaginationParams, searchParams *models.UserSearch) (*database.PaginatedResult, error) {
	var data []models.SearchWithSettings
	var count int64

	db := z.database.Debug().Table(z.repoConfig.UserTable)

	// temporary date format fix. Will be removed when dbs are ready and synchronized with the date types -->
	dateFields := []string{"created_at", "updated_at", "deleted_at"}
	z.builders.ConvertDateFields(dateFields, z.builders.StructToMap(searchParams), "toUnix")
	// <--

	whereScope := scopes.ApplySearchFilters(searchParams, z.repoConfig.UserTable, z.dslSearchEnabled)

	query := db.Scopes(
		z.JoinWithUserSettings,
		z.SelectFieldsWithSettings([]string{"user_id,language,theme,currency,favorite_pairs"}, reflect.TypeOf(models.UserSettings{})),
		whereScope,
		scopes.OrderBy(paginationParams.SortBy, paginationParams.SortOrder),
	)

	countQuery := z.database.Table(z.repoConfig.UserTable).Scopes(
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
