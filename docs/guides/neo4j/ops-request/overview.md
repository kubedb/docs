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

# Neo4j Ops Request

This guide lists the Neo4j operations currently documented for KubeDB.

## Before You Begin

- Deploy Neo4j first using the [quickstart guide](/docs/guides/neo4j/quickstart/quickstart.md).
- Review the operation-specific pages before applying changes in production.

## Supported Ops Requests

- [Reconfigure](/docs/guides/neo4j/reconfigure/overview.md)
- [HorizontalScaling](/docs/guides/neo4j/scaling/horizontal-scaling/overview.md)
- [VerticalScaling](/docs/guides/neo4j/scaling/vertical-scaling/overview.md)
- [VolumeExpansion](/docs/guides/neo4j/volume-expansion/overview.md)
- [UpdateVersion](/docs/guides/neo4j/update-version/overview.md)
- [ReconfigureTLS](/docs/guides/neo4j/reconfigure-tls/overview.md)
- [RotateAuth](/docs/guides/neo4j/rotate-auth/overview.md)
- [Restart](/docs/guides/neo4j/restart/restart.md)

## How Ops Requests Work

Create a `Neo4jOpsRequest` for the target database, wait for the request status to move through validation and execution phases, and then verify both the `Neo4jOpsRequest` and the `Neo4j` resources.

## Next Steps

- Choose the specific operation page that matches your intended change.
- Apply one operation at a time and wait for completion before starting the next.
