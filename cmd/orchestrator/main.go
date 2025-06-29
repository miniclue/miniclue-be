package main

import (
	"context"
	"database/sql"
	"flag"
	"os/signal"
	"strconv"
	"syscall"

	"app/internal/config"
	"app/internal/logger"
	"app/internal/orchestrator/embedding"
	"app/internal/orchestrator/explanation"
	"app/internal/orchestrator/ingestion"
	"app/internal/orchestrator/summary"
	"app/internal/pgmq"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Parse mode flag
	mode := flag.String("mode", "", "Orchestrator mode: ingestion|embedding|explanation|summary")
	flag.Parse()

	// Initialize logger
	logger := logger.New()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Warn().Msg("Warning: no .env file found")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Msgf("Error loading config: %v", err)
	}

	// Initialize DB connection
	dsn := "host=" + cfg.DBHost +
		" port=" + strconv.Itoa(cfg.DBPort) +
		" user=" + cfg.DBUser +
		" password=" + cfg.DBPassword +
		" dbname=" + cfg.DBName +
		" sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatal().Msgf("Failed to open DB connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal().Msgf("Failed to ping DB: %v", err)
	}
	logger.Info().Msg("Database connection established")

	// Initialize PGMQ client
	pgmqClient := pgmq.New(db)
	logger.Info().Msg("PGMQ client initialized")

	// Set up context with graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Dispatch to the selected orchestrator
	var runErr error
	switch *mode {
	case "ingestion":
		runErr = ingestion.Run(ctx, logger, pgmqClient)
	case "embedding":
		runErr = embedding.Run(ctx, logger, pgmqClient)
	case "explanation":
		runErr = explanation.Run(ctx, logger, pgmqClient)
	case "summary":
		runErr = summary.Run(ctx, logger, pgmqClient)
	default:
		logger.Fatal().Msgf("Invalid mode: %s", *mode)
	}

	if runErr != nil {
		logger.Fatal().Msgf("%s orchestrator failed: %v", *mode, runErr)
	}

	logger.Info().Msgf("%s orchestrator stopped gracefully", *mode)
}
