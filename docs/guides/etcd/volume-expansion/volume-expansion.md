---
title: Etcd Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: etcd-volume-expansion-ops
    name: Expand Storage Volume
    parent: etcd-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Volume Expansion of an Etcd Cluster

This guide will show you how to use the `KubeDB` Ops-manager operator to expand the data volumes of
an `Etcd` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
  configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps
  [here](/docs/setup/README.md). Etcd support is an **alpha** feature, so the operators must be
  installed with the `Etcd` feature gate turned on (`--set featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Volume Expansion Overview](/docs/guides/etcd/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd/volume-expansion](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/volume-expansion)
> directory of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Expand Volume of an Etcd Cluster

### Prepare the Etcd Cluster

At first, verify that your cluster has a `StorageClass` that supports volume expansion:

```bash
$ kubectl get storageclass
NAME                 PROVISIONER          RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   driver.standard.io   Delete          Immediate           true                   93s
```

`standard` has `ALLOWVOLUMEEXPANSION: true`, so we can use it.

#### Deploy Etcd

Below is the YAML of the `Etcd` CR that we are going to create, with 1GB volumes:

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
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Etcd` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/volume-expansion/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Now, wait until `etcd-cluster` has status `Ready`. i.e,

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    4m
```

Let's check the volume size from the `PetSet` and from the PVCs:

```bash
$ kubectl get petset -n demo etcd-cluster \
  -o jsonpath='{.spec.volumeClaimTemplates[0].spec.resources.requests.storage}'
1Gi

$ kubectl get pvc -n demo -l app.kubernetes.io/instance=etcd-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
NAME                  SC         SIZE
data-etcd-cluster-0   standard   1Gi
data-etcd-cluster-1   standard   1Gi
data-etcd-cluster-2   standard   1Gi
```

We are now ready to apply an `EtcdOpsRequest` to expand these volumes.

### Online Volume Expansion

`Online` mode grows the volumes underneath the running member pods. Nothing is restarted, so the
cluster keeps serving and quorum is never at risk.

#### Create EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-volume-exp-online
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: etcd-cluster
  volumeExpansion:
    mode: Online
    etcd: 2Gi
  timeout: 20m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are expanding the volumes of the `etcd-cluster` database.
- `spec.type` specifies that we are performing `VolumeExpansion`.
- `spec.volumeExpansion.etcd` is the desired size of each member's data volume. It must be larger
  than the current request — the webhook rejects a shrink.
- `spec.volumeExpansion.mode` is **required**: `Online` or `Offline`. See
  [Volume Expansion Modes](/docs/guides/etcd/volume-expansion/overview.md#volume-expansion-modes).

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/volume-expansion/etcdops-volume-exp-online.yaml
etcdopsrequest.ops.kubedb.com/etcd-volume-exp-online created
```

#### Verify the Volumes Expanded Successfully

Let's wait for the `EtcdOpsRequest` to become `Successful`:

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME                     TYPE              STATUS       AGE
etcd-volume-exp-online   VolumeExpansion   Successful   4m11s
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-volume-exp-online
Name:         etcd-volume-exp-online
Namespace:    demo
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-08-15T11:04:31Z
  Generation:          1
  Resource Version:    143901
  UID:                 9b2e73aa-1c05-4f8a-8c53-2e0d6c4a91b2
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   etcd-cluster
  Timeout:  20m
  Type:     VolumeExpansion
  Volume Expansion:
    Etcd:  2Gi
    Mode:  Online
Status:
  Conditions:
    Last Transition Time:  2026-08-15T11:04:31Z
    Message:               Volume Expansion is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-08-15T11:04:41Z
    Message:               3
    Observed Generation:   1
    Reason:                PetSetReplicasBeforeExpansion
    Status:                True
    Type:                  PetSetReplicasBeforeExpansion
    Last Transition Time:  2026-08-15T11:07:26Z
    Message:               is p v c expanded; ConditionStatus:True; PodName:data-etcd-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  IsPVCExpanded--data-etcd-cluster-1
    Last Transition Time:  2026-08-15T11:07:41Z
    Message:               is pet set deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetDeleted
    Last Transition Time:  2026-08-15T11:07:46Z
    Message:               Successfully expanded the etcd data volumes
    Observed Generation:   1
    Reason:                UpdateEtcdNodePVCs
    Status:                True
    Type:                  UpdateEtcdNodePVCs
    Last Transition Time:  2026-08-15T11:07:51Z
    Message:               Successfully updated the Etcd storage request
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-08-15T11:08:01Z
    Message:               PetSet has been recreated with the expanded volume claim template
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2026-08-15T11:08:03Z
    Message:               Successfully expanded the etcd data volumes
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age    From                         Message
  ----     ------                                    ----   ----                         -------
  Normal   PauseDatabase                             4m11s  KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-cluster
  Warning  is p v c expanded; ConditionStatus:False    3m56s  KubeDB Ops-manager Operator  is p v c expanded; ConditionStatus:False
  Warning  is p v c expanded; ConditionStatus:True     76s    KubeDB Ops-manager Operator  is p v c expanded; ConditionStatus:True
  Warning  is pet set deleted; ConditionStatus:True  61s    KubeDB Ops-manager Operator  is pet set deleted; ConditionStatus:True
  Normal   UpdateEtcdNodePVCs                        56s    KubeDB Ops-manager Operator  Successfully expanded the etcd data volumes
  Normal   UpdateDatabase                            51s    KubeDB Ops-manager Operator  Successfully updated the Etcd storage request
  Normal   ReadyPetSets                              41s    KubeDB Ops-manager Operator  PetSet has been recreated with the expanded volume claim template
  Normal   ResumeDatabase                            39s    KubeDB Ops-manager Operator  Successfully resumed Etcd demo/etcd-cluster
  Normal   Successful                                39s    KubeDB Ops-manager Operator  Successfully expanded the etcd data volumes
```

Two conditions in there are worth explaining:

- `PetSetReplicasBeforeExpansion` carries the member count (`3`) as its **message**. The operator has
  to delete and recreate the `PetSet` (`spec.volumeClaimTemplates` is immutable), and a newly created
  `PetSet` is seeded with a single replica, so the member count is stashed here and restored right
  after the recreation while the database is still paused.
- `IsPVCExpanded--<pvc>` is the gate on the CSI driver. KubeDB patches
  `spec.resources.requests.storage` on each PVC and then waits until `status.capacity` catches up; a
  driver that never grows the volume is what makes this condition stay `False`.

Now let's verify from the `PetSet` and the PVCs that the volumes have been expanded:

```bash
$ kubectl get petset -n demo etcd-cluster \
  -o jsonpath='{.spec.volumeClaimTemplates[0].spec.resources.requests.storage}'
2Gi

$ kubectl get pvc -n demo -l app.kubernetes.io/instance=etcd-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
NAME                  SC         SIZE
data-etcd-cluster-0   standard   2Gi
data-etcd-cluster-1   standard   2Gi
data-etcd-cluster-2   standard   2Gi

$ kubectl get etcd -n demo etcd-cluster \
  -o jsonpath='{.spec.storage.resources.requests.storage}'
2Gi
```

Because this was an `Online` expansion, the member pods were never deleted — check their age against
the ops request's:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
NAME             READY   STATUS    RESTARTS   AGE
etcd-cluster-0   1/1     Running   0          18m
etcd-cluster-1   1/1     Running   0          17m
etcd-cluster-2   1/1     Running   0          17m
```

### Offline Volume Expansion

If your CSI driver cannot grow a mounted volume, use `Offline` mode. The operator scales the `PetSet`
to `0` and waits for every member pod to terminate before it touches the PVCs — so the **cluster is
fully unavailable for the duration**. etcd itself is fine with this: the Raft log and the snapshot
live on the very volumes being expanded, so a full stop loses no data.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-volume-exp-offline
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: etcd-cluster
  volumeExpansion:
    mode: Offline
    etcd: 3Gi
  timeout: 20m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/volume-expansion/etcdops-volume-exp-offline.yaml
etcdopsrequest.ops.kubedb.com/etcd-volume-exp-offline created
```

The condition list is the same as for `Online`, with one extra gate at the beginning reflecting the
shutdown — the operator scales the `PetSet` to zero and then waits for every member pod to go away:

```yaml
    Message:  are pods deleted; ConditionStatus:True
    Status:   True
    Type:     ArePodsDeleted
```

> A `ScalePetSet` condition exists too, but only ever appears with `Status: False`, and only if
> reading or patching the `PetSet` failed. On a clean run you will not see it at all.

Verify the same way as above; the member pods will be new (their `AGE` resets), because they were
terminated and recreated.

## Troubleshooting

- **`volume expansion is not applicable to an ephemeral etcd`** — `spec.storageType` is `Ephemeral`;
  there is no PVC to expand.
- **`IsPVCExpanded--<pvc>` stays `False`** — the `StorageClass` does not have
  `allowVolumeExpansion: true`, or the CSI driver cannot grow a mounted volume. In the latter case,
  retry with `mode: Offline`.
- **The webhook rejects the request** — you cannot shrink a volume, and
  `spec.volumeExpansion.etcd` may not be empty.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-volume-exp-online etcd-volume-exp-offline
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- Move the volumes to a different `StorageClass` with
  [Storage Migration](/docs/guides/etcd/storage-migration/storage-migration.md).
- Scale the cluster [vertically](/docs/guides/etcd/scaling/vertical-scaling/vertical-scaling.md) or
  [horizontally](/docs/guides/etcd/scaling/horizontal-scaling/horizontal-scaling.md).
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
