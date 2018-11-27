#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG_REGISTRY=oliver006
IMG=redis_exporter
TAG=v0.21.1
# Available image tags: https://hub.docker.com/r/oliver006/redis_exporter/tags/
# Prometheus officially suggested exporter list: https://prometheus.io/docs/instrumenting/exporters/

docker pull $IMG_REGISTRY/$IMG:$TAG

docker tag $IMG_REGISTRY/$IMG:$TAG "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
