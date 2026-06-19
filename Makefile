.PHONY: build test lint up down
# build compiles all packages.ForgeOps - Phase 0 Implementation (Foundations & Bootstrap)3
build:
	go build ./...
# test runs the unit test suite.
test:
	go test ./...
# lint runs the configured linters.
lint:
	golangci-lint run
# up creates the local cluster and runs the webhook server (wired in a later phase).
up:
	@echo "cluster bring-up is implemented in Phase 3"
# down tears the local cluster down (wired in a later phase).
down:
	@echo "cluster teardown is implemented in Phase 3"

fmt:
	golangci-lint fmt

check: fmt lint
