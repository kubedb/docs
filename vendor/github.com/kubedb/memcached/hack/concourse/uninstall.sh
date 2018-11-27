#!/usr/bin/env bash

set -x

# uninstall operator
./hack/deploy/setup.sh --uninstall --purge
./hack/deploy/setup.sh --uninstall --purge

# remove docker images
source "hack/libbuild/common/lib.sh"
detect_tag ''

# delete docker image on exit
./hack/libbuild/docker.py del_tag $DOCKER_REGISTRY mc-operator $TAG
