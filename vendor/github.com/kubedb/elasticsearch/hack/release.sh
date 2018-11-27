#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubedb/elasticsearch"

export APPSCODE_ENV=prod

pushd $REPO_ROOT

rm -rf dist

./hack/docker/es-operator/make.sh
./hack/docker/es-operator/make.sh release

rm dist/.tag

popd
