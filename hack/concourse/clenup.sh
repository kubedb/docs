#!/bin/bash

set -x

mkdir -p /root/.kube
cp configs/config /root/.kube/config
cp configs/pharmer /bin/pharmer
cp -r configs/.pharmer /root/.pharmer
cp configs/kubectl /bin/kubectl
kubectl get nodes

export ClusterProvider=${ClusterProvider:-}

# name of cluster
pushd operator
export NAME=operator-$(git rev-parse --short HEAD)
export OPERATOR_TAG=$(git rev-parse --abbrev-ref HEAD)-$ClusterProvider
popd

if [ "$ClusterProvider" = "aws" ]; then
    kops delete cluster --name $NAME --yes
elif [[ "$ClusterProvider" = "aks" || "$ClusterProvider" = "acs" ]]; then
    az group delete --name $NAME --yes --no-wait
else
    pharmer get cluster
    pharmer delete cluster $NAME
    pharmer get cluster
    sleep 300
    pharmer apply $NAME
    pharmer get cluster
fi

# delete docker image on exit
curl -LO https://raw.githubusercontent.com/appscodelabs/libbuild/master/docker.py
chmod +x docker.py
./docker.py del_tag kubedbci operator $OPERATOR_TAG
