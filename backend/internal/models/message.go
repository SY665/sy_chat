package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
	MessageRoleTool      = "tool"
)

type Message struct {
	ID             string `gorm:"type:char(36);primaryKey"`
	ConversationID string `gorm:"type:char(36);not null;index:idx_messages_conversation_created,priority:1"`

	Conversation Conversation `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Role        string         `gorm:"type:varchar(20);not null"`
	Content     string         `gorm:"type:longtext;not null"`
	Thinking    *string        `gorm:"type:longtext"`
	ToolCalls   datatypes.JSON `gorm:"type:json"`
	ToolResults datatypes.JSON `gorm:"type:json"`
	Attachments datatypes.JSON `gorm:"type:json"`
	CreatedAt   time.Time      `gorm:"not null;index:idx_messages_conversation_created,priority:2"`
}

func (message *Message) BeforeCreate(_ *gorm.DB) error {
	if message.ID == "" {
		message.ID = newID()
	}
	return nil
}
