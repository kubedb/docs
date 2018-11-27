#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubedb/postgres"

export APPSCODE_ENV=prod

pushd $REPO_ROOT

rm -rf dist

./hack/docker/pg-operator/make.sh
./hack/docker/pg-operator/make.sh release

rm dist/.tag

popd
