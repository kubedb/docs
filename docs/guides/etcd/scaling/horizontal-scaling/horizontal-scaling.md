---
title: Horizontal Scaling Etcd
menu:
  docs_{{ .version }}:
    identifier: etcd-horizontal-scaling-ops
    name: Scale Horizontally
    parent: etcd-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Etcd

This guide will show you how to use the `KubeDB` Ops-manager operator to change the number of members
of an `Etcd` cluster.

Before you run any of this, read
[Horizontal Scaling Overview](/docs/guides/etcd/scaling/horizontal-scaling/overview.md). Scaling an
etcd cluster is a **sequence of Raft membership changes**, not an instant replica-count change, and
the guide below will make a lot more sense once you know why.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
  configured to communicate with your cluster. If you do not already have a cluster, you can create
  one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps
  [here](/docs/setup/README.md). Etcd support is an **alpha** feature, so the operators must be
  installed with the `Etcd` feature gate turned on (`--set featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/etcd/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd/scaling/horizontal-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/scaling/horizontal-scaling)
> directory of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Deploy a 3-Member Etcd Cluster

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/scaling/horizontal-scaling/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Now, wait until `etcd-cluster` has status `Ready`. i.e,

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    4m
```

Let's check the member count from the `Etcd` object and from the `PetSet`:

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.replicas}'
3

$ kubectl get petset -n demo etcd-cluster -o jsonpath='{.spec.replicas}'
3
```

Those two are Kubernetes' view of the world. etcd's own view — the one that actually decides quorum —
is the member list:

```bash
$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 member list -w table
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
|        ID        | STATUS  |      NAME      |                             PEER ADDRS                              |                            CLIENT ADDRS                             | IS LEARNER |
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
| 1a2b3c4d5e6f7a8b | started | etcd-cluster-0 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 2b3c4d5e6f7a8b9c | started | etcd-cluster-1 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 3c4d5e6f7a8b9c0d | started | etcd-cluster-2 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
```

Three voting members (`IS LEARNER` is `false` for all of them), so quorum is 2 and the cluster
tolerates one failure.

> If you have enabled etcd's RBAC authentication (for example by running a `RotateAuth`
> `EtcdOpsRequest`), add `--user=root:<password>` to the `etcdctl` calls, with the password read from
> the auth Secret:
> `kubectl get secret -n demo etcd-cluster-auth -o jsonpath='{.data.password}' | base64 -d`.
> If TLS is enabled, use the `https://` scheme and pass the client certificate flags.

## Scale Up: 3 → 5 Members

### Create the EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: etcd-cluster
  horizontalScaling:
    replicas: 5
  timeout: 20m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are scaling the `etcd-cluster` database.
- `spec.type` specifies that we are performing `HorizontalScaling`.
- `spec.horizontalScaling.replicas` is the desired **number of etcd members** after scaling. It must
  be at least `1`; odd values are strongly recommended (see
  [Choosing a Member Count](/docs/guides/etcd/scaling/horizontal-scaling/overview.md#choosing-a-member-count)).
- `spec.timeout` bounds each polling step. Give a scale up room: every new member has to replicate the
  whole keyspace before it can be promoted, so a large cluster legitimately takes a while.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/scaling/horizontal-scaling/etcdops-hscale-up.yaml
etcdopsrequest.ops.kubedb.com/etcd-hscale-up created
```

### What Happens Now

The ops request itself does very little: it patches `spec.replicas: 5` on the `Etcd` object and then
waits. The actual member-by-member work is driven by the **Provisioner** operator's membership
reconciler, which runs continuously and performs exactly one membership mutation per pass:

| Pass | Provisioner action                                                       | Voting members | Learners |
|------|--------------------------------------------------------------------------|:--------------:|:--------:|
| 1    | ConfigMap rewritten, `MemberAddAsLearner` for `etcd-cluster-3`, `PetSet` → 4 | 3            | 1        |
| 2..n | Poll `etcd-cluster-3`'s applied revision against the leader's             | 3              | 1        |
| n+1  | `MemberPromote` `etcd-cluster-3`                                          | 4              | 0        |
| n+2  | ConfigMap rewritten, `MemberAddAsLearner` for `etcd-cluster-4`, `PetSet` → 5 | 4            | 1        |
| ...  | Poll, then `MemberPromote` `etcd-cluster-4`                               | 5              | 0        |

Note the invariant across every row: **at most one learner exists at a time, and quorum is computed
only over the voting members.** While `etcd-cluster-3` is catching up, the cluster still needs 2 of
its 3 voters — the in-flight member cannot make things worse.

You can watch the sequence in the Provisioner operator's log:

```bash
$ kubectl logs -n kubedb -l app.kubernetes.io/name=etcd-operator -f | grep etcd-cluster
etcd demo/etcd-cluster: added etcd-cluster-3 as a learner
etcd demo/etcd-cluster: promoted learner etcd-cluster-3 to a voting member
etcd demo/etcd-cluster: added etcd-cluster-4 as a learner
etcd demo/etcd-cluster: promoted learner etcd-cluster-4 to a voting member
```

And in the member list, mid-flight — note `IS LEARNER: true` on the member that is catching up:

```bash
$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 member list -w table
+------------------+-----------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
|        ID        |  STATUS   |      NAME      |                             PEER ADDRS                              |                            CLIENT ADDRS                             | IS LEARNER |
+------------------+-----------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
| 1a2b3c4d5e6f7a8b | started   | etcd-cluster-0 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 2b3c4d5e6f7a8b9c | started   | etcd-cluster-1 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 3c4d5e6f7a8b9c0d | started   | etcd-cluster-2 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 4d5e6f7a8b9c0d1e | started   | etcd-cluster-3 | http://etcd-cluster-3.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-3.etcd-cluster-pods.demo.svc.cluster.local:2379 |       true |
+------------------+-----------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
```

> **The `Etcd` object may report `Critical` while a learner is catching up.** That is expected, not an
> error: a learner cannot serve the linearizable read behind etcd's health endpoint, so its pod never
> becomes `Ready`, and KubeDB reports `Critical` for "the client endpoint works but not every member
> is ready". The phase returns to `Ready` once the learner is promoted. This is also exactly why
> KubeDB gates promotion on the learner's **applied revision**, not on pod readiness — gating on
> readiness would deadlock the scale up permanently.

### Verify the Scale Up

Let's wait for the `EtcdOpsRequest` to become `Successful`:

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME             TYPE                STATUS       AGE
etcd-hscale-up   HorizontalScaling   Successful   6m41s
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-hscale-up
Name:         etcd-hscale-up
Namespace:    demo
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-08-15T10:02:11Z
  Generation:          1
  Resource Version:    131204
  UID:                 5f6c0f21-95a2-4c7d-9a13-2f1b8e6d4c77
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  etcd-cluster
  Horizontal Scaling:
    Replicas:  5
  Timeout:     20m
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2026-08-15T10:02:11Z
    Message:               Horizontal Scaling is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-08-15T10:02:16Z
    Message:               Successfully updated the desired member count
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-08-15T10:05:41Z
    Message:               is pet set scaled; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetScaled
    Last Transition Time:  2026-08-15T10:07:56Z
    Message:               etcd learner promoted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  EtcdLearnerPromoted
    Last Transition Time:  2026-08-15T10:08:01Z
    Message:               etcd cluster healthy; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  EtcdClusterHealthy
    Last Transition Time:  2026-08-15T10:08:06Z
    Message:               The etcd member list has converged on the requested member count
    Observed Generation:   1
    Reason:                EtcdMemberListReady
    Status:                True
    Type:                  EtcdMemberListReady
    Last Transition Time:  2026-08-15T10:08:08Z
    Message:               Successfully horizontally scaled the etcd cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                        Age    From                         Message
  ----     ------                                        ----   ----                         -------
  Warning  is pet set scaled; ConditionStatus:False      6m36s  KubeDB Ops-manager Operator  is pet set scaled; ConditionStatus:False
  Warning  etcd member list ready; ConditionStatus:False 6m31s  KubeDB Ops-manager Operator  etcd member list ready; ConditionStatus:False
  Warning  etcd learner promoted; ConditionStatus:False  3m06s  KubeDB Ops-manager Operator  etcd learner promoted; ConditionStatus:False
  Warning  etcd learner promoted; ConditionStatus:True   52s    KubeDB Ops-manager Operator  etcd learner promoted; ConditionStatus:True
  Normal   EtcdMemberListReady                           42s    KubeDB Ops-manager Operator  The etcd member list has converged on the requested member count
  Normal   Successful                                    40s    KubeDB Ops-manager Operator  Successfully horizontally scaled the etcd cluster
```

Reading that condition list:

- `UpdateDatabase` is the whole "action" the ops request takes — it wrote `spec.replicas: 5`.
- `IsPetSetScaled`, `EtcdLearnerPromoted`, `EtcdClusterHealthy` and `EtcdMemberListReady` are the
  **convergence gates**. They flip to `False` while the poller is waiting and back to `True` when the
  condition is met, which is why you see both polarities of `etcd learner promoted` in the events —
  the ops request observed a learner mid-promotion and waited for it. On a `3 → 5` scale up you will
  typically see that pair of events twice, once per added member.
- `EtcdMemberListReady` goes `True` as soon as the `PetSet` carries 5 replicas and etcd reports
  exactly 5 members. The no-learner and quorum checks that follow it report through
  `EtcdLearnerPromoted` and `EtcdClusterHealthy`. It is the request's own top-level
  `HorizontalScaling` condition — and the `Successful` phase — that means all four passed together.
- There is **no** `PauseDatabase` / `ResumeDatabase` event pair here. `HorizontalScaling` is the one
  `EtcdOpsRequest` type that does not pause the database, because the Provisioner operator is the
  component doing the membership work.

Now let's verify the result from all three viewpoints:

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.replicas}'
5

$ kubectl get petset -n demo etcd-cluster -o jsonpath='{.spec.replicas}'
5

$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
NAME             READY   STATUS    RESTARTS   AGE
etcd-cluster-0   1/1     Running   0          21m
etcd-cluster-1   1/1     Running   0          20m
etcd-cluster-2   1/1     Running   0          20m
etcd-cluster-3   1/1     Running   0          6m
etcd-cluster-4   1/1     Running   0          3m
```

And, most importantly, from etcd itself — five members, **all voting**:

```bash
$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 member list -w table
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
|        ID        | STATUS  |      NAME      |                             PEER ADDRS                              |                            CLIENT ADDRS                             | IS LEARNER |
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
| 1a2b3c4d5e6f7a8b | started | etcd-cluster-0 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 2b3c4d5e6f7a8b9c | started | etcd-cluster-1 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 3c4d5e6f7a8b9c0d | started | etcd-cluster-2 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 4d5e6f7a8b9c0d1e | started | etcd-cluster-3 | http://etcd-cluster-3.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-3.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 5e6f7a8b9c0d1e2f | started | etcd-cluster-4 | http://etcd-cluster-4.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-4.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
```

Quorum is now 3 of 5, so the cluster tolerates two simultaneous failures.

## Scale Down: 5 → 3 Members

### Create the EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: etcd-cluster
  horizontalScaling:
    replicas: 3
  timeout: 20m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/scaling/horizontal-scaling/etcdops-hscale-down.yaml
etcdopsrequest.ops.kubedb.com/etcd-hscale-down created
```

### Removal Comes Before the PetSet Shrinks

This is the part that is easy to get wrong and that KubeDB gets right for you. For each member being
dropped, the Provisioner operator:

1. Picks the **highest-ordinal** member — `etcd-cluster-4` first. Members are mapped back to `PetSet`
   ordinals by peer URL, because a member that has not finished starting reports an empty name.
2. Calls **`MemberRemove` on that member while its pod is still running**. The cluster commits the
   configuration change, and its quorum requirement immediately drops from 3-of-5 to 3-of-4.
3. Rewrites the cluster-state ConfigMap and only *then* shrinks the `PetSet` to 4, which deletes the
   pod.

Then the whole thing repeats for `etcd-cluster-3`. One member per pass — never two removals in
flight.

If the pod were deleted first and the member removed afterwards, there would be a window in which
etcd still counted a member whose process no longer exists: a 5-member cluster requiring 3 votes with
only 4 live members, one failure away from losing quorum for no reason at all. Removing first closes
that window.

```bash
$ kubectl logs -n kubedb -l app.kubernetes.io/name=etcd-operator -f | grep etcd-cluster
etcd demo/etcd-cluster: removed member 6804325408493731870 (etcd-cluster-4)
etcd demo/etcd-cluster: removed member 5571401537629186833 (etcd-cluster-3)
```

### Verify the Scale Down

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME               TYPE                STATUS       AGE
etcd-hscale-down   HorizontalScaling   Successful   2m14s
```

```bash
$ kubectl describe etcdopsrequest -n demo etcd-hscale-down
...
Status:
  Conditions:
    Message:  Horizontal Scaling is in progress
    Reason:   Running
    Status:   True
    Type:     Running
    Message:  Successfully updated the desired member count
    Reason:   UpdateDatabase
    Status:   True
    Type:     UpdateDatabase
    Message:  is pet set scaled; ConditionStatus:True
    Status:   True
    Type:     IsPetSetScaled
    Message:  etcd cluster healthy; ConditionStatus:True
    Status:   True
    Type:     EtcdClusterHealthy
    Message:  The etcd member list has converged on the requested member count
    Reason:   EtcdMemberListReady
    Status:   True
    Type:     EtcdMemberListReady
    Message:  Successfully horizontally scaled the etcd cluster
    Reason:   Successful
    Status:   True
    Type:     Successful
  Phase:      Successful
```

Note that the `EtcdLearnerPromoted` gate means something different on the way down: nothing is ever
added as a learner during a scale down, so the check is simply "no member is still a learner", and it
passes on the first poll. It still records a `True` condition — you just never see the `False`
polarity that a scale up produces while it waits for a learner to catch up. The scale down also
opens with an `EtcdMemberRemoved` condition, where a scale up opens with `EtcdMemberAdded`.

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.replicas}'
3

$ kubectl get petset -n demo etcd-cluster -o jsonpath='{.spec.replicas}'
3

$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 member list -w table
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
|        ID        | STATUS  |      NAME      |                             PEER ADDRS                              |                            CLIENT ADDRS                             | IS LEARNER |
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
| 1a2b3c4d5e6f7a8b | started | etcd-cluster-0 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 2b3c4d5e6f7a8b9c | started | etcd-cluster-1 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| 3c4d5e6f7a8b9c0d | started | etcd-cluster-2 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
```

The `PetSet` no longer manages `etcd-cluster-3` and `etcd-cluster-4`, but their PVCs are **retained**:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=etcd-cluster \
  -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,SIZE:.status.capacity.storage
NAME                  STATUS    SIZE
data-etcd-cluster-0   Bound     1Gi
data-etcd-cluster-1   Bound     1Gi
data-etcd-cluster-2   Bound     1Gi
data-etcd-cluster-3   Bound     1Gi
data-etcd-cluster-4   Bound     1Gi
```

Delete them by hand if you want the storage back — KubeDB keeps them so a later scale up (or a
mistaken scale down) does not destroy data.

## Troubleshooting

- **The ops request sits in `Progressing` with `etcd learner promoted; ConditionStatus:False`.**
  The new member is still replicating. Check its log (`kubectl logs -n demo etcd-cluster-3 -c etcd`)
  and give the request a longer `spec.timeout`. Promotion happens once the learner's applied revision
  reaches 90% of the leader's; a large keyspace or a slow disk simply takes longer.
- **The new pod is `Running` but never `Ready`.** Expected while it is a learner — see the note
  above. If it stays that way after promotion, look at the member's own log.
- **The `PetSet` replica count and the etcd member list disagree.** The provisioner repairs this
  itself on the next pass, realigning the `PetSet` with etcd's member list (`realigning petset
  replicas (N) with the etcd member count (M)` in the operator log). etcd's member list is the
  authority.
- **Another `EtcdOpsRequest` is already running.** Only one ops request per database may be in
  `Progressing`; the new one waits and retries rather than racing it.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-hscale-up etcd-hscale-down
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- Scale the cluster [vertically](/docs/guides/etcd/scaling/vertical-scaling/vertical-scaling.md).
- [Expand the volume](/docs/guides/etcd/volume-expansion/volume-expansion.md) of an Etcd cluster.
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
