package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// GitHub Secrets
	// Supabase
	DBConnectionString string `envconfig:"DB_CONNECTION_STRING" required:"true"`
	JWTSecret          string `envconfig:"SUPABASE_JWT_SECRET" required:"true"`
	S3URL              string `envconfig:"SUPABASE_S3_URL" required:"true"`
	S3Bucket           string `envconfig:"SUPABASE_S3_BUCKET" required:"true"`
	S3Region           string `envconfig:"SUPABASE_S3_REGION" required:"true"`
	S3AccessKey        string `envconfig:"SUPABASE_S3_ACCESS_KEY" required:"true"`
	S3SecretKey        string `envconfig:"SUPABASE_S3_SECRET_KEY" required:"true"`

	// Pub/Sub
	DLQEndpointURL                string `envconfig:"DLQ_ENDPOINT_URL"`
	PubSubPushServiceAccountEmail string `envconfig:"PUBSUB_PUSH_SERVICE_ACCOUNT_EMAIL"`

	// Pub/Sub Publisher - GCP Project IDs
	GCPProjectIDLocal   string `envconfig:"GCP_PROJECT_ID_LOCAL"`
	GCPProjectIDStaging string `envconfig:"GCP_PROJECT_ID_STAGING"`
	GCPProjectIDProd    string `envconfig:"GCP_PROJECT_ID_PROD"`

	// Local Secrets
	// Environment
	Environment string `envconfig:"ENV"`

	// API
	Port string `envconfig:"PORT" default:"8080"`

	// Pub/Sub
	PubSubIngestionTopic string `envconfig:"PUBSUB_INGESTION_TOPIC" default:"ingestion"`
	PubSubEmulatorHost   string `envconfig:"PUBSUB_EMULATOR_HOST"`

	// Pub/Sub - Push Endpoint URLs
	// For local dev, these are derived from constants in the setup-pubsub command.
	// For staging/prod, these should be the base URL of your API gateway (e.g., https://api.miniclue.com)
	APIBaseURLStaging string `envconfig:"API_BASE_URL_STAGING"`
	APIBaseURLProd    string `envconfig:"API_BASE_URL_PROD"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetGCPProjectID returns the appropriate GCP project ID based on the environment.
// Uses the same logic as Pub/Sub: local if emulator host is set, otherwise staging (preferred) or prod.
func (c *Config) GetGCPProjectID() string {
	if c.PubSubEmulatorHost != "" {
		return c.GCPProjectIDLocal
	}
	if c.GCPProjectIDStaging != "" {
		return c.GCPProjectIDStaging
	}
	return c.GCPProjectIDProd
}
