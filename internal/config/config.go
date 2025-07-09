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

	// Pub/Sub ingestion topic
	GCPProjectID         string `envconfig:"GCP_PROJECT_ID" required:"true"`
	PubSubIngestionTopic string `envconfig:"PUBSUB_INGESTION_TOPIC" default:"ingestion"`
	PubSubEmulatorHost   string `envconfig:"PUBSUB_EMULATOR_HOST"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
