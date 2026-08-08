package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"sy_chat/internal/models"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUsernameRequired   = errors.New("username is required")
	ErrUsernameInvalid    = errors.New("username is invalid")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailRequired      = errors.New("email is required")
	ErrEmailInvalid       = errors.New("email is invalid")
	ErrEmailExists        = errors.New("email already exists")
	ErrPasswordTooShort   = errors.New("password is too short")
	ErrPasswordTooLong    = errors.New("password is too long")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserIDRequired     = errors.New("userid is required")
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// Repository 描述认证业务需要的用户数据库操作。
type Repository interface {
	Create(ctx context.Context, user *models.User) error
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByIdentifier(ctx context.Context, identifier string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type LoginInput struct {
	Identifier string
	Password   string
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// Register 校验注册信息，并使用 bcrypt 保存密码哈希。
func (service *Service) Register(
	ctx context.Context,
	input RegisterInput,
) (*models.User, error) {
	username := strings.ToLower(strings.TrimSpace(input.Username))
	email := strings.ToLower(strings.TrimSpace(input.Email))

	if username == "" {
		return nil, ErrUsernameRequired
	}

	if !usernamePattern.MatchString(username) {
		return nil, ErrUsernameInvalid
	}

	if email == "" {
		return nil, ErrEmailRequired
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return nil, ErrEmailInvalid
	}

	if utf8.RuneCountInString(input.Password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// bcrypt 最多接受 72 字节的密码。
	if len([]byte(input.Password)) > 72 {
		return nil, ErrPasswordTooLong
	}

	_, err = service.repository.FindByUsername(ctx, username)
	switch {
	case err == nil:
		return nil, ErrUsernameExists
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, fmt.Errorf("check username: %w", err)
	}

	_, err = service.repository.FindByEmail(ctx, email)
	switch {
	case err == nil:
		return nil, ErrEmailExists
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, fmt.Errorf("check email: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	password := string(passwordHash)
	name := username

	user := &models.User{
		Username: &username,
		Email:    &email,
		Password: &password,
		Name:     &name,
	}

	if err := service.repository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}
	return user, nil
}

// Login 验证用户名或邮箱及密码，成功后返回用户。
func (service *Service) Login(
	ctx context.Context,
	input LoginInput,
) (*models.User, error) {
	identifier := strings.ToLower(strings.TrimSpace(input.Identifier))

	if identifier == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := service.repository.FindByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find login user: %w", err)
	}

	if user.Password == nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(*user.Password),
		[]byte(input.Password),
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// CurrentUser 根据 JWT 中的用户 ID 查询当前用户。
func (service *Service) CurrentUser(
	ctx context.Context,
	userID string,
) (*models.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	user, err := service.repository.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find current user: %w", err)
	}

	return user, nil
}
