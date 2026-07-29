---
title: Weaviate Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-storage-overview
    name: Overview
    parent: weaviate-autoscaler-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Storage Autoscaling

This guide will give you an overview of how KubeDB autoscales the storage of a `Weaviate` cluster using a `WeaviateAutoscaler`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Volume Expansion](/docs/guides/weaviate/volume-expansion/overview.md)

## How Storage Autoscaling Works

KubeDB provides a `WeaviateAutoscaler` CRD to automatically expand the storage of a Weaviate cluster when the volumes start filling up. Storage autoscaling requires a `StorageClass` that supports volume expansion (`allowVolumeExpansion: true`).

The storage autoscaling process consists of the following steps:

1. The user creates a `WeaviateAutoscaler` CR with a `spec.storage.weaviate` block describing the trigger, the usage threshold, and the scaling factor.

2. The `KubeDB` Autoscaler operator watches the PVC usage of the Weaviate pods.

3. When a volume's used space crosses `usageThreshold` (a percentage of the volume capacity), the Autoscaler operator creates a `WeaviateOpsRequest` of type `VolumeExpansion`, increasing the volume size by `scalingThreshold` percent.

4. The `KubeDB` Ops Manager applies the `VolumeExpansion` ops request, expanding the PVCs.

The relevant fields under `spec.storage.weaviate` are:

- `trigger` — `On` or `Off`, enables/disables storage autoscaling.
- `usageThreshold` — the percentage of used space that triggers expansion.
- `scalingThreshold` — the percentage by which the volume is expanded each time.
- `expansionMode` — `Online` or `Offline`.

In the next doc, we are going to show a step-by-step guide on autoscaling the storage of a Weaviate cluster.
