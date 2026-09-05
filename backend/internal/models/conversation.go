package models

import (
	"time"

	"gorm.io/gorm"
)

const DefaultConversationTitle = "新对话"

type Conversation struct {
	ID     string `gorm:"type:char(36);primaryKey"`
	Title  string `gorm:"type:varchar(255);not null;default:新对话"`
	UserID string `gorm:"type:char(36);not null;index:idx_conversations_user_updated,priority:1;index:idx_conversations_user_pinned,priority:1"`

	User     User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Messages []Message `gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null;index:idx_conversations_user_updated,priority:2,sort:desc"`

	ShareToken   *string `gorm:"type:varchar(64);uniqueIndex"`
	IsShared     bool    `gorm:"not null;default:false"`
	SharedAt     *time.Time
	ViewCount    int `gorm:"not null;default:0"`
	LastViewedAt *time.Time
	IsPinned     bool       `gorm:"not null;default:false;index:idx_conversations_user_pinned,priority:2"`
	PinnedAt     *time.Time `gorm:"index:idx_conversations_user_pinned,priority:3,sort:desc"`
}

func (conversation *Conversation) BeforeCreate(_ *gorm.DB) error {
	if conversation.ID == "" {
		conversation.ID = newID()
	}
	return nil
}
