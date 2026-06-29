---
title: PostgreSQL Database Migration Guide
menu:
  docs_{{ .version }}:
    identifier: pg-migration-database
    name: PostgreSQL Database Migration
    parent: pg-migration
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL Database Migration

This guide will show you how to use `KubeDB` Migrator to migrate an existing `PostgreSQL` database to a KubeDB-managed `PostgreSQL` instance with minimal downtime. The tool supports migration from a wide range of source environments — including Amazon RDS, CloudNativePG (CNPG), Zalando PostgreSQL Operator, Bitnami Helm charts, and self-hosted PostgreSQL instances.

The migration operates in three phases:

1. **Schema migration** — extracts DDL (tables, indexes, functions, etc.) from the source using `pg_dump`.
2. **Initial data copy** — performs a full bulk copy of all table data to the target.
3. **Live streaming** — uses PostgreSQL logical replication to continuously apply source changes to the target, keeping both databases in sync until cutover.

A brief downtime occurs only during the final cutover when application endpoints are redirected to the target database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` operator with the Migrator operator enabled in your cluster following the steps [here](/docs/operatormanual/migration/).

- The source `PostgreSQL` instance must be network-reachable from within your Kubernetes cluster.

- The source `PostgreSQL` instance must have `wal_level` set to `logical`. The database user provided for migration must have the `REPLICATION` privilege. There is no single procedure to configure this — it depends on your deployment environment.

  <details>
  <summary>How to configure this on your source instance?</summary>

  **Self-hosted PostgreSQL**

  ```sql
  -- Set WAL level to logical (requires a PostgreSQL restart to take effect)
  ALTER SYSTEM SET wal_level = 'logical';

  -- Grant the replication privilege to the migration user
  ALTER USER <migration-user> WITH REPLICATION;
  ```

  **AWS RDS** <br>

  Set `rds.logical_replication = 1` in your [RDS Parameter Group](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts.General.FeatureSupport.LogicalReplication) and reboot the instance. Then grant the replication privilege via SQL as shown above.

  <br> <br> **Azure Database for PostgreSQL** <br>

  Set `azure.replication_support = logical` under **Server Parameters** in the [Azure Portal](https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-logical) and restart the server. Then grant the replication privilege via SQL as shown above.

  <br> <br> **Google Cloud SQL** <br>

  Enable the `cloudsql.logical_decoding` flag via the [Cloud Console database flags](https://cloud.google.com/sql/docs/postgres/replication/configure-logical-replication). Then grant the replication privilege via SQL as shown above.

  <br> <br> **CloudNativePG (CNPG)** <br>

  Add `wal_level: logical` under `postgresql` parameters in the `Cluster` spec.

  </details>

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/postgres/concepts/appbinding/)
    - [PostgreSQL](/docs/guides/postgres/concepts/postgres)
    - [Migrator](/docs/guides/postgres/concepts/migrator)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Database

We will use an **AWS RDS PostgreSQL** instance as the source. Connect to it as the admin user to verify the prerequisites and set up the migration user and test data.

### Verify prerequisites

```bash
$ psql -h <rds-endpoint>.rds.amazonaws.com -U postgres -p 5432
```

```sql
SHOW wal_level;
 wal_level
-----------
 logical
(1 row)
```

> **Note:** On AWS RDS, set `rds.logical_replication = 1` in a custom RDS Parameter Group and reboot the instance. See the prerequisites section above for details.

### Create a dedicated migration user

```sql
-- Create the user
CREATE USER migrator WITH PASSWORD '<password>';

-- Logical replication: allows creating publications and replication slots
-- (AWS RDS-specific role; on self-hosted PostgreSQL use: ALTER ROLE migrator WITH REPLICATION;)
GRANT rds_replication TO migrator;

-- pg_dump access: read all schemas and table data (PostgreSQL 14+)
GRANT pg_read_all_data TO migrator;

-- Verify
\du migrator
```

> **Note (AWS RDS):** `rds_replication` is the AWS-managed equivalent of the `REPLICATION` privilege. `pg_read_all_data` (PostgreSQL 14+) covers all `SELECT` access needed by `pg_dump` across all schemas and tables — no per-table grants needed. For PostgreSQL 13 or earlier, grant `SELECT` on each table explicitly.

The `migrator` user is referenced in the Kubernetes secret and AppBinding for the rest of this guide.

### Create the source database and seed data

```sql
-- Create and switch to the shop database
CREATE DATABASE shop;
\c shop

-- Create the orders table
CREATE TABLE orders (
  id            SERIAL PRIMARY KEY,
  customer_name VARCHAR(100) NOT NULL,
  product       VARCHAR(100) NOT NULL,
  quantity      INT          NOT NULL DEFAULT 1,
  status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
  created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- Allow the migrator user to connect to this database
-- (pg_read_all_data already covers USAGE on schemas and SELECT on all tables/sequences)
GRANT CONNECT ON DATABASE shop TO migrator;

-- Seed initial data
INSERT INTO orders (customer_name, product, quantity, status) VALUES
  ('Alice', 'Laptop',     1, 'shipped'),
  ('Bob',   'Headphones', 2, 'pending'),
  ('Carol', 'Keyboard',   3, 'delivered');

SELECT * FROM orders;
 id | customer_name |  product   | quantity |  status   |       created_at
----+---------------+------------+----------+-----------+------------------------
  1 | Alice         | Laptop     |        1 | shipped   | 2026-06-29 08:00:00
  2 | Bob           | Headphones |        2 | pending   | 2026-06-29 08:00:01
  3 | Carol         | Keyboard   |        3 | delivered | 2026-06-29 08:00:02
(3 rows)
```

## Prepare Source Connection Information

First, create an authentication secret using the `migrator` user credentials:

```bash
$ kubectl create secret generic source-postgres-auth -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=migrator \
                --from-literal=password=<password>
```

If your database has TLS enabled, create a secret with the CA certificate:

```bash
kubectl create secret generic ca-secret \
  --from-file=ca.crt=$CERT_PATH/ca.crt \
  --namespace=demo
```

Now create an `AppBinding` with the necessary information. The Migrator operator reads the source PostgreSQL connection information from this AppBinding CR. Use the following YAML to create your AppBinding:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: source-postgres
  namespace: demo
spec:
  type: postgresql
  version: "17.4"
  clientConfig:
    url: "postgresql://<rds-endpoint>.rds.amazonaws.com:5432"
  secret:
    name: source-postgres-auth
  tlsSecret: # omit if TLS is disabled
    name: ca-secret
```

Here,

- `spec.clientConfig.url` is the connection URL of the source PostgreSQL instance.
- `spec.secret.name` is the reference to the secret we created earlier, containing the PostgreSQL authentication information.

> For a `KubeDB`-managed database, an `AppBinding` is created by default. So there is no need to create one for the target database.

## Create Target PostgreSQL Database

KubeDB implements a `Postgres` CRD to define the specification of a PostgreSQL database. Follow the `Postgres` object to create the target database.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: target-postgres
  namespace: demo
spec:
  version: "17.4"
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 20Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/migration/target-postgres.yaml
postgres.kubedb.com/target-postgres created
```

> Note: Adjust the `resources.requests.storage` based on the source database size.

Wait until `target-postgres` has status `Ready`.

## Apply Migrator CR

To migrate the database we have to create a `Migrator` CR. Below is the YAML of the `Migrator` CR that we are going to create:

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: postgres-migrate
  namespace: demo
spec:
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
  source:
    postgres:
      connectionInfo:
        appbinding:
          name: source-postgres
          namespace: demo
        dbName: shop
        maxConnections: 100
      pgDump:
        schemaOnly: true
      logicalReplication:
        copyData: true
        publication:
          name: "pub"
          mode: default
        subscription:
          name: "sub"
  target:
    postgres:
      connectionInfo:
        appBinding:
          name: target-postgres
          namespace: demo
        dbName: shop
        maxConnections: 100
```

Here we connect to and migrate the `shop` database. Schema is extracted via `pg_dump` (`pgDump.schemaOnly: true`) and data is replicated using PostgreSQL logical replication with publication `pub` on the source and subscription `sub` on the target. For a full description of every field, see the [Migrator CRD reference](/docs/guides/postgres/concepts/migrator).

## Watch Migration Progress

Let's wait for the `LAG` to reach near zero. Run the following command to watch `Migrator` CR:

```bash
Every 2.0s: kubectl get migrator -n demo

NAME               PHASE     DBTYPE     STAGE       LAG   PROGRESS   AGE
postgres-migrate   Running   postgres   Streaming   0B               4h36m
```

### Verify initial snapshot on target

Once the migrator reaches the `Streaming` stage, exec into the KubeDB target pod and confirm all seed rows were copied over:

```bash
$ kubectl exec -it -n demo target-postgres-0 -- psql -U postgres -d shop
```

```sql
SELECT * FROM orders;
 id | customer_name |  product   | quantity |  status   |       created_at
----+---------------+------------+----------+-----------+------------------------
  1 | Alice         | Laptop     |        1 | shipped   | 2026-06-29 08:00:00
  2 | Bob           | Headphones |        2 | pending   | 2026-06-29 08:00:01
  3 | Carol         | Keyboard   |        3 | delivered | 2026-06-29 08:00:02
(3 rows)
```

### Test live CDC streaming

With the migrator still running, connect to the **source RDS** instance and run some DML:

```bash
$ psql -h <rds-endpoint>.rds.amazonaws.com -U migrator -d shop -p 5432
```

```sql
-- Insert a new order
INSERT INTO orders (customer_name, product, quantity, status)
VALUES ('Dave', 'Mouse', 1, 'pending');

-- Mark Bob's headphones as delivered
UPDATE orders SET status = 'delivered' WHERE id = 2;

-- Remove the already-shipped laptop order
DELETE FROM orders WHERE id = 1;
```

Wait a few seconds for logical replication to propagate, then re-query the **target**:

```sql
SELECT * FROM orders;
 id | customer_name |  product   | quantity |  status   |       created_at
----+---------------+------------+----------+-----------+------------------------
  2 | Bob           | Headphones |        2 | delivered | 2026-06-29 08:00:01
  3 | Carol         | Keyboard   |        3 | delivered | 2026-06-29 08:00:02
  4 | Dave          | Mouse      |        1 | pending   | 2026-06-29 08:10:00
(3 rows)
```

The INSERT, UPDATE, and DELETE are all reflected on the target — logical replication is working correctly.

## Cutover

Once the `LAG` drops to near zero, stop all writes to the source database. Wait until the `LAG` reaches exactly zero — at that point both databases are fully in sync.

Now delete the `Migrator` CR to stop the migration process:

```bash
$ kubectl delete migrator -n demo postgres-migrate
migrator.migrator.kubedb.com "postgres-migrate" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `PostgreSQL` database. The migration is complete.
