---
title: Refreshing Database Branches
menu:
  docs_{{ .version }}:
    identifier: branching-refresh
    name: Refresh
    parent: branching
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing Database Branches

A branch starts as a point-in-time copy of its source. It does not continuously receive source changes. Set `spec.schedule.cron` to periodically replace the target data with a fresh source snapshot, keeping development or test environments current.

Refresh applies to all supported database engines. This tutorial uses PostgreSQL to demonstrate a same-cluster branch in a different namespace. You will deploy and seed a source, create a scheduled Branch, change data on both databases, and verify what the next refresh preserves and replaces. No database from another tutorial is required.

## What happens during refresh

1. At the scheduled time, Courier advances the refresh generation and takes fresh source snapshots. It keeps the existing target running until the new snapshots are ready.
2. For a target in another namespace, Courier imports the snapshot references into that namespace.
3. Courier pauses the target, removes its old cloned volumes, and creates new volumes from the fresh snapshots.
4. Courier unpauses the target, and KubeDB brings the database back up using the new volumes.
5. After the target and any configured preparation steps complete, the Branch returns to `Ready`. Courier releases the previous generation's snapshots.

If configured, [root password reset](/docs/operatormanual/branching/reset-root-password.md) and [post-actions](/docs/operatormanual/branching/post-actions.md) run again on every refresh before the Branch returns to `Ready`.

**Refresh discards branch-only writes and makes the target temporarily unavailable while its storage is replaced and the database restarts.** It replaces the branch's data; it does not merge changes back into the source. The source continues to serve its own workload. Choose a schedule that fits the target application's maintenance requirements.

## Before You Begin

- Complete the [single-cluster Courier installation](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass supporting copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure PostgreSQL version `18.3` is available in the installed KubeDB catalog, and use branching-capable KubeDB and database images.
- Run the commands against the same Kubernetes context. Your account must be able to create databases and Branches, inspect their status and pods, and open a shell in database pods.

The manifests use the StorageClass name `ceph-rbd`. Substitute your Ceph RBD class consistently in both manifests. The source lives in `demo` and the target in `dev`; reserve the database names below for this tutorial.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `refresh-source.yaml`. It creates a three-replica PostgreSQL source named `refresh-source`:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: refresh-source
  namespace: demo
spec:
  version: "18.3"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Apply it and wait for the source to become ready:

```shell
kubectl apply -f refresh-source.yaml
kubectl -n demo wait postgres/refresh-source \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Find the source primary:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=refresh-source,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Example output:

```text
refresh-source-0
```

Use the returned pod name in the following commands; this example uses `refresh-source-0`.

## Seed and inspect the source data

From your terminal, open a shell in the source primary:

```shell
kubectl -n demo exec -it refresh-source-0 -c postgres -- bash
```

Inside the pod, connect to database `postgres`. Local socket connections in this setup use trust authentication, so no password is required:

```shell
psql -U postgres -d postgres
```

At the `postgres=#` prompt, create five customer records:

```sql
CREATE TABLE customers (
    id integer PRIMARY KEY,
    full_name text NOT NULL,
    email text NOT NULL
);
INSERT INTO customers (id, full_name, email) VALUES
    (1, 'Customer 1', 'customer1@example.com'),
    (2, 'Customer 2', 'customer2@example.com'),
    (3, 'Customer 3', 'customer3@example.com'),
    (4, 'Customer 4', 'customer4@example.com'),
    (5, 'Customer 5', 'customer5@example.com');
```

Inspect the source data:

```sql
SELECT * FROM customers ORDER BY id;
```

Expected result:

```text
 id | full_name  |         email
----+------------+-----------------------
  1 | Customer 1 | customer1@example.com
  2 | Customer 2 | customer2@example.com
  3 | Customer 3 | customer3@example.com
  4 | Customer 4 | customer4@example.com
  5 | Customer 5 | customer5@example.com
(5 rows)
```

Leave `psql` with `\q`, then run `exit` in the pod shell to return to your terminal.

## Create a scheduled Branch

Save the following as `branch-refresh.yaml`. The Branch lives in `demo` and creates `dev/refresh-branch`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: refresh-source
    namespace: demo
  target:
    namespace: dev
    name: refresh-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  schedule:
    cron: "*/10 * * * *"
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

The initial copy starts when you create the Branch. After it becomes ready, `*/10 * * * *` schedules refreshes at ten-minute clock boundaries: `:00`, `:10`, `:20`, and so on. It does not mean ten minutes after creation. For example, a branch ready at 10:02 has its next refresh scheduled for 10:10 in the Courier controller's timezone.

Apply the manifest and wait for the initial copy:

```shell
kubectl apply -f branch-refresh.yaml
kubectl -n demo wait branch/refresh-branch \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
kubectl -n demo get branch refresh-branch
```

Illustrative output:

```text
NAME             PHASE   MODE    TARGET           FRESHNESS   AGE
refresh-branch   Ready   Local   refresh-branch   2s          3m
```

Outputs below illustrate successful runs; pod names, timestamps, ages, and durations vary between environments.

### Check the initial copy

Find the target primary from your terminal:

```shell
kubectl -n dev get pods -l app.kubernetes.io/instance=refresh-branch,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Use the returned name below; this example uses `refresh-branch-0`:

```shell
kubectl -n dev exec -it refresh-branch-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Query the branch:

```sql
SELECT * FROM customers ORDER BY id;
```

It should contain the same five rows:

```text
 id | full_name  |         email
----+------------+-----------------------
  1 | Customer 1 | customer1@example.com
  2 | Customer 2 | customer2@example.com
  3 | Customer 3 | customer3@example.com
  4 | Customer 4 | customer4@example.com
  5 | Customer 5 | customer5@example.com
(5 rows)
```

Leave `psql` with `\q`, then run `exit` to return to your terminal.

### Check the next refresh time

Before changing either database, inspect the current generation and schedule:

```shell
kubectl -n demo get branch refresh-branch \
  -o custom-columns='PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT-REFRESH:.status.nextRefreshTime'
```

For an initial copy completed at 10:02 UTC, with Courier using UTC:

```text
PHASE   GENERATION   LAST-SUCCESS           NEXT-REFRESH
Ready   <none>       2026-11-10T10:02:00Z   2026-11-10T10:10:00Z
```

The initial generation is `0`; it may appear as `<none>` because the API omits zero values. Record the generation and successful completion time. If the next deadline is close, let that cycle finish before starting the following steps, then record the new values. Make both edits and inspect the data before the next refresh starts.

## Change source and branch data

### Update the source

From your terminal, find the source primary again in case a failover occurred:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=refresh-source,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open a shell in that pod, substituting its name if different:

```shell
kubectl -n demo exec -it refresh-source-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Change customer `1` on the source and inspect all five rows:

```sql
UPDATE customers SET full_name = 'Updated Customer 1' WHERE id = 1;
SELECT * FROM customers ORDER BY id;
```

Expected result:

```text
UPDATE 1
 id |     full_name      |         email
----+--------------------+-----------------------
  1 | Updated Customer 1 | customer1@example.com
  2 | Customer 2         | customer2@example.com
  3 | Customer 3         | customer3@example.com
  4 | Customer 4         | customer4@example.com
  5 | Customer 5         | customer5@example.com
(5 rows)
```

Leave `psql` with `\q`, then run `exit` to return to your terminal.

### Make a branch-only edit

Find the target primary again:

```shell
kubectl -n dev get pods -l app.kubernetes.io/instance=refresh-branch,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open its pod shell, using the returned name:

```shell
kubectl -n dev exec -it refresh-branch-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Change customer `5` on the branch, then inspect the data:

```sql
UPDATE customers SET full_name = 'Branch-only Customer 5' WHERE id = 5;
SELECT * FROM customers ORDER BY id;
```

Before refresh, the branch should return:

```text
UPDATE 1
 id |       full_name        |         email
----+------------------------+-----------------------
  1 | Customer 1             | customer1@example.com
  2 | Customer 2             | customer2@example.com
  3 | Customer 3             | customer3@example.com
  4 | Customer 4             | customer4@example.com
  5 | Branch-only Customer 5 | customer5@example.com
(5 rows)
```

Customer `1` still has its original name: the source update has not reached the branch. Customer `5` contains the branch-only edit that the next refresh will discard.

Leave `psql` with `\q`, then run `exit` to return to your terminal.

## Wait for the next completed refresh

Watch the Branch from your terminal:

```shell
kubectl -n demo get branch refresh-branch \
  -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' \
  --watch
```

An initial generation `0` followed by a successful refresh might look like:

```text
NAME             PHASE        GENERATION   LAST-SUCCESS
refresh-branch   Ready        <none>       2026-11-10T10:02:00Z
refresh-branch   Refreshing   1            2026-11-10T10:02:00Z
refresh-branch   Ready        1            2026-11-10T10:11:30Z
```

Wait until the generation is later than the one you recorded, the phase is `Ready`, and `LAST-SUCCESS` has advanced. Then press **Ctrl+C** to stop watching. Intermediate events may repeat or pass too quickly to observe. Do not treat the initial `Ready` row or a generation increase alone as a completed refresh.

Inspect the full status:

```shell
kubectl -n demo get branch refresh-branch -o yaml
```

For the example above, the refresh-related status would include:

```yaml
status:
  history:
    - at: "2026-11-10T10:02:00Z"
      result: Succeeded
    - at: "2026-11-10T10:11:30Z"
      result: Succeeded
  lastRefreshTime: "2026-11-10T10:11:30Z"
  lastSuccessfulRefreshTime: "2026-11-10T10:11:30Z"
  mode: Local
  nextRefreshTime: "2026-11-10T10:20:00Z"
  phase: Ready
  refreshGeneration: 1
  snapshot:
    generation: 1
    ready: true
    strategy: VolumeSnapshot
```

The two history entries represent the initial copy and one successful refresh. The current snapshot generation matches the refreshed branch. The snapshot strategy may differ with your storage configuration; this excerpt omits the snapshot members and other status fields.

The `FRESHNESS` column in the default `kubectl get branch` output measures time since `lastSuccessfulRefreshTime`. It increases between successful runs and resets when another run completes. It is a completion-age indicator; use `status.snapshot.members[].creationTime`, when supplied by the CSI driver, to inspect when source snapshots were actually taken.

## Verify the refreshed data

### Check the branch

Refresh replaces storage and restarts the target. From your terminal, discover the current target primary rather than reusing an existing pod session:

```shell
kubectl -n dev get pods -l app.kubernetes.io/instance=refresh-branch,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open its pod shell, substituting the returned name:

```shell
kubectl -n dev exec -it refresh-branch-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Inspect all five customers:

```sql
SELECT * FROM customers ORDER BY id;
```

Expected result:

```text
 id |     full_name      |         email
----+--------------------+-----------------------
  1 | Updated Customer 1 | customer1@example.com
  2 | Customer 2         | customer2@example.com
  3 | Customer 3         | customer3@example.com
  4 | Customer 4         | customer4@example.com
  5 | Customer 5         | customer5@example.com
(5 rows)
```

The source update to customer `1` is now present. The branch-only edit to customer `5` is gone, and the table still contains five rows.

Leave `psql` with `\q`, then run `exit` to return to your terminal.

### Confirm the source was not changed by refresh

Find the source primary again:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=refresh-source,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open its pod shell, using the returned name:

```shell
kubectl -n demo exec -it refresh-source-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Query the source:

```sql
SELECT * FROM customers ORDER BY id;
```

It should still contain the source-side update and no branch-only edit:

```text
 id |     full_name      |         email
----+--------------------+-----------------------
  1 | Updated Customer 1 | customer1@example.com
  2 | Customer 2         | customer2@example.com
  3 | Customer 3         | customer3@example.com
  4 | Customer 4         | customer4@example.com
  5 | Customer 5         | customer5@example.com
(5 rows)
```

Leave `psql` with `\q`, then run `exit` to return to your terminal.

## Schedule and history field reference

| Field | Type | Default or omitted behavior | Purpose |
|---|---|---|---|
| `spec.schedule` | Object | Omitted: create a one-time branch without scheduled refreshes. | Enables periodic replacement of branch data. |
| `spec.schedule.cron` | String | Required when `schedule` is present; no default expression. | Sets the refresh cadence. The example uses `*/10 * * * *`. |
| `spec.historyLimit` | Object | Omitted: keep the last `3` successful and `2` failed run records. | Bounds entries in `status.history` independently for each outcome. |
| `spec.historyLimit.success` | Integer (int32) | `3` when omitted. | Number of successful run records to keep, including initial creation. Set `0` to retain none. |
| `spec.historyLimit.failed` | Integer (int32) | `2` when omitted. | Number of failed run records to keep. Set `0` to retain none. |

History limits retain **outcome records**, not database copies or snapshots for rollback. With the example limits, status can hold up to five entries: three successful and two failed. A fourth success removes the oldest success without consuming the failed-run allowance. Snapshot cleanup follows the refresh lifecycle independently of these limits.

### Cron syntax and timezone

Use five fields separated by spaces:

```text
minute hour day-of-month month day-of-week
```

| Expression | Schedule |
|---|---|
| `*/10 * * * *` | Every ten minutes, at minutes `0`, `10`, `20`, `30`, `40`, and `50`. |
| `0 2 * * *` | Every day at 02:00. |
| `0 2 * * 1` | Every Monday at 02:00. |

Without an explicit timezone prefix, expressions use the Courier controller process's local timezone, which may differ from your workstation. The parser also accepts a prefix such as `CRON_TZ=Etc/UTC 0 2 * * *` to specify a timezone. There is no separate `spec.schedule.timeZone` field.

After a run completes, Courier computes the next future cron deadline. Refreshes run sequentially for a Branch; a short interval does not make an in-progress replacement complete sooner or start a second replacement concurrently.

## Refresh status reference

These fields are written by Courier; do not set them in your manifest.

| Field | Type | When absent or initial | Meaning |
|---|---|---|---|
| `status.phase` | String | May be absent before reconciliation. | Current lifecycle phase. A scheduled local replacement uses `Refreshing`, then returns to `Ready` on success. |
| `status.refreshGeneration` | Integer (int64) | Initial generation is `0`, which may be omitted from output. | Advances when a scheduled refresh begins. It identifies an attempt, not proof of successful completion. |
| `status.lastRefreshTime` | Timestamp | Absent until a run outcome is recorded. | Time the latest attempt's outcome was recorded, whether successful or failed. It is not the refresh start time. |
| `status.lastSuccessfulRefreshTime` | Timestamp | Absent until the first successful copy completes. | Latest successful completion time, including initial creation. Drives `FRESHNESS`; a failed attempt does not advance it. |
| `status.nextRefreshTime` | Timestamp | Absent for a one-time branch or before the next deadline is computed. | Next scheduled refresh deadline. It does not predict when replacement will finish. |
| `status.history` | Array of run records | Empty before recorded outcomes, or when both limits are `0`. | Bounded successful and failed outcomes, ordered oldest to newest. |
| `status.history[].at` | Timestamp | Recorded with the outcome. | Time that run finished successfully or was recorded as failed. |
| `status.history[].result` | String | Recorded with the outcome. | `Succeeded` or `Failed`. |
| `status.history[].message` | String | Optional; normally absent on success. | Human-readable detail about a failed attempt. |
| `status.snapshot.generation` | Integer (int64) | Generation `0` may be omitted. | Generation of the source snapshot set recorded for the branch. Compare it with the refresh generation and completion status. |

## Apply the same strategy across clusters

The same schedule and history settings apply to a cross-cluster Branch. After completing [cross-cluster Courier setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm), use `spec.target.clusterName` to select the registered target cluster and apply the original Branch on the source cluster. See the [PostgreSQL cross-cluster walkthrough](/docs/guides/postgres/branch/cross-cluster/index.md) for the complete configuration.

The source Courier agent owns the schedule and takes the next generation's snapshots. OCM delivers the new generation through the hub, the target agent performs replacement, and completion is reported back to the source Branch. Inspect scheduling and freshness on the source cluster and target pods and data on the target cluster. Both agents and the hub delivery path must remain available.

Both clusters must use **the same Ceph backend, reachable over the network**, with compatible CSI configuration and snapshot support. Matching StorageClass names alone are insufficient; the target must be able to resolve the source snapshot handles. See [cross-cluster backend compatibility](/docs/operatormanual/branching/snapshots.md#cross-cluster-backend-compatibility).

## Cleanup

Delete the scheduled Branch first and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch refresh-branch --wait=true --timeout=15m
```

With `deletionPolicy: Delete`, Courier removes the target database, cloned volumes, auth Secret, and managed snapshots. The source remains available.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete postgres refresh-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources.
