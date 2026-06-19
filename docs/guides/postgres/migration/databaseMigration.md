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
  <summary>How to configure this on your source instance</summary>

  **Self-hosted PostgreSQL**
  ```sql
  -- Set WAL level to logical (requires a PostgreSQL restart to take effect)
  ALTER SYSTEM SET wal_level = 'logical';

  -- Grant the replication privilege to the migration user
  ALTER USER <migration-user> WITH REPLICATION;
  ```

  **AWS RDS**
  Set `rds.logical_replication = 1` in your [RDS Parameter Group](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts.General.FeatureSupport.LogicalReplication) and reboot the instance. Then grant the replication privilege via SQL as shown above.

  **Azure Database for PostgreSQL**
  Set `azure.replication_support = logical` under **Server Parameters** in the [Azure Portal](https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-logical) and restart the server. Then grant the replication privilege via SQL as shown above.

  **Google Cloud SQL**
  Enable the `cloudsql.logical_decoding` flag via the [Cloud Console database flags](https://cloud.google.com/sql/docs/postgres/replication/configure-logical-replication). Then grant the replication privilege via SQL as shown above.

  **CloudNativePG (CNPG)**
  Add `wal_level: logical` under `postgresql` parameters in the `Cluster` spec.

  </details>

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/postgres/concepts/appbinding/)
    - [PostgreSQL](/docs/guides/postgres/concepts/postgres)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Connection Information

First, create an authentication secret to communicate with the source PostgreSQL database:

```bash
$ kubectl create secret generic source-postgres-auth -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=<username> \
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
    url: "postgresql://host:port"
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
$ kubectl apply -f target-postgres.yaml
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
        dbName: postgres
        maxConnections: 100
      pgDump:
        schemaOnly: true
      logicalReplication:
        copyData: true
        publication:
          name: "pub"
        subscription:
          name: "sub"
  target:
    postgres:
      connectionInfo:
        appBinding:
          name: target-postgres
          namespace: demo
        dbName: postgres
        maxConnections: 100
```

Here,

**`spec.source` / `spec.target` — connectionInfo:**
- `appBinding.name` / `appBinding.namespace` — references the `AppBinding` for the source or target PostgreSQL instance.
- `dbName` — the database used as the initial connection entry point.
- `maxConnections` — limits the number of concurrent connections the migrator opens to this PostgreSQL instance.

**`spec.source.pgDump` — schema migration phase:**
- `schemaOnly: true` — uses `pg_dump` to extract and apply only the DDL (schema) to the target, without any row data.

**`spec.source.logicalReplication` — data copy and streaming phase:**
- `copyData: true` — performs an initial bulk copy of all table data to the target before streaming begins.
- `publication.name` — the name of the PostgreSQL publication created on the source database to track changes.
- `subscription.name` — the name of the PostgreSQL subscription created on the target database to receive those changes.

## Watch Migration Progress

Let's wait for the `LAG` to reach near zero. Run the following command to watch `Migrator` CR:

```bash
Every 2.0s: kubectl get migrator -n demo

NAME               PHASE     DBTYPE     STAGE       LAG   PROGRESS   AGE
postgres-migrate   Running   postgres   Streaming   0B               4h36m
```

## Cutover

Once the `LAG` drops to near zero, stop all writes to the source database. Wait until the `LAG` reaches exactly zero — at that point both databases are fully in sync.

Now delete the `Migrator` CR to stop the migration process:

```bash
$ kubectl delete migrator -n demo postgres-migrate
migrator.migrator.kubedb.com "postgres-migrate" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `PostgreSQL` database. The migration is complete.
