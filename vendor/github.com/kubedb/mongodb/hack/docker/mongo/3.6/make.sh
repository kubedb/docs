#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubedb/mongodb

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mongo
SUFFIX=v1
DB_VERSION=3.6
TAG="$DB_VERSION-$SUFFIX"

DIST=$REPO_ROOT/dist
mkdir -p $DIST

build() {
  pushd "$REPO_ROOT/hack/docker/mongo/$DB_VERSION"

  # Download Peer-finder
  # ref: peer-finder: https://github.com/kubernetes/contrib/tree/master/peer-finder
  # wget peer-finder: https://github.com/kubernetes/charts/blob/master/stable/mongodb-replicaset/install/Dockerfile#L18
  wget -qO peer-finder http://storage.googleapis.com/kubernetes-release/pets/peer-finder
  chmod +x peer-finder

  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  rm peer-finder
  popd
}

binary_repo $@
