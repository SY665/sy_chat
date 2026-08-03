package handlers

import (
	"net/http"
	"sy_chat/internal/response"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func Health(c *gin.Context) {
	response.JSON(c, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "sy-chat-api",
	})
}
