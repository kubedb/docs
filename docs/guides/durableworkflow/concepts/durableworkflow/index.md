---
title: DurableWorkflow CRD
menu:
  docs_{{ .version }}:
    identifier: dw-durableworkflow-concepts
    name: DurableWorkflow
    parent: dw-concepts-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DurableWorkflow

## What is DurableWorkflow

`DurableWorkflow` is a Kubernetes `Custom Resource Definition` (CRD). It provides declarative configuration for a [Temporal](https://temporal.io) or [Cadence](https://cadenceworkflow.io) cluster in a Kubernetes native way. You only need to describe the desired cluster in a `DurableWorkflow` object, and the KubeDB operator will create the Kubernetes objects in the desired state for you. The engine is not a field of its own: it is the distribution of the [DurableWorkflowVersion](/docs/guides/durableworkflow/concepts/catalog/index.md) named in `spec.version`.

## DurableWorkflow Spec

A `DurableWorkflow` object has the following fields in the `spec` section.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal
  namespace: demo
spec:
  version: temporal-1.32.1
  numHistoryShards: 16
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
  ui:
    enabled: true
  adminTools:
    enabled: true
  monitor:
    agent: prometheus.io/builtin
  deletionPolicy: WipeOut
```

### spec.version

`spec.version` is a required field that names a `DurableWorkflowVersion`, for example `temporal-1.32.1` or `cadence-1.4.2`. You cannot change it by editing the object after it is `Ready`. Use an [UpdateVersion](/docs/guides/durableworkflow/update-version/overview.md) ops request. A change of the engine is not an update and is rejected.

### spec.numHistoryShards

The number of history shards. If you omit it, the default of the `DurableWorkflowVersion` is used, which is `16`. It is immutable.

### spec.topology

`spec.topology` has one entry for each role: `frontend`, `history`, `matching`, `worker` and `internalFrontend` (Temporal only, rejected for Cadence). Every entry has:

- `replicas`: the number of pods. The default is `1`, the minimum is `1`.
- `podTemplate`: the usual KubeDB pod template, for resources, scheduling, security context and so on. The container is named `durableworkflow`.

### spec.persistence

`spec.persistence` references the stores that hold all state.

- `default` is required. It is the store for workflow state: Postgres, MySQL or Cassandra.
- `visibility` is optional. When omitted, the server of the default store is used with a separate database. Postgres and MySQL work for both engines. Elasticsearch and OpenSearch work for Temporal only. Temporal does not accept Cassandra for visibility.
- `kafkaRef` is for Cadence with a search visibility store. That combination is **not supported yet**, and the field is rejected for Temporal. In practice, do not set it.

Every store has these fields:

| Field            | Description                                                                                                                         |
|------------------|-------------------------------------------------------------------------------------------------------------------------------------|
| `name`, `namespace` | The `AppBinding` of a KubeDB database, or of an external one. Required.                                                           |
| `type`           | `Postgres`, `MySQL`, `Cassandra`, `Elasticsearch` or `OpenSearch`. Required only if the `AppBinding` does not identify the engine. |
| `databaseName`   | The database, keyspace or index name. Defaults to `<name>` for the default store and `<name>_visibility` for visibility.             |
| `createDatabase` | Defaults to `true`: the operator creates the database and a dedicated user with the admin credentials of the `AppBinding`. If `false`, both must exist and `authSecret` is required. |
| `authSecret`     | The credentials the servers connect with. The operator generates the Secret when `createDatabase` is true. `rotateAfter` is read by the [recommendation engine](/docs/guides/durableworkflow/recommendation/recommendation.md). |
| `cassandra`      | `replicationFactor` (default `1`) and `datacenter`, used when the operator creates a keyspace.                                       |

See [Persistence](/docs/guides/durableworkflow/configuration/persistence.md).

### spec.configuration

Custom configuration, given in a Secret (`secretName`) or inline (`inline`). Two keys are used: `server.yaml` (merged over the static configuration, rolls the pods) and `dynamic_config.yaml` (applied live). See [Custom Configuration](/docs/guides/durableworkflow/configuration/custom-config.md).

### spec.tls and spec.frontendClientAuth

`spec.tls.issuerRef` is a cert-manager issuer. The certificates have the aliases `internode`, `frontend` and `client`. `spec.frontendClientAuth` is `Required` or `NotRequired`. It defaults to `Required` for Temporal and `NotRequired` for Cadence, and a Cadence cluster cannot use `Required`. See [TLS](/docs/guides/durableworkflow/tls/tls.md).

### spec.authorization

JWT authorization: `jwksURL` or `publicKeySecret`, plus claim settings that differ by engine. See [Authorization](/docs/guides/durableworkflow/authorization/authorization.md).

### spec.ui and spec.adminTools

`spec.ui` has `enabled`, `replicas`, `podTemplate` and, for Cadence only, `auth` (`Disabled` or `JWT`). `spec.adminTools` has `enabled` and `podTemplate`. See [UI and Admin Tools](/docs/guides/durableworkflow/configuration/ui-and-admin-tools.md).

### spec.serviceTemplates

Optional templates for the Services. The `primary` alias is the frontend Service and `stats` is the metrics Service.

### spec.monitor

Monitoring with Prometheus, see [Monitoring](/docs/guides/durableworkflow/monitoring/overview.md).

### spec.deletionPolicy

| Value            | Servers                                | Databases                                                                                      |
|------------------|----------------------------------------|------------------------------------------------------------------------------------------------|
| `DoNotTerminate` | Deletion is blocked                    | Kept                                                                                           |
| `Halt`           | Removed, the credential Secrets stay   | Kept, with the user                                                                            |
| `Delete` (default) | Removed                              | Kept, with the user                                                                            |
| `WipeOut`        | Removed                                | The databases, keyspaces, indices and users that the operator created are dropped. A database the operator did not create (`createDatabase: false`) and the database server itself are never touched. |

### spec.halted

`spec.halted: true` removes all server pods and keeps the configuration, the Secrets and the databases. Set it back to `false` to start again.

### spec.healthChecker

Attributes of the KubeDB health checker. The defaults are `periodSeconds: 10`, `timeoutSeconds: 10` and `failureThreshold: 3`.

## DurableWorkflow Status

`status.phase` is `Provisioning`, `Ready`, `NotReady` and so on, like for the other databases. `status.distribution` is the resolved engine. `status.conditions` holds the conditions of the cluster, like `Ready`. `status.persistence.default` and `status.persistence.visibility` report whether the database was created (`created`), its `schemaVersion` and the server version whose schema was applied (`schemaServerVersion`).

## Next Steps

- Learn about the [DurableWorkflowVersion](/docs/guides/durableworkflow/concepts/catalog/index.md) and [DurableWorkflowOpsRequest](/docs/guides/durableworkflow/concepts/durableworkflow-ops-request/index.md) CRDs.
- Run a cluster with the [quickstart](/docs/guides/durableworkflow/quickstart/prerequisites.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
