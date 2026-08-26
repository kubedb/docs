---
title: Defragment Etcd Backends
menu:
  docs_{{ .version }}:
    identifier: etcd-maintenance-defragment
    name: Defragment
    parent: etcd-maintenance
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Defragment the Backends of an Etcd Cluster

KubeDB supports defragmenting the backend database file of every etcd member through an
`EtcdOpsRequest` of type `Defragment`. Defragmentation is what actually returns disk space to the
operating system after keys have been deleted or the keyspace history has been compacted — etcd
does not do that on its own.

For the background on why compaction alone does not free disk, see the
[Maintenance Operations Overview](/docs/guides/etcd/maintenance/overview.md#defragmentation).

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

Before defragmenting, it is worth looking at the current backend sizes so that you have something
to compare against afterwards. `etcdctl endpoint status` reports each member's `DB SIZE` (the
size of the file on disk) and `DB SIZE IN USE` (the part of it that holds live data). A large gap
between the two is exactly what defragmentation removes.

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- etcdctl \
    --endpoints=http://etcd-cluster-0.etcd-cluster-pods.demo.svc:2379,http://etcd-cluster-1.etcd-cluster-pods.demo.svc:2379,http://etcd-cluster-2.etcd-cluster-pods.demo.svc:2379 \
    endpoint status -w table
```

> If the cluster has TLS or authentication enabled, add the corresponding `etcdctl` flags
> (`--cacert`/`--cert`/`--key`, and `--user`) and use the `https` scheme.

## Apply the Defragment OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-defragment
  namespace: demo
spec:
  type: Defragment
  databaseRef:
    name: etcd-cluster
  defragment: {}
```

- `spec.type` specifies the type of the ops request, here `Defragment`.
- `spec.databaseRef` holds the name of the `Etcd` object. It must live in the same namespace as
  the ops request.
- `spec.defragment` carries **no fields**. Defragmentation always applies to every member; there
  is nothing to select or configure. The empty object `{}` is all you write.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/maintenance/defragment.yaml
etcdopsrequest.ops.kubedb.com/etcd-defragment created
```

### A note on `spec.timeout`

`spec.timeout` is optional here, and it is worth understanding what it does before setting it. It
does **not** bound the request as a whole — it is the budget for each individual step, measured from
the moment that step first reported "not done yet". So on a three member cluster, each member's
defragmentation and each quorum re-check gets its own full `spec.timeout`.

When `spec.timeout` is omitted, each step gets a default budget of 10 minutes, and the `Defragment`
RPC issued against a single member is additionally capped at 10 minutes of its own. Unless you have
a specific reason — a backend large enough that one member takes longer than ten minutes to rewrite —
leaving `spec.timeout` unset is the safer choice.

## What the operator does

`Defragment` never pauses the `Etcd` object, never patches the PetSet and never restarts a pod.
It is a sequence of direct RPCs against the running members. The sequencing is the important
part, because defragmentation **blocks the member it runs against** for its whole duration:

1. **Resolve the leader.** For a multi-member cluster the operator first determines which pod
   currently holds the Raft leadership.
2. **Order the members, leader last.** Followers are visited in ordinal order, and the leader is
   appended at the end. This is the same reason a rolling restart visits the leader last: it
   keeps the cluster's write path available for as long as possible and avoids provoking an
   election in the middle of the run.
3. **Defragment one member at a time.** For each member in turn, the operator opens a client
   pinned to that member's own endpoint and calls `Defragment` against it. It never issues the
   call to more than one member concurrently.
4. **Gate on quorum between members.** After each member finishes, the operator re-checks that
   the cluster still has quorum before touching the next one. If it does not, the operator waits
   rather than proceeding — so a cluster that is already degraded does not get pushed over the
   edge by the defragmentation itself.
5. **Disarm the alarms.** Once every member has been defragmented, the operator lists the
   cluster's active alarms and disarms them. This is a genuine, necessary step, not a formality:
   when a member's backend grows past `--quota-backend-bytes`
   (`spec.configuration.tuning.quotaBackendBytes`), etcd raises a cluster-wide `NOSPACE` alarm and
   goes read-only until it is disarmed — and disarming it *before* the space was actually
   reclaimed would simply re-raise it. Doing it after the defragmentation is what makes the
   cluster writable again. On a cluster with no active alarms, this step is a harmless no-op.

## Progress and status

```bash
$ kubectl get etcdopsrequest -n demo
NAME              TYPE         STATUS       AGE
etcd-defragment   Defragment   Successful   4m
```

```bash
$ kubectl describe etcdopsrequest -n demo etcd-defragment
```

The full object of a clean run looks like this:

```yaml
$ kubectl get etcdopsrequest -n demo etcd-defragment -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  creationTimestamp: "2026-02-11T12:20:03Z"
  generation: 1
  name: etcd-defragment
  namespace: demo
  resourceVersion: "70412"
  uid: 9c4b7e11-1a26-4d0f-8b70-2e3f4a5b6c77
spec:
  apply: IfReady
  databaseRef:
    name: etcd-cluster
  defragment: {}
  maxRetries: 1
  type: Defragment
status:
  conditions:
  - lastTransitionTime: "2026-02-11T12:20:03Z"
    message: Defragmenting the etcd backends
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2026-02-11T12:23:47Z"
    message: Successfully defragmented every etcd member
    observedGeneration: 1
    reason: EtcdDefragmented
    status: "True"
    type: EtcdDefragmented
  - lastTransitionTime: "2026-02-11T12:23:52Z"
    message: Successfully disarmed the etcd alarms
    observedGeneration: 1
    reason: EtcdAlarmCleared
    status: "True"
    type: EtcdAlarmCleared
  - lastTransitionTime: "2026-02-11T12:23:53Z"
    message: Successfully defragmented the etcd cluster
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

Two conditions carry the outcome:

- **`EtcdDefragmented`** — every member has been defragmented. This is a single, cluster-wide
  condition that only flips to `True` after the last member (the leader) is done.
- **`EtcdAlarmCleared`** — the alarm-disarming step has run.

### Per-member progress conditions

The walk across members also reports progress per member, using the same
`<ConditionType>--<podName>` shape KubeDB uses elsewhere. `EtcdDefragmented--<pod>` and
`EtcdClusterHealthy--<pod>` are written for **every** member on every run — they are what the
operator reads back to know which members it has already defragmented, so a restart of the operator
resumes rather than starts over. While a step is still in flight the same condition is present with
`Status: False`, which is common on a large backend where the member is busy being rewritten:

```yaml
  - lastTransitionTime: "2026-02-11T12:21:11Z"
    message: 'etcd defragmented; ConditionStatus:True; PodName:etcd-cluster-1'
    observedGeneration: 1
    status: "True"
    type: EtcdDefragmented--etcd-cluster-1
  - lastTransitionTime: "2026-02-11T12:21:16Z"
    message: 'etcd cluster healthy; ConditionStatus:True; PodName:etcd-cluster-1'
    observedGeneration: 1
    status: "True"
    type: EtcdClusterHealthy--etcd-cluster-1
  - lastTransitionTime: "2026-02-11T12:22:30Z"
    message: 'etcd defragmented; ConditionStatus:True; PodName:etcd-cluster-2'
    observedGeneration: 1
    status: "True"
    type: EtcdDefragmented--etcd-cluster-2
  - lastTransitionTime: "2026-02-11T12:23:47Z"
    message: 'etcd defragmented; ConditionStatus:True; PodName:etcd-cluster-0'
    observedGeneration: 1
    status: "True"
    type: EtcdDefragmented--etcd-cluster-0
```

Reading those in order tells you the walk order the operator chose. In the excerpt above,
`etcd-cluster-0` was the Raft leader, so it was defragmented last.

## Verify the reclaimed space

Run `endpoint status` again and compare `DB SIZE` with what you recorded before:

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- etcdctl \
    --endpoints=http://etcd-cluster-0.etcd-cluster-pods.demo.svc:2379,http://etcd-cluster-1.etcd-cluster-pods.demo.svc:2379,http://etcd-cluster-2.etcd-cluster-pods.demo.svc:2379 \
    endpoint status -w table
```

`DB SIZE` should now be close to `DB SIZE IN USE` on every member. How much it drops depends
entirely on how much dead history there was — if you have never compacted, defragmentation has
little to reclaim, because the old revisions are still live data as far as the backend is
concerned. Which is the whole point of pairing the two: see
[Compact](/docs/guides/etcd/maintenance/compact.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-defragment
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- [Compact the keyspace history of an Etcd cluster](/docs/guides/etcd/maintenance/compact.md) —
  run this *before* a defragmentation to give it something to reclaim.
- [Move the Raft leadership of an Etcd cluster](/docs/guides/etcd/maintenance/move-leader.md).
- Detail concepts of the [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
