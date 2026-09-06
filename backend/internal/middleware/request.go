package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	requestIDContextKey = "requestID"
	requestIDHeader     = "X-Request-ID"
)

// RequestID 为每个请求生成标识，也允许客户端传入合法的 UUID。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(requestIDHeader))
		if _, err := uuid.Parse(requestID); err != nil {
			requestID = uuid.NewString()
		}

		c.Set(requestIDContextKey, requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

func CurrentRequestID(c *gin.Context) string {
	requestID, _ := c.Get(requestIDContextKey)
	value, _ := requestID.(string)
	return value
}

// RequestLogger 在请求完成后记录状态码、耗时与请求 ID。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		slog.Info(
			"HTTP request",
			"request_id", CurrentRequestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(startedAt),
			"client_ip", c.ClientIP(),
		)
	}
}
