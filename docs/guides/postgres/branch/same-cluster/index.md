---
title: Branch a PostgreSQL Database | KubeDB
description: Create an instant, writable copy of a running KubeDB PostgreSQL database in the same cluster
menu:
  docs_{{ .version }}:
    identifier: guides-pg-branch-same-cluster
    name: Same Cluster Branching
    parent: guides-pg-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch a PostgreSQL Database in the Same Cluster

This guide walks through branching a running KubeDB `PostgreSQL` database into a second, fully independent `PostgreSQL` database in the same Kubernetes cluster.

You will deploy a source database, branch it, inspect what KubeDB built, connect to the branch, confirm it carries the source's data while staying isolated from it, and finally tear it down.

This page covers the simplest case: a standalone database branched into its own namespace, reusing the source credential, taken once. Once that works, [Customizing a PostgreSQL Branch](/docs/guides/postgres/branch/customization/index.md) shows the optional fields — a separate credential, anonymizing post-actions, scheduled refresh, a different target namespace, replicated clusters, and snapshot-class pinning.

If you have not read [PostgreSQL Branching Overview](/docs/guides/postgres/branch/overview/index.md) yet, start there — it explains what a branch is and how KubeDB builds one.

## Before You Begin

- You need a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with it.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md), with the **KubeDB Courier** operator enabled by adding `--set kubedb-courier.enabled=true` to the install command. A fresh install brings the `Branch` CRD with it.

  If instead you are enabling Courier on an **existing** install, note that `helm upgrade` does not apply CRDs, so you have to add the `Branch` CRD yourself:

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kubedb/apimachinery/refs/heads/master/crds/courier.kubedb.com_branches.yaml
  ```

  Then check that the CRD your cluster ended up with is the one this guide describes:

  ```bash
  $ kubectl explain branch.spec.postActions
  GROUP:      courier.kubedb.com
  KIND:       Branch
  VERSION:    v1alpha1

  FIELD: postActions <[]Object>
  ```

  If that command reports the field does not exist, your `Branch` CRD predates the current schema and the [customization guide](/docs/guides/postgres/branch/customization/index.md) will not apply cleanly — re-apply the CRD above.

- Branching is built on CSI volume snapshots. Your cluster needs:
    - The [external-snapshotter](https://github.com/kubernetes-csi/external-snapshotter) CRDs and snapshot controller.
    - A CSI driver that supports volume snapshots (AWS EBS, GCE PD, Ceph RBD, TopoLVM, Longhorn, Rook, …), backing the StorageClass your database uses.
    - A `VolumeSnapshotClass` for that driver.

- You should be familiar with the following concepts:
    - [PostgreSQL](/docs/guides/postgres/concepts/postgres.md)
    - [AppBinding](/docs/guides/postgres/concepts/appbinding.md)
    - [Branch](/docs/guides/postgres/concepts/branch.md)
    - [PostgreSQL Branching Overview](/docs/guides/postgres/branch/overview/index.md)

How much of the snapshot stack you have to install yourself depends on your platform.

<details>
<summary><b>Setting up the snapshot stack on your platform.</b></summary>

**Amazon EKS** <br>

The EBS CSI driver ships as an add-on, but the snapshot CRDs and controller do not — install them, then create a `VolumeSnapshotClass` for `ebs.csi.aws.com`:

```bash
kubectl apply -k "https://github.com/kubernetes-csi/external-snapshotter/client/config/crd?ref=v8.2.0"
kubectl apply -k "https://github.com/kubernetes-csi/external-snapshotter/deploy/kubernetes/snapshot-controller?ref=v8.2.0"
```

<br> **Google GKE** <br>

The snapshot controller is built in. Run `kubectl get volumesnapshotclass` and confirm a class exists for `pd.csi.storage.gke.io`; create one if the list is empty.

<br> **Azure AKS** <br>

The snapshot controller is built in. Run `kubectl get volumesnapshotclass` and confirm a class exists for the Azure Disk CSI driver; create one if the list is empty.

<br> **On-prem (Rook/Ceph, Longhorn, TopoLVM, …)** <br>

Install the external-snapshotter CRDs and controller as shown for EKS above, then create a `VolumeSnapshotClass` naming your driver. For TopoLVM, use a **thin** device-class pool — thick volumes cannot be snapshotted.

</details>

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/branch/same-cluster/examples](/docs/guides/postgres/branch/same-cluster/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Verify the Prerequisites

Check that the Courier operator is running and the `Branch` CRD is installed. Courier runs in whichever namespace you installed KubeDB into, so search by label:

```bash
$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=kubedb-courier
NAMESPACE   NAME                              READY   STATUS    RESTARTS   AGE
courier     kubedb-courier-7dcc86b9d5-8pjqh   1/1     Running   0          23h

$ kubectl get crd branches.courier.kubedb.com
NAME                          CREATED AT
branches.courier.kubedb.com   2026-07-14T11:44:22Z
```

Check that the snapshot stack is present:

```bash
$ kubectl get crd | grep snapshot.storage.k8s.io
volumegroupsnapshotclasses.groupsnapshot.storage.k8s.io        2026-07-14T11:56:50Z
volumegroupsnapshotcontents.groupsnapshot.storage.k8s.io       2026-07-14T11:56:50Z
volumegroupsnapshots.groupsnapshot.storage.k8s.io              2026-07-14T11:56:50Z
volumesnapshotclasses.snapshot.storage.k8s.io                  2026-07-14T11:56:50Z
volumesnapshotcontents.snapshot.storage.k8s.io                 2026-07-14T11:56:50Z
volumesnapshots.snapshot.storage.k8s.io                        2026-07-14T11:56:50Z
```

Finally, confirm you have a `VolumeSnapshotClass` whose driver matches the StorageClass you will use for the database:

```bash
$ kubectl get volumesnapshotclass
NAME                       DRIVER             DELETIONPOLICY   AGE
ceph-rbd-vsc               rbd.csi.ceph.com   Delete           3d4h
topolvm-provisioner-thin   topolvm.io         Delete           22d
topolvm-vsc-explicit       topolvm.io         Delete           7d3h

$ kubectl get sc topolvm-provisioner-thin
NAME                       PROVISIONER   RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
topolvm-provisioner-thin   topolvm.io    Delete          WaitForFirstConsumer   true                   22d
```

This cluster has snapshot classes for two drivers. What matters is that the StorageClass and the `VolumeSnapshotClass` you use are backed by the **same** one: here the `topolvm-provisioner-thin` StorageClass is provisioned by `topolvm.io`, and the `topolvm-provisioner-thin` snapshot class snapshots that same driver, so a snapshot taken from one of its volumes can be restored into another. (`topolvm-vsc-explicit` is a second, non-default class for the same driver, used on the [customization page](/docs/guides/postgres/branch/customization/index.md#pin-a-volumesnapshotclass) to demonstrate pinning.)

This guide uses TopoLVM throughout. Substitute your own StorageClass and its matching snapshot class — on EKS that would be a `gp3` StorageClass with a `VolumeSnapshotClass` for `ebs.csi.aws.com`.

## Deploy the Source Database

Let's deploy a sample `PostgreSQL` database that will act as our "production" database.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sample-postgres
  namespace: demo
spec:
  version: "16.1"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: topolvm-provisioner-thin
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Create it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/same-cluster/examples/sample-postgres.yaml
postgres.kubedb.com/sample-postgres created
```

Wait until it is ready:

```bash
$ kubectl get pg -n demo sample-postgres
NAME              VERSION   STATUS   AGE
sample-postgres   16.1      Ready    24s
```

Now insert some data so we have something to branch. We will create a `customers` table with 1000 rows containing e-mail addresses and phone numbers — the kind of data you would not want to hand to a developer unmasked.

```bash
$ kubectl exec -it -n demo sample-postgres-0 -c postgres -- psql -U postgres
```

```sql
CREATE TABLE customers (
    id        SERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    email     TEXT NOT NULL,
    phone     TEXT NOT NULL
);

INSERT INTO customers (full_name, email, phone)
SELECT 'Customer ' || i,
       'customer' || i || '@example.com',
       '+1-555-' || LPAD(i::TEXT, 4, '0')
FROM generate_series(1, 1000) AS i;

SELECT count(*) FROM customers;
```

```
 count
-------
  1000
(1 row)
```

## Create a Branch

Below is the YAML of the `Branch` object we are going to create:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: dev-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: sample-postgres
    namespace: demo
  target:
    namespace: demo
    name: dev-postgres
    storageClassName: topolvm-provisioner-thin
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

Here,

- `spec.source.databaseRef` points to the KubeDB database to branch. `spec.source.namespace` defaults to the `Branch`'s own namespace when omitted.
- `spec.target` describes **only what differs** from the source. By default the rest — the PostgreSQL version, the replica count, the custom configuration, the pod template — is inherited from `sample-postgres`. (Two other top-level fields also change the result, both covered on the [customization page](/docs/guides/postgres/branch/customization/index.md): `spec.resetRootPassword` replaces the inherited credential, and `spec.postActions` mutates the cloned data.)
    - `spec.target.name` / `spec.target.namespace`: the `Postgres` object to create.
    - `spec.target.storageClassName`: the StorageClass for the branch's volumes. It must be backed by the same CSI driver as the source.
    - `spec.target.resources`: CPU/memory for the branch — usually smaller than production.
    - `spec.target.clusterName` is omitted, which selects a same-cluster branch.
- `spec.deletionPolicy: Delete` means deleting this `Branch` also removes the branched database. Use `Orphan` to keep it.

For a full description of every field, see the [Branch CRD reference](/docs/guides/postgres/concepts/branch.md).

Create the `Branch`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/same-cluster/examples/dev-branch.yaml
branch.courier.kubedb.com/dev-branch created
```

Watch it progress through its phases:

```bash
$ kubectl get branch -n demo -w
NAME         PHASE          MODE    TARGET         FRESHNESS   AGE
dev-branch                          dev-postgres               0s
dev-branch   Snapshotting   Local   dev-postgres               0s
dev-branch   Snapshotting   Local   dev-postgres               10s
dev-branch   Cloning        Local   dev-postgres               10s
dev-branch   Provisioning   Local   dev-postgres               10s
dev-branch   Provisioning   Local   dev-postgres               40s
dev-branch   Ready          Local   dev-postgres   0s          40s
```

The branch was ready in **40 seconds**: `Snapshotting` (take a CSI snapshot of the source volume) → `Cloning` (create the branch's volume from that snapshot) → `Provisioning` (KubeDB brings PostgreSQL up on the cloned volume) → `Ready`. `MODE: Local` confirms this is a same-cluster branch.

Most of that time went into starting the target database, not into copying data. Because the clone is copy-on-write, the snapshot and clone steps take about as long for a 1 TB database as for this 1 GB one.

## Inspect the Branch

`status` records exactly what the branch is made of:

```bash
$ kubectl get branch -n demo dev-branch -o yaml
```

```yaml
status:
  conditions:
  - lastTransitionTime: "2026-08-06T10:05:33Z"
    message: source snapshot br-dev-branch-data-sample-postgres-0-0 is readyToUse
    observedGeneration: 1
    reason: SnapshotReady
    status: "True"
    type: SnapshotReady
  - lastTransitionTime: "2026-08-06T10:05:33Z"
    message: target Database created and adopted the cloned PVC
    observedGeneration: 1
    reason: TargetCreated
    status: "True"
    type: TargetCreated
  - lastTransitionTime: "2026-08-06T10:06:03Z"
    message: branched Database reached Ready
    observedGeneration: 1
    reason: TargetReady
    status: "True"
    type: TargetReady
  - lastTransitionTime: "2026-08-06T10:06:03Z"
    message: branch is Ready (refresh gen 0)
    observedGeneration: 1
    reason: Ready
    status: "True"
    type: Ready
  history:
  - at: "2026-08-06T10:06:03Z"
    result: Succeeded
  lastRefreshTime: "2026-08-06T10:06:03Z"
  lastSuccessfulRefreshTime: "2026-08-06T10:06:03Z"
  mode: Local
  phase: Ready
  resources:
    authSecret: dev-postgres-auth
    clonedPVCs:
    - data-dev-postgres-0
  snapshot:
    members:
    - creationTime: "2026-08-06T10:05:23Z"
      name: br-dev-branch-data-sample-postgres-0-0
      readyToUse: true
      restoreSize: 1Gi
      sourcePVC: data-sample-postgres-0
    ready: true
    strategy: VolumeSnapshot
```

- `status.snapshot` records the snapshot set the current copy was made from — the `strategy` used (`VolumeSnapshot` per volume, or `VolumeGroupSnapshot` for a group-consistent snapshot), and one `members` entry per source data volume.
- `status.resources` lists everything the branch owns, which is what teardown removes.
- `status.history` keeps the outcome of each run, bounded by `spec.historyLimit`.

The branched database itself is an ordinary KubeDB `Postgres`:

```bash
$ kubectl get pg -n demo
NAME              VERSION   STATUS   AGE
dev-postgres      16.1      Ready    46s
sample-postgres   16.1      Ready    104s
```

It has its own Services, auth Secret, and `AppBinding`:

```bash
$ kubectl get svc,secret,appbinding -n demo -l app.kubernetes.io/instance=dev-postgres
NAME                        TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                               AGE
service/dev-postgres        ClusterIP   10.43.54.140   <none>        5432/TCP,2379/TCP                     46s
service/dev-postgres-pods   ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP,2384/TCP   46s

NAME                       TYPE                       DATA   AGE
secret/dev-postgres-auth   kubernetes.io/basic-auth   2      46s

NAME                                              TYPE                  VERSION   AGE
appbinding.appcatalog.appscode.com/dev-postgres   kubedb.com/postgres   16.1      43s
```

And it is annotated with where it came from:

```bash
$ kubectl get pg -n demo dev-postgres -o jsonpath='{.metadata.annotations.kubedb\.com/branched-from}'
{"cluster":"","source":"demo/sample-postgres"}
```

### The Storage Layer

The branch's volume was cloned from the snapshot rather than provisioned empty — its `dataSource` proves it:

```bash
$ kubectl get volumesnapshot -n demo
NAME                                     READYTOUSE   SOURCEPVC                RESTORESIZE   SNAPSHOTCLASS              CREATIONTIME   AGE
br-dev-branch-data-sample-postgres-0-0   true         data-sample-postgres-0   1Gi           topolvm-provisioner-thin   56s            56s

$ kubectl get pvc -n demo data-dev-postgres-0 -o jsonpath='{.spec.dataSource}'
{"apiGroup":"snapshot.storage.k8s.io","kind":"VolumeSnapshot","name":"br-dev-branch-data-sample-postgres-0-0"}
```

## Connect to the Branch

Connect the same way you would to any KubeDB `PostgreSQL` database — through its Service `dev-postgres`, with the credentials in its own `dev-postgres-auth` Secret:

```bash
$ kubectl get secret -n demo dev-postgres-auth -o jsonpath='{.data.username}' | base64 -d
postgres
$ kubectl get secret -n demo dev-postgres-auth -o jsonpath='{.data.password}' | base64 -d   # prints the branch's password
```

For a quick look, exec into the branch's Pod:

```bash
$ kubectl exec -it -n demo dev-postgres-0 -c postgres -- psql -U postgres -c "SELECT count(*) FROM customers;"
 count
-------
  1000
(1 row)

$ kubectl exec -it -n demo dev-postgres-0 -c postgres -- psql -U postgres -c "SELECT * FROM customers ORDER BY id LIMIT 3;"
 id | full_name  |         email         |    phone
----+------------+-----------------------+-------------
  1 | Customer 1 | customer1@example.com | +1-555-0001
  2 | Customer 2 | customer2@example.com | +1-555-0002
  3 | Customer 3 | customer3@example.com | +1-555-0003
(3 rows)
```

All 1000 rows captured in the snapshot are present on the branch.

## Verify the Branch Is Isolated

The branch and the source share storage blocks, but only until either of them writes. Let's prove they are fully independent by writing a row to each and looking for it on the other side.

Write a row on the **branch**:

```bash
$ kubectl exec -it -n demo dev-postgres-0 -c postgres -- psql -U postgres \
    -c "INSERT INTO customers (full_name, email, phone) VALUES ('Dev Only', 'dev@example.com', '+1-555-9999');"
INSERT 0 1
```

Write a different row on the **source**:

```bash
$ kubectl exec -it -n demo sample-postgres-0 -c postgres -- psql -U postgres \
    -c "INSERT INTO customers (full_name, email, phone) VALUES ('Prod Only', 'prod@example.com', '+1-555-8888');"
INSERT 0 1
```

Neither row crosses over:

```bash
$ kubectl exec -it -n demo dev-postgres-0 -c postgres -- psql -U postgres \
    -c "SELECT count(*) AS visible_on_branch FROM customers WHERE full_name = 'Prod Only';"
 visible_on_branch
-------------------
                 0
(1 row)

$ kubectl exec -it -n demo sample-postgres-0 -c postgres -- psql -U postgres \
    -c "SELECT count(*) AS visible_on_source FROM customers WHERE full_name = 'Dev Only';"
 visible_on_source
-------------------
                 0
(1 row)
```

Both sides now hold 1001 rows — the same 1000 they started from, plus the one row each side added for itself:

```bash
$ kubectl exec -n demo sample-postgres-0 -c postgres -- psql -U postgres -tAc "SELECT count(*) FROM customers;"
1001
$ kubectl exec -n demo dev-postgres-0 -c postgres -- psql -U postgres -tAc "SELECT count(*) FROM customers;"
1001
```

The branch is a real, separate database. You can drop tables, run a destructive migration, or corrupt it entirely — the source is unaffected.

## Delete a Branch

What happens to the branched database when you delete the `Branch` is controlled by `spec.deletionPolicy`.

### `Delete` — remove everything (default)

`dev-branch` was created with the default policy, so deleting it takes the branched database with it:

```bash
$ kubectl delete branch -n demo dev-branch --wait=false
branch.courier.kubedb.com "dev-branch" deleted from demo namespace

# Courier completes Branch deletion asynchronously. Wait for its finalizer
# before checking the resources it cleaned up.
$ kubectl wait --for=delete branch/dev-branch -n demo --timeout=2m
```

Teardown removes the target database and everything the branch created:

```bash
$ kubectl get pg -n demo dev-postgres
Error from server (NotFound): postgreses.kubedb.com "dev-postgres" not found

$ kubectl get secret -n demo dev-postgres-auth
Error from server (NotFound): secrets "dev-postgres-auth" not found

$ kubectl get pvc -n demo data-dev-postgres-0
NAME                  STATUS        VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS               VOLUMEATTRIBUTESCLASS   AGE
data-dev-postgres-0   Terminating   pvc-355e402e-c8e7-4b4d-a9de-acb84c7a74e3   1Gi        RWO            topolvm-provisioner-thin   <unset>                 90s
```

The cloned volume finishes deleting asynchronously once the Pod releases it.

The branch's `VolumeSnapshot` is gone too:

```bash
$ kubectl get volumesnapshot -n demo
No resources found in demo namespace.
```

The source is untouched — still the 1001 rows we left it at:

```bash
$ kubectl exec -n demo sample-postgres-0 -c postgres -- psql -U postgres -tAc "SELECT count(*) FROM customers;"
1001
```

If the branch had post-actions configured, its `<target>-post-action` Job would be removed here too — see the [customization page](/docs/guides/postgres/branch/customization/index.md#mask-sensitive-data-before-handing-the-branch-out).

### `Orphan` — keep the branched database

Set `spec.deletionPolicy: Orphan` to promote the branch to a standalone database when the `Branch` goes away. This is how you graduate a branch into a permanent environment.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/branch/same-cluster/examples/dev-branch-keep.yaml
branch.courier.kubedb.com/dev-branch-keep created

# Wait until the target exists before recording its UID.
$ kubectl wait --for=jsonpath='{.status.phase}'=Ready branch/dev-branch-keep -n demo --timeout=5m
```

Note the target's UID before deleting the `Branch`, so we can prove it is the same object afterwards and not a re-created one:

```bash
$ kubectl get pg -n demo dev-postgres-keep -o jsonpath='{.metadata.uid}'
d0a85da7-ced2-4e94-b768-bd5622cdbf06

$ kubectl delete branch -n demo dev-branch-keep --wait=false
branch.courier.kubedb.com "dev-branch-keep" deleted from demo namespace

# Wait for Courier to finish removing the Branch before checking the promoted target.
$ kubectl wait --for=delete branch/dev-branch-keep -n demo --timeout=2m
```

The database survives, with its branch ownership metadata stripped:

```bash
$ kubectl get pg -n demo dev-postgres-keep
NAME                VERSION   STATUS   AGE
dev-postgres-keep   16.1      Ready    41s

$ kubectl get pg -n demo dev-postgres-keep \
    -o jsonpath='uid={.metadata.uid}{"\n"}branched-from={.metadata.annotations.kubedb\.com/branched-from}{"\n"}ownerRefs={.metadata.ownerReferences}{"\n"}'
uid=d0a85da7-ced2-4e94-b768-bd5622cdbf06
branched-from=
ownerRefs=
```

Same UID — it was promoted in place, not re-created — with no `branched-from` annotation and no owner reference. Its volume and credential are kept, so it keeps serving:

```bash
$ kubectl get secret -n demo dev-postgres-keep-auth
NAME                     TYPE                       DATA   AGE
dev-postgres-keep-auth   kubernetes.io/basic-auth   2      41s

$ kubectl exec -n demo dev-postgres-keep-0 -c postgres -- psql -U postgres -tAc "SELECT count(*) FROM customers;"
1001
```

From here it is an ordinary KubeDB `Postgres` — delete it with `kubectl delete pg` when you are done.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `Branch` stuck in `Snapshotting`; the `VolumeSnapshot` never becomes `readyToUse` | No `VolumeSnapshotClass` for the source volume's CSI driver, or the driver does not support snapshots. Check `kubectl describe volumesnapshot -n <ns> br-<branch>-…` and set `spec.volumeSnapshotClassName` explicitly. |
| `Branch` stuck in `Cloning`; the target PVC stays `Pending` | The target StorageClass is backed by a different CSI driver than the source, so the driver cannot restore the snapshot. Point `spec.target.storageClassName` at a class using the same driver. |
| `Branch` stuck in `Provisioning` | The branched database is not coming up. Inspect it like any other KubeDB database: `kubectl describe pg -n <ns> <target>` and check its Pod logs. |
| A one-shot `Branch` stays `Failed` after you fixed the underlying problem | Without `spec.schedule` a failure is terminal by design. Edit the `Branch` (or delete and re-create it) to trigger a retry. |
| The branch shows stale data | A branch is a point-in-time copy. Add [`spec.schedule.cron`](/docs/guides/postgres/branch/customization/index.md#keep-the-branch-fresh) to refresh it, or re-create it. |

For problems with the optional fields — post-actions, scheduled refresh, cross-namespace targets — see the [customization page's troubleshooting table](/docs/guides/postgres/branch/customization/index.md#troubleshooting).

## Cleaning up

```bash
$ kubectl delete branch -n demo --all
$ kubectl delete pg -n demo dev-postgres-keep
$ kubectl delete pg -n demo sample-postgres
$ kubectl delete ns demo
```

Deleting a `Branch` with the default `Delete` policy removes the branched database, its volumes, its credential, the snapshots, and any post-action Jobs. The database left behind by the `Orphan` policy — and the source database itself — have to be deleted directly, as above.

## Next Steps

- Customize your branch: [a separate credential, post-actions, scheduled refresh, cross-namespace targets, HA, and snapshot-class pinning](/docs/guides/postgres/branch/customization/index.md).
- Need the branch on a *different* Kubernetes cluster? See [Cross-Cluster Branching](/docs/guides/postgres/branch/cross-cluster/index.md).
- Read the [PostgreSQL Branching Overview](/docs/guides/postgres/branch/overview/index.md) for how branching works under the hood.
- See every field of the [Branch CRD](/docs/guides/postgres/concepts/branch.md).
- Learn about [PostgreSQL backup and restore](/docs/guides/postgres/backup/kubestash/overview/index.md) — branching is for disposable copies, backups are for durability. They solve different problems.
- Migrating an existing database *into* KubeDB instead? See the [PostgreSQL Database Migration guide](/docs/guides/postgres/migration/databasemigration.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
