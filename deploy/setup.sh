#!/usr/bin/env bash
set -euo pipefail

# One-time GCP provisioning for GoNotify's production Cloud Run Job.
# Safe to re-run: skips/updates resources that already exist.
#
# Prerequisites:
#   - gcloud CLI installed and authenticated (`gcloud auth login`)
#   - secrets/application.yaml and secrets/service_account.json present locally (see CLAUDE.md)
#
# After this script, run deploy/deploy.sh to build the image, create the Cloud Run Job, and wire
# up the Cloud Scheduler trigger (that part needs the job to exist first, so it lives in deploy.sh).

PROJECT_ID="bolu-286408"
REGION="europe-west1"
REPO="gonotify"
RUNNER_SA="gonotify-runner"
SCHEDULER_SA="gonotify-scheduler-invoker"
SECRET_SA_JSON="gonotify-google-service-account"
SECRET_TWILIO_SID="gonotify-twilio-account-sid"
SECRET_TWILIO_TOKEN="gonotify-twilio-auth-token"

cd "$(dirname "$0")/.."

if [[ ! -f secrets/application.yaml || ! -f secrets/service_account.json ]]; then
  echo "secrets/application.yaml and secrets/service_account.json must exist locally (see CLAUDE.md)." >&2
  exit 1
fi

yaml_val() {
  sed -nE "s/^[[:space:]]*${1}:[[:space:]]*\"?([^\"]*)\"?[[:space:]]*\$/\1/p" secrets/application.yaml | tail -1
}

TWILIO_ACCOUNT_SID=$(yaml_val account_sid)
TWILIO_AUTH_TOKEN=$(yaml_val auth_token)

gcloud config set project "$PROJECT_ID" >/dev/null

echo "Enabling required APIs..."
gcloud services enable \
  run.googleapis.com \
  cloudscheduler.googleapis.com \
  secretmanager.googleapis.com \
  artifactregistry.googleapis.com

echo "Creating Artifact Registry repo..."
gcloud artifacts repositories describe "$REPO" --location="$REGION" >/dev/null 2>&1 || \
  gcloud artifacts repositories create "$REPO" \
    --repository-format=docker \
    --location="$REGION" \
    --description="GoNotify container images"

echo "Creating/updating secrets..."
if gcloud secrets describe "$SECRET_SA_JSON" >/dev/null 2>&1; then
  gcloud secrets versions add "$SECRET_SA_JSON" --data-file=secrets/service_account.json >/dev/null
else
  gcloud secrets create "$SECRET_SA_JSON" --data-file=secrets/service_account.json --replication-policy=automatic >/dev/null
fi

create_string_secret() {
  local name="$1" value="$2"
  if gcloud secrets describe "$name" >/dev/null 2>&1; then
    printf '%s' "$value" | gcloud secrets versions add "$name" --data-file=- >/dev/null
  else
    printf '%s' "$value" | gcloud secrets create "$name" --data-file=- --replication-policy=automatic >/dev/null
  fi
}
create_string_secret "$SECRET_TWILIO_SID" "$TWILIO_ACCOUNT_SID"
create_string_secret "$SECRET_TWILIO_TOKEN" "$TWILIO_AUTH_TOKEN"

echo "Creating runtime service account..."
gcloud iam service-accounts describe "${RUNNER_SA}@${PROJECT_ID}.iam.gserviceaccount.com" >/dev/null 2>&1 || \
  gcloud iam service-accounts create "$RUNNER_SA" --display-name="GoNotify Cloud Run Job runtime"

for secret in "$SECRET_SA_JSON" "$SECRET_TWILIO_SID" "$SECRET_TWILIO_TOKEN"; do
  gcloud secrets add-iam-policy-binding "$secret" \
    --member="serviceAccount:${RUNNER_SA}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor" >/dev/null
done

echo "Creating scheduler-invoker service account..."
gcloud iam service-accounts describe "${SCHEDULER_SA}@${PROJECT_ID}.iam.gserviceaccount.com" >/dev/null 2>&1 || \
  gcloud iam service-accounts create "$SCHEDULER_SA" --display-name="GoNotify Cloud Scheduler invoker"

echo
echo "Setup done. Next: run deploy/deploy.sh to build+push the image, create the Cloud Run Job,"
echo "and wire up the Cloud Scheduler trigger."