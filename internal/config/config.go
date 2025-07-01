package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	DBHost      string `envconfig:"DB_HOST" required:"true"`
	DBPort      int    `envconfig:"DB_PORT" default:"5432"`
	DBUser      string `envconfig:"DB_USER" required:"true"`
	DBPassword  string `envconfig:"DB_PASSWORD" required:"true"`
	DBName      string `envconfig:"DB_NAME" required:"true"`
	JWTSecret   string `envconfig:"SUPABASE_LOCAL_JWT_SECRET" required:"true"`
	S3URL       string `envconfig:"SUPABASE_LOCAL_S3_URL" required:"true"`
	S3Bucket    string `envconfig:"SUPABASE_LOCAL_S3_BUCKET" required:"true"`
	S3Region    string `envconfig:"SUPABASE_LOCAL_S3_REGION" required:"true"`
	S3AccessKey string `envconfig:"SUPABASE_LOCAL_S3_ACCESS_KEY" required:"true"`
	S3SecretKey string `envconfig:"SUPABASE_LOCAL_S3_SECRET_KEY" required:"true"`

	// Python service settings
	PythonServiceBaseURL string `envconfig:"PYTHON_SERVICE_BASE_URL" required:"true"`

	// Ingestion orchestrator settings
	IngestionQueueName           string `envconfig:"INGESTION_QUEUE_NAME" default:"ingestion_queue"`
	IngestionPollTimeoutSec      int    `envconfig:"INGESTION_POLL_TIMEOUT_SEC" default:"30"`
	IngestionPollMaxMsg          int    `envconfig:"INGESTION_POLL_MAX_MSG" default:"1"`
	IngestionMaxRetries          int    `envconfig:"INGESTION_MAX_RETRIES" default:"5"`
	IngestionBackoffInitialSec   int    `envconfig:"INGESTION_BACKOFF_INITIAL_SEC" default:"1"`
	IngestionBackoffMaxSec       int    `envconfig:"INGESTION_BACKOFF_MAX_SEC" default:"60"`
	IngestionRequestTimeoutSec   int    `envconfig:"INGESTION_REQUEST_TIMEOUT_SEC" default:"300"`
	IngestionDeadLetterQueueName string `envconfig:"INGESTION_DEAD_LETTER_QUEUE_NAME" default:"ingestion_queue_dlq"`

	// Embedding orchestrator settings
	EmbeddingQueueName           string `envconfig:"EMBEDDING_QUEUE_NAME" default:"embedding_queue"`
	EmbeddingPollTimeoutSec      int    `envconfig:"EMBEDDING_POLL_TIMEOUT_SEC" default:"30"`
	EmbeddingPollMaxMsg          int    `envconfig:"EMBEDDING_POLL_MAX_MSG" default:"1"`
	EmbeddingMaxRetries          int    `envconfig:"EMBEDDING_MAX_RETRIES" default:"5"`
	EmbeddingBackoffInitialSec   int    `envconfig:"EMBEDDING_BACKOFF_INITIAL_SEC" default:"1"`
	EmbeddingBackoffMaxSec       int    `envconfig:"EMBEDDING_BACKOFF_MAX_SEC" default:"60"`
	EmbeddingRequestTimeoutSec   int    `envconfig:"EMBEDDING_REQUEST_TIMEOUT_SEC" default:"120"`
	EmbeddingDeadLetterQueueName string `envconfig:"EMBEDDING_DEAD_LETTER_QUEUE_NAME" default:"embedding_queue_dlq"`

	// Explanation orchestrator settings
	ExplanationQueueName           string `envconfig:"EXPLANATION_QUEUE_NAME" default:"explanation_queue"`
	ExplanationPollTimeoutSec      int    `envconfig:"EXPLANATION_POLL_TIMEOUT_SEC" default:"30"`
	ExplanationPollMaxMsg          int    `envconfig:"EXPLANATION_POLL_MAX_MSG" default:"1"`
	ExplanationMaxRetries          int    `envconfig:"EXPLANATION_MAX_RETRIES" default:"5"`
	ExplanationBackoffInitialSec   int    `envconfig:"EXPLANATION_BACKOFF_INITIAL_SEC" default:"1"`
	ExplanationBackoffMaxSec       int    `envconfig:"EXPLANATION_BACKOFF_MAX_SEC" default:"60"`
	ExplanationRequestTimeoutSec   int    `envconfig:"EXPLANATION_REQUEST_TIMEOUT_SEC" default:"120"`
	ExplanationDeadLetterQueueName string `envconfig:"EXPLANATION_DEAD_LETTER_QUEUE_NAME" default:"explanation_queue_dlq"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
