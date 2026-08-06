---
title: MSSQLServerMigration CRD
menu:
  docs_{{ .version }}:
    identifier: ms-migrator-concepts
    name: MSSQLServerMigration
    parent: ms-concepts
    weight: 27
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServerMigration

## What is MSSQLServerMigration

`MSSQLServerMigration` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to
migrate an existing database — such as one running on an external or managed instance — into a
KubeDB-managed database. You only need to describe the source and target databases in a `MSSQLServerMigration`
object, and the kubedb-courier operator will run the migration Job that copies the data and keeps
the target in sync until you cut over.

`MSSQLServerMigration` is the MSSQL-specific migration CRD (`courier.kubedb.com/v1alpha1`) whose
`spec.source` and `spec.target` describe the MSSQL source and target directly. KubeDB uses
[SqlPackage](https://learn.microsoft.com/en-us/sql/tools/sqlpackage/sqlpackage) for schema migration
and built-in CDC (change data capture) for streaming changes.

## MSSQLServerMigration Spec

As with all other Kubernetes objects, a `MSSQLServerMigration` needs `apiVersion`, `kind`, and `metadata`
fields. It also needs a `.spec` section. Below is an example `MSSQLServerMigration` object for migrating an
MSSQL Server database.

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
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 10001
```

### spec.source

`spec.source` is a required field that describes the MSSQL Server database being migrated **from**.

### spec.target

`spec.target` is a required field that describes the KubeDB-managed MSSQL Server database being migrated **into**.

### spec.source.connectionInfo / spec.target.connectionInfo

`connectionInfo` tells the migration how to connect to the MSSQL Server instance.

- `appBinding` — references an `AppBinding` that holds the connection information for this MSSQL
  instance. An `AppBinding` is a KubeDB resource that decouples the connection details (endpoint,
  credentials, TLS) from the consumer; create one with the necessary information and reference it
  here. This is currently the only supported way to provide connection details.
  - `name` — name of the AppBinding.
  - `namespace` — namespace of the AppBinding.
- `database` — the database used as the initial connection entry point (typically `master`).
- `maxConnections` — limits the number of concurrent connections the migration opens to this MSSQL
  instance.
- `encrypt` — encrypts the connection to the MSSQL Server instance.
- `trustServerCertificate` — when `true`, trusts the server's TLS certificate without validating it
  against a CA. Use only for testing with self-signed certificates — for production connections,
  leave this `false` and provide a CA certificate via the AppBinding's `tlsSecret` instead.

> For a `KubeDB`-managed database, an `AppBinding` is created by default, so you usually only need to
> create one for the source. Learn more about [AppBinding](/docs/guides/mssqlserver/concepts/appbinding.md).

### spec.source.schema

`schema` configures the schema migration phase. All fields are optional unless noted.

- `enabled` — enables or disables the schema migration phase. Defaults to `true`.
- `database` — list of database names to include in the schema migration. When empty, all user
  databases are included.
- `schema` — list of SQL Server schemas (e.g. `dbo`) to include.
- `excludeSchema` — list of SQL Server schemas to exclude.
- `table` — list of schema-qualified tables (e.g. `dbo.Users`) to include.
- `excludeTable` — list of schema-qualified tables to exclude.

### spec.source.snapshot

`snapshot` configures the bulk data copy phase. All fields are optional unless noted.

- `enabled` — enables or disables the snapshot phase. Defaults to `true`.
- `pipeline` — controls the parallelism of the snapshot phase:
  - `workers` — number of reader workers reading data from the source. Defaults to `3`.
  - `sinkers` — number of writer workers writing data to the target. Defaults to `5`.
  - `buffer` — buffer size per worker. Defaults to `16`.
  - `read_batch_size` — number of rows to read in a single batch from the source. Defaults to `1000`.
  - `write_batch_size` — number of rows to write in a single batch to the target. Defaults to `100`.

### spec.source.streaming

`streaming` configures the CDC (change data capture) streaming phase. All fields are optional unless noted.

- `enabled` — enables or disables the streaming phase. Defaults to `true`.
- `pollInterval` — how often CDC changes are polled from the source.
- `autoEnableCDC` — when `true`, the migration enables CDC on the source database and on every
  selected table using the standard `sys.sp_cdc_enable_db` / `sys.sp_cdc_enable_table` procedures.
  These require the `sysadmin` role and are **not** available on AWS RDS for SQL Server — on RDS,
  pre-enable CDC yourself with `EXEC msdb.dbo.rds_cdc_enable_db '<database>'` and set
  `autoEnableCDC: false`. When `false`, CDC must already be enabled on the source before the
  migration starts.
- `batchSize` — number of CDC change rows to batch together per write to the target. Defaults to `1000`.

### spec.jobDefaults

`spec.jobDefaults` is an optional field that sets default settings for the migration Job.

- `imagePullPolicy` — the image pull policy for the migration Job. Defaults to `IfNotPresent`.
- `backoffLimit` — the number of retries before the Job is marked as failed. Defaults to `6`.
- `ttlSecondsAfterFinished` — the TTL (in seconds) for cleaning up a completed Job.
- `activeDeadlineSeconds` — the duration (in seconds) relative to its start time that the Job may be
  active before the system tries to terminate it.

### spec.jobTemplate

`spec.jobTemplate` is an optional field that holds runtime configuration for the migration Job pod
(a `PodTemplateSpec`). Use it to set pod-level settings such as `securityContext`, `nodeSelector`,
`resources`, `serviceAccountName`, and so on.

## MSSQLServerMigration Phases

The migration proceeds through three phases:

1. **Schema** — copies the database schema (tables, indexes, stored procedures, etc.) from source to
   target using SqlPackage.
2. **Snapshot** — bulk copies all existing row data from source to target using parallel workers and
   sinkers.
3. **Streaming** — tails the MSSQL transaction log using CDC and continuously applies changes to the
   target, keeping it in sync with the source until cutover.

These phases run sequentially: schema first, then snapshot, then streaming.

## MSSQLServerMigration Status

`status` reflects the observed state of the migration.

- `status.phase` — the current phase of the migration. One of:
  - `Pending` — the migration has not started yet.
  - `Running` — the migration is in progress.
  - `Succeeded` — the migration completed successfully.
  - `Failed` — the migration failed.
- `status.progress` — the current progress of the migration:
  - `dbType` — the type of database being migrated.
  - `info` — additional progress information, including the current `Stage`, `Lag`, and `Progress`
    (these are surfaced as columns in `kubectl get mssqlservermigrations`).
- `status.conditions` — an array of conditions describing the migration's state over time (for
  example, `MigratorJobTriggered`, `MigrationRunning`, `MigrationSucceeded`, `MigrationFailed`).

## Next Steps

- Migrate an MSSQL Server database step by step with the [MSSQL Server Database Migration](/docs/guides/mssqlserver/migration/databasemigration.md) guide.
- Learn about the [AppBinding](/docs/guides/mssqlserver/concepts/appbinding.md) concept.
- Learn about the MSSQLServer CRD [here](/docs/guides/mssqlserver/concepts/mssqlserver.md).
