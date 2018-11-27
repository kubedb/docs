#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kubedb
REPO_NAME=mongodb
OPERATOR_NAME=mg-operator
APP_LABEL=kubedb #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci
export K8S_VERSION=v1.11.2

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
./hack/deploy/setup.sh --docker-registry=${DOCKER_REGISTRY}

./hack/make.py test e2e \
  --v=1 \
  --storageclass=${StorageClass:-standard} \
  --selfhosted-operator=true \
  --docker-registry=${DOCKER_REGISTRY} \
  --ginkgo.flakeAttempts=2 \
  --db-version=3.6-v1

#./hack/make.py test e2e \
#  --v=1 \
#  --storageclass=$StorageClass \
#  --selfhosted-operator=true \
#  --docker-registry=${DOCKER_REGISTRY} \
#  --ginkgo.flakeAttempts=2 \
#  --db-version=3.4
