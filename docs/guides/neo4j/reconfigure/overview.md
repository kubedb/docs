---
title: Reconfiguring Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-overview
    name: Overview
    parent: neo4j-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Neo4j Overview

This page explains how KubeDB Ops-manager applies configuration changes to Neo4j using `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How Reconfigure Works

For a `Neo4jOpsRequest` with `spec.type: Reconfigure`, KubeDB Ops-manager:

1. Validates configuration inputs from `spec.configuration`.
2. Resolves custom config secret and inline `applyConfig` values.
3. Pauses conflicting reconciliations.
4. Merges or replaces Neo4j config based on request fields.
5. Restarts relevant pods to apply new configuration.
6. Verifies pod/database health and marks the request `Successful`.

## Next Step

Follow the detailed guide: [Reconfigure Neo4j Cluster](/docs/guides/neo4j/reconfigure/reconfigure.md).
