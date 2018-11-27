#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubedb/mongodb

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mongodb_exporter
IMG_REGISTRY=dcu
TAG=latest

# Take 1st 8 letters of hash as a shorten hash. Get hash without cloning: https://stackoverflow.com/a/24750310/4628962
COMMIT_HASH=`git ls-remote https://github.com/${IMG_REGISTRY}/${IMG}.git | grep HEAD | awk '{ print substr($1,1,8)}'`

build() {
  pushd "$REPO_ROOT/hack/docker/mongodb_exporter/$TAG"

  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$COMMIT_HASH ."
  echo $cmd; $cmd

  local cmd="docker tag $DOCKER_REGISTRY/$IMG:$COMMIT_HASH $DOCKER_REGISTRY/$IMG:$TAG"
  echo $cmd; $cmd

  popd
}

docker_push() {
  local cmd="docker push $DOCKER_REGISTRY/$IMG:$COMMIT_HASH"
  echo $cmd; $cmd

  local cmd="docker push $DOCKER_REGISTRY/$IMG:$TAG"
  echo $cmd; $cmd
}

binary_repo $@
