#!/bin/bash -u

root_dir=$(cd "$(dirname "$0")/.." && pwd)
dist_dir="${root_dir}/dist"

pushd "${root_dir}"
rm -rf "${dist_dir}/chart"

# Create Template
kubebuilder edit --plugins=helm/v1-alpha

# Copy template files
cp -r helm/chart/* "${dist_dir}/chart/"

# Create package
helm package "${dist_dir}/chart" --destination "${dist_dir}"

popd
