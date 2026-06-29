---
title: Migrator CRD
menu:
  docs_{{ .version }}:
    identifier: mg-migrator-concepts
    name: Migrator
    parent: mg-concepts-mongodb
    weight: 27
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Migrator

## What is Migrator

`Migrator` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to
migrate an existing database — such as one running on an external or managed instance — into a
KubeDB-managed database. You only need to describe the source and target databases in a `Migrator`
object, and the KubeDB Migrator operator will run the migration Job that copies the data and keeps
the target in sync until you cut over.

`Migrator` is a single shared CRD (`migrator.kubedb.com/v1alpha1`). Its `spec.source` and
`spec.target` each carry a per-database sub-spec (`mysql`, `mariadb`, `postgres`, `mongodb`). This
page describes the `Migrator` object for a **MongoDB** source and target. KubeDB uses
[mongoshake](https://github.com/alibaba/MongoShake) to perform MongoDB migrations.

## Migrator Spec

As with all other Kubernetes objects, a `Migrator` needs `apiVersion`, `kind`, and `metadata`
fields. It also needs a `.spec` section. Below is an example `Migrator` object for migrating a
MongoDB database.

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: mongodb-migrate
  namespace: demo
spec:
  source:
    mongodb:
      connectionInfo:
        appBinding:
          name: mgo-source
          namespace: demo
      mongoshake:
        syncMode: all
        extraConfiguration:
          full_sync.executor.insert_on_dup_update: "true"
  target:
    mongodb:
      connectionInfo:
        appBinding:
          name: mgo-destination
          namespace: demo
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
per-database sub-spec; for a MongoDB migration you set `spec.source.mongodb`.

### spec.target

`spec.target` is a required field that describes the KubeDB-managed database being migrated **into**.
It holds a per-database sub-spec; for a MongoDB migration you set `spec.target.mongodb`.

### spec.source.mongodb.connectionInfo

`connectionInfo` (also under `spec.target.mongodb`) tells the Migrator how to connect to the MongoDB
instance. There are two ways to provide the connection details — set **either** `appBinding` **or**
`url`:

- `appBinding` — references an `AppBinding` that holds the connection information for this MongoDB
  instance. An `AppBinding` is a KubeDB resource that decouples the connection details (endpoint,
  credentials, TLS) from the consumer; create one with the necessary information and reference it
  here. This is the recommended approach.
  - `name` — name of the AppBinding.
  - `namespace` — namespace of the AppBinding.
- `url` — the database connection string (for example `mongodb://user:password@host:27017/`). Use
  this as an alternative to `appBinding` when you want to provide the endpoint inline instead of
  through an AppBinding.
- `dbName` — the database used as the initial connection entry point.
- `maxConnections` — limits the number of concurrent connections the Migrator opens to this MongoDB
  instance.
- `tls` — paths to PEM files for a TLS-enabled connection. You can set the following fields:
  - `caFile` — path to the PEM-encoded CA certificate file.
  - `certFile` — path to the PEM-encoded client certificate (for mutual TLS).
  - `keyFile` — path to the PEM-encoded client private key (for mutual TLS).
  - `insecureSkipVerify` — disables server certificate and hostname verification.
  - `serverName` — overrides the hostname used for TLS SNI and certificate verification.

> For a `KubeDB`-managed database, an `AppBinding` is created by default, so you usually only need to
> create one for the source. Learn more about [AppBinding](/docs/guides/mongodb/concepts/appbinding/).

### spec.source.mongodb.mongoshake

`mongoshake` configures how the migration is performed. All fields are optional unless noted.

- `syncMode` — controls the synchronization mode. One of:
  - `all` — full synchronization followed by incremental (oplog) synchronization.
  - `full` — full synchronization only.
  - `incr` — incremental synchronization only.
- `filterOpTypes` — oplog operation types to include, for example `i` (insert), `u` (update),
  `d` (delete).
- `filterNamespaceBlack` — namespaces (`db.collection` or `db`) to exclude. Multiple entries are
  separated by `;`, for example `db1.col1;db2`.
- `filterNamespaceWhite` — namespaces (`db.collection` or `db`) to include exclusively. When set,
  only the listed namespaces are migrated.
- `filterPassSpecialDb` — special system databases to include, for example `admin;local;config`.
  Collection-level filtering is not supported here.
- `filterDdlEnable` — controls whether DDL operations are included. When disabled, only oplog
  operations (`i`/`u`/`d`) are synced; when enabled, DDL operations such as create index or drop
  database are also included.
- `filterOplogGids` — enables filtering of the oplog by GID.
- `checkpointStartPosition` — initial oplog position as a UTC timestamp. Used only when no checkpoint
  exists.
- `transformNamespace` — maps source namespaces to destination namespaces
  (`fromDb.fromCollection:toDb.toCollection` or `fromDb:toDb`). Multiple mappings are separated by
  `;`, for example `db1.col1:db2.col1;db3:db4`.
- `extraConfiguration` — additional raw `mongoshake` configuration as key-value pairs passed directly
  without schema validation. For example, `full_sync.executor.insert_on_dup_update: "true"` uses
  upsert instead of insert during full sync to handle duplicate key errors gracefully.

### spec.jobDefaults

`spec.jobDefaults` is an optional field that sets default settings for the migration Job.

- `imagePullPolicy` — the image pull policy for the Migrator Job. Defaults to `IfNotPresent`.
- `backoffLimit` — the number of retries before the Job is marked as failed. Defaults to `6`.
- `ttlSecondsAfterFinished` — the TTL (in seconds) for cleaning up a completed Job.
- `activeDeadlineSeconds` — the duration (in seconds) relative to its start time that the Job may be
  active before the system tries to terminate it.

### spec.jobTemplate

`spec.jobTemplate` is an optional field that holds runtime configuration for the migration Job pod
(a `PodTemplateSpec`). Use it to set pod-level settings such as `securityContext`, `nodeSelector`,
`resources`, `serviceAccountName`, and so on.

## Migrator Status

`status` reflects the observed state of the migration.

- `status.phase` — the current phase of the migration. One of:
  - `Pending` — the migration has not started yet.
  - `Running` — the migration is in progress.
  - `Succeeded` — the migration completed successfully.
  - `Failed` — the migration failed.
- `status.progress` — the current progress of the migration:
  - `dbType` — the type of database being migrated.
  - `info` — additional progress information, including the current `Stage`, `Lag`, and `Progress`
    (these are surfaced as columns in `kubectl get migrator`).
- `status.conditions` — an array of conditions describing the migration's state over time (for
  example, `MigratorJobTriggered`, `MigrationRunning`, `MigrationSucceeded`, `MigrationFailed`).

## Next Steps

- Migrate a MongoDB database step by step with the [MongoDB Database Migration](/docs/guides/mongodb/migration/databaseMigration.md) guide.
- Learn about the [AppBinding](/docs/guides/mongodb/concepts/appbinding/) concept.
- Learn about the MongoDB CRD [here](/docs/guides/mongodb/concepts/mongodb).
