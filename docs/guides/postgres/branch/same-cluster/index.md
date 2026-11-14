---
title: Branch PostgreSQL into Another Namespace
menu:
  docs_{{ .version }}:
    identifier: pg-branch-same-cluster
    name: Same Cluster
    parent: guides-pg-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch PostgreSQL into Another Namespace

Same-cluster branching creates an independent PostgreSQL database from a point-in-time snapshot of an existing database within the same Kubernetes cluster. You can place the branch in another namespace to prepare development, testing, or data analysis environments using existing data.

The branch accepts independent reads and writes without changing the source database. Source changes are not continuously replicated; you can configure scheduled refreshes to replace branch data with a fresh source snapshot.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md) to understand how branches work.
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass supporting copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure PostgreSQL version `18.3` is available in the installed KubeDB catalog and use branching-capable KubeDB and database images.
- Run the commands against the same Kubernetes context.

The manifests use the StorageClass name `ceph-rbd`; substitute your Ceph RBD class consistently. Your account must be able to create databases and Branches, inspect their resources, and open a shell in the database pods. The source lives in `demo`, and the target lives in `dev`.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
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

Apply it and wait for readiness:

```shell
kubectl apply -f source-postgres.yaml
kubectl -n demo wait postgres/prod-pg \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Find the elected primary:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Example output:

```text
prod-pg-0
```

Use the returned pod name below if it differs. Pod names, ages, and durations in this guide are illustrative.

## Seed and inspect the source data

From your terminal, open a shell in the source primary:

```shell
kubectl -n demo exec -it prod-pg-0 -c postgres -- bash
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

## Create the Branch

Save the following as `branch-same-cluster.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
    namespace: demo
  target:
    namespace: dev
    name: pg-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

The Branch lives in `demo`, references `prod-pg`, and creates `pg-dev` in `dev`. Omitting `spec.target.clusterName` selects the same-cluster workflow. Courier imports source snapshot references into `dev` before creating target PVCs, because a PVC references snapshots within its own namespace.

`spec.target.storageClassName` selects the target Ceph class; `spec.target.resources` sets target PostgreSQL container requests. See the [PostgreSQL Branch specification](/docs/guides/postgres/branch/overview/index.md) for field details.

Apply the manifest and wait for the **Branch** to become ready:

```shell
kubectl apply -f branch-same-cluster.yaml
kubectl -n demo wait branch/pg-dev \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo get branch pg-dev
```

Representative output:

```text
NAME       PHASE   MODE        TARGET     FRESHNESS   AGE
pg-dev     Ready   Local       pg-dev     2s          3m
```

Inspect the target resources:

```shell
kubectl -n dev get postgres pg-dev
kubectl -n dev get pvc,volumesnapshot
```

## Verify the copied data

From your terminal, find the branch primary:

```shell
kubectl -n dev get pods -l app.kubernetes.io/instance=pg-dev,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Example output:

```text
pg-dev-0
```

Use the returned name to open its pod shell:

```shell
kubectl -n dev exec -it pg-dev-0 -c postgres -- bash
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
kubectl -n demo get pods -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open its pod shell, substituting the returned name if different:

```shell
kubectl -n demo exec -it prod-pg-0 -c postgres -- bash
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
kubectl -n dev get pods -l app.kubernetes.io/instance=pg-dev,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open its pod shell using the returned name:

```shell
kubectl -n dev exec -it pg-dev-0 -c postgres -- bash
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

Delete the Branch first and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch pg-dev --wait=true --timeout=15m
kubectl -n dev get postgres,pvc,volumesnapshot
```

With `deletionPolicy: Delete`, Courier removes the target database and managed storage. The source remains available. See [deletion policies](/docs/guides/postgres/branch/customization/index.md#deletion-policies) to retain a branch independently.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete postgres prod-pg --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources. Each feature tutorial below deploys its own source database.

## Next Steps

- [Refresh a PostgreSQL branch](/docs/guides/postgres/branch/refresh.md) on a schedule and verify how replacement affects branch data.
- [Reset the branch root password](/docs/guides/postgres/branch/reset-root-password.md) and verify that source credentials cannot authenticate to the branch.
- [Run post-actions](/docs/guides/postgres/branch/post-actions.md) to anonymize customer data and remove a sensitive table before the branch becomes ready.
