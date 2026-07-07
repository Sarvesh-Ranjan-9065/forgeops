#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

# registry.sh starts a local container registry and connects it to the Kind
# network so cluster nodes can pull locally-built images. Preview hostnames use
# the pattern pr-<n>-<svc>.127-0-0-1.nip.io, which resolves to 127.0.0.1 with no
# extra DNS configuration.

reg_name="kind-registry"
reg_port="5000"

# Start the registry container if it is not already running.
if [ -z "$(docker ps -q --filter "name=^/${reg_name}$")" ]; then
  docker run -d --restart=always \
    -p "127.0.0.1:${reg_port}:5000" \
    --name "${reg_name}" \
    registry:2
fi

# Connect the registry to the Kind network (no-op if already connected).
docker network connect "kind" "${reg_name}" 2>/dev/null || true

# Advertise the registry to the cluster per KEP-1755.
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

echo "Local registry available at localhost:${reg_port}"
echo "Tag images as localhost:${reg_port}/<svc>:<tag> and docker push them."