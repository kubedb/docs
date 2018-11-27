#!/bin/bash
set -xeou pipefail

IMG=elasticsearch-tools
SUFFIX=v1
TAG="5.6-$SUFFIX"
PATCH="5.6.4-$SUFFIX"

docker pull "$DOCKER_REGISTRY/$IMG:$PATCH"

docker tag "$DOCKER_REGISTRY/$IMG:$PATCH" "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
