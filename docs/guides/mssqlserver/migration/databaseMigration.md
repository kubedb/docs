---
title: MSSQL Server Database Migration Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mssqlserver-migration-database
    name: MSSQL Server Database Migration
    parent: guides-mssqlserver-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQL Server Database Migration

This guide will show you how to use `KubeDB` Migration to migrate an existing `MSSQL Server` database — such as one running on AWS RDS for SQL Server or any external instance — entirely into a KubeDB-managed `MSSQLServer` with minimal downtime.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` operator with the kubedb-courier operator enabled in your cluster following the steps [here](/docs/operatormanual/migration/).

- The source `MSSQL Server` instance must be network-reachable from within your Kubernetes cluster.

- The source `MSSQL Server` database should have the `sa` user or a user with `sysadmin` privileges for the migration.

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/mssqlserver/concepts/appbinding.md)
    - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
    - [MSSQLServerMigration](/docs/guides/mssqlserver/concepts/migrator.md)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Database

We will use an **AWS RDS for SQL Server** instance as the source. Below is how to verify the prerequisites, set up the migration user, and insert test data.

<details>
<summary><b>Configuring your source instance.</b></summary>

<br> **AWS RDS for SQL Server** <br>

AWS RDS for SQL Server provides a fully managed instance. Refer to the [AWS documentation](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html) for creating an RDS SQL Server instance. For migration, you need:

- The RDS endpoint (e.g., `mydb.xxxx.us-east-1.rds.amazonaws.com`)
- The `sa` user password (or a user with `sysadmin` privileges)
- The source must be accessible from your Kubernetes cluster (configure the VPC security group to allow inbound connections on port `1433` from your cluster's CIDR range)

**Self-hosted SQL Server** <br>

For a self-hosted SQL Server instance, make sure TCP/IP is enabled on port `1433` and the instance is accessible from your Kubernetes cluster. Use the `sa` account or create a login with `sysadmin` server role.

**Azure SQL Managed Instance** <br>

Azure SQL Managed Instance works similarly with a public or private endpoint. Use the `sa` user credentials for the migration.

See the official [AWS RDS SQL Server](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html) docs for more details.

</details>

### Create a database and seed data

Connect to the source instance using `sqlcmd` or your preferred SQL client:

```bash
$ sqlcmd -S <rds-endpoint> -U sa -P '<password>' -C
```

```sql
-- Create a test database
CREATE DATABASE RestaurantMigrationDB;
GO

-- Switch to the new database
USE RestaurantMigrationDB;
GO

-- Create a table
CREATE TABLE Customers (
    CustomerID INT IDENTITY(1,1) PRIMARY KEY,
    Name NVARCHAR(100) NOT NULL,
    Email NVARCHAR(100),
    City NVARCHAR(50),
    CreatedAt DATETIME2 DEFAULT GETDATE()
);
GO

-- Insert sample data
INSERT INTO Customers (Name, Email, City) VALUES
('Alice', 'alice@example.com', 'NYC'),
('Bob', 'bob@example.com', 'SF'),
('Carol', 'carol@example.com', 'Chicago');
GO

-- Verify
SELECT * FROM Customers;
GO
```

Expected output:

```text
CustomerID  Name   Email              City     CreatedAt
----------- ------ ------------------ -------- -----------------------
1           Alice  alice@example.com  NYC      2026-07-10 12:00:00.000
2           Bob    bob@example.com    SF       2026-07-10 12:00:00.000
3           Carol  carol@example.com  Chicago  2026-07-10 12:00:00.000
```

## Prepare Source Connection Information

First, create an authentication secret using the source `sa` user credentials:

```bash
$ kubectl create secret generic source-mssql-auth -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=sa \
                --from-literal=password=<password>
```

If your source database has TLS enabled (RDS and Azure SQL use TLS by default), create a secret with the CA certificate:

```bash
kubectl create secret generic ca-secret \
  --from-file=ca.crt=<path-to-ca-cert> \
  --namespace=demo
```

> **Note:** For mTLS, also include the client certificate and key by appending <br> `--from-file=tls.crt=<path-to-client-cert>` <br> `--from-file=tls.key=<path-to-client-key>` <br> to the command above.

Now create an `AppBinding` with the necessary information. The kubedb-courier operator reads the source MSSQL Server connection information from this AppBinding CR. Use the following YAML to create your AppBinding:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mssqlserver-source
  namespace: demo
spec:
  type: kubedb.com/mssqlserver
  clientConfig:
    url: "sqlserver://<rds-endpoint>:1433"
  secret:
    name: source-mssql-auth
  tlsSecret: # omit if TLS is disabled
    name: ca-secret
```

Here,

- `spec.type` is set to `kubedb.com/mssqlserver`.
- `spec.clientConfig.url` is the connection URL of the source MSSQL Server instance in the format `sqlserver://<host>:<port>`.
- `spec.secret.name` references the secret we created earlier, containing the MSSQL authentication information.
- `spec.tlsSecret.name` references the TLS CA certificate secret (optional — omit if TLS is not required).

> For a `KubeDB`-managed database, an `AppBinding` is created by default. So there is no need to create one for the target database.

## Create a TLS Certificate Issuer (for the Target)

KubeDB-managed MSSQL Servers use TLS certificates by default. If you don't have a cert-manager issuer in the `demo` namespace, create one:

```bash
# Create a self-signed CA
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ca.key -out ca.crt \
    -subj "/CN=mssqlserver-ca"

# Create a TLS secret
$ kubectl create secret tls mssqlserver-ca \
    --cert=ca.crt --key=ca.key \
    -n demo
```

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mssqlserver-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: mssqlserver-ca
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/migration/source-issuer.yaml
```

> **Note:** This step is only required if you don't already have a cert-manager issuer configured in the namespace. KubeDB uses cert-manager to issue TLS certificates for the MSSQL Server pods.

## Create Target MSSQL Server Database

KubeDB implements a `MSSQLServer` CRD to define the specification of an MSSQL Server database. Use the following `MSSQLServer` object to create the target database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-standalone
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Developer
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 20Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/migration/target-mssqlserver.yaml
mssqlserver.kubedb.com/mssqlserver-standalone created
```

> Note: Adjust the `resources.requests.storage` based on the source database size.

Wait until `mssqlserver-standalone` has status `Ready`.

```bash
$ kubectl get mssqlserver -n demo
NAME                     VERSION     STATUS   AGE
mssqlserver-standalone   2022-cu12   Ready    5m
```

## Apply MSSQLServerMigration CR

To migrate the database we have to create an `MSSQLServerMigration` CR. KubeDB uses [SqlPackage](https://learn.microsoft.com/en-us/sql/tools/sqlpackage/sqlpackage) for schema migration and built-in CDC for streaming changes. Below is the YAML of the `MSSQLServerMigration` CR:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: MSSQLServerMigration
metadata:
  name: mssqlserver-migration
  namespace: demo
spec:
  source:
    connectionInfo:
      appBinding:
        name: mssqlserver-source
        namespace: demo
      database: master
    schema:
      enabled: true
      database:
        - RestaurantMigrationDB
    snapshot:
      enabled: true
      pipeline:
        workers: 3
        sinkers: 5
        buffer: 16
        read_batch_size: 1000
        write_batch_size: 100
    streaming:
      enabled: true
      autoEnableCDC: true
      batchSize: 1000
  target:
    connectionInfo:
      appBinding:
        name: mssqlserver-standalone
        namespace: demo
      database: master
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/migration/mssqlserver-migration.yaml
mssqlservermigration.courier.kubedb.com/mssqlserver-migration created
```

Here,

**`spec.source` / `spec.target` — connectionInfo:**
- `appBinding.name` / `appBinding.namespace` — references the `AppBinding` for the source or target MSSQL Server instance.
- `database` — the database used as the initial connection entry point (typically `master`).

**`spec.source.schema` — schema migration:**
- `enabled: true` — enables the schema migration phase.
- `database` — list of databases to migrate. The schema (tables, indexes, stored procedures, etc.) is copied from source to target.

**`spec.source.snapshot` — bulk data copy:**
- `enabled: true` — enables the snapshot phase.
- `pipeline.workers` — number of parallel reader workers (default: 3).
- `pipeline.sinkers` — number of parallel writer workers (default: 5).

**`spec.source.streaming` — CDC change streaming:**
- `enabled: true` — enables the streaming phase.
- `autoEnableCDC: true` — automatically enables CDC on the source database and tables.

For a full description of every field, see the [MSSQLServerMigration CRD reference](/docs/guides/mssqlserver/concepts/migrator.md).

## Watch Migration Progress

Let's wait for the Migration to finish the schema and snapshot phases and enter the streaming phase. Run the following command to watch `MSSQLServerMigration` CR:

```bash
$ watch kubectl get mssqlservermigrations -n demo
```

During the **Schema** stage:

```bash
NAME                    PHASE     STAGE    LAG   PROGRESS   AGE
mssqlserver-migration   Running   Schema          0.0       12s
```

During the **Snapshot** stage, you'll see progress advancing to 100%:

```bash
NAME                    PHASE     STAGE      LAG   PROGRESS   AGE
mssqlserver-migration   Running   Snapshot         45.3       45s
```

When the **Streaming** stage begins with `LAG` at zero, both databases are fully in sync:

```bash
NAME                    PHASE     STAGE       LAG   PROGRESS   AGE
mssqlserver-migration   Running   Streaming   0     100.0      2m
```

### View detailed progress via pod logs

You can also see stage-wise progress and per-database details by checking the migration pod logs:

```bash
$ kubectl logs -n demo migrator-<migration-pod-name>
```

Example output during the snapshot stage:

```log
2026-07-10T12:00:30.632Z  INFO  mssqlserver/mssqlserver.go:144  Starting snapshot for database  {"database": "RestaurantMigrationDB"}
2026-07-10T12:00:45.123Z  INFO  mssqlserver/mssqlserver.go:148  Completed snapshot for database  {"database": "RestaurantMigrationDB"}
```

Example output during streaming — showing per-database CDC lag:

```log
2026-07-10T12:01:30.632Z  INFO  mssqlserver/mssqlserver.go:329  Starting CDC streaming for database  {"database": "RestaurantMigrationDB"}
```

### Verify initial snapshot on target

Once the migration reaches the `Streaming` stage, exec into the KubeDB target pod and confirm all seed documents were copied over:

```bash
$ kubectl exec -it -n demo mssqlserver-standalone-0 -- /opt/mssql-tools18/bin/sqlcmd \
    -S localhost -U sa -P '<sa-password>' -C -Q "SELECT * FROM RestaurantMigrationDB.dbo.Customers"
```

Expected output:

```text
CustomerID  Name   Email              City     CreatedAt
----------- ------ ------------------ -------- -----------------------
1           Alice  alice@example.com  NYC      2026-07-10 12:00:00.000
2           Bob    bob@example.com    SF       2026-07-10 12:00:00.000
3           Carol  carol@example.com  Chicago  2026-07-10 12:00:00.000
```

> To get the `sa` password for the target MSSQL Server, run: <br> `kubectl get secret -n demo mssqlserver-standalone-auth -o jsonpath='{.data.password}' | base64 -d`

### Test live CDC streaming

With the migration still running, connect to the **source AWS RDS** instance and run some DML:

```bash
$ sqlcmd -S <rds-endpoint> -U sa -P '<password>' -C
```

```sql
USE RestaurantMigrationDB;
GO

-- Insert a new customer
INSERT INTO Customers (Name, Email, City) VALUES
('Dave', 'dave@example.com', 'Boston');
GO

-- Update Bob's city
UPDATE Customers SET City = 'Seattle' WHERE Name = 'Bob';
GO

-- Delete Alice
DELETE FROM Customers WHERE Name = 'Alice';
GO
```

Wait a few seconds for the CDC events to propagate, then re-query the **target**:

```bash
$ kubectl exec -it -n demo mssqlserver-standalone-0 -- /opt/mssql-tools18/bin/sqlcmd \
    -S localhost -U sa -P '<sa-password>' -C -Q "SELECT * FROM RestaurantMigrationDB.dbo.Customers"
```

Expected output:

```text
CustomerID  Name   Email              City     CreatedAt
----------- ------ ------------------ -------- -----------------------
2           Bob    bob@example.com    Seattle  2026-07-10 12:00:00.000
3           Carol  carol@example.com  Chicago  2026-07-10 12:00:00.000
4           Dave   dave@example.com   Boston   2026-07-10 12:15:00.000
```

The INSERT, UPDATE, and DELETE are all reflected on the target — CDC streaming is working correctly.

## Cutover

Once the `LAG` drops to near zero, stop all writes to the source database. Wait until the `LAG` reaches exactly zero — at that point both databases are fully in sync.

Now delete the `MSSQLServerMigration` CR to stop the migration process:

```bash
$ kubectl delete mssqlservermigrations -n demo mssqlserver-migration
mssqlservermigration.courier.kubedb.com "mssqlserver-migration" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `MSSQLServer` database. The migration is complete.
