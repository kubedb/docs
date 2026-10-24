---
title: Persistence
menu:
  docs_{{ .version }}:
    identifier: dw-persistence-configuration
    name: Persistence
    parent: dw-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Persistence

A `DurableWorkflow` has no volumes. Every workflow, timer and task queue lives in the databases that `spec.persistence` refers to, so the availability and the backup of your data are the availability and the backup of those databases.

## Which database for which store

| Engine                         | Default store    | Visibility store | Notes                                                         |
|--------------------------------|:----------------:|:----------------:|---------------------------------------------------------------|
| KubeDB Postgres (or external)  | Temporal, Cadence | Temporal, Cadence | Verified, used in all guides                                  |
| KubeDB MySQL (or external)     | Temporal, Cadence | Temporal, Cadence | Verified                                                      |
| KubeDB Cassandra               | Temporal, Cadence | Cadence only      | Temporal does not accept Cassandra for visibility. Not yet verified on a cluster. |
| KubeDB Elasticsearch / OpenSearch | no             | Temporal only     | Not yet verified on a cluster. Cadence needs Kafka for it and is not supported yet. |

The `DurableWorkflowVersion` lists the supported engines, and the webhook rejects the others.

## What the operator creates

With the default `createDatabase: true`, the operator uses the admin credentials in the `AppBinding` to create:

| Store                                | Database or keyspace | User               | Secret                       |
|--------------------------------------|----------------------|--------------------|------------------------------|
| default                              | `<name>`             | `dw_<name>`        | `<name>-default-auth`        |
| visibility (not set, same server)    | `<name>_visibility`  | `dw_<name>` (shared) | `<name>-default-auth`      |
| visibility (set to another store)    | `<name>_visibility`  | `dw_<name>_visibility` | `<name>-visibility-auth` |

Dashes and dots in the name become underscores in user names and in Cassandra keyspaces. `spec.persistence.<store>.databaseName` overrides the database name. The servers connect with this dedicated user and do not use the admin credentials. The schema is applied by a Job before any server pod is created.

For Elasticsearch and OpenSearch the operator creates the index (from `databaseName`, default `<name>_visibility`), but it does not create a user. The servers use the account of the `AppBinding`.

## Use a database you prepared

If your DBA creates the databases and the user, set `createDatabase: false` and give the credentials in `authSecret`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: temporal-existing-auth
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: temporal_app
  password: change-me
---
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-existing
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
      databaseName: temporal_existing
      createDatabase: false
      authSecret:
        kind: Secret
        name: temporal-existing-auth
    visibility:
      name: pg
      namespace: demo
      databaseName: temporal_existing_visibility
      createDatabase: false
      authSecret:
        kind: Secret
        name: temporal-existing-auth
  deletionPolicy: Delete
```

The `AppBinding` is still needed, for the address of the server and for its admin Secret. With `createDatabase: false`, the operator does not drop anything under `WipeOut`, and [RotateAuth](/docs/guides/durableworkflow/rotateauth/overview.md) has nothing to rotate, because the user is yours.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/persistence/existing-database.yaml
```

## MySQL

Refer to the `AppBinding` of the KubeDB `MySQL`. Nothing else changes:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/persistence/mysql.yaml
```

## Cassandra

{{< notice type="warning" message="Cassandra as a store is implemented but has not been verified on a cluster yet." >}}

`cassandra.replicationFactor` (default `1`) and `cassandra.datacenter` are used when the operator creates the keyspace. This example keeps the workflow state in Cassandra and the visibility records in Postgres:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/persistence/temporal-cassandra.yaml
```

## Elasticsearch and OpenSearch visibility (Temporal)

{{< notice type="warning" message="An Elasticsearch or OpenSearch visibility store is implemented but has not been verified on a cluster yet." >}}

Name the `AppBinding` of a KubeDB `Elasticsearch` or `OpenSearch` in `visibility`, and set `type` if the `AppBinding` does not identify the engine:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-es
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
    visibility:
      name: es
      namespace: demo
      type: Elasticsearch
      databaseName: temporal_es_visibility
  deletionPolicy: WipeOut
```

Cadence with Elasticsearch or OpenSearch visibility is **not supported**. The webhook rejects it, as it needs Kafka and a visibility consumer that the operator does not set up.

## Deletion and backup

`deletionPolicy: WipeOut` drops the databases, keyspaces, indices and users that the operator created. `Delete` and `Halt` keep them. The database server is never deleted by the operator.

KubeDB does not back up a `DurableWorkflow`. To protect the data, back up the default store and the visibility store with the backup of their own KubeDB database. They are not snapshotted together. Restore into a new `DurableWorkflow` of the same engine, the same version line and the same `numHistoryShards`.

## Next Steps

- [Custom configuration](/docs/guides/durableworkflow/configuration/custom-config.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
