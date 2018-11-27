#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/appscode/osm"

export APPSCODE_ENV=prod

pushd $REPO_ROOT

rm -rf dist

./hack/docker/setup.sh
./hack/docker/setup.sh release
./hack/make.py push

rm dist/.tag

popd
