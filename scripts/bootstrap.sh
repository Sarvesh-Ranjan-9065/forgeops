#!/usr/bin/env bash
# Validates developer environment and checks installed tools.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

log_info "Verifying ForgeOps development toolchain..."

tools=("go" "docker" "kubectl" "kind" "helm" "golangci-lint" "trivy")

for tool in "${tools[@]}"; do
    if command -v "$tool" >/dev/null 2>&1; then
        version=""
        case "$tool" in
            go)            version="$(go version)" ;;
            docker)        version="$(docker --version)" ;;
            kubectl)       version="$(kubectl version --client 2>/dev/null | head -n 1)" ;;
            kind)          version="$(kind --version)" ;;
            helm)          version="$(helm version --short 2>/dev/null || helm version)" ;;
            golangci-lint) version="$(golangci-lint --version)" ;;
            trivy)         version="$(trivy --version 2>/dev/null | head -n 1)" ;;
            *)             version="$($tool --version 2>/dev/null | head -n 1)" ;;
        esac
        log_success "  ${tool}: ${version}"
    else
        log_warn "  ${tool}: Not installed"
    fi
done

log_info "Toolchain check complete."
