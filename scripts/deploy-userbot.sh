#!/bin/bash
# Deploy Rooted Userbot to Cloud Run.
# Usage: ./scripts/deploy-userbot.sh
set -e

IMAGE="us-central1-docker.pkg.dev/doodlegen-app-2026/rooted/rooted-userbot"
PROJECT="doodlegen-app-2026"
REGION="us-central1"
SERVICE="rooted-userbot"
BUCKET="rooted-userbot-session"

# Create GCS bucket for session persistence (idempotent, versioned)
if ! gsutil ls "gs://$BUCKET" 2>/dev/null; then
  gsutil mb -l "$REGION" -p "$PROJECT" "gs://$BUCKET"
  gsutil versioning set on "gs://$BUCKET"
  echo "Created versioned bucket gs://$BUCKET"
fi

# Upload local session file if it exists and bucket is empty
if [ -f "userbot/rooted_userbot.session" ]; then
  gsutil cp "userbot/rooted_userbot.session" "gs://$BUCKET/userbot/rooted_userbot.session"
  echo "Session file uploaded to GCS"
fi

echo "=== Building ==="
docker build --platform linux/amd64 -t "$IMAGE" -f userbot/Dockerfile userbot/

echo "=== Pushing ==="
docker push "$IMAGE"

echo "=== Deploying ==="
gcloud run deploy "$SERVICE" \
  --image "$IMAGE:latest" \
  --region "$REGION" \
  --project "$PROJECT" \
  --port 8090 \
  --min-instances 1 \
  --max-instances 1 \
  --cpu 1 \
  --memory 256Mi \
  --set-env-vars "API_ID=31814871,API_HASH=aafb049d9b3330cce4627a38d61039e5,PHONE_NUMBER=+2348080578741,GCS_SESSION_BUCKET=$BUCKET,BACKEND_URL=https://rooted-api-643943133167.us-central1.run.app" \
  --set-secrets "DATABASE_URL=rooted-database-url:latest" \
  --add-cloudsql-instances "doodlegen-app-2026:us-central1:storyink-db" \
  --no-allow-unauthenticated

echo "=== Getting service URL ==="
USERBOT_URL=$(gcloud run services describe "$SERVICE" --region "$REGION" --project "$PROJECT" --format 'value(status.url)')
echo "Userbot: $USERBOT_URL"
echo ""
echo "Now update the API with USERBOT_URL:"
echo "  gcloud run services update rooted-api --region $REGION --project $PROJECT --set-env-vars USERBOT_URL=$USERBOT_URL"
