#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubedb/elasticsearch"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=kibana
TAG=6.3.0

build() {
    pushd "$REPO_ROOT/hack/docker/$IMG/$TAG"

    local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
    echo $cmd; $cmd

    popd
}

binary_repo $@
