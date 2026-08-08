package conversation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sy_chat/internal/models"
	"unicode/utf8"
)

const (
	DefaultTitle   = "新对话"
	MaxTitleLength = 255
)

var (
	ErrUserIDRequired         = errors.New("user ID is required")
	ErrConversationIDRequired = errors.New("conversation ID is required")
	ErrTitleRequired          = errors.New("conversation title is required")
	ErrTitleTooLong           = errors.New("conversation title is too long")
)

// Repository 描述 Service 所需要的数据库操作。
type Repository interface {
	Create(
		ctx context.Context,
		conversation *models.Conversation,
	) error

	FindByID(
		ctx context.Context,
		id string,
		userID string,
	) (*models.Conversation, error)

	ListByUserID(
		ctx context.Context,
		userID string,
	) ([]models.Conversation, error)

	UpdateTitle(
		ctx context.Context,
		id string,
		userID string,
		title string,
	) error

	Delete(
		ctx context.Context,
		id string,
		userID string,
	) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// Create 创建对话，标题为空时使用默认标题。
func (service *Service) Create(
	ctx context.Context,
	userID string,
	title string,
) (*models.Conversation, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = DefaultTitle
	}

	if utf8.RuneCountInString(title) > MaxTitleLength {
		return nil, ErrTitleTooLong
	}

	conversation := &models.Conversation{
		UserID: userID,
		Title:  title,
	}

	if err := service.repository.Create(ctx, conversation); err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	return conversation, nil
}

// Get 查询一个对话及其消息。
func (service *Service) Get(
	ctx context.Context,
	id string,
	userID string,
) (*models.Conversation, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrConversationIDRequired
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	conversation, err := service.repository.FindByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get conversation: %w", err)
	}

	return conversation, nil
}

// List 查询指定用户的对话列表。
func (service *Service) List(
	ctx context.Context,
	userID string,
) ([]models.Conversation, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	conversations, err := service.repository.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}

	return conversations, nil
}

// UpdateTitle 修改对话标题。
func (service *Service) UpdateTitle(
	ctx context.Context,
	id string,
	userID string,
	title string,
) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrConversationIDRequired
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ErrUserIDRequired
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleRequired
	}

	if utf8.RuneCountInString(title) > MaxTitleLength {
		return ErrTitleTooLong
	}

	if err := service.repository.UpdateTitle(ctx, id, userID, title); err != nil {
		return fmt.Errorf("update conversation title: %w", err)
	}

	return nil
}

// Delete 删除指定用户的对话。
func (service *Service) Delete(
	ctx context.Context,
	id string,
	userID string,
) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrConversationIDRequired
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ErrUserIDRequired
	}

	if err := service.repository.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}

	return nil
}
