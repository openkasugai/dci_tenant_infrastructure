#!/bin/sh
# update-deps.sh
#
# Bulk-update direct dependencies listed in go.mod.
#
# Overview:
#   - k8s.io/*, sigs.k8s.io/cluster-api, and sigs.k8s.io/controller-runtime are
#     pinned to specific versions for compatibility and excluded from go get -u.
#   - All other direct dependencies are upgraded with go get -u.
#   - After the bulk upgrade, pinned versions are re-applied and go mod tidy is run.
#
# Prerequisites:
#   Modules referenced via local replace directives (e.g. ../cluster-api-provider-dci,
#   ../resource-pool-manager) must exist on disk before running this script.
#
# Usage:
#   ./update-deps.sh
#
# Module versions and excluded modules can be customized via environment variables:
#   K8S_MODULE_VERSION=v0.36.3 \
#   CONTROLLER_RUNTIME_VERSION=v0.24.1 \
#   CLUSTER_API_VERSION=v1.14.0 \
#   EXCLUDE_MODULES="gorm.io/gorm gorm.io/gen" \
#   ./update-deps.sh

set -eu

# Module versions (can be overridden by environment variables)
K8S_MODULE_VERSION="${K8S_MODULE_VERSION:-v0.36.3}"
CONTROLLER_RUNTIME_VERSION="${CONTROLLER_RUNTIME_VERSION:-v0.24.1}"
CLUSTER_API_VERSION="${CLUSTER_API_VERSION:-v1.14.0}"

# Additional modules to exclude from update (space-separated).
# Can be overridden by environment variable, e.g.:
#   EXCLUDE_MODULES="gorm.io/gen gorm.io/gorm my.org/foo" ./update-deps.sh
EXCLUDE_MODULES="${EXCLUDE_MODULES:-cluster-api-provider-dci resource-pool-manager}"

cd "$(dirname "$0")"

# Build exclude pattern: fixed k8s-related modules + user-defined ones
_FIXED_EXCLUDE='k8s\.io/(api|apiextensions-apiserver|apimachinery|cli-runtime|client-go)|sigs\.k8s\.io/cluster-api|sigs\.k8s\.io/controller-runtime'
_EXTRA_EXCLUDE=""
for _mod in ${EXCLUDE_MODULES}; do
    _escaped=$(printf '%s' "${_mod}" | sed 's/\./\\./g')
    _EXTRA_EXCLUDE="${_EXTRA_EXCLUDE}|${_escaped}"
done
_EXCLUDE_PATTERN="${_FIXED_EXCLUDE}${_EXTRA_EXCLUDE}"

echo "=== Updating direct dependencies (excluding k8s-related and: ${EXCLUDE_MODULES}) ==="
grep -P '^\t\S' go.mod \
    | grep -v '// indirect' \
    | grep -Ev "${_EXCLUDE_PATTERN}" \
    | awk '{print $1}' \
    | xargs go get -u

# --- Pin k8s.io/* (only modules present as direct deps in go.mod) ---
echo "=== Pinning k8s.io/* to ${K8S_MODULE_VERSION} ==="
_k8s_to_pin=$(awk -v v="${K8S_MODULE_VERSION}" \
    '$1 ~ /^k8s\.io\/(api|apiextensions-apiserver|apimachinery|cli-runtime|client-go)$/ && !/indirect/ {printf "%s@%s\n", $1, v}' \
    go.mod)
if [ -n "${_k8s_to_pin}" ]; then
    go get ${_k8s_to_pin}
fi

# --- Pin sigs.k8s.io/cluster-api* (all matching direct deps, e.g. /test) ---
echo "=== Pinning sigs.k8s.io/cluster-api* to ${CLUSTER_API_VERSION} ==="
_capi_to_pin=$(awk -v v="${CLUSTER_API_VERSION}" \
    '$1 ~ /^sigs\.k8s\.io\/cluster-api/ && !/indirect/ {printf "%s@%s\n", $1, v}' \
    go.mod)
if [ -n "${_capi_to_pin}" ]; then
    go get ${_capi_to_pin}
fi

# --- Pin sigs.k8s.io/controller-runtime (if present as direct dep) ---
echo "=== Pinning sigs.k8s.io/controller-runtime to ${CONTROLLER_RUNTIME_VERSION} ==="
if awk '$1=="sigs.k8s.io/controller-runtime" && !/indirect/{found=1} END{exit !found}' go.mod; then
    go get sigs.k8s.io/controller-runtime@${CONTROLLER_RUNTIME_VERSION}
fi

echo "=== Running go mod tidy ==="
go mod tidy

echo "Done"
