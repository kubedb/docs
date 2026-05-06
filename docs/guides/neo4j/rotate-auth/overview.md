---
title: Rotating Neo4j Credentials
menu:
  docs_{{ .version }}:
    identifier: neo4j-rotate-auth-overview
    name: Overview
    parent: neo4j-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Neo4j Overview

This page explains how KubeDB Ops-manager rotates Neo4j credentials using `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How RotateAuth Works

For a `Neo4jOpsRequest` with `spec.type: RotateAuth`, KubeDB Ops-manager:

1. Validates rotate-auth request and target database.
2. Uses one of the supported credential sources:
   - operator-managed generated secret,
   - user-provided secret from `spec.authentication.secretRef`.
3. Rotates credentials in Neo4j and updates auth secret state.
4. Ensures database authentication remains healthy.
5. Marks the request `Successful` when rotation completes.

## Next Step

Follow the detailed guide: [Rotate Auth for Neo4j](/docs/guides/neo4j/rotate-auth/rotateauth.md).
