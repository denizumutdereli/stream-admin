package user

import (
	"io"
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	service "github.com/denizumutdereli/stream-admin/internal/service/administrator/users"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AdminUserHandler interface {
	CreateSuperAdmin(c *gin.Context)

	// admin users
	CreateAdminUser(c *gin.Context)
	UpdateAdminUser(c *gin.Context)
	DeleteAdminUser(c *gin.Context)
	GetAdminUsers(c *gin.Context)
	VerifyAdminUser(c *gin.Context)
}

type adminUserHandler struct {
	usersService service.AdminUserService
	config       *config.Config
	logger       *zap.Logger
	builders     builders.BuilderService
	mid_         types.QueryParams
}

func NewAdminUserHandler(userService *service.AdminUserService, cfg *config.Config, builders builders.BuilderService) AdminUserHandler {
	return &adminUserHandler{usersService: *userService, config: cfg, logger: cfg.Logger, builders: builders}
}

func (h *adminUserHandler) CreateSuperAdmin(c *gin.Context) {
	err := h.usersService.CreateSuperAdmin()
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Super admin successfully created",
	})
}

func (h *adminUserHandler) CreateAdminUser(c *gin.Context) {
	var adminUser models.AdministratorUser

	if err := c.ShouldBindJSON(&adminUser); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "No data in request body", http.StatusBadRequest)
		} else {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		}
		return
	}

	if err := models.ValidateAdminUser(&adminUser); err != nil {
		h.logger.Error("Validation error", zap.Error(err))

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)
			for _, errField := range validationErrors {
				errorMessages[errField.Field()] = errField.Translate(nil)
			}
			utils.IfErrorExistReturnWithErrorDetails(c, err, "Validation error", errorMessages, http.StatusBadRequest)
		} else {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Validation error", http.StatusBadRequest)
		}
		return
	}

	if !utils.PhoneValidator(adminUser.PhoneNumber) {
		utils.IfErrorExistReturnWithErrorExplanation(c, nil, "Invalid phone number", http.StatusBadRequest)
		return
	}

	added, err := h.usersService.CreateAdminUser(c.Request.Context(), &adminUser)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Admin user successfully created. Let's verify.",
		"data":    added,
	})
}

func (h *adminUserHandler) UpdateAdminUser(c *gin.Context) {
	var adminUser models.AdministratorUser

	if err := c.ShouldBindJSON(&adminUser); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}
		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := models.ValidateAdminUser(&adminUser); err != nil {
		h.logger.Error("Validation error", zap.Error(err))

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)
			for _, errField := range validationErrors {
				errorMessages[errField.Field()] = errField.Translate(nil)
			}
			utils.IfErrorExistReturnWithErrorDetails(c, err, "Validation error", errorMessages, http.StatusBadRequest)
		} else {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Validation error", http.StatusBadRequest)
		}

		return
	}

	updatedPolicy, err := h.usersService.UpdateAdminUser(c.Request.Context(), &adminUser)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Admin user policy successfully updated",
		"data":    updatedPolicy,
	})
}

func (h *adminUserHandler) DeleteAdminUser(c *gin.Context) {
	userID := c.Param("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin User ID is required"})
		return
	}

	err := h.usersService.DeleteAdminUser(c.Request.Context(), userID)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Admin user successfully deleted",
	})
}

func (h *adminUserHandler) GetAdminUsers(c *gin.Context) {
	var queryParams models.AdministratorUserSearch
	dqlQuery := make([]types.QueryCondition, 0)

	bind := h.builders.NewHandleBinding(c, &queryParams, &dqlQuery).BindQuery().BindDSL().BindPagination(&h.mid_.Pagination).Validate()

	if err := bind.GetError(); err != nil {
		if msgs := bind.GetErrorMessages(); len(msgs) > 0 {
			utils.IfErrorExistReturnWithErrorDetails(c, err, "Error in query parameters", msgs, http.StatusBadRequest)
		} else {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Bad request", http.StatusBadRequest)
		}
		return
	}

	queryParams.DSLSearchOperator = &dqlQuery

	paginatedResults, err := h.usersService.GetAdminUsers(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}

func (ac *adminUserHandler) VerifyAdminUser(c *gin.Context) {
	var loginInfo struct {
		Username string `json:"username" binding:"required"`
		OTP      string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginInfo); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}

		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	_, err := ac.usersService.VerifyAdminUser(loginInfo.OTP, loginInfo.Username)

	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account successfully verified"})

}
