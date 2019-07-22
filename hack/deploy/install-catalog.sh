#!/bin/bash
set -eou pipefail

crds=(
  dormantdatabases.kubedb.com
  elasticsearches.kubedb.com
  etcds.kubedb.com
  memcacheds.kubedb.com
  mongodbs.kubedb.com
  mysqls.kubedb.com
  postgreses.kubedb.com
  redises.kubedb.com
  snapshots.kubedb.com
  elasticsearchversions.catalog.kubedb.com
  etcdversions.catalog.kubedb.com
  memcachedversions.catalog.kubedb.com
  mongodbversions.catalog.kubedb.com
  mysqlversions.catalog.kubedb.com
  postgresversions.catalog.kubedb.com
  redisversions.catalog.kubedb.com
  appbindings.appcatalog.appscode.com
)

echo "checking kubeconfig context"
kubectl config current-context || {
  echo "Set a context (kubectl use-context <context>) out of the following:"
  echo
  kubectl config get-contexts
  exit 1
}
echo ""

OS=""
ARCH=""
DOWNLOAD_URL=""
DOWNLOAD_DIR=""
TEMP_DIRS=()
ONESSL=""

GOPATH=$(go env GOPATH)
REPO_ROOT=`git rev-parse --show-toplevel`
INSTALLER_ROOT="$GOPATH/src/kubedb.dev/installer"

pushd $REPO_ROOT

# http://redsymbol.net/articles/bash-exit-traps/
function cleanup() {
  rm -rf ca.crt ca.key server.crt server.key
  # remove temporary directories
  for dir in "${TEMP_DIRS[@]}"; do
    rm -rf "${dir}"
  done
}
trap cleanup EXIT

# detect operating system
# ref: https://raw.githubusercontent.com/helm/helm/master/scripts/get
function detectOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    cygwin* | mingw* | msys*) OS='windows';;
  esac
}

# detect machine architecture
function detectArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

detectOS
detectArch

# download file pointed by DOWNLOAD_URL variable
# store download file to the directory pointed by DOWNLOAD_DIR variable
# you have to sent the output file name as argument. i.e. downloadFile myfile.tar.gz
function downloadFile() {
  if curl --output /dev/null --silent --head --fail "$DOWNLOAD_URL"; then
    curl -fsSL ${DOWNLOAD_URL} -o $DOWNLOAD_DIR/$1
  else
    echo "File does not exist"
    exit 1
  fi
}

onessl_found() {
  # https://stackoverflow.com/a/677212/244009
  if [ -x "$(command -v onessl)" ]; then
    onessl wait-until-has -h >/dev/null 2>&1 || {
      # old version of onessl found
      echo "Found outdated onessl"
      return 1
    }
    export ONESSL=onessl
    return 0
  fi
  return 1
}

# download onessl if it does not exist
onessl_found || {
  echo "Downloading onessl ..."

  ARTIFACT="https://github.com/kubepack/onessl/releases/download/0.12.0"
  ONESSL_BIN=onessl-${OS}-${ARCH}
  case "$OS" in
    cygwin* | mingw* | msys*)
      ONESSL_BIN=${ONESSL_BIN}.exe
    ;;
  esac

  DOWNLOAD_URL=${ARTIFACT}/${ONESSL_BIN}
  DOWNLOAD_DIR="$(mktemp -dt onessl-XXXXXX)"
  TEMP_DIRS+=($DOWNLOAD_DIR) # store DOWNLOAD_DIR to cleanup later

  downloadFile $ONESSL_BIN # downloaded file name will be saved as the value of ONESSL_BIN variable

  export ONESSL=${DOWNLOAD_DIR}/${ONESSL_BIN}
  chmod +x $ONESSL
}

source "$REPO_ROOT/hack/deploy/settings"

echo "waiting until kubedb crds are ready"
for crd in "${crds[@]}"; do
  $ONESSL wait-until-ready crd ${crd} || {
    echo "$crd crd failed to be ready"
    exit 1
  }
done

cat $INSTALLER_ROOT/deploy/kubedb-catalog/* | $ONESSL envsubst | kubectl apply -f - || true
