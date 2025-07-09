.PHONY: all build run swagger clean fmt build-worker worker-ingestion worker-embedding worker-explanation worker-summary

# Default target
all: build

# Build the application
build:
	go build -o bin/app ./cmd/app

# Run the application
run: build
	./bin/app

# Build the setup-pubsub command
build-setup-pubsub:
	go build -o bin/setup-pubsub ./cmd/setup-pubsub

# Run the setup-pubsub command
setup-pubsub: build-setup-pubsub
	./bin/setup-pubsub

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
	rm -f bin/setup-pubsub
	rm -rf docs/swagger
