#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)

REPO_ROOT="$GOPATH/src/github.com/kubedb/operator"
CLI_ROOT="$GOPATH/src/github.com/kubedb/cli"

pushd $REPO_ROOT

source "$REPO_ROOT/hack/deploy/settings"
source "$REPO_ROOT/hack/libbuild/common/lib.sh"

export APPSCODE_ENV=${APPSCODE_ENV:-prod}
export KUBEDB_SCRIPT="curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.8.0-beta.3/"

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
echo "${KUBEDB_SCRIPT}hack/deploy/kubedb.sh | bash -s -- "$@""
${KUBEDB_SCRIPT}hack/deploy/kubedb.sh | bash -s -- "$@"

if [ `pwd` = "$CLI_ROOT" ]; then
    popd
fi
popd
