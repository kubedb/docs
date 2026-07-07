---
title: Volume Expansion DocumentDB
menu:
  docs_{{ .version }}:
    identifier: dc-volume-expansion-details
    name: Volume Expansion
    parent: dc-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Volume Expansion of DocumentDB

When a `DocumentDB` database is provisioned on an **expandable** StorageClass you can grow its
data volumes in place with a `DocumentDBOpsRequest` of type `VolumeExpansion` — no
backup/restore and no manual PVC editing required. This guide expands a 3-node cluster from
`5Gi` to `10Gi` per replica.

> [!IMPORTANT]
> The StorageClass must allow volume expansion (`allowVolumeExpansion: true`). This guide uses
> `longhorn`, which is expandable. The default `local-path` StorageClass on many clusters is
> **not** expandable — check with `kubectl get sc` and use an expandable class.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo` (`kubectl create ns demo`).
- Deploy a `DocumentDB` cluster (`documentdb-cls-sample`) on an **expandable** StorageClass
  (`longhorn`) and wait for it to become `Ready`.

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## PVCs before

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=documentdb-cls-sample \
    -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName,STATUS:.status.phase
```
NAME                           SIZE   SC         STATUS
data-documentdb-cls-sample-0   5Gi    longhorn   Bound
data-documentdb-cls-sample-1   5Gi    longhorn   Bound
data-documentdb-cls-sample-2   5Gi    longhorn   Bound

## Create the VolumeExpansion OpsRequest

`mode: Offline` tells the operator to recreate the pods around the resize (the PetSet is deleted
and recreated so the larger PVCs are picked up cleanly). The `documentdb` field carries the new
target size:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: documentdb-cls-sample
  volumeExpansion:
    mode: Offline
    documentdb: 10Gi
```

```bash
kubectl apply -f cluster-volume-expansion.yaml
```
documentdbopsrequest.ops.kubedb.com/documentdb-cls-volume-expansion created

```bash
kubectl get dcops -n demo documentdb-cls-volume-expansion
```
NAME                              TYPE              STATUS       AGE
documentdb-cls-volume-expansion   VolumeExpansion   Successful   4m49s

The status conditions walk through the offline-expansion mechanics: the operator deletes the
PetSet, then for each replica it deletes the pod, expands the PVC, recreates the pod, and waits
for it to become ready, before finally recreating the PetSet:

```bash
kubectl get dcops -n demo documentdb-cls-volume-expansion \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
```
Running=True :: Volume Expansion is in progress
DeletePetset=True :: delete petset; ConditionStatus:True
IsPvcData-documentdb-cls-sample-0Updated=True :: is pvc data-documentdb-cls-sample-0 updated; ConditionStatus:True
CreatePod=True :: create pod; ConditionStatus:True
IsPodReady=True :: is pod ready; ConditionStatus:True
IsPvcData-documentdb-cls-sample-2Updated=True :: is pvc data-documentdb-cls-sample-2 updated; ConditionStatus:True
IsPvcData-documentdb-cls-sample-1Updated=True :: is pvc data-documentdb-cls-sample-1 updated; ConditionStatus:True
VolumeExpansion=True :: Offline Volume Expansion performed successfully in DocumentDB pods
ReadyPetSets=True :: PetSet is recreated
Successful=True :: Successfully Expanded Volume.

## PVCs after

All three data volumes are now `10Gi`, and the `DocumentDB` object's storage request is updated
to match:

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=documentdb-cls-sample \
    -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName,STATUS:.status.phase
```
NAME                           SIZE   SC         STATUS
data-documentdb-cls-sample-0   10Gi   longhorn   Bound
data-documentdb-cls-sample-1   10Gi   longhorn   Bound
data-documentdb-cls-sample-2   10Gi   longhorn   Bound

```bash
kubectl get docdb -n demo documentdb-cls-sample -o jsonpath='{.spec.storage.resources.requests.storage}'
```
10Gi

The cluster is healthy and serving traffic after the expansion:

```bash
PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo documentdb-cls-sample-0 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
```
{ ok: 1 }

## Standalone

The same `DocumentDBOpsRequest` works for a standalone (`replicas: 1`) instance on an expandable
StorageClass — point `spec.databaseRef.name` at `documentdb-sa-sample`. On this build standalone
instances did not finish bootstrapping (see the [Restart](/docs/guides/documentdb/restart/)
guide), so the standalone expansion could not be exercised live; the cluster procedure applies
once a standalone instance is healthy.

## Cleaning Up

```bash
kubectl delete documentdbopsrequest -n demo documentdb-cls-volume-expansion
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- [Storage migration](/docs/guides/documentdb/storage-migration/) to a different StorageClass.
- [Storage autoscaling](/docs/guides/documentdb/autoscaler/storage/) of a DocumentDB cluster.
