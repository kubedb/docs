---
title: Branch MariaDB Across Clusters
menu:
  docs_{{ .version }}:
    identifier: mariadb-branch-cross-cluster
    name: Cross Cluster
    parent: guides-mariadb-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch MariaDB Across Clusters

Cross-cluster branching creates an independent MariaDB database on another Kubernetes cluster from a point-in-time snapshot of an existing database. This lets you prepare development, testing, or data analysis environments on a separate cluster using existing data.

The branch accepts independent reads and writes without changing the source database. Source changes are not continuously replicated; you can configure scheduled refreshes to replace branch data with a fresh source snapshot.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md) to understand how branches work.
- Complete [Courier cross-cluster setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm) in an existing OCM environment.
- Have `kubectl` contexts named `hub`, `source`, and `target`. The OCM ManagedCluster names in this tutorial are `source` and `target`.
- Install compatible branching-capable KubeDB and database images with valid licenses on both database clusters. MariaDB version `12.3.2` must be available in both installed KubeDB catalogs.
- Configure compatible Ceph RBD storage and snapshot classes on both clusters. Source snapshots must be resolvable on the target against the shared backend. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

Using two independent Ceph backends does not satisfy the storage requirement. Confirm [CSI/backend compatibility](/docs/operatormanual/branching/snapshots.md#cross-cluster-backend-compatibility) before creating the source.

The manifests use the StorageClass name `ceph-rbd`; substitute your Ceph RBD class consistently. Your account must be able to create databases and Branches, inspect their resources, and open a shell in the database pods. Commands identify the cluster with `--context source`, `--context target`, or `--context hub`.

Create the namespaces if they do not already exist:

```shell
kubectl --context source create namespace demo
kubectl --context target create namespace dev
```

## Deploy the source database

Save the following as `source-mariadb.yaml`. It creates a three-replica Galera MariaDB source named `mariadb-source`:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-source
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

Apply it on the source cluster and wait for readiness:

```shell
kubectl --context source apply -f source-mariadb.yaml
kubectl --context source -n demo wait mariadb/mariadb-source \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Read the source database's auth Secret name from your terminal:

```shell
kubectl --context source -n demo get mariadb mariadb-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mariadb-source-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl --context source -n demo get secret mariadb-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the source connections below. The MariaDB root username is `root`.

## Seed and inspect the source data

From your terminal, open the source pod shell:

```shell
kubectl --context source -n demo exec -it mariadb-source-0 -c mariadb -- bash
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

## Create the cross-cluster Branch

Save the following as `branch-cross-cluster.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mariadb-cross
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MariaDB
      name: mariadb-source
    namespace: demo
  target:
    clusterName: target
    namespace: dev
    name: mariadb-cross
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

`spec.target.clusterName` is the target's **OCM ManagedCluster name**, which can differ from a kubeconfig context name. Apply the original Branch on the **source** cluster. Courier routes the request through the hub; on the target, Courier prepares cloned PVCs and the MariaDB resource, and KubeDB provisions the database using those PVCs.

`spec.target.storageClassName` selects the target Ceph class; `spec.target.resources` sets target MariaDB container requests. See the [MariaDB Branch specification](/docs/guides/mariadb/branch/overview/index.md) for field details.

Apply the manifest and wait for the **Branch** to become ready:

```shell
kubectl --context source apply -f branch-cross-cluster.yaml
kubectl --context source -n demo wait branch/mariadb-cross \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
kubectl --context source -n demo get branch mariadb-cross
```

Representative output:

```text
NAME            PHASE   MODE        TARGET          FRESHNESS   AGE
mariadb-cross   Ready   Initiator   mariadb-cross   2s          3m
```

Inspect the target resources on the target cluster:

```shell
kubectl --context target -n dev get mariadb mariadb-cross
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

Read the branch database's auth Secret name from your terminal:

```shell
kubectl --context target -n dev get mariadb mariadb-cross -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mariadb-cross-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl --context target -n dev get secret mariadb-cross-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the branch connections below. The MariaDB root username is `root`.

Open the branch pod shell from your terminal:

```shell
kubectl --context target -n dev exec -it mariadb-cross-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the branch password at the prompt:

```shell
mariadb -u root -p branchdb
```

Query the copied data:

```sql
SELECT * FROM customers ORDER BY id;
```

The branch should contain the same five rows:

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

Keep this `mariadb` session open for the next step. These connections verify copied data; password-isolation checks are covered in the [root password reset tutorial](/docs/guides/mariadb/branch/reset-root-password.md).

## Verify independent writes

### Change the branch

In the same branch `mariadb` session, update customer `5`:

```sql
UPDATE customers SET full_name = 'Branch-only Customer 5' WHERE id = 5;
```

Expected output:

```text
Query OK, 1 row affected
```

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

### Change and inspect the source

```shell
kubectl --context source -n demo exec -it mariadb-source-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the source password at the prompt:

```shell
mariadb -u root -p branchdb
```

Update customer `1` and inspect all five rows:

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

Customer `5` retains its original value on the source: the branch-only edit did not affect it.

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

### Check the branch again

```shell
kubectl --context target -n dev exec -it mariadb-cross-0 -c mariadb -- bash
```

Inside the pod, start the MariaDB client and enter the branch password at the prompt:

```shell
mariadb -u root -p branchdb
```

Query all five customers:

```sql
SELECT * FROM customers ORDER BY id;
```

Expected result:

```text
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

Customer `1` still has its original value, and customer `5` retains the branch-only edit. The branch and source accept independent writes. A [refresh](/docs/guides/mariadb/branch/refresh.md) replaces branch data with a new source snapshot; source changes are not continuously replicated to the branch.

Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

## Cleanup

Delete the original Branch on the source cluster while the hub, agents, and storage backend remain available:

```shell
kubectl --context source -n demo delete branch mariadb-cross --wait=true --timeout=15m
kubectl --context target -n dev get mariadb,pvc,volumesnapshot
```

With `deletionPolicy: Delete`, Courier removes the target database and managed storage. The source remains available. See [deletion policies](/docs/guides/mariadb/branch/customization/index.md#deletion-policies) to retain a branch independently.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl --context source -n demo delete mariadb mariadb-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources. Each feature tutorial below deploys its own source database.

## Next Steps

- [Refresh a MariaDB branch](/docs/guides/mariadb/branch/refresh.md) on a schedule and verify how replacement affects branch data.
- [Reset the branch root password](/docs/guides/mariadb/branch/reset-root-password.md) and verify that source credentials cannot authenticate to the branch.
- [Run post-actions](/docs/guides/mariadb/branch/post-actions.md) to anonymize customer data and remove a sensitive table before the branch becomes ready.
