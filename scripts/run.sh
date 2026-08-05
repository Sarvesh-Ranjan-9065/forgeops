#!/usr/bin/env bash
# Starts the forge-webhook server locally after ensuring required environment variables exist.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

require_cmd "go"

cd "${REPO_ROOT}"

# Load .env file if present
if [ -f ".env" ]; then
    log_info "Loading environment variables from .env file..."
    set -a
    source .env
    set +a
fi

if [ -z "${FORGE_WEBHOOK_SECRET:-}" ]; then
    log_warn "FORGE_WEBHOOK_SECRET is not set in environment. Falling back to default 'devsecret'..."
    export FORGE_WEBHOOK_SECRET="devsecret"
fi

if [ -z "${FORGE_GITHUB_TOKEN:-}" ]; then
    log_warn "FORGE_GITHUB_TOKEN is not set in environment."
fi

log_info "Starting forge-webhook server on :8080..."
exec go run ./cmd/forge-webhook
