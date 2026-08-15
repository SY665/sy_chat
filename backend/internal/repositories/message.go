package repositories

import (
	"context"
	"fmt"
	"sy_chat/internal/models"
	"time"

	"gorm.io/gorm"
)

// MessageRepository 负责消息写入和查询。
type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

// CreateExchange 在同一事务中保存用户消息和 AI 回复。
//
// 如果任意一步失败，事务会回滚，不会只留下其中一条消息。
func (repository *MessageRepository) CreateExchange(
	ctx context.Context,
	userID string,
	title string,
	userMessage *models.Message,
	assistantMessage *models.Message,
) error {
	err := repository.db.WithContext(ctx).Transaction(
		func(transaction *gorm.DB) error {
			// 同时检查对话 ID 和用户 ID，避免向其他用户的对话写入消息。
			var conversation models.Conversation

			if err := transaction.
				Select("id", "title").
				Where(
					"id = ? AND user_id = ?",
					userMessage.ConversationID,
					userID,
				).
				First(&conversation).
				Error; err != nil {
				return fmt.Errorf("find message conversation: %w", err)
			}

			if err := transaction.Create(userMessage).Error; err != nil {
				return fmt.Errorf("create user message: %w", err)
			}

			if err := transaction.Create(assistantMessage).Error; err != nil {
				return fmt.Errorf("create assistant message: %w", err)
			}

			updates := map[string]any{
				"updated_at": time.Now(),
			}

			// 只自动修改仍然使用默认标题的对话。
			// 用户手动修改标题后，发送消息不会覆盖它。
			if conversation.Title == models.DefaultConversationTitle {
				updates["title"] = title
			}

			// 新消息产生后更新对话时间，使其回到最近对话列表前面。
			if err := transaction.
				Model(&models.Conversation{}).
				Where("id = ?", conversation.ID).
				Updates(updates).
				Error; err != nil {
				return fmt.Errorf("update conversation time: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("create message exchange: %w", err)
	}
	return nil
}

// ListByConversation 查询对话消息，并再次校验对话所属用户。
func (repository *MessageRepository) ListByConversation(
	ctx context.Context,
	conversationID string,
	userID string,
) ([]models.Message, error) {
	var messages []models.Message

	err := repository.db.
		WithContext(ctx).
		Model(&models.Message{}).
		Joins(
			"JOIN conversations ON conversations.id = messages.conversation_id",
		).
		Where(
			"messages.conversation_id = ? AND conversations.user_id = ?",
			conversationID,
			userID,
		).
		Order("messages.created_at ASC").
		Find(&messages).
		Error

	if err != nil {
		return nil, fmt.Errorf("list conversation messages: %w", err)
	}

	return messages, nil
}
