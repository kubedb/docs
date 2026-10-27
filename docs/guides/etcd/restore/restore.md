---
title: Etcd In-place Restore
menu:
  docs_{{ .version }}:
    identifier: etcd-restore-ops
    name: Restore into an Existing Cluster
    parent: etcd-restore
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Restore a Snapshot into an Existing Etcd Cluster

This guide shows how to restore a [KubeStash](https://kubestash.com/) snapshot into an `Etcd` cluster
that already exists, using an `EtcdOpsRequest` of type `Restore`. KubeDB takes the cluster apart down
to a single empty volume, writes the snapshot into it with a KubeStash `RestoreSession`, brings the
seed member back on the restored data, and lets the ordinary membership reconciliation regrow the
cluster from there.

> **This replaces the entire keyspace of the database.** Everything currently stored in
> `etcd-cluster` is discarded and replaced by the contents of the snapshot, and every member's
> current data directory is thrown away. Full data loss is not a failure mode here — it is the
> normal, successful outcome of this operation. There is no undo, so make sure the snapshot you are
> restoring really is the one you want **before** you create the request.

> **Restoring into a brand-new cluster instead?** Use the bootstrap-time restore
> (`spec.init.archiver` on a new `Etcd` object) documented in
> [Snapshot Backup & Restore](/docs/guides/etcd/backup/kubestash/snapshot/index.md#restore). It is
> strictly safer, because there is no existing data to lose. See
> [Two Kinds of Restore](/docs/guides/etcd/restore/overview.md#two-kinds-of-restore).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
  configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps
  [here](/docs/setup/README.md). Etcd support is an **alpha** feature, so the operators must be
  installed with the `Etcd` feature gate turned on (`--set featureGates.Etcd=true`).

- Install [KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/) — the restore is
  driven by a KubeStash `RestoreSession`.

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md)
  - [In-place Restore Overview](/docs/guides/etcd/restore/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd/restore](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/restore)
> directory of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Prerequisite: a Repository with a Snapshot in It

This guide assumes you already have a running `Etcd` cluster that is backed up by an
[EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md), and therefore a KubeStash `Repository`
holding at least one successful **full** snapshot. If you do not, follow
[Snapshot Backup & Restore](/docs/guides/etcd/backup/kubestash/snapshot/index.md) first — the backup
half of that guide is exactly what this one consumes.

```bash
$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    2h

$ kubectl get repository -n demo
NAME             INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
etcd-full-repo   true        2                2.204MiB   Ready   6m32s                    1h

$ kubectl get snapshots -n demo -l kubestash.com/repo-name=etcd-full-repo
NAME                                     REPOSITORY       SESSION       SNAPSHOT-TIME   DELETION-POLICY   PHASE       AGE
etcd-full-repo-etcd-cluster-full-1755    etcd-full-repo   full-backup   2026-08-15T08:00:00Z   Delete     Succeeded   66m
etcd-full-repo-etcd-cluster-full-1756    etcd-full-repo   full-backup   2026-08-15T09:00:00Z   Delete     Succeeded   6m
```

Only snapshots whose `spec.type` is `Full` and whose phase is `Succeeded` are candidates. They are
matched **in the database's own namespace**, by repository name.

## Create a Restore EtcdOpsRequest

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-restore
  namespace: demo
spec:
  type: Restore
  databaseRef:
    name: etcd-cluster
  restore:
    fullDBRepository:
      name: etcd-full-repo
      namespace: demo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  timeout: 30m
  apply: Always
```

Here,

- `spec.restore.fullDBRepository` names the KubeStash `Repository` to restore the etcd snapshot
  from. It is **required and has no default**: unlike the bootstrap-time restore, this operation
  destroys data that is already there, so there is no safe repository to guess at.
- `spec.restore.recoveryTimestamp` is **optional** and omitted above, so the **latest** available
  snapshot is used. Set it to pick a point in time instead — the operator restores the newest
  successful full snapshot taken *at or before* that instant:

  ```yaml
    restore:
      fullDBRepository:
        name: etcd-full-repo
        namespace: demo
      recoveryTimestamp: "2026-08-15T08:30:00Z"
  ```

  Note this is snapshot selection, not point-in-time recovery. etcd has no continuous archiving
  primitive, so you can only land on an instant at which a snapshot was actually taken — everything
  written after it is gone.
- `spec.restore.encryptionSecret` refers to the Secret holding the encryption key the snapshot was
  backed up with. Omit it if the backup was not encrypted.
- `spec.timeout` is **required** for `Restore`. Every step is bounded by it, and the
  `RestoreSession` has no other deadline — so size it against how large the snapshot is, not against
  how long a pod takes to restart.
- `spec.apply: Always` is the safe choice here. The default, `IfReady`, holds the request until the
  database is `Ready`, which a degraded cluster may never become. Keep `IfReady` only if you are
  restoring into a perfectly healthy cluster and want that as an extra guard.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/restore/etcdops-restore.yaml
etcdopsrequest.ops.kubedb.com/etcd-restore created
```

## Verify the Restore

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME           TYPE      STATUS       AGE
etcd-restore   Restore   Successful   9m
```

While it runs you can watch the cluster collapse to a single member and come back. The `PetSet` is
deleted up front (with orphaned pods), the other members are discarded, and the seed member's claim
is replaced by an empty one for the snapshot to be written into:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
No resources found in demo namespace.

$ kubectl get restoresession -n demo
NAME                            REPOSITORY       FAILURE-POLICY   PHASE     DURATION   AGE
etcd-restore-snapshot-restorer  etcd-full-repo                    Running              41s
```

The `RestoreSession` is named after the **ops request**, not after the database. That is deliberate:
an `Etcd` that was itself bootstrapped from a backup already owns a succeeded session under the
bootstrap name, and a session is never rewritten once it exists — reusing that name would make the
operation read back `Succeeded` without a single byte having been restored.

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-restore
Name:         etcd-restore
Namespace:    demo
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-08-15T10:31:52Z
  Generation:          1
  Resource Version:    231904
  UID:                 6b2f4c8d-2c11-4f0a-8f77-1d3e5a9b7c02
Spec:
  Apply:  Always
  Database Ref:
    Name:  etcd-cluster
  Restore:
    Encryption Secret:
      Name:       encrypt-secret
      Namespace:  demo
    Full DB Repository:
      Name:       etcd-full-repo
      Namespace:  demo
  Timeout:        30m
  Type:           Restore
Status:
  Conditions:
    Last Transition Time:  2026-08-15T10:31:55Z
    Message:               In-place restore of the etcd keyspace is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-08-15T10:32:11Z
    Message:               Orphaned the PetSet of Etcd demo/etcd-cluster
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestorePetSetDeleted
    Last Transition Time:  2026-08-15T10:33:04Z
    Message:               Discarded every etcd member outside the seed etcd-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestoreMembersDiscarded
    Last Transition Time:  2026-08-15T10:33:41Z
    Message:               Replaced the data volume of etcd-cluster-0 with an empty PVC data-etcd-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestoreSeedPVCWiped
    Last Transition Time:  2026-08-15T10:33:49Z
    Message:               Created RestoreSession demo/etcd-restore-snapshot-restorer to restore Snapshot etcd-full-repo-etcd-cluster-full-1756 into PVC data-etcd-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestoreSessionCreated
    Last Transition Time:  2026-08-15T10:36:12Z
    Message:               RestoreSession demo/etcd-restore-snapshot-restorer restored the snapshot into PVC data-etcd-cluster-0
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestoreSnapshotApplied
    Last Transition Time:  2026-08-15T10:36:16Z
    Message:               Rewrote the cluster state ConfigMap for a single member cluster
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestoreClusterStateSeeded
    Last Transition Time:  2026-08-15T10:36:22Z
    Message:               Recreated the PetSet of Etcd demo/etcd-cluster around the restored seed member
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestorePetSetRestored
    Last Transition Time:  2026-08-15T10:37:05Z
    Message:               etcd-cluster-0 is serving the restored keyspace as a healthy single member cluster
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestoreSingleMemberHealthy
    Last Transition Time:  2026-08-15T10:37:09Z
    Message:               Restored the reclaim policy of every volume the restore discarded
    Observed Generation:   1
    Status:                True
    Type:                  EtcdRestoreVolumesReclaimed
    Last Transition Time:  2026-08-15T10:37:13Z
    Message:               Successfully restored Etcd demo/etcd-cluster from a snapshot of repository etcd-full-repo
    Observed Generation:   1
    Reason:                RestoreEtcdSnapshot
    Status:                True
    Type:                  RestoreEtcdSnapshot
    Last Transition Time:  2026-08-15T10:37:18Z
    Message:               Successfully restored Etcd demo/etcd-cluster from the snapshot
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason               Age   From                         Message
  ----    ------               ----  ----                         -------
  Normal  PauseDatabase        5m    KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-cluster
  Normal  RestoreEtcdSnapshot  25s   KubeDB Ops-manager Operator  Successfully restored Etcd demo/etcd-cluster from a snapshot of repository etcd-full-repo
  Normal  ResumeDatabase       20s   KubeDB Ops-manager Operator  Successfully resumed Etcd demo/etcd-cluster
  Normal  Successful           20s   KubeDB Ops-manager Operator  Successfully restored Etcd demo/etcd-cluster from the snapshot
```

The per-member and per-resource progress conditions (`EtcdRestoreMembersDiscarded--…`,
`PVCDeleted--…`, `CheckPVCDelete--…`, `PodDeleted--…`, `PodCreated--…`, and so on) are elided
above. They are written for every step the restore completes, not only for steps that had to wait,
and are useful when a restore stalls.

## Watch the Cluster Regrow

The restore itself stops at a healthy **one-member** cluster. Growing it back to `spec.replicas` is
not its job — once the `Etcd` object is resumed, the ordinary membership reconciliation notices one
member against three replicas and adds the others back through exactly the same
[learner-add/promote path](/docs/guides/etcd/scaling/horizontal-scaling/overview.md#learners) an
ordinary scale up uses. Each new member joins as a learner, streams the **restored** keyspace from
`etcd-cluster-0`, and is promoted to a voting member once it has caught up:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-cluster
NAME             READY   STATUS    RESTARTS   AGE
etcd-cluster-0   1/1     Running   0          4m
etcd-cluster-1   1/1     Running   0          92s
etcd-cluster-2   0/1     Running   0          11s

$ kubectl get etcd -n demo
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Critical 2h
```

> The `Critical` phase is expected here: the client endpoint is reachable and quorum is healthy, but
> `etcd-cluster-2` is still catching up as a learner. It becomes `Ready` once the last member is
> promoted.

## Verify the Restored Data

Read back a key you know was in the snapshot:

```bash
$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 get /kubedb --prefix -w fields | grep -E '^"(Key|Value)"'
"Key" : "/kubedb/demo"
"Value" : "hello"
```

And confirm every member is serving the same restored keyspace:

```bash
$ kubectl exec -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 endpoint status --cluster -w table
+---------------------------------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|                              ENDPOINT                               |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+---------------------------------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| http://etcd-cluster-0.etcd-cluster-pods.demo.svc.cluster.local:2379 | 8e9e05c52164694d |   3.6.4 |   20 kB |      true |      false |         2 |         31 |                 31 |        |
| http://etcd-cluster-1.etcd-cluster-pods.demo.svc.cluster.local:2379 | b2c6679ac05f2cf1 |   3.6.4 |   20 kB |     false |      false |         2 |         31 |                 31 |        |
| http://etcd-cluster-2.etcd-cluster-pods.demo.svc.cluster.local:2379 | ffc1c9a1a4b6ef2b |   3.6.4 |   20 kB |     false |      false |         2 |         31 |                 31 |        |
+---------------------------------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

## Troubleshooting

- **`spec.restore.fullDBRepository is required for an in-place restore: it replaces the whole
  keyspace of the database, so there is no safe repository to default to`** — name the repository.
  There is deliberately no default.
- **`spec.timeout is required for an in-place restore`** — add `spec.timeout`. It is not optional
  for this ops type.
- **`no successful full Snapshot found in repository <name> for Etcd demo/etcd-cluster`** — the
  repository holds no `Succeeded`, `Full` snapshot in the database's namespace. Check the repository
  name and that a backup has actually completed.
- **`no full Snapshot in repository <name> was taken at or before the requested recoveryTimestamp`**
  — your `recoveryTimestamp` predates every snapshot. Pick a later instant, or drop the field to use
  the latest.
- **`RestoreSession demo/<name>-snapshot-restorer failed to restore the snapshot into PVC
  data-etcd-cluster-0`** — this is **not** retried. A `RestoreSession` is a one-shot object and is
  never rewritten once it exists, so every later pass would read the same failure back. Look at the
  session (`kubectl describe restoresession -n demo <name>-snapshot-restorer`) and its pods for the
  real cause — a wrong or missing `encryptionSecret` is the usual one — then delete the failed ops
  request and create a fresh one. Nothing has been reclaimed at that point, so the old volumes are
  still there.
- **`can not restore Etcd demo/etcd-cluster from a snapshot: spec.storageType is Ephemeral, so there
  is no PVC to restore into before the seed member starts`** — an in-place restore needs a durable
  volume to write the snapshot into before etcd starts.
- **The request is never picked up at all** — if `spec.apply` is left at its `IfReady` default and
  the cluster is degraded, the ops request waits for an `Etcd` that may never become `Ready`. Use
  `apply: Always`.

## Cleaning Up

`etcd-cluster` and the `demo` namespace are prerequisites this guide assumes were already there (see
[Prerequisite: a Repository with a Snapshot in It](#prerequisite-a-repository-with-a-snapshot-in-it))
rather than resources it created, so the default cleanup only removes the `EtcdOpsRequest`:

```bash
kubectl delete etcdopsrequest -n demo etcd-restore
```

If `demo` really was a disposable environment set up just for this walkthrough, you can additionally
remove the cluster and the namespace:

```bash
kubectl delete etcd -n demo etcd-cluster
kubectl delete ns demo
```

## Next Steps

- Set up scheduled snapshot [backups](/docs/guides/etcd/backup/kubestash/overview/index.md) with an
  [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md).
- Bootstrap a **new** cluster from a snapshot instead: see
  [Snapshot Backup & Restore](/docs/guides/etcd/backup/kubestash/snapshot/index.md#restore).
- [Recover from a permanent quorum loss](/docs/guides/etcd/recover-from-quorum-loss/overview.md) when
  the data is fine but the cluster is not.
- Detail concepts of the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) object.
