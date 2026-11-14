---
title: DurableWorkflowVersion CRD
menu:
  docs_{{ .version }}:
    identifier: dw-catalog-concepts
    name: DurableWorkflowVersion
    parent: dw-concepts-durableworkflow
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DurableWorkflowVersion

## What is DurableWorkflowVersion

`DurableWorkflowVersion` is a Kubernetes `Custom Resource Definition` (CRD). It specifies the engine, the container images, the ports, the schema tooling and the rollout settings that KubeDB uses for a [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md). A new upstream release of Temporal or Cadence needs, in most cases, only a new `DurableWorkflowVersion`.

`DurableWorkflowVersion` is a cluster scoped object. The KubeDB catalog installs these objects, and the short name is `dwversion`.

## DurableWorkflowVersion Spec

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DurableWorkflowVersion
metadata:
  name: temporal-1.32.1
spec:
  version: 1.32.1
  distribution: Temporal
  defaultHistoryShards: 16
  server:
    image: docker.io/temporalio/server:1.32.1
  adminTools:
    image: docker.io/temporalio/admin-tools:1.32.1
  ui:
    image: docker.io/temporalio/ui:2.54.1
  ports:
    frontend: {grpc: 7233, membership: 6933, http: 7243}
    internalFrontend: {grpc: 7236, membership: 6936, http: 7246}
    history: {grpc: 7234, membership: 6934}
    matching: {grpc: 7235, membership: 6935}
    worker: {grpc: 7239, membership: 6939}
    metrics: 9090
    ui: 8080
  schema:
    tool: temporal-tools
    dir: /etc/temporal/schema
    stores:
    - {type: Postgres, default: true, visibility: true}
    - {type: MySQL, default: true, visibility: true}
    - {type: Cassandra, default: true}
    - {type: Elasticsearch, visibility: true}
    - {type: OpenSearch, visibility: true}
  securityContext:
    runAsUser: 1000
  rollout:
    roleOrder: [history, matching, internalFrontend, frontend, worker]
    frontendShutdownFailHealthCheckDuration: 10s
    frontendShutdownDrainDuration: 60s
    matchingShutdownDrainDuration: 10s
    historyShutdownDrainDuration: 30s
    preStopSleep: 10s
    historyMinInterval: 60s
    postHopSettle: 10m
    onlineSchemaUpdate: true
  updateConstraints:
    allowlist:
    - "> 1.32.1, < 1.33.0"
```

### spec.version and spec.distribution

`spec.version` is the engine version, for example `1.32.1`. `spec.distribution` is `Temporal` or `Cadence`. A `DurableWorkflow` takes its engine from here. The name of the object is what `DurableWorkflow.spec.version` refers to.

### spec.server, spec.adminTools and spec.ui

The images. `server` runs every role. `adminTools` runs the schema Job and the admin tools pod, and for Cadence it is the server image, because `cadence-server update-schema` and the Cadence CLI are in it. `ui` is the web UI image (Temporal UI or cadence-web).

### spec.ports

The ports of every role, see [Roles and High Availability](/docs/guides/durableworkflow/clustering/roles-and-ha.md#ports).

### spec.schema

How the schema is applied. `tool` is `temporal-tools` or `cadence-server`. `stores` lists the engines that the version supports, and whether each can be the default store, the visibility store, or both. The webhook rejects a store that is not in the list.

### spec.rollout

What zero-downtime rolling needs: the order of the roles, the drain durations that the servers need to leave their rings, the sleep before SIGTERM, the minimum time between two history pods, and whether the schema update can run while the previous version is serving (`onlineSchemaUpdate`, default `true`). See [Update Version](/docs/guides/durableworkflow/update-version/overview.md).

### spec.updateConstraints

An `allowlist` of the versions that this one can be updated to. See [Update Version](/docs/guides/durableworkflow/update-version/overview.md).

### spec.publicClientTLS

`true` only for a server build that applies TLS to the public client of the Cadence worker service. Upstream Cadence 1.4.2 has no such field, and without it the worker cannot reach a TLS frontend. The build is an AppsCode patch. **The webhook rejects TLS on a Cadence version that does not set it.** Temporal does not need it. See [TLS](/docs/guides/durableworkflow/tls/tls.md#cadence).

### spec.defaultHistoryShards, spec.deprecated, spec.endOfLife

`defaultHistoryShards` is used when a `DurableWorkflow` omits `numHistoryShards` (`16`). `deprecated` and `endOfLife` mark versions that should not be used.

## Next Steps

- Learn about the [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md) CRD.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
