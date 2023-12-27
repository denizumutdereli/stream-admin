package handler

import (
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/assets"
	"github.com/denizumutdereli/stream-admin/internal/service"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
)

type AssetsRestHandler interface {
	GetSearchParameters(c *gin.Context)
	GetCoins(c *gin.Context)
	GetAssets(c *gin.Context)
	GetNetworks(c *gin.Context)
}

type assetsRestHandler struct {
	StreamService service.AssetsService
	config        *config.Config
	builders      builders.BuilderService
	mid_          types.QueryParams
}

func NewAssetsRestHandler(s service.AssetsService, cfg *config.Config, builders builders.BuilderService) AssetsRestHandler {
	return &assetsRestHandler{StreamService: s, config: cfg, builders: builders}
}

func (h *assetsRestHandler) GetSearchParameters(c *gin.Context) {
	searchParameters, err := h.StreamService.GetSearchParameters()
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, searchParameters)
}

func (h *assetsRestHandler) GetCoins(c *gin.Context) {
	var queryParams models.AssetsCoinsSearch
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
	paginatedResults, err := h.StreamService.GetCoins(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}

func (h *assetsRestHandler) GetAssets(c *gin.Context) {
	var queryParams models.AssetsSearch
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
	paginatedResults, err := h.StreamService.GetAssets(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}

func (h *assetsRestHandler) GetNetworks(c *gin.Context) {
	var queryParams models.AssetsNetworksSearch
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
	paginatedResults, err := h.StreamService.GetNetworks(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}
