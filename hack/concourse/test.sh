#!/usr/bin/env bash
# shellcheck disable=SC1090,SC1091

set -eoux pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/kubedb/operator"
export ClusterProvider=${ClusterProvider:-digitalocean}
export StorageClass=${StorageClass:-standard}

# copy operator to $GOPATH
mkdir -p "$GOPATH"/src/github.com/kubedb
cp -r operator "$GOPATH"/src/github.com/kubedb

# install all the dependencies and prepeare cluster
source "$REPO_ROOT/hack/concourse/dependencies.sh"
source "$REPO_ROOT/hack/concourse/cluster.sh"

# build and push operator docker-image
pushd "$GOPATH"/src/github.com/kubedb/operator

# changed name of branch
# this is necessary because operator image tag is based on branch name
# for parallel tests, if two test build image of same tag, it'll create problem
# one test may finish early and delete image while other is using it
git branch -m "$(git rev-parse --abbrev-ref HEAD)-$ClusterProvider"

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
echo "======================TESTING REDIS=============================="
git clone https://github.com/kubedb/redis
pushd redis
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=3); then
    EXIT_CODE=1
fi
popd
sleep 120

kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true
echo ""
echo "::::::::::::::::::::::::::: Describe Nodes :::::::::::::::::::::::::::"
echo ""
kubectl get nodes || true
echo ""
kubectl describe nodes || true

# test memcached
echo "======================TESTING MEMCACHED=============================="
git clone https://github.com/kubedb/memcached
pushd memcached
if ! (./hack/make.py test e2e --v=1 --selfhosted-operator=true --ginkgo.flakeAttempts=3); then
    EXIT_CODE=1
fi
popd
sleep 120

kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true
echo ""
echo "::::::::::::::::::::::::::: Describe Nodes :::::::::::::::::::::::::::"
echo ""
kubectl get nodes || true
echo ""
kubectl describe nodes || true

# test elasticsearch
echo "======================TESTING ELASTICSEARCH============================="
git clone https://github.com/kubedb/elasticsearch
pushd elasticsearch
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=3); then
    EXIT_CODE=1
fi
popd
sleep 120

kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true
echo ""
echo "::::::::::::::::::::::::::: Describe Nodes :::::::::::::::::::::::::::"
echo ""
kubectl get nodes || true
echo ""
kubectl describe nodes || true

# test postgres
echo "======================TESTING POSTGRES=============================="
git clone https://github.com/kubedb/postgres
pushd postgres
cp /tmp/.env hack/config/.env
./hack/docker/postgres/9.6.7/make.sh build
./hack/docker/postgres/9.6.7/make.sh push
./hack/docker/postgres/9.6/make.sh
./hack/docker/postgres/10.2/make.sh build
./hack/docker/postgres/10.2/make.sh push
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=3); then
    EXIT_CODE=1
fi
popd
sleep 120

kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true
echo ""
echo "::::::::::::::::::::::::::: Describe Nodes :::::::::::::::::::::::::::"
echo ""
kubectl get nodes || true
echo ""
kubectl describe nodes || true

# test mongodb
echo "======================TESTING MONGODB=============================="
git clone https://github.com/kubedb/mongodb
pushd mongodb
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=3); then
    EXIT_CODE=1
fi
popd
sleep 120

kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true
echo ""
echo "::::::::::::::::::::::::::: Describe Nodes :::::::::::::::::::::::::::"
echo ""
kubectl get nodes || true
echo ""
kubectl describe nodes || true

# test mysql
echo "======================TESTING MYSQL=============================="
git clone https://github.com/kubedb/mysql
pushd mysql
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass="$StorageClass" --selfhosted-operator=true --ginkgo.flakeAttempts=3); then
    EXIT_CODE=1
fi
popd

kubectl get pods --all-namespaces
kubectl describe pods -n kube-system -l app=kubedb || true
echo ""
echo "::::::::::::::::::::::::::: Describe Nodes :::::::::::::::::::::::::::"
echo ""
kubectl get nodes || true
echo ""
kubectl describe nodes || true

popd

exit $EXIT_CODE
