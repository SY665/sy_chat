package middleware

import (
	"net/http"
	"sy_chat/internal/response"
	"sy_chat/internal/services/auth"

	"github.com/gin-gonic/gin"
)

const userIDContextKey = "authenticatedUserID"

// RequireAuth 验证 Cookie 中的 JWT，并把用户 ID 写入请求上下文。
func RequireAuth(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(auth.SessionCookieName)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		userID, err := tokenManager.Parse(token)
		if err != nil {
			abortUnauthorized(c)
			return
		}

		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

// CurrentUserID 读取认证中间件写入的用户 ID。
func CurrentUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(userIDContextKey)
	if !exists {
		return "", false
	}

	userID, ok := value.(string)
	return userID, ok && userID != ""
}

func abortUnauthorized(c *gin.Context) {
	response.Error(
		c,
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		"请先登录",
	)
	c.Abort()
}
