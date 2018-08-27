#!/usr/bin/env bash

set -x

# uninstall operator
./hack/deploy/setup.sh --uninstall --purge
./hack/deploy/setup.sh --uninstall --purge

# remove creds
rm -rf /gcs.json
rm -rf /tmp/.env

# remove docker images
source "hack/libbuild/common/lib.sh"
detect_tag ''

# delete docker image on exit
curl -LO https://raw.githubusercontent.com/appscodelabs/libbuild/master/docker.py
chmod +x docker.py
./docker.py del_tag $DOCKER_REGISTRY operator $TAG
