package conversation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sy_chat/internal/models"
	toolservice "sy_chat/internal/services/tools"
	"time"
	"unicode/utf8"
)

const (
	DefaultTitle        = models.DefaultConversationTitle
	MaxTitleLength      = 255
	MaxBatchDeleteCount = 100
)

var (
	ErrUserIDRequired          = errors.New("user ID is required")
	ErrConversationIDRequired  = errors.New("conversation ID is required")
	ErrConversationIDsRequired = errors.New("conversation IDs are required")
	ErrTooManyConversationIDs  = errors.New("too many conversation IDs")
	ErrTitleRequired           = errors.New("conversation title is required")
	ErrTitleTooLong            = errors.New("conversation title is too long")
	ErrShareTokenRequired      = errors.New("share token is required")
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

	FindSharedByToken(
		ctx context.Context,
		token string,
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

	UpdatePinned(
		ctx context.Context,
		id string,
		userID string,
		isPinned bool,
		pinnedAt *time.Time,
	) error

	UpdateSharing(
		ctx context.Context,
		id string,
		userID string,
		shareToken *string,
		isShared bool,
		sharedAt *time.Time,
	) error

	Delete(
		ctx context.Context,
		id string,
		userID string,
	) ([]models.Message, error)

	DeleteMany(
		ctx context.Context,
		ids []string,
		userID string,
	) (int64, []models.Message, error)
}

type ShareResult struct {
	Token    string
	SharedAt time.Time
}

type Service struct {
	repository   Repository
	imageCleaner toolservice.GeneratedImageCleaner
}

func NewService(
	repository Repository,
	imageCleaner toolservice.GeneratedImageCleaner,
) *Service {
	return &Service{
		repository:   repository,
		imageCleaner: imageCleaner,
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

// GetShared 使用公开令牌读取分享会话，不需要用户身份。
func (service *Service) GetShared(
	ctx context.Context,
	token string,
) (*models.Conversation, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrShareTokenRequired
	}

	conversation, err := service.repository.FindSharedByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get shared conversation: %w", err)
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

// UpdatePinned 更新对话的置顶状态。
func (service *Service) UpdatePinned(
	ctx context.Context,
	id string,
	userID string,
	isPinned bool,
) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrConversationIDRequired
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ErrUserIDRequired
	}

	var pinnedAt *time.Time
	if isPinned {
		now := time.Now()
		pinnedAt = &now
	}

	if err := service.repository.UpdatePinned(
		ctx,
		id,
		userID,
		isPinned,
		pinnedAt,
	); err != nil {
		return fmt.Errorf("update conversation pinned state: %w", err)
	}

	return nil
}

// Share 为属于当前用户的会话生成公开分享令牌。
func (service *Service) Share(
	ctx context.Context,
	id string,
	userID string,
) (*ShareResult, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrConversationIDRequired
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	token, err := newShareToken()
	if err != nil {
		return nil, fmt.Errorf("generate share token: %w", err)
	}

	sharedAt := time.Now()
	if err := service.repository.UpdateSharing(
		ctx,
		id,
		userID,
		&token,
		true,
		&sharedAt,
	); err != nil {
		return nil, fmt.Errorf("share conversation: %w", err)
	}

	return &ShareResult{
		Token:    token,
		SharedAt: sharedAt,
	}, nil
}

// Unshare 关闭公开分享，并使旧令牌永久失效。
func (service *Service) Unshare(
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

	if err := service.repository.UpdateSharing(
		ctx,
		id,
		userID,
		nil,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("unshare conversation: %w", err)
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

	deletedMessages, err := service.repository.Delete(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}

	// 数据库删除已经提交，图片清理失败只记录日志。
	for _, message := range deletedMessages {
		if err := toolservice.DeleteGeneratedImages(
			service.imageCleaner,
			message.ToolResults,
		); err != nil {
			slog.Warn(
				"clean generated images after conversation delete",
				"conversation_id",
				id,
				"error",
				err,
			)
		}
	}

	return nil
}

// DeleteMany 校验并删除当前用户选择的多条对话。
func (service *Service) DeleteMany(
	ctx context.Context,
	ids []string,
	userID string,
) (int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, ErrUserIDRequired
	}

	if len(ids) == 0 {
		return 0, ErrConversationIDsRequired
	}
	if len(ids) > MaxBatchDeleteCount {
		return 0, ErrTooManyConversationIDs
	}

	normalizedIDs := make([]string, 0, len(ids))
	seenIDs := make(map[string]struct{}, len(ids))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			return 0, ErrConversationIDsRequired
		}

		if _, exists := seenIDs[id]; exists {
			continue
		}

		seenIDs[id] = struct{}{}
		normalizedIDs = append(normalizedIDs, id)
	}

	deletedCount, deletedMessages, err := service.repository.DeleteMany(
		ctx,
		normalizedIDs,
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"batch delete conversations: %w",
			err,
		)
	}

	// 批量删除事务已经提交，逐条清理消息关联的生成图片。
	for _, message := range deletedMessages {
		if err := toolservice.DeleteGeneratedImages(
			service.imageCleaner,
			message.ToolResults,
		); err != nil {
			slog.Warn(
				"clean generated images after batch conversation delete",
				"deleted_conversation_count",
				deletedCount,
				"error",
				err,
			)
		}
	}

	return deletedCount, nil
}

func newShareToken() (string, error) {
	const tokenBytes = 16

	value := make([]byte, tokenBytes)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("read secure random bytes: %w", err)
	}

	return hex.EncodeToString(value), nil
}
