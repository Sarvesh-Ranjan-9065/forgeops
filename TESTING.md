# ForgeOps Testing & Execution Guide

This document serves as your complete manual for building, testing, linting, and running ForgeOps end-to-end. It covers how the scripts work, what you should see, and how to verify that your local Kubernetes control plane is working perfectly.

---

## 1. Local Development Checks (The Basics)

Before you push any code or attempt an end-to-end test, always ensure the project is healthy. 

### What to run:
```bash
make check
make test
```

### What these do:
* **`make check`**: Runs `go fmt ./...` to format your Go code, and then triggers `scripts/lint.sh` (which runs `golangci-lint` and lints your generated Helm charts).
* **`make test`**: Runs the Go unit test suite (`bash scripts/test.sh`).

### Where to look / What to expect:
* You should see output indicating that all linters passed cleanly.
* The test suite will output standard Go test results (e.g., `ok github.com/sarvesh-ranjan-9065/forgeops/internal/...`).
* **Note on Golden Files:** If you intentionally change the scaffold templates, some tests may fail because the output no longer matches the expected "golden" files. If this happens, run `go test ./internal/scaffold/... -update` to regenerate them.

---

## 2. Infrastructure Setup (Booting the Cluster)

ForgeOps runs its ephemeral environments on a local Kubernetes (`Kind`) cluster combined with a local Docker registry and an NGINX ingress.

### When to run:
Run this once when you start working for the day, or when you need a fresh cluster.

### What to run:
```bash
make up
```

### What this does:
This script (`scripts/up.sh`) does the heavy lifting:
1. Creates a Kind cluster named `forgeops-cluster`.
2. Starts a local Docker registry (`registry:2`) on port `5000`.
3. Connects the registry to the Kind cluster's network.
4. Installs the NGINX Ingress Controller.

### Where to look / What to expect:
* Run `docker ps` to verify that `kind-control-plane` and `registry` are running.
* Run `kubectl get pods -n ingress-nginx` and wait until the ingress controller is marked `Running` and `Ready`.

---

## 3. The End-to-End Live Workflow

This is how you verify the entire PR lifecycle (Phase 4 & 5) works in real-time.

### Step 3a: Start the Webhook Server
In a dedicated terminal tab, start the control plane.

**What to run:**
```bash
make run
```
**What to expect:** The server will boot on port `8080`, and you'll see startup logs in JSON format from `slog`. Keep this terminal visible.

### Step 3b: Scaffold a Service & Trigger a PR (Provisioning)

**CRITICAL RULE:** Do NOT open pull requests against the `ForgeOps` repository itself to test the webhook. ForgeOps is a control plane meant to manage *other* repositories. Testing against the `ForgeOps` repository pollutes the control plane's source code with preview environments that don't belong to it.

Instead, generate a fresh target repository and test against that:
1. **Scaffold a target repo:** Run `go run ./cmd/forge new --name demo-svc --push` (Ensure `FORGE_GITHUB_TOKEN` and `FORGE_GITHUB_OWNER` are set in your environment). This creates a new repository on your GitHub account.
2. **Configure the Webhook via Cloudflare Tunnel:**
   * Install `cloudflared` (e.g., `sudo pacman -S cloudflared` on Arch).
   * Run the tunnel in a new terminal: `cloudflared tunnel --url http://localhost:8080`.
   * In GitHub, go to your new `demo-svc` repository settings -> Webhooks -> Add webhook.
   * **Payload URL:** Paste the public URL provided by Cloudflare and **CRITICALLY add `/webhook`** at the end (e.g., `https://random-words.trycloudflare.com/webhook`).
   * **Content type:** Set to `application/json`.
   * **Secret:** Set to the exact same value as `FORGE_WEBHOOK_SECRET` in your `.env` file.
   * **Events:** Check "Let me select individual events" and select **Pull requests**.
3. **Trigger:** Open a Pull Request against the new `demo-svc` repository.

**Where to look:**
1. **Terminal:** Watch your `make run` logs. You should see it receive an `opened` event, enqueue the PR, build the image, push to the local registry, and deploy via Helm.
2. **Kubernetes:** Open another terminal and run `kubectl get ns -l forgeops.io/pr`. You should see a new namespace named something like `pr-1-demo-svc`.
3. **Pods:** Run `kubectl get pods -n pr-1-demo-svc`. Ensure the pod is running.
4. **GitHub:** Check the PR comments. You should see a new sticky comment from your bot containing a link like `http://pr-1-demo-svc.127-0-0-1.nip.io`.
5. **Browser:** Click that link! It routes natively to your local cluster via `nip.io` and the NGINX ingress. You should see your scaffolded Go service live.

### Step 3c: Merge or Close the PR (Teardown)
Go back to GitHub and **Close** or **Merge** the PR.

**Where to look:**
1. **Terminal:** The webhook server logs will show a `closed` event, triggering the teardown sequence.
2. **Kubernetes:** Run `kubectl get ns -l forgeops.io/pr` again. The namespace should vanish (or be marked `Terminating`). The Helm release is cleanly uninstalled.

---

## 4. Advanced: Testing the Reconciler & Garbage Collection

ForgeOps contains background loops that heal the cluster (Phase 5). Here is how you test them.

### Test Drift Reconciliation
1. Open a PR so a namespace is created.
2. Kill the webhook server (`Ctrl+C` on `make run`).
3. Delete the namespace manually: `kubectl delete ns pr-<n>-<svc>`.
4. Start the webhook server again (`make run`).
5. **Where to look:** Watch the logs. Within 5 minutes, the startup reconciler will query GitHub, notice the PR is open but the namespace is missing, and automatically recreate the environment!

### Test TTL Garbage Collection
The GC collector looks for namespaces older than 24 hours. To test this without waiting a day:
1. Open a PR so a namespace is created.
2. Manually edit the creation timestamp label on the namespace to pretend it was created 2 days ago:
   `kubectl label ns pr-<n>-<svc> forgeops.io/created-at=<old-unix-timestamp> --overwrite`
3. **Where to look:** Watch the `make run` logs. The GC ticker (which runs every 10 minutes) will eventually wake up, spot the expired namespace, and delete it. 

---

## 5. Teardown (Cleaning Up)

When you are done for the day and want to reclaim resources on your laptop.

### What to run:
```bash
make down
```

### What this does:
This script (`scripts/down.sh`) completely destroys the Kind cluster and stops the local registry container.

### Where to look:
* `kind get clusters` should return empty.
* `docker ps` will show the registry container is gone.
