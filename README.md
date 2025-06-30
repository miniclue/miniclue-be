# miniclue-be

A backend service for the miniclue application, providing APIs for managing courses and lectures using Go, Supabase, and AI-driven processing pipelines.

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [Running the Application](#running-the-application)
  - [API Server](#api-server)
  - [Orchestrator](#orchestrator)
- [API Endpoints](#api-endpoints)
- [Database Migrations](#database-migrations)
- [Testing](#testing)
- [Contributing](#contributing)

## Features

- RESTful API built in Go (1.22+) using `net/http` and `ServeMux`
- Supabase PostgreSQL database with migration scripts
- User authentication and authorization middleware
- Orchestration pipelines for embeddings, summaries, and explanations
- Background task queue with PG message queue client
- Structured logging using standard library

## Project Structure

```
miniclue-be/
├── cmd/
│   ├── app/          # Main API server entrypoint
│   └── orchestrator/ # Background orchestrator service
├── internal/         # Application code
│   ├── api/v1/       # DTOs, handlers, router
│   ├── config/       # Configuration loader
│   ├── middleware/   # Logging and auth middleware
│   ├── model/        # Database models
│   ├── repository/   # Database access layer
│   ├── service/      # Business logic
│   └── orchestrator/ # AI pipelines
├── supabase/         # Supabase config and migrations
├── go.mod            # Module definition
├── go.sum            # Dependency checksums
├── README.md         # Project overview
└── PLAN.md           # Development plan
```

## Getting Started

### Prerequisites

- Go 1.22+ installed

### Installation

```bash
git clone https://github.com/your-username/miniclue-be.git
cd miniclue-be
go mod download
```

### Configuration

1.  Set up your Supabase project locally or in the cloud.
2.  Export the required environment variables. You can create a `.env` file and source it.

```bash
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="your-db-password"
export DB_NAME="postgres"
export SUPABASE_LOCAL_JWT_SECRET="your-super-secret-jwt-token"
export SUPABASE_LOCAL_S3_URL="http://localhost:9000"
export SUPABASE_LOCAL_S3_BUCKET="your-s3-bucket"
export SUPABASE_LOCAL_S3_REGION="us-east-1"
export SUPABASE_LOCAL_S3_ACCESS_KEY="your-s3-access-key"
export SUPABASE_LOCAL_S3_SECRET_KEY="your-s3-secret-key"
export PYTHON_SERVICE_BASE_URL="http://localhost:8000"
```

## Running the Application

### API Server

To build and run the main API server:

```bash
make run
```

### Orchestrator

To run the background orchestrator for a specific task:

```bash
# For ingestion
make run-orchestrator-ingestion

# For embedding
make run-orchestrator-embedding

# For explanation
make run-orchestrator-explanation

# For summary
make run-orchestrator-summary
```

## Makefile Commands

This project uses a `Makefile` to streamline common development tasks.

- `make build`: Builds the API server binary.
- `make run`: Builds and runs the API server.
- `make build-orchestrator`: Builds the orchestrator binary.
- `make run-orchestrator-ingestion`: Runs the orchestrator in ingestion mode.
- `make run-orchestrator-embedding`: Runs the orchestrator in embedding mode.
- `make run-orchestrator-explanation`: Runs the orchestrator in explanation mode.
- `make run-orchestrator-summary`: Runs the orchestrator in summary mode.
- `make fmt`: Formats all Go source code.
- `make swagger`: Generates Swagger API documentation.
- `make clean`: Removes generated binaries and documentation.
- `go test ./...`: Run unit and integration tests (no make command).

## API Endpoints

Refer to `internal/api/v1/router/router.go` for detailed endpoint documentation.
You can also generate Swagger documentation by running `make swagger`.
