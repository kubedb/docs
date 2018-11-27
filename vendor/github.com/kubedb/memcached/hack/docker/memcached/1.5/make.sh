#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=memcached
SUFFIX=v1
TAG="1.5-$SUFFIX"
PATCH="1.5.4-$SUFFIX"

docker pull "$DOCKER_REGISTRY/$IMG:$PATCH"

docker tag "$DOCKER_REGISTRY/$IMG:$PATCH" "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
