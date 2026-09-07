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

// ChatRequest 描述发送聊天消息需要的参数。
type ChatRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
	Message        string `json:"message" binding:"required"`
	// 允许旧版请求不传模型，由 Service 使用默认模型。
	ModelID        string `json:"modelId"`
	EnableThinking bool   `json:"enableThinking"`
}

// ChatMessageData 描述聊天接口返回的一条已保存消息。
type ChatMessageData struct {
	ID        string  `json:"id"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	Thinking  *string `json:"thinking,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

// ChatData 包含一次问答产生的用户消息和 AI 消息。
type ChatData struct {
	UserMessage      ChatMessageData `json:"userMessage"`
	AssistantMessage ChatMessageData `json:"assistantMessage"`
}

// TruncateMessagesData 描述消息截断操作的结果。
type TruncateMessagesData struct {
	ConversationID string `json:"conversationId"`
	MessageID      string `json:"messageId"`
	DeletedCount   int64  `json:"deletedCount"`
}

// ChatHandler 负责处理聊天相关的 HTTP 请求。
type ChatHandler struct {
	chatService *chat.Service
}

// NewChatHandler 创建聊天处理器并注入聊天服务。
func NewChatHandler(chatService *chat.Service) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// Send godoc
// @Summary 发送聊天消息
// @Description 保存用户消息，并以普通 JSON 响应返回完整的 AI 回复。
// @Tags Chat
// @Accept json
// @Produce json
// @Param request body ChatRequest true "聊天参数"
// @Success 200 {object} response.Envelope{data=ChatData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /chat [post]
func (handler *ChatHandler) Send(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request ChatRequest

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
			EnableThinking: request.EnableThinking,
		},
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, ChatData{
		UserMessage:      newChatMessageData(exchange.UserMessage),
		AssistantMessage: newChatMessageData(exchange.AssistantMessage),
	})
}

// Stream godoc
// @Summary 流式发送聊天消息
// @Description 使用 SSE 持续返回 AI 回复。chunk 事件包含增量文本，done 事件包含最终消息，error 事件表示流内错误。
// @Tags Chat
// @Accept json
// @Produce text/event-stream
// @Param request body ChatRequest true "聊天参数"
// @Success 200 {string} string "SSE 事件流：chunk、done、error"
// @Failure 400,401,404,500,501 {object} response.Envelope
// @Router /chat/stream [post]
func (handler *ChatHandler) Stream(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request ChatRequest
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
			EnableThinking: request.EnableThinking,
		},
		func(chunk chat.StreamChunk) error {
			if !streamStarted {
				prepareSSEHeaders(c)
				streamStarted = true
			}

			if chunk.Thinking != "" {
				if err := writeSSEEvent(c, "thinking", gin.H{
					"content": chunk.Thinking,
				}); err != nil {
					return err
				}
			}

			if chunk.Content != "" {
				if err := writeSSEEvent(c, "chunk", gin.H{
					"content": chunk.Content,
				}); err != nil {
					return err
				}
			}

			return nil
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
	if err := writeSSEEvent(c, "done", ChatData{
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

// TruncateFromMessage godoc
// @Summary 截断对话消息
// @Description 删除目标用户消息以及它之后的全部消息，用于重试和编辑重发。
// @Tags Chat
// @Produce json
// @Param id path string true "对话 ID"
// @Param messageId path string true "作为截断起点的用户消息 ID"
// @Success 200 {object} response.Envelope{data=TruncateMessagesData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /conversations/{id}/messages/{messageId}/tail [delete]
func (handler *ChatHandler) TruncateFromMessage(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	conversationID := c.Param("id")
	messageID := c.Param("messageId")

	deletedCount, err := handler.chatService.TruncateFromUserMessage(
		c.Request.Context(),
		chat.TruncateInput{
			UserID:         userID,
			ConversationID: conversationID,
			MessageID:      messageID,
		},
	)
	if err != nil {
		// 在该接口中，记录不存在表示目标用户消息或所属对话不存在。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"MESSAGE_BRANCH_NOT_FOUND",
				"用户消息或所属对话不存在",
			)
			return
		}

		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, TruncateMessagesData{
		ConversationID: conversationID,
		MessageID:      messageID,
		DeletedCount:   deletedCount,
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

	case errors.Is(err, chat.ErrMessageIDRequired):
		response.Error(
			c,
			http.StatusBadRequest,
			"MESSAGE_ID_REQUIRED",
			"缺少消息 ID",
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

	case errors.Is(err, chat.ErrThinkingNotSupported):
		response.Error(
			c,
			http.StatusBadRequest,
			"THINKING_NOT_SUPPORTED",
			"当前模型不支持思考模式",
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

func newChatMessageData(message *models.Message) ChatMessageData {
	return ChatMessageData{
		ID:        message.ID,
		Role:      message.Role,
		Content:   message.Content,
		Thinking:  message.Thinking,
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
