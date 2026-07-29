---
title: PostgresMigration CRD
menu:
  docs_{{ .version }}:
    identifier: pg-migrator-concepts
    name: PostgresMigration
    parent: pg-concepts-postgres
    weight: 26
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgresMigration

## What is PostgresMigration

`PostgresMigration` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to
migrate an existing database — such as one running on an external or managed instance — into a
KubeDB-managed database. You only need to describe the source and target databases in a `PostgresMigration`
object, and the kubedb-courier operator will run the migration Job that copies the data and keeps
the target in sync until you cut over.

`PostgresMigration` is the PostgreSQL-specific migration CRD (`courier.kubedb.com/v1alpha1`) whose
`spec.source` and `spec.target` describe the PostgreSQL source and target directly.

## PostgresMigration Spec

As with all other Kubernetes objects, a `PostgresMigration` needs `apiVersion`, `kind`, and `metadata`
fields. It also needs a `.spec` section. Below is an example `PostgresMigration` object for migrating a
PostgreSQL database.

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: PostgresMigration
metadata:
  name: postgres-migrate
  namespace: demo
spec:
  source:
    connectionInfo:
      appBinding:
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
    connectionInfo:
      appBinding:
        name: target-postgres
        namespace: demo
      dbName: shop
      maxConnections: 100
  jobDefaults:
    imagePullPolicy: IfNotPresent
    backoffLimit: 6
    ttlSecondsAfterFinished: 3600
    activeDeadlineSeconds: 86400
  jobTemplate:
    spec:
      securityContext:
        fsGroup: 65534
```

### spec.source

`spec.source` is a required field that describes the database being migrated **from**. It holds the
PostgreSQL source fields directly under `spec.source`.

### spec.target

`spec.target` is a required field that describes the KubeDB-managed database being migrated **into**.
It holds the PostgreSQL target fields directly under `spec.target`.

### spec.source.connectionInfo

`connectionInfo` (also under `spec.target`) tells the PostgresMigration how to connect to the
PostgreSQL instance. There are two ways to provide the connection details — set **either**
`appBinding` **or** `url`:

- `appBinding` — references an `AppBinding` that holds the connection information for this PostgreSQL
  instance. An `AppBinding` is a KubeDB resource that decouples the connection details (endpoint,
  credentials, TLS) from the consumer; create one with the necessary information and reference it
  here. This is the recommended approach.
  - `name` — name of the AppBinding.
  - `namespace` — namespace of the AppBinding.
- `url` — the database connection string (for example
  `postgres://user:password@host:5432/postgres`). Use this as an alternative to `appBinding` when you
  want to provide the endpoint inline instead of through an AppBinding.
- `dbName` — the database used as the initial connection entry point.
- `maxConnections` — limits the number of concurrent connections the Migration opens to this
  PostgreSQL instance.
- `tls` — paths to PEM files for a TLS-enabled connection. You can set the following fields:
  - `caFile` — path to the PEM-encoded CA certificate file.
  - `certFile` — path to the PEM-encoded client certificate (for mutual TLS).
  - `keyFile` — path to the PEM-encoded client private key (for mutual TLS).
  - `insecureSkipVerify` — disables server certificate and hostname verification.
  - `serverName` — overrides the hostname used for TLS SNI and certificate verification.

> For a `KubeDB`-managed database, an `AppBinding` is created by default, so you usually only need to
> create one for the source. Learn more about [AppBinding](/docs/guides/postgres/concepts/appbinding.md).

### spec.source.pgDump

`pgDump` configures the schema migration phase, which uses `pg_dump` to extract object definitions
from the source. These fields map directly to `pg_dump` command-line options.

- `schemaOnly` — dump only the schema (DDL), no data. Equivalent to `pg_dump --schema-only`.
- `schema` — list of schemas to dump. Equivalent to `pg_dump --schema=<schema>`.
- `excludeSchema` — list of schemas to exclude. Equivalent to `pg_dump --exclude-schema=<schema>`.
- `table` — list of tables to dump. Equivalent to `pg_dump --table=<table>`.
- `excludeTable` — list of tables to exclude. Equivalent to `pg_dump --exclude-table=<table>`.
- `extraOptions` — additional raw `pg_dump` command-line flags not explicitly modeled by the fields
  above.

### spec.source.logicalReplication

`logicalReplication` configures the data copy and streaming phases using PostgreSQL
[logical replication](https://www.postgresql.org/docs/current/logical-replication.html).

- `copyData` — performs an initial bulk copy of all table data to the target before streaming begins.
  Defaults to `true`.
- `publication` — the publication created on the source to track changes:
  - `name` — identifier of the PostgreSQL publication.
  - `mode` — how tables are selected for the publication. One of:
    - `default` — manual selection with filtering behavior similar to `pg_dump`.
    - `table` — publishes only the specified tables (`FOR TABLE ...`).
    - `allTable` — publishes all tables in the database (`FOR ALL TABLES`).
    - `tableInSchema` — publishes all tables within the specified schemas (`FOR TABLES IN SCHEMA ...`).
  - `args` — additional publication parameters depending on `mode` (for example, table names when
    `mode=table`, or schema names when `mode=tableInSchema`).
- `subscription` — the subscription created on the target to receive those changes:
  - `name` — identifier of the PostgreSQL subscription.

### spec.jobDefaults

`spec.jobDefaults` is an optional field that sets default settings for the migration Job.

- `imagePullPolicy` — the image pull policy for the Migration Job. Defaults to `IfNotPresent`.
- `backoffLimit` — the number of retries before the Job is marked as failed. Defaults to `6`.
- `ttlSecondsAfterFinished` — the TTL (in seconds) for cleaning up a completed Job.
- `activeDeadlineSeconds` — the duration (in seconds) relative to its start time that the Job may be
  active before the system tries to terminate it.

### spec.jobTemplate

`spec.jobTemplate` is an optional field that holds runtime configuration for the migration Job pod
(a `PodTemplateSpec`). Use it to set pod-level settings such as `securityContext`, `nodeSelector`,
`resources`, `serviceAccountName`, and so on.

## PostgresMigration Status

`status` reflects the observed state of the migration.

- `status.phase` — the current phase of the migration. One of:
  - `Pending` — the migration has not started yet.
  - `Running` — the migration is in progress.
  - `Succeeded` — the migration completed successfully.
  - `Failed` — the migration failed.
- `status.progress` — the current progress of the migration:
  - `dbType` — the type of database being migrated.
  - `info` — additional progress information, including the current `Stage`, `Lag`, and `Progress`
    (these are surfaced as columns in `kubectl get postgresmigrations`).
- `status.conditions` — an array of conditions describing the migration's state over time (for
  example, `MigratorJobTriggered`, `MigrationRunning`, `MigrationSucceeded`, `MigrationFailed`).

## Next Steps

- Migrate a PostgreSQL database step by step with the [PostgreSQL Database Migration](/docs/guides/postgres/migration/databaseMigration.md) guide.
- Learn about the [AppBinding](/docs/guides/postgres/concepts/appbinding.md) concept.
- Learn about the Postgres CRD [here](/docs/guides/postgres/concepts/postgres.md).
