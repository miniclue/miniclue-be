.PHONY: all build run swagger clean

# Default target
all: build

# Build the application
build:
	go build -o bin/miniclue-be ./cmd/app

# Run the application
run: build
	./bin/miniclue-be

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init --parseDependency --parseInternal --generalInfo cmd/app/main.go --output docs/swagger
	@echo "Swagger documentation generated in docs/swagger"

# Clean generated files
clean:
	rm -f bin/miniclue-be
	rm -rf docs/swagger