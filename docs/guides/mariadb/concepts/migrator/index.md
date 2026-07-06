---
title: Migration CRD
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-concepts-migrator
    name: Migration
    parent: guides-mariadb-concepts
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Migration

## What is Migration

`Migration` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to
migrate an existing database — such as one running on an external or managed instance — into a
KubeDB-managed database. You only need to describe the source and target databases in a `Migration`
object, and the kubedb-courier operator will run the migration Job that copies the data and keeps
the target in sync until you cut over.

`Migration` is a single shared CRD (`courier.kubedb.com/v1alpha1`). Its `spec.source` and
`spec.target` each carry a per-database sub-spec (`mysql`, `mariadb`, `postgres`, `mongodb`). This
page describes the `Migration` object for a **MariaDB** source and target.

## Migration Spec

As with all other Kubernetes objects, a `Migration` needs `apiVersion`, `kind`, and `metadata`
fields. It also needs a `.spec` section. Below is an example `Migration` object for migrating a
MariaDB database.

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Migration
metadata:
  name: mariadb-migrate
  namespace: demo
spec:
  source:
    mariadb:
      connectionInfo:
        appBinding:
          name: source-mariadb
          namespace: demo
        dbName: "mysql"
        maxConnections: 100
      schema:
        enabled: true
        database: [] # databases to include
        excludeDatabase: [] # databases to exclude
      snapshot:
        enabled: true
        pipeline:
          workers: 3
          sinkers: 4
          buffer: 12
          read_batch_size: 1000
          write_batch_size: 200
      streaming:
        enabled: true
  target:
    mariadb:
      connectionInfo:
        appBinding:
          name: target-mariadb
          namespace: demo
        dbName: "mysql"
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

`spec.source` is a required field that describes the database being migrated **from**. It holds a
per-database sub-spec; for a MariaDB migration you set `spec.source.mariadb`.

### spec.target

`spec.target` is a required field that describes the KubeDB-managed database being migrated **into**.
It holds a per-database sub-spec; for a MariaDB migration you set `spec.target.mariadb`.

### spec.source.mariadb.connectionInfo

`connectionInfo` (also under `spec.target.mariadb`) tells the Migration how to connect to the MariaDB
instance. There are two ways to provide the connection details — set **either** `appBinding` **or**
`url`:

- `appBinding` — references an `AppBinding` that holds the connection information for this MariaDB
  instance. An `AppBinding` is a KubeDB resource that decouples the connection details (endpoint,
  credentials, TLS) from the consumer; create one with the necessary information and reference it
  here. This is the recommended approach.
  - `name` — name of the AppBinding.
  - `namespace` — namespace of the AppBinding.
- `url` — the database connection string (for example `mysql://user:password@host:3306/`). Use this
  as an alternative to `appBinding` when you want to provide the endpoint inline instead of through an
  AppBinding.
- `dbName` — the internal database used as the initial connection entry point.
- `maxConnections` — limits the number of concurrent connections the Migration opens to this MariaDB
  instance.
- `tls` — paths to PEM files for a TLS-enabled connection. You can set the following fields:
  - `caFile` — path to the PEM-encoded CA certificate file.
  - `certFile` — path to the PEM-encoded client certificate (for mutual TLS).
  - `keyFile` — path to the PEM-encoded client private key (for mutual TLS).
  - `insecureSkipVerify` — disables server certificate and hostname verification.
  - `serverName` — overrides the hostname used for TLS SNI and certificate verification.

> For a `KubeDB`-managed database, an `AppBinding` is created by default, so you usually only need to
> create one for the source. Learn more about [AppBinding](/docs/guides/mariadb/concepts/appbinding/).

### spec.source.mariadb.schema

`schema` configures the schema migration phase, which recreates database and table definitions on the
target before any data is copied.

- `enabled` — enables the schema migration phase.
- `database` — list of databases to include. An empty list means all databases except the system
  databases (`mysql`, `information_schema`, `performance_schema`, `sys`).
- `excludeDatabase` — list of databases to exclude from migration.

### spec.source.mariadb.snapshot

`snapshot` configures the initial bulk snapshot phase, which copies the existing rows from the source
to the target.

- `enabled` — enables the bulk snapshot phase.
- `pipeline` — tunes the parallel copy pipeline:
  - `workers` — number of parallel workers, each processing a separate table concurrently. Defaults
    to `3`.
  - `sinkers` — number of parallel write workers pushing data to the target for each worker. Defaults
    to `3`.
  - `buffer` — size of the in-memory queue (in records) between readers and writers. Defaults to `10`.
  - `read_batch_size` — number of rows fetched per read batch from the source. Defaults to `5000`.
  - `write_batch_size` — number of rows written per batch to the target. Defaults to `500`.

### spec.source.mariadb.streaming

`streaming` configures the change-data-capture (CDC) phase.

- `enabled` — enables CDC streaming after the snapshot completes, keeping the target continuously in
  sync with ongoing changes on the source until you cut over.

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

## Migration Status

`status` reflects the observed state of the migration.

- `status.phase` — the current phase of the migration. One of:
  - `Pending` — the migration has not started yet.
  - `Running` — the migration is in progress.
  - `Succeeded` — the migration completed successfully.
  - `Failed` — the migration failed.
- `status.progress` — the current progress of the migration:
  - `dbType` — the type of database being migrated.
  - `info` — additional progress information, including the current `Stage`, `Lag`, and `Progress`
    (these are surfaced as columns in `kubectl get migration`).
- `status.conditions` — an array of conditions describing the migration's state over time (for
  example, `MigratorJobTriggered`, `MigrationRunning`, `MigrationSucceeded`, `MigrationFailed`).

## Next Steps

- Migrate a MariaDB database step by step with the [MariaDB Database Migration](/docs/guides/mariadb/migration/databaseMigration.md) guide.
- Learn about the [AppBinding](/docs/guides/mariadb/concepts/appbinding/) concept.
- Learn about the MariaDB CRD [here](/docs/guides/mariadb/concepts/mariadb).
