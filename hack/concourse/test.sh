#!/usr/bin/env bash

set -eoux pipefail

REPO_NAME=operator

# get concourse-common
pushd $REPO_NAME
git status
git subtree pull --prefix hack/concourse/common https://github.com/kubedb/concourse-common.git master --squash -m 'concourse'
popd

source $REPO_NAME/hack/concourse/common/init.sh

pushd "$GOPATH"/src/github.com/kubedb/$REPO_NAME

./hack/builddeps.sh
export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci

./hack/docker/setup.sh build
./hack/docker/setup.sh push
popd

cp creds/gcs.json /gcs.json

# create config/.env file that have all necessary creds
cp creds/.env /tmp/.env

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
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd
sleep 120

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

cowsay -f tux "describe nodes"
kubectl get nodes || true
kubectl describe nodes || true

# test memcached
cowsay -f tux "testing memcached"
git clone https://github.com/kubedb/memcached
pushd memcached
if ! (./hack/make.py test e2e --v=1 --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd
sleep 120

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

cowsay -f tux "describe nodes"
kubectl get nodes || true
kubectl describe nodes || true

# test elasticsearch
cowsay -f tux "testing elasticsearch"
git clone https://github.com/kubedb/elasticsearch
pushd elasticsearch
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd
sleep 120

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

cowsay -f tux "describe nodes"
kubectl get nodes || true
kubectl describe nodes || true

# test postgres
cowsay -f tux "testing postgres"
git clone https://github.com/kubedb/postgres
pushd postgres
cp /tmp/.env hack/config/.env
./hack/docker/postgres/9.6.7/make.sh build
./hack/docker/postgres/9.6.7/make.sh push
./hack/docker/postgres/9.6/make.sh
./hack/docker/postgres/10.2/make.sh build
./hack/docker/postgres/10.2/make.sh push
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd
sleep 120

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

cowsay -f tux "describe nodes"
kubectl get nodes || true
kubectl describe nodes || true

# test mongodb
cowsay -f tux "testing mongodb"
git clone https://github.com/kubedb/mongodb
pushd mongodb
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd
sleep 120

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

cowsay -f tux "describe nodes"
kubectl get nodes || true
kubectl describe nodes || true

# test mysql
cowsay -f tux "testing mysql"
git clone https://github.com/kubedb/mysql
pushd mysql
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=2); then
  EXIT_CODE=1
fi
popd

cowsay -f tux "describe pods"
kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true

cowsay -f tux "describe nodes"
kubectl get nodes || true
kubectl describe nodes || true

popd

exit $EXIT_CODE
