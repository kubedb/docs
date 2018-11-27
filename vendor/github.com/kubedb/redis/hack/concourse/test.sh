#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kubedb
REPO_NAME=redis
OPERATOR_NAME=rd-operator
APP_LABEL=kubedb #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/libbuild/concourse/init.sh

pushd "$GOPATH"/src/github.com/kubedb/$REPO_NAME

# build and push docker-image
./hack/builddeps.sh
./hack/dev/update-docker.sh

# clean the cluster in case previous operator exists
./hack/deploy/setup.sh --uninstall --purge

# run tests
./hack/deploy/setup.sh --docker-registry=${DOCKER_REGISTRY}
./hack/make.py test e2e \
  --v=1 \
  --storageclass=${StorageClass:-standard} \
  --selfhosted-operator=true \
  --docker-registry=${DOCKER_REGISTRY} \
  --ginkgo.flakeAttempts=2
