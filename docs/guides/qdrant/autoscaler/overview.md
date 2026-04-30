---
title: Qdrant Autoscaler Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-overview
    name: Overview
    parent: qdrant-autoscaler
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Qdrant Autoscaler

This guide summarizes the autoscaling documentation status for Qdrant.

## Before You Begin

- Deploy Qdrant first using the [quickstart guide](/docs/guides/qdrant/quickstart/quickstart.md).
- Verify feature availability in your installed KubeDB release before applying any autoscaler examples.

## Available Autoscaling Modes

This repository does not currently contain a `QdrantAutoscaler` Go type or CRD.

The compute and storage autoscaler pages are retained as placeholders so the guide tree is complete, but they do not represent CRD-validated manifests in the current repo.

## Related Guides

- [Compute Autoscaler](/docs/guides/qdrant/autoscaler/compute/overview.md)
- [Storage Autoscaler](/docs/guides/qdrant/autoscaler/storage/overview.md)

## Next Steps

- Start with conservative limits and thresholds.
- Review autoscaler recommendations before enabling broad production rollouts.
