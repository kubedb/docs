---
title: Managed Persistence
menu:
  docs_{{ .version }}:
    identifier: dw-managed-persistence-configuration
    name: Managed Persistence
    parent: dw-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Managed Persistence

Instead of pointing a `DurableWorkflow` to a database that you created, you can let the operator create it. Set `spec.persistence.managed` and leave `spec.persistence.default` out. The operator then creates a KubeDB `Postgres` (the default) or `MySQL` named `<name>-db` in the namespace of the `DurableWorkflow`, and uses it for both the default and the visibility store.

{{< notice type="info" message="Verified on a kind cluster: Temporal on a managed Postgres (the child is named <name>-db) and Cadence on a managed MySQL both reached Ready." >}}

## Before You Begin

- Install KubeDB with the `DurableWorkflow` feature gate, and have a Postgres or MySQL catalog entry (`PostgresVersion` or `MySQLVersion`) installed. See [Prepare Dependencies](/docs/guides/durableworkflow/quickstart/prerequisites.md) for the namespace. You do not need the `pg` Postgres of the quickstart.
- You should be familiar with [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md) and [Persistence](/docs/guides/durableworkflow/configuration/persistence.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Temporal on a managed Postgres

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-managed
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
    managed:
      engine: Postgres
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 20Gi
  deletionPolicy: WipeOut
```

Here,

- `managed.engine` is `Postgres` (default) or `MySQL`.
- `managed.version` is the name of a `PostgresVersion` or `MySQLVersion`. If you leave it out, the operator picks the newest catalog entry that is official, not deprecated, not end of life and needs no license, and writes the choice to the spec on creation.
- `managed.replicas` is `1` (default) or `3` for a highly available database. Postgres forms a streaming replication cluster. MySQL forms a group replication cluster and needs `3` to `9` replicas, or `1` for a single server.
- `managed.storage` is the volume of each database pod. The size is required.
- `managed.podTemplate` is the optional pod template of the database pods.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/persistence/managed-postgres.yaml
durableworkflow.kubedb.com/temporal-managed created
```

### What happens

1. The operator creates the `Postgres` `temporal-managed-db`, with an owner reference to the `DurableWorkflow`, and sets its deletion policy to the one of the `DurableWorkflow`.
2. The `DurableWorkflow` stays in `Provisioning` while the database is not `Ready`. The condition `ManagedStoreReady` says why.
3. Once the database is `Ready` and KubeDB has published its `AppBinding` (same name, `temporal-managed-db`), the operator continues as it does with a reference: it creates the databases and the user, applies the schema, and creates the PetSets.

```bash
$ kubectl get dw,postgres -n demo
NAME                                         VERSION           STATUS         AGE
durableworkflow.kubedb.com/temporal-managed  temporal-1.32.1   Provisioning   40s

NAME                                      VERSION   STATUS   AGE
postgres.kubedb.com/temporal-managed-db   17.2      Ready    40s

```

Once the servers are up, `status.phase` is `Ready`.

## Cadence on a managed MySQL

Set `engine: MySQL`. Nothing else changes:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/persistence/managed-mysql.yaml
```

## Connection limit

The servers cap their connection pool at 5 connections per store per pod. The operator sets `max_connections` of the database when it creates it:

```
max_connections = 2 x (pods + one extra pod per role) x 2 stores x 5 + 25
```

It is never below the stock limit of the engine (`100` for Postgres, `151` for MySQL). The factor 2 leaves room to scale the roles out later. If you scale the roles out a lot, raise the limit with a Reconfigure ops request ([Postgres](/docs/guides/postgres/reconfigure/overview.md), [MySQL](/docs/guides/mysql/reconfigure/overview/)) on the database `<name>-db`; the operator warns when the budget is exceeded. See also [the connection budget](/docs/guides/durableworkflow/quickstart/prerequisites.md#the-connection-budget).

## Managing the database afterwards

- `spec.persistence.managed` is **immutable**. It cannot be added to an existing `DurableWorkflow` (that one already uses the databases it references), cannot be removed, and cannot be edited. The webhook rejects all three.
- After creation, the database belongs to you. Only its labels, owner and deletion policy are kept in sync. Change its version, replicas, resources, storage or configuration with ops requests on `<name>-db` (`PostgresOpsRequest` or `MySQLOpsRequest`: UpdateVersion, VerticalScaling, HorizontalScaling, VolumeExpansion, Reconfigure, ...).
- An `UpdateVersion` of the `DurableWorkflow` never touches the database.
- `spec.persistence.visibility` cannot be set, because the database holds the visibility store as well. `spec.persistence.default` cannot carry a name or `createDatabase: false`, and a `type` there must match the engine of the database.
- Only Postgres and MySQL can be managed. Cassandra, Elasticsearch and OpenSearch stay references.
- If the database disappears after the schema was applied, the operator does not create an empty one over the lost data. It reports an error instead.
- Once the workflow state exists, a database that is briefly not `Ready` (a failover or a restart) does not stop the reconciliation of the servers, so ops requests of the `DurableWorkflow` keep working.

## Deletion

The `deletionPolicy` of the `DurableWorkflow` is copied to the database and applied when the `DurableWorkflow` is deleted. The operator deletes the database before it releases the `DurableWorkflow`.

| `deletionPolicy`   | Volumes and secrets of the database                |
|--------------------|----------------------------------------------------|
| `DoNotTerminate`   | Deletion of the `DurableWorkflow` is blocked       |
| `Halt`             | Volumes and secrets are kept                       |
| `Delete` (default) | Secrets are kept, volumes are removed              |
| `WipeOut`          | Everything is removed                              |

## Cleaning up

```bash
kubectl delete dw -n demo temporal-managed
```

The child `Postgres` is deleted with it, under the policy above.

## Next Steps

- Use a [database you prepared](/docs/guides/durableworkflow/configuration/persistence.md) instead.
- Create a [WorkflowNamespace](/docs/guides/durableworkflow/namespaces/workflownamespace.md) in the cluster.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
