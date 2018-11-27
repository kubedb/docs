#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kubedb
REPO_NAME=memcached
OPERATOR_NAME=mc-operator
APP_LABEL=kubedb #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/libbuild/concourse/init.sh

pushd "$GOPATH"/src/github.com/$ORG_NAME/$REPO_NAME

# build and push docker-image
./hack/builddeps.sh
./hack/dev/update-docker.sh

# uninstall any previous existing configuration
./hack/deploy/setup.sh --uninstall --purge

# run tests
./hack/deploy/setup.sh --docker-registry=kubedbci
./hack/make.py test e2e \
  --v=1 \
  --selfhosted-operator=true \
  --ginkgo.flakeAttempts=2
