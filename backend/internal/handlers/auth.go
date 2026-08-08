package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"sy_chat/internal/middleware"
	"sy_chat/internal/models"
	"sy_chat/internal/response"
	"sy_chat/internal/services/auth"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const sessionCookieName = "sy_chat_session"

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// authUserData 只返回前端需要的公开用户信息。
type authUserData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

type authData struct {
	User authUserData `json:"user"`
}

type AuthHandler struct {
	authService  *auth.Service
	tokenManager *auth.TokenManager
	tokenMaxAge  time.Duration
	cookieSecure bool
}

func NewAuthHandler(
	authService *auth.Service,
	tokenManager *auth.TokenManager,
	tokenMaxAge time.Duration,
	cookieSecure bool,
) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		tokenManager: tokenManager,
		tokenMaxAge:  tokenMaxAge,
		cookieSecure: cookieSecure,
	}
}

// Register 创建用户，并在注册成功后自动登录。
func (handler *AuthHandler) Register(c *gin.Context) {
	var request registerRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_REGISTER_INPUT",
			"用户名、邮箱或密码格式不正确",
		)
		return
	}

	user, err := handler.authService.Register(
		c.Request.Context(),
		auth.RegisterInput{
			Username: request.Username,
			Email:    request.Email,
			Password: request.Password,
		},
	)

	if err != nil {
		handler.handleRegisterError(c, err)
		return
	}

	if err := handler.startSession(c, user); err != nil {
		slog.Error("start registration session", "error", err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"注册成功，但登录状态创建失败",
		)
		return
	}

	response.JSON(c, http.StatusCreated, authData{
		User: newAuthUserData(user),
	})
}

// Login 验证账号密码，并将 JWT 写入 HttpOnly Cookie。
func (handler *AuthHandler) Login(c *gin.Context) {
	var request loginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_LOGIN_INPUT",
			"请输入用户名或邮箱和密码",
		)
		return
	}

	user, err := handler.authService.Login(
		c.Request.Context(),
		auth.LoginInput{
			Identifier: request.Identifier,
			Password:   request.Password,
		},
	)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			response.Error(
				c,
				http.StatusUnauthorized,
				"INVALID_CREDENTIALS",
				"账号或密码错误",
			)
			return
		}

		slog.Error("login user", "error", err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"登录失败，请稍后重试",
		)
		return
	}

	if err := handler.startSession(c, user); err != nil {
		slog.Error("start login session", "error", err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"登录状态创建失败",
		)
		return
	}

	response.JSON(c, http.StatusOK, authData{
		User: newAuthUserData(user),
	})
}

// Logout 删除浏览器中的登录 Cookie。
func (handler *AuthHandler) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	response.JSON(c, http.StatusOK, gin.H{
		"message": "已退出登录",
	})
}

func (handler *AuthHandler) startSession(
	c *gin.Context,
	user *models.User,
) error {
	token, err := handler.tokenManager.Generate(user.ID)
	if err != nil {
		return err
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(handler.tokenMaxAge.Seconds()),
		Expires:  time.Now().Add(handler.tokenMaxAge),
		HttpOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (handler *AuthHandler) handleRegisterError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, auth.ErrUsernameExists):
		response.Error(c, http.StatusConflict, "USERNAME_EXISTS", "用户名已存在")

	case errors.Is(err, auth.ErrEmailExists):
		response.Error(c, http.StatusConflict, "EMAIL_EXISTS", "邮箱已存在")

	case errors.Is(err, auth.ErrUsernameRequired),
		errors.Is(err, auth.ErrUsernameInvalid):
		response.Error(
			c,
			http.StatusBadRequest,
			"INVALID_USERNAME",
			"用户名只能包含字母、数字和下划线，长度为 3 到 30 个字符",
		)

	case errors.Is(err, auth.ErrEmailRequired),
		errors.Is(err, auth.ErrEmailInvalid):
		response.Error(c, http.StatusBadRequest, "INVALID_EMAIL", "邮箱格式不正确")

	case errors.Is(err, auth.ErrPasswordTooShort):
		response.Error(
			c,
			http.StatusBadRequest,
			"PASSWORD_TOO_SHORT",
			"密码不能少于 8 个字符",
		)

	case errors.Is(err, auth.ErrPasswordTooLong):
		response.Error(
			c,
			http.StatusBadRequest,
			"PASSWORD_TOO_LONG",
			"密码不能超过 72 个字节",
		)

	default:
		slog.Error("register user", "error", err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"注册失败，请稍后重试",
		)
	}
}

func newAuthUserData(user *models.User) authUserData {
	return authUserData{
		ID:       user.ID,
		Username: stringValue(user.Username),
		Email:    stringValue(user.Email),
		Name:     stringValue(user.Name),
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

// Me 返回当前登录用户的公开信息。
func (handler *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(
			c,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"请先登录",
		)
		return
	}
	user, err := handler.authService.CurrentUser(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(
				c,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"登录用户不存在",
			)
			return
		}

		slog.Error("get current user", "error", err)
		response.Error(
			c,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"获取用户信息失败",
		)
		return
	}

	response.JSON(c, http.StatusOK, authData{
		User: newAuthUserData(user),
	})
}
