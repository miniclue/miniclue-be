package router

import (
	"app/internal/api/v1/handler"
	"app/internal/config"
	"app/internal/middleware"
	"app/internal/pubsub"
	"app/internal/repository"
	"app/internal/service"
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsmiddleware "github.com/aws/smithy-go/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

func New(cfg *config.Config, logger zerolog.Logger) (http.Handler, *pgxpool.Pool, error) {
	logger.Info().Msg("Router initialized")

	// Log environment variables for debugging
	logger.Info().Str("environment", cfg.Environment).Msg("App environment loaded")
	logger.Info().Str("db_connection_string_port_check", getPortFromDSN(cfg.DBConnectionString)).Msg("DB connection string port")

	// 2. Open DB connection (connection pooling)
	dsn := cfg.DBConnectionString
	var pool *pgxpool.Pool
	var err error

	if cfg.Environment == "development" {
		// For development, use a direct connection but still manage it with pgxpool for consistency.
		if !strings.Contains(dsn, "sslmode") {
			separator := " "
			if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
				if strings.Contains(dsn, "?") {
					separator = "&"
				} else {
					separator = "?"
				}
			}
			dsn += separator + "sslmode=disable"
		}
		pool, err = pgxpool.New(context.Background(), dsn)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create connection pool for development")
			return nil, nil, err
		}
	} else {
		// For staging/production, use the transaction pooler with prepared statements disabled.
		dbConfig, parseErr := pgxpool.ParseConfig(dsn)
		if parseErr != nil {
			logger.Fatal().Err(parseErr).Msg("Failed to parse DB connection string for production")
			return nil, nil, parseErr
		}
		dbConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

		pool, err = pgxpool.NewWithConfig(context.Background(), dbConfig)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create connection pool for production")
			return nil, nil, err
		}
	}

	// Ping the database to ensure connection is valid
	if err := pool.Ping(context.Background()); err != nil {
		logger.Fatal().Err(err).Msg("Failed to ping DB")
		return nil, nil, err
	}
	logger.Info().Msg("Database connection successful")

	// 3. Initialize S3 client
	s3Config, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
		awsconfig.WithAPIOptions([]func(*awsmiddleware.Stack) error{removeDisableGzip()}),
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

	// 5. Initialize Pub/Sub publisher
	pubSubPublisher, err := pubsub.NewPublisher(context.Background(), cfg)
	if err != nil {
		logger.Fatal().Msgf("Failed to create Pub/Sub publisher: %v", err)
		return nil, nil, err
	}

	// 6. Initialize repositories & services & handlers
	userRepo := repository.NewUserRepo(pool)
	lectureRepo := repository.NewLectureRepository(pool)
	usageRepo := repository.NewUsageRepo(pool) // new usage events repo
	courseRepo := repository.NewCourseRepo(pool)
	summaryRepo := repository.NewSummaryRepository(pool)
	explanationRepo := repository.NewExplanationRepository(pool)
	noteRepo := repository.NewNoteRepository(pool)
	dlqRepo := repository.NewDLQRepository(pool)
	subscriptionRepo := repository.NewSubscriptionRepo(pool)

	subscriptionSvc := service.NewSubscriptionService(subscriptionRepo, logger)
	stripeSvc := service.NewStripeService(cfg, userRepo, subscriptionSvc, logger)
	userSvc := service.NewUserService(userRepo, courseRepo, lectureRepo, subscriptionRepo, stripeSvc, logger)
	lectureSvc := service.NewLectureService(lectureRepo, userRepo, usageRepo, s3Client, cfg.S3Bucket, pubSubPublisher, cfg.PubSubIngestionTopic, logger)
	courseSvc := service.NewCourseService(courseRepo, lectureSvc, logger)
	summarySvc := service.NewSummaryService(summaryRepo, logger)
	explanationSvc := service.NewExplanationService(explanationRepo, logger)
	noteSvc := service.NewNoteService(noteRepo, logger)
	dlqSvc := service.NewDLQService(dlqRepo, logger)

	userHandler := handler.NewUserHandler(userSvc, validate, logger)
	courseHandler := handler.NewCourseHandler(courseSvc, validate, logger)
	lectureHandler := handler.NewLectureHandler(lectureSvc, courseSvc, summarySvc, explanationSvc, noteSvc, validate, cfg.S3URL, cfg.S3Bucket, logger)
	dlqHandler := handler.NewDLQHandler(dlqSvc, logger)
	subscriptionHandler := handler.NewSubscriptionHandler(stripeSvc, subscriptionSvc, logger)
	stripeHandler := handler.NewStripeHandler(stripeSvc)

	// 7. Initialize middleware
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)
	isLocalDev := cfg.PubSubEmulatorHost != ""
	pubsubAuthMiddleware := middleware.PubSubAuthMiddleware(isLocalDev, cfg.DLQEndpointURL, cfg.PubSubPushServiceAccountEmail, logger)
	subscriptionMiddleware := middleware.SubscriptionLimitMiddleware(subscriptionSvc)

	// 8. Create ServeMux router
	mux := http.NewServeMux()

	// Create a subrouter for API v1 with the /api/v1 prefix
	apiV1Mux := http.NewServeMux()
	userHandler.RegisterRoutes(apiV1Mux, authMiddleware)
	courseHandler.RegisterRoutes(apiV1Mux, authMiddleware)
	lectureHandler.RegisterRoutes(apiV1Mux, authMiddleware, subscriptionMiddleware)
	dlqHandler.RegisterRoutes(apiV1Mux, pubsubAuthMiddleware)
	subscriptionHandler.RegisterRoutes(apiV1Mux, authMiddleware)
	stripeHandler.RegisterRoutes(apiV1Mux)

	// Mount the API v1 routes under /v1
	mux.Handle("/v1/", http.StripPrefix("/v1", apiV1Mux))

	// Add Swagger documentation
	mux.HandleFunc("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger/swagger.json")
	})
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs/swagger/swagger-ui"))))

	// Redirect /api/* to /v1/* for backward compatibility
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/api/")
		http.Redirect(w, r, "/v1/"+rest, http.StatusMovedPermanently)
	})

	// Redirect all other root-level requests to /v1/{path}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Avoid redirect loops by checking if already under /v1 or /swagger or /api
		if strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/swagger/") || strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/v1"+r.URL.Path, http.StatusMovedPermanently)
	})

	// 8. Apply CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins for development
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            false, // Enable debug logging for CORS
	})

	return middleware.LoggerMiddleware(c.Handler(mux)), pool, nil
}

// getPortFromDSN is a helper function to extract the port from a DSN string.
// It is intended for debugging purposes.
func getPortFromDSN(dsn string) string {
	parts := strings.Split(dsn, ":")
	for i, part := range parts {
		if strings.Contains(part, "@") {
			// This part contains user:pass@host, next part is port
			if len(parts) > i+1 {
				portAndDB := strings.Split(parts[i+1], "/")
				if len(portAndDB) > 0 {
					return portAndDB[0]
				}
			}
		}
	}
	return "not_found"
}

// removeDisableGzip is a workaround for S3 signature errors with some S3-compatible services.
// See: https://github.com/supabase/storage/issues/577
func removeDisableGzip() func(*awsmiddleware.Stack) error {
	return func(stack *awsmiddleware.Stack) error {
		// Only remove the middleware if it exists.
		// This makes the client more robust, especially for operations like presigned URLs
		// that might inspect the middleware stack.
		if _, ok := stack.Finalize.Get("DisableAcceptEncodingGzip"); ok {
			_, err := stack.Finalize.Remove("DisableAcceptEncodingGzip")
			return err
		}
		return nil
	}
}
