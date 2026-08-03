package main

import (
	"log/slog"
	"net/http"
	"os"
	"sy_chat/internal/config"
	"sy_chat/internal/routes"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load application config", "error", err)
		os.Exit(1)
	}
	router := routes.RouterInit(cfg)

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
