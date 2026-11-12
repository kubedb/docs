---
title: Branch MySQL into Another Namespace
menu:
  docs_{{ .version }}:
    identifier: mysql-branch-same-cluster
    name: Same Cluster
    parent: guides-mysql-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch MySQL into Another Namespace

Same-cluster branching creates an independent MySQL database from a point-in-time snapshot of an existing database within the same Kubernetes cluster. You can place the branch in another namespace to prepare development, testing, or data analysis environments using existing data.

The branch accepts independent reads and writes without changing the source database. Source changes are not continuously replicated; you can configure scheduled refreshes to replace branch data with a fresh source snapshot.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md) to understand how branches work.
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass supporting copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure MySQL version `8.4.8` is available in the installed KubeDB catalog and use branching-capable KubeDB and database images.
- Run the commands against the same Kubernetes context.

The manifests use the StorageClass name `ceph-rbd`; substitute your Ceph RBD class consistently. Your account must be able to create databases and Branches, inspect their resources, and open a shell in the database pods. The source lives in `demo`, and the target lives in `dev`.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `source-mysql.yaml`. It creates a three-replica Group Replication MySQL source named `mysql-source`:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-source
  namespace: demo
spec:
  version: "8.4.8"
  replicas: 3
  topology:
    mode: GroupReplication
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
kubectl apply -f source-mysql.yaml
kubectl -n demo wait mysql/mysql-source \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Read the source database's auth Secret name from your terminal:

```shell
kubectl -n demo get mysql mysql-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mysql-source-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl -n demo get secret mysql-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the source connections below. The MySQL root username is `root`.

## Seed and inspect the source data

From your terminal, open the source pod shell:

```shell
kubectl -n demo exec -it mysql-source-0 -c mysql -- bash
```

Inside the pod, start the MySQL client and enter the source password at the prompt:

```shell
mysql -u root -p
```

At the `mysql>` prompt, create five customer records:

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

Leave `mysql` with `\q`, then run `exit` in the pod shell to return to your terminal.

## Create the Branch

Save the following as `branch-same-cluster.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mysql-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MySQL
      name: mysql-source
    namespace: demo
  target:
    namespace: dev
    name: mysql-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

The Branch lives in `demo`, references `mysql-source`, and creates `mysql-dev` in `dev`. Omitting `spec.target.clusterName` selects the same-cluster workflow. Courier imports source snapshot references into `dev` before creating target PVCs, because a PVC references snapshots within its own namespace.

`spec.target.storageClassName` selects the target Ceph class; `spec.target.resources` sets target MySQL container requests. See the [MySQL Branch specification](/docs/guides/mysql/branch/overview/index.md) for field details.

Apply the manifest and wait for the **Branch** to become ready:

```shell
kubectl apply -f branch-same-cluster.yaml
kubectl -n demo wait branch/mysql-dev \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo get branch mysql-dev
```

Representative output:

```text
NAME        PHASE   MODE    TARGET      FRESHNESS   AGE
mysql-dev   Ready   Local   mysql-dev   2s          3m
```

Inspect the target resources:

```shell
kubectl -n dev get mysql mysql-dev
kubectl -n dev get pvc,volumesnapshot
```

## Verify the copied data

Read the branch database's auth Secret name from your terminal:

```shell
kubectl -n dev get mysql mysql-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mysql-dev-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl -n dev get secret mysql-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the branch connections below. The MySQL root username is `root`.

Open the branch pod shell from your terminal:

```shell
kubectl -n dev exec -it mysql-dev-0 -c mysql -- bash
```

Inside the pod, start the MySQL client and enter the branch password at the prompt:

```shell
mysql -u root -p branchdb
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

Keep this `mysql` session open for the next step. These connections verify copied data; password-isolation checks are covered in the [root password reset tutorial](/docs/guides/mysql/branch/reset-root-password.md).

## Verify independent writes

### Change the branch

In the same branch `mysql` session, update customer `5`:

```sql
UPDATE customers SET full_name = 'Branch-only Customer 5' WHERE id = 5;
```

Expected output:

```text
Query OK, 1 row affected
```

Leave `mysql` with `\q`, then run `exit` to return to your terminal.

### Change and inspect the source

```shell
kubectl -n demo exec -it mysql-source-0 -c mysql -- bash
```

Inside the pod, start the MySQL client and enter the source password at the prompt:

```shell
mysql -u root -p branchdb
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

Leave `mysql` with `\q`, then run `exit` to return to your terminal.

### Check the branch again

```shell
kubectl -n dev exec -it mysql-dev-0 -c mysql -- bash
```

Inside the pod, start the MySQL client and enter the branch password at the prompt:

```shell
mysql -u root -p branchdb
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

Customer `1` still has its original value, and customer `5` retains the branch-only edit. The branch and source accept independent writes. A [refresh](/docs/guides/mysql/branch/refresh.md) replaces branch data with a new source snapshot; source changes are not continuously replicated to the branch.

Leave `mysql` with `\q`, then run `exit` to return to your terminal.

## Cleanup

Delete the Branch first and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch mysql-dev --wait=true --timeout=15m
kubectl -n dev get mysql,pvc,volumesnapshot
```

With `deletionPolicy: Delete`, Courier removes the target database and managed storage. The source remains available. See [deletion policies](/docs/guides/mysql/branch/customization/index.md#deletion-policies) to retain a branch independently.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete mysql mysql-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources. Each feature tutorial below deploys its own source database.

## Next Steps

- [Refresh a MySQL branch](/docs/guides/mysql/branch/refresh.md) on a schedule and verify how replacement affects branch data.
- [Reset the branch root password](/docs/guides/mysql/branch/reset-root-password.md) and verify that source credentials cannot authenticate to the branch.
- [Run post-actions](/docs/guides/mysql/branch/post-actions.md) to anonymize customer data and remove a sensitive table before the branch becomes ready.
