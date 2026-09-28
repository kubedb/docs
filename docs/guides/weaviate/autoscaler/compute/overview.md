---
title: Weaviate Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-compute-overview
    name: Overview
    parent: weaviate-autoscaler-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Compute Resource Autoscaling

This guide will give you an overview of how KubeDB autoscales the compute resources (CPU and Memory) of a `Weaviate` cluster using a `WeaviateAutoscaler`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Vertical Scaling](/docs/guides/weaviate/scaling/vertical-scaling/overview.md)

## How Compute Autoscaling Works

KubeDB provides a `WeaviateAutoscaler` CRD to automatically scale the compute resources of a Weaviate cluster. It is backed by a `VerticalPodAutoscaler` (VPA) that observes the actual resource usage of the Weaviate pods (requires `metrics-server`).

The compute autoscaling process consists of the following steps:

1. The user creates a `WeaviateAutoscaler` CR with a `spec.compute.weaviate` block describing the trigger, the min/max allowed resources, and the controlled resources.

2. The `KubeDB` Autoscaler operator creates a `VerticalPodAutoscaler` for the cluster and watches the recommendations it produces.

3. When the recommended resources differ from the current resources by more than `resourceDiffPercentage` (and the pods are older than `podLifeTimeThreshold`), the Autoscaler operator creates a `WeaviateOpsRequest` of type `VerticalScaling`.

4. The `KubeDB` Ops Manager applies the `VerticalScaling` ops request, updating the pod resources within the `minAllowed`/`maxAllowed` bounds.

The relevant fields under `spec.compute.weaviate` are:

- `trigger` — `On` or `Off`, enables/disables compute autoscaling.
- `podLifeTimeThreshold` — the minimum age of a Pod before a recommendation can be applied.
- `resourceDiffPercentage` — the minimum percentage change required before a new recommendation is applied.
- `minAllowed` / `maxAllowed` — the lower and upper bounds for the autoscaled resources.
- `controlledResources` — the resource types to autoscale (e.g. `cpu`, `memory`).
- `containerControlledValues` — whether to control `RequestsAndLimits` or just `Requests`.

In the next doc, we are going to show a step-by-step guide on autoscaling the compute resources of a Weaviate cluster.
