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

// CreateExchange 在同一事务中保存用户消息和可选的 AI 回复。
//
// assistantMessage 为 nil 时只保存用户消息，用于生成开始前就被停止的情况。
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

			// 已经收到部分回复时保存 AI 消息；尚未收到内容时只保存用户消息。
			if assistantMessage != nil {
				if err := transaction.Create(assistantMessage).Error; err != nil {
					return fmt.Errorf("create assistant message: %w", err)
				}
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

// ListRecentByConversation 查询当前用户对话中的最近消息。
//
// 查询前先校验对话归属，避免非法对话 ID 触发外部 AI 调用。
func (repository *MessageRepository) ListRecentByConversation(
	ctx context.Context,
	conversationID string,
	userID string,
	limit int,
) ([]models.Message, error) {
	db := repository.db.WithContext(ctx)
	var conversation models.Conversation

	if err := db.
		Select("id").
		Where("id = ? AND user_id = ?", conversationID, userID).
		First(&conversation).
		Error; err != nil {
		return nil, fmt.Errorf(
			"find conversation for message context: %w",
			err,
		)
	}

	if limit <= 0 {
		return []models.Message{}, nil
	}

	var messages []models.Message

	// 先倒序取最近的 N 条，避免长对话把全部历史发送给模型。
	if err := db.
		Where("conversation_id = ?", conversation.ID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).
		Error; err != nil {
		return nil, fmt.Errorf(
			"list recent conversation messages: %w",
			err,
		)
	}

	// AI 上下文需要按照真实对话顺序排列。
	for left, right := 0, len(messages)-1; left < right; left, right =
		left+1, right-1 {
		messages[left], messages[right] =
			messages[right], messages[left]
	}

	return messages, nil
}

// DeleteFromUserMessage 删除指定用户消息以及它之后的全部消息。
//
// 查询目标消息时同时校验会话归属，并使用事务保证删除与会话更新时间同步。
func (repository *MessageRepository) DeleteFromUserMessage(
	ctx context.Context,
	conversationID string,
	userID string,
	messageID string,
) (int64, error) {
	var deletedCount int64

	err := repository.db.WithContext(ctx).Transaction(
		func(transaction *gorm.DB) error {
			var target models.Message

			// 删除边界只能是当前用户会话中的用户消息。
			if err := transaction.
				Table("messages").
				Select(
					"messages.id",
					"messages.conversation_id",
					"messages.created_at",
				).
				Joins(
					"JOIN conversations ON conversations.id = messages.conversation_id",
				).
				Where("messages.id = ?", messageID).
				Where("messages.conversation_id = ?", conversationID).
				Where("messages.role = ?", models.MessageRoleUser).
				Where("conversations.user_id = ?", userID).
				Take(&target).
				Error; err != nil {
				return fmt.Errorf(
					"find truncation start message: %w",
					err,
				)
			}

			// 用户消息和对应 AI 消息的创建时间有先后顺序，因此可以按时间截断。
			result := transaction.
				Where("conversation_id = ?", target.ConversationID).
				Where("created_at >= ?", target.CreatedAt).
				Delete(&models.Message{})
			if result.Error != nil {
				return fmt.Errorf(
					"delete messages from target: %w",
					result.Error,
				)
			}

			deletedCount = result.RowsAffected

			// 截断本身也是一次会话变更，需要刷新会话排序时间。
			if err := transaction.
				Model(&models.Conversation{}).
				Where(
					"id = ? AND user_id = ?",
					target.ConversationID,
					userID,
				).
				Update("updated_at", time.Now()).
				Error; err != nil {
				return fmt.Errorf(
					"update truncated conversation time: %w",
					err,
				)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf(
			"delete messages from user message: %w",
			err,
		)
	}

	return deletedCount, nil
}
