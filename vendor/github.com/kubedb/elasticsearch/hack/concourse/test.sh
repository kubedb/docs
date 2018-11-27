#!/usr/bin/env bash

set -eoux pipefail

#K8S_VERSION=v1.11.2
ORG_NAME=kubedb
REPO_NAME=elasticsearch
OPERATOR_NAME=es-operator
APP_LABEL=kubedb #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/libbuild/concourse/init.sh

cp creds/gcs.json /gcs.json
cp creds/.env $GOPATH/src/github.com/$ORG_NAME/$REPO_NAME/hack/config/.env

pushd "$GOPATH"/src/github.com/$ORG_NAME/$REPO_NAME

./hack/builddeps.sh
./hack/dev/update-docker.sh

# clean the cluster in case previous operator exists
./hack/deploy/setup.sh --uninstall --purge

# run tests
source ./hack/deploy/setup.sh --docker-registry=${DOCKER_REGISTRY}

./hack/make.py test e2e \
  --v=1 \
  --storageclass=${StorageClass:-starndard} \
  --es-version=6.2.4-v1 \
  --selfhosted-operator=true \
  --docker-registry=${DOCKER_REGISTRY} \
  --ginkgo.flakeAttempts=2

#./hack/make.py test e2e \
#  --v=1 \
#  --storageclass=${StorageClass:-starndard} \
#  --es-version=6.3.0 \
#  --selfhosted-operator=true \
#  --docker-registry=${DOCKER_REGISTRY} \
#  --ginkgo.flakeAttempts=2

#./hack/make.py test e2e \
#  --v=1 \
#  --storageclass=${StorageClass:-starndard} \
#  --es-version=5.6.4 \
#  --selfhosted-operator=true \
#  --docker-registry=${DOCKER_REGISTRY} \
#  --ginkgo.flakeAttempts=2
