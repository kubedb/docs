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

# http://redsymbol.net/articles/bash-exit-traps/
function cleanup() {
  rm -rf $ONESSL ca.crt ca.key server.crt server.key
}
trap cleanup EXIT


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

onessl_found || {
  echo "Downloading onessl ..."
  # ref: https://stackoverflow.com/a/27776822/244009
  case "$(uname -s)" in
    Darwin)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-darwin-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    Linux)
      curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-linux-amd64
      chmod +x onessl
      export ONESSL=./onessl
      ;;

    CYGWIN* | MINGW* | MSYS*)
      curl -fsSL -o onessl.exe https://github.com/kubepack/onessl/releases/download/0.10.0/onessl-windows-amd64.exe
      chmod +x onessl.exe
      export ONESSL=./onessl.exe
      ;;
    *)
      echo 'other OS'
      ;;
  esac
}

GOPATH=$(go env GOPATH)
REPO_ROOT=`git rev-parse --show-toplevel`
INSTALLER_ROOT="$GOPATH/src/github.com/kubedb/installer"

source "$REPO_ROOT/hack/deploy/settings"

echo "waiting until kubedb crds are ready"
for crd in "${crds[@]}"; do
  $ONESSL wait-until-ready crd ${crd} || {
    echo "$crd crd failed to be ready"
    exit 1
  }
done

cat $INSTALLER_ROOT/deploy/kubedb-catalog/* | $ONESSL envsubst | kubectl apply -f - || true
