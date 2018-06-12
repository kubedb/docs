#!/bin/bash

set -eoux pipefail

export DOCKER_USER=${DOCKER_USER:-}
export DOCKER_PASS=${DOCKER_PASS:-}
export ClusterProvider=${ClusterProvider:-}

# start docker and log-in to docker-hub
entrypoint.sh
set +x
docker login --username=$DOCKER_USER --password=$DOCKER_PASS
set -x

# copy operator to $GOPATH
mkdir -p $GOPATH/src/github.com/kubedb
cp -r operator $GOPATH/src/github.com/kubedb

# build and push operator docker-image
pushd $GOPATH/src/github.com/kubedb/operator

# changed name of branch
# this is necessary because operator image tag is based on branch name
# for parallel tests, if two test build image of same tag, it'll create problem
# one test may finish early and delete image while other is using it
git branch -m $(git rev-parse --abbrev-ref HEAD)-$ClusterProvider

./hack/builddeps.sh
export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

./hack/docker/setup.sh build
./hack/docker/setup.sh push
popd
