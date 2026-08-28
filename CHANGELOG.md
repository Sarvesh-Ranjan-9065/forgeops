# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/) and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-08-29

### Added
- `forge` CLI that scaffolds a service (Dockerfile, Helm chart, CI) and can create
  and push a GitHub repository.
- `forge-webhook` controller that provisions a per-PR preview environment on a
  local Kind cluster: namespace, image build/push, Helm deploy, readiness wait,
  and a sticky PR comment with the URL.
- Lifecycle management: teardown on PR close, startup/periodic reconciliation, and
  a 24h TTL garbage collector.
- Observability: Prometheus metrics (webhook events, active envs, provision
  duration, reconcile errors, GC deletions) and a local Prometheus/Grafana stack.
- Security: HMAC webhook verification and a least-privilege RBAC manifest.
- CI: lint, test, build, Trivy vulnerability scan, and a no-emoji source gate.