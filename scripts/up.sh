#!/usr/bin/env bash
# Creates the local Kind cluster, installs ingress-nginx, and bootstraps the local registry.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

require_cmd "kind"
require_cmd "kubectl"
require_cmd "docker"

log_info "Starting ForgeOps local infrastructure..."

CLUSTER_CONFIG="${REPO_ROOT}/deploy/kind/cluster.yaml"
REGISTRY_SCRIPT="${REPO_ROOT}/deploy/kind/registry.sh"

if kind get clusters 2>/dev/null | grep -q "^kind$"; then
    log_info "Kind cluster 'kind' is already running."
else
    log_info "Creating Kind cluster using ${CLUSTER_CONFIG}..."
    kind create cluster --config "${CLUSTER_CONFIG}"
    log_success "Kind cluster created."
fi

log_info "Deploying ingress-nginx controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

log_info "Bootstrapping local container registry..."
if [ -f "${REGISTRY_SCRIPT}" ]; then
    bash "${REGISTRY_SCRIPT}"
else
    log_warn "Registry script not found at ${REGISTRY_SCRIPT}"
fi

log_success "ForgeOps local infrastructure is UP and READY."
