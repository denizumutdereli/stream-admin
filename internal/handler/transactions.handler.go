package handler

import (
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/builders"
	"github.com/denizumutdereli/stream-admin/internal/config"
	models "github.com/denizumutdereli/stream-admin/internal/models/transactions"
	"github.com/denizumutdereli/stream-admin/internal/service"
	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
)

type TransactionsRestHandler interface {
	GetSearchParameters(c *gin.Context)
	GetFiatTransactions(c *gin.Context)
	GetCryptoTransactions(c *gin.Context)
	GetCryptoWallets(c *gin.Context)
}

type transactionsRestHandler struct {
	StreamService service.TransactionService
	config        *config.Config
	builders      builders.BuilderService
	mid_          types.QueryParams
}

func NewTransactionsRestHandler(s service.TransactionService, cfg *config.Config, builders builders.BuilderService) TransactionsRestHandler {
	return &transactionsRestHandler{StreamService: s, config: cfg, builders: builders, mid_: types.QueryParams{}}
}

func (h *transactionsRestHandler) GetSearchParameters(c *gin.Context) {
	searchParameters, err := h.StreamService.GetSearchParameters()
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, searchParameters)
}

func (h *transactionsRestHandler) GetFiatTransactions(c *gin.Context) {
	var queryParams models.FiatTransactionsSearch
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
	paginatedResults, err := h.StreamService.GetFiatTransactions(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}

func (h *transactionsRestHandler) GetCryptoTransactions(c *gin.Context) {
	var queryParams models.CryptoTransactionsSearch
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
	paginatedResults, err := h.StreamService.GetCryptoTransactions(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}

func (h *transactionsRestHandler) GetCryptoWallets(c *gin.Context) {
	var queryParams models.CryptoWalletsSearch
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
	paginatedResults, err := h.StreamService.GetCryptoWallets(&h.mid_.Pagination, &queryParams)
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, paginatedResults)
}
