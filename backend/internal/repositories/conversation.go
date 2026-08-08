package repositories

import (
	"context"
	"fmt"
	"sy_chat/internal/models"

	"gorm.io/gorm"
)

// ConversationRepository 负责 conversations 表的数据库操作。
type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{
		db: db,
	}
}

// Create 创建一条新对话。
func (repository *ConversationRepository) Create(
	ctx context.Context,
	conversation *models.Conversation,
) error {
	if err := repository.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return fmt.Errorf("create conversation: %w", err)
	}
	return nil
}

// FindByID 查询属于指定用户的对话，并按照时间顺序加载消息。
func (repository *ConversationRepository) FindByID(
	ctx context.Context,
	id string,
	userID string,
) (*models.Conversation, error) {
	var conversation models.Conversation

	err := repository.db.
		WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ? AND user_id = ?", id, userID).
		First(&conversation).
		Error

	if err != nil {
		return nil, fmt.Errorf("find conversation: %w", err)
	}
	return &conversation, nil
}

// ListByUserID 查询用户的对话列表，置顶对话优先显示。
func (repository *ConversationRepository) ListByUserID(
	ctx context.Context,
	userID string,
) ([]models.Conversation, error) {
	var conversations []models.Conversation

	err := repository.db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_pinned DESC").
		Order("pinned_at DESC").
		Order("updated_at DESC").
		Find(&conversations).
		Error
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}

	return conversations, nil
}

// UpdateTitle 修改属于指定用户的对话标题。
func (repository *ConversationRepository) UpdateTitle(
	ctx context.Context,
	id string,
	userID string,
	title string,
) error {
	result := repository.db.
		WithContext(ctx).
		Model(&models.Conversation{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("title", title)
	if result.Error != nil {
		return fmt.Errorf("update conversation title: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除属于指定用户的对话。
func (repository *ConversationRepository) Delete(
	ctx context.Context,
	id string,
	userID string,
) error {
	result := repository.db.
		WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Conversation{})
	if result.Error != nil {
		return fmt.Errorf("delete conversation: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
