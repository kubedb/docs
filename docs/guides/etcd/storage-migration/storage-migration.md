---
title: Etcd Storage Migration
menu:
  docs_{{ .version }}:
    identifier: etcd-storage-migration-ops
    name: Migrate StorageClass
    parent: etcd-storage-migration
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Migration of an Etcd Cluster

This guide shows how to move an `Etcd` cluster's data volumes to a different `StorageClass` using an
`EtcdOpsRequest` of type `StorageMigration`. KubeDB provisions new PVCs on the target `StorageClass`,
copies the data across with a migrator `Job`, swaps the volumes in and recreates each member pod —
**followers first, the Raft leader last**, waiting for a healthy quorum between members.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
  configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps
  [here](/docs/setup/README.md). Etcd support is an **alpha** feature, so the operators must be
  installed with the `Etcd` feature gate turned on (`--set featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Storage Migration Overview](/docs/guides/etcd/storage-migration/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd/storage-migration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/storage-migration)
> directory of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Prepare Source and Target StorageClasses

You need a **source** and a **target** `StorageClass`. This guide migrates between two
[Longhorn](https://longhorn.io/) StorageClasses, `longhorn-single` and `longhorn-single-migrated`:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: longhorn-single-migrated
provisioner: driver.longhorn.io
allowVolumeExpansion: true
reclaimPolicy: Delete
volumeBindingMode: Immediate
parameters:
  numberOfReplicas: "1"
  staleReplicaTimeout: "30"
  fsType: ext4
  dataLocality: disabled
  dataEngine: v1
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/storage-migration/longhorn-single.yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/storage-migration/longhorn-single-migrated.yaml
```

> If the **source** `StorageClass` uses `volumeBindingMode: WaitForFirstConsumer`, the target must use
> it too — the validating webhook rejects the mismatch, because a migration that changes the binding
> mode cannot guarantee the new volume lands in the right topology.

## Deploy an Etcd Cluster on the Source StorageClass

Below is the YAML of the `Etcd` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-cluster
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "longhorn-single"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/storage-migration/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Wait until `etcd-cluster` is `Ready`, then note the source `StorageClass` of the PVCs:

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    5m

$ kubectl get pvc -n demo -l app.kubernetes.io/instance=etcd-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
NAME                  SC                SIZE
data-etcd-cluster-0   longhorn-single   2Gi
data-etcd-cluster-1   longhorn-single   2Gi
data-etcd-cluster-2   longhorn-single   2Gi
```

## Create a StorageMigration EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: etcd-cluster
  migration:
    storageClassName: longhorn-single-migrated
  timeout: 30m
  apply: IfReady
```

Here,

- `spec.migration.storageClassName` is the **target** `StorageClass`.
- `spec.timeout` is **required** for `StorageMigration`. Every step is bounded by it, and the
  data-copy `Job` has no other deadline — so size it against how much data each member holds, not
  against how long a pod takes to restart.
- `spec.migration.oldPVReclaimPolicy` (optional) controls what happens to the source PVs once their
  claims have been renamed away. Set it to `Retain` if you want to keep the old volumes.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/storage-migration/etcdops-storage-migration.yaml
etcdopsrequest.ops.kubedb.com/etcd-storage-migration created
```

## Verify the Migration

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME                     TYPE               STATUS       AGE
etcd-storage-migration   StorageMigration   Successful   14m
```

While it runs you can watch the members being migrated one at a time. The `PetSet` is deleted up
front (with orphaned pods, so the members keep serving), and then each member's pod goes away and
comes back on the new volume:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
NAME                            READY   STATUS      RESTARTS   AGE
etcd-cluster-0                  1/1     Running     0          21m
etcd-cluster-2                  1/1     Running     0          20m
migrator-etcd-cluster-1-xxxxx   0/1     Completed   0          38s
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-storage-migration
Name:         etcd-storage-migration
Namespace:    demo
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-08-15T12:15:07Z
  Generation:          1
  Resource Version:    158773
  UID:                 c40b7c19-6d33-4f2e-b9b7-8f8f2b8a4d10
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  etcd-cluster
  Migration:
    Storage Class Name:  longhorn-single-migrated
  Timeout:               30m
  Type:                  StorageMigration
Status:
  Conditions:
    Last Transition Time:  2026-08-15T12:15:07Z
    Message:               StorageClass migration is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-08-15T12:15:22Z
    Message:               pet set deleted; ConditionStatus:True; PodName:etcd-cluster
    Observed Generation:   1
    Status:                True
    Type:                  PetSetDeleted--etcd-cluster
    Last Transition Time:  2026-08-15T12:19:47Z
    Message:               PVC migration completed for etcd-cluster-1
    Observed Generation:   1
    Reason:                PodMigrationCompleted-etcd-cluster-1
    Status:                True
    Type:                  PodMigrationCompleted-etcd-cluster-1
    Last Transition Time:  2026-08-15T12:24:12Z
    Message:               PVC migration completed for etcd-cluster-2
    Observed Generation:   1
    Reason:                PodMigrationCompleted-etcd-cluster-2
    Status:                True
    Type:                  PodMigrationCompleted-etcd-cluster-2
    Last Transition Time:  2026-08-15T12:28:31Z
    Message:               PVC migration completed for etcd-cluster-0
    Observed Generation:   1
    Reason:                PodMigrationCompleted-etcd-cluster-0
    Status:                True
    Type:                  PodMigrationCompleted-etcd-cluster-0
    Last Transition Time:  2026-08-15T12:28:36Z
    Message:               Successfully migrated the StorageClass of every etcd member
    Observed Generation:   1
    Reason:                MigrateEtcdStorage
    Status:                True
    Type:                  MigrateEtcdStorage
    Last Transition Time:  2026-08-15T12:28:41Z
    Message:               Successfully updated the Etcd storage class
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-08-15T12:28:44Z
    Message:               Successfully migrated the etcd StorageClass
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason              Age  From                         Message
  ----    ------              ---- ----                         -------
  Normal  PauseDatabase       14m  KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-cluster
  Normal  MigrateEtcdStorage  20s  KubeDB Ops-manager Operator  Successfully migrated the StorageClass of every etcd member
  Normal  UpdateDatabase      15s  KubeDB Ops-manager Operator  Successfully updated the Etcd storage class
  Normal  ResumeDatabase      12s  KubeDB Ops-manager Operator  Successfully resumed Etcd demo/etcd-cluster
  Normal  Successful          12s  KubeDB Ops-manager Operator  Successfully migrated the etcd StorageClass
```

Note the order of the `PodMigrationCompleted-*` conditions: the followers (`etcd-cluster-1`,
`etcd-cluster-2` in this run) complete before the Raft leader (`etcd-cluster-0` here). Which pod is
the leader is discovered at runtime, so the order you observe depends on where leadership sits when
the migration starts. Between two members, the operator waits for the restarted one to be `Ready` and
for the cluster to report a healthy quorum — that gate is what keeps a 3-member cluster from ever
having two members down at once.

The per-resource progress conditions (`PVCCreated--…`, `PodDeleted--…`, `JobCreated--…`,
`PVCDeleted--…`, `PodReady--…`, and so on) are elided above. They are written for every step the
migration completes, not only for steps that had to wait, and are useful when a migration stalls.

Confirm the data PVCs are now bound to the target `StorageClass` and the cluster is `Ready`:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=etcd-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
NAME                  SC                         SIZE
data-etcd-cluster-0   longhorn-single-migrated   2Gi
data-etcd-cluster-1   longhorn-single-migrated   2Gi
data-etcd-cluster-2   longhorn-single-migrated   2Gi

$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.storage.storageClassName}'
longhorn-single-migrated

$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    43m
```

And that etcd itself still has all three voting members with a healthy quorum:

```bash
$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 endpoint health --cluster -w table
+---------------------------------------------------------------------+--------+-------------+-------+
|                              ENDPOINT                               | HEALTH |    TOOK     | ERROR |
+---------------------------------------------------------------------+--------+-------------+-------+
| http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 |   true | 4.512103ms  |       |
| http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 |   true | 5.284917ms  |       |
| http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 |   true | 5.901442ms  |       |
+---------------------------------------------------------------------+--------+-------------+-------+
```

## Troubleshooting

- **`spec.timeout is required for a storage migration`** — add `spec.timeout`. It is not optional for
  this ops type.
- **`storage class <name> not found`** — the target `StorageClass` must exist before the request is
  created.
- **`volume binding mode should be WaitForFirstConsumer for <name> storageClass`** — the source class
  binds late and the target does not. Pick a target with a matching `volumeBindingMode`.
- **`storage migration is not applicable to an ephemeral etcd`** — `spec.storageType` is `Ephemeral`;
  there are no PVCs to migrate.
- **The migration stalls on one member** — look at the migrator `Job`
  (`kubectl logs -n demo job/migrator-<pod-name>`); it is the `rsync` doing the copying. Bear in mind
  the whole request, including that copy, is bounded by `spec.timeout`.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-storage-migration
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
# the StorageClasses are cluster-scoped, so deleting the namespace does not remove them
kubectl delete storageclass longhorn-single longhorn-single-migrated
```

## Next Steps

- [Expand the volume](/docs/guides/etcd/volume-expansion/volume-expansion.md) of an Etcd cluster.
- Scale the cluster [vertically](/docs/guides/etcd/scaling/vertical-scaling/vertical-scaling.md) or
  [horizontally](/docs/guides/etcd/scaling/horizontal-scaling/horizontal-scaling.md).
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
