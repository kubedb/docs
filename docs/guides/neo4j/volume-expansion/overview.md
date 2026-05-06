---
title: Expanding Neo4j Storage
menu:
  docs_{{ .version }}:
    identifier: neo4j-volume-expansion-overview
    name: Overview
    parent: neo4j-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Volume Expansion Overview

This page explains how KubeDB Ops-manager expands Neo4j data volumes using `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).
- Your StorageClass must support `allowVolumeExpansion: true`.

## How Volume Expansion Works

For a `Neo4jOpsRequest` with `spec.type: VolumeExpansion`, KubeDB Ops-manager:

1. Validates requested size from `spec.volumeExpansion.server`.
2. Validates expansion mode from `spec.volumeExpansion.mode`.
3. Pauses conflicting reconciliations.
4. Expands the target PVCs to the requested size.
5. Reconciles Neo4j state based on online/offline mode requirements.
6. Marks request `Successful` after PVC and pod health checks.

## Next Step

Follow the detailed guide: [Expand Neo4j Volume](/docs/guides/neo4j/volume-expansion/volume-expansion.md).
