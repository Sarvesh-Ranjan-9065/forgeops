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
 # up creates the local cluster and runs the webhook server.
 up:
	kind create cluster --config deploy/kind/cluster.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	bash deploy/kind/registry.sh
 
 # down tears the local cluster down.
 down:
	kind delete cluster

fmt:
	golangci-lint fmt

check: fmt lint

start:
	docker start kind-control-plane kind-registry

stop:
	docker stop kind-control-plane kind-registry

