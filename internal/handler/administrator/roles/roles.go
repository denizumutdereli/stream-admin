package roles

import (
	"io"
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	rolesService "github.com/denizumutdereli/stream-admin/internal/service/administrator/roles"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.uber.org/zap"
)

type AdminUserRolesHandler interface {
	CreateAdminRole(c *gin.Context)
	GetAdminRoles(c *gin.Context)
	AttachPoliciesToRole(c *gin.Context)
}

type adminUserRolesHandler struct {
	rolesService rolesService.AdminUserRolesService
	config       *config.Config
	logger       *zap.Logger
	builders     builders.BuilderService
	mid_         types.QueryParams
}

func NewAdminUserRolesHandler(userService *rolesService.AdminUserRolesService, cfg *config.Config, builders builders.BuilderService) AdminUserRolesHandler {
	return &adminUserRolesHandler{rolesService: *userService, config: cfg, logger: cfg.Logger, builders: builders}
}

func (h *adminUserRolesHandler) CreateAdminRole(c *gin.Context) {
	var adminRole models.AdministratorRole

	if err := c.ShouldBindJSON(&adminRole); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}

		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := models.ValidateAdminRole(&adminRole); err != nil {
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

	added, err := h.rolesService.CreateAdminRole(c.Request.Context(), &adminRole)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Admin role successfully created",
		"data":    added,
	})
}

func (h *adminUserRolesHandler) GetAdminRoles(c *gin.Context) {
	var queryParams models.AdministratorRoleSearch
	dqlQuery := make([]types.QueryCondition, 0)

	bind := h.builders.NewHandleBinding(c, &queryParams, &dqlQuery).BindQuery().BindDSL().BindPagination(&h.mid_.Pagination).Validate()

	if err := bind.GetError(); err != nil {
		if msgs := bind.GetErrorMessages(); len(msgs) > 0 {
			utils.IfErrorExistReturnWithErrorDetails(c, err, "Validation error", msgs, http.StatusBadRequest)
		} else {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Bad request", http.StatusBadRequest)
		}
		return
	}

	queryParams.DSLSearchOperator = &dqlQuery

	paginatedResults, err := h.rolesService.GetAdminRoles(&h.mid_.Pagination, &queryParams)

	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}

func (h *adminUserRolesHandler) AttachPoliciesToRole(c *gin.Context) {
	var request struct {
		RoleID    string   `json:"role_id"`
		PolicyIDs []string `json:"policy_ids"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}

		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	role, err := h.rolesService.AttachPoliciesToRole(c.Request.Context(), request.RoleID, request.PolicyIDs)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Policies successfully attached to role",
		"data":    role,
	})
}
