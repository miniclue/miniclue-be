package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Local & Github Secrets (Fill up for local development)
	DBConnectionString   string `envconfig:"DB_CONNECTION_STRING" required:"true"`
	JWTSecret            string `envconfig:"SUPABASE_JWT_SECRET" required:"true"`
	S3URL                string `envconfig:"SUPABASE_S3_URL" required:"true"`
	S3Bucket             string `envconfig:"SUPABASE_S3_BUCKET" required:"true"`
	S3Region             string `envconfig:"SUPABASE_S3_REGION" required:"true"`
	S3AccessKey          string `envconfig:"SUPABASE_S3_ACCESS_KEY" required:"true"`
	S3SecretKey          string `envconfig:"SUPABASE_S3_SECRET_KEY" required:"true"`
	Environment          string `envconfig:"ENV" default:"development"`
	PubSubIngestionTopic string `envconfig:"PUBSUB_INGESTION_TOPIC" default:"ingestion"`
	PythonServiceBaseURL string `envconfig:"PYTHON_SERVICE_BASE_URL" required:"true"`

	// Local Secrets (Fill up for local development)
	Port                       string `envconfig:"PORT" default:"8080"`
	PubSubEmulatorHost         string `envconfig:"PUBSUB_EMULATOR_HOST"`
	SupabaseAuthGoogleClientID string `envconfig:"SUPABASE_AUTH_GOOGLE_CLIENT_ID"`
	SupabaseAuthGoogleSecret   string `envconfig:"SUPABASE_AUTH_GOOGLE_SECRET"`

	// GitHub Secrets (No need to fill up for local development)
	DLQEndpointURL                string `envconfig:"DLQ_ENDPOINT_URL"`
	PubSubPushServiceAccountEmail string `envconfig:"PUBSUB_PUSH_SERVICE_ACCOUNT_EMAIL"`
	GCPProjectID                  string `envconfig:"GCP_PROJECT_ID"`
	GCPRegion                     string `envconfig:"GCP_REGION"`
	GCPServiceAccount             string `envconfig:"GCP_SERVICE_ACCOUNT"`
	GCPWorkloadIdentityProvider   string `envconfig:"GCP_WORKLOAD_IDENTITY_PROVIDER"`
	SupabaseDBPassword            string `envconfig:"SUPABASE_DB_PASSWORD"`
	SupabaseProjectID             string `envconfig:"SUPABASE_PROJECT_ID"`

	// GCP Setup Pub/Sub Script Secrets
	APIBaseURLStaging    string `envconfig:"API_BASE_URL_STAGING"`
	APIBaseURLProd       string `envconfig:"API_BASE_URL_PROD"`
	PythonBaseURLStaging string `envconfig:"PYTHON_BASE_URL_STAGING"`
	PythonBaseURLProd    string `envconfig:"PYTHON_BASE_URL_PROD"`
	GCPProjectIDLocal    string `envconfig:"GCP_PROJECT_ID_LOCAL"`
	GCPProjectIDStaging  string `envconfig:"GCP_PROJECT_ID_STAGING"`
	GCPProjectIDProd     string `envconfig:"GCP_PROJECT_ID_PROD"`
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
