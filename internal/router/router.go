package router

import (
	"app/internal/config"
	"app/internal/handler"
	"app/internal/logger"
	"app/internal/middleware"
	"app/internal/repository"
	"app/internal/service"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
)

func New(cfg *config.Config) (http.Handler, *sql.DB, error) {
	// 1. Initialize logger
	logger := logger.New()
	logger.Info().Msg("Router initialized")

	// 2. Open DB connection (connection pooling)
	dsn :=
		"host=" + cfg.DBHost +
			" port=" + strconv.Itoa(cfg.DBPort) +
			" user=" + cfg.DBUser +
			" password=" + cfg.DBPassword +
			" dbname=" + cfg.DBName +
			" sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatal().Msgf("Failed to open DB connection: %v", err)
		return nil, nil, err
	}
	// Set reasonable connection pool limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// 3. Initialize validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// 4. Initialize repositories & services & handlers
	userRepo := repository.NewUserRepo(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc, validate)

	// 4. Initialize middleware
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	// 5. Create ServeMux router
	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux, authMiddleware)

	return mux, db, nil
}
