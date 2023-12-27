package policy

import (
	"io"
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	rolePolicyModels "github.com/denizumutdereli/stream-admin/internal/models/administrator/policy"
	service "github.com/denizumutdereli/stream-admin/internal/service/administrator/policy"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AdminPolicyHandler interface {
	CreateAdminRolePolicy(c *gin.Context)
	UpdateAdminRolePolicy(c *gin.Context)
	DeleteAdminRolePolicy(c *gin.Context)
	GetAdminRolePolicies(c *gin.Context)
}

type adminPolicyHandler struct {
	policyService service.AdminPolicyService
	config        *config.Config
	logger        *zap.Logger
	builders      builders.BuilderService
	mid_          types.QueryParams
}

func NewAdminPolicyHandler(policyService *service.AdminPolicyService, cfg *config.Config, builders builders.BuilderService) AdminPolicyHandler {
	return &adminPolicyHandler{policyService: *policyService, config: cfg, logger: cfg.Logger, builders: builders}
}

func (h *adminPolicyHandler) CreateAdminRolePolicy(c *gin.Context) {
	var adminRolePolicy rolePolicyModels.AdministratorRolePolicy

	if err := c.ShouldBindJSON(&adminRolePolicy); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}

		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := rolePolicyModels.ValidateAdminRolePolicy(&adminRolePolicy); err != nil {
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

	added, err := h.policyService.CreateAdminRolePolicy(c.Request.Context(), &adminRolePolicy)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"message": "Admin role policy successfully created",
		"data":    added,
	})
}

func (h *adminPolicyHandler) UpdateAdminRolePolicy(c *gin.Context) {
	var adminRolePolicy rolePolicyModels.AdministratorRolePolicy

	if err := c.ShouldBindJSON(&adminRolePolicy); err != nil {
		if err == io.EOF {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Request body is empty", http.StatusBadRequest)
			return
		}
		utils.IfErrorExistReturnWithErrorExplanation(c, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := rolePolicyModels.ValidateAdminRolePolicy(&adminRolePolicy); err != nil {
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

	updatedPolicy, err := h.policyService.UpdateAdminRolePolicy(c.Request.Context(), &adminRolePolicy)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Admin role policy successfully updated",
		"data":    updatedPolicy,
	})
}

func (h *adminPolicyHandler) DeleteAdminRolePolicy(c *gin.Context) {
	policyID := c.Param("policy_id")

	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}

	err := h.policyService.DeleteAdminRolePolicy(c.Request.Context(), policyID)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Admin role policy successfully deleted",
	})
}

func (h *adminPolicyHandler) GetAdminRolePolicies(c *gin.Context) {
	var queryParams rolePolicyModels.AdministratorRolePolicySearch
	dqlQuery := make([]types.QueryCondition, 0)

	bind := h.builders.NewHandleBinding(c, &queryParams, &dqlQuery).BindQuery().BindDSL().BindPagination(&h.mid_.Pagination).Validate()

	if err := bind.GetError(); err != nil {
		if msgs := bind.GetErrorMessages(); len(msgs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"errors": msgs})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	queryParams.DSLSearchOperator = &dqlQuery

	paginatedResults, err := h.policyService.GetAdminRolePolicies(&h.mid_.Pagination, &queryParams)

	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
	}

	c.JSON(http.StatusOK, paginatedResults)
}
