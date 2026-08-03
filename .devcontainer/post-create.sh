#!/usr/bin/env bash
set -e

echo "=================================================="
echo "      Setting up ForgeOps Dev Workspace          "
echo "=================================================="

# Dynamically align docker group GID with host docker.sock
DOCKER_SOCK="/var/run/docker.sock"
if [ -S "$DOCKER_SOCK" ]; then
    DOCKER_GID=$(stat -c '%g' "$DOCKER_SOCK")
    if ! getent group "$DOCKER_GID" >/dev/null 2>&1; then
        sudo groupadd -g "$DOCKER_GID" docker-host
    fi
    GROUP_NAME=$(getent group "$DOCKER_GID" | cut -d: -f1)
    sudo usermod -aG "$GROUP_NAME" sarvesh
fi

# Ensure workspace volume mount permissions are correct
sudo chown -R sarvesh:sarvesh /go /home/sarvesh/.cache /home/sarvesh/.commandhistory 2>/dev/null || true

# Setup persistent shell history
if ! grep -q 'HISTFILE=/home/sarvesh/.commandhistory/.bash_history' ~/.bashrc 2>/dev/null; then
    mkdir -p /home/sarvesh/.commandhistory
    touch /home/sarvesh/.commandhistory/.bash_history
    echo 'export HISTFILE=/home/sarvesh/.commandhistory/.bash_history' >> ~/.bashrc
    echo 'export PROMPT_COMMAND="history -a; $PROMPT_COMMAND"' >> ~/.bashrc
fi

# Add GOPATH bin to PATH in bashrc if not present
if ! grep -q 'export PATH=$PATH:$(go env GOPATH)/bin' ~/.bashrc 2>/dev/null; then
    echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
fi

# Ensure deploy scripts are executable
if [ -f "deploy/kind/registry.sh" ]; then
    chmod +x deploy/kind/registry.sh
fi

# Download Go module dependencies
echo "==> Downloading Go dependencies..."
go mod download || true

# Welcome banner in bashrc
if ! grep -q 'ForgeOps Container Environment' ~/.bashrc 2>/dev/null; then
    cat << 'EOF' >> ~/.bashrc

echo ""
echo "--------------------------------------------------"
echo " Welcome to ForgeOps Containerized Workspace!     "
echo " All development operations run isolated here.    "
echo "--------------------------------------------------"
echo ""
EOF
fi

echo "==> Checking installed tool versions:"
echo "  - Go:            $(go version 2>/dev/null || echo 'Not found')"
echo "  - Docker CLI:    $(docker --version 2>/dev/null || echo 'Not found')"
echo "  - kubectl:       $(kubectl version --client --short 2>/dev/null || kubectl version --client 2>/dev/null | head -n 1 || echo 'Not found')"
echo "  - Kind:          $(kind --version 2>/dev/null || echo 'Not found')"
echo "  - Helm:          $(helm version --short 2>/dev/null || echo 'Not found')"
echo "  - golangci-lint: $(golangci-lint --version 2>/dev/null || echo 'Not found')"
echo "  - Trivy:         $(trivy --version 2>/dev/null | head -n 1 || echo 'Not found')"

echo "=================================================="
echo " Workspace Setup Complete!                        "
echo "=================================================="
