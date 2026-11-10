---
title: Branch PostgreSQL Across Clusters
menu:
  docs_{{ .version }}:
    identifier: pg-branch-cross-cluster
    name: Cross Cluster
    parent: guides-pg-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch PostgreSQL Across Clusters

This guide creates an HA PostgreSQL 18.3 source on the `source` managed cluster and branches it into namespace `dev` on `target`. Both clusters access the **same network-reachable Ceph RBD backend** with copy-on-write snapshot and clone support.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md) to understand how branches work.
- Complete [Courier cross-cluster setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm) in an existing OCM environment.
- Have `kubectl` contexts named `hub`, `source`, and `target`. The OCM ManagedCluster names in this tutorial are `source` and `target`.
- Install compatible branching-capable KubeDB and database images with valid licenses on both database clusters. PostgreSQL version `18.3` must be available in both installed KubeDB catalogs.
- Configure compatible Ceph RBD storage and snapshot classes on both clusters. Source snapshots must be resolvable on the target against the shared backend. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

Using two independent Ceph backends does not satisfy the storage requirement. Confirm [CSI/backend compatibility](/docs/operatormanual/branching/snapshots.md#cross-cluster-backend-compatibility) before creating the source.

The manifests use the StorageClass name `ceph-rbd`; substitute your Ceph RBD class consistently. Your account must be able to create databases and Branches, inspect their resources, and open a shell in the database pods. Commands identify the cluster with `--context source`, `--context target`, or `--context hub`.

Create the namespaces if they do not already exist:

```shell
kubectl --context source create namespace demo
kubectl --context target create namespace dev
```

## Deploy the source database

Save the following as `source-postgres.yaml`. It creates a three-replica PostgreSQL source named `prod-pg`:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: prod-pg
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

Apply it on the source cluster and wait for readiness:

```shell
kubectl --context source apply -f source-postgres.yaml
kubectl --context source -n demo wait postgres/prod-pg \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Find the elected primary:

```shell
kubectl --context source -n demo get pods -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Example output:

```text
prod-pg-0
```

Use the returned pod name below if it differs. Pod names, ages, and durations in this guide are illustrative.

## Seed and inspect the source data

From your terminal, open a shell in the source primary:

```shell
kubectl --context source -n demo exec -it prod-pg-0 -c postgres -- bash
```

Inside the pod, connect to database `postgres`. Local socket connections in this setup use trust authentication, so no password is required:

```shell
psql -U postgres -d postgres
```

At the `postgres=#` create five customer records:

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

## Create the cross-cluster Branch

Save the following as `branch-cross-cluster.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-cross
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
    namespace: demo
  target:
    clusterName: target
    namespace: dev
    name: pg-cross
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

`spec.target.clusterName` is the target's **OCM ManagedCluster name**, which can differ from a kubeconfig context name. Apply the original Branch on the **source** cluster. Courier routes the request through the hub; on the target, Courier prepares cloned PVCs and the PostgreSQL resource, and KubeDB provisions the database using those PVCs.

`spec.target.storageClassName` selects the target Ceph class; `spec.target.resources` sets target PostgreSQL container requests. See the [PostgreSQL Branch specification](/docs/guides/postgres/branch/overview/index.md) for field details.

Apply the manifest and wait for the **Branch** to become ready:

```shell
kubectl --context source apply -f branch-cross-cluster.yaml
kubectl --context source -n demo wait branch/pg-cross \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
kubectl --context source -n demo get branch pg-cross
```

Representative output:

```text
NAME       PHASE   MODE        TARGET     FRESHNESS   AGE
pg-cross   Ready   Initiator   pg-cross   2s          3m
```

Inspect the target resources on the target cluster:

```shell
kubectl --context target -n dev get postgres pg-cross
kubectl --context target -n dev get pvc,volumesnapshot
```

To inspect delivery and target execution from your terminal:

```shell
kubectl --context hub -n source get branchwork
kubectl --context hub -n target get manifestwork
kubectl --context target -n dev get branch
```

Manage the branch through its original source resource. Courier owns the delivered target resources.

## Verify the copied data

From your terminal, find the branch primary on the target cluster:

```shell
kubectl --context target -n dev get pods -l app.kubernetes.io/instance=pg-cross,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Example output:

```text
pg-cross-0
```

Use the returned name to open its pod shell:

```shell
kubectl --context target -n dev exec -it pg-cross-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Query the copied data:

```sql
SELECT * FROM customers ORDER BY id;
```

The branch should contain the same five rows:

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

Keep this `psql` session open for the next step. These local connections verify data; password-isolation checks are covered in the [root password reset tutorial](/docs/guides/postgres/branch/reset-root-password.md).

## Verify independent writes

### Change the branch

In the same branch `psql` session, update customer `5`:

```sql
UPDATE customers SET full_name = 'Branch-only Customer 5' WHERE id = 5;
```

Expected output:

```text
UPDATE 1
```

Leave `psql` with `\q`, then run `exit` to return to your terminal.

### Change and inspect the source

Find the source primary again in case a failover occurred:

```shell
kubectl --context source -n demo get pods -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open its pod shell, substituting the returned name if different:

```shell
kubectl --context source -n demo exec -it prod-pg-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Update customer `1` and inspect all five rows:

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

Customer `5` retains its original value on the source: the branch-only edit did not affect it.

Leave `psql` with `\q`, then run `exit` to return to your terminal.

### Check the branch again

From your terminal, find the current branch primary:

```shell
kubectl --context target -n dev get pods -l app.kubernetes.io/instance=pg-cross,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open its pod shell using the returned name:

```shell
kubectl --context target -n dev exec -it pg-cross-0 -c postgres -- bash
```

Inside the pod, start `psql`:

```shell
psql -U postgres -d postgres
```

Query all five customers:

```sql
SELECT * FROM customers ORDER BY id;
```

Expected result:

```text
 id |       full_name        |         email
----+------------------------+-----------------------
  1 | Customer 1             | customer1@example.com
  2 | Customer 2             | customer2@example.com
  3 | Customer 3             | customer3@example.com
  4 | Customer 4             | customer4@example.com
  5 | Branch-only Customer 5 | customer5@example.com
(5 rows)
```

Customer `1` still has its original value, and customer `5` retains the branch-only edit. The branch and source accept independent writes. A [refresh](/docs/guides/postgres/branch/refresh.md) replaces branch data with a new source snapshot; source changes are not continuously replicated to the branch.

Leave `psql` with `\q`, then run `exit` to return to your terminal.

## Cleanup

Delete the original Branch on the source cluster while the hub, agents, and storage backend remain available:

```shell
kubectl --context source -n demo delete branch pg-cross --wait=true --timeout=15m
kubectl --context target -n dev get postgres,pvc,volumesnapshot
```

With `deletionPolicy: Delete`, Courier removes the target database and managed storage. The source remains available. See [deletion policies](/docs/guides/postgres/branch/customization/index.md#deletion-policies) to retain a branch independently.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl --context source -n demo delete postgres prod-pg --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources. Each feature tutorial below deploys its own source database.

## Next Steps

- [Refresh a PostgreSQL branch](/docs/guides/postgres/branch/refresh.md) on a schedule and verify how replacement affects branch data.
- [Reset the branch root password](/docs/guides/postgres/branch/reset-root-password.md) and verify that source credentials cannot authenticate to the branch.
- [Run post-actions](/docs/guides/postgres/branch/post-actions.md) to anonymize customer data and remove a sensitive table before the branch becomes ready.
