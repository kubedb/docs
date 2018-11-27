#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=redis
SUFFIX=v1
DB_VERSION=4.0.6
TAG="$DB_VERSION-$SUFFIX"

docker pull $IMG:$DB_VERSION-alpine

docker tag $IMG:$DB_VERSION-alpine "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
