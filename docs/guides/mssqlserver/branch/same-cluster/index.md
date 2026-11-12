---
title: Branch MSSQL Server into Another Namespace
menu:
  docs_{{ .version }}:
    identifier: mssqlserver-branch-same-cluster
    name: Same-Cluster
    parent: guides-mssqlserver-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch MSSQL Server into Another Namespace

Same-cluster branching creates an independent MSSQL Server deployment from a point-in-time snapshot within the same Kubernetes cluster. Place the branch in another namespace to prepare development, testing, or data analysis environments using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

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

Save the following as `mssql-source-issuer.yaml`. Namespace is omitted so the same file creates independent issuers and CA Secrets in `demo` and `dev`:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-source-ca-selfsigned
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: mssql-source-ca
spec:
  isCA: true
  commonName: mssql-source-ca
  secretName: mssql-source-ca
  privateKey:
    algorithm: RSA
    size: 2048
  issuerRef:
    name: mssql-source-ca-selfsigned
    kind: Issuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssql-source-ca
spec:
  ca:
    secretName: mssql-source-ca
```

```shell
kubectl -n demo apply -f mssql-source-issuer.yaml
kubectl -n demo wait certificate/mssql-source-ca --for=condition=Ready --timeout=5m
kubectl -n demo wait issuer/mssql-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev apply -f mssql-source-issuer.yaml
kubectl -n dev wait certificate/mssql-source-ca --for=condition=Ready --timeout=5m
kubectl -n dev wait issuer/mssql-source-ca --for=condition=Ready --timeout=5m
```

Save the following as `mssql-source.yaml`. It creates a three-replica Availability Group:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mssql-source-config
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
  name: mssql-source
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
    secretName: mssql-source-config
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-source-ca
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
kubectl -n demo apply -f mssql-source.yaml
kubectl -n demo wait mssqlserver/mssql-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

`agdb1` is a database managed by the Availability Group. Seed data there so it remains available through the primary Service after a failover. The configuration Secret limits core-dump size; `ACCEPT_EULA` and `MSSQL_PID` follow the SQL Server container requirements.

## Seed and inspect the source data

Discover the auth Secret referenced by `mssql-source`:

```shell
kubectl -n demo get mssqlserver mssql-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-source.demo.svc,1433 -U sa -No
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

## Create the Branch

Save the following as `mssql-dev.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-dev
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
      name: mssql-source-ca
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f mssql-dev.yaml
kubectl -n demo wait branch/mssql-dev --for=condition=Ready --timeout=60m
kubectl -n demo get branch mssql-dev
```

Courier creates `mssqlserver/mssql-dev` and its cloned PVCs in `dev`; KubeDB provisions the target database.

## Verify the copied data

Discover the auth Secret referenced by `mssql-dev`:

```shell
kubectl -n dev get mssqlserver mssql-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-dev-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-dev.dev.svc,1433 -U sa -No
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

### Check the target cluster

```shell
kubectl -n dev get pods,pvc -l app.kubernetes.io/instance=mssql-dev
```

The target retains the Availability Group shape. Check the roles and database synchronization from the target primary Service:

Discover the auth Secret referenced by `mssql-dev`:

```shell
kubectl -n dev get mssqlserver mssql-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-dev-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-dev.dev.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

```sql
SELECT ar.replica_server_name, ars.role_desc, ars.connected_state_desc
FROM sys.dm_hadr_availability_replica_states AS ars
JOIN sys.availability_replicas AS ar ON ar.replica_id = ars.replica_id;
GO
SELECT DB_NAME(database_id) AS database_name, synchronization_state_desc
FROM sys.dm_hadr_database_replica_states
WHERE DB_NAME(database_id) = 'agdb1';
GO
```

Expect one `PRIMARY`, two `SECONDARY` replicas, `CONNECTED` replica states, and `SYNCHRONIZED` rows for `agdb1` after recovery completes. Replica names must refer to the branch.

Run `QUIT` at the `sqlcmd` prompt, then `exit` in the pod shell to return to your terminal.

## Verify independent writes

### Change the branch

Discover the auth Secret referenced by `mssql-dev`:

```shell
kubectl -n dev get mssqlserver mssql-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-dev-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-dev.dev.svc,1433 -U sa -No
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

### Change and inspect the source

Discover the auth Secret referenced by `mssql-source`:

```shell
kubectl -n demo get mssqlserver mssql-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret mssql-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret mssql-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it mssql-source-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-source.demo.svc,1433 -U sa -No
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

The source retains the original customer `5`; the branch edit did not reach it.

### Check the branch again

Discover the auth Secret referenced by `mssql-dev`:

```shell
kubectl -n dev get mssqlserver mssql-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
mssql-dev-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n dev get secret mssql-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `sa`. Enter the decoded password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it mssql-dev-0 -c mssql -- bash
```

Inside the pod, start the database client through the primary Service. This also works when pod `-0` is a secondary:

```shell
/opt/mssql-tools18/bin/sqlcmd -S mssql-dev.dev.svc,1433 -U sa -No
```

Enter the password decoded above. `-No` makes client encryption optional for this `clientTLS: false` example. Run SQL batches at the `sqlcmd` prompt; `GO` submits each batch.

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

The branch retains its own customer `5` and the original customer `1`. The source update did not reach it.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch mssql-dev --wait=true --timeout=30m
kubectl -n demo delete mssqlserver mssql-source --wait=true --timeout=30m
```

After both databases are deleted, remove the tutorial configuration and issuers:

```shell
kubectl -n demo delete secret mssql-source-config
kubectl -n demo delete -f mssql-source-issuer.yaml
kubectl -n dev delete -f mssql-source-issuer.yaml
kubectl -n demo delete secret mssql-source-ca
kubectl -n dev delete secret mssql-source-ca
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/mssqlserver/branch/refresh.md).
- [Reset the branch root password](/docs/guides/mssqlserver/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/mssqlserver/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/mssqlserver/branch/customization/index.md).
