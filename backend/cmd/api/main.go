package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sy_chat/internal/config"
	"sy_chat/internal/database"
	"sy_chat/internal/routes"
)

const shutdownTimeout = 10 * time.Second

// @title SY Chat API
// @version 1.0
// @description SY Chat 的认证、会话、聊天、模型和分享接口。
// @description 登录成功后，浏览器通过 sy_chat_session HttpOnly Cookie 自动访问受保护接口。
// @BasePath /api/v1
// @schemes http
func main() {
	if err := run(); err != nil {
		slog.Error("API server exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load application config: %w", err)
	}

	gormDB, err := database.OpenMySQL(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connect to mysql: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("get mysql connection pool: %w", err)
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Error("close mysql connection", "error", err)
		}
	}()

	slog.Info("MySQL connection established")

	// 开发环境启动时自动同步数据表结构。
	if cfg.AppEnv == "development" {
		if err := database.Migrate(gormDB); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
		slog.Info("Database migration completed")
	}

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           routes.RouterInit(cfg, gormDB),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	slog.Info(
		"API server started",
		"environment",
		cfg.AppEnv,
		"address",
		"http://localhost:"+cfg.ServerPort,
	)

	signalContext, stopSignals := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stopSignals()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)

	case <-signalContext.Done():
		slog.Info("API server shutdown started")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownContext); err != nil {
		// 超时后强制关闭仍未结束的连接，例如长时间无响应的 AI 请求。
		_ = server.Close()
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverErrors; err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("stop HTTP server: %w", err)
	}

	slog.Info("API server stopped")
	return nil
}
