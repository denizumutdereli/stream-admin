package logs

import (
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/administrator"
	logService "github.com/denizumutdereli/stream-admin/internal/service/administrator/logs"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
)

type AdminLogsRestHandler interface {
	GetAll(c *gin.Context)
}

type adminLogsRestHandler struct {
	StreamService logService.AdminLogsService
	config        *config.Config
	builders      builders.BuilderService
	mid_          types.QueryParams
}

func NewAdminLogsRestHandler(s logService.AdminLogsService, cfg *config.Config, builders builders.BuilderService) AdminLogsRestHandler {
	return &adminLogsRestHandler{StreamService: s, config: cfg, builders: builders}
}

func (h *adminLogsRestHandler) GetAll(c *gin.Context) {
	var queryParams models.AdministratorLogsSearch
	dqlQuery := make([]types.QueryCondition, 0)

	bind := h.builders.NewHandleBinding(c, &queryParams, &dqlQuery).BindQuery().BindDSL().BindPagination(&h.mid_.Pagination).Validate()

	if err := bind.GetError(); err != nil {
		if msgs := bind.GetErrorMessages(); len(msgs) > 0 {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Error in query parameters", http.StatusBadRequest)
		} else {
			utils.IfErrorExistReturnWithErrorExplanation(c, err, "Bad request", http.StatusBadRequest)
		}
		return
	}

	queryParams.DSLSearchOperator = &dqlQuery

	paginatedResults, err := h.StreamService.GetAll(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithErrorExplanation(c, err, "error while fetching data", err.StatusCode())
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}
