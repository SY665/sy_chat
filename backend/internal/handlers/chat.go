package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"sy_chat/internal/middleware"
	"sy_chat/internal/models"
	"sy_chat/internal/response"
	"sy_chat/internal/services/chat"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 前端请求
type chatRequest struct {
	ConversationID string `json:"conversationID" binding:"required"`
	Message        string `json:"message" binding:"required"`
}

// 处理数据
type chatMessageData struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// 返回数据
type chatData struct {
	UserMessage      chatMessageData `json:"userMessage"`
	AssistantMessage chatMessageData `json:"assistantMessage"`
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

// Send 保存用户消息和暂时生成的 Go 回复。
func (handler *ChatHandler) Send(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

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

	exchange, err := handler.chatService.Send(
		c.Request.Context(),
		chat.SendInput{
			UserID:         userID,
			ConversationID: request.ConversationID,
			Content:        request.Message,
		},
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, chatData{
		UserMessage:      newChatMessageData(exchange.UserMessage),
		AssistantMessage: newChatMessageData(exchange.AssistantMessage),
	})
}

func (handler *ChatHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, chat.ErrMessageRequired):
		response.Error(c, http.StatusBadRequest, "EMPTY_MESSAGE", "消息不能为空")

	case errors.Is(err, chat.ErrMessageTooLong):
		response.Error(
			c,
			http.StatusBadRequest,
			"MESSAGE_TOO_LONG",
			"消息不能超过 20000 个字符",
		)

	case errors.Is(err, chat.ErrConversationIDRequired):
		response.Error(
			c,
			http.StatusBadRequest,
			"CONVERSATION_ID_REQUIRED",
			"缺少对话 ID",
		)

	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(
			c,
			http.StatusNotFound,
			"CONVERSATION_NOT_FOUND",
			"对话不存在",
		)

	default:
		slog.Error("send chat message", "error", err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"发送消息失败，请稍后重试",
		)
	}
}

func newChatMessageData(message *models.Message) chatMessageData {
	return chatMessageData{
		ID:        message.ID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt.Format(time.RFC3339Nano),
	}
}
