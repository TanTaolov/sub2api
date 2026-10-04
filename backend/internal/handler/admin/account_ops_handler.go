package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AccountOpsHandler struct {
	svc     *service.AccountOpsService
	email   *service.EmailService
	gateway *service.OpenAIGatewayService
}

func NewAccountOpsHandler(svc *service.AccountOpsService, email *service.EmailService) *AccountOpsHandler {
	return &AccountOpsHandler{svc: svc, email: email}
}

// SetOpenAIGatewayService injects the service that owns the Prism adapter
// configuration. It is attached after construction, matching the account
// handler, so no generated wire file has to change.
func (h *AccountOpsHandler) SetOpenAIGatewayService(gateway *service.OpenAIGatewayService) {
	if h != nil {
		h.gateway = gateway
	}
}
func (h *AccountOpsHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "Account alert configuration unavailable")
		return
	}
	smtp, err := h.email.GetSMTPConfig(c.Request.Context())
	ready := err == nil && smtp != nil && smtp.Host != "" && smtp.From != ""
	dropped, failures := h.svc.RuntimeCounters()
	response.Success(c, gin.H{"config": cfg, "smtp_configured": ready, "dropped_signals": dropped, "storage_failures": failures})
}
func (h *AccountOpsHandler) SaveConfig(c *gin.Context) {
	var cfg service.AccountOpsConfig
	if c.ShouldBindJSON(&cfg) != nil {
		response.BadRequest(c, "Invalid alert configuration")
		return
	}
	if err := service.ValidateAccountOpsConfig(cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.SaveConfig(c.Request.Context(), cfg); err != nil {
		response.Error(c, http.StatusServiceUnavailable, "Could not save account alert configuration")
		return
	}
	response.Success(c, cfg)
}
func (h *AccountOpsHandler) List(c *gin.Context) {
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		response.BadRequest(c, "Invalid offset")
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		response.BadRequest(c, "Invalid limit")
		return
	}
	events, err := h.svc.List(c.Request.Context(), offset, limit+1)
	if err != nil {
		response.Error(c, http.StatusServiceUnavailable, "Account alerts unavailable")
		return
	}
	more := len(events) > limit
	if more {
		events = events[:limit]
	}
	response.Success(c, gin.H{"items": events, "has_more": more})
}

// PrismStatus reports the gateway-side Prism adapter configuration shown in
// 智能运维. It is a read-only snapshot of startup configuration: it never probes
// the adapter, never replays a request, and never returns the bridge key.
// GET /api/v1/admin/account-ops/prism/status
func (h *AccountOpsHandler) PrismStatus(c *gin.Context) {
	if h == nil || h.gateway == nil {
		response.Error(c, http.StatusServiceUnavailable, "Prism status unavailable")
		return
	}
	response.Success(c, h.gateway.PrismBrowserStatus())
}
