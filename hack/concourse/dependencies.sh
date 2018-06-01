#!/bin/bash
set -eoux pipefail

export DOCKER_USER=${DOCKER_USER:-}
export DOCKER_PASS=${DOCKER_PASS:-}

# start docker and log-in to docker-hub
entrypoint.sh
docker login --username=${DOCKER_USER} --password=${DOCKER_PASS}
docker run hello-world

# install python pip
apt-get update > /dev/null
apt-get install -y python python-pip > /dev/null

# install kubectl
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &> /dev/null
chmod +x ./kubectl
mv ./kubectl /bin/kubectl

# install onessl
curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-linux-amd64
chmod +x onessl
mv onessl /usr/local/bin/

# install pharmer
curl -LO https://cdn.appscode.com/binaries/pharmer/0.1.0-rc.4/pharmer-linux-amd64
chmod +x pharmer-linux-amd64
mv pharmer-linux-amd64 /bin/pharmer
