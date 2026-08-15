---
title: Etcd Recover from Quorum Loss
menu:
  docs_{{ .version }}:
    identifier: etcd-recover-from-quorum-loss-ops
    name: Recover from Quorum Loss
    parent: etcd-recover-from-quorum-loss
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Recover an Etcd Cluster from a Permanent Quorum Loss

This guide shows how to rebuild an `Etcd` cluster that has permanently lost its Raft quorum, using
an `EtcdOpsRequest` of type `RecoverFromQuorumLoss`. KubeDB keeps the single surviving member you
confirm, force-boots it as a brand new one-member cluster with etcd's own `--force-new-cluster`
procedure, and lets the ordinary membership reconciliation regrow the cluster from there.

> **This is a destructive, break-glass procedure.** Every member except the confirmed survivor has
> its data discarded permanently, and any write the lost majority had committed but the survivor had
> not yet applied is gone for good. It is never triggered automatically. Read the
> [overview](/docs/guides/etcd/recover-from-quorum-loss/overview.md) — especially
> [when *not* to use it](/docs/guides/etcd/recover-from-quorum-loss/overview.md#when-to-use-it) —
> before you run it.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
  configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps
  [here](/docs/setup/README.md). Etcd support is an **alpha** feature, so the operators must be
  installed with the `Etcd` feature gate turned on (`--set featureGates.Etcd=true`).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Recover from Quorum Loss Overview](/docs/guides/etcd/recover-from-quorum-loss/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd/recover-from-quorum-loss](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/recover-from-quorum-loss)
> directory of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Deploy an Etcd Cluster

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
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/recover-from-quorum-loss/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Wait until `etcd-cluster` is `Ready` and all three voting members are healthy:

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    6m

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

## Recognise a Permanent Quorum Loss

For this walkthrough, assume the two nodes hosting `etcd-cluster-0` and `etcd-cluster-2` were lost
together with their disks — their pods and their `PersistentVolumeClaim`s are gone and are not
coming back. Only `etcd-cluster-1` is left, and 1 out of 3 is not a quorum.

The `Etcd` object goes `NotReady`:

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS     AGE
etcd-cluster   3.6.4     NotReady   41m
```

and the Provisioner operator's health checker raises the `QuorumLost` condition on it:

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{range .status.conditions[?(@.type=="QuorumLost")]}{.type}{"\t"}{.status}{"\t"}{.reason}{"\n"}{.message}{"\n"}{end}'
QuorumLost	True	RaftQuorumLost
The Etcd: demo/etcd-cluster has lost its Raft quorum: 2 of 3 members are not answering. http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379: context deadline exceeded; http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379: context deadline exceeded
```

> **This condition is only a signal.** KubeDB will never act on it by itself. It is also the gate on
> the recovery: an ops request created against an `Etcd` that does *not* report `QuorumLost=True`
> right now is refused outright.

Before going any further, convince yourself the loss really is permanent — that the missing members'
volumes are gone, not merely unmounted, and that no node is about to come back. If the members can
return, waiting costs you nothing and loses no data; this procedure does both.

## Create a RecoverFromQuorumLoss EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-recover-quorum
  namespace: demo
spec:
  type: RecoverFromQuorumLoss
  databaseRef:
    name: etcd-cluster
  recoverFromQuorumLoss: {}
  timeout: 30m
  apply: Always
```

Here,

- `spec.recoverFromQuorumLoss.member` is **omitted**, so the operator picks the reachable member with
  the highest Raft applied index — the survivor that is least behind, and therefore the one that
  loses the fewest writes. Set it to a pod name (`etcd-cluster-1`) or a bare ordinal (`"1"`) if you
  want to choose yourself.
- `spec.recoverFromQuorumLoss.confirmMember` is **deliberately left unset for now**. It is the hard
  confirmation gate: nothing destructive happens until it names exactly the survivor the operator
  resolved. We fill it in a moment, once we have read that name.
- `spec.timeout` is **required** for `RecoverFromQuorumLoss`. Every step is bounded by it, and
  several — waiting for the rebuilt member to come up, waiting for a volume to rebind — have no
  other deadline at all.
- `spec.apply: Always` is what you want here. The default, `IfReady`, holds the request until the
  database is `Ready` — which, by definition, a cluster that has lost its quorum is not.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/recover-from-quorum-loss/etcdops-recover-quorum.yaml
etcdopsrequest.ops.kubedb.com/etcd-recover-quorum created
```

## Read the Resolved Survivor

The operator admits the request, pauses the database, resolves the survivor, and then **stops and
waits for you**:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                  TYPE                    STATUS        AGE
etcd-recover-quorum   RecoverFromQuorumLoss   Progressing   45s
```

The resolved choice is reported in two places. As a condition:

```bash
$ kubectl get etcdopsrequest -n demo etcd-recover-quorum -o jsonpath='{range .status.conditions[?(@.type=="EtcdQuorumLossAwaitingConfirmation")]}{.message}{"\n"}{end}'
Resolved etcd-cluster-1 as the surviving etcd member to rebuild demo/etcd-cluster from. Set spec.recoverFromQuorumLoss.confirmMember to "etcd-cluster-1" to allow the recovery to discard every other member's data; it currently reads "".
```

and as a companion condition that records the decision permanently:

```bash
$ kubectl get etcdopsrequest -n demo etcd-recover-quorum -o jsonpath='{range .status.conditions[*]}{.type}{"\n"}{end}' | grep EtcdQuorumLoss
EtcdQuorumLossAdmitted
EtcdQuorumLossSurvivorResolved--etcd-cluster-1
EtcdQuorumLossAwaitingConfirmation
```

The request will sit here **indefinitely**. This wait is a gate, not a retry: it never times out and
never fails the request, because what it is waiting for is a person reading the resolved survivor
and agreeing to it.

> The survivor is resolved **exactly once** and is never re-picked on a later pass. If
> `etcd-cluster-1` is not the member you want to keep, delete this ops request and create a new one
> with `spec.recoverFromQuorumLoss.member` naming the one you do want — there is no way to change
> the answer in place.

## Confirm the Survivor

Now that you have read the name, agree to it:

```bash
$ kubectl patch etcdopsrequest -n demo etcd-recover-quorum --type=merge \
    -p '{"spec":{"recoverFromQuorumLoss":{"confirmMember":"etcd-cluster-1"}}}'
etcdopsrequest.ops.kubedb.com/etcd-recover-quorum patched
```

The value has to match the resolved pod name **exactly**. Anything else — including the ordinal
form, if the operator resolved a pod name — leaves the request waiting.

From this point on the recovery proceeds on its own, and it is destructive.

## Verify the Recovery

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME                  TYPE                    STATUS       AGE
etcd-recover-quorum   RecoverFromQuorumLoss   Successful   12m
```

While it runs you can watch the cluster shrink to a single member and come back. The `PetSet` is
deleted up front (with orphaned pods), the survivor's volume is rebound under ordinal 0's claim
name, and ordinal 0 is booted with `--force-new-cluster`:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
NAME             READY   STATUS    RESTARTS   AGE
etcd-cluster-0   1/1     Running   0          38s

$ kubectl get pod -n demo etcd-cluster-0 -o jsonpath='{.spec.containers[0].args}' | tr ',' '\n' | tail -2
"--listen-metrics-urls=http://0.0.0.0:2381"
"--force-new-cluster"]
```

The flag is taken back off in a later step — it is not self-clearing, and left in place it would
rewrite the data directory into a fresh single-member cluster on every subsequent restart.

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-recover-quorum
Name:         etcd-recover-quorum
Namespace:    demo
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-08-15T09:02:11Z
  Generation:          2
  Resource Version:    204518
  UID:                 f2ad7a1e-1c94-4a4e-9d1e-4a2f7b6a9c31
Spec:
  Apply:  Always
  Database Ref:
    Name:  etcd-cluster
  Recover From Quorum Loss:
    Confirm Member:  etcd-cluster-1
  Timeout:           30m
  Type:              RecoverFromQuorumLoss
Status:
  Conditions:
    Last Transition Time:  2026-08-15T09:02:14Z
    Message:               Etcd demo/etcd-cluster reports QuorumLost=True
    Observed Generation:   1
    Status:                True
    Type:                  EtcdQuorumLossAdmitted
    Last Transition Time:  2026-08-15T09:02:16Z
    Message:               Recovery from the etcd quorum loss is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-08-15T09:02:31Z
    Message:               etcd quorum loss survivor resolved; ConditionStatus:True; PodName:etcd-cluster-1
    Observed Generation:   1
    Status:                True
    Type:                  EtcdQuorumLossSurvivorResolved--etcd-cluster-1
    Last Transition Time:  2026-08-15T09:06:04Z
    Message:               Confirmed etcd-cluster-1 as the surviving etcd member
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossAwaitingConfirmation
    Last Transition Time:  2026-08-15T09:06:09Z
    Message:               Etcd demo/etcd-cluster has still lost its quorum and etcd-cluster-0 is present; rebuilding it from etcd-cluster-1
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossPreflight
    Last Transition Time:  2026-08-15T09:06:22Z
    Message:               pet set deleted; ConditionStatus:True; PodName:etcd-cluster
    Observed Generation:   2
    Status:                True
    Type:                  PetSetDeleted--etcd-cluster
    Last Transition Time:  2026-08-15T09:07:41Z
    Message:               Relocated the volume of the surviving member etcd-cluster-1 onto etcd-cluster-0
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossPVCRelocated
    Last Transition Time:  2026-08-15T09:08:12Z
    Message:               Discarded every etcd member outside the surviving etcd-cluster-1
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossStaleMembersRemoved
    Last Transition Time:  2026-08-15T09:08:16Z
    Message:               Rewrote the cluster state ConfigMap for a single member cluster
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossClusterStateSeeded
    Last Transition Time:  2026-08-15T09:08:24Z
    Message:               Recreated etcd-cluster-0 with --force-new-cluster
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossForceNewClusterBoot
    Last Transition Time:  2026-08-15T09:09:02Z
    Message:               etcd-cluster-0 is serving as a healthy single member cluster
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossSingleMemberHealthy
    Last Transition Time:  2026-08-15T09:09:44Z
    Message:               Recreated etcd-cluster-0 without --force-new-cluster
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossFlagRemoved
    Last Transition Time:  2026-08-15T09:09:48Z
    Message:               Restored the reclaim policy of every volume the recovery touched
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossVolumesReclaimed
    Last Transition Time:  2026-08-15T09:09:53Z
    Message:               Restored the PetSet of Etcd demo/etcd-cluster around the rebuilt member
    Observed Generation:   2
    Status:                True
    Type:                  EtcdQuorumLossPetSetRestored
    Last Transition Time:  2026-08-15T09:09:57Z
    Message:               Successfully rebuilt the etcd cluster from the surviving member etcd-cluster-1
    Observed Generation:   2
    Reason:                RecoverEtcdFromQuorumLoss
    Status:                True
    Type:                  RecoverEtcdFromQuorumLoss
    Last Transition Time:  2026-08-15T09:10:02Z
    Message:               Successfully recovered the etcd cluster from the loss of its quorum
    Observed Generation:   2
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     2
  Phase:                   Successful
Events:
  Type    Reason                     Age    From                         Message
  ----    ------                     ----   ----                         -------
  Normal  PauseDatabase              12m    KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-cluster
  Normal  RecoverEtcdFromQuorumLoss  22s    KubeDB Ops-manager Operator  Successfully rebuilt the etcd cluster from the surviving member etcd-cluster-1
  Normal  ResumeDatabase             17s    KubeDB Ops-manager Operator  Successfully resumed Etcd demo/etcd-cluster
  Normal  Successful                 17s    KubeDB Ops-manager Operator  Successfully recovered the etcd cluster from the loss of its quorum
```

The per-resource progress conditions (`PetSetDeleted--…`, `PVCDeleted--…`, `GetPod--…`, `PodDeleted--…`, `PodCreated--…`, `PatchPV--…`,
and so on) are elided above. They are written for every step the recovery completes, not only for
steps that had to wait, and are useful when a recovery stalls.

## Watch the Cluster Regrow

The recovery itself stops at a healthy **one-member** cluster. Growing it back to `spec.replicas` is
not its job — once the `Etcd` object is resumed, the ordinary membership reconciliation notices one
member against three replicas and adds the others back through exactly the same
[learner-add/promote path](/docs/guides/etcd/scaling/horizontal-scaling/overview.md#learners) an
ordinary scale up uses. Each new member joins as a learner, streams the recovered keyspace from
`etcd-cluster-0`, and is promoted to a voting member once it has caught up:

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS     AGE
etcd-cluster   3.6.4     Critical   1h

$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
NAME             READY   STATUS    RESTARTS   AGE
etcd-cluster-0   1/1     Running   0          6m
etcd-cluster-1   1/1     Running   0          2m
etcd-cluster-2   0/1     Running   0          14s
```

The `Critical` phase here is expected — the client endpoint is serving, but not every member is
ready yet because a learner is still catching up. Once all three are voting members the cluster goes
back to `Ready`:

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    1h

$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 member list -w table
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
|        ID        | STATUS  |      NAME      |                             PEER ADDRS                              |                            CLIENT ADDRS                             | IS LEARNER |
+------------------+---------+----------------+---------------------------------------------------------------------+---------------------------------------------------------------------+------------+
| 8e9e05c52164694d | started | etcd-cluster-0 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| b2c6679ac05f2cf1 | started | etcd-cluster-1 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
| ffc1c9a1a4b6ef2b | started | etcd-cluster-2 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2380 | http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 |      false |
```

Note that the member IDs and the cluster ID are new: `--force-new-cluster` minted a brand new
cluster, which is exactly why the discarded members could never have rejoined with their old data
directories.

Finally, check that the `QuorumLost` condition has flipped back:

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{range .status.conditions[?(@.type=="QuorumLost")]}{.status}{"\t"}{.reason}{"\n"}{end}'
False	RaftQuorumHealthy
```

## Troubleshooting

- **`Etcd demo/etcd-cluster does not currently report QuorumLost=True, so it has not permanently
  lost its Raft quorum; refusing to run the destructive quorum loss recovery`** — the cluster still
  has quorum. This refusal is terminal: it marks the request `Failed` immediately rather than
  retrying, because retrying cannot change the answer.
- **`Etcd demo/etcd-cluster has regained its Raft quorum on its own; refusing to run the destructive
  quorum loss recovery`** — the cluster healed while the request was waiting for your confirmation.
  That is the good outcome; nothing was destroyed.
- **The request sits at `Progressing` and never moves** — check the
  `EtcdQuorumLossAwaitingConfirmation` condition. It is almost always waiting for
  `spec.recoverFromQuorumLoss.confirmMember` to match the resolved survivor exactly.
- **`spec.timeout is required for a recovery from quorum loss`** — add `spec.timeout`. It is not
  optional for this ops type.
- **`recovery from quorum loss is not applicable to an ephemeral etcd`** — `spec.storageType` is
  `Ephemeral`, so the members' data died with their pods and there is nothing to rescue.
- **`no member of etcd demo/etcd-cluster answered; there is no surviving member to rebuild the
  cluster from`** — nothing is left to recover from. Restore from a backup instead; see
  [in-place Restore](/docs/guides/etcd/restore/restore.md).
- **`the requested surviving member etcd-cluster-2 is not answering; a cluster cannot be rebuilt
  from a member that cannot be reached`** — the member named in
  `spec.recoverFromQuorumLoss.member` is one of the lost ones. Pick a member that still answers, or
  omit the field and let the operator choose.
- **The database is never picked up at all** — if `spec.apply` is left at its `IfReady` default, the
  ops request waits for an `Etcd` that will never become `Ready`. Use `apply: Always`.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-recover-quorum
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- [Restore a snapshot](/docs/guides/etcd/restore/restore.md) into an existing Etcd cluster when the
  data itself is the problem.
- Take scheduled snapshot [backups](/docs/guides/etcd/backup/kubestash/overview/index.md) so you
  always have something to fall back on.
- Scale the cluster [horizontally](/docs/guides/etcd/scaling/horizontal-scaling/horizontal-scaling.md)
  to raise the number of failures it can tolerate.
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
