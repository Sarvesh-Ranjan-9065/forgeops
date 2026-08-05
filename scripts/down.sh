#!/usr/bin/env bash
# Tears down the local Kind cluster and removes local registry container.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

require_cmd "kind"
require_cmd "docker"

log_info "Tearing down ForgeOps local infrastructure..."

if kind get clusters 2>/dev/null | grep -q "^kind$"; then
    kind delete cluster
    log_success "Deleted Kind cluster."
else
    log_info "No active Kind cluster found."
fi

if docker ps -a --format '{{.Names}}' | grep -q "^kind-registry$"; then
    log_info "Stopping and removing kind-registry container..."
    docker stop kind-registry >/dev/null 2>&1 || true
    docker rm kind-registry >/dev/null 2>&1 || true
    log_success "Removed container registry."
fi

log_success "ForgeOps local infrastructure tear down complete."
