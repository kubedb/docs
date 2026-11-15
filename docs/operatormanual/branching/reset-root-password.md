---
title: Resetting a Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: branching-reset-root-password
    name: ResetRootPassword
    parent: branching
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a Branch Root Password

Set `spec.resetRootPassword: true` to give a branch its own root credential. Giving the branch its own root password reduces the risk of exposing or reusing production credentials in development and test environments. The source database retains its original credential.

This feature applies to all supported database engines. This tutorial uses PostgreSQL to demonstrate a same-cluster branch in a different namespace. You will deploy and seed the source, create a Branch with password reset enabled, compare both passwords, and verify authentication through the database Services. No database from another tutorial is required.

## Execution order

Without this option, Courier copies the source credential into a separate target Secret. Different Secret names therefore do not necessarily mean different passwords.

With `resetRootPassword: true`, Courier coordinates credential isolation before running any post-actions or reporting the Branch `Ready`. The database-specific implementation determines whether credentials are generated during provisioning or changed through a database operation. Engines that use an OpsRequest for rotation require KubeDB Ops-manager support on the target cluster.

For PostgreSQL, Courier leaves the target auth Secret for KubeDB to generate during provisioning. The branching-capable PostgreSQL startup applies that credential to the cloned data. This path does not require a password-rotation OpsRequest.

## Before You Begin

- Complete the [single-cluster Courier installation](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD.
- Have a working Ceph RBD StorageClass supporting copy-on-write cloning and a compatible snapshot class. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Ensure PostgreSQL version `18.3` is available in the installed KubeDB catalog, and use branching-capable KubeDB and database images.
- Run the commands against the same Kubernetes context. Your account must be able to create databases and Branches, read auth Secrets, inspect pods, and open a shell in database pods.
- Allow the source pod to connect to the source and branch database Services on port `5432` for the authentication checks.

The manifests use the StorageClass name `ceph-rbd`. Substitute your Ceph RBD class consistently in both manifests. The source lives in `demo` and the target in `dev`; reserve the database names below for this tutorial.

Create the namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `reset-source.yaml`. It creates a three-replica PostgreSQL source named `reset-source`:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: reset-source
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
kubectl apply -f reset-source.yaml
kubectl -n demo wait postgres/reset-source \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
```

Find the source primary:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=reset-source,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Example output:

```text
reset-source-0
```

Use the returned pod name in the following commands; this example uses `reset-source-0`.

## Seed and inspect the source data

From your terminal, open a shell in the source primary:

```shell
kubectl -n demo exec -it reset-source-0 -c postgres -- bash
```

Inside the pod, connect to database `postgres`. Local socket connections in this setup use trust authentication, so no password is required for seeding:

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

## Record the source password

Read the auth Secret name from the source database:

```shell
kubectl -n demo get postgres reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output:

```text
reset-source-auth
```

Use the returned Secret name below if it differs. Decode the password:

```shell
kubectl -n demo get secret reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Illustrative output:

```text
source-example-password
```

Keep the actual value available for the comparison and authentication checks. The passwords shown here are illustrative; KubeDB generates your installation's values.

## Create a Branch with password reset enabled

Save the following as `branch-reset-root-password.yaml`. The Branch lives in `demo` and creates `dev/reset-branch`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: reset-source
    namespace: demo
  target:
    namespace: dev
    name: reset-branch
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
kubectl -n demo wait branch/reset-branch \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
kubectl -n demo get branch reset-branch
```

Illustrative output:

```text
NAME           PHASE   MODE    TARGET         FRESHNESS   AGE
reset-branch   Ready   Local   reset-branch   2s          3m
```

Check the conditions:

```shell
kubectl -n demo get branch reset-branch \
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

## Compare source and branch passwords

From your terminal, discover both databases' auth Secret names:

```shell
kubectl -n demo get postgres reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get postgres reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

Example output, in command order:

```text
reset-source-auth
reset-branch-auth
```

Use those returned names in the commands below. Read the **source** password again:

```shell
kubectl -n demo get secret reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

It should match the value recorded before creating the Branch:

```text
source-example-password
```

Read the **branch** password:

```shell
kubectl -n dev get secret reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

It should differ from the source password:

```text
branch-example-password
```

This comparison demonstrates that the branch has a different password and the source password remains unchanged. Applications connecting to the branch should use the target database's auth Secret after the Branch reaches `Ready`.

## Verify authentication through the Services

Use TCP connections through the database Services for password checks. The earlier local socket connection used trust authentication and cannot demonstrate password isolation. `-h` selects the Service, and `-W` makes `psql` prompt for a password without putting it in a command.

From your terminal, find the source primary again in case a failover occurred:

```shell
kubectl -n demo get pods -l app.kubernetes.io/instance=reset-source,kubedb.com/role=primary -o jsonpath='{.items[0].metadata.name}{"\n"}'
```

Open a shell in that pod, substituting the returned name if different:

```shell
kubectl -n demo exec -it reset-source-0 -c postgres -- bash
```

Run the following authentication checks from this pod shell. The example uses PostgreSQL's generated root user, `postgres`.

### Connect to the branch with its own password

Connect through the branch Service in `dev`:

```shell
psql -h reset-branch.dev.svc -p 5432 -U postgres -d postgres -W
```

At `Password for user postgres:`, enter the **branch password** decoded above. It is not echoed. Once connected, query the copied data:

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

The branch password authenticates successfully, and all five source rows are present. Run `\q` to return to the pod shell; keep this shell open for the next check.

### Confirm the source password is rejected by the branch

From the same pod shell, repeat the connection:

```shell
psql -h reset-branch.dev.svc -p 5432 -U postgres -d postgres -W
```

This time, enter the **source password** at the prompt. The connection should fail with an error containing:

```text
FATAL:  password authentication failed for user "postgres"
```

The failed command returns to the pod shell. A DNS error, connection timeout, or connection refusal does not establish password isolation; the expected result is a password-authentication rejection from PostgreSQL.

### Confirm the source credential and data still work

From the same pod shell, connect through the source Service:

```shell
psql -h reset-source.demo.svc -p 5432 -U postgres -d postgres -W
```

Enter the original **source password** at the prompt. Once connected, inspect the source data:

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

The original password still authenticates to the source, and its data is unchanged. Leave `psql` with `\q`, then run `exit` to return to your terminal.

## Field reference

| Field | Type | Default or omitted behavior | Purpose |
|---|---|---|---|
| `spec.resetRootPassword` | Boolean | `false`: the target receives the source root credential in a separate auth Secret. | Set `true` to give the branch its own root credential. Courier completes credential isolation before reporting the Branch `Ready`. The source credential remains unchanged. |

For PostgreSQL, the root user in this example is `postgres`. Read the generated credential from the target database's `spec.authSecret.name`; the Branch also records the target Secret name in `status.resources.authSecret`.

## Using password reset with other branch features

A [scheduled refresh](/docs/operatormanual/branching/refresh.md) replaces branch data and reapplies credential isolation when enabled. Wait for the refreshed Branch to return to `Ready`, then read its current auth Secret before connecting.

[Post-actions](/docs/operatormanual/branching/post-actions.md) run after credential isolation and receive the branch's current connection credentials automatically.

## Cleanup

Delete this tutorial's Branch and wait for Courier to finish cleanup:

```shell
kubectl -n demo delete branch reset-branch --wait=true --timeout=15m
```

With `deletionPolicy: Delete`, Courier removes the target database, cloned volumes, auth Secret, and managed snapshots. The source remains available.

Remove the tutorial source when finished. Its `WipeOut` policy deletes its data:

```shell
kubectl -n demo delete postgres reset-source --wait=true --timeout=15m
```

These commands leave the `demo` and `dev` namespaces available for other resources.
