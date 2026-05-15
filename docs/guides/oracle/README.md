---
title: Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-readme
    name: Oracle
    parent: guides-oracle
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/oracle/
aliases:
  - /docs/{{ .version }}/guides/oracle/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Oracle Features

| Features                                          | Availability |
|---------------------------------------------------|:------------:|
| Standalone Oracle Deployment                      |   &#10003;   |
| Oracle Data Guard (Primary/Standby Topology)     |   &#10003;   |
| Synchronous and Asynchronous Replication          |   &#10003;   |
| Fast-Start Failover (FSFO)                        |   &#10003;   |
| Automatic Failover Workflow                       |   &#10003;   |
| Persistent Storage with PVC                       |   &#10003;   |
| Pod-Level Resource and Security Customization     |   &#10003;   |
| Custom Container Image via OracleVersion Catalog  |   &#10003;   |

## Life Cycle of an Oracle Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/oracle/oracle_lifecycle.png">
</p>

## User Guide

- [Quickstart Oracle](/docs/guides/oracle/quickstart/guide.md) with KubeDB Operator.
- Learn Oracle deployment and behavior through the [Oracle CRD Concepts](/docs/guides/oracle/concepts/oracle.md) guide.
- Explore high availability and automatic recovery using [Failover and Disaster Recovery](/docs/guides/oracle/failover/overview.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## A Guide to Oracle Operations in KubeDB

The `Oracle` custom resource helps you manage day-2 operations for Oracle databases on Kubernetes through a declarative API. You can define your target topology, resource profile, and failover behavior, and KubeDB continuously reconciles runtime state with your desired state.

### Choose the Right Deployment Mode

KubeDB Oracle supports two deployment modes:

- **Standalone**: A single-instance Oracle deployment, typically used for development, test, and low-complexity workloads.
- **DataGuard**: A multi-node deployment with primary/standby roles for high availability and disaster recovery.

The [Quickstart guide](/docs/guides/oracle/quickstart/guide.md) walks through creating an Oracle instance and validating the generated resources.

### Understand Oracle CRD Structure

The [Oracle Concepts guide](/docs/guides/oracle/concepts/oracle.md) explains important fields in the `Oracle` spec, including:

- Base settings like `version`, `mode`, `edition`, and `replicas`.
- Data Guard settings such as `protectionMode`, `syncMode`, and failover observer configuration.
- Storage and pod template customization for production-ready runtime behavior.
- Lifecycle controls, including `deletionPolicy`.

Use this guide as the authoritative reference when designing manifests for either standalone or Data Guard deployments.

### Plan for High Availability and Failover

For production Oracle clusters, Data Guard is the key reliability building block. The [failover overview](/docs/guides/oracle/failover/overview.md) details:

- How primary and standby members are coordinated.
- How redo transport and apply flows affect recovery characteristics.
- How Fast-Start Failover (FSFO) and observer behavior impact automatic failover.
- Practical failure simulation scenarios and expected cluster behavior.

This is the best starting point for validating RTO/RPO expectations before promoting Oracle workloads to production.

### Recommended Workflow

For new users and platform teams, follow this sequence:

1. Deploy a baseline instance from [Quickstart](/docs/guides/oracle/quickstart/guide.md).
2. Review CRD-level controls in [Oracle Concepts](/docs/guides/oracle/concepts/oracle.md).
3. Enable and test recovery paths using [Failover and Disaster Recovery](/docs/guides/oracle/failover/overview.md).

By following this progression, you can move from initial deployment to production-grade high availability with a clear operational model.


