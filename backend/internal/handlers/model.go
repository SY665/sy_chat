package handlers

import (
	"net/http"
	"strings"

	"sy_chat/internal/response"

	"github.com/gin-gonic/gin"
)

type modelData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

type ModelHandler struct {
	models         []modelData
	defaultModelID string
}

func NewModelHandler(provider string, modelIDs []string, defaultModelID string) *ModelHandler {
	if provider == "local" {
		return &ModelHandler{
			models: []modelData{
				{
					ID:       "local",
					Name:     "本地模拟回复",
					Provider: provider,
				},
			},
			defaultModelID: "local",
		}
	}

	models := make([]modelData, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, modelData{
			ID:       modelID,
			Name:     modelDisplayName(modelID),
			Provider: provider,
		})
	}
	return &ModelHandler{
		models:         models,
		defaultModelID: defaultModelID,
	}
}

func modelDisplayName(modelID string) string {
	parts := strings.Split(modelID, "/")
	if len(parts) == 0 {
		return modelID
	}

	return parts[len(parts)-1]
}

// List 只返回公开的模型信息，不包含 API Key 等服务端配置。
func (handler *ModelHandler) List(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{
		"models":         handler.models,
		"defaultModelId": handler.defaultModelID,
	})
}
