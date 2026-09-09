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

// CreateConversationRequest 描述创建对话时可以提交的数据。
type CreateConversationRequest struct {
	Title string `json:"title"`
}

// ConversationMessageData 描述对话详情中的单条消息。
type ConversationMessageData struct {
	ID          string                  `json:"id"`
	Role        string                  `json:"role"`
	Content     string                  `json:"content"`
	Thinking    *string                 `json:"thinking,omitempty"`
	Attachments []models.FileAttachment `json:"attachments,omitempty"`
	ToolEvents  []ToolEventData         `json:"toolEvents,omitempty"`
	CreatedAt   time.Time               `json:"createdAt"`
}

// ConversationDetailData 描述包含消息记录的完整对话。
type ConversationDetailData struct {
	ID         string                    `json:"id"`
	Title      string                    `json:"title"`
	IsPinned   bool                      `json:"isPinned"`
	IsShared   bool                      `json:"isShared"`
	ShareToken *string                   `json:"shareToken,omitempty"`
	SharedAt   *time.Time                `json:"sharedAt,omitempty"`
	CreatedAt  time.Time                 `json:"createdAt"`
	UpdatedAt  time.Time                 `json:"updatedAt"`
	Messages   []ConversationMessageData `json:"messages"`
}

// ConversationSummaryData 描述对话列表中的摘要信息。
type ConversationSummaryData struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	IsPinned  bool      `json:"isPinned"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ConversationListData 包含当前用户的对话列表。
type ConversationListData struct {
	Conversations []ConversationSummaryData `json:"conversations"`
}

// UpdateConversationTitleRequest 描述修改对话标题的参数。
type UpdateConversationTitleRequest struct {
	Title string `json:"title" binding:"required"`
}

// UpdateConversationPinnedRequest 描述对话的新置顶状态。
type UpdateConversationPinnedRequest struct {
	// 使用指针区分 false 和请求中没有传入 isPinned。
	IsPinned *bool `json:"isPinned" binding:"required"`
}

// ConversationTitleData 是修改标题后的响应数据。
type ConversationTitleData struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ConversationPinnedData 是修改置顶状态后的响应数据。
type ConversationPinnedData struct {
	ID       string `json:"id"`
	IsPinned bool   `json:"isPinned"`
}

// ConversationDeleteData 是删除对话后的响应数据。
type ConversationDeleteData struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// BatchDeleteConversationsRequest 描述批量删除的会话 ID。
type BatchDeleteConversationsRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// BatchDeleteConversationsData 描述批量删除结果。
type BatchDeleteConversationsData struct {
	DeletedCount int64 `json:"deletedCount"`
}

// ConversationShareData 包含公开分享令牌和创建时间。
type ConversationShareData struct {
	ShareToken string    `json:"shareToken"`
	SharedAt   time.Time `json:"sharedAt"`
}

// PublicShareMessageData 描述公开页面允许展示的消息。
type PublicShareMessageData struct {
	ID         string          `json:"id"`
	Role       string          `json:"role"`
	Content    string          `json:"content"`
	Thinking   *string         `json:"thinking,omitempty"`
	ToolEvents []ToolEventData `json:"toolEvents,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

// PublicShareData 描述无需登录即可读取的分享内容。
type PublicShareData struct {
	Title     string                   `json:"title"`
	OwnerName string                   `json:"ownerName"`
	SharedAt  *time.Time               `json:"sharedAt"`
	Messages  []PublicShareMessageData `json:"messages"`
	ViewCount int                      `json:"viewCount"`
}

// ConversationUnshareData 是关闭分享后的响应数据。
type ConversationUnshareData struct {
	ID       string `json:"id"`
	IsShared bool   `json:"isShared"`
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

// Create godoc
// @Summary 创建对话
// @Description 为当前登录用户创建一个新对话；标题为空时使用默认标题。
// @Tags Conversations
// @Accept json
// @Produce json
// @Param request body CreateConversationRequest true "创建对话参数"
// @Success 201 {object} response.Envelope{data=ConversationSummaryData}
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Router /conversations [post]
func (handler *ConversationHandler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request CreateConversationRequest
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

// List godoc
// @Summary 获取对话列表
// @Description 返回当前登录用户的全部对话，置顶对话优先排列。
// @Tags Conversations
// @Produce json
// @Success 200 {object} response.Envelope{data=ConversationListData}
// @Failure 401 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Router /conversations [get]
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

	data := make([]ConversationSummaryData, 0, len(conversations))
	for index := range conversations {
		data = append(
			data,
			newConversationSummaryData(&conversations[index]),
		)
	}

	response.JSON(c, http.StatusOK, ConversationListData{
		Conversations: data,
	})
}

// Get godoc
// @Summary 获取对话详情
// @Description 返回指定对话及其全部消息。
// @Tags Conversations
// @Produce json
// @Param id path string true "对话 ID"
// @Success 200 {object} response.Envelope{data=ConversationDetailData}
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Failure 404 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Router /conversations/{id} [get]
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

	messages := make([]ConversationMessageData, 0, len(conversation.Messages))
	for index := range conversation.Messages {
		message := &conversation.Messages[index]

		messages = append(messages, ConversationMessageData{
			ID:          message.ID,
			Role:        message.Role,
			Content:     message.Content,
			Thinking:    message.Thinking,
			Attachments: decodeMessageAttachments(message),
			ToolEvents:  decodeMessageToolEvents(message),
			CreatedAt:   message.CreatedAt,
		})
	}

	response.JSON(c, http.StatusOK, ConversationDetailData{
		ID:         conversation.ID,
		Title:      conversation.Title,
		IsPinned:   conversation.IsPinned,
		CreatedAt:  conversation.CreatedAt,
		UpdatedAt:  conversation.UpdatedAt,
		Messages:   messages,
		IsShared:   conversation.IsShared,
		ShareToken: conversation.ShareToken,
		SharedAt:   conversation.SharedAt,
	})
}

// UpdateTitle godoc
// @Summary 修改对话标题
// @Tags Conversations
// @Accept json
// @Produce json
// @Param id path string true "对话 ID"
// @Param request body UpdateConversationTitleRequest true "新标题"
// @Success 200 {object} response.Envelope{data=ConversationTitleData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /conversations/{id} [patch]
func (handler *ConversationHandler) UpdateTitle(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request UpdateConversationTitleRequest
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

	response.JSON(c, http.StatusOK, ConversationTitleData{
		ID:    c.Param("id"),
		Title: strings.TrimSpace(request.Title),
	})
}

// UpdatePinned godoc
// @Summary 修改对话置顶状态
// @Tags Conversations
// @Accept json
// @Produce json
// @Param id path string true "对话 ID"
// @Param request body UpdateConversationPinnedRequest true "置顶状态"
// @Success 200 {object} response.Envelope{data=ConversationPinnedData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /conversations/{id}/pin [patch]
func (handler *ConversationHandler) UpdatePinned(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request UpdateConversationPinnedRequest
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

	response.JSON(c, http.StatusOK, ConversationPinnedData{
		ID:       c.Param("id"),
		IsPinned: *request.IsPinned,
	})
}

// GetShared godoc
// @Summary 获取公开分享
// @Description 根据分享令牌读取公开对话，不需要登录。
// @Tags Shares
// @Produce json
// @Param token path string true "分享令牌"
// @Success 200 {object} response.Envelope{data=PublicShareData}
// @Failure 400,404,500 {object} response.Envelope
// @Router /shares/{token} [get]
func (handler *ConversationHandler) GetShared(c *gin.Context) {
	conversation, err := handler.conversationService.GetShared(
		c.Request.Context(),
		c.Param("token"),
	)
	if err != nil {
		switch {
		case errors.Is(err, conversationservice.ErrShareTokenRequired):
			response.Error(
				c,
				http.StatusBadRequest,
				"SHARE_TOKEN_REQUIRED",
				"缺少分享令牌",
			)

		case errors.Is(err, gorm.ErrRecordNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				"SHARE_NOT_FOUND",
				"分享不存在或已失效",
			)

		default:
			slog.Error("get shared conversation", "error", err)
			response.Error(
				c,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"读取分享内容失败",
			)
		}
		return
	}

	messages := make(
		[]PublicShareMessageData,
		0,
		len(conversation.Messages),
	)
	for index := range conversation.Messages {
		message := &conversation.Messages[index]

		// 系统和工具消息可能含有内部信息，不放入公开页面。
		if message.Role != models.MessageRoleUser &&
			message.Role != models.MessageRoleAssistant {
			continue
		}

		messages = append(messages, PublicShareMessageData{
			ID:         message.ID,
			Role:       message.Role,
			Content:    message.Content,
			Thinking:   message.Thinking,
			ToolEvents: decodeMessageToolEvents(message),
			CreatedAt:  message.CreatedAt,
		})
	}

	response.JSON(c, http.StatusOK, PublicShareData{
		Title:     conversation.Title,
		OwnerName: publicOwnerName(conversation.User),
		SharedAt:  conversation.SharedAt,
		ViewCount: conversation.ViewCount,
		Messages:  messages,
	})
}

// Share godoc
// @Summary 创建公开分享
// @Description 为当前用户的指定对话创建或返回公开分享令牌。
// @Tags Shares
// @Produce json
// @Param id path string true "对话 ID"
// @Success 200 {object} response.Envelope{data=ConversationShareData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /conversations/{id}/share [post]
func (handler *ConversationHandler) Share(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	result, err := handler.conversationService.Share(
		c.Request.Context(),
		c.Param("id"),
		userID,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, ConversationShareData{
		ShareToken: result.Token,
		SharedAt:   result.SharedAt,
	})
}

// Unshare godoc
// @Summary 关闭公开分享
// @Description 使指定对话现有的公开分享令牌失效。
// @Tags Shares
// @Produce json
// @Param id path string true "对话 ID"
// @Success 200 {object} response.Envelope{data=ConversationUnshareData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /conversations/{id}/share [delete]
func (handler *ConversationHandler) Unshare(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	err := handler.conversationService.Unshare(
		c.Request.Context(),
		c.Param("id"),
		userID,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, ConversationUnshareData{
		ID:       c.Param("id"),
		IsShared: false,
	})
}

// DeleteMany godoc
// @Summary 批量删除对话
// @Description 删除当前用户选择的多条对话及其全部消息。
// @Tags Conversations
// @Accept json
// @Produce json
// @Param input body BatchDeleteConversationsRequest true "会话 ID 列表"
// @Success 200 {object} response.Envelope{data=BatchDeleteConversationsData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /conversations/batch [delete]
func (handler *ConversationHandler) DeleteMany(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
		return
	}

	var request BatchDeleteConversationsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_BATCH_DELETE_INPUT",
			"请提供需要删除的会话 ID",
		)
		return
	}

	deletedCount, err := handler.conversationService.DeleteMany(
		c.Request.Context(),
		request.IDs,
		userID,
	)
	if err != nil {
		handler.handleError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, BatchDeleteConversationsData{
		DeletedCount: deletedCount,
	})
}

// Delete godoc
// @Summary 删除对话
// @Tags Conversations
// @Produce json
// @Param id path string true "对话 ID"
// @Success 200 {object} response.Envelope{data=ConversationDeleteData}
// @Failure 400,401,404,500 {object} response.Envelope
// @Router /conversations/{id} [delete]
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

	response.JSON(c, http.StatusOK, ConversationDeleteData{
		ID:      c.Param("id"),
		Message: "对话已删除",
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

	case errors.Is(err, conversationservice.ErrConversationIDsRequired):
		response.Error(
			c,
			http.StatusBadRequest,
			"CONVERSATION_IDS_REQUIRED",
			"请至少选择一个需要删除的对话",
		)

	case errors.Is(err, conversationservice.ErrTooManyConversationIDs):
		response.Error(
			c,
			http.StatusBadRequest,
			"TOO_MANY_CONVERSATION_IDS",
			"单次最多删除 100 个对话",
		)

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
) ConversationSummaryData {
	return ConversationSummaryData{
		ID:        conversation.ID,
		Title:     conversation.Title,
		IsPinned:  conversation.IsPinned,
		CreatedAt: conversation.CreatedAt,
		UpdatedAt: conversation.UpdatedAt,
	}
}

func publicOwnerName(user models.User) string {
	if user.Name != nil && strings.TrimSpace(*user.Name) != "" {
		return strings.TrimSpace(*user.Name)
	}

	if user.Username != nil && strings.TrimSpace(*user.Username) != "" {
		return strings.TrimSpace(*user.Username)
	}

	return "用户"
}
