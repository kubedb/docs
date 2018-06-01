#!/bin/bash

set -x -e

# start docker and log-in to docker-hub
entrypoint.sh
docker login --username=$DOCKER_USER --password=$DOCKER_PASS
docker run hello-world

# install dependencies
apt-get update > /dev/null
apt-get install -y python python-pip awscli > /dev/null

# install kubectl
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &> /dev/null
chmod +x ./kubectl
mv ./kubectl /bin/kubectl

# install onessl
curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-linux-amd64 \
  && chmod +x onessl \
  && mv onessl /usr/local/bin/

# install kops
curl -Lo kops https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod +x ./kops
mv ./kops /usr/local/bin/

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
    kops delete cluster --name ${NAME} --yes

    # delete docker image on exit
    curl -LO https://raw.githubusercontent.com/appscodelabs/libbuild/master/docker.py || true
    chmod +x docker.py || true
    ./docker.py del_tag kubedbci operator $CUSTOM_OPERATOR_TAG || true
}
trap cleanup EXIT

## create cluster using kops
# aws credentials for kops user
export AWS_ACCESS_KEY_ID=$KOPS_AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY=$KOPS_AWS_SECRET_ACCESS_KEY

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

# copy operator to $GOPATH
mkdir -p $GOPATH/src/github.com/kubedb
cp -r operator $GOPATH/src/github.com/kubedb

# build and push operator while cluster is ready
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


# wait for cluster to be ready
while [ "$(kops validate cluster | tail -1)" != "Your cluster ${NAME} is ready" ]; do
    sleep 60
done

# check if cluster is working as expected
kops validate cluster
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
if ! (./hack/make.py test e2e --v=1 --storageclass=gp2 --selfhosted-operator=true); then
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
if ! (./hack/make.py test e2e --v=1 --storageclass=gp2 --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

# test mysql
echo "======================TESTING MYSQL=============================="
git clone https://github.com/kubedb/mysql
pushd mysql
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass=gp2 --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

# test elasticsearch
echo "======================TESTING ELASTICSEARCH============================="
git clone https://github.com/kubedb/elasticsearch
pushd elasticsearch
cp /tmp/.env hack/config/.env
if ! (./hack/make.py test e2e --v=1 --storageclass=gp2 --selfhosted-operator=true); then
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
if ! (./hack/make.py test e2e --v=1 --storageclass=gp2 --selfhosted-operator=true); then
    EXIT_CODE=1
fi
popd

kubectl describe pods -n kube-system -l app=kubedb || true

popd

exit $EXIT_CODE
