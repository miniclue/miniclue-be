# MiniClue API Gateway (`miniclue-be`)

The Go-based backend service responsible for authentication, orchestration, and serving as the primary API gateway for the MiniClue platform.

**Role in Stack:**

- **API Gateway:** Entry point for frontend traffic, handling routing, validation, and CORS.
- **Authentication:** Validates Supabase JWTs and enforces user-based access control.
- **Orchestration:** Triggers AI pipelines (ingestion, embedding, image analysis) by publishing messages to Google Cloud Pub/Sub.
- **Management:** Handles CRUD operations for courses, lectures, and chat history via Supabase Postgres.

## 🛠 Prerequisites

- **Go 1.24+**
- **Docker & Docker Compose** (For local Pub/Sub emulator)
- **Supabase CLI** (For local database management)
- **Google Cloud SDK** (For local authentication and Secret Manager)

## 🚀 Quick Start

> See [CONTRIBUTING.md](https://github.com/miniclue/miniclue-info/blob/main/CONTRIBUTING.md) for full details on how to setup and contribute to the project.

1. **Fork & Clone**

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/your-username/miniclue-be.git
cd miniclue-be
git remote add upstream https://github.com/miniclue/miniclue-be.git
go mod download
```

2. **Environment Setup**
   Copy the example config:

```bash
cp .env.example .env
```

_Ensure you populate all fields as stated in the `.env.example` file. For local development, you will need a GCP project to use Secret Manager._

3. **Local Infrastructure**

```bash
# Start Supabase (Postgres & Storage)
supabase start

# Start Pub/Sub Emulator
docker-compose up -d

# Setup Topics and Subscriptions
make setup-pubsub-local
```

4. **Run Locally**

```bash
make run
# Service will run at http://127.0.0.1:8080
```

## 📝 Pull Request Process

1. Create a new branch for your feature or bugfix: `git checkout -b feature/my-cool-improvement`.
2. Ensure your code follows the coding standards and project architecture.
3. Update Swagger documentation if you changed API handlers: `make swagger`.
4. Push to your fork: `git push origin feature/my-cool-improvement`.
5. Submit a Pull Request from your fork to the original repository's `main` branch.
6. Once your PR is approved and merged into `main`, the CI/CD pipeline will automatically deploy it to the [staging environment](https://stg.api.miniclue.com) for verification.
7. Once a new release is created, the CI/CD pipeline will automatically deploy it to the [production environment](https://api.miniclue.com).

> Note: Merging of PR and creation of release will be done by repo maintainers.
