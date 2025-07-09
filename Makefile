.PHONY: all build run swagger clean fmt build-worker worker-ingestion worker-embedding worker-explanation worker-summary

# Default target
all: build

# Build the application
build:
	go build -o bin/app ./cmd/app

# Run the application
run: build
	./bin/app

# Build the setup-pubsub command for the local environment
build-setup-pubsub-local:
	go build -o bin/setup-pubsub-local ./cmd/setup-pubsub-local

# Run the setup for the local Pub/Sub emulator.
# This will delete all existing topics and subscriptions and create new ones.
.PHONY: setup-pubsub-local
setup-pubsub-local: build-setup-pubsub-local
	./bin/setup-pubsub-local

# Deploy Pub/Sub resources to staging or production.
# Usage: make deploy-pubsub env=staging
#        make deploy-pubsub env=production
.PHONY: deploy-pubsub
deploy-pubsub:
	./scripts/setup_pubsub.sh $(env)

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init --parseDependency --parseInternal --generalInfo cmd/app/main.go --output docs/swagger
	@echo "Swagger documentation generated in docs/swagger"

# Clean generated files
clean: 
	rm -f bin/app
	rm -f bin/setup-pubsub-local
	rm -rf docs/swagger
