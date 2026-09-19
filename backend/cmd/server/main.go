// Package main is the entry point for the Live Polling API server.
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

	"github.com/live-polling-app/backend/config"
	"github.com/live-polling-app/backend/database"
	"github.com/live-polling-app/backend/routes"
	"github.com/live-polling-app/backend/websocket"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize MongoDB
	var db *database.DB
	if cfg.SkipDB {
		slog.Warn("SKIP_DB=true is set; skipping MongoDB initialization")
	} else {
		mongoDB, err := database.Connect(ctx, cfg)
		if err != nil {
			slog.Error("failed to initialize mongodb", "error", err.Error())
			slog.Error("Ensure MongoDB is running at MONGODB_URI or set SKIP_DB=true in .env for offline mode")
			os.Exit(1)
		}
		db = mongoDB
		defer func() {
			closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer closeCancel()
			if err := db.Close(closeCtx); err != nil {
				slog.Error("error closing mongodb connection", "error", err.Error())
			}
		}()
	}

	// Initialize Redis
	var redisClient *database.RedisClient
	if cfg.SkipRedis {
		slog.Warn("SKIP_REDIS=true is set; skipping Redis initialization")
	} else {
		rc, err := database.ConnectRedis(ctx, cfg)
		if err != nil {
			slog.Warn("failed to connect to redis (realtime pub/sub cache disabled)", "error", err.Error())
		} else {
			redisClient = rc
			defer func() {
				if err := redisClient.Close(); err != nil {
					slog.Error("error closing redis connection", "error", err.Error())
				}
			}()
		}
	}

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub(redisClient)

	router := routes.Setup(cfg, db, redisClient, wsHub)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("server starting", "port", cfg.Port, "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		slog.Error("server startup error", "error", err.Error())
		os.Exit(1)

	case sig := <-shutdown:
		slog.Info("shutdown signal received", "signal", sig.String())

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("forcing server shutdown", "error", err.Error())
			_ = server.Close()
			os.Exit(1)
		}
		slog.Info("server shut down gracefully")
	}
}
