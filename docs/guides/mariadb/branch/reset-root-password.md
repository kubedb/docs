---
title: Resetting a MariaDB Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: mariadb-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-mariadb-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a MariaDB Branch Root Password

Set `spec.resetRootPassword: true` to give a MariaDB branch its own root credential. Giving the branch its own root password reduces the risk of exposing or reusing production credentials in development and test environments. The source database retains its original credential.

This tutorial demonstrates a same-cluster MariaDB branch in a different namespace. You will deploy and seed the source, create a Branch with password reset enabled, compare both passwords, and verify authentication through the database Services. No database from another tutorial is required.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md) to understand how branches work.
- Complete the [single-cluster Courier installation](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass supporting copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure MariaDB version `12.3.2` is available in the installed KubeDB catalog, and use branching-capable KubeDB and database images.
- Run the commands against the same Kubernetes context. Your account must be able to create databases and Branches, read auth Secrets, inspect pods, and open a shell in database pods.
- Install KubeDB Ops Manager with MariaDB RotateAuth support and the `MariaDBOpsRequest` CRD. Courier creates the rotation request automatically.
- Allow the source pod to connect to the source and branch database Services on port `3306` for the authentication checks.

The manifests use the StorageClass name `ceph-rbd`. Substitute your Ceph RBD class consistently in both manifests. The source lives in `demo` and the target in `dev`; reserve the database names below for this tutorial.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `mariadb-reset-source.yaml`. It creates a three-replica Galera MariaDB source named `mariadb-reset-source`:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-reset-source
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
kubectl apply -f mariadb-reset-source.yaml
kubectl -n demo wait mariadb/mariadb-reset-source \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

## Record the source password

Read the auth Secret name from the source database:

```shell
kubectl -n demo get mariadb mariadb-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mariadb-reset-source-auth
```

Use the returned Secret name below if it differs. Decode the password:

```shell
kubectl -n demo get secret mariadb-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Illustrative output:

```text
source-example-password
```

Keep the actual value available for the comparison and authentication checks. The passwords shown here are illustrative; KubeDB generates your installation's values.

## Seed and inspect the source data

From your terminal, open the source pod shell:

```shell
kubectl -n demo exec -it mariadb-reset-source-0 -c mariadb -- bash
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

## Create a Branch with password reset enabled

Save the following as `branch-reset-root-password.yaml`. The Branch lives in `demo` and creates `dev/mariadb-reset-branch`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mariadb-reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MariaDB
      name: mariadb-reset-source
    namespace: demo
  target:
    namespace: dev
    name: mariadb-reset-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  resetRootPassword: true
  deletionPolicy: Delete
```

Omitting `spec.target.clusterName` keeps this branch in the same cluster (`Local` mode). `resetRootPassword: true` requests a separate root credential for the target.

Apply the manifest and wait for the Branch to become ready:

```shell
kubectl apply -f branch-reset-root-password.yaml
kubectl -n demo wait branch/mariadb-reset-branch \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
kubectl -n demo get branch mariadb-reset-branch
```

Illustrative output:

```text
NAME                   PHASE   MODE    TARGET                 FRESHNESS   AGE
mariadb-reset-branch   Ready   Local   mariadb-reset-branch   2s          3m
```

Check the conditions:

```shell
kubectl -n demo get branch mariadb-reset-branch \
  -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Successful completion includes these conditions; their order may vary:

```text
SnapshotReady           True
TargetCreated           True
TargetReady             True
RootPasswordReset       True
Ready                   True
```

`RootPasswordReset: True` records completion of credential isolation. The following checks compare the Secrets and verify that the database accepts the intended password.

## Inspect the password rotation

Courier creates the initial rotation request in the target namespace. Inspect its result from your terminal:

```shell
kubectl -n dev get mariadbopsrequest mariadb-reset-branch-branch-reset-0 \
  -o custom-columns='NAME:.metadata.name,TYPE:.spec.type,DATABASE:.spec.databaseRef.name,PHASE:.status.phase'
```

Expected values:

```text
NAME                                  TYPE         DATABASE               PHASE
mariadb-reset-branch-branch-reset-0   RotateAuth   mariadb-reset-branch   Successful
```

The suffix `0` identifies the initial generation. A configured refresh creates a separate rotation request for each new generation. Wait for the Branch's `RootPasswordReset` condition and `Ready` phase before using its credential.

## Compare source and branch passwords

From your terminal, discover both databases' auth Secret names:

```shell
kubectl -n demo get mariadb mariadb-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get mariadb mariadb-reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output, in command order:

```text
mariadb-reset-source-auth
mariadb-reset-branch-auth
```

Use the returned Secret names below if they differ. Read the **source** password again:

```shell
kubectl -n demo get secret mariadb-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

It should match the value recorded before creating the Branch:

```text
source-example-password
```

Read the **branch** password:

```shell
kubectl -n dev get secret mariadb-reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

It should differ from the source password:

```text
branch-example-password
```

This comparison demonstrates that the branch has a different password and the source password remains unchanged. Applications connecting to the branch should use the target database's auth Secret after the Branch reaches `Ready`.

## Verify authentication through the Services

Use TCP connections through the database Services for password checks. `--protocol=TCP` selects TCP, `-h` selects the Service, and `-p` prompts for a password. Each check below explicitly identifies which password to enter.

From your terminal, open the source pod shell:

```shell
kubectl -n demo exec -it mariadb-reset-source-0 -c mariadb -- bash
```

Run the following authentication checks from this pod shell. The example uses MariaDB's root user, `root`.

### Connect to the branch with its own password

Connect through the branch Service in `dev`:

```shell
mariadb --protocol=TCP -h mariadb-reset-branch.dev.svc -P 3306 -u root -p branchdb
```

At `Enter password:`, enter the **branch password** decoded above. It is not echoed. Once connected, query the copied data:

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

The branch password authenticates successfully, and all five source rows are present. Run `\q` to return to the pod shell; keep this shell open for the next check.

### Confirm the source password is rejected by the branch

From the same pod shell, repeat the connection:

```shell
mariadb --protocol=TCP -h mariadb-reset-branch.dev.svc -P 3306 -u root -p branchdb
```

This time, enter the **source password** at the prompt. The connection should fail with an error containing:

```text
ERROR 1045 (28000): Access denied for user 'root'@'<client-address>' (using password: YES)
```

The client address in the error varies. The failed command returns to the pod shell. A DNS error, connection timeout, or connection refusal does not establish password isolation; the expected result is a password-authentication rejection from MariaDB.

### Confirm the source credential and data still work

From the same pod shell, connect through the source Service:

```shell
mariadb --protocol=TCP -h mariadb-reset-source.demo.svc -P 3306 -u root -p branchdb
```

Enter the original **source password** at the prompt. Once connected, inspect the source data:

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

The original password still authenticates to the source, and its data is unchanged. Leave `mariadb` with `\q`, then run `exit` to return to your terminal.

For the option's default and behavior, see the common [root password reset field reference](/docs/operatormanual/branching/reset-root-password.md#field-reference).

## Using password reset with other branch features

A [scheduled refresh](/docs/guides/mariadb/branch/refresh.md) replaces branch data and reapplies credential isolation when enabled. Wait for the refreshed Branch to return to `Ready`, then read its current auth Secret before connecting.

[Post-actions](/docs/guides/mariadb/branch/post-actions.md) run after credential isolation and receive the branch's current connection credentials automatically.

## Cleanup

Delete this tutorial's Branch and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch mariadb-reset-branch --wait=true --timeout=15m
```

With `deletionPolicy: Delete`, Courier removes the target database, cloned volumes, auth Secret, and managed snapshots. The source remains available.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete mariadb mariadb-reset-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources.
