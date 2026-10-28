package admin

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func profitURLRemarkError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrUpstreamURLRemarkValidation) || errors.Is(err, service.ErrUpstreamURLRemarkConflict) {
		response.ErrorFrom(c, err)
	} else {
		response.InternalError(c, "URL 渠道备注操作失败，请重试")
	}
}
func (h *ChannelHandler) ReadProfitURLRemark(c *gin.Context) {
	out, err := h.channelService.ReadUpstreamURLRemark(c.Request.Context(), c.Query("upstream"))
	if err != nil {
		profitURLRemarkError(c, err)
		return
	}
	response.Success(c, out)
}
func (h *ChannelHandler) SaveProfitURLRemark(c *gin.Context) {
	var req struct {
		Upstream string  `json:"upstream"`
		Remark   *string `json:"remark"`
		Revision *int64  `json:"revision"`
	}
	// Both empty text and revision zero are meaningful; omission is an error.
	if c.ShouldBindJSON(&req) != nil || req.Remark == nil {
		response.ErrorFrom(c, service.ErrUpstreamURLRemarkValidation)
		return
	}
	m := service.UpstreamURLRemarkMutation{Upstream: req.Upstream, Remark: *req.Remark, Revision: req.Revision}
	if err := service.ValidateUpstreamURLRemarkMutation(&m); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	identityHash := sha256.Sum256([]byte(m.Upstream))
	middleware.SetAuditAction(c, "admin.profit_url_remark.update")
	middleware.SetAuditExtra(c, map[string]any{"source": "profit_url_remark", "config_version": *m.Revision, "filter_hash": hex.EncodeToString(identityHash[:])})
	out, err := h.channelService.SaveUpstreamURLRemark(c.Request.Context(), m)
	if err != nil {
		profitURLRemarkError(c, err)
		return
	}
	response.Success(c, out)
}
