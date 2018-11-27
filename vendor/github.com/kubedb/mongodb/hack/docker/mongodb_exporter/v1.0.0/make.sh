#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubedb/mongodb

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mongodb_exporter
TAG=v1.0.0

build() {
  pushd "$REPO_ROOT/hack/docker/mongodb_exporter/$TAG"

  # Download mongodb_exporter. github repo: https://github.com/dcu/mongodb_exporter
  # Prometheus Exporters link: https://prometheus.io/docs/instrumenting/exporters/
  wget -O mongodb_exporter https://github.com/dcu/mongodb_exporter/releases/download/$TAG/mongodb_exporter-linux-amd64
  chmod +x mongodb_exporter

  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  rm mongodb_exporter
  popd
}

binary_repo $@
