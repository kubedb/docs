#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubedb/memcached"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=memcached
SUFFIX=v1
DB_VERSION=1.5.4
TAG="$DB_VERSION-$SUFFIX"

build() {
  pushd "$REPO_ROOT/hack/docker/memcached/$DB_VERSION"

  chmod +x start.sh
  cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  popd
}

binary_repo $@
