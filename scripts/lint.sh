#!/usr/bin/env bash
# Runs golangci-lint and helm chart linter checks.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

require_cmd "golangci-lint"

cd "${REPO_ROOT}"

log_info "Running golangci-lint..."
golangci-lint run
log_success "Go linters passed cleanly."

if [ -d "out/demo/helm" ] && command -v helm >/dev/null 2>&1; then
    log_info "Linting generated demo Helm chart..."
    helm lint out/demo/helm
    log_success "Helm chart lint passed."
fi
