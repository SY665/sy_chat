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

// RegisterRequest 描述用户注册参数。
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest 描述用户名或邮箱登录参数。
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

// AuthUserData 只返回前端需要的公开用户信息。
type AuthUserData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// AuthData 是认证接口成功后的响应数据。
type AuthData struct {
	User AuthUserData `json:"user"`
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

// Register godoc
// @Summary 注册用户
// @Description 创建用户，并通过 HttpOnly Cookie 自动建立登录会话。
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} response.Envelope{data=AuthData}
// @Failure 400 {object} response.Envelope
// @Failure 409 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Router /auth/register [post]
func (handler *AuthHandler) Register(c *gin.Context) {
	var request RegisterRequest

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

	response.JSON(c, http.StatusCreated, AuthData{
		User: newAuthUserData(user),
	})
}

// Login godoc
// @Summary 登录
// @Description 验证用户名或邮箱和密码，并写入 sy_chat_session HttpOnly Cookie。
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} response.Envelope{data=AuthData}
// @Failure 400 {object} response.Envelope
// @Failure 401 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Router /auth/login [post]
func (handler *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest

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

	response.JSON(c, http.StatusOK, AuthData{
		User: newAuthUserData(user),
	})
}

// Logout godoc
// @Summary 退出登录
// @Description 删除浏览器中的 sy_chat_session 登录 Cookie。
// @Tags Auth
// @Produce json
// @Success 200 {object} response.Envelope
// @Router /auth/logout [post]
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

func newAuthUserData(user *models.User) AuthUserData {
	return AuthUserData{
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

// Me godoc
// @Summary 获取当前用户
// @Description 根据 sy_chat_session 登录 Cookie 返回当前用户信息。
// @Tags Auth
// @Produce json
// @Success 200 {object} response.Envelope{data=AuthData}
// @Failure 401 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Router /auth/me [get]
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

	response.JSON(c, http.StatusOK, AuthData{
		User: newAuthUserData(user),
	})
}
