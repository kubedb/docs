#!/bin/bash

set -eoux pipefail

export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci
export ClusterProvider=${ClusterProvider:-}

mkdir -p /root/.kube
cp configs/config /root/.kube
cp configs/kubectl /bin/kubectl
kubectl get nodes

# copy operator to $GOPATH
mkdir -p $GOPATH/src/github.com/kubedb
cp -r operator $GOPATH/src/github.com/kubedb

pushd $GOPATH/src/github.com/kubedb/operator
git branch -m $(git rev-parse --abbrev-ref HEAD)-$ClusterProvider
./hack/deploy/setup.sh --docker-registry=kubedbci
popd
