---
title: Neo4j Ops Request Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-ops-request-overview
    name: Overview
    parent: neo4j-ops-request
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Ops Request

This page gives an overview of how KubeDB Ops-manager handles day-2 operations for Neo4j through `Neo4jOpsRequest`.

## Before You Begin

- Deploy Neo4j first using the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md).
- Be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How the Operator Processes `Neo4jOpsRequest`

The following diagram shows how KubeDB Ops-manager processes `Neo4jOpsRequest` for day-2 operations. Open the image in a new tab to see the enlarged version.

<figure>
  <img alt="Neo4j OpsRequest operational flow" src="/docs/images/neo4j/operational-view.png">
  <figcaption>Fig: Neo4j OpsRequest operational flow</figcaption>
</figure>

When you create a `Neo4jOpsRequest`, KubeDB Ops-manager performs the operation in phases:

1. Validates `spec.type` and operation-specific fields.
2. Resolves the target database from `spec.databaseRef`.
3. Pauses conflicting reconciliations for safe execution.
4. Applies the requested operation (for example scaling, restart, reconfigure, TLS update).
5. Updates status conditions and marks `.status.phase` as `Successful` or `Failed`.
6. Resumes normal reconciliation after operation completion.

## Supported Ops Requests

- [Reconfigure](/docs/guides/neo4j/reconfigure/overview.md): Update Neo4j configuration values or custom config secret references.
- [Horizontal Scaling](/docs/guides/neo4j/scaling/horizontal-scaling/overview.md): Add or remove Neo4j server members.
- [Vertical Scaling](/docs/guides/neo4j/scaling/vertical-scaling/overview.md): Update CPU and memory requests/limits.
- [Volume Expansion](/docs/guides/neo4j/volume-expansion/overview.md): Expand PVC size for Neo4j data volumes.
- [StorageClass Migration](/docs/guides/neo4j/migration/storageMigration.md): Migrate database PVCs from one StorageClass to another.
- [Update Version](/docs/guides/neo4j/update-version/overview.md): Upgrade Neo4j to a target `Neo4jVersion`.
- [Reconfigure TLS](/docs/guides/neo4j/reconfigure-tls/overview.md): Rotate, remove, or re-issue TLS configuration.
- [Rotate Auth](/docs/guides/neo4j/rotate-auth/overview.md): Rotate database credentials using generated or user-provided secrets.
- [Restart](/docs/guides/neo4j/restart/restart.md): Restart Neo4j pods in a controlled way.


## Next Step

Choose the operation-specific guide for the step-by-step manifest and verification workflow.
