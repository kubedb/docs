#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubedb/etcd

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

IMG=etcd
TAG=3.2.13

DIST="$REPO_ROOT/dist"
mkdir -p "$DIST"

build_binary() {
  pushd $REPO_ROOT
  ./hack/builddeps.sh
  ./hack/make.py build etcd-operator
  popd
}

build_docker() {
  pushd "$REPO_ROOT/hack/docker/etcd/$TAG"

  # Copy etcd-operator
  cp "$DIST/etcd-operator/etcd-operator-alpine-amd64" etcd-operator
  chmod 755 etcd-operator

  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  rm etcd-operator
  popd
}

build() {
  build_binary
  build_docker
}

binary_repo $@
