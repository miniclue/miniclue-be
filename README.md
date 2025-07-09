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
2.  Export the required environment variables. See `.env.example` for reference.

## Running the Application

### 1. Start Local Services (Local Development Only)

This project uses Docker Compose to run the Google Cloud Pub/Sub emulator.

```bash
docker-compose up -d
```

### 2. Set Up Pub/Sub Environment

This step uses a Go program to configure Pub/Sub topics and subscriptions. It can target your local emulator, staging, or production.

**Important**: Before running for `staging` or `production`, you must authenticate with Google Cloud:

```bash
gcloud auth application-default login
```

The account you use must have the `Pub/Sub Editor` role on the target GCP project.

To run the setup:

```bash
# For local development (resets all topics/subscriptions)
make setup-pubsub env=local

# For staging (creates or updates resources, does not delete)
make setup-pubsub env=staging

# For production (creates or updates resources, does not delete)
make setup-pubsub env=production
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
