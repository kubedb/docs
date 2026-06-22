---
title: HanaDB StorageClass Migration
menu:
  docs_{{ .version }}:
    identifier: hanadb-migration-storageclass
    name: StorageClass Migration
    parent: hanadb-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# HanaDB StorageClass Migration

This guide shows how to migrate HanaDB persistent volumes from one `StorageClass` to another using a `HanaDBOpsRequest`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB by following the [setup guide](/docs/setup/README.md).
- Create a namespace:

```bash
kubectl create ns demo
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/migration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/migration).

## Install Longhorn

This example uses Longhorn as the source and target storage backend.

```bash
helm repo add longhorn https://charts.longhorn.io
helm repo update

helm install longhorn longhorn/longhorn \
  --namespace longhorn-system \
  --create-namespace
```

Wait until Longhorn pods are ready:

```bash
kubectl get pods -n longhorn-system
```

Create source and target StorageClasses:

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/migration/longhorn-single.yaml
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/migration/longhorn-single-migrated.yaml
```

## Deploy HanaDB

The Longhorn example includes an init container that fixes ownership of the fresh volume before SAP HANA starts.

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/migration/hanadb-cluster.yaml
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Apply Storage Migration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: hanadb-cluster
  migration:
    storageClassName: longhorn-single-migrated
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/migration/hdbops-storage-migration.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-storage-migration --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Verify

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=hanadb-cluster
kubectl get petsets.apps.k8s.appscode.com -n demo hanadb-cluster -o jsonpath='{.spec.volumeClaimTemplates[?(@.metadata.name=="data")].spec.storageClassName}'
```

## Cleanup

```bash
kubectl delete hdbops -n demo hdbops-storage-migration
kubectl delete hanadb -n demo hanadb-cluster
kubectl delete ns demo
```
