package handler

import (
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/orders"
	"github.com/denizumutdereli/stream-admin/internal/service"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
)

type OrdersRestHandler interface {
	GetAll(c *gin.Context)
}

type ordersRestHandler struct {
	StreamService service.OrdersService
	config        *config.Config
	builders      builders.BuilderService
	mid_          types.QueryParams
}

func NewOrdersRestHandler(s service.OrdersService, cfg *config.Config, builders builders.BuilderService) OrdersRestHandler {
	return &ordersRestHandler{StreamService: s, config: cfg, builders: builders}
}

func (h *ordersRestHandler) GetAll(c *gin.Context) {
	var queryParams models.OrderSearch
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
	paginatedResults, err := h.StreamService.GetAll(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}
