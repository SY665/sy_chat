package main

import (
	"log/slog"
	"net/http"
	"os"
	"sy_chat/internal/config"
	"sy_chat/internal/database"
	"sy_chat/internal/routes"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load application config", "error", err)
		os.Exit(1)
	}

	gormDB, err := database.OpenMySQL(cfg.DatabaseDSN)
	if err != nil {
		slog.Error("connect to mysql", "error", err)
		os.Exit(1)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		slog.Error("get mysql connection pool", "error", err)
		os.Exit(1)
	}

	// 程序退出时释放数据库连接池。
	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.Error("close mysql connection", "error", err)
		}
	}()

	slog.Info("MySQL connection established")

	// 开发环境启动时自动同步数据表结构。
	if cfg.AppEnv == "development" {
		if err := database.Migrate(gormDB); err != nil {
			slog.Error("migrate database", "error", err)
			os.Exit(1)
		}
		slog.Info("Database migration completed")
	}

	router := routes.RouterInit(cfg, gormDB)

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info(
		"API server started",
		"environment",
		cfg.AppEnv,
		"address",
		"http://localhost:"+cfg.ServerPort,
	)

	if err := server.ListenAndServe(); err != nil {
		slog.Error("API server stopped", "error", err)
		os.Exit(1)
	}
}
