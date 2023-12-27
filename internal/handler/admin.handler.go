package handler

import (
	"net/http"

	"github.com/denizumutdereli/stream-admin/internal/config"
	adminService "github.com/denizumutdereli/stream-admin/internal/service"
	"github.com/denizumutdereli/stream-admin/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type adminRestHander struct {
	StreamService adminService.AdminService
	Config        *config.Config
}

type AdminRestHandler interface {
	Live(c *gin.Context)
	Read(c *gin.Context)
	Metrics(c *gin.Context)
	Configs(c *gin.Context)
}

func NewAdminRestHandler(s adminService.AdminService, cfg *config.Config) AdminRestHandler {
	return &adminRestHander{StreamService: s, Config: cfg}
}

func (h *adminRestHander) Live(c *gin.Context) {
	statusCode, ok, err := h.StreamService.Live(c.Request.Context())
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(statusCode, gin.H{"status": ok})
}

func (h *adminRestHander) Read(c *gin.Context) {
	statusCode, ok, err := h.StreamService.Live(c.Request.Context())
	if err != nil {
		utils.IfErrorExistReturnWithError(c, err)
		return
	}

	c.JSON(statusCode, gin.H{"status": ok})
}

func (s *adminRestHander) Metrics(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}

func (s *adminRestHander) Configs(c *gin.Context) {
	fullConfig := s.Config

	type ConfigResponse struct {
		Status bool           `json:"status"`
		Data   *config.Config `json:"data"`
	}

	c.JSON(http.StatusOK, &ConfigResponse{Status: true, Data: fullConfig})
}
