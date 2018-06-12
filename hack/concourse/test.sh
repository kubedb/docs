#!/bin/bash

set -eoux pipefail

apt-get update &> /dev/null
apt-get install -y git &> /dev/null

mkdir -p /root/.kube
cp configs/config /root/.kube/
cp configs/kubectl /bin/kubectl
kubectl get nodes

export ROOT=$(pwd)
export DB=${DB:-}

# create config/.env file that have all necessary creds
cp creds/gcs.json /gcs.json

mkdir -p $GOPATH/src/github.com/kubedb
pushd $GOPATH/src/github.com/kubedb
git clone https://github.com/kubedb/$DB
cd $DB

if [ -d "hack/config" ]; then
    cp $ROOT/creds/.env hack/config/.env
fi

if [ "$DB" = "memcached" ]; then
    ./hack/make.py test e2e --v=1 --selfhosted-operator=true
else
    ./hack/make.py test e2e --v=1 --storageclass=$StorageClass --selfhosted-operator=true
fi
popd
