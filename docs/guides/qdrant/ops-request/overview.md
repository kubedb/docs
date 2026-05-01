---
title: Qdrant Ops Request Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-ops-request-overview
    name: Overview
    parent: qdrant-ops-request
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Day-2 Operations

This guide provides an overview of the day-2 operational workflows that KubeDB supports for `Qdrant` databases via the `QdrantOpsRequest` CRD.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## Supported Operations

KubeDB supports the following day-2 operations for Qdrant:

| Operation | Description |
|-----------|-------------|
| [UpdateVersion](/docs/guides/qdrant/update-version/overview.md) | Update the version of a running Qdrant database |
| [HorizontalScaling](/docs/guides/qdrant/scaling/horizontal-scaling/overview.md) | Scale the number of Qdrant nodes up or down |
| [VerticalScaling](/docs/guides/qdrant/scaling/vertical-scaling/overview.md) | Update CPU and memory resources of Qdrant nodes |
| [VolumeExpansion](/docs/guides/qdrant/volume-expansion/overview.md) | Expand the persistent volume claim size of Qdrant nodes |
| [Reconfigure](/docs/guides/qdrant/reconfigure/overview.md) | Reconfigure a running Qdrant database with new configuration |
| [ReconfigureTLS](/docs/guides/qdrant/reconfigure-tls/overview.md) | Add, rotate, or remove TLS certificates for Qdrant |
| [Restart](/docs/guides/qdrant/restart/restart.md) | Restart the Qdrant database pods in a rolling fashion |
| [RotateAuth](/docs/guides/qdrant/rotate-auth/overview.md) | Rotate the authentication credentials of a Qdrant database |

## How Ops Requests Work

All day-2 operations for Qdrant are performed through the `QdrantOpsRequest` CRD. The general workflow is:

1. The user creates a `QdrantOpsRequest` CR with the desired operation type and parameters.
2. `KubeDB-ops-manager` operator watches for `QdrantOpsRequest` CRs.
3. When it finds one, it pauses the `Qdrant` object to prevent conflicting operations.
4. The operator performs the requested operation (e.g., updates images, scales nodes, expands volumes).
5. After the operation completes successfully, the operator updates the `Qdrant` object and resumes it.
6. The `QdrantOpsRequest` status transitions to `Successful`.

> **Note:** Only one `QdrantOpsRequest` should be active at a time for a given `Qdrant` database. Wait for one operation to complete before starting another.
