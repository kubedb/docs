---
title: Running MSSQL Server Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: mssqlserver-branch-post-actions
    name: Post-Actions
    parent: guides-mssqlserver-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running MSSQL Server Branch Post-Actions

Post-actions prepare a branch before it becomes ready for use. They can mask personal fields, drop sensitive tables, or create test fixtures. This guide masks five customer rows and drops a sensitive table without changing the source.

## Execution and readiness

Each `postActions` entry creates a Job in the target namespace. This example uses one entry and one Job. Courier waits for target database readiness, then runs the action; the Branch becomes `Ready` only after it succeeds. A configured password reset would run before the action.

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

Save the following as `mssql-pa-source-issuer.yaml`. Namespace is omitted so the same file creates independent issuers and CA Secrets in `demo` and `dev`:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-pa-source-ca-selfsigned
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: mssql-pa-source-ca
spec:
  isCA: true
  commonName: mssql-pa-source-ca
  secretName: mssql-pa-source-ca
  privateKey:
    algorithm: RSA
    size: 2048
  issuerRef:
    name: mssql-pa-source-ca-selfsigned
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-pa-source-ca
spec:
  ca:
    secretName: mssql-pa-source-ca
```

```shell
kubectl -n demo apply -f mssql-pa-source-issuer.yaml
kubectl -n demo wait certificate/mssql-pa-source-ca --for=condition=Ready --timeout=5m
kubectl -n demo wait issuer/mssql-pa-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev apply -f mssql-pa-source-issuer.yaml
kubectl -n dev wait certificate/mssql-pa-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev wait issuer/mssql-pa-source-ca --for=condition=Ready --timeout=5m
```

Save the following as `mssql-pa-source.yaml`. It creates a three-replica Availability Group:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mssql-pa-source-config
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
  name: mssql-pa-source
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
    secretName: mssql-pa-source-config
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-pa-source-ca
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
kubectl -n demo apply -f mssql-pa-source.yaml
kubectl -n demo wait mssqlserver/mssql-pa-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

`agdb1` is a database managed by the Availability Group. Seed data there so it remains available through the primary Service after a failover. The configuration Secret limits core-dump size; `ACCEPT_EULA` and `MSSQL_PID` follow the SQL Server container requirements.

## Seed and inspect the source data

Discover the auth Secret referenced by `mssql-pa-source`:

```shell
kubectl -n demo get mssqlserver mssql-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-pa-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-pa-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-pa-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-pa-source.demo.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

Result tables below are formatted for readability. Create and seed the table in `agdb1`. `CHECKPOINT` flushes dirty data pages before taking the example snapshot.

```sql
USE agdb1;
GO
CREATE TABLE dbo.customers (
    id int PRIMARY KEY,
    full_name nvarchar(100) NOT NULL,
    email nvarchar(255) NOT NULL,
    phone nvarchar(30) NOT NULL
);
INSERT INTO dbo.customers (id, full_name, email, phone) VALUES
    (1, 'Customer 1', 'customer1@example.com', '+1-555-0001'),
    (2, 'Customer 2', 'customer2@example.com', '+1-555-0002'),
    (3, 'Customer 3', 'customer3@example.com', '+1-555-0003'),
    (4, 'Customer 4', 'customer4@example.com', '+1-555-0004'),
    (5, 'Customer 5', 'customer5@example.com', '+1-555-0005');
CREATE TABLE dbo.sensitive_tokens (id int PRIMARY KEY, token nvarchar(100) NOT NULL);
INSERT INTO dbo.sensitive_tokens VALUES (1, 'sample-token-1'), (2, 'sample-token-2');
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
id full_name  email                 phone
-- ---------- --------------------- -----------
1  Customer 1 customer1@example.com +1-555-0001
2  Customer 2 customer2@example.com +1-555-0002
3  Customer 3 customer3@example.com +1-555-0003
4  Customer 4 customer4@example.com +1-555-0004
5  Customer 5 customer5@example.com +1-555-0005

(5 rows affected)
```

The source also contains two sensitive tokens:

```sql
SELECT * FROM dbo.sensitive_tokens ORDER BY id;
GO
```

```text
id token
-- --------------
1  sample-token-1
2  sample-token-2

(2 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

## Create a Branch with post-actions

Save the following as `mssql-pa-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-pa-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-pa-branch
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
      name: mssql-pa-source-ca
  postActions:
  - image: mcr.microsoft.com/mssql-tools:latest
    jobDefaults:
      imagePullPolicy: IfNotPresent
      backoffLimit: 3
      activeDeadlineSeconds: 900
      ttlSecondsAfterFinished: 3600
    jobTemplate:
      metadata:
        labels:
          branch-action: mssql-pa-branch-prepare
      spec:
        # First operation: mask customer data.
        # Second operation: drop the sensitive table after masking succeeds.
        args: [/opt/mssql-tools/bin/sqlcmd, -S, "$(MSSQL_HOST),$(MSSQL_PORT)", -U, "$(MSSQL_USER)", -C, -b, -d, agdb1, -Q,
               "SET XACT_ABORT ON; BEGIN TRANSACTION; UPDATE dbo.customers SET full_name = 'Masked Customer', email = CONCAT('user', id, '@example.invalid'), phone = '+1-555-0000'; DROP TABLE IF EXISTS dbo.sensitive_tokens; COMMIT TRANSACTION;"]
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mssql-pa-branch.yaml
kubectl -n demo wait branch/mssql-pa-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch mssql-pa-branch
```

Courier creates `mssqlserver/mssql-pa-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

The tools image runs the supplied `sqlcmd` executable. Courier sets `SQLCMDPASSWORD`, so the password is not supplied in the arguments. `-b` reports SQL errors as a failed process. `SET XACT_ABORT ON` and the transaction ensure the masking and table removal succeed together; a failed update prevents a successful commit. `-C` trusts the server certificate for this tutorial connection.

## Inspect the action Job

The Branch passes through `ActionsRunning`; the phase can finish before you observe it. Find its Job by Courier’s labels:

```shell
kubectl -n dev get jobs -l courier.kubedb.com/branch=mssql-pa-branch,courier.kubedb.com/branch-namespace=demo
kubectl -n dev logs job/mssql-pa-branch-post-action -c actions
kubectl -n dev get pods -l branch-action=mssql-pa-branch-prepare
```

Expect Job status `Complete` with `1/1` completions. The SQL output reports five updated rows.

```shell
kubectl -n demo get branch mssql-pa-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
kubectl -n demo get branch mssql-pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

Expect `PostActionsCompleted=True` and `Ready=True`. The resource references identify the target auth Secret, cloned PVCs, and the action Job.

## Verify that only the branch was changed

### Inspect the branch

Discover the auth Secret referenced by `mssql-pa-branch`:

```shell
kubectl -n dev get mssqlserver mssql-pa-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-pa-branch-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-pa-branch-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-pa-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-pa-branch-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-pa-branch.dev.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name       email                 phone
-- --------------- --------------------- -----------
1  Masked Customer user1@example.invalid +1-555-0000
2  Masked Customer user2@example.invalid +1-555-0000
3  Masked Customer user3@example.invalid +1-555-0000
4  Masked Customer user4@example.invalid +1-555-0000
5  Masked Customer user5@example.invalid +1-555-0000

(5 rows affected)
```

```sql
SELECT CASE WHEN OBJECT_ID('dbo.sensitive_tokens', 'U') IS NULL THEN 1 ELSE 0 END AS sensitive_table_removed;
GO
```

```text
sensitive_table_removed
-----------------------
1
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

### Inspect the source

Discover the auth Secret referenced by `mssql-pa-source`:

```shell
kubectl -n demo get mssqlserver mssql-pa-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-pa-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-pa-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-pa-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-pa-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-pa-source.demo.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
USE agdb1;
GO
SELECT * FROM dbo.customers ORDER BY id;
GO
```

```text
id full_name  email                 phone
-- ---------- --------------------- -----------
1  Customer 1 customer1@example.com +1-555-0001
2  Customer 2 customer2@example.com +1-555-0002
3  Customer 3 customer3@example.com +1-555-0003
4  Customer 4 customer4@example.com +1-555-0004
5  Customer 5 customer5@example.com +1-555-0005

(5 rows affected)
```

```sql
SELECT * FROM dbo.sensitive_tokens ORDER BY id;
GO
```

```text
id token
-- --------------
1  sample-token-1
2  sample-token-2

(2 rows affected)
```

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

## Connection environment

| Variable | Value used by the action |
|---|---|
| `MSSQL_HOST` | Target primary Service |
| `MSSQL_PORT` | 1433 |
| `MSSQL_USER` | Username from the target auth Secret |
| `SQLCMDPASSWORD` | Target password, read automatically by sqlcmd |
| `MSSQL_PASSWORD` | The same target password, for other clients |

See the common [action configuration and Job defaults](/docs/operatormanual/branching/post-actions.md#post-action-field-reference) for the feature fields. The image must execute its supplied arguments; the Branch action template does not expose a container `command` override.

## Use post-actions with refresh or another cluster

With [refresh](/docs/guides/mssqlserver/branch/refresh.md), actions run again against each new copy before readiness. Generation `1` uses a Job named `mssql-pa-branch-post-action-1`; older generation Jobs can be removed by Courier. Design actions for repetition and use logs before cleanup. For [cross-cluster branching](/docs/guides/mssqlserver/branch/cross-cluster/index.md), the Job runs on the target cluster; inspect it there and use the target namespace in any explicit connection address.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mssql-pa-branch --wait=true --timeout=30m
kubectl -n demo delete mssqlserver mssql-pa-source --wait=true --timeout=30m
```

After both databases are deleted, remove the tutorial configuration and issuers:

```shell
kubectl -n demo delete secret mssql-pa-source-config
kubectl -n demo delete -f mssql-pa-source-issuer.yaml
kubectl -n dev delete -f mssql-pa-source-issuer.yaml
kubectl -n demo delete secret mssql-pa-source-ca
kubectl -n dev delete secret mssql-pa-source-ca
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
