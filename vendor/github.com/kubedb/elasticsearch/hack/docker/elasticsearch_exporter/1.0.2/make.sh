#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}

IMG_REGISTRY=justwatch
IMG=elasticsearch_exporter
TAG=1.0.2

docker pull "$IMG_REGISTRY/$IMG:$TAG"

docker tag "$IMG_REGISTRY/$IMG:$TAG" "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
