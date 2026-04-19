#!/bin/bash
# Deploy Rooted API to Cloud Run with ALL env vars.
# Usage: source .env.deploy && ./scripts/deploy-api.sh
#
# Requires .env.deploy file (gitignored) with:
#   R2_ACCOUNT_ID=xxx
#   R2_ACCESS_KEY_ID=xxx
#   R2_SECRET_ACCESS_KEY=xxx
set -e

IMAGE="us-central1-docker.pkg.dev/doodlegen-app-2026/rooted/rooted-api"
PROJECT="doodlegen-app-2026"
REGION="us-central1"
SERVICE="rooted-api"

# Check required vars
if [ -z "$R2_ACCOUNT_ID" ]; then
  echo "Error: R2_ACCOUNT_ID not set. Run: source .env.deploy"
  exit 1
fi

echo "=== Building ==="
docker build --platform linux/amd64 -t "$IMAGE" .

echo "=== Pushing ==="
docker push "$IMAGE"

echo "=== Deploying ==="
gcloud run deploy "$SERVICE" \
  --image "$IMAGE:latest" \
  --region "$REGION" \
  --project "$PROJECT" \
  --set-env-vars "ENV=production,ADMIN_TELEGRAM_IDS=6576003873,R2_ACCOUNT_ID=$R2_ACCOUNT_ID,R2_ACCESS_KEY_ID=$R2_ACCESS_KEY_ID,R2_SECRET_ACCESS_KEY=$R2_SECRET_ACCESS_KEY,R2_BUCKET_NAME=rooted-media,R2_PUBLIC_URL=https://rooted-media.r2.dev,TELEGRAM_WEBAPP_URL=https://rooted-miniapp.pages.dev" \
  --set-secrets "DATABASE_URL=rooted-database-url:latest,TELEGRAM_BOT_TOKEN=rooted-bot-token:latest,JWT_SECRET=rooted-jwt-secret:latest,REDIS_URL=rooted-redis-url:latest" \
  --add-cloudsql-instances "doodlegen-app-2026:us-central1:storyink-db"

echo "=== Done ==="
echo "API: https://rooted-api-643943133167.us-central1.run.app"
