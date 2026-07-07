---
title: Volume Expansion HanaDB
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-volume-expansion-volume-expansion
    name: Volume Expansion
    parent: guides-hanadb-volume-expansion
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Volume Expansion of HanaDB

This guide shows how to grow the data volumes of a HanaDB using a `HanaDBOpsRequest` of type
`VolumeExpansion`.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/volume-expansion](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/volume-expansion) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- **Volume expansion requires a `StorageClass` that supports expansion** (`allowVolumeExpansion: true`).
  Verify this before you start:

```bash
kubectl get storageclass
```
NAME                       PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)       rancher.io/local-path   Delete          WaitForFirstConsumer   false                  19d
longhorn-single            driver.longhorn.io      Delete          Immediate              true                   1h

> The `local-path` provisioner used in the other guides does **not** support volume expansion
> (`ALLOWVOLUMEEXPANSION` is `false`). For this guide, deploy the database on an expansion-capable
> StorageClass such as [Longhorn](https://longhorn.io/) (`longhorn-single` above).

## Deploy a HanaDB on an Expandable StorageClass

This guide uses a System Replication cluster on `longhorn-single`; the same `VolumeExpansion`
`HanaDBOpsRequest` works for a standalone database as well.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: Durable
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  podTemplate:
    spec:
      containers:
      - name: hanadb
        resources:
          requests:
            cpu: "1500m"
            memory: "8Gi"
          limits:
            cpu: "4"
            memory: "14Gi"
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
    storageClassName: longhorn-single
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/volume-expansion/system-replication-ops.yaml
```
hanadb.kubedb.com/hanadb-cluster created

Wait until `hanadb-cluster` is `Ready`, then check the current PVC sizes:

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=hanadb-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
```
NAME                            SC                SIZE
data-hanadb-cluster-0           longhorn-single   64Gi
data-hanadb-cluster-1           longhorn-single   64Gi
data-hanadb-cluster-arbiter-0   longhorn-single   2Gi

## Create a VolumeExpansion HanaDBOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: hanadb-cluster
  volumeExpansion:
    mode: Online
    hanadb: 65Gi
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/volume-expansion/volume-expansion.yaml
```
hanadbopsrequest.ops.kubedb.com/hdbops-volume-expansion created

Here,

- `spec.volumeExpansion.hanadb` is the new size of each HANA data volume.
- `spec.volumeExpansion.mode` is `Online` (expand in place) or `Offline` (recreate pods). `Online`
  requires the StorageClass/CSI driver to support online expansion.

## Verify the Expansion

```bash
kubectl get hdbops -n demo hdbops-volume-expansion
```
NAME                      TYPE              STATUS       AGE
hdbops-volume-expansion   VolumeExpansion   Successful   2m36s

```bash
kubectl describe hdbops -n demo hdbops-volume-expansion
```
...
Status:
  Conditions:
    Message:  HanaDBOpsRequest has started to expand volume of HanaDB nodes.
    Reason:   VolumeExpansion
    Status:   True
    Type:     VolumeExpansion
    Message:  Successfully paused database
    Reason:   DatabasePauseSucceeded
    Status:   True
    Type:     DatabasePauseSucceeded
    Message:  PetSet is recreated
    Reason:   ReadyPetSets
    Status:   True
    Type:     ReadyPetSets
    Message:  Successfully completed volume expansion for HanaDB.
    Reason:   Successful
    Status:   True
    Type:     Successful
  Phase:      Successful

Confirm the data PVCs grew to the requested size (the small arbiter volume is unchanged):

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=hanadb-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
```
NAME                            SC                SIZE
data-hanadb-cluster-0           longhorn-single   65Gi
data-hanadb-cluster-1           longhorn-single   65Gi
data-hanadb-cluster-arbiter-0   longhorn-single   2Gi

## Cleaning Up

```bash
kubectl delete hdbops -n demo hdbops-volume-expansion
```

```bash
kubectl delete hanadb.kubedb.com -n demo hanadb-cluster
```

```bash
kubectl delete ns demo
```

## Next Steps

- Move data to a different StorageClass with [Storage Migration](/docs/guides/hanadb/storage-migration/storage-migration.md).
- Review the [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md).
