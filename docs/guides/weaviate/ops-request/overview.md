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

# Weaviate Ops Request

This guide lists the Weaviate operation categories currently documented in the guide tree.

## Before You Begin

- Deploy Weaviate first using the [quickstart guide](/docs/guides/weaviate/quickstart/quickstart.md).
- Review the operation-specific pages for documentation status.
- This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD.

## Documented Operation Categories

- [Reconfigure](/docs/guides/weaviate/reconfigure/overview.md)
- [VerticalScaling](/docs/guides/weaviate/scaling/vertical-scaling/overview.md)
- [VolumeExpansion](/docs/guides/weaviate/volume-expansion/overview.md)
- [UpdateVersion](/docs/guides/weaviate/update-version/overview.md)
- [RotateAuth](/docs/guides/weaviate/rotate-auth/overview.md)
- [Restart](/docs/guides/weaviate/restart/restart.md)

## How Ops Requests Work

There is no CRD-backed `WeaviateOpsRequest` schema in this repository today, so no validated operation manifest can be generated from the current source tree.

## Next Steps

- Choose the specific operation page that matches your intended change.
- Apply one operation at a time and verify support against your installed release.
