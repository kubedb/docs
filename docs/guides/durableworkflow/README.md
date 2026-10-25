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
| Visibility store: Postgres, MySQL; Elasticsearch / OpenSearch for Temporal, and for Cadence with Kafka |   &#10003;   |
| Operator-created databases and a dedicated database user per store                                 |   &#10003;   |
| Managed persistence: the operator creates the Postgres or MySQL itself (`spec.persistence.managed`) |   &#10003;   |
| `WorkflowNamespace`: Temporal namespaces and Cadence domains as Kubernetes objects                 |   &#10003;   |
| Archival of closed workflows to S3, GCS or a filestore                                             |   &#10003;   |
| Per-index application user for Elasticsearch / OpenSearch                                          |   &#10003;   |
| Custom configuration: static `server.yaml` and live `dynamic_config.yaml`                          |   &#10003;   |
| TLS: `internode`, `frontend` and `client` certificates ( [Cert Manager](https://cert-manager.io/docs/) ) |   &#10003;   |
| JWT authorization                                                                                  |   &#10003;   |
| Web UI and admin tools pod                                                                         |   &#10003;   |
| Zero-downtime Restart, UpdateVersion, Reconfigure, scaling, RotateAuth on a highly available cluster |   &#10003;   |
| Reconfigure TLS: rotate, update certificates, enable, remove, change issuer                        |   &#10003;   |
| Compute autoscaling per role                                                                       |   &#10003;   |
| Recommendations: rotate TLS, rotate auth, version update                                           |   &#10003;   |
| Monitoring with Prometheus                                                                         |   &#10003;   |
| GitOps: UpdateVersion, HorizontalScaling, VerticalScaling, Reconfigure, ReconfigureTLS              |   &#10003;   |
| Backup and restore with KubeStash (PostgreSQL stores verified)                                       |   &#10003;   |

## Verification status

Be aware of what has been exercised on a real cluster.

- **Verified** on a kind cluster, with a load generator that started and signaled a workflow every 0.3 seconds against a highly available cluster (2 replicas per role) on KubeDB Postgres, with **0 failed calls** during: Restart, UpdateVersion (Temporal and Cadence), HorizontalScaling up and down, VerticalScaling, dynamic and static Reconfigure, certificate rotation (Temporal and Cadence with TLS) and RotateAuth (Temporal).
- **Verified**: Temporal and Cadence both also run on KubeDB MySQL. JWT authorization works for both engines: a call without a token and a call with a reader token are denied, and a call with an admin token is allowed.
- **Verified** on a kind cluster: managed persistence (Temporal on a managed Postgres, Cadence on a managed MySQL, both `Ready`), and `WorkflowNamespace` (create, retention update, delete, archival enabled on a namespace, and a Cadence domain with the retention rounded up to whole days).
- **Implemented, not yet verified on a cluster**: backup and restore of MySQL and Cassandra stores, Elasticsearch and OpenSearch as the visibility store (with Kafka for Cadence, and the per-index user), archival to a real bucket, GitOps (it depends on kubedb/gitops#97), the autoscaler and the recommendations. The pages that cover them say so.

## Supported Versions

KubeDB supports the following `DurableWorkflowVersion` objects:

| Name              | Engine   | Server image                     |
|-------------------|----------|----------------------------------|
| `temporal-1.32.1` | Temporal | `temporalio/server:1.32.1`       |
| `cadence-1.4.2`   | Cadence  | `ubercadence/server:v1.4.2`      |

## Limitations

The following are **not** implemented. There are no guides for them.

- Multi-cluster replication.
- Volume expansion and storage migration. The servers have no PVCs. Run them on the database.
- Grafana dashboards.

Also note: a Cadence server (1.4.2) cannot delete a domain, so deleting a `WorkflowNamespace` leaves the domain deprecated, see [WorkflowNamespace](/docs/guides/durableworkflow/concepts/workflownamespace/index.md).

## User Guide

- [Quickstart](/docs/guides/durableworkflow/quickstart/prerequisites.md) with Temporal and Cadence.
- Learn about the [roles and high availability](/docs/guides/durableworkflow/clustering/roles-and-ha.md).
- Choose and configure the [persistence stores](/docs/guides/durableworkflow/configuration/persistence.md), or let the operator create the database with [managed persistence](/docs/guides/durableworkflow/configuration/managed-persistence.md).
- Create [namespaces and domains](/docs/guides/durableworkflow/namespaces/workflownamespace.md) with `WorkflowNamespace`.
- Configure [archival](/docs/guides/durableworkflow/configuration/archival.md) and [advanced visibility](/docs/guides/durableworkflow/configuration/advanced-visibility.md) (Elasticsearch / OpenSearch, Kafka).
- [Custom configuration](/docs/guides/durableworkflow/configuration/custom-config.md), [UI and admin tools](/docs/guides/durableworkflow/configuration/ui-and-admin-tools.md).
- [TLS](/docs/guides/durableworkflow/tls/tls.md) and [JWT authorization](/docs/guides/durableworkflow/authorization/authorization.md).
- Day-2 operations: [Restart](/docs/guides/durableworkflow/restart/restart.md), [Reconfigure](/docs/guides/durableworkflow/reconfigure/overview.md), [Reconfigure TLS](/docs/guides/durableworkflow/reconfigure_tls/overview.md), [Rotate Authentication](/docs/guides/durableworkflow/rotateauth/overview.md), [Scaling](/docs/guides/durableworkflow/scaling/overview.md) and [Update Version](/docs/guides/durableworkflow/update-version/overview.md).
- [Backup and restore](/docs/guides/durableworkflow/backup-restore/backup-restore.md), [GitOps](/docs/guides/durableworkflow/gitops/gitops.md), [Autoscaler](/docs/guides/durableworkflow/autoscaler/autoscaler.md), [Monitoring](/docs/guides/durableworkflow/monitoring/overview.md) and [Recommendation](/docs/guides/durableworkflow/recommendation/recommendation.md).
- Detailed [concepts](/docs/guides/durableworkflow/concepts/durableworkflow/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
