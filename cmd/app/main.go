package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/internal/config"
	"app/internal/logger"
	"app/internal/api/v1/router"

	"github.com/joho/godotenv"
)

func main() {
	logger := logger.New()

	// 1. Load configuration
	if err := godotenv.Load(); err != nil {
		logger.Warn().Msg("Warning: no .env file found")
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Error loading config: %v", err)
	}

	// 2. Build router (and get DB connection)
	r, db, err := router.New(cfg)
	if err != nil {
		logger.Fatal().Msgf("Failed to build router: %v", err)
	}
	defer db.Close()

	// 3. Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 4. Start server in a goroutine
	go func() {
		logger.Info().Msgf("🚀 Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Msgf("Listen: %s\n", err)
		}
	}()

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutdown signal received, exiting...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Msgf("Server forced to shutdown: %v", err)
	}
	logger.Info().Msg("Server shut down gracefully")
}
