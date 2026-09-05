package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            string  `gorm:"type:char(36);primaryKey"`
	Username      *string `gorm:"type:varchar(50);uniqueIndex"`
	Password      *string `json:"-" gorm:"type:varchar(255)"`
	Email         *string `gorm:"type:varchar(255);uniqueIndex"`
	EmailVerified *time.Time
	Name          *string        `gorm:"type:varchar(100)"`
	Image         *string        `gorm:"type:text"`
	APIKey        *string        `json:"-" gorm:"type:varchar(255)"`
	CreatedAt     time.Time      `gorm:"not null"`
	Conversations []Conversation `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// BeforeCreate 在写入数据库前自动补充主键。
func (user *User) BeforeCreate(_ *gorm.DB) error {
	if user.ID == "" {
		user.ID = newID()
	}
	return nil
}
