#!/bin/bash
# Deploy Rooted API to Cloud Run with ALL env vars.
# Usage: ./scripts/deploy-api.sh
set -e

IMAGE="us-central1-docker.pkg.dev/doodlegen-app-2026/rooted/rooted-api"
PROJECT="doodlegen-app-2026"
REGION="us-central1"
SERVICE="rooted-api"

echo "=== Building ==="
docker build --platform linux/amd64 -t "$IMAGE" .

echo "=== Pushing ==="
docker push "$IMAGE"

echo "=== Deploying ==="
gcloud run deploy "$SERVICE" \
  --image "$IMAGE:latest" \
  --region "$REGION" \
  --project "$PROJECT" \
  --set-env-vars "ENV=production,ADMIN_TELEGRAM_IDS=6576003873,R2_ACCOUNT_ID=87736e341d8552320ec9c5a30d71f463,R2_ACCESS_KEY_ID=91b8b75a8c2209043dbfc40c9473e646,R2_SECRET_ACCESS_KEY=c42ec5e08944be41930a55473fe45746852b629a411b52be0d5b9091b3e32403,R2_BUCKET_NAME=rooted-media,R2_PUBLIC_URL=https://rooted-media.r2.dev,TELEGRAM_WEBAPP_URL=https://rooted-miniapp.pages.dev" \
  --set-secrets "DATABASE_URL=rooted-database-url:latest,TELEGRAM_BOT_TOKEN=rooted-bot-token:latest,JWT_SECRET=rooted-jwt-secret:latest,REDIS_URL=rooted-redis-url:latest" \
  --add-cloudsql-instances "doodlegen-app-2026:us-central1:storyink-db"

echo "=== Done ==="
echo "API: https://rooted-api-643943133167.us-central1.run.app"
