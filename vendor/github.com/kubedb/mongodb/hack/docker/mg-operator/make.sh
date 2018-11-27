#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
SRC=$GOPATH/src
BIN=$GOPATH/bin
ROOT=$GOPATH
REPO_ROOT=$GOPATH/src/github.com/kubedb/mongodb

source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

APPSCODE_ENV=${APPSCODE_ENV:-dev}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mg-operator

DIST=$GOPATH/src/github.com/kubedb/mongodb/dist
mkdir -p $DIST
if [ -f "$DIST/.tag" ]; then
  export $(cat $DIST/.tag | xargs)
fi

clean() {
  pushd $REPO_ROOT/hack/docker/mg-operator
  rm -f mg-operator Dockerfile
  popd
}

build_binary() {
  pushd $REPO_ROOT
  ./hack/builddeps.sh
  ./hack/make.py build mg-operator
  detect_tag $DIST/.tag
  popd
}

build_docker() {
  pushd $REPO_ROOT/hack/docker/mg-operator
  cp $DIST/mg-operator/mg-operator-alpine-amd64 mg-operator
  chmod 755 mg-operator

  cat >Dockerfile <<EOL
FROM alpine:3.8

RUN set -x \
  && apk add --update --no-cache ca-certificates

COPY mg-operator /usr/bin/mg-operator

USER nobody:nobody
ENTRYPOINT ["mg-operator"]
EOL
  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  rm mg-operator Dockerfile
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
