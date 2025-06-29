.PHONY: all build run swagger clean fmt build-worker worker-ingestion worker-embedding worker-explanation worker-summary

# Default target
all: build

# Build the application
build:
	go build -o bin/app ./cmd/app

# Run the application
run: build
	./bin/app

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
	rm -rf docs/swagger

# Build orchestrator
build-orchestrator:
	go build -o bin/orchestrator ./cmd/orchestrator

# Run the orchestrator for ingestion
run-orchestrator-ingestion: build-orchestrator
	./bin/orchestrator --mode ingestion

# Run the orchestrator for embedding
run-orchestrator-embedding: build-orchestrator
	./bin/orchestrator --mode embedding

# Run the orchestrator for explanation
run-orchestrator-explanation: build-orchestrator
	./bin/orchestrator --mode explanation

# Run the orchestrator for summary
run-orchestrator-summary: build-orchestrator
	./bin/orchestrator --mode summary