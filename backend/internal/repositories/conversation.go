package repositories

import (
	"context"
	"fmt"
	"sy_chat/internal/models"
	"time"

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

// FindSharedByToken 查询公开会话，并原子增加浏览次数。
func (repository *ConversationRepository) FindSharedByToken(
	ctx context.Context,
	token string,
) (*models.Conversation, error) {
	var conversation models.Conversation

	err := repository.db.
		WithContext(ctx).
		Preload("User").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("share_token = ? AND is_shared = ?", token, true).
		First(&conversation).
		Error
	if err != nil {
		return nil, fmt.Errorf("find shared conversation: %w", err)
	}

	viewedAt := time.Now()
	result := repository.db.
		WithContext(ctx).
		Model(&models.Conversation{}).
		Where(
			"id = ? AND share_token = ? AND is_shared = ?",
			conversation.ID,
			token,
			true,
		).
		UpdateColumns(map[string]any{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_viewed_at": viewedAt,
		})
	if result.Error != nil {
		return nil, fmt.Errorf("record shared conversation view: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	conversation.ViewCount++
	conversation.LastViewedAt = &viewedAt

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

func (repository *ConversationRepository) UpdatePinned(
	ctx context.Context,
	id string,
	userID string,
	isPinned bool,
	pinnedAt *time.Time,
) error {
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conversation models.Conversation

		// 先确认会话存在且属于当前用户，避免把“状态未变化”误判为记录不存在。
		if err := tx.
			Select("id").
			Where("id = ? AND user_id = ?", id, userID).
			First(&conversation).Error; err != nil {
			return fmt.Errorf("find conversation before pin update: %w", err)
		}

		if err := tx.Model(&conversation).Updates(map[string]any{
			"is_pinned": isPinned,
			"pinned_at": pinnedAt,
		}).Error; err != nil {
			return fmt.Errorf("update conversation pin fields: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("update conversation pinned state: %w", err)
	}

	return nil
}

// UpdateSharing 更新会话分享状态，并校验会话归属。
func (repository *ConversationRepository) UpdateSharing(
	ctx context.Context,
	id string,
	userID string,
	shareToken *string,
	isShared bool,
	sharedAt *time.Time,
) error {
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conversation models.Conversation

		if err := tx.
			Select("id").
			Where("id = ? AND user_id = ?", id, userID).
			First(&conversation).Error; err != nil {
			return fmt.Errorf("find conversation before sharing update: %w", err)
		}

		if err := tx.Model(&conversation).Updates(map[string]any{
			"share_token": shareToken,
			"is_shared":   isShared,
			"shared_at":   sharedAt,
		}).Error; err != nil {
			return fmt.Errorf("update conversation sharing fields: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("update conversation sharing state: %w", err)
	}

	return nil
}

// Delete 在事务中删除属于指定用户的对话及其全部消息。
// 返回消息快照，供事务提交后清理关联的本地资源。
func (repository *ConversationRepository) Delete(
	ctx context.Context,
	id string,
	userID string,
) ([]models.Message, error) {
	var deletedMessages []models.Message

	err := repository.db.WithContext(ctx).Transaction(
		func(transaction *gorm.DB) error {
			var conversation models.Conversation

			// 删除消息前先校验对话归属，避免操作其他用户的数据。
			if err := transaction.
				Select("id").
				Where("id = ? AND user_id = ?", id, userID).
				First(&conversation).
				Error; err != nil {
				return fmt.Errorf(
					"find conversation before delete: %w",
					err,
				)
			}

			// 数据库提交后才能删除图片，因此先保存工具结果快照。
			if err := transaction.
				Select("tool_results").
				Where("conversation_id = ?", conversation.ID).
				Find(&deletedMessages).
				Error; err != nil {
				return fmt.Errorf(
					"list conversation messages before delete: %w",
					err,
				)
			}

			if err := transaction.
				Where("conversation_id = ?", conversation.ID).
				Delete(&models.Message{}).
				Error; err != nil {
				return fmt.Errorf(
					"delete conversation messages: %w",
					err,
				)
			}

			if err := transaction.
				Delete(&conversation).
				Error; err != nil {
				return fmt.Errorf(
					"delete conversation: %w",
					err,
				)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"delete conversation transaction: %w",
			err,
		)
	}

	return deletedMessages, nil
}

// DeleteMany 在同一事务中删除属于指定用户的多条对话及其消息。
// 返回消息快照，供事务提交后清理关联的本地资源。
func (repository *ConversationRepository) DeleteMany(
	ctx context.Context,
	ids []string,
	userID string,
) (int64, []models.Message, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	var deletedCount int64
	var deletedMessages []models.Message

	err := repository.db.WithContext(ctx).Transaction(
		func(transaction *gorm.DB) error {
			var ownedCount int64

			// 必须确认所有 ID 都属于当前用户，避免只删除其中一部分。
			if err := transaction.
				Model(&models.Conversation{}).
				Where("id IN ? AND user_id = ?", ids, userID).
				Count(&ownedCount).
				Error; err != nil {
				return fmt.Errorf(
					"count conversations before batch delete: %w",
					err,
				)
			}

			if ownedCount != int64(len(ids)) {
				return gorm.ErrRecordNotFound
			}

			// 所有会话归属确认后，再保存即将删除消息的工具结果。
			if err := transaction.
				Select("tool_results").
				Where("conversation_id IN ?", ids).
				Find(&deletedMessages).
				Error; err != nil {
				return fmt.Errorf(
					"list messages before batch delete: %w",
					err,
				)
			}

			if err := transaction.
				Where("conversation_id IN ?", ids).
				Delete(&models.Message{}).
				Error; err != nil {
				return fmt.Errorf(
					"delete conversation messages in batch: %w",
					err,
				)
			}

			result := transaction.
				Where("id IN ? AND user_id = ?", ids, userID).
				Delete(&models.Conversation{})
			if result.Error != nil {
				return fmt.Errorf(
					"delete conversations in batch: %w",
					result.Error,
				)
			}

			deletedCount = result.RowsAffected
			return nil
		},
	)
	if err != nil {
		return 0, nil, fmt.Errorf(
			"batch delete conversation transaction: %w",
			err,
		)
	}

	return deletedCount, deletedMessages, nil
}
