---
title: Refreshing a MSSQL Server Branch
menu:
  docs_{{ .version }}:
    identifier: mssqlserver-branch-refresh
    name: Refresh
    parent: guides-mssqlserver-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a MSSQL Server Branch

Set `spec.schedule.cron` to periodically replace branch data with a fresh source snapshot. This guide deploys its own source and scheduled branch, changes both sides, and verifies what a completed refresh preserves and replaces.

## What happens during refresh

1. Courier advances the refresh generation and takes fresh source snapshots.
2. It imports snapshot references into the target namespace, pauses the target, and replaces its cloned volumes.
3. KubeDB restarts the target using the new volumes.
4. After the target and any configured preparation complete, the Branch returns to `Ready`; Courier releases the previous generation’s snapshots.

**Refresh discards branch-only writes and temporarily interrupts the target.** It replaces data rather than merging changes. The source remains independent. If configured, [password reset](/docs/guides/mssqlserver/branch/reset-root-password.md) and [post-actions](/docs/guides/mssqlserver/branch/post-actions.md) run again on each refresh.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `MSSQLServerVersion` `2025-cu0` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Install cert-manager in the cluster. Allow capacity for three source and three target SQL Server pods, each requesting 3 GiB of memory and one CPU, plus operator and Job overhead.
- Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.

The source uses namespace `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the names in this tutorial and create namespaces if they do not already exist.

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

### Prepare certificate issuers

SQL Server requires an issuer for its supporting components even with `clientTLS: false`. The Availability Group also uses certificates for its replica endpoints. Courier does not install target issuers.

Save the following as `mssql-refresh-source-issuer.yaml`. Namespace is omitted so the same file creates independent issuers and CA Secrets in `demo` and `dev`:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-refresh-source-ca-selfsigned
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: mssql-refresh-source-ca
spec:
  isCA: true
  commonName: mssql-refresh-source-ca
  secretName: mssql-refresh-source-ca
  privateKey:
    algorithm: RSA
    size: 2048
  issuerRef:
    name: mssql-refresh-source-ca-selfsigned
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-refresh-source-ca
spec:
  ca:
    secretName: mssql-refresh-source-ca
```

```shell
kubectl -n demo apply -f mssql-refresh-source-issuer.yaml
kubectl -n demo wait certificate/mssql-refresh-source-ca --for=condition=Ready --timeout=5m
kubectl -n demo wait issuer/mssql-refresh-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev apply -f mssql-refresh-source-issuer.yaml
kubectl -n dev wait certificate/mssql-refresh-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev wait issuer/mssql-refresh-source-ca --for=condition=Ready --timeout=5m
```

Save the following as `mssql-refresh-source.yaml`. It creates a three-replica Availability Group:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mssql-refresh-source-config
  namespace: demo
stringData:
  mssql.conf: |
    [coredump]
    coredumptype = mini
    captureminiandfull = false
---
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-refresh-source
  namespace: demo
spec:
  version: 2025-cu0
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
      - agdb1
  configuration:
    secretName: mssql-refresh-source-config
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-refresh-source-ca
    clientTLS: false
  podTemplate:
    spec:
      containers:
      - name: mssql
        resources:
          requests:
            cpu: '1'
            memory: 3Gi
          limits:
            memory: 3Gi
        env:
        - name: ACCEPT_EULA
          value: "Y"
        - name: MSSQL_PID
          value: Evaluation
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

```shell
kubectl -n demo apply -f mssql-refresh-source.yaml
kubectl -n demo wait mssqlserver/mssql-refresh-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

`agdb1` is a database managed by the Availability Group. Seed data there so it remains available through the primary Service after a failover. The configuration Secret limits core-dump size; `ACCEPT_EULA` and `MSSQL_PID` follow the SQL Server container requirements.

## Seed and inspect the source data

Discover the auth Secret referenced by `mssql-refresh-source`:

```shell
kubectl -n demo get mssqlserver mssql-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-refresh-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-refresh-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-refresh-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-refresh-source.demo.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

Result tables below are formatted for readability. Create and seed the table in `agdb1`. `CHECKPOINT` flushes dirty data pages before taking the example snapshot.

```sql
USE agdb1;
GO
CREATE TABLE dbo.customers (
    id int PRIMARY KEY,
    full_name nvarchar(100) NOT NULL,
    email nvarchar(255) NOT NULL
);
INSERT INTO dbo.customers (id, full_name, email) VALUES
    (1, 'Customer 1', 'customer1@example.com'),
    (2, 'Customer 2', 'customer2@example.com'),
    (3, 'Customer 3', 'customer3@example.com'),
    (4, 'Customer 4', 'customer4@example.com'),
    (5, 'Customer 5', 'customer5@example.com');
CHECKPOINT;
GO
```

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name  email
-- ---------- ---------------------
1  Customer 1 customer1@example.com
2  Customer 2 customer2@example.com
3  Customer 3 customer3@example.com
4  Customer 4 customer4@example.com
5  Customer 5 customer5@example.com

(5 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

## Create a scheduled Branch

The example runs on ten-minute clock boundaries, such as `10:00`, `10:10`, and `10:20`; it is not a ten-minute timer from creation.

Save the following as `mssql-refresh-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-refresh-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-refresh-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: '1'
        memory: 3Gi
      limits:
        memory: 3Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-refresh-source-ca
  schedule:
    cron: '*/10 * * * *'
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mssql-refresh-branch.yaml
kubectl -n demo wait branch/mssql-refresh-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch mssql-refresh-branch
```

Courier creates `mssqlserver/mssql-refresh-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

### Check the initial copy

Discover the auth Secret referenced by `mssql-refresh-branch`:

```shell
kubectl -n dev get mssqlserver mssql-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-refresh-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-refresh-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-refresh-branch-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-refresh-branch.dev.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name  email
-- ---------- ---------------------
1  Customer 1 customer1@example.com
2  Customer 2 customer2@example.com
3  Customer 3 customer3@example.com
4  Customer 4 customer4@example.com
5  Customer 5 customer5@example.com

(5 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

### Record the current generation and schedule

```shell
kubectl -n demo get branch mssql-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT:.status.nextRefreshTime'
```

Record the generation and successful completion timestamp before making the following edits. Initial generation `0` can appear as `<none>`. Make both edits before the next scheduled run; if one starts while you are editing, finish the edits and wait for a subsequent generation. Runs that are still active can cause scheduled ticks to be skipped.

## Change source and branch data

### Update the source

Discover the auth Secret referenced by `mssql-refresh-source`:

```shell
kubectl -n demo get mssqlserver mssql-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-refresh-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-refresh-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-refresh-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-refresh-source.demo.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
UPDATE dbo.customers SET full_name = 'Updated Customer 1' WHERE id = 1;
CHECKPOINT;
GO
```

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name          email
-- ------------------ ---------------------
1  Updated Customer 1 customer1@example.com
2  Customer 2         customer2@example.com
3  Customer 3         customer3@example.com
4  Customer 4         customer4@example.com
5  Customer 5         customer5@example.com

(5 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

### Make a branch-only change

Discover the auth Secret referenced by `mssql-refresh-branch`:

```shell
kubectl -n dev get mssqlserver mssql-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-refresh-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-refresh-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-refresh-branch-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-refresh-branch.dev.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
UPDATE dbo.customers SET full_name = 'Branch-only Customer 5' WHERE id = 5;
CHECKPOINT;
GO
```

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name              email
-- ---------------------- ---------------------
1  Customer 1             customer1@example.com
2  Customer 2             customer2@example.com
3  Customer 3             customer3@example.com
4  Customer 4             customer4@example.com
5  Branch-only Customer 5 customer5@example.com

(5 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

## Wait for the next completed refresh

```shell
kubectl -n demo get branch mssql-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' --watch
```

Illustrative output:

```text
NAME                   PHASE        GENERATION   LAST-SUCCESS
mssql-refresh-branch   Ready        <none>       2026-11-11T10:02:00Z
mssql-refresh-branch   Refreshing   1            2026-11-11T10:02:00Z
mssql-refresh-branch   Ready        1            2026-11-11T10:18:00Z
```

Wait for **a later generation, phase `Ready`, and an advanced `LAST-SUCCESS` timestamp**. An initial `Ready` row or an increased generation alone does not establish completion. Press **Ctrl+C** to stop watching. Recovery duration depends on the engine and available resources.

```shell
kubectl -n demo get branch mssql-refresh-branch -o yaml
```

Inspect `status.history`, `lastSuccessfulRefreshTime`, and `snapshot.generation`. With the group-snapshot prerequisite above, `snapshot.strategy` is `VolumeGroupSnapshot`; per-volume fallback can occur if no compatible group class is available. `FRESHNESS` measures time since successful completion, not the source snapshot’s age.

## Verify the refreshed data

### Check the branch

Open a new pod/client session after recovery; the previous session may have been disconnected.

Discover the auth Secret referenced by `mssql-refresh-branch`:

```shell
kubectl -n dev get mssqlserver mssql-refresh-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-refresh-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-refresh-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-refresh-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-refresh-branch-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-refresh-branch.dev.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name          email
-- ------------------ ---------------------
1  Updated Customer 1 customer1@example.com
2  Customer 2         customer2@example.com
3  Customer 3         customer3@example.com
4  Customer 4         customer4@example.com
5  Customer 5         customer5@example.com

(5 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

Customer `1` has the source update, customer `5` has its original value, and five records remain. The branch-only edit was discarded.

### Confirm the source remains unchanged by refresh

Discover the auth Secret referenced by `mssql-refresh-source`:

```shell
kubectl -n demo get mssqlserver mssql-refresh-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-refresh-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-refresh-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-refresh-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-refresh-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-refresh-source.demo.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name          email
-- ------------------ ---------------------
1  Updated Customer 1 customer1@example.com
2  Customer 2         customer2@example.com
3  Customer 3         customer3@example.com
4  Customer 4         customer4@example.com
5  Customer 5         customer5@example.com

(5 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

The source retains its own update and never receives the branch-only change.

## Apply the same strategy across clusters

Use the [cross-cluster tutorial](/docs/guides/mssqlserver/branch/cross-cluster/index.md) and add the same schedule to its Branch. Manage the original Branch on the source cluster, inspect the database on the target, and wait for a later completed generation before checking data.

See the common [schedule and history fields](/docs/operatormanual/branching/refresh.md#schedule-and-history-field-reference) and [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) for configuration details.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mssql-refresh-branch --wait=true --timeout=30m
kubectl -n demo delete mssqlserver mssql-refresh-source --wait=true --timeout=30m
```

After both databases are deleted, remove the tutorial configuration and issuers:

```shell
kubectl -n demo delete secret mssql-refresh-source-config
kubectl -n demo delete -f mssql-refresh-source-issuer.yaml
kubectl -n dev delete -f mssql-refresh-source-issuer.yaml
kubectl -n demo delete secret mssql-refresh-source-ca
kubectl -n dev delete secret mssql-refresh-source-ca
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
