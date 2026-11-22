---
title: Refreshing a MariaDB Branch
menu:
  docs_{{ .version }}:
    identifier: mariadb-branch-refresh
    name: Refresh
    parent: guides-mariadb-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a MariaDB Branch

A MariaDB branch starts as a point-in-time copy of its source. It does not continuously receive source changes. Set `spec.schedule.cron` to periodically replace the target data with a fresh source snapshot, keeping development or test environments current.

This tutorial demonstrates a same-cluster MariaDB branch in a different namespace. You will deploy and seed a source, create a scheduled Branch, change data on both databases, and verify what the next refresh preserves and replaces. No database from another tutorial is required.

## What happens during refresh

1. At the scheduled time, Courier advances the refresh generation and takes fresh source snapshots. It keeps the existing target running until the new snapshots are ready.
2. For a target in another namespace, Courier imports the snapshot references into that namespace.
3. Courier pauses the target, removes its old cloned volumes, and creates new volumes from the fresh snapshots.
4. Courier unpauses the target, and KubeDB brings the database back up using the new volumes.
5. After the target and any configured preparation steps complete, the Branch returns to `Ready`. Courier releases the previous generation's snapshots.

If configured, [root password reset](/docs/guides/mariadb/branch/reset-root-password.md) and [post-actions](/docs/guides/mariadb/branch/post-actions.md) run again on every refresh before the Branch returns to `Ready`.

**Refresh discards branch-only writes and makes the target temporarily unavailable while its storage is replaced and the database restarts.** It replaces the branch's data; it does not merge changes back into the source. The source continues to serve its own workload. Choose a schedule that fits the target application's maintenance requirements.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md) to understand how branches work.
- Complete the [single-cluster Courier installation](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass supporting copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure MariaDB version `12.3.2` is available in the installed KubeDB catalog, and use branching-capable KubeDB and database images.
- Run the commands against the same Kubernetes context. Your account must be able to create databases and Branches, inspect their status and pods, and open a shell in database pods.

The manifests use the StorageClass name `ceph-rbd`. Substitute your Ceph RBD class consistently in both manifests. The source lives in `demo` and the target in `dev`; reserve the database names below for this tutorial.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `mariadb-refresh-source.yaml`. It creates a three-replica Galera MariaDB source named `mariadb-refresh-source`:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-refresh-source
  namespace: demo
spec:
  version: "12.3.2"
  replicas: 3
  topology:
    mode: GaleraCluster
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
kubectl apply -f mariadb-refresh-source.yaml
kubectl -n demo wait mariadb/mariadb-refresh-source \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Read the source database's auth Secret name from your terminal:

```shell
kubectl -n demo get mariadb mariadb-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mariadb-refresh-source-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl -n demo get secret mariadb-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the source connections below. The MariaDB root username is `root`.

## Seed and inspect the source data

From your terminal, open the source pod shell:

```shell
kubectl -n demo exec -it mariadb-refresh-source-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the source password at the prompt:

```shell
mariadb -u root -p
```

At the MariaDB prompt, create five customer records:

```sql
CREATE DATABASE branchdb;
USE branchdb;

CREATE TABLE customers (
    id integer PRIMARY KEY,
    full_name varchar(100) NOT NULL,
    email varchar(255) NOT NULL
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
+----+------------+-----------------------+
| id | full_name  | email                 |
+----+------------+-----------------------+
|  1 | Customer 1 | customer1@example.com |
|  2 | Customer 2 | customer2@example.com |
|  3 | Customer 3 | customer3@example.com |
|  4 | Customer 4 | customer4@example.com |
|  5 | Customer 5 | customer5@example.com |
+----+------------+-----------------------+
5 rows in set
```

Leave `mariadb` with `\q`, then run `exit` in the pod shell to return to your terminal.

## Create a scheduled Branch

Save the following as `branch-refresh.yaml`. The Branch lives in `demo` and creates `dev/mariadb-refresh-branch`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mariadb-refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MariaDB
      name: mariadb-refresh-source
    namespace: demo
  target:
    namespace: dev
    name: mariadb-refresh-branch
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
kubectl -n demo wait branch/mariadb-refresh-branch \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
kubectl -n demo get branch mariadb-refresh-branch
```

Illustrative output:

```text
NAME                     PHASE   MODE    TARGET                   FRESHNESS   AGE
mariadb-refresh-branch   Ready   Local   mariadb-refresh-branch   2s          3m
```

Outputs below illustrate successful runs; pod names, timestamps, ages, and durations vary between environments.

### Check the initial copy

Read the branch database's auth Secret name from your terminal:

```shell
kubectl -n dev get mariadb mariadb-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mariadb-refresh-branch-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl -n dev get secret mariadb-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the branch connections below. The MariaDB root username is `root`.

Open the branch pod shell from your terminal:

```shell
kubectl -n dev exec -it mariadb-refresh-branch-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the branch password at the prompt:

```shell
mariadb -u root -p branchdb
```

Query the branch:

```sql
SELECT * FROM customers ORDER BY id;
```

It should contain the same five rows:

```text
+----+------------+-----------------------+
| id | full_name  | email                 |
+----+------------+-----------------------+
|  1 | Customer 1 | customer1@example.com |
|  2 | Customer 2 | customer2@example.com |
|  3 | Customer 3 | customer3@example.com |
|  4 | Customer 4 | customer4@example.com |
|  5 | Customer 5 | customer5@example.com |
+----+------------+-----------------------+
5 rows in set
```

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

### Check the next refresh time

Before changing either database, inspect the current generation and schedule:

```shell
kubectl -n demo get branch mariadb-refresh-branch \
  -o custom-columns='PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT-REFRESH:.status.nextRefreshTime'
```

For an initial copy completed at 10:02 UTC, with Courier using UTC:

```text
PHASE   GENERATION   LAST-SUCCESS           NEXT-REFRESH
Ready   <none>       2026-11-11T10:02:00Z   2026-11-11T10:10:00Z
```

The initial generation is `0`; it may appear as `<none>` because the API omits zero values. Record the generation and successful completion time. If the next deadline is close, let that cycle finish before starting the following steps, then record the new values. Make both edits and inspect the data before the next refresh starts.

## Change source and branch data

### Update the source

```shell
kubectl -n demo exec -it mariadb-refresh-source-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the source password at the prompt:

```shell
mariadb -u root -p branchdb
```

Change customer `1` on the source and inspect all five rows:

```sql
UPDATE customers SET full_name = 'Updated Customer 1' WHERE id = 1;
SELECT * FROM customers ORDER BY id;
```

Expected result:

```text
Query OK, 1 row affected
+----+--------------------+-----------------------+
| id | full_name          | email                 |
+----+--------------------+-----------------------+
|  1 | Updated Customer 1 | customer1@example.com |
|  2 | Customer 2         | customer2@example.com |
|  3 | Customer 3         | customer3@example.com |
|  4 | Customer 4         | customer4@example.com |
|  5 | Customer 5         | customer5@example.com |
+----+--------------------+-----------------------+
5 rows in set
```

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

### Make a branch-only edit

```shell
kubectl -n dev exec -it mariadb-refresh-branch-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the branch password at the prompt:

```shell
mariadb -u root -p branchdb
```

Change customer `5` on the branch, then inspect the data:

```sql
UPDATE customers SET full_name = 'Branch-only Customer 5' WHERE id = 5;
SELECT * FROM customers ORDER BY id;
```

Before refresh, the branch should return:

```text
Query OK, 1 row affected
+----+------------------------+-----------------------+
| id | full_name              | email                 |
+----+------------------------+-----------------------+
|  1 | Customer 1             | customer1@example.com |
|  2 | Customer 2             | customer2@example.com |
|  3 | Customer 3             | customer3@example.com |
|  4 | Customer 4             | customer4@example.com |
|  5 | Branch-only Customer 5 | customer5@example.com |
+----+------------------------+-----------------------+
5 rows in set
```

Customer `1` still has its original name: the source update has not reached the branch. Customer `5` contains the branch-only edit that the next refresh will discard.

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

## Wait for the next completed refresh

Watch the Branch from your terminal:

```shell
kubectl -n demo get branch mariadb-refresh-branch \
  -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' \
  --watch
```

An initial generation `0` followed by a successful refresh might look like:

```text
NAME                     PHASE        GENERATION   LAST-SUCCESS
mariadb-refresh-branch   Ready        <none>       2026-11-11T10:02:00Z
mariadb-refresh-branch   Refreshing   1            2026-11-11T10:02:00Z
mariadb-refresh-branch   Ready        1            2026-11-11T10:11:30Z
```

Wait until the generation is later than the one you recorded, the phase is `Ready`, and `LAST-SUCCESS` has advanced. Then press **Ctrl+C** to stop watching. Intermediate events may repeat or pass too quickly to observe. Do not treat the initial `Ready` row or a generation increase alone as a completed refresh.

Inspect the full status:

```shell
kubectl -n demo get branch mariadb-refresh-branch -o yaml
```

For the example above, the refresh-related status would include:

```yaml
status:
  history:
    - at: "2026-11-11T10:02:00Z"
      result: Succeeded
    - at: "2026-11-11T10:11:30Z"
      result: Succeeded
  lastRefreshTime: "2026-11-11T10:11:30Z"
  lastSuccessfulRefreshTime: "2026-11-11T10:11:30Z"
  mode: Local
  nextRefreshTime: "2026-11-11T10:20:00Z"
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

Refresh restarts the target. Open a new pod shell after the Branch returns to `Ready`:

```shell
kubectl -n dev exec -it mariadb-refresh-branch-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the branch password at the prompt:

```shell
mariadb -u root -p branchdb
```

Inspect all five customers:

```sql
SELECT * FROM customers ORDER BY id;
```

Expected result:

```text
+----+--------------------+-----------------------+
| id | full_name          | email                 |
+----+--------------------+-----------------------+
|  1 | Updated Customer 1 | customer1@example.com |
|  2 | Customer 2         | customer2@example.com |
|  3 | Customer 3         | customer3@example.com |
|  4 | Customer 4         | customer4@example.com |
|  5 | Customer 5         | customer5@example.com |
+----+--------------------+-----------------------+
5 rows in set
```

The source update to customer `1` is now present. The branch-only edit to customer `5` is gone, and the table still contains five rows.

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

### Confirm the source was not changed by refresh

```shell
kubectl -n demo exec -it mariadb-refresh-source-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the source password at the prompt:

```shell
mariadb -u root -p branchdb
```

Query the source:

```sql
SELECT * FROM customers ORDER BY id;
```

It should still contain the source-side update and no branch-only edit:

```text
+----+--------------------+-----------------------+
| id | full_name          | email                 |
+----+--------------------+-----------------------+
|  1 | Updated Customer 1 | customer1@example.com |
|  2 | Customer 2         | customer2@example.com |
|  3 | Customer 3         | customer3@example.com |
|  4 | Customer 4         | customer4@example.com |
|  5 | Customer 5         | customer5@example.com |
+----+--------------------+-----------------------+
5 rows in set
```

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

For field definitions, defaults, and cron syntax, see the common [schedule and history reference](/docs/operatormanual/branching/refresh.md#schedule-and-history-field-reference). The [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) explains the fields reported by Courier.

## Apply the same strategy across clusters

The same schedule and history settings apply to a cross-cluster Branch. After completing [cross-cluster Courier setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm), use `spec.target.clusterName` to select the registered target cluster and apply the original Branch on the source cluster. See the [MariaDB cross-cluster walkthrough](/docs/guides/mariadb/branch/cross-cluster/index.md) for the complete configuration.

The source Courier agent owns the schedule and takes the next generation's snapshots. OCM delivers the new generation through the hub, the target agent performs replacement, and completion is reported back to the source Branch. Inspect scheduling and freshness on the source cluster and target pods and data on the target cluster. Both agents and the hub delivery path must remain available.

Both clusters must use **the same Ceph backend, reachable over the network**, with compatible CSI configuration and snapshot support. Matching StorageClass names alone are insufficient; the target must be able to resolve the source snapshot handles. See [cross-cluster backend compatibility](/docs/operatormanual/branching/snapshots.md#cross-cluster-backend-compatibility).

## Cleanup

Delete the scheduled Branch first and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch mariadb-refresh-branch --wait=true --timeout=15m
```

With `deletionPolicy: Delete`, Courier removes the target database, cloned volumes, auth Secret, and managed snapshots. The source remains available.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete mariadb mariadb-refresh-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources.
