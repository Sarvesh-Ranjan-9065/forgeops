#!/usr/bin/env bash
# Runs the test suite for ForgeOps. Supports --update and --coverage flags.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
source "${SCRIPT_DIR}/lib/common.sh"

require_cmd "go"

UPDATE_GOLDEN=false
COVERAGE=false

for arg in "$@"; do
    case "$arg" in
        --update)
            UPDATE_GOLDEN=true
            ;;
        --coverage)
            COVERAGE=true
            ;;
        *)
            log_error "Unknown argument: $arg"
            log_info "Usage: bash scripts/test.sh [--update] [--coverage]"
            exit 1
            ;;
    esac
done

cd "${REPO_ROOT}"

if [ "${UPDATE_GOLDEN}" = true ]; then
    log_info "Regenerating scaffold golden files..."
    go test ./internal/scaffold/... -update
    log_success "Golden files updated."
    exit 0
fi

if [ "${COVERAGE}" = true ]; then
    log_info "Running Go test suite with coverage profile..."
    go test -v -coverprofile=coverage.out ./...
    log_success "Test coverage written to coverage.out"
else
    log_info "Running Go test suite..."
    go test ./...
    log_success "All unit and golden file tests passed."
fi
