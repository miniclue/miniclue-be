#!/bin/bash

# This script configures Pub/Sub topics and subscriptions using the gcloud CLI.
# It is designed to be environment-aware, targeting staging or production
# GCP projects.

set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration ---
TOPICS=("ingestion" "embedding" "explanation" "summary" "image-analysis")
SEVEN_DAYS_S="604800s" # 7 days in seconds for gcloud
THIRTY_ONE_DAYS="31d"
ACK_DEADLINE=180
MIN_RETRY_S="10s"
MAX_RETRY_S="600s"
MAX_DELIVERY_ATTEMPTS=5

# --- Function Definitions ---
function check_deps() {
  if ! command -v gcloud &>/dev/null; then
    echo "ERROR: gcloud command could not be found. Please install the Google Cloud SDK."
    exit 1
  fi
  if ! command -v jq &>/dev/null; then
    echo "ERROR: jq command could not be found. Please install jq (e.g., 'brew install jq')."
    exit 1
  fi
}

# --- Script Start ---
check_deps

ENV=${1}
if [ -z "$ENV" ]; then
  echo "ERROR: Environment not specified."
  echo "Usage: $0 <staging|production>"
  exit 1
fi

echo "Starting Pub/Sub setup for environment: $ENV"

# Load environment variables from .env file
if [ -f .env ]; then
  set -o allexport
  source .env
  set +o allexport
else
  echo "Warning: .env file not found. Relying on exported environment variables."
fi

# --- Environment-Specific Setup ---
PROJECT_FLAG=""
GCLOUD_OPTS=""
API_BASE_URL=""

case "$ENV" in
"staging")
  if [ -z "$GCP_PROJECT_ID_STAGING" ] || [ -z "$API_BASE_URL_STAGING" ]; then
    echo "ERROR: GCP_PROJECT_ID_STAGING and API_BASE_URL_STAGING must be set."
    exit 1
  fi
  PROJECT_FLAG="--project=$GCP_PROJECT_ID_STAGING"
  API_BASE_URL=$API_BASE_URL_STAGING
  PYTHON_API_URL="$API_BASE_URL/v1/pubsub"
  GATEWAY_API_URL="$API_BASE_URL/v1/pubsub/dlq"
  ;;
"production")
  if [ -z "$GCP_PROJECT_ID_PROD" ] || [ -z "$API_BASE_URL_PROD" ]; then
    echo "ERROR: GCP_PROJECT_ID_PROD and API_BASE_URL_PROD must be set."
    exit 1
  fi
  PROJECT_FLAG="--project=$GCP_PROJECT_ID_PROD"
  API_BASE_URL=$API_BASE_URL_PROD
  PYTHON_API_URL="$API_BASE_URL/v1/pubsub"
  GATEWAY_API_URL="$API_BASE_URL/v1/pubsub/dlq"
  ;;
*)
  echo "ERROR: Invalid environment '$ENV'. Must be one of: staging, production."
  exit 1
  ;;
esac

# --- Resource Creation/Update Loop ---
for TOPIC_NAME in "${TOPICS[@]}"; do
  # Generate names
  SUB_SUFFIX=""
  [ "$ENV" == "staging" ] && SUB_SUFFIX="-stg"
  [ "$ENV" == "production" ] && SUB_SUFFIX="-prod"
  DLQ_TOPIC="${TOPIC_NAME}-dlq"
  MAIN_TOPIC="${TOPIC_NAME}"
  MAIN_SUB="${TOPIC_NAME}-sub${SUB_SUFFIX}"
  DLQ_SUB="${TOPIC_NAME}-dlq-sub${SUB_SUFFIX}"
  PUSH_ENDPOINT="${PYTHON_API_URL}/${TOPIC_NAME}"

  echo -e "\n--- Ensuring resources for topic: $MAIN_TOPIC (Env: $ENV) ---"

  # Create/Update Topics
  gcloud pubsub topics describe "$MAIN_TOPIC" $PROJECT_FLAG &>/dev/null || gcloud pubsub topics create "$MAIN_TOPIC" $PROJECT_FLAG
  gcloud pubsub topics update "$MAIN_TOPIC" $PROJECT_FLAG --message-retention-duration="$SEVEN_DAYS_S"

  gcloud pubsub topics describe "$DLQ_TOPIC" $PROJECT_FLAG &>/dev/null || gcloud pubsub topics create "$DLQ_TOPIC" $PROJECT_FLAG
  gcloud pubsub topics update "$DLQ_TOPIC" $PROJECT_FLAG --message-retention-duration="$SEVEN_DAYS_S"

  # Create/Update Main Subscription
  if gcloud pubsub subscriptions describe "$MAIN_SUB" $PROJECT_FLAG &>/dev/null; then
    echo "Updating existing subscription: $MAIN_SUB"
    gcloud pubsub subscriptions update "$MAIN_SUB" $PROJECT_FLAG \
      --push-endpoint="$PUSH_ENDPOINT" --ack-deadline="$ACK_DEADLINE"
  else
    echo "Creating new subscription: $MAIN_SUB"
    gcloud pubsub subscriptions create "$MAIN_SUB" $PROJECT_FLAG \
      --topic="$MAIN_TOPIC" --push-endpoint="$PUSH_ENDPOINT" \
      --ack-deadline="$ACK_DEADLINE" --expiration-period="$THIRTY_ONE_DAYS" \
      --dead-letter-topic="$DLQ_TOPIC" --max-delivery-attempts="$MAX_DELIVERY_ATTEMPTS" \
      --min-retry-delay="$MIN_RETRY_S" --max-retry-delay="$MAX_RETRY_S"
  fi

  # Create/Update DLQ Subscription
  if gcloud pubsub subscriptions describe "$DLQ_SUB" $PROJECT_FLAG &>/dev/null; then
    echo "Updating existing subscription: $DLQ_SUB"
    gcloud pubsub subscriptions update "$DLQ_SUB" $PROJECT_FLAG \
      --push-endpoint="$GATEWAY_API_URL" --ack-deadline="$ACK_DEADLINE" \
      --min-retry-delay="$MIN_RETRY_S" --max-retry-delay="$MAX_RETRY_S"
  else
    echo "Creating new subscription: $DLQ_SUB"
    gcloud pubsub subscriptions create "$DLQ_SUB" $PROJECT_FLAG \
      --topic="$DLQ_TOPIC" --push-endpoint="$GATEWAY_API_URL" \
      --ack-deadline="$ACK_DEADLINE" --expiration-period="$THIRTY_ONE_DAYS" \
      --min-retry-delay="$MIN_RETRY_S" --max-retry-delay="$MAX_RETRY_S"
  fi
done

echo -e "\nPub/Sub setup for $ENV environment complete." 