---
title: Neo4j Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-vertical-scaling-overview
    name: Overview
    parent: neo4j-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Vertical Scaling Overview

This page explains how KubeDB Ops-manager updates Neo4j pod resources using `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How Vertical Scaling Works

For a `Neo4jOpsRequest` with `spec.type: VerticalScaling`, KubeDB Ops-manager:

1. Validates CPU/memory values from `spec.verticalScaling.server.resources`.
2. Pauses conflicting reconciliations.
3. Applies updated requests/limits to Neo4j server pods.
4. Performs controlled restarts where necessary.
5. Waits for pods to become healthy with new resources.
6. Marks the request `Successful` after reconciliation.

## Next Step

Follow the detailed guide: [Scale Neo4j Vertically](/docs/guides/neo4j/scaling/vertical-scaling/scale-vertically/index.md).
