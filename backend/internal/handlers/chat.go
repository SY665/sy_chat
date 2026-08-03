package handlers

import (
	"net/http"
	"strings"
	"sy_chat/internal/response"
	"sy_chat/internal/services/chat"

	"github.com/gin-gonic/gin"
)

// 前端请求
type chatRequest struct {
	Message string `json:"message" binding:"required,max=20000"`
}

// 返回数据
type chatData struct {
	Reply string `json:"reply"`
}

// 接口输入输出
type ChatHandler struct {
	chatService *chat.Service
}

// 创建聊天handler，注入service
func NewChatHandler(chatService *chat.Service) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// 校验输入，调用Service
func (h *ChatHandler) Send(c *gin.Context) {
	var request chatRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"消息内容不能为空，且不能超过 20000 个字符",
		)
		return
	}

	message := strings.TrimSpace(request.Message)
	if message == "" {
		response.Error(
			c,
			http.StatusBadRequest,
			"EMPTY_MESSAGE",
			"消息内容不能为空",
		)
		return
	}

	reply := h.chatService.GenerateReply(message)

	response.JSON(c, http.StatusOK, chatData{
		Reply: reply,
	})
}
