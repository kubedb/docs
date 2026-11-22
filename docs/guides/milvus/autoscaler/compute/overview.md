---
title: Milvus Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-autoscaler-compute-overview
    name: Overview
    parent: milvus-autoscaler-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus Compute Resource Autoscaling

This guide will give an overview on how the KubeDB Autoscaler operator autoscales the compute resources (CPU/memory) of a `Milvus` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusAutoscaler](/docs/guides/milvus/concepts/milvusautoscaler.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

## How Compute Autoscaling Works

A `MilvusAutoscaler` of type `compute` watches the resource usage of the Milvus pods (via the metrics server and the autoscaler's built-in recommender) and, when usage drifts far enough from the requested resources, it creates a `VerticalScaling` `MilvusOpsRequest` to right-size the pods.

`spec.compute` is keyed by component:

- **Standalone:** `node`.
- **Distributed:** `proxy`, `mixcoord`, `datanode`, `querynode`, `streamingnode`.

Each block supports:

```yaml
spec:
  compute:
    node:                          # or proxy/mixcoord/datanode/querynode/streamingnode
      trigger: "On"
      podLifeTimeThreshold: 10m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 100m
        memory: 1Gi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
```

- **`trigger`** — `On`/`Off`; enables autoscaling for the component.
- **`minAllowed`/`maxAllowed`** — the bounds the autoscaler stays within.
- **`resourceDiffPercentage`** — how far current resources must drift from the recommendation before an ops request is created.
- **`podLifeTimeThreshold`** — minimum pod age before it is considered for scaling.

Every autoscaler-created OpsRequest restarts the pods it scales, and a freshly started Milvus briefly uses much more CPU. `podLifeTimeThreshold: 10m` and `resourceDiffPercentage: 20` keep that startup spike from triggering another re-scale right after a restart. Keep `minAllowed.memory` at `1Gi` or more: the autoscaler sets the memory limit equal to the request, and Milvus needs more than `256Mi`.
- **`controlledResources`** — which resources are managed.

The flow is:

1. A user creates a `MilvusAutoscaler` with `spec.compute`.
2. The autoscaler's built-in recommender produces resource recommendations and records them in the autoscaler's `status.vpas` (no `VerticalPodAutoscaler` object is created).
3. When the recommendation differs from the current resources by more than `resourceDiffPercentage`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest` (subject to `spec.opsRequestOptions`).
4. The Ops-manager operator applies the vertical scaling as usual, restarting the affected pods.

> **Prerequisite:** a **metrics server** must be installed in the cluster for the autoscaler to produce recommendations.

In the next doc, we will see a step-by-step guide on compute autoscaling of a Milvus database.
