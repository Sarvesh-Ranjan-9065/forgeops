FROM golang:1.26-bookworm

# Build arguments for pinned toolchain versions
ARG KUBECTL_VERSION="v1.30.2"
ARG KIND_VERSION="v0.27.0"
ARG HELM_VERSION="v3.15.4"
ARG GOLANGCI_LINT_VERSION="v1.64.5"
ARG TRIVY_VERSION="0.59.1"

# Avoid interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive
ENV PATH="/usr/local/bin:/go/bin:${PATH}"

# Install core development utilities
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    git \
    make \
    bash \
    ca-certificates \
    tar \
    sudo \
    jq \
    gnupg \
    lsb-release \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Create non-root developer user (sarvesh) with passwordless sudo access
RUN groupadd --gid 1000 sarvesh \
    && useradd --uid 1000 --gid 1000 -m -s /bin/bash sarvesh \
    && echo "sarvesh ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/sarvesh \
    && chmod 0440 /etc/sudoers.d/sarvesh

# Install Docker CLI (for interacting with host Docker daemon)
RUN mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends docker-ce-cli \
    && rm -rf /var/lib/apt/lists/*

# Install pinned kubectl
RUN curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 -o /usr/local/bin/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
    && chmod +x /usr/local/bin/kubectl

# Install pinned Kind
RUN curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 -o /usr/local/bin/kind "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-linux-amd64" \
    && chmod +x /usr/local/bin/kind

# Install pinned Helm
RUN curl -fsSL --retry 5 --retry-all-errors --retry-delay 2 -o /tmp/helm.tar.gz "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" \
    && tar -xzf /tmp/helm.tar.gz -C /tmp \
    && mv /tmp/linux-amd64/helm /usr/local/bin/helm \
    && rm -rf /tmp/helm.tar.gz /tmp/linux-amd64

# Install pinned golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} \
    && mv /go/bin/golangci-lint /usr/local/bin/

# Install Trivy vulnerability scanner via official apt repository
RUN mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://aquasecurity.github.io/trivy-repo/deb/public.key | gpg --dearmor -o /etc/apt/keyrings/trivy.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" | tee /etc/apt/sources.list.d/trivy.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends trivy \
    && rm -rf /var/lib/apt/lists/*

# Prepare persistent volume directories and fix ownership for sarvesh user
RUN mkdir -p /go/pkg/mod /home/sarvesh/.cache/go-build /home/sarvesh/.commandhistory \
    && chown -R sarvesh:sarvesh /go /home/sarvesh

USER sarvesh
WORKDIR /workspace
