package handler

import (
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/users"
	"github.com/denizumutdereli/stream-admin/internal/service"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
)

type UsersRestHandler interface {
	GetSearchUserParameters(c *gin.Context)
	GetSearchKYCParameters(c *gin.Context)
	GetUsers(c *gin.Context)
	// GetUserDetailsBuilder(c *gin.Context)
	GetKYC(c *gin.Context)
}

type usersRestHandler struct {
	StreamService service.UsersService
	config        *config.Config
	builders      builders.BuilderService
	mid_          types.QueryParams
}

func NewUsersRestHandler(s service.UsersService, cfg *config.Config, builders builders.BuilderService) UsersRestHandler {
	return &usersRestHandler{StreamService: s, config: cfg, builders: builders, mid_: types.QueryParams{}}
}

func (h *usersRestHandler) GetSearchUserParameters(c *gin.Context) {
	searchParameters, err := h.StreamService.GetSearchUserParameters()
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, searchParameters)
}

func (h *usersRestHandler) GetSearchKYCParameters(c *gin.Context) {
	searchParameters, err := h.StreamService.GetSearchKYCParameters()
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, searchParameters)
}

func (h *usersRestHandler) GetUsers(c *gin.Context) {
	var queryParams models.UserSearch
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
	paginatedResults, err := h.StreamService.GetUsers(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}

// func (h *usersRestHandler) GetUserDetailsBuilder(c *gin.Context) {
// 	userIdStr := c.Param("user_id")

// 	userId, err := strconv.Atoi(userIdStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
// 		return
// 	}

// 	var queryParams models.SearchWithFullJoins
// 	dqlQuery := make([]types.QueryCondition, 0)
// 	bind := h.builders.NewHandleBinding(c, &queryParams, &dqlQuery).BindQuery().BindDSL().BindPagination(&h.mid_.Pagination).Validate()

// 	if err := bind.GetError(); err != nil {
// 		if msgs := bind.GetErrorMessages(); len(msgs) > 0 {
// 			c.JSON(http.StatusBadRequest, gin.H{"errors": msgs})
// 		} else {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		}
// 		return
// 	}

// 	queryParams.DSLSearchOperator = &dqlQuery

// 	includeDetails := userTypes.NewUserDetailsIncludingWithDefaults()
// 	// TODO -> bind.go
// 	if err := c.ShouldBindQuery(&includeDetails); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
// 		return
// 	}

// 	result, err := h.StreamService.GetUserDetailsBuilder(userId, includeDetails, &h.mid_.Pagination)

// 	if err != nil {
// 		fmt.Println(err)
// 		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "message": "No data found matching the criteria", "error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, result)
// }

func (h *usersRestHandler) GetKYC(c *gin.Context) {
	var queryParams models.UserKYCSearch
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
	paginatedResults, err := h.StreamService.GetKYC(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}
