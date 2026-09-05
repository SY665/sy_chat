package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	ConversationID string `json:"conversationId" binding:"required"`
	Message        string `json:"message" binding:"required"`
	// 允许旧版请求不传模型，后续由 Service 使用默认模型。
	ModelID string `json:"modelId"`
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
			ModelID:        request.ModelID,
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

// Stream 使用 SSE 将 AI 生成的增量内容持续发送给浏览器。
func (handler *ChatHandler) Stream(c *gin.Context) {
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

	streamStarted := false

	exchange, err := handler.chatService.Stream(
		c.Request.Context(),
		chat.SendInput{
			UserID:         userID,
			ConversationID: request.ConversationID,
			Content:        request.Message,
			ModelID:        request.ModelID,
		},
		func(content string) error {
			if !streamStarted {
				prepareSSEHeaders(c)
				streamStarted = true
			}

			return writeSSEEvent(c, "chunk", gin.H{
				"content": content,
			})
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) ||
			errors.Is(c.Request.Context().Err(), context.Canceled) {
			return
		}

		if !streamStarted {
			handler.handleError(c, err)
			return
		}

		slog.Error("stream chat message", "error", err)

		// 只有客户端仍连接时才尝试发送流内错误。
		_ = writeSSEEvent(c, "error", gin.H{
			"code":    "STREAM_FAILED",
			"message": "生成回复失败，请稍后重试",
		})
		return
	}

	if !streamStarted {
		prepareSSEHeaders(c)
	}

	// done 事件提供数据库生成的真实消息 ID 和时间。
	if err := writeSSEEvent(c, "done", chatData{
		UserMessage: newChatMessageData(
			exchange.UserMessage,
		),
		AssistantMessage: newChatMessageData(
			exchange.AssistantMessage,
		),
	}); err != nil {
		// 客户端已断开时，写入失败属于预期行为。
		if c.Request.Context().Err() == nil {
			slog.Error(
				"write chat stream completion",
				"error",
				err,
			)
		}
	}
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

	case errors.Is(err, chat.ErrModelUnavailable):
		response.Error(
			c,
			http.StatusBadRequest,
			"MODEL_UNAVAILABLE",
			"所选模型不可用，请刷新模型列表",
		)

	case errors.Is(err, chat.ErrStreamingNotSupported):
		response.Error(
			c,
			http.StatusNotImplemented,
			"STREAMING_NOT_SUPPORTED",
			"当前 AI 服务暂不支持流式回复",
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

func prepareSSEHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 禁止反向代理缓存流式响应。
	c.Header("X-Accel-Buffering", "no")
}

func writeSSEEvent(
	c *gin.Context,
	event string,
	data any,
) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode SSE event: %w", err)
	}

	if _, err := fmt.Fprintf(
		c.Writer,
		"event: %s\ndata: %s\n\n",
		event,
		payload,
	); err != nil {
		return fmt.Errorf("write SSE event: %w", err)
	}

	// 立即刷新缓冲区，否则浏览器可能等到请求结束才收到内容。
	c.Writer.Flush()

	return nil
}
