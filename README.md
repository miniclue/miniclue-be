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
- [API Endpoints](#api-endpoints)
- [Database Migrations](#database-migrations)
- [Testing](#testing)
- [Contributing](#contributing)

## Features

- RESTful API built in Go (1.22+) using `net/http` and `ServeMux`
- Supabase PostgreSQL database with migration scripts
- User authentication and authorization middleware
- Orchestration pipelines for embeddings, summaries, and explanations
- Push-based Google Cloud Pub/Sub handlers for asynchronous processing

## Project Structure

```
miniclue-be/
├── cmd/
│   ├── app/          # Main API server entrypoint
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
- Docker and Docker Compose installed

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
# Supabase Local Development
DB_HOST=localhost
DB_PORT=54322
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=postgres
SUPABASE_LOCAL_JWT_SECRET=super-secret-jwt-token-with-at-least-32-characters-long
SUPABASE_LOCAL_S3_URL=http://localhost:54324
SUPABASE_LOCAL_S3_BUCKET=storage
SUPABASE_LOCAL_S3_REGION=us-east-1
SUPABASE_LOCAL_S3_ACCESS_KEY=owner
SUPABASE_LOCAL_S3_SECRET_KEY=owner

# Google Cloud Pub/Sub Emulator
GCP_PROJECT_ID=miniclue-dev
PUBSUB_INGESTION_TOPIC=ingestion
PUBSUB_EMULATOR_HOST=localhost:8085
```

## Running the Application

### 1. Start Local Services

This project uses Docker Compose to run the Google Cloud Pub/Sub emulator.

```bash
docker-compose up -d
```

### 2. Set Up Local Pub/Sub Environment

After starting the emulator for the first time (or to reset it), you must create the necessary topics and subscriptions. A helper command is provided for this.

```bash
make setup-pubsub
```

### 3. Run the API Server

To build and run the main API server:

```bash
make run
```

The API server will now be running and connected to the local Pub/Sub emulator.

## API Endpoints

Refer to `internal/api/v1/router/router.go` for detailed endpoint documentation.
You can also generate Swagger documentation by running `make swagger`.

## Full CI/CD Workflow

1. Developer writes code, tests locally, and commits to a feature branch.
2. Developer opens a PR from the feature branch to main.
3. Code is reviewed by a reviewer.
4. Once approved, PR is merged to main.
5. GitHub Actions workflow builds and deploys to staging.
6. Developer tests in staging.
7. If no issues are detected in staging, developer manually deploys to production using Github Actions.
