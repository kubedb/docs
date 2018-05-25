#!/bin/bash

set -x -e

# start docker and log-in to docker-hub
entrypoint.sh
docker login --username=$DOCKER_USER --password=$DOCKER_PASS
docker run hello-world

# install python pip
apt-get update > /dev/null
apt-get install -y python python-pip > /dev/null

# install kubectl
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &> /dev/null
chmod +x ./kubectl
mv ./kubectl /bin/kubectl

# install onessl
curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-linux-amd64 \
  && chmod +x onessl \
  && mv onessl /usr/local/bin/

# install pharmer
pushd /tmp
curl -LO https://cdn.appscode.com/binaries/pharmer/0.1.0-rc.4/pharmer-linux-amd64
chmod +x pharmer-linux-amd64
mv pharmer-linux-amd64 /bin/pharmer
popd

function cleanup {
    # Workload Descriptions if the test fails
    if [ $? -ne 0 ]; then
        echo ""
        kubectl describe deploy -n kube-system -l app=kubedb || true
        echo ""
        echo ""
        kubectl describe replicasets -n kube-system -l app=kubedb || true
        echo ""
        echo ""
        kubectl describe pods -n kube-system -l app=kubedb || true
    fi

    # delete cluster on exit
    pharmer get cluster || true
    pharmer delete cluster $NAME || true
    pharmer get cluster || true
    sleep 120 || true
    pharmer apply $NAME || true
    pharmer get cluster || true

    # delete docker image on exit
    curl -LO https://raw.githubusercontent.com/appscodelabs/libbuild/master/docker.py || true
    chmod +x docker.py || true
    ./docker.py del_tag kubedbci operator $CUSTOM_OPERATOR_TAG || true
}
trap cleanup EXIT

# name of the cluster
# nameing is based on repo+commit_hash
pushd operator
NAME=operator-$(git rev-parse --short HEAD)
popd

# copy operator to $GOPATH
mkdir -p $GOPATH/src/github.com/kubedb
cp -r operator $GOPATH/src/github.com/kubedb

pushd $GOPATH/src/github.com/kubedb/operator

# build and push operator docker-image
./hack/builddeps.sh
export APPSCODE_ENV=dev
export DOCKER_REGISTRY=kubedbci
pip install git+https://github.com/ellisonbg/antipackage.git#egg=antipackage
go get -u golang.org/x/tools/cmd/goimports
go get github.com/Masterminds/glide
go get github.com/sgotti/glide-vc
go get github.com/onsi/ginkgo/ginkgo
go install github.com/onsi/ginkgo/ginkgo

./hack/docker/setup.sh build
./hack/docker/setup.sh push

popd

# create cluster using pharmer
pharmer create credential --from-file=creds/gcs/gke.json --provider=GoogleCloud cred
pharmer create cluster $NAME --provider=gke --zone=us-central1-f --nodes=n1-standard-2=1 --credential-uid=cred --v=10 --kubernetes-version=1.10.2-gke.3
pharmer apply $NAME

# gcloud-sdk
pushd /tmp
curl -LO https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-202.0.0-linux-x86_64.tar.gz
tar --extract --file google-cloud-sdk-202.0.0-linux-x86_64.tar.gz
CLOUDSDK_CORE_DISABLE_PROMPTS=1 ./google-cloud-sdk/install.sh
source /tmp/google-cloud-sdk/path.bash.inc
popd
gcloud auth activate-service-account --key-file creds/gcs/gke.json
gcloud container clusters get-credentials $NAME --zone us-central1-f --project k8s-qa
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=k8s-qa@k8s-qa.iam.gserviceaccount.com

# wait for cluster to be ready
sleep 300
kubectl get nodes


export CRED_DIR=$(pwd)/creds/gcs/gcs.json

# create config/.env file that have all necessary creds
cat > /tmp/.env <<EOF
AWS_ACCESS_KEY_ID=$AWS_KEY_ID
AWS_SECRET_ACCESS_KEY=$AWS_SECRET

GOOGLE_PROJECT_ID=$GCE_PROJECT_ID
GOOGLE_APPLICATION_CREDENTIALS=$CRED_DIR

AZURE_ACCOUNT_NAME=$AZURE_ACCOUNT_NAME
AZURE_ACCOUNT_KEY=$AZURE_ACCOUNT_KEY

OS_AUTH_URL=$OS_AUTH_URL
OS_TENANT_ID=$OS_TENANT_ID
OS_TENANT_NAME=$OS_TENANT_NAME
OS_USERNAME=$OS_USERNAME
OS_PASSWORD=$OS_PASSWORD
OS_REGION_NAME=$OS_REGION_NAME

S3_BUCKET_NAME=$S3_BUCKET_NAME
GCS_BUCKET_NAME=$GCS_BUCKET_NAME
AZURE_CONTAINER_NAME=$AZURE_CONTAINER_NAME
SWIFT_CONTAINER_NAME=$SWIFT_CONTAINER_NAME
EOF

pushd $GOPATH/src/github.com/kubedb

# deploy operator
pushd operator
./hack/deploy/setup.sh --docker-registry=kubedbci
popd

EXIT_CODE=0

# test redis
echo "======================TESTING REDIS=============================="
git clone https://github.com/kubedb/redis
pushd redis
if ! (./hack/make.py test e2e --v=1 --storageclass=standard --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

# test memcached
echo "======================TESTING MEMCACHED=============================="
git clone https://github.com/kubedb/memcached
pushd memcached
if ! (./hack/make.py test e2e --v=1 --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

# test mongodb
echo "======================TESTING MONGODB=============================="
git clone https://github.com/kubedb/mongodb
pushd mongodb
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass=standard --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

# test mysql
echo "======================TESTING MYSQL=============================="
git clone https://github.com/kubedb/mysql
pushd mysql
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass=standard --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

# test elasticsearch
echo "======================TESTING ELASTICSEARCH============================="
git clone https://github.com/kubedb/elasticsearch
pushd elasticsearch
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass=standard --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

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
if ! (./hack/make.py test e2e --v=1 --storageclass=standard --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

popd

exit $EXIT_CODE
