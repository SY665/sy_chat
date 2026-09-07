package handlers

import (
	"net/http"
	"strings"

	"sy_chat/internal/response"

	"github.com/gin-gonic/gin"
)

// ModelData 描述前端可以选择的 AI 模型及其能力。
type ModelData struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Provider         string `json:"provider"`
	SupportsThinking bool   `json:"supportsThinking"`
}

// ModelListData 包含可用模型以及服务端默认模型。
type ModelListData struct {
	Models         []ModelData `json:"models"`
	DefaultModelID string      `json:"defaultModelId"`
}

type ModelHandler struct {
	models         []ModelData
	defaultModelID string
}

func NewModelHandler(provider string, modelIDs []string, defaultModelID string, thinkingModelIDs []string) *ModelHandler {
	thinkingModels := make(
		map[string]struct{},
		len(thinkingModelIDs),
	)
	for _, modelID := range thinkingModelIDs {
		thinkingModels[modelID] = struct{}{}
	}

	if provider == "local" {
		return &ModelHandler{
			models: []ModelData{
				{
					ID:               "local",
					Name:             "本地模拟回复",
					Provider:         provider,
					SupportsThinking: false,
				},
			},
			defaultModelID: "local",
		}
	}

	models := make([]ModelData, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		_, supportsThinking := thinkingModels[modelID]

		models = append(models, ModelData{
			ID:               modelID,
			Name:             modelDisplayName(modelID),
			Provider:         provider,
			SupportsThinking: supportsThinking,
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

// List godoc
// @Summary 获取可用模型
// @Description 返回前端允许选择的模型，不包含 API Key 等服务端配置。
// @Tags Models
// @Produce json
// @Success 200 {object} response.Envelope{data=ModelListData}
// @Failure 401,500 {object} response.Envelope
// @Router /models [get]
func (handler *ModelHandler) List(c *gin.Context) {
	response.JSON(c, http.StatusOK, ModelListData{
		Models:         handler.models,
		DefaultModelID: handler.defaultModelID,
	})
}
