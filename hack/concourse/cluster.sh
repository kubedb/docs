#!/bin/bash

set -eoux pipefail

export ClusterProvider=${ClusterProvider:-digitalocean}

function cleanup {
    set +eoux pipefail

    # Workload Descriptions if the test fails
    if [ $? -ne 0 ]; then
        echo ""
        kubectl describe deploy -n kube-system -l app=kubedb
        echo ""
        echo ""
        kubectl describe replicasets -n kube-system -l app=kubedb
        echo ""
        echo ""
        kubectl describe pods -n kube-system -l app=kubedb
    fi

    # delete cluster on exit
    if [ "${ClusterProvider}" = "aws" ]; then
        kops delete cluster --name ${NAME} --yes
    else
        pharmer get cluster
        pharmer delete cluster ${NAME}
        pharmer get cluster
        sleep 120
        pharmer apply ${NAME}
        pharmer get cluster
    fi

    # delete docker image on exit
    curl -LO https://raw.githubusercontent.com/appscodelabs/libbuild/master/docker.py
    chmod +x docker.py
    ./docker.py del_tag kubedbci operator ${CUSTOM_OPERATOR_TAG}
}
trap cleanup EXIT

function pharmer_common {
    export CredProvider=${CredProvider:-DigitalOcean}
    export ZONE=${ZONE:-nyc1}
    export NODE=${NODE:-2gb}
    export K8S_VERSION=${K8S_VERSION:-1.10.0}

    # name of the cluster
    pushd operator
    export NAME=operator-$(git rev-parse --short HEAD)
    popd

    # create cluster using pharmer
    pharmer create credential --from-file=creds/${ClusterProvider}.json --provider=${CredProvider} cred
    pharmer create cluster ${NAME} --provider=${ClusterProvider} --zone=${ZONE} --nodes=${NODE}=1 --credential-uid=cred --v=10 --kubernetes-version=${K8S_VERSION}
    pharmer apply ${NAME} || true
    pharmer apply ${NAME}
}

function prepare_gke {
    pharmer_common

    pushd /tmp
    curl -LO https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-202.0.0-linux-x86_64.tar.gz
    tar --extract --file google-cloud-sdk-202.0.0-linux-x86_64.tar.gz
    CLOUDSDK_CORE_DISABLE_PROMPTS=1 ./google-cloud-sdk/install.sh
    source /tmp/google-cloud-sdk/path.bash.inc
    popd
    gcloud auth activate-service-account --key-file creds/gke.json
    gcloud container clusters get-credentials ${NAME} --zone us-central1-f --project k8s-qa
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=k8s-qa@k8s-qa.iam.gserviceaccount.com

    # wait for cluster to be ready
    sleep 120
}

function prepare_aws {
    # install kops
    curl -Lo kops https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
    chmod +x ./kops
    mv ./kops /usr/local/bin/

    # install awscli
    apt-get update &> /dev/null
    apt-get install -y awscli &> /dev/null

    ## create cluster using kops
    # aws credentials for kops user
    export AWS_ACCESS_KEY_ID=${KOPS_AWS_ACCESS_KEY_ID:-}
    export AWS_SECRET_ACCESS_KEY=${KOPS_AWS_SECRET_ACCESS_KEY:-}

    # name of the cluster
    pushd operator
    export NAME=operator-$(git rev-parse --short HEAD).k8s.local
    popd

    # use s3 bucket for cluster state storage
    export KOPS_STATE_STORE=s3://kubedbci

    # check avability
    aws ec2 describe-availability-zones --region us-east-1

    # generate ssh-keys without prompt
    ssh-keygen -q -t rsa -N '' -f /root/.ssh/id_rsa

    # generate cluster configuration
    kops create cluster --zones us-east-1a  --node-count 1 ${NAME}

    # build cluster
    kops update cluster ${NAME} --yes

    # wait for cluster to be ready
    while [ "$(kops validate cluster | tail -1)" != "Your cluster ${NAME} is ready" ]; do
        sleep 60
    done

    export StorageClass="gp2"
}

function prepare_aks {
    # download azure cli
    AZ_REPO=$(lsb_release -cs)
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $AZ_REPO main" | \
    tee /etc/apt/sources.list.d/azure-cli.list
    curl -L https://packages.microsoft.com/keys/microsoft.asc | apt-key add -
    apt-get install -y apt-transport-https
    apt-get update && apt-get install -y azure-cli

    # login with service principal
    set +x
    az login --service-principal --username $APP_ID --password $PASSWORD --tenant $TENANT_ID &> /dev/null
    set -x

    # create cluster
    #pharmer_common
    az provider register -n Microsoft.Network
    az provider register -n Microsoft.Storage
    az provider register -n Microsoft.Compute
    az provider register -n Microsoft.ContainerService

    az group create --name $NAME --location eastus
    az aks create --resource-group $NAME --name $NAME --node-count 1 --generate-ssh-keys
    az aks get-credentials --resource-group myResourceGroup --name myAKSCluster

    kubectl get nodes

    export StorageClass="default"
}

export StorageClass="standard"

# prepare cluster
if [ "${ClusterProvider}" = "gke" ]; then
    prepare_gke
elif [ "${ClusterProvider}" = "aws" ]; then
    prepare_aws
elif [ "${ClusterProvider}" = "aks" ]; then
    prepare_aks
fi

kubectl get nodes
