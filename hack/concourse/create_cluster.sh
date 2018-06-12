#!/bin/bash

set -eoux pipefail

# install pharmer
curl -LO https://cdn.appscode.com/binaries/pharmer/0.1.0-rc.4/pharmer-linux-amd64
chmod +x pharmer-linux-amd64
cp pharmer-linux-amd64 /bin/pharmer

# install kubectl
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &> /dev/null
chmod +x ./kubectl
mv ./kubectl /bin/kubectl

# name of the cluster
pushd operator
export NAME=operator-$(git rev-parse --short HEAD)
popd

export ClusterProvider=${ClusterProvider:-}
export StorageClass="standard"

if [ "$ClusterProvider" = "digitalocean" ]; then
    export CredProvider="DigitalOcean"
    export ZONE="nyc1"
    export NODE="4gb"
    export K8S_VERSION="v1.10.0"

    prepare_pharmer
elif [ "$ClusterProvider" = "gke" ]; then
    export CredProvider="GoogleCloud"
    export ZONE="us-central1-f"
    export NODE="n1-standard-2"
    export K8S_VERSION="1.10.2-gke.3"

    prepare_pharmer
elif [ "$ClusterProvider" = "aks" ]; then
    export CredProvider="Azure"
    export ZONE="eastus"
    export NODE="Standard_DS2_v2"
    export K8S_VERSION="1.9.6"

    prepare_aks
elif [ "$ClusterProvider" = "acs" ]; then
    export ZONE="westcentralus"
    export NODE="Standard_DS2_v2"
    export K8S_VERSION="1.10.3"

    prepare_acs
elif [ "$ClusterProvider" = "aws" ]; then
    prepare_aws
else
    echo "unknown provider"
    exit 1
fi

kubectl get nodes
cp /root/.kube/config configs/
cp /bin/pharmer configs/
cp /bin/kubectl configs/
cp -r /root/.pharmer configs

# create cluster using pharmer
function prepare_pharmer {
    pharmer create credential --from-file=creds/${ClusterProvider}.json --provider=${CredProvider} cred
    pharmer create cluster ${NAME} --provider=${ClusterProvider} --zone=${ZONE} --nodes=${NODE}=1 --credential-uid=cred --v=10 --kubernetes-version=${K8S_VERSION}
    pharmer apply ${NAME} || true
    pharmer apply ${NAME}
    pharmer use cluster ${NAME}
    sleep 120
}

# prepare cluster for aws using kops
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
    set +x
    export AWS_ACCESS_KEY_ID=${KOPS_AWS_ACCESS_KEY_ID:-}
    export AWS_SECRET_ACCESS_KEY=${KOPS_AWS_SECRET_ACCESS_KEY:-}
    set -x

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
    end=$((SECONDS+900))
    while [ $SECONDS -lt $end ]; do
        if (kops validate cluster); then
            break
        else
            sleep 60
        fi
    done

    export StorageClass="gp2"
}

function azure_common {
    export StorageClass="default"

    # download azure cli
    AZ_REPO=$(lsb_release -cs)
    echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ $AZ_REPO main" | \
    tee /etc/apt/sources.list.d/azure-cli.list
    curl -L https://packages.microsoft.com/keys/microsoft.asc | apt-key add -
    apt-get install -y apt-transport-https &> /dev/null
    apt-get update &> /dev/null
    apt-get install -y azure-cli &> /dev/null

    # login with service principal
    set +x
    az login --service-principal --username $APP_ID --password $PASSWORD --tenant $TENANT_ID &> /dev/null
    az group create --name $NAME --location $ZONE
    set -x
}

function prepare_aks {
    azure_common
    set +x
    az aks create --resource-group $NAME --name $NAME --service-principal $APP_ID --client-secret $PASSWORD --generate-ssh-keys --node-vm-size $NODE --node-count 1 --kubernetes-version $K8S_VERSION &> /dev/null
    set -x
    az aks get-credentials --resource-group $NAME --name $NAME
}

function prepare_acs {
    azure_common
    set +x
    az acs create --orchestrator-type kubernetes --orchestrator-version $K8S_VERSION --resource-group $NAME --name $NAME --master-vm-size $NODE --agent-vm-size $NODE --agent-count 1 --service-principal $APP_ID --client-secret $PASSWORD --generate-ssh-keys &> /dev/null
    set -x
    az acs kubernetes get-credentials --resource-group $NAME --name $NAME
}
