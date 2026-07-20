#!/usr/bin/env bash
set -euo pipefail

# Build, push, and deploy the GoNotify Cloud Run Job, then (re)wire the Cloud Scheduler trigger.
# Run after any code change to ship it to production. Requires deploy/setup.sh to have been run
# at least once first.

PROJECT_ID="bolu-286408"
REGION="europe-west1"
REPO="gonotify"
JOB_NAME="gonotify-reminder"
RUNNER_SA="gonotify-runner@${PROJECT_ID}.iam.gserviceaccount.com"
SCHEDULER_SA="gonotify-scheduler-invoker@${PROJECT_ID}.iam.gserviceaccount.com"
SCHEDULER_JOB="gonotify-daily-trigger"
SECRET_SA_JSON="gonotify-google-service-account"
SECRET_TWILIO_SID="gonotify-twilio-account-sid"
SECRET_TWILIO_TOKEN="gonotify-twilio-auth-token"
IMAGE="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO}/gonotify:$(date +%Y%m%d%H%M%S)"

cd "$(dirname "$0")/.."

if [[ ! -f secrets/application.yaml ]]; then
  echo "secrets/application.yaml must exist locally (see CLAUDE.md)." >&2
  exit 1
fi

yaml_val() {
  sed -nE "s/^[[:space:]]*${1}:[[:space:]]*\"?([^\"]*)\"?[[:space:]]*\$/\1/p" secrets/application.yaml | tail -1
}

SHEET_ID=$(yaml_val sheet_id)
SHEET_RANGE=$(yaml_val sheet_range)
WHATSAPP_FROM=$(yaml_val whatsapp_from)

echo "Building and pushing ${IMAGE}..."
gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet
# --platform linux/amd64: Cloud Run requires a linux/amd64 image manifest; without this, building
# on Apple Silicon produces an arm64 image (or an attestation-bearing multi-arch index) that Cloud
# Run rejects. --provenance=false avoids buildx wrapping the image in an OCI index for attestations.
docker build --platform linux/amd64 --provenance=false -t "$IMAGE" .
docker push "$IMAGE"

echo "Deploying Cloud Run Job ${JOB_NAME}..."
gcloud run jobs deploy "$JOB_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --image="$IMAGE" \
  --service-account="$RUNNER_SA" \
  --max-retries=0 \
  --set-env-vars="GOOGLE_SHEET_ID=${SHEET_ID},GOOGLE_SHEET_RANGE=${SHEET_RANGE},TWILIO_WHATSAPP_FROM=${WHATSAPP_FROM},GOOGLE_APPLICATION_CREDENTIALS=/secrets/service-account.json" \
  --set-secrets="/secrets/service-account.json=${SECRET_SA_JSON}:latest,TWILIO_ACCOUNT_SID=${SECRET_TWILIO_SID}:latest,TWILIO_AUTH_TOKEN=${SECRET_TWILIO_TOKEN}:latest"

echo "Granting Cloud Scheduler invoker permission on the job..."
gcloud run jobs add-iam-policy-binding "$JOB_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --member="serviceAccount:${SCHEDULER_SA}" \
  --role="roles/run.invoker" >/dev/null

RUN_URI="https://${REGION}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${PROJECT_ID}/jobs/${JOB_NAME}:run"

echo "Creating/updating Cloud Scheduler trigger (${SCHEDULER_JOB}, 07:00 UTC daily)..."
if gcloud scheduler jobs describe "$SCHEDULER_JOB" --location="$REGION" >/dev/null 2>&1; then
  gcloud scheduler jobs update http "$SCHEDULER_JOB" \
    --location="$REGION" \
    --schedule="0 7 * * *" \
    --time-zone="UTC" \
    --uri="$RUN_URI" \
    --http-method=POST \
    --oauth-service-account-email="$SCHEDULER_SA"
else
  gcloud scheduler jobs create http "$SCHEDULER_JOB" \
    --location="$REGION" \
    --schedule="0 7 * * *" \
    --time-zone="UTC" \
    --uri="$RUN_URI" \
    --http-method=POST \
    --oauth-service-account-email="$SCHEDULER_SA"
fi

echo
echo "Done. Deployed image: ${IMAGE}"
echo "Trigger a run manually any time with:"
echo "  gcloud run jobs execute ${JOB_NAME} --region ${REGION} --project ${PROJECT_ID}"