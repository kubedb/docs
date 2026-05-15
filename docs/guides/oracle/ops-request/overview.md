---
title: Oracle Ops Request Overview
menu:
  docs_{{ .version }}:
    identifier: oracle-ops-request-overview
    name: Overview
    parent: oracle-ops-request
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Oracle Day-2 Operations

This guide provides an overview of the day-2 operational workflows that KubeDB supports for `Oracle` databases via the `OracleOpsRequest` CRD.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## Supported Operations

KubeDB supports the following day-2 operations for Oracle:

| Operation | Description |
|-----------|-------------|
| [UpdateVersion](/docs/guides/oracle/update-version/overview.md) | Update the version of a running Oracle database |
| [HorizontalScaling](/docs/guides/oracle/scaling/horizontal-scaling/overview.md) | Scale the number of Oracle nodes up or down |
| [VerticalScaling](/docs/guides/oracle/scaling/vertical-scaling/overview.md) | Update CPU and memory resources of Oracle nodes |
| [VolumeExpansion](/docs/guides/oracle/volume-expansion/overview.md) | Expand the persistent volume claim size of Oracle nodes |
| [Reconfigure](/docs/guides/oracle/reconfigure/overview.md) | Reconfigure a running Oracle database with new configuration |
| [ReconfigureTLS](/docs/guides/oracle/reconfigure-tls/overview.md) | Add, rotate, or remove TLS certificates for Oracle |
| [Restart](/docs/guides/oracle/restart/restart.md) | Restart the Oracle database pods in a rolling fashion |
| [RotateAuth](/docs/guides/oracle/rotate-auth/overview.md) | Rotate the authentication credentials of a Oracle database |

## How Ops Requests Work

All day-2 operations for Oracle are performed through the `OracleOpsRequest` CRD. The general workflow is:

1. The user creates a `OracleOpsRequest` CR with the desired operation type and parameters.
2. `KubeDB-ops-manager` operator watches for `OracleOpsRequest` CRs.
3. When it finds one, it pauses the `Oracle` object to prevent conflicting operations.
4. The operator performs the requested operation (e.g., updates images, scales nodes, expands volumes).
5. After the operation completes successfully, the operator updates the `Oracle` object and resumes it.
6. The `OracleOpsRequest` status transitions to `Successful`.

> **Note:** Only one `OracleOpsRequest` should be active at a time for a given `Oracle` database. Wait for one operation to complete before starting another.
