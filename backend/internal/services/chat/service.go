package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sy_chat/internal/models"
	"time"
	"unicode/utf8"
)

const MaxMessageLength = 20000

var (
	ErrUserIDRequired         = errors.New("user ID is required")
	ErrConversationIDRequired = errors.New("conversation ID is required")
	ErrMessageRequired        = errors.New("message is required")
	ErrMessageTooLong         = errors.New("message is too long")
)

// Repository 描述聊天业务需要的消息数据库操作。
type Repository interface {
	CreateExchange(
		ctx context.Context,
		userID string,
		title string,
		userMessage *models.Message,
		assistantMessage *models.Message,
	) error
}

type SendInput struct {
	UserID         string
	ConversationID string
	Content        string
}

type Exchange struct {
	UserMessage      *models.Message
	AssistantMessage *models.Message
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// Send 校验消息、生成回复并持久化完整的一轮对话。
func (service *Service) Send(
	ctx context.Context,
	input SendInput,
) (*Exchange, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	conversationID := strings.TrimSpace(input.ConversationID)
	if conversationID == "" {
		return nil, ErrConversationIDRequired
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, ErrMessageRequired
	}

	if utf8.RuneCountInString(content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	reply := service.GenerateReply(content)
	now := time.Now()

	userMessage := &models.Message{
		ConversationID: conversationID,
		Role:           models.MessageRoleUser,
		Content:        content,
		CreatedAt:      now,
	}

	assistantMessage := &models.Message{
		ConversationID: conversationID,
		Role:           models.MessageRoleAssistant,
		Content:        reply,
		CreatedAt:      now.Add(time.Millisecond),
	}

	title := createConversationTitle(content)

	if err := service.repository.CreateExchange(
		ctx,
		userID,
		title,
		userMessage,
		assistantMessage,
	); err != nil {
		return nil, fmt.Errorf("save chat exchange: %w", err)
	}

	return &Exchange{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
	}, nil
}

// createConversationTitle 使用第一条消息生成简短标题。
func createConversationTitle(content string) string {
	const maxTitleLength = 30

	// 去除换行和连续空格，避免标题破坏侧边栏布局。
	normalized := strings.Join(strings.Fields(content), " ")
	characters := []rune(normalized)

	if len(characters) <= maxTitleLength {
		return normalized
	}

	return string(characters[:maxTitleLength]) + "…"
}

// 生成回复
func (*Service) GenerateReply(message string) string {
	return fmt.Sprintf(
		"我收到了你的消息：“%s”。这条回复来自 Go 后端。",
		message,
	)
}
