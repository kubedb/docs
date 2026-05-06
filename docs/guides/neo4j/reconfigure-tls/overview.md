---
title: Reconfiguring Neo4j TLS
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-tls-overview
    name: Overview
    parent: neo4j-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Neo4j TLS Overview

This page explains how KubeDB Ops-manager manages TLS updates for Neo4j using `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How ReconfigureTLS Works

For a `Neo4jOpsRequest` with `spec.type: ReconfigureTLS`, KubeDB Ops-manager:

1. Validates TLS operation fields under `spec.tls`.
2. Handles one of the supported actions:
   - rotate certificates (`rotateCertificates`),
   - remove TLS (`remove`),
   - issue/re-issue TLS from `issuerRef`.
3. Pauses conflicting reconciliations.
4. Updates secrets/config and restarts affected members as required.
5. Verifies connectivity and pod health.
6. Marks the request `Successful` after reconciliation.

## Next Step

Follow the detailed guide: [Reconfigure TLS in Neo4j](/docs/guides/neo4j/reconfigure-tls/reconfigure-tls.md).
