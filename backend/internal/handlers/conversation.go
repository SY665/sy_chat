package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sy_chat/internal/middleware"
	"sy_chat/internal/models"
	"sy_chat/internal/response"
	conversationservice "sy_chat/internal/services/conversation"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createConversationRequest struct {
	Title string `json:"title"`
}

type updateConversationTitleRequest struct {
	Title string `json:"title" binding:"required"`
}

type updateConversationPinnedRequest struct {
	// 使用指针区分 false 和请求中没有传入 isPinned。
	IsPinned *bool `json:"isPinned" binding:"required"`
}

type conversationMessageData struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Thinking  *string   `json:"thinking,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type conversationDetailData struct {
	ID        string                    `json:"id"`
	Title     string                    `json:"title"`
	IsPinned  bool                      `json:"isPinned"`
	CreatedAt time.Time                 `json:"createdAt"`
	UpdatedAt time.Time                 `json:"updatedAt"`
	Messages  []conversationMessageData `json:"messages"`
}

// conversationSummaryData 是对话列表和创建接口返回的公开数据。
type conversationSummaryData struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	IsPinned  bool      `json:"isPinned"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type conversationListData struct {
	Conversations []conversationSummaryData `json:"conversations"`
}

type ConversationHandler struct {
	conversationService *conversationservice.Service
}

func NewConversationHandler(
	conversationService *conversationservice.Service,
) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
	}
}

// Create 为当前登录用户创建对话。
func (handler *ConversationHandler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request createConversationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"请求内容格式不正确",
		)
		return
	}

	conversation, err := handler.conversationService.Create(
		c.Request.Context(),
		userID,
		request.Title,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(
		c,
		http.StatusCreated,
		newConversationSummaryData(conversation),
	)
}

// List 返回当前登录用户的全部对话。
func (handler *ConversationHandler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	conversations, err := handler.conversationService.List(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	data := make([]conversationSummaryData, 0, len(conversations))
	for index := range conversations {
		data = append(
			data,
			newConversationSummaryData(&conversations[index]),
		)
	}

	response.JSON(c, http.StatusOK, conversationListData{
		Conversations: data,
	})
}

// Get 返回当前用户的指定对话及其消息。
func (handler *ConversationHandler) Get(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	conversation, err := handler.conversationService.Get(
		c.Request.Context(),
		c.Param("id"),
		userID,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	messages := make([]conversationMessageData, 0, len(conversation.Messages))
	for index := range conversation.Messages {
		message := &conversation.Messages[index]

		messages = append(messages, conversationMessageData{
			ID:        message.ID,
			Role:      message.Role,
			Content:   message.Content,
			Thinking:  message.Thinking,
			CreatedAt: message.CreatedAt,
		})
	}

	response.JSON(c, http.StatusOK, conversationDetailData{
		ID:        conversation.ID,
		Title:     conversation.Title,
		IsPinned:  conversation.IsPinned,
		CreatedAt: conversation.CreatedAt,
		UpdatedAt: conversation.UpdatedAt,
		Messages:  messages,
	})
}

// UpdateTitle 修改当前用户的对话标题。
func (handler *ConversationHandler) UpdateTitle(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request updateConversationTitleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_TITLE",
			"请输入对话标题",
		)
		return
	}

	err := handler.conversationService.UpdateTitle(
		c.Request.Context(),
		c.Param("id"),
		userID,
		request.Title,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"id":    c.Param("id"),
		"title": strings.TrimSpace(request.Title),
	})
}

// UpdatePinned 修改当前用户对话的置顶状态。
func (handler *ConversationHandler) UpdatePinned(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request updateConversationPinnedRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_PINNED_STATE",
			"请提供正确的置顶状态",
		)
		return
	}

	err := handler.conversationService.UpdatePinned(
		c.Request.Context(),
		c.Param("id"),
		userID,
		*request.IsPinned,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"id":       c.Param("id"),
		"isPinned": *request.IsPinned,
	})
}

// Delete 删除当前用户的指定对话。
func (handler *ConversationHandler) Delete(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	err := handler.conversationService.Delete(
		c.Request.Context(),
		c.Param("id"),
		userID,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"id":      c.Param("id"),
		"message": "对话已删除",
	})
}

func (handler *ConversationHandler) handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(
			c,
			http.StatusNotFound,
			"CONVERSATION_NOT_FOUND",
			"对话不存在",
		)

	case errors.Is(err, conversationservice.ErrConversationIDRequired):
		response.Error(
			c,
			http.StatusBadRequest,
			"CONVERSATION_ID_REQUIRED",
			"缺少对话 ID",
		)

	case errors.Is(err, conversationservice.ErrTitleRequired):
		response.Error(
			c,
			http.StatusBadRequest,
			"TITLE_REQUIRED",
			"对话标题不能为空",
		)
	case errors.Is(err, conversationservice.ErrTitleTooLong):
		response.Error(
			c,
			http.StatusBadRequest,
			"TITLE_TOO_LONG",
			"对话标题不能超过 255 个字符",
		)

	case errors.Is(err, conversationservice.ErrUserIDRequired):
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")

	default:
		slog.Error("handle conversation request", "error", err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"对话操作失败，请稍后重试",
		)
	}
}

func newConversationSummaryData(
	conversation *models.Conversation,
) conversationSummaryData {
	return conversationSummaryData{
		ID:        conversation.ID,
		Title:     conversation.Title,
		IsPinned:  conversation.IsPinned,
		CreatedAt: conversation.CreatedAt,
		UpdatedAt: conversation.UpdatedAt,
	}
}
