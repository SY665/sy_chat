package repositories

import (
	"context"
	"fmt"
	"sy_chat/internal/models"

	"gorm.io/gorm"
)

// UserRepository 负责 users 表的数据库操作。
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create 创建用户。传入的 Password 必须已经完成哈希处理。
func (repository *UserRepository) Create(
	ctx context.Context,
	user *models.User,
) error {
	if err := repository.
		db.
		WithContext(ctx).
		Create(user).
		Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// FindByID 根据主键查询用户。
func (repository *UserRepository) FindByID(
	ctx context.Context,
	id string,
) (*models.User, error) {
	var user models.User

	err := repository.db.
		WithContext(ctx).
		Where("id = ?", id).
		First(&user).
		Error

	if err != nil {
		return nil, fmt.Errorf("find user by ID: %w", err)
	}

	return &user, nil
}

// FindByUsername 根据用户名查询用户。
func (repository *UserRepository) FindByUsername(
	ctx context.Context,
	username string,
) (*models.User, error) {
	var user models.User

	err := repository.db.
		WithContext(ctx).
		Where("username = ?", username).
		First(&user).
		Error

	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return &user, nil
}

// FindByEmail 根据邮箱查询用户。
func (repository *UserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	var user models.User

	err := repository.db.
		WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error

	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}

// FindByIdentifier 允许用户使用用户名或邮箱登录。
func (repository *UserRepository) FindByIdentifier(
	ctx context.Context,
	identifier string,
) (*models.User, error) {
	var user models.User

	err := repository.db.
		WithContext(ctx).
		Where("username = ? OR email = ?", identifier, identifier).
		First(&user).
		Error

	if err != nil {
		return nil, fmt.Errorf("find user by identifier: %w", err)
	}

	return &user, nil
}
