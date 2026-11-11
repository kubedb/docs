---
title: Resetting a MSSQL Server Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: mssqlserver-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-mssqlserver-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a MSSQL Server Branch Root Password

Set `spec.resetRootPassword: true` to give the branch a separate `sa` password. Rotating branch credentials reduces the risk of exposing or reusing production credentials in development and test environments. The source retains its original password.

This independent tutorial deploys a source and a same-cluster branch in another namespace, compares their passwords, and verifies which credentials each deployment accepts.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `MSSQLServerVersion` `2025-cu0` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Install KubeDB Ops Manager and the `MSSQLServerOpsRequest` CRD; password reset uses a `RotateAuth` request on the target.
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

Save the following as `mssql-reset-source-issuer.yaml`. Namespace is omitted so the same file creates independent issuers and CA Secrets in `demo` and `dev`:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-reset-source-ca-selfsigned
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: mssql-reset-source-ca
spec:
  isCA: true
  commonName: mssql-reset-source-ca
  secretName: mssql-reset-source-ca
  privateKey:
    algorithm: RSA
    size: 2048
  issuerRef:
    name: mssql-reset-source-ca-selfsigned
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-reset-source-ca
spec:
  ca:
    secretName: mssql-reset-source-ca
```

```shell
kubectl -n demo apply -f mssql-reset-source-issuer.yaml
kubectl -n demo wait certificate/mssql-reset-source-ca --for=condition=Ready --timeout=5m
kubectl -n demo wait issuer/mssql-reset-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev apply -f mssql-reset-source-issuer.yaml
kubectl -n dev wait certificate/mssql-reset-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev wait issuer/mssql-reset-source-ca --for=condition=Ready --timeout=5m
```

Save the following as `mssql-reset-source.yaml`. It creates a three-replica Availability Group:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mssql-reset-source-config
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
  name: mssql-reset-source
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
    secretName: mssql-reset-source-config
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-reset-source-ca
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
kubectl -n demo apply -f mssql-reset-source.yaml
kubectl -n demo wait mssqlserver/mssql-reset-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

`agdb1` is a database managed by the Availability Group. Seed data there so it remains available through the primary Service after a failover. The configuration Secret limits core-dump size; `ACCEPT_EULA` and `MSSQL_PID` follow the SQL Server container requirements.

## Record the source credential

Before creating the Branch, record the decoded source password for comparison and the later rejection check.

Discover the auth Secret referenced by `mssql-reset-source`:

```shell
kubectl -n demo get mssqlserver mssql-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-reset-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-reset-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

## Seed and inspect the source data

Discover the auth Secret referenced by `mssql-reset-source`:

```shell
kubectl -n demo get mssqlserver mssql-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-reset-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-reset-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-reset-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-reset-source.demo.svc,1433 -U sa -No
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

## Create a Branch with password reset enabled

Save the following as `mssql-reset-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-reset-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-reset-branch
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
      name: mssql-reset-source-ca
  resetRootPassword: true
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mssql-reset-branch.yaml
kubectl -n demo wait branch/mssql-reset-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch mssql-reset-branch
```

Courier creates `mssqlserver/mssql-reset-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

Inspect the initial generation’s rotation request in `dev`:

```shell
kubectl -n dev get mssqlserveropsrequest mssql-reset-branch-branch-reset-0 -o custom-columns='NAME:.metadata.name,TYPE:.spec.type,PHASE:.status.phase'
```

```text
NAME                                  TYPE         PHASE
mssql-reset-branch-branch-reset-0   RotateAuth   Successful
```

Check the Branch conditions:

```shell
kubectl -n demo get branch mssql-reset-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Expect `RootPasswordReset=True` and `Ready=True`. The following checks establish that the intended credentials actually authenticate.

## Compare source and branch passwords

Discover both Secret names:

```shell
kubectl -n demo get mssqlserver mssql-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get mssqlserver mssql-reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-reset-source-auth
mssql-reset-branch-auth
```

Use the returned names if they differ. Read each password separately:

```shell
kubectl -n demo get secret mssql-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
source-example-password
```

```shell
kubectl -n dev get secret mssql-reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
branch-example-password
```

The displayed strings are illustrative. Your source password must match the value recorded before branching, while the branch password must differ.

## Verify authentication through the Services

Open the source pod shell from your terminal:

```shell
kubectl -n demo exec -it mssql-reset-source-0 -c mssql -- bash
```

### Connect with the branch credential

Run this command inside the pod and enter the **branch password** at the password prompt:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-reset-branch.dev.svc,1433 -U sa -No
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

Run `QUIT` to leave `sqlcmd`, keeping the pod shell open.

### Confirm the source credential is rejected by the branch

Repeat the command below, this time entering the **source password recorded before branching** at the prompt. The endpoint is still the branch Service:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-reset-branch.dev.svc,1433 -U sa -No
```

```text
Sqlcmd: Error: Microsoft ODBC Driver 18 for SQL Server : Login failed for user 'sa'.
```

A timeout or DNS error does not verify credential isolation; expect an authentication rejection from the branch.

### Confirm the source credential still works

Connect to the source Service and enter the **original source password**:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-reset-source.demo.svc,1433 -U sa -No
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

## Use password reset with other branch features

On each [refresh](/docs/guides/mssqlserver/branch/refresh.md), the cloned data brings back the source credential and Courier performs a new rotation before readiness. Re-read the branch Secret after refresh. Configured [post-actions](/docs/guides/mssqlserver/branch/post-actions.md) run after password reset and receive branch credentials. For another cluster, use the same setting on the [cross-cluster Branch](/docs/guides/mssqlserver/branch/cross-cluster/index.md) and inspect the rotation request on the target.

See the common [password-reset field reference](/docs/operatormanual/branching/reset-root-password.md#field-reference).

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mssql-reset-branch --wait=true --timeout=30m
kubectl -n demo delete mssqlserver mssql-reset-source --wait=true --timeout=30m
```

After both databases are deleted, remove the tutorial configuration and issuers:

```shell
kubectl -n demo delete secret mssql-reset-source-config
kubectl -n demo delete -f mssql-reset-source-issuer.yaml
kubectl -n dev delete -f mssql-reset-source-issuer.yaml
kubectl -n demo delete secret mssql-reset-source-ca
kubectl -n dev delete secret mssql-reset-source-ca
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
