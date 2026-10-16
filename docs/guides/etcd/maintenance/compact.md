---
title: Compact Etcd Keyspace History
menu:
  docs_{{ .version }}:
    identifier: etcd-maintenance-compact
    name: Compact
    parent: etcd-maintenance
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Compact the Keyspace History of an Etcd Cluster

KubeDB supports compacting the keyspace history of an etcd cluster through an `EtcdOpsRequest` of
type `Compact`. Compaction discards superseded versions of keys, which is what stops a busy
cluster's MVCC history from growing without bound.

For the background on what a revision is and what compaction throws away, see the
[Maintenance Operations Overview](/docs/guides/etcd/maintenance/overview.md#compaction).

> **Important:** compaction on its own does **not** shrink anything on disk. It makes the space
> reusable inside the backend file; returning it to the filesystem requires a
> [Defragment](/docs/guides/etcd/maintenance/defragment.md) afterwards. Plan for the pair.

## Before You Begin

- At first, you need a Kubernetes cluster, and the `kubectl` command-line tool must be configured
  to communicate with your cluster. If you do not already have a cluster, you can create one
  using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the `KubeDB` Provisioner and Ops-manager operators in your cluster following the steps
  [here](/docs/setup/README.md). etcd support is behind an alpha feature gate, so make sure the
  operators are installed with `Etcd=true` enabled (Helm value `featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Maintenance Operations Overview](/docs/guides/etcd/maintenance/overview.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd)
> folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Etcd

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Wait until the cluster reports `Ready`,

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    2m
```

## Apply the Compact OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-compact
  namespace: demo
spec:
  type: Compact
  databaseRef:
    name: etcd-cluster
  compact: {}
  timeout: 5m
```

- `spec.type` specifies the type of the ops request, here `Compact`.
- `spec.databaseRef` holds the name of the `Etcd` object. It must live in the same namespace as
  the ops request.
- `spec.compact` is **required** for this type, even when it is empty — a `Compact` request
  without a `spec.compact` block is rejected by the validating webhook. Write `compact: {}` when
  you have nothing to configure.
- `spec.compact.revision` is optional. Omitting it, as above, means "compact up to the current
  revision at execution time": the operator reads the cluster's current revision when the request
  runs and compacts to that. This is what you want for the routine "reclaim everything that is
  safe to reclaim right now" case.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/compact.yaml
etcdopsrequest.ops.kubedb.com/etcd-compact created
```

### Compacting to a specific revision

If you have consumers that read or watch from a known past revision, compacting to "now" would
break them — a read below the compacted revision fails with
`mvcc: required revision has been compacted`. In that case, name the revision explicitly:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-compact-revision
  namespace: demo
spec:
  type: Compact
  databaseRef:
    name: etcd-cluster
  compact:
    revision: 128000
  timeout: 5m
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/compact-revision.yaml
etcdopsrequest.ops.kubedb.com/etcd-compact-revision created
```

The webhook rejects a negative `revision`. You can find the cluster's current revision — a useful
starting point for choosing one — from any member:

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- etcdctl \
    --endpoints=http://etcd-cluster-0.etcd-cluster-pods.demo.svc:2379 \
    endpoint status -w table
```

> If the cluster has TLS or authentication enabled, add the corresponding `etcdctl` flags
> (`--cacert`/`--cert`/`--key`, and `--user`) and use the `https` scheme.

## What the operator does

`Compact` never pauses the `Etcd` object, never patches the PetSet and never restarts a pod.

Unlike defragmentation, compaction is a **cluster-wide** operation: the request goes through Raft
and is replicated to every member, so the operator issues it exactly **once**, against any member
it can reach — there is no per-member walk.

The steps are:

1. If `spec.compact.revision` is unset (or not positive), read the cluster's current revision
   from the header of a linearizable response, and use that.
2. Call etcd's `Compact` RPC for that revision, with the physical option set — meaning the call
   only returns once the compaction has actually been applied to the backend, not merely
   accepted.
3. Treat "already compacted" as success. If the requested revision has already been compacted by
   an earlier request or by automatic compaction, the cluster is already in the state you asked
   for, so the operator records the step as done rather than failing.

## Progress and status

```bash
$ kubectl get etcdopsrequest -n demo
NAME           TYPE      STATUS       AGE
etcd-compact   Compact   Successful   40s
```

```bash
$ kubectl describe etcdopsrequest -n demo etcd-compact
```

The full object looks like this once it has finished:

```yaml
$ kubectl get etcdopsrequest -n demo etcd-compact -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  creationTimestamp: "2026-02-11T12:02:41Z"
  generation: 1
  name: etcd-compact
  namespace: demo
  resourceVersion: "68120"
  uid: 4d1c9a77-6b3e-42fa-9c05-8f1a2b3c4d66
spec:
  apply: IfReady
  compact: {}
  databaseRef:
    name: etcd-cluster
  maxRetries: 1
  timeout: 5m
  type: Compact
status:
  conditions:
  - lastTransitionTime: "2026-02-11T12:02:41Z"
    message: Compacting the etcd keyspace history
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2026-02-11T12:02:49Z"
    message: Successfully compacted the etcd keyspace history
    observedGeneration: 1
    reason: EtcdCompacted
    status: "True"
    type: EtcdCompacted
  - lastTransitionTime: "2026-02-11T12:02:50Z"
    message: Successfully compacted the etcd keyspace history
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

The condition to look for is **`EtcdCompacted`**. Once it is `True`, the compaction has been
applied cluster-wide and the ops request is marked `Successful`.

## Reclaiming the disk space

At this point the old revisions are gone from the keyspace, but every member's backend file is
still exactly as large as it was. To actually return that space to the filesystem, follow up with
a defragmentation:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/defragment.yaml
etcdopsrequest.ops.kubedb.com/etcd-defragment created
```

Only one `EtcdOpsRequest` per database can be `Progressing` at a time, so wait for the `Compact`
request to reach `Successful` before applying the `Defragment` one. See
[Defragment the backends of an Etcd cluster](/docs/guides/etcd/maintenance/defragment.md) for the
full walkthrough.

## Automatic compaction

For the routine case you usually do not want to apply an ops request by hand at all. etcd can
compact on a schedule, and KubeDB exposes that on the `Etcd` object:

```yaml
spec:
  configuration:
    tuning:
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
```

`autoCompactionMode` is `periodic` or `revision`, and `autoCompactionRetention` is interpreted
accordingly — a duration of history to keep, or a number of revisions. These map directly onto
etcd's `--auto-compaction-mode` and `--auto-compaction-retention` flags. Changing them on an
existing cluster is a [Reconfigure](/docs/guides/etcd/reconfigure/reconfigure.md) operation.

With automatic compaction configured, keep the `Compact` ops request for the cases it is good at:
compacting right now ahead of a defragmentation, or compacting to a revision you have chosen
deliberately.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-compact etcd-compact-revision etcd-defragment
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- [Defragment the member backends of an Etcd cluster](/docs/guides/etcd/maintenance/defragment.md) —
  the follow-up that actually frees disk space.
- [Move the Raft leadership of an Etcd cluster](/docs/guides/etcd/maintenance/move-leader.md).
- Detail concepts of the [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
