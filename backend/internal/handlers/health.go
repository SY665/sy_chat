package handlers

import (
	"net/http"
	"sy_chat/internal/response"

	"github.com/gin-gonic/gin"
)

// HealthResponse 描述健康检查返回的数据。
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Health godoc
// @Summary 检查服务健康状态
// @Description 返回 SY Chat API 当前是否可以正常响应。
// @Tags System
// @Produce json
// @Success 200 {object} response.Envelope{data=HealthResponse}
// @Router /health [get]
func Health(c *gin.Context) {
	response.JSON(c, http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "sy-chat-api",
	})
}
