#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=${GOPATH}/src/github.com/kubedb/elasticsearch

export DB_UPDATE=1
export TOOLS_UPDATE=1
export EXPORTER_UPDATE=1
export OPERATOR_UPDATE=1
export KIBANA_UPDATE=0

show_help() {
  echo "update-docker.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help                       show brief help"
  echo "    --db-only                    update only database images"
  echo "    --tools-only                 update only database-tools images"
  echo "    --operator-only              update only operator image"
  echo "    --exporter-only              update only database-exporter images"
  echo "    --kibana-only                update only kibana images"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      exit 0
      ;;
    --db-only)
      export DB_UPDATE=1
      export TOOLS_UPDATE=0
      export OPERATOR_UPDATE=0
      export EXPORTER_UPDATE=0
      shift
      ;;
    --tools-only)
      export DB_UPDATE=0
      export TOOLS_UPDATE=1
      export OPERATOR_UPDATE=0
      export EXPORTER_UPDATE=0
      shift
      ;;
    --exporter-only)
      export DB_UPDATE=0
      export TOOLS_UPDATE=0
      export EXPORTER_UPDATE=1
      export OPERATOR_UPDATE=0
      shift
      ;;
    --operator-only)
      export DB_UPDATE=0
      export TOOLS_UPDATE=0
      export EXPORTER_UPDATE=0
      export OPERATOR_UPDATE=1
      shift
      ;;
    --kibana-only)
      export DB_UPDATE=0
      export TOOLS_UPDATE=0
      export EXPORTER_UPDATE=0
      export OPERATOR_UPDATE=0
      export KIBANA_UPDATE=1
      shift
      ;;
    *)
      show_help
      exit 1
      ;;
  esac
done

dbversions=(
  5.6.4
  5.6
  6.2.4
  6.2
  6.3.0
  6.3
  6.4.0
  6.4
)

exporters=(
  1.0.2
)

kibanaimages=(
  6.3.0
)

echo ""
env | sort | grep -e DOCKER_REGISTRY -e APPSCODE_ENV || true
echo ""

if [ "$DB_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing database images" || true
  for db in "${dbversions[@]}"; do
    ${REPO_ROOT}/hack/docker/elasticsearch/${db}/make.sh build
    ${REPO_ROOT}/hack/docker/elasticsearch/${db}/make.sh push
  done
fi

if [ "$TOOLS_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing database-tools images" || true
  for db in "${dbversions[@]}"; do
    ${REPO_ROOT}/hack/docker/elasticsearch-tools/${db}/make.sh build
    ${REPO_ROOT}/hack/docker/elasticsearch-tools/${db}/make.sh push
  done
fi

if [ "$EXPORTER_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing database-exporter images" || true
  for exporter in "${exporters[@]}"; do
    ${REPO_ROOT}/hack/docker/elasticsearch_exporter/${exporter}/make.sh
  done
fi

if [ "$OPERATOR_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing Operator images" || true
  ${REPO_ROOT}/hack/docker/es-operator/make.sh build
  ${REPO_ROOT}/hack/docker/es-operator/make.sh push
fi

if [ "$KIBANA_UPDATE" -eq 1 ]; then
  cowsay -f tux "Processing Kibana images" || true
  for kibana in "${kibanaimages[@]}"; do
    ${REPO_ROOT}/hack/docker/kibana/${kibana}/make.sh build
    ${REPO_ROOT}/hack/docker/kibana/${kibana}/make.sh push
  done
fi
