package router

import (
	"app/internal/api/v1/handler"
	"app/internal/config"
	"app/internal/logger"
	"app/internal/middleware"
	"app/internal/repository"
	"app/internal/service"
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
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

	// Ping the database to ensure connection is valid
	if err := db.Ping(); err != nil {
		logger.Fatal().Msgf("Failed to ping DB: %v", err)
		return nil, nil, err
	}
	logger.Info().Msg("Database connection successful")

	// Set reasonable connection pool limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// 3. Initialize S3 client
	s3Config, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
	)
	if err != nil {
		logger.Fatal().Msgf("Failed to load S3 config: %v", err)
	}
	s3Client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3URL)
		o.UsePathStyle = true
	})

	// 4. Initialize validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// 5. Initialize repositories & services & handlers
	userRepo := repository.NewUserRepo(db)
	lectureRepo := repository.NewLectureRepository(db)
	courseRepo := repository.NewCourseRepo(db)
	summaryRepo := repository.NewSummaryRepository(db)
	explanationRepo := repository.NewExplanationRepository(db)
	noteRepo := repository.NewNoteRepository(db)

	userSvc := service.NewUserService(userRepo, courseRepo, lectureRepo)
	lectureSvc := service.NewLectureService(lectureRepo, s3Client, cfg.S3Bucket)
	courseSvc := service.NewCourseService(courseRepo, lectureSvc)
	summarySvc := service.NewSummaryService(summaryRepo)
	explanationSvc := service.NewExplanationService(explanationRepo)
	noteSvc := service.NewNoteService(noteRepo)

	userHandler := handler.NewUserHandler(userSvc, validate)
	courseHandler := handler.NewCourseHandler(courseSvc, validate)
	lectureHandler := handler.NewLectureHandler(lectureSvc, courseSvc, summarySvc, explanationSvc, noteSvc, validate, cfg.S3URL, cfg.S3Bucket)

	// 6. Initialize middleware
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	// 7. Create ServeMux router
	mux := http.NewServeMux()

	// Create a subrouter for API v1 with the /api/v1 prefix
	apiV1Mux := http.NewServeMux()
	userHandler.RegisterRoutes(apiV1Mux, authMiddleware)
	courseHandler.RegisterRoutes(apiV1Mux, authMiddleware)
	lectureHandler.RegisterRoutes(apiV1Mux, authMiddleware)

	// Mount the API v1 routes under /api/v1
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1Mux))

	// Add Swagger documentation
	mux.HandleFunc("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger/swagger.json")
	})
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs/swagger/swagger-ui"))))

	// Handle /api and all its subpaths
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Get the rest of the path after /api
		restOfPath := r.URL.Path[4:] // Remove "/api" from the beginning
		http.Redirect(w, r, "/api/v1"+restOfPath, http.StatusMovedPermanently)
	})

	// 8. Apply CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins for development
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false, // Enable debug logging for CORS
	})

	return middleware.LoggerMiddleware(c.Handler(mux)), db, nil
}
