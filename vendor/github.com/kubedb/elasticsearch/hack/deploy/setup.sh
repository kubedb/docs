#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
export KUBEDB_DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
export KUBEDB_NAMESPACE=${KUBEDB_NAMESPACE:-kube-system}
export MINIKUBE=0
export MINIKUBE_RUN=0
export SELF_HOSTED=1
export ARGS="" # Forward arguments to installer script

REPO_ROOT="$GOPATH/src/github.com/kubedb/elasticsearch"
CLI_ROOT="$GOPATH/src/github.com/kubedb/cli"

pushd $REPO_ROOT

# https://stackoverflow.com/a/677212/244009
if [[ ! -z "$(command -v onessl)" ]]; then
  export ONESSL=onessl
else
  # ref: https://stackoverflow.com/a/27776822/244009
  case "$(uname -s)" in
    Darwin)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.7.0/onessl-darwin-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    Linux)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.7.0/onessl-linux-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    CYGWIN* | MINGW32* | MSYS*)
      curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.7.0/onessl-windows-amd64.exe
      chmod +x onessl.exe
      export ONESSL=./onessl.exe
      ;;
    *)
      echo 'other OS'
      ;;
  esac
fi

source "$REPO_ROOT/hack/deploy/settings"
source "$REPO_ROOT/hack/libbuild/common/lib.sh"

export KUBE_CA=$($ONESSL get kube-ca | $ONESSL base64)
export APPSCODE_ENV=${APPSCODE_ENV:-prod}
export KUBEDB_SCRIPT="curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.9.0-rc.0/"

show_help() {
  echo "setup.sh - setup kubedb operator"
  echo " "
  echo "setup.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help          show brief help"
  echo "    --selfhosted    deploy operator cluster."
  echo "    --minikube      setup configurations for minikube to run operator in localhost"
  echo "    --run           run operator in localhost and connect with minikube. only works with --minikube flag"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      ARGS="$ARGS $1" # also show helps of "CLI repo" installer script
      shift
      ;;
    --docker-registry*)
      export KUBEDB_DOCKER_REGISTRY=$(echo $1 | sed -e 's/^[^=]*=//g')
      ARGS="$ARGS $1"
      shift
      ;;
    --minikube)
      export APPSCODE_ENV=dev
      export MINIKUBE=1
      export SELF_HOSTED=0
      shift
      ;;
    --run)
      export MINIKUBE_RUN=1
      shift
      ;;
    --selfhosted)
      export MINIKUBE=0
      export SELF_HOSTED=1
      shift
      ;;
    *)
      ARGS="$ARGS $1"
      shift
      ;;
  esac
done

# If APPSCODE_ENV==dev , use cli repo locally to run the installer script.
# Update "CLI_BRANCH" in deploy/settings file to pull a particular CLI repo branch.
if [ "$APPSCODE_ENV" = "dev" ]; then
  detect_tag ''
  export KUBEDB_SCRIPT="cat $CLI_ROOT/"
  export CUSTOM_OPERATOR_TAG=$TAG
  echo ""

  if [[ ! -d $CLI_ROOT ]]; then
    echo ">>> Cloning cli repo"
    git clone -b $CLI_BRANCH https://github.com/kubedb/cli.git "${CLI_ROOT}"
    pushd $CLI_ROOT
  else
    pushd $CLI_ROOT
    detect_tag ''
    if [[ $git_branch != $CLI_BRANCH ]]; then
      git fetch --all
      git checkout $CLI_BRANCH
    fi
    git pull --ff-only origin $CLI_BRANCH #Pull update from remote only if there will be no conflict.
  fi
fi

echo ""
env | sort | grep -e KUBEDB* -e APPSCODE*
echo ""

if [ "$SELF_HOSTED" -eq 1 ]; then
  echo "${KUBEDB_SCRIPT}hack/deploy/kubedb.sh | bash -s -- --operator-name=es-operator $ARGS"
  ${KUBEDB_SCRIPT}hack/deploy/kubedb.sh | bash -s -- --operator-name=es-operator ${ARGS}
fi

if [ "$MINIKUBE" -eq 1 ]; then
  cat $CLI_ROOT/hack/deploy/validating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
  cat $CLI_ROOT/hack/deploy/mutating-webhook.yaml | $ONESSL envsubst | kubectl apply -f -
  cat $REPO_ROOT/hack/dev/apiregistration.yaml | $ONESSL envsubst | kubectl apply -f -
  # Following line may give error if DBVersions CRD already not created
  cat $CLI_ROOT/hack/deploy/kubedb-catalog/elasticsearch.yaml | $ONESSL envsubst | kubectl apply -f - || true

  if [ "$MINIKUBE_RUN" -eq 1 ]; then
    $REPO_ROOT/hack/make.py
    es-operator run --v=4 \
      --secure-port=8443 \
      --enable-status-subresource=true \
      --enable-mutating-webhook=true \
      --enable-validating-webhook=true \
      --kubeconfig="$HOME/.kube/config" \
      --authorization-kubeconfig="$HOME/.kube/config" \
      --authentication-kubeconfig="$HOME/.kube/config"
  fi
fi

if [ $(pwd) = "$CLI_ROOT" ]; then
  popd
fi
popd
