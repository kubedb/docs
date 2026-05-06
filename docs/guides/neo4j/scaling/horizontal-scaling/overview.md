---
title: Neo4j Horizontal Scaling
menu:
  docs_{{ .version }}:
    identifier: neo4j-horizontal-scaling-overview
    name: Overview
    parent: neo4j-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Horizontal Scaling Overview

This page explains how KubeDB Ops-manager performs horizontal scaling for Neo4j using `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How Horizontal Scaling Works

For a `Neo4jOpsRequest` with `spec.type: HorizontalScaling`, KubeDB Ops-manager:

1. Validates the requested server count in `spec.horizontalScaling.server`.
2. Pauses conflicting reconciliations for safe scale execution.
3. Updates the target Neo4j server replica count.
4. Applies reallocation policy from `spec.horizontalScaling.reallocate`.
5. Waits for members and database hosting to reach a healthy state.
6. Marks the operation `Successful` and resumes normal reconciliation.

Use Cypher views to verify topology and allocation after scaling:

- `SHOW DATABASE <name>` for allocation status.
- `SHOW SERVERS` for hosting distribution.

## Next Step

Follow the detailed guide: [Scale Neo4j Horizontally](/docs/guides/neo4j/scaling/horizontal-scaling/scale-horizontally/index.md).
