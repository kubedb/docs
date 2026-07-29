---
title: Milvus Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-autoscaler-storage-overview
    name: Overview
    parent: milvus-autoscaler-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus Storage Autoscaling

This guide will give an overview on how the KubeDB Autoscaler operator autoscales the persistent storage of a `Milvus` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusAutoscaler](/docs/guides/milvus/concepts/milvusautoscaler.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

## How Storage Autoscaling Works

A `MilvusAutoscaler` of type `storage` watches PVC usage and, when a volume crosses the configured usage threshold, creates a `VolumeExpansion` `MilvusOpsRequest` to grow the volume.

`spec.storage` is keyed by the workload that carries persistent storage:

- **Standalone:** `node`.
- **Distributed:** `streamingnode` — among the distributed roles, only `streamingnode` has a persistent volume, so it is the sole storage-autoscaling target.

```yaml
spec:
  storage:
    streamingnode:                 # or 'node' for standalone
      trigger: "On"
      usageThreshold: 34
      expansionMode: "Online"
      scalingRules:
        - appliesUpto: "100Ti"
          threshold: "50%"
```

- **`trigger`** — `On`/`Off`.
- **`usageThreshold`** — percentage of disk usage that triggers expansion.
- **`expansionMode`** — `Online` or `Offline` (passed through to the generated `VolumeExpansion` ops request).
- **`scalingRules`** — how much to grow, optionally varying by current size.

The flow is:

1. A user creates a `MilvusAutoscaler` with `spec.storage`.
2. The autoscaler reads PVC usage from Prometheus.
3. When usage exceeds `usageThreshold`, the autoscaler creates a `VolumeExpansion` `MilvusOpsRequest` sized per `scalingRules`.
4. The Ops-manager operator performs the volume expansion as usual.

> **Prerequisites:** Prometheus must be collecting volume metrics, and the PVC's `StorageClass` must have `allowVolumeExpansion: true`.

In the next doc, we will see a step-by-step guide on storage autoscaling of a Milvus database.
