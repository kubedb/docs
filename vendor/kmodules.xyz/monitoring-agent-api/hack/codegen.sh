#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=kmodules.xyz/monitoring-agent-api
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"

pushd $REPO_ROOT

mkdir -p "$REPO_ROOT"/api/api-rules

# Generate deep copies
docker run --rm -ti -u $(id -u):$(id -g) \
    -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
    -w "$DOCKER_REPO_ROOT" \
    appscode/gengo:release-1.12 deepcopy-gen \
    --v 1 --logtostderr \
    --go-header-file "hack/gengo/boilerplate.go.txt" \
    --input-dirs "$PACKAGE_NAME/api/v1" \
    --output-file-base zz_generated.deepcopy

# Generate openapi
docker run --rm -ti -u $(id -u):$(id -g) \
    -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
    -w "$DOCKER_REPO_ROOT" \
    appscode/gengo:release-1.12 openapi-gen \
    --v 1 --logtostderr \
    --go-header-file "hack/gengo/boilerplate.go.txt" \
    --input-dirs "$PACKAGE_NAME/api/v1" \
    --output-package "$PACKAGE_NAME/api/v1" \
    --report-filename api/api-rules/violation_exceptions.list

popd
