---
title: Weaviate Ops Request Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-ops-request-overview
    name: Overview
    parent: weaviate-ops-request
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Day-2 Operations

This guide provides an overview of the day-2 operational workflows that KubeDB supports for `Weaviate` databases via the `WeaviateOpsRequest` CRD.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)

## Supported Operations

KubeDB supports the following day-2 operations for Weaviate:

| Operation | Description |
|-----------|-------------|
| [UpdateVersion](/docs/guides/weaviate/update-version/overview.md) | Update the version of a running Weaviate database |
| [VerticalScaling](/docs/guides/weaviate/scaling/vertical-scaling/overview.md) | Update CPU and memory resources of Weaviate nodes |
| [VolumeExpansion](/docs/guides/weaviate/volume-expansion/overview.md) | Expand the persistent volume claim size of Weaviate nodes |
| [Reconfigure](/docs/guides/weaviate/reconfigure/overview.md) | Reconfigure a running Weaviate database with new configuration |
| [Restart](/docs/guides/weaviate/restart/restart.md) | Restart the Weaviate database pods in a rolling fashion |
| [RotateAuth](/docs/guides/weaviate/rotate-auth/overview.md) | Rotate the API key credentials of a Weaviate database |

## How Ops Requests Work

All day-2 operations for Weaviate are performed through the `WeaviateOpsRequest` CRD. The general workflow is:

1. The user creates a `WeaviateOpsRequest` CR with the desired operation type and parameters.
2. `KubeDB-ops-manager` operator watches for `WeaviateOpsRequest` CRs.
3. When it finds one, it pauses the `Weaviate` object to prevent conflicting operations.
4. The operator performs the requested operation (e.g., updates images, scales nodes, expands volumes).
5. After the operation completes successfully, the operator updates the `Weaviate` object and resumes it.
6. The `WeaviateOpsRequest` status transitions to `Successful`.

> **Note:** Only one `WeaviateOpsRequest` should be active at a time for a given `Weaviate` database. Wait for one operation to complete before starting another.
