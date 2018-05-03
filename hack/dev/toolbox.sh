#!/bin/bash
set -eou pipefail

crds=(mongodbs mysqls memcacheds redises postgreses elasticsearches snapshots dormantdatabases)

export KUBEDB_UNINSTALL=0
export KUBEDB_PURGE=0
export KUBEDB_NAMESPACE=kube-system

show_help() {
    echo "toolbox.sh - deb tool for kubedb operator"
    echo " "
    echo "toolbox.sh [options]"
    echo " "
    echo "options:"
    echo "-h, --help                         show brief help"
    echo "    --uninstall                    uninstall kubedb"
    echo "    --purge                        purges kubedb crd objects and crds"
}

while test $# -gt 0; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        --namespace*)
            export KUBEDB_NAMESPACE=`echo $1 | sed -e 's/^[^=]*=//g'`
            shift
            ;;
        --uninstall)
            export KUBEDB_UNINSTALL=1
            shift
            ;;
        --purge)
            export KUBEDB_PURGE=1
            shift
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
done

if [ "$KUBEDB_UNINSTALL" -eq 1 ]; then
# delete webhooks and apiservices
    kubectl delete validatingwebhookconfiguration -l app=kubedb || true
    kubectl delete mutatingwebhookconfiguration -l app=kubedb || true
    kubectl delete apiservice -l app=kubedb
    # delete kubedb operator
    kubectl delete deployment -l app=kubedb --namespace $KUBEDB_NAMESPACE
    kubectl delete service -l app=kubedb --namespace $KUBEDB_NAMESPACE
    kubectl delete endpoints -l app=kubedb --namespace $KUBEDB_NAMESPACE
    kubectl delete secret -l app=kubedb --namespace $KUBEDB_NAMESPACE
    # delete RBAC objects, if --rbac flag was used.
    kubectl delete serviceaccount -l app=kubedb --namespace $KUBEDB_NAMESPACE
    kubectl delete clusterrolebindings -l app=kubedb
    kubectl delete clusterrole -l app=kubedb
    kubectl delete rolebindings -l app=kubedb --namespace $KUBEDB_NAMESPACE
    kubectl delete role -l app=kubedb --namespace $KUBEDB_NAMESPACE

    # https://github.com/kubernetes/kubernetes/issues/60538
    if [ "$KUBEDB_PURGE" -eq 1 ]; then
        for crd in "${crds[@]}"; do
            pairs=($(kubectl get ${crd}.kubedb.com --all-namespaces -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.namespace} {end}' || true))
            total=${#pairs[*]}

            for (( i=0; i<$total; i+=2 )); do
                name=${pairs[$i]}
                namespace=${pairs[$i + 1]}
                # remove finalizers
                kubectl patch ${crd}.kubedb.com $name -n $namespace  -p '{"metadata":{"finalizers":[]}}' --type=merge
                # delete crd object
                echo "deleting ${crd} $namespace/$name"
                kubectl delete ${crd}.kubedb.com $name -n $namespace || true
            done

            # delete crd
            kubectl delete crd ${crd}.kubedb.com || true
        done
    fi

    echo
    echo "Successfully Cleaned KubeDB Stuffs!"
    exit 0
fi
