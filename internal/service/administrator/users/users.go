package users

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/denizumutdereli/stream-admin/internal/caesar"
	appErrors "github.com/denizumutdereli/stream-admin/internal/common"
	"github.com/denizumutdereli/stream-admin/internal/config"
	"github.com/denizumutdereli/stream-admin/internal/database"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	users "github.com/denizumutdereli/stream-admin/internal/repository/administrator/users"
	"github.com/denizumutdereli/stream-admin/internal/transport"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"go.uber.org/zap"
)

type AdminUserService interface {
	CreateSuperAdmin() appErrors.Error

	// admin users
	CreateAdminUser(ctx context.Context, admin_user *models.AdministratorUser) (*models.AdministratorUser, appErrors.Error)
	UpdateAdminUser(ctx context.Context, adminRolePolicy *models.AdministratorUser) (*models.AdministratorUser, appErrors.Error)
	DeleteAdminUser(ctx context.Context, userID string) appErrors.Error

	GetAdminUsers(paginationParams *types.PaginationParams, queryParams *models.AdministratorUserSearch) (*database.PaginatedResult, appErrors.Error)
	VerifyAdminUser(userOTP, username string) (bool, appErrors.Error)
	GetAdminActiveUsersVPNAddresses() ([]string, appErrors.Error)
	InitiateOTP(phone, username string, debug bool) (string, appErrors.Error)
	VerifyOTP(phone_key, userOTP string) (bool, appErrors.Error)

	SetUserLock(username string) (bool, appErrors.Error)
}

type adminUsersService struct {
	ctx         context.Context
	cancel      context.CancelFunc
	userRepo    users.AdminUsersRepository
	config      *config.Config
	logger      *zap.Logger
	caesar      caesar.CaesarManager
	restClient  *transport.Client
	redisClient *transport.RedisManager
	debug       bool
}

func NewAdminUsersService(userRepo *users.AdminUsersRepository, caesar caesar.CaesarManager, redis *transport.RedisManager, config *config.Config) AdminUserService {
	service := &adminUsersService{
		userRepo:    *userRepo,
		config:      config,
		logger:      config.Logger,
		caesar:      caesar,
		restClient:  transport.NewRestClient(config.OTPServiceApi, config.Logger),
		redisClient: redis,
		debug:       false,
	}

	if service.config.Test == "true" {
		service.debug = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	service.ctx = ctx
	service.cancel = cancel

	return service
}

/* Admin setup ------------------------------------------------------------------------------------------------------ */

func (s *adminUsersService) CreateSuperAdmin() appErrors.Error {
	err := s.userRepo.CreateInitialSuperAdmin()
	if err != nil {
		s.logger.Error("error creating admin user", zap.Error(err))
		return appErrors.AppError(http.StatusServiceUnavailable, "", "error creating super admin", err)
	}
	return nil
}

/* Admin users ------------------------------------------------------------------------------------------------------ */

func (s *adminUsersService) CreateAdminUser(ctx context.Context, admin_user *models.AdministratorUser) (*models.AdministratorUser, appErrors.Error) {
	isNew, err := s.userRepo.CreateAdminUser(admin_user)
	if err != nil {
		s.logger.Error("error creating admin user", zap.Error(err))
		return nil, appErrors.AppError(http.StatusBadRequest, "", "error creating admin user", err)
	}

	if !isNew {
		return nil, appErrors.AppError(http.StatusConflict, "", "user already exists", nil)
	}

	_, err = s.InitiateOTP(admin_user.PhoneNumber, admin_user.Username, s.debug)
	if err != nil {
		return &models.AdministratorUser{}, appErrors.AppError(http.StatusServiceUnavailable, "", "otp code could not send", nil)
	}

	return admin_user, nil
}

func (s *adminUsersService) UpdateAdminUser(ctx context.Context, adminUser *models.AdministratorUser) (*models.AdministratorUser, appErrors.Error) {

	response, err := s.userRepo.UpdateAdminUser(adminUser)
	if err != nil {
		s.logger.Error("error updating admin user", zap.Error(err))
		return nil, appErrors.AppError(http.StatusInternalServerError, "", "error updating user", err)
	}

	return response, nil
}

func (s *adminUsersService) DeleteAdminUser(ctx context.Context, policyID string) appErrors.Error {
	_, err := s.userRepo.DeleteAdminUser(policyID)
	if err != nil {
		s.logger.Error("error deleting admin user", zap.Error(err))
		return appErrors.AppError(http.StatusInternalServerError, "", "error deleting admin user", err)
	}

	return nil
}

func (s *adminUsersService) GetAdminUsers(paginationParams *types.PaginationParams, queryParams *models.AdministratorUserSearch) (*database.PaginatedResult, appErrors.Error) {
	data, err := s.userRepo.GetAdminUsers(paginationParams, queryParams)
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *adminUsersService) VerifyAdminUser(userOTP, username string) (bool, appErrors.Error) {

	user, err := s.userRepo.FindAdminUserByAdminUsername(username)
	if err != nil {
		return false, appErrors.AppError(http.StatusConflict, "", "credentials are not valid", err)
	}

	phone_key := fmt.Sprintf("%s%s%s", user.PhoneNumber, username, "_admin_verify")

	_, err = s.VerifyOTP(phone_key, userOTP)
	if err != nil {
		return false, appErrors.AppError(http.StatusConflict, "", "invalid or expired otp", err)
	}

	err = s.redisClient.DeleteKey(context.Background(), phone_key)
	if err != nil {
		return false, appErrors.AppError(http.StatusInternalServerError, "", "unable to delete redis key", err)
	}

	user.Status = models.UserStatusVerified
	_, err = s.userRepo.UpdateAdminUser(&user)

	if err != nil {
		return false, appErrors.AppError(http.StatusServiceUnavailable, "", "unable to update admin user verify", err)
	}

	return true, nil
}

func (s *adminUsersService) GetAdminActiveUsersVPNAddresses() ([]string, appErrors.Error) {
	data, err := s.userRepo.GetAdminActiveUsersVPNAddresses()
	if err != nil {
		return nil, appErrors.AppError(http.StatusInternalServerError, "", err.Error(), err)
	}

	return data, nil
}

func (s *adminUsersService) InitiateOTP(phone, username string, debug bool) (string, appErrors.Error) {
	var otp string
	var err error

	if !debug {
		otp, err = s.caesar.GenerateOTP()
		if err != nil {
			s.logger.Error("Error generating OTP", zap.Error(err))
			return "", appErrors.AppError(http.StatusServiceUnavailable, "", "error generating otp code", err)
		}

		type SMSServiceMessage struct {
			Recipient string `json:"recipient"`
			Message   string `json:"message"`
		}

		requestBody := &SMSServiceMessage{
			Recipient: phone,
			Message:   fmt.Sprintf("Your %s code is: %s", s.config.AppName, otp),
		}

		headers := make(map[string]string)
		headers["Content-Type"] = "application/json"
		headers["X-Secret"] = s.config.OTPServiceKey
		response, err := s.restClient.DoRequest("POST", "", requestBody, headers)
		if err != nil {
			return "", appErrors.AppError(http.StatusServiceUnavailable, "", "OTP service call problem", err)
		}

		if response.StatusCode != http.StatusOK {
			s.logger.Error("SMS service failed to send the message", zap.Int("status", response.StatusCode), zap.String("response:", string(response.Body)))
			return "", appErrors.AppError(http.StatusServiceUnavailable, "", "failed to send the OTP code", nil)
		}
	} else {
		otp = "999900" // ref config Test bool
	}

	phone_key := fmt.Sprintf("%s%s%s", phone, username, "_admin_verify")

	err = s.caesar.StoreOTP(phone_key, otp, time.Now().Add(time.Duration(s.config.OTPCodesInMinutes+1)*time.Minute))
	if err != nil {
		return "", appErrors.AppError(http.StatusServiceUnavailable, "", "error storing the OTP code", err)
	}

	return otp, nil
}

func (s *adminUsersService) VerifyOTP(phone_key, userOTP string) (bool, appErrors.Error) {
	validOTP, err := s.caesar.RetrieveOTP(phone_key)
	if err != nil {
		return false, appErrors.AppError(http.StatusServiceUnavailable, "", "error retrieving the OTP code from service", err)
	}

	fmt.Println("--", userOTP, validOTP, phone_key, "-->")

	if userOTP != validOTP {
		return false, appErrors.AppError(http.StatusConflict, "", "otp code is incorrect", nil)
	}

	return true, nil
}

func (s *adminUsersService) SetUserLock(username string) (bool, appErrors.Error) {
	// TODO: record events on nats and wsserver
	findUser, err := s.userRepo.FindAdminUserByAdminUsername(username)
	if err != nil {
		return false, appErrors.AppError(http.StatusInternalServerError, "", "error retrieving the user with username", err)
	}

	operationCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.DefaultFuncsTimeOutInSeconds)*time.Second)
	defer cancel()

	lockKey := fmt.Sprintf("user-lock:%s", findUser.UserID)

	err = s.redisClient.SetKeyValue(operationCtx, lockKey, true,
		time.Duration(s.config.DefaultPanelLockPeriodInMinutes)*time.Minute)
	if err != nil {
		s.logger.Error("Error setting user lock", zap.String("userId", findUser.UserID))
		return false, appErrors.AppError(http.StatusServiceUnavailable, "", "Error setting user lock", err)
	}

	return true, nil
}

/* ------------------------------------------------------------------------------------------------------------------ */
