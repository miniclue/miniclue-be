package router

import (
	"app/internal/api/v1/handler"
	"app/internal/config"
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

	// Create a subrouter for API v1 with the /api/v1 prefix
	apiV1Mux := http.NewServeMux()
	userHandler.RegisterRoutes(apiV1Mux, authMiddleware)

	// Mount the API v1 routes under /api/v1
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1Mux))

	// Handle /api and all its subpaths
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Get the rest of the path after /api
		restOfPath := r.URL.Path[4:] // Remove "/api" from the beginning
		http.Redirect(w, r, "/api/v1"+restOfPath, http.StatusMovedPermanently)
	})

	return mux, db, nil
}
