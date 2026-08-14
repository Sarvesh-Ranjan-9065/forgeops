.PHONY: build test lint check fmt up down run bootstrap help

# build compiles all packages.
build:
	go build ./...

# test runs the unit test suite.
test:
	bash scripts/test.sh

# lint runs the configured linters.
lint:
	bash scripts/lint.sh

# fmt formats Go source files.
fmt:
	go fmt ./...

# check runs formatters and linters.
check: fmt lint

# up creates the local cluster and boots infrastructure.
up:
	bash scripts/up.sh

# down tears down the local cluster and infrastructure.
down:
	bash scripts/down.sh

# run starts the webhook server locally.
run:
	bash scripts/run.sh

# bootstrap verifies the installed development toolchain.
bootstrap:
	bash scripts/bootstrap.sh



# check-emoji fails if any Go source contains non-ASCII characters.
check-emoji:
	bash scripts/check-no-emoji.sh


# help prints available make targets.
help:
	@echo "Available targets:"
	@echo "  make build      - Compile all Go packages"
	@echo "  make test       - Run unit and golden file tests"
	@echo "  make lint       - Run golangci-lint code checks"
	@echo "  make fmt        - Format code using golangci-lint fmt"
	@echo "  make check      - Run formatting and linting"
	@echo "  make up         - Spin up Kind cluster, ingress, and local registry"
	@echo "  make down       - Tear down Kind cluster and local registry"
	@echo "  make run        - Run forge-webhook server locally"
	@echo "  make bootstrap  - Check installed toolchain versions"
	@echo "  make check-emoji - Check for non-ASCII characters in Go source files"