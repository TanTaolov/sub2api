package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *OpenAIOAuthReauthHandler) ListAuthConnections(c *gin.Context) {
	id, ok := parseReauthAccountID(c)
	if !ok {
		return
	}
	result, err := h.service.ListAuthConnections(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
func (h *OpenAIOAuthReauthHandler) CreateAuthConnection(c *gin.Context) {
	id, ok := parseReauthAccountID(c)
	if !ok {
		return
	}
	var input service.CreateOpenAIAuthConnectionInput
	if c.ShouldBindJSON(&input) != nil {
		response.BadRequest(c, "Connection kind and label are required")
		return
	}
	result, err := h.service.CreateAuthConnection(c.Request.Context(), id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
func (h *OpenAIOAuthReauthHandler) UpdateAuthConnection(c *gin.Context) {
	id, ok := parseReauthAccountID(c)
	if !ok {
		return
	}
	var input service.UpdateOpenAIAuthConnectionInput
	if c.ShouldBindJSON(&input) != nil {
		response.BadRequest(c, "Connection revision and action are required")
		return
	}
	result, err := h.service.UpdateAuthConnection(c.Request.Context(), id, c.Param("connection_id"), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
