# miniclue-be

A Go-based backend service for MiniClue, providing lecture processing, AI-powered explanations, and subscription management.

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

## Local Development

### 1. Start Local Services (Local Development Only)

This project uses Docker Compose to run the Google Cloud Pub/Sub emulator.

```bash
docker-compose up -d
```

### 2. Set Up Google Cloud Secret Manager (Local Development)

Secret Manager requires a real GCP project even for local development. You'll need to create a dedicated GCP project for local development.

**Step 1: Create a GCP Project for Local Development**

```bash
# Create a new GCP project (or use an existing one)
gcloud projects create miniclue-gcp-local-sr --name="MiniClue - Local SR"
gcloud config set project miniclue-gcp-local-sr
```

**Step 2: Enable Secret Manager API**

```bash
gcloud services enable secretmanager.googleapis.com --project=miniclue-gcp-local-sr
```

**Step 3: Set Up Authentication**

For local development, use Application Default Credentials (your personal account). This authenticates your local machine to access Secret Manager:

```bash
gcloud auth application-default login
```

**Important**: Grant your personal account the necessary permissions:

```bash
# Grant yourself Secret Manager Admin role for local development
gcloud projects add-iam-policy-binding miniclue-gcp-local-sr \
  --member="user:$(gcloud config get-value account)" \
  --role="roles/secretmanager.admin"
```

**Step 4: Set Environment Variables**

**Note**: Service accounts are only needed for production/staging environments where services run without user interaction. For local development, Application Default Credentials with your personal account is sufficient.

Add to your `.env` file:

```bash
# GCP Project IDs
GCP_PROJECT_ID_LOCAL=miniclue-gcp-local-sr
GCP_PROJECT_ID_STAGING=miniclue-gcp-stg
GCP_PROJECT_ID_PROD=miniclue-gcp-prod

# Environment
ENV=development
```

**Note**: For production/staging, the AI service (`miniclue-ai`) also needs access to Secret Manager. Ensure the Python service's service account has the `Secret Manager Secret Accessor` role in the same project. For local development, the AI service can also use Application Default Credentials.

### 3. Set Up Pub/Sub Environment

This step uses a Go program to configure Pub/Sub topics and subscriptions. It can target your local emulator, staging, or production.

**Important**: Before running for `staging` or `production`, you must authenticate with Google Cloud:

```bash
gcloud auth application-default login
```

The account you use must have the `Pub/Sub Editor` role on the target GCP project.

To run the setup:

```bash
# For local development (resets all topics/subscriptions)
make setup-pubsub-local

# For staging (creates or updates resources, does not delete)
make deploy-pubsub env=staging

# For production (creates or updates resources, does not delete)
make deploy-pubsub env=production
```

### 4. Run the API Server

To build and run the main API server:

```bash
make run
```

The API server will now be running and connected to the local Pub/Sub emulator.

### 5. Update swagger documentation

```bash
make swagger
```

This will generate the `swagger.json` file in the `docs` directory.

### 6. Update local Supabase database

- Make updates to the `supabase/schemas/schema.sql` file.
- Run `supabase db diff -f [filename]` to generate a migration file.
- Run `supabase migration up` to apply the migration to the local database.

## CI/CD Workflow

### Staging Environment

1. A developer writes code on a feature branch and opens a Pull Request to `main`.
2. After code review and approval, the PR is merged.
3. The merge to `main` automatically triggers a GitHub Actions workflow (`cd.yml`).
4. This workflow builds a Docker image tagged with the commit SHA and deploys it to the **staging** environment.

### Production Environment

1. After changes are verified in staging, a release can be deployed to production.
2. A developer creates and pushes a semantic version git tag (e.g., `v1.2.3`) from the `main` branch.
   ```bash
   # From the main branch
   git tag -a v1.0.0 -m "Release notes"
   git push origin v1.0.0
   ```
3. Pushing the tag automatically triggers the release workflow (`release.yml`).
4. This workflow builds a Docker image tagged with the version (e.g., `v1.0.0`) and deploys it to the **production** environment.
