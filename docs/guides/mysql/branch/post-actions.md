---
title: Running MySQL Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: mysql-branch-post-actions
    name: Post-Actions
    parent: guides-mysql-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running MySQL Branch Post-Actions

Post-actions prepare cloned MySQL data for the people and applications that will use a branch. Define containers in `spec.postActions`, and Courier runs them as Kubernetes Jobs against the target database. For example, you can:

- Mask names, email addresses, and other personal information before giving developers or AI agents access.
- Drop tables containing sensitive information that the branch does not need.
- Delete confidential rows or data outside the scope of a test.
- Insert test fixtures or application-specific seed data.

These changes run against the branch database and leave the source data unchanged.

This tutorial runs one MySQL post-action Job that anonymizes five customer records, then drops a table containing sensitive tokens. You will deploy and seed the source, create the Branch, inspect the Job, and compare the source and branch data. No database from another tutorial is required.

## Execution and readiness

Courier runs the actions **in list order**, creating a separate Kubernetes Job for each action in the target namespace. The next action starts only after the preceding Job succeeds. Actions start after the target database is ready and requested credential isolation is complete. The Branch stays in `ActionsRunning` until all actions succeed, then becomes `Ready`.

The target database Service can exist before the actions finish. Applications and automation consuming prepared data should wait for the **Branch** to become `Ready`. Branch readiness provides a completion signal; it does not restrict connections to the Service while data preparation is running.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md) to understand how branches work.
- Complete the [single-cluster Courier installation](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass with copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure MySQL version `8.4.8` is available in the installed KubeDB catalog, and use branching-capable KubeDB and database images.
- Run the commands below against the same Kubernetes context. Your account must be able to create databases and Branches, inspect Jobs and pods, and open a shell in the database pods.

The examples use the StorageClass name `ceph-rbd`. Substitute your Ceph RBD class consistently in the source and Branch manifests. The namespaces are `demo` for the source and `dev` for the target; use the database names below for this tutorial only.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `post-actions-source.yaml`. It creates a three-replica Group Replication MySQL source named `mysql-pa-source`:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-pa-source
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

Apply the manifest:

```shell
kubectl apply -f post-actions-source.yaml
kubectl -n demo wait mysql/mysql-pa-source \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Read the source database's auth Secret name from your terminal:

```shell
kubectl -n demo get mysql mysql-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mysql-pa-source-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl -n demo get secret mysql-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the source connections below. The MySQL root username is `root`.

## Seed and inspect the source data

From your terminal, open the source pod shell:

```shell
kubectl -n demo exec -it mysql-pa-source-0 -c mysql -- bash
```

Inside the pod, start the MySQL client and enter the source password at the prompt:

```shell
mysql -u root -p
```

At the `mysql>` prompt, create five customer records and a table of sample sensitive tokens:

```sql
CREATE DATABASE branchdb;
USE branchdb;

CREATE TABLE customers (
    id integer PRIMARY KEY,
    full_name varchar(100) NOT NULL,
    email varchar(255) NOT NULL,
    phone varchar(30) NOT NULL
);
INSERT INTO customers (id, full_name, email, phone) VALUES
    (1, 'Customer 1', 'customer1@example.com', '+1-555-0001'),
    (2, 'Customer 2', 'customer2@example.com', '+1-555-0002'),
    (3, 'Customer 3', 'customer3@example.com', '+1-555-0003'),
    (4, 'Customer 4', 'customer4@example.com', '+1-555-0004'),
    (5, 'Customer 5', 'customer5@example.com', '+1-555-0005');

CREATE TABLE sensitive_tokens (
    id integer PRIMARY KEY,
    token varchar(255) NOT NULL
);
INSERT INTO sensitive_tokens (id, token) VALUES
    (1, 'sample-token-1'),
    (2, 'sample-token-2');
```

In the same `mysql` session, inspect the original data:

```sql
SELECT * FROM customers ORDER BY id;
SELECT * FROM sensitive_tokens ORDER BY id;
```

Expected results:

```text
+----+------------+-----------------------+-------------+
| id | full_name  | email                 | phone       |
+----+------------+-----------------------+-------------+
|  1 | Customer 1 | customer1@example.com | +1-555-0001 |
|  2 | Customer 2 | customer2@example.com | +1-555-0002 |
|  3 | Customer 3 | customer3@example.com | +1-555-0003 |
|  4 | Customer 4 | customer4@example.com | +1-555-0004 |
|  5 | Customer 5 | customer5@example.com | +1-555-0005 |
+----+------------+-----------------------+-------------+
5 rows in set

+----+----------------+
| id | token          |
+----+----------------+
|  1 | sample-token-1 |
|  2 | sample-token-2 |
+----+----------------+
2 rows in set
```

Leave `mysql` with `\q`, then run `exit` in the pod shell to return to your terminal.

## Create a Branch with post-actions

Save the following as `branch-post-actions.yaml`. The Branch lives in `demo` and creates the target database `dev/mysql-pa-branch`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mysql-pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MySQL
      name: mysql-pa-source
    namespace: demo
  target:
    namespace: dev
    name: mysql-pa-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  postActions:
    - image: ghcr.io/appscode-images/mysql:8.4.8-oracle
      jobDefaults:
        imagePullPolicy: IfNotPresent
        backoffLimit: 3
        activeDeadlineSeconds: 900
        ttlSecondsAfterFinished: 3600
      jobTemplate:
        metadata:
          labels:
            branch-action: mysql-pa-anonymize
        spec:
          # First statement: anonymize customer data.
          # Second statement: drop the sensitive table after anonymization succeeds.
          args: [mysql, --protocol=TCP, "--user=$(MYSQL_USER)", --database=branchdb, --verbose,
                 --execute, "UPDATE customers SET full_name = 'Masked Customer', email = CONCAT('user', id, '@example.invalid'), phone = '+1-555-0000'; DROP TABLE IF EXISTS sensitive_tokens;"]
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
  deletionPolicy: Delete
```

The single `postActions` entry creates one Job. Its `mysql` process executes the SQL statements in order: anonymize the customer data, then drop the sensitive table. Both operations are safe to repeat: the update produces the same values on each run, and the drop allows the table to be absent.

`--database=branchdb` selects the application database. The client uses Courier's injected host, port, and password. Kubernetes expands `$(MYSQL_USER)` in the container arguments to supply the username; this is not a host-shell command substitution. Without `--force`, a SQL error stops execution. If the update fails, the drop does not run and the container exits unsuccessfully. `--verbose` prints the executed SQL for inspection.

Apply the Branch and wait for all preparation to finish:

```shell
kubectl apply -f branch-post-actions.yaml
kubectl -n demo wait branch/mysql-pa-branch \
  --for=condition=Ready --timeout=20m
kubectl -n demo get branch mysql-pa-branch
```

Representative output after successful completion:

```text
NAME              PHASE   MODE    TARGET            FRESHNESS   AGE
mysql-pa-branch   Ready   Local   mysql-pa-branch   2s          5m
```

The Branch passes through `ActionsRunning` while the Job executes. It may finish too quickly for you to see that phase in an interactive watch. Names and data results below follow the manifests; ages, durations, and pod names vary between runs.

## Inspect the action Job

Courier labels its Jobs with the Branch name and source namespace. List this Branch's Jobs in the target namespace:

```shell
kubectl -n dev get jobs \
  -l courier.kubedb.com/branch=mysql-pa-branch,courier.kubedb.com/branch-namespace=demo
```

Representative output:

```text
NAME                          STATUS     COMPLETIONS   DURATION   AGE
mysql-pa-branch-post-action   Complete   1/1           13s        30s
```

The Job contains one container named `actions`. Inspect its logs:

```shell
kubectl -n dev logs job/mysql-pa-branch-post-action -c actions
```

The verbose log includes both statements after a successful run:

```text
--------------
UPDATE customers SET full_name = 'Masked Customer', email = CONCAT('user', id, '@example.invalid'), phone = '+1-555-0000'
--------------

--------------
DROP TABLE IF EXISTS sensitive_tokens
--------------
```

The custom label in `jobTemplate.metadata` identifies the action pod:

```shell
kubectl -n dev get pods -l branch-action=mysql-pa-anonymize
```

Check the Branch's conditions:

```shell
kubectl -n demo get branch mysql-pa-branch \
  -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Successful completion includes these conditions; their display order may vary:

```text
SnapshotReady           True
TargetCreated           True
TargetReady             True
PostActionsCompleted    True
Ready                   True
```

Inspect the resource references recorded by Courier:

```shell
kubectl -n demo get branch mysql-pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

The `authSecret` and `clonedPVCs` fields identify target resources. `postActionJob` identifies `mysql-pa-branch-post-action` here. For Branches with multiple `postActions` entries, this field references the first Job; use the Job labels to find all of them.

## Verify that only the branch was changed

Read the branch database's auth Secret name from your terminal:

```shell
kubectl -n dev get mysql mysql-pa-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
mysql-pa-branch-auth
```

Use the returned Secret name below if it differs. Decode its password:

```shell
kubectl -n dev get secret mysql-pa-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

Keep this password for the branch connections below. The MySQL root username is `root`.

Open the branch pod shell from your terminal:

```shell
kubectl -n dev exec -it mysql-pa-branch-0 -c mysql -- bash
```

Inside the pod, start the MySQL client and enter the branch password at the prompt:

```shell
mysql -u root -p branchdb
```

At the `mysql>` prompt, inspect all five customers and check that the sensitive table was removed:

```sql
SELECT * FROM customers ORDER BY id;
SELECT COUNT(*) = 0 AS sensitive_table_removed
FROM information_schema.tables
WHERE table_schema = 'branchdb' AND table_name = 'sensitive_tokens';
```

Expected results:

```text
+----+-----------------+-----------------------+-------------+
| id | full_name       | email                 | phone       |
+----+-----------------+-----------------------+-------------+
|  1 | Masked Customer | user1@example.invalid | +1-555-0000 |
|  2 | Masked Customer | user2@example.invalid | +1-555-0000 |
|  3 | Masked Customer | user3@example.invalid | +1-555-0000 |
|  4 | Masked Customer | user4@example.invalid | +1-555-0000 |
|  5 | Masked Customer | user5@example.invalid | +1-555-0000 |
+----+-----------------+-----------------------+-------------+
5 rows in set

+-------------------------+
| sensitive_table_removed |
+-------------------------+
|                       1 |
+-------------------------+
1 row in set
```

The five customer records are preserved with anonymized values, and the sensitive table no longer exists on the branch. Leave `mysql` with `\q`, then run `exit` in the pod shell to return to your terminal.

```shell
kubectl -n demo exec -it mysql-pa-source-0 -c mysql -- bash
```

Inside the pod, start the MySQL client and enter the source password at the prompt:

```shell
mysql -u root -p branchdb
```

Query the original tables:

```sql
SELECT * FROM customers ORDER BY id;
SELECT * FROM sensitive_tokens ORDER BY id;
```

The source returns the same five original customer records and two sample tokens shown in [Seed and inspect the source data](#seed-and-inspect-the-source-data). This verifies both that anonymization was limited to the branch and that dropping its sensitive table did not remove the source table.

Leave `mysql` with `\q`, then run `exit` in the pod shell to return to your terminal.

For action fields, template configuration, and Job defaults, see the common [post-action field reference](/docs/operatormanual/branching/post-actions.md#post-action-field-reference).

## Connection environment

Courier injects the following MySQL connection variables into each action container:

| Variable | Value or source | Use |
|---|---|---|
| `MYSQL_HOST` | Target Service name, `mysql-pa-branch`. | Connects to the branch in the Job's target namespace. |
| `MYSQL_TCP_PORT` | `3306`. | MySQL Service port. |
| `MYSQL_USER` | `username` key of the target auth Secret. | Passed explicitly through `--user=$(MYSQL_USER)` in the container arguments. |
| `MYSQL_PWD` | `password` key of the target auth Secret. | Supplies the branch's current password through a Secret reference. |

The application database is selected explicitly with `--database=branchdb`. Keep passwords out of image layers, SQL text, and logs.

## Using post-actions with refresh or another cluster

A [scheduled refresh](/docs/guides/mysql/branch/refresh.md) replaces branch data from a fresh source snapshot and reruns the full action list before the Branch returns to `Ready`. Write actions that can be repeated against freshly cloned data and retried without producing unwanted changes. The deterministic update and `DROP TABLE IF EXISTS` in this tutorial meet that requirement.

Refresh generations have separate Jobs. For example, generation `1` uses `mysql-pa-branch-post-action-gen1` to run both commands. Inspect the current generation rather than assuming the initial Job name remains current. Courier removes older action Jobs after successful replacement.

The same feature applies to [cross-cluster branching](/docs/guides/mysql/branch/cross-cluster/index.md). Apply the original Branch on the source cluster; action Jobs execute in the target namespace on the **target** cluster, where the target Service, auth Secret, and any referenced ServiceAccounts or image pull Secrets must be available. Inspect Jobs and logs using the target cluster context. The source Branch reports readiness after target-side preparation completes.

## Cleanup

Delete this tutorial's Branch and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch mysql-pa-branch --wait=true --timeout=15m
```

With `deletionPolicy: Delete`, Courier removes the target database, cloned storage, auth Secret, snapshots, and action Jobs. The source remains available.

Remove the tutorial source when you are finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete mysql mysql-pa-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources.
