#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
SRC=$GOPATH/src
BIN=$GOPATH/bin
ROOT=$GOPATH
REPO_ROOT=$GOPATH/src/github.com/kubedb/elasticsearch

source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

APPSCODE_ENV=${APPSCODE_ENV:-dev}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=es-operator

DIST=$GOPATH/src/github.com/kubedb/elasticsearch/dist
mkdir -p $DIST
if [ -f "$DIST/.tag" ]; then
  export $(cat $DIST/.tag | xargs)
fi

clean() {
  pushd $REPO_ROOT/hack/docker/es-operator
  rm -f es-operator Dockerfile
  popd
}

build_binary() {
  pushd $REPO_ROOT
  ./hack/builddeps.sh
  ./hack/make.py build es-operator
  detect_tag $DIST/.tag
  popd
}

build_docker() {
  pushd $REPO_ROOT/hack/docker/es-operator
  cp $DIST/es-operator/es-operator-alpine-amd64 es-operator
  chmod 755 es-operator

  cat >Dockerfile <<EOL
FROM alpine:3.8

RUN set -x \
  && apk add --update --no-cache ca-certificates openssl openjdk8-jre-base

COPY es-operator /usr/bin/es-operator

ENTRYPOINT ["es-operator"]
EOL
  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  rm es-operator Dockerfile
  popd
}

build() {
  build_binary
  build_docker
}

docker_push() {
  if [ "$APPSCODE_ENV" = "prod" ]; then
    echo "Nothing to do in prod env. Are you trying to 'release' binaries to prod?"
    exit 1
  fi
  if [ "$TAG_STRATEGY" = "git_tag" ]; then
    echo "Are you trying to 'release' binaries to prod?"
    exit 1
  fi
  hub_canary
}

docker_release() {
  if [ "$APPSCODE_ENV" != "prod" ]; then
    echo "'release' only works in PROD env."
    exit 1
  fi
  if [ "$TAG_STRATEGY" != "git_tag" ]; then
    echo "'apply_tag' to release binaries and/or docker images."
    exit 1
  fi
  hub_up
}

source_repo $@
