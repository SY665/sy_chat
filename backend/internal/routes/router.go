// Package routes 负责注册中间件和 HTTP 路由。
package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"sy_chat/internal/config"
	"sy_chat/internal/handlers"
	"sy_chat/internal/services/chat"
)

// RouterInit 创建并配置 Gin 路由。
func RouterInit(cfg config.Config) *gin.Engine {
	// 使用 gin.New 可以明确决定启用哪些全局中间件。
	router := gin.New()

	// Logger 记录请求，Recovery 防止 panic 导致整个服务退出。
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		// 使用明确的前端地址，不使用 "*"，避免任意网站访问 API。
		AllowOrigins: []string{
			cfg.FrontendOrigin,
		},

		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},

		// 后续 JWT 存入 HttpOnly Cookie 时，浏览器需要携带凭证。
		AllowCredentials: true,

		// 浏览器可缓存预检请求，减少 OPTIONS 请求数量。
		MaxAge: 12 * time.Hour,
	}))

	chatService := chat.NewService()
	chatHandler := handlers.NewChatHandler(chatService)

	// 所有业务接口统一使用 /api/v1 前缀，方便未来升级 API。
	api := router.Group("/api/v1")
	{
		api.GET("/health", handlers.Health)
		api.POST("/chat", chatHandler.Send)
	}

	return router
}
