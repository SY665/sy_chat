package database

import (
	"fmt"
	"sy_chat/internal/models"

	"gorm.io/gorm"
)

// Migrate 创建或更新开发环境所需的数据表。
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.Conversation{},
		&models.Message{},
	)
	if err != nil {
		return fmt.Errorf("auto migrate database: %w", err)
	}

	return nil
}
