#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# demo.sh drives the end-to-end ForgeOps flow for a recorded demo. Record it with:
#   asciinema rec forgeops.cast -c "bash scripts/demo.sh"
#   agg forgeops.cast docs/demo.gif
# It assumes FORGE_GITHUB_TOKEN, FORGE_GITHUB_OWNER, and FORGE_WEBHOOK_SECRET are
# exported and that a tunnel is already forwarding GitHub webhooks to :8080.

svc="demo-svc"

echo "== bringing up the local platform =="
make up

echo "== scaffolding and pushing ${svc} =="
go run ./cmd/forge new --name "${svc}" --push

echo "== open a PR on ${FORGE_GITHUB_OWNER}/${svc}, then watch the preview appear =="
kubectl get ns -l forgeops.io/pr -w &
watch_pid=$!
sleep 60
kill "${watch_pid}" 2>/dev/null || true

echo "== preview URL =="
echo "http://pr-1-${svc}.127-0-0-1.nip.io"