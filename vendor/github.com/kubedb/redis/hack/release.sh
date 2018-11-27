#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubedb/redis"

export APPSCODE_ENV=prod

pushd $REPO_ROOT

rm -rf dist

./hack/docker/rd-operator/make.sh
./hack/docker/rd-operator/make.sh release

rm dist/.tag

popd
