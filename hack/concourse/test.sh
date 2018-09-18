#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kubedb
REPO_NAME=operator
APP_LABEL=kubedb #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/libbuild/concourse/init.sh

# create config/.env file that have all necessary creds
cp creds/gcs.json /gcs.json
cp creds/.env /tmp/.env

pushd "$GOPATH"/src/github.com/$ORG_NAME/$REPO_NAME
./hack/builddeps.sh
./hack/docker/setup.sh build
./hack/docker/setup.sh push
popd

pushd "$GOPATH"/src/github.com/kubedb

# deploy operator
pushd operator
source ./hack/deploy/setup.sh --docker-registry=kubedbci
popd

EXIT_CODE=0

# test redis
cowsay -f tux "testing redis"
git clone https://github.com/kubedb/redis
pushd redis
./hack/dev/update-docker.sh
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --docker-registry="$DOCKER_REGISTRY" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

# test memcached
cowsay -f tux "testing memcached"
git clone https://github.com/kubedb/memcached
pushd memcached
./hack/dev/update-docker.sh
if ! (./hack/make.py test e2e --v=1 --docker-registry="$DOCKER_REGISTRY" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

# test elasticsearch
cowsay -f tux "testing elasticsearch"
git clone https://github.com/kubedb/elasticsearch
pushd elasticsearch
cp /tmp/.env hack/config/.env
./hack/dev/update-docker.sh
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --docker-registry="$DOCKER_REGISTRY" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
rm -rf hack/config/.env
popd

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

# test postgres
cowsay -f tux "testing postgres"
git clone https://github.com/kubedb/postgres
pushd postgres
cp /tmp/.env hack/config/.env
./hack/dev/update-docker.sh
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --docker-registry="$DOCKER_REGISTRY" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
rm -rf hack/config/.env
popd

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

# test mongodb
cowsay -f tux "testing mongodb"
git clone https://github.com/kubedb/mongodb
pushd mongodb
cp /tmp/.env hack/config/.env
./hack/dev/update-docker.sh
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --docker-registry="$DOCKER_REGISTRY" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
rm -rf hack/config/.env
popd

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

# test mysql
cowsay -f tux "testing mysql"
git clone https://github.com/kubedb/mysql
pushd mysql
cp /tmp/.env hack/config/.env
./hack/dev/update-docker.sh
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --docker-registry="$DOCKER_REGISTRY" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
rm -rf hack/config/.env
popd

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

popd

exit $EXIT_CODE
