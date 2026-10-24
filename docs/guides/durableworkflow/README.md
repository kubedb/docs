---
title: DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-readme-durableworkflow
    name: DurableWorkflow
    parent: dw-durableworkflow-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/durableworkflow/
aliases:
  - /docs/{{ .version }}/guides/durableworkflow/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## DurableWorkflow Features

`DurableWorkflow` is one KubeDB kind for two durable workflow engines, the way `Elasticsearch` covers Elasticsearch and OpenSearch. The `DurableWorkflowVersion` named in `spec.version` selects the engine:

- **Temporal**, version `temporal-1.32.1`
- **Cadence**, version `cadence-1.4.2`

The server roles (frontend, history, matching, worker, and the Temporal-only internal frontend) are stateless. They keep no volumes. All workflow state lives in a database that you point to with an `AppBinding`.

| Features                                                                                           | Availability |
|----------------------------------------------------------------------------------------------------|:------------:|
| Temporal and Cadence from one kind                                                                 |   &#10003;   |
| Stateless roles: frontend, history, matching, worker, internalFrontend (Temporal only)             |   &#10003;   |
| Default store: KubeDB Postgres, MySQL, Cassandra (through an AppBinding)                           |   &#10003;   |
| Visibility store: Postgres, MySQL; Elasticsearch / OpenSearch for Temporal only                    |   &#10003;   |
| Operator-created databases and a dedicated database user per store                                 |   &#10003;   |
| Custom configuration: static `server.yaml` and live `dynamic_config.yaml`                          |   &#10003;   |
| TLS: `internode`, `frontend` and `client` certificates ( [Cert Manager](https://cert-manager.io/docs/) ) |   &#10003;   |
| JWT authorization                                                                                  |   &#10003;   |
| Web UI and admin tools pod                                                                         |   &#10003;   |
| Zero-downtime Restart, UpdateVersion, Reconfigure, scaling, RotateAuth on a highly available cluster |   &#10003;   |
| Reconfigure TLS: rotate, update certificates, enable, remove, change issuer                        |   &#10003;   |
| Compute autoscaling per role                                                                       |   &#10003;   |
| Recommendations: rotate TLS, rotate auth, version update                                           |   &#10003;   |
| Monitoring with Prometheus                                                                         |   &#10003;   |

## Verification status

Be aware of what has been exercised on a real cluster.

- **Verified** on a kind cluster, with a load generator that started and signaled a workflow every 0.3 seconds against a highly available cluster (2 replicas per role) on KubeDB Postgres, with **0 failed calls** during: Restart, UpdateVersion (Temporal and Cadence), HorizontalScaling up and down, VerticalScaling, dynamic and static Reconfigure, certificate rotation (Temporal and Cadence with TLS) and RotateAuth (Temporal).
- **Verified**: Temporal and Cadence both also run on KubeDB MySQL. JWT authorization works for both engines: a call without a token and a call with a reader token are denied, and a call with an admin token is allowed.
- **Implemented, not yet verified on a cluster**: Cassandra as a store, Elasticsearch and OpenSearch as the Temporal visibility store, the autoscaler and the recommendations. The pages that cover them say so.

## Supported Versions

KubeDB supports the following `DurableWorkflowVersion` objects:

| Name              | Engine   | Server image                     |
|-------------------|----------|----------------------------------|
| `temporal-1.32.1` | Temporal | `temporalio/server:1.32.1`       |
| `cadence-1.4.2`   | Cadence  | `ubercadence/server:v1.4.2`      |

## Limitations

The following are **not** implemented. There are no guides for them.

- GitOps.
- Backup and restore. All state is in the databases, so back up the databases with their own KubeDB backup. See [Persistence](/docs/guides/durableworkflow/configuration/persistence.md).
- Multi-cluster replication.
- Volume expansion and storage migration. The servers have no PVCs. Run them on the database.
- Managed mode. The database must already exist and be reachable through an `AppBinding`. KubeDB does not create it for you.
- Cadence with an Elasticsearch or OpenSearch visibility store. It needs Kafka and is rejected by the webhook.
- The operator does not register a Temporal namespace or a Cadence domain. Create one after the cluster is `Ready`, see the [quickstart](/docs/guides/durableworkflow/quickstart/temporal.md).

## User Guide

- [Quickstart](/docs/guides/durableworkflow/quickstart/prerequisites.md) with Temporal and Cadence.
- Learn about the [roles and high availability](/docs/guides/durableworkflow/clustering/roles-and-ha.md).
- Choose and configure the [persistence stores](/docs/guides/durableworkflow/configuration/persistence.md).
- [Custom configuration](/docs/guides/durableworkflow/configuration/custom-config.md), [UI and admin tools](/docs/guides/durableworkflow/configuration/ui-and-admin-tools.md).
- [TLS](/docs/guides/durableworkflow/tls/tls.md) and [JWT authorization](/docs/guides/durableworkflow/authorization/authorization.md).
- Day-2 operations: [Restart](/docs/guides/durableworkflow/restart/restart.md), [Reconfigure](/docs/guides/durableworkflow/reconfigure/overview.md), [Reconfigure TLS](/docs/guides/durableworkflow/reconfigure_tls/overview.md), [Rotate Authentication](/docs/guides/durableworkflow/rotateauth/overview.md), [Scaling](/docs/guides/durableworkflow/scaling/overview.md) and [Update Version](/docs/guides/durableworkflow/update-version/overview.md).
- [Autoscaler](/docs/guides/durableworkflow/autoscaler/autoscaler.md), [Monitoring](/docs/guides/durableworkflow/monitoring/overview.md) and [Recommendation](/docs/guides/durableworkflow/recommendation/recommendation.md).
- Detailed [concepts](/docs/guides/durableworkflow/concepts/durableworkflow/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
