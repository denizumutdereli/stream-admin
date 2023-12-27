package users

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/database"
	orderModels "github.com/denizumutdereli/stream-admin/internal/models/orders"
	transactionModels "github.com/denizumutdereli/stream-admin/internal/models/transactions"
	models "github.com/denizumutdereli/stream-admin/internal/models/users"
	ordersRepos "github.com/denizumutdereli/stream-admin/internal/repository/orders"
	transactionsRepos "github.com/denizumutdereli/stream-admin/internal/repository/transactions"
	"github.com/denizumutdereli/stream-admin/internal/types"
	userTypes "github.com/denizumutdereli/stream-admin/internal/types/users"
	"go.uber.org/zap"
)

func (z *usersRepository) GetUserDetailsBuilder(
	userId int,
	includeDetails *userTypes.UserDetailsIncluding,
	paginationParams *types.PaginationParams,
	orderRepo ordersRepos.OrdersRepository,
	transactionRepo transactionsRepos.TransactionRepository) (*database.DataResult, error) {
	data, err := z.GetUser(userId)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	errChan := make(chan error, 5)

	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second) // 10 seconds as an example
	defer cancel()                                                 // Important to avoid context leak

	if includeDetails.KYC != nil && *includeDetails.KYC {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := z.LoadKYCInfo(timeoutCtx, data)
			errChan <- err
		}()
	}

	if includeDetails.Banks != nil && *includeDetails.Banks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := z.LoadBankInfo(timeoutCtx, data)
			errChan <- err
		}()
	}

	// if includeDetails.Orders != nil && *includeDetails.Orders {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		err := z.LoadOrdersInfo(timeoutCtx, data, paginationParams, orderRepo)
	// 		errChan <- err
	// 	}()
	// }

	// if includeDetails.FiatTransactions != nil && *includeDetails.FiatTransactions {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		err := z.LoadTransactionsInfo(timeoutCtx, data, paginationParams, transactionRepo, "fiat")
	// 		errChan <- err
	// 	}()
	// }

	// if includeDetails.CryptoTransactions != nil && *includeDetails.CryptoTransactions {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		err := z.LoadTransactionsInfo(timeoutCtx, data, paginationParams, transactionRepo, "crypto")
	// 		errChan <- err
	// 	}()
	// }

	// if includeDetails.CryptoWallets != nil && *includeDetails.CryptoWallets {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		err := z.LoadTransactionsInfo(timeoutCtx, data, paginationParams, transactionRepo, "wallets")
	// 		errChan <- err
	// 	}()
	// }

	if ctxErr := timeoutCtx.Err(); ctxErr != nil {
		z.logger.Error("context finished with error", zap.Error(ctxErr))
		return nil, ctxErr
	}

	wg.Wait()
	close(errChan)

	var errorsSlice []error
	for err := range errChan {
		if err != nil {
			errorsSlice = append(errorsSlice, err)
		}
	}

	if ctxErr := timeoutCtx.Err(); ctxErr != nil {
		z.logger.Error("context finished with error", zap.Error(ctxErr))
		return nil, ctxErr
	}

	if len(errorsSlice) > 0 {
		for _, err := range errorsSlice {
			z.logger.Error("error received processing user details", zap.Error(err))
		}
		return nil, fmt.Errorf("multiple errors occurred in GetUserDetailsBuilder: %v", errorsSlice)
	}

	result := database.SingleDataResult(*data)
	return result, nil
}

func (z *usersRepository) GetUser(userId int) (*models.SearchWithFullJoins, error) {
	var data models.SearchWithFullJoins

	whereScope := fmt.Sprintf("%s.id= ?", z.repoConfig.UserTable)
	db := z.database.Debug().Table(z.repoConfig.UserTable).Where(whereScope, userId)

	query := db.Scopes(
		z.JoinWithUserSettings,
		z.SelectFieldsWithSettings([]string{"user_id,language,theme,currency,favorite_pairs"}, reflect.TypeOf(models.UserSettings{})),
	)

	if err := query.Find(&data).Error; err != nil {
		return nil, errors.New("no user found")
	}

	return &data, nil
}

func (z *usersRepository) LoadKYCInfo(ctx context.Context, data *models.SearchWithFullJoins) error {
	userID := *data.ID

	var KycData []models.UserKYCSearch
	if err := z.database.WithContext(ctx).Table(z.repoConfig.KycTable).Where("user_id = ?", userID).Find(&KycData).Error; err != nil {
		if err == context.DeadlineExceeded {
			z.logger.Error("timeout reached when loading KYC data", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}
		z.logger.Error("error loading KYC data", zap.Int64("user_id", userID), zap.Error(err))
		return err
	}

	for i := range KycData {
		if KycData[i].ID == nil {
			continue
		}

		kycID := *KycData[i].ID
		var kycFiles []models.UserKYCFile
		if err := z.database.WithContext(ctx).Table(z.repoConfig.UserFileTable).Where("kyc_id = ?", kycID).Find(&kycFiles).Error; err != nil {
			if err == context.DeadlineExceeded {
				z.logger.Error("timeout reached when loading KYC files", zap.Int64("kyc_id", kycID), zap.Error(err))
				continue
			}
			z.logger.Error("error loading KYC files", zap.Int64("kyc_id", kycID), zap.Error(err))
			continue
		}

		var filesData []models.FilesOfKYCs
		for _, kycFile := range kycFiles {
			var file models.FilesOfKYCs
			if err := z.database.WithContext(ctx).Table(z.repoConfig.FileServiceTable).Where("id = ?", kycFile.FileID).Find(&file).Error; err != nil {
				if err == context.DeadlineExceeded {
					z.logger.Error("timeout reached when loading file service files", zap.Int64("file_id", kycFile.FileID), zap.Error(err))
					continue
				}
				z.logger.Error("error loading file service files", zap.Int64("file_id", kycFile.FileID), zap.Error(err))
				continue
			}
			filesData = append(filesData, file)
		}

		KycData[i].UserFiles = &filesData
		KycData[i].KYCFiles = &kycFiles
	}
	data.KycData = &KycData
	return nil
}

func (z *usersRepository) LoadBankInfo(ctx context.Context, data *models.SearchWithFullJoins) error {
	userID := *data.ID
	// Bank information  -> TODO: userbank accounts table on fiat_manager has user_id fields as STRING!! -> correct it
	var BankData []models.UserBankAccounts

	// table wrong user_id definition temporarily fixing here!!
	userIDStr := strconv.FormatInt(userID, 10)

	if err := z.database.WithContext(ctx).Table(z.repoConfig.UserBankAccountsTable).Where("user_id = ?", userIDStr).Find(&BankData).Error; err != nil {
		if err == context.DeadlineExceeded {
			z.logger.Error("timeout reached when loading User bank data", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}
		z.logger.Error("error loading User bank data", zap.Int64("user_id", userID), zap.Error(err))
	} else {
		data.UserBanks = &BankData
	}

	return nil
}

func (z *usersRepository) LoadOrdersInfo(
	ctx context.Context,
	data *models.SearchWithFullJoins,
	paginationParams *types.PaginationParams,
	ordersRepo ordersRepos.OrdersRepository) error {
	userID := *data.ID
	var OrderData *database.PaginatedResult

	OrderData, err := ordersRepo.GetAll(paginationParams, &orderModels.OrderSearch{UserID: &userID})
	if err != nil {
		z.logger.Error("error loading User orders data", zap.Int64("user_id", userID), zap.Error(err))
		return err
	}

	data.UserOrders = OrderData

	return nil
}

func (z *usersRepository) LoadTransactionsInfo(
	ctx context.Context,
	data *models.SearchWithFullJoins,
	paginationParams *types.PaginationParams,
	transactionRepo transactionsRepos.TransactionRepository,
	actualType string) error {

	if data.ID == nil {
		return errors.New("data is nil")
	}

	userID := *data.ID
	var TransactionData *database.PaginatedResult
	var err error

	switch actualType {
	case "fiat":
		TransactionData, err = transactionRepo.GetFiatTransactions(paginationParams, &transactionModels.FiatTransactionsSearch{UserID: &userID})
		if err != nil {
			z.logger.Error("error loading fiat transactions data", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}
		data.FiatTransactions = TransactionData
	case "crypto":
		TransactionData, err = transactionRepo.GetCryptoTransactions(paginationParams, &transactionModels.CryptoTransactionsSearch{UserID: &userID})
		if err != nil {
			z.logger.Error("error loading crypto transactions data", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}
		data.CryptoTransactions = TransactionData
	case "wallets": // crypto wallets
		TransactionData, err = transactionRepo.GetCryptoWallets(paginationParams, &transactionModels.CryptoWalletsSearch{UserID: &userID})
		if err != nil {
			z.logger.Error("error loading crypto wallets data", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}
		data.CryptoWallets = TransactionData
	default:
		z.logger.Warn("wrong transaction type on user detail page", zap.Int64("user_id", userID))
	}

	return nil
}
