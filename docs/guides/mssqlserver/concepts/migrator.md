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

`connectionInfo` tells the migration how to connect to the MSSQL Server instance. There are two ways to provide the connection details — set **either** `appBinding` **or** direct connection parameters:

- `appBinding` — references an `AppBinding` that holds the connection information for this MSSQL
  instance. An `AppBinding` is a KubeDB resource that decouples the connection details (endpoint,
  credentials, TLS) from the consumer; create one with the necessary information and reference it
  here. This is the recommended approach.
  - `name` — name of the AppBinding.
  - `namespace` — namespace of the AppBinding.
- `database` — the database used as the initial connection entry point (typically `master`).
- `maxConnections` — limits the number of concurrent connections the migration opens to this MSSQL
  instance.
- `tls` — paths to PEM files for a TLS-enabled connection. You can set the following fields:
  - `caFile` — path to the PEM-encoded CA certificate file.
  - `certFile` — path to the PEM-encoded client certificate (for mutual TLS).
  - `keyFile` — path to the PEM-encoded client private key (for mutual TLS).
  - `insecureSkipVerify` — disables server certificate and hostname verification.
  - `serverName` — overrides the hostname used for TLS SNI and certificate verification.

> For a `KubeDB`-managed database, an `AppBinding` is created by default, so you usually only need to
> create one for the source. Learn more about [AppBinding](/docs/guides/mssqlserver/concepts/appbinding.md).

### spec.source.schema

`schema` configures the schema migration phase. All fields are optional unless noted.

- `enabled` — enables or disables the schema migration phase. Defaults to `true`.
- `database` — list of database names to include in the schema migration. When empty, all user
  databases are included.
- `excludeDatabase` — list of database names to exclude from the schema migration.

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
- `autoEnableCDC` — when `true`, the migration automatically enables CDC on the source database
  and on every selected table. Set to `false` if CDC is already pre-configured on the source.
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

- Migrate an MSSQL Server database step by step with the [MSSQL Server Database Migration](/docs/guides/mssqlserver/migration/databaseMigration.md) guide.
- Learn about the [AppBinding](/docs/guides/mssqlserver/concepts/appbinding.md) concept.
- Learn about the MSSQLServer CRD [here](/docs/guides/mssqlserver/concepts/mssqlserver.md).
