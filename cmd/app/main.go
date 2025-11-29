package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/internal/api/v1/router"
	"app/internal/config"
	"app/internal/logger"

	"github.com/joho/godotenv"
)

// @title MiniClue API
// @version 1.0
// @description MiniClue API documentation
// @host localhost:8080
// @BasePath /api/v1
// @Schemes http https

func main() {
	// It's important to load the .env file before anything else,
	// so that components like the logger can initialize correctly based on the environment.
	// We deliberately ignore the error from godotenv.Load(), as it's expected
	// that a .env file may not be present in all environments (e.g., production).
	_ = godotenv.Load()

	logger := logger.New()

	// 1. Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Error loading config: %v", err)
	}

	// 2. Build router (and get DB connection)
	r, pool, err := router.New(cfg, logger)
	if err != nil {
		logger.Fatal().Msgf("Failed to build router: %v", err)
	}
	defer pool.Close()

	// Determine port for HTTP server.
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}

	// 3. Create HTTP server
	// WriteTimeout is set to 5 minutes to accommodate long-running streaming responses
	// (e.g., AI chat streaming which involves query rewriting, RAG retrieval, and LLM streaming)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	// 4. Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Msgf("Listen: %s\n", err)
		}
	}()

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Msgf("Server forced to shutdown: %v", err)
	}
}
