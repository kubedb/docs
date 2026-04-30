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

# Qdrant Ops Request

This guide lists the Qdrant operations currently documented for KubeDB.

## Before You Begin

- Deploy Qdrant first using the [quickstart guide](/docs/guides/qdrant/quickstart/quickstart.md).
- Review the operation-specific pages before applying changes in production.

## Supported Ops Requests

- [HorizontalScaling](/docs/guides/qdrant/scaling/horizontal-scaling/overview.md)
- [Reconfigure](/docs/guides/qdrant/reconfigure/overview.md)
- [ReconfigureTLS](/docs/guides/qdrant/reconfigure-tls/overview.md)
- [Restart](/docs/guides/qdrant/restart/restart.md)
- [RotateAuth](/docs/guides/qdrant/rotate-auth/overview.md)
- [VerticalScaling](/docs/guides/qdrant/scaling/vertical-scaling/overview.md)
- [VolumeExpansion](/docs/guides/qdrant/volume-expansion/overview.md)
- [UpdateVersion](/docs/guides/qdrant/update-version/overview.md)

## How Ops Requests Work

Create a `QdrantOpsRequest` for the target database, wait for the request to complete, and verify both the request object and the database status before moving to the next change.

## Next Steps

- Choose the specific operation page that matches your intended change.
- Apply one operation at a time and wait for completion before starting the next.
