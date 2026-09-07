// Package routes 负责注册中间件和 HTTP 路由。
package routes

import (
	"time"

	_ "sy_chat/docs"
	"sy_chat/internal/config"
	"sy_chat/internal/handlers"
	"sy_chat/internal/middleware"
	"sy_chat/internal/repositories"
	aiservice "sy_chat/internal/services/ai"
	"sy_chat/internal/services/auth"
	"sy_chat/internal/services/chat"
	conversationservice "sy_chat/internal/services/conversation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// RouterInit 创建并配置 Gin 路由。
func RouterInit(cfg config.Config, db *gorm.DB) *gin.Engine {
	// 使用 gin.New 可以明确决定启用哪些全局中间件。
	router := gin.New()

	// Logger 记录请求，Recovery 防止 panic 导致整个服务退出。请求 ID 必须先于日志中间件执行。
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger())
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

		ExposeHeaders: []string{
			"X-Request-ID",
		},

		// 后续 JWT 存入 HttpOnly Cookie 时，浏览器需要携带凭证。
		AllowCredentials: true,

		// 浏览器可缓存预检请求，减少 OPTIONS 请求数量。
		MaxAge: 12 * time.Hour,
	}))

	// Swagger 仅在非生产环境开放，避免生产环境直接暴露接口文档。
	if cfg.AppEnv != "production" {
		router.GET(
			"/swagger/*any",
			ginSwagger.WrapHandler(swaggerFiles.Handler),
		)
	}

	messageRepository := repositories.NewMessageRepository(db)
	aiProvider := newAIProvider(cfg)

	defaultModelID := cfg.SiliconFlowModel
	availableModelIDs := cfg.SiliconFlowModels
	thinkingModelIDs := cfg.SiliconFlowThinkingModels

	if cfg.AIProvider == "local" {
		defaultModelID = "local"
		availableModelIDs = []string{"local"}
		thinkingModelIDs = nil
	}

	chatService := chat.NewService(
		messageRepository,
		aiProvider,
		defaultModelID,
		availableModelIDs,
		thinkingModelIDs,
	)
	chatHandler := handlers.NewChatHandler(chatService)

	modelHandler := handlers.NewModelHandler(
		cfg.AIProvider,
		availableModelIDs,
		defaultModelID,
		thinkingModelIDs,
	)

	userRepository := repositories.NewUserRepository(db)
	authService := auth.NewService(userRepository)
	tokenManager := auth.NewTokenManager(
		cfg.JWTSecret,
		cfg.JWTExpiresIn,
	)
	authHandler := handlers.NewAuthHandler(
		authService,
		tokenManager,
		cfg.JWTExpiresIn,
		cfg.CookieSecure,
	)

	conversationRepository := repositories.NewConversationRepository(db)
	conversationService := conversationservice.NewService(
		conversationRepository,
	)
	conversationHandler := handlers.NewConversationHandler(
		conversationService,
	)

	// 所有业务接口统一使用 /api/v1 前缀，方便未来升级 API。
	api := router.Group("/api/v1")
	{
		// 分享读取接口通过随机令牌授权，不要求登录。
		api.GET("/shares/:token", conversationHandler.GetShared)

		api.GET(
			"/models",
			middleware.RequireAuth(tokenManager),
			modelHandler.List,
		)
		api.GET("/health", handlers.Health)

		api.POST(
			"/attachments",
			middleware.RequireAuth(tokenManager),
			handlers.UploadAttachment,
		)

		api.POST(
			"/chat",
			middleware.RequireAuth(tokenManager),
			chatHandler.Send,
		)

		api.POST(
			"/chat/stream",
			middleware.RequireAuth(tokenManager),
			chatHandler.Stream,
		)

		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/logout", authHandler.Logout)

			protectedAuthRoutes := authRoutes.Group("")
			protectedAuthRoutes.Use(middleware.RequireAuth(tokenManager))
			{
				protectedAuthRoutes.GET("/me", authHandler.Me)
			}
		}

		conversationRoutes := api.Group("/conversations")
		conversationRoutes.Use(middleware.RequireAuth(tokenManager))
		{
			conversationRoutes.POST("", conversationHandler.Create)
			conversationRoutes.GET("", conversationHandler.List)
			conversationRoutes.GET("/:id", conversationHandler.Get)
			conversationRoutes.DELETE("/:id/messages/:messageId/tail", chatHandler.TruncateFromMessage)
			conversationRoutes.PATCH("/:id", conversationHandler.UpdateTitle)
			conversationRoutes.PATCH("/:id/pin", conversationHandler.UpdatePinned)
			conversationRoutes.POST("/:id/share", conversationHandler.Share)
			conversationRoutes.DELETE("/:id/share", conversationHandler.Unshare)
			conversationRoutes.DELETE("/:id", conversationHandler.Delete)
		}
	}

	return router
}

// newAIProvider 是应用的 AI 依赖装配点。
// Provider 的选择只由启动配置决定，业务层不读取环境变量。
func newAIProvider(cfg config.Config) aiservice.Provider {
	if cfg.AIProvider == "siliconflow" {
		return aiservice.NewSiliconFlowProvider(
			cfg.SiliconFlowAPIKey,
			cfg.SiliconFlowBaseURL,
			cfg.SiliconFlowModel,
			cfg.AIRequestTimeout,
		)
	}

	return aiservice.NewLocalProvider()
}
