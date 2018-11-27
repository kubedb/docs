#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}

IMG=mysql
SUFFIX=v1

DB_VERSION=5.7
TAG="$DB_VERSION-$SUFFIX"

docker pull $IMG:$DB_VERSION

docker tag $IMG:$DB_VERSION "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
