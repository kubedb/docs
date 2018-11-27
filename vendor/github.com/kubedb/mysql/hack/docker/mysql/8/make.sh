#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}

IMG=mysql
SUFFIX=v1

DB_VERSION=8
PATCH=8.0.3
TAG="$DB_VERSION-$SUFFIX"

docker pull $IMG:$PATCH

docker tag $IMG:$PATCH "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"