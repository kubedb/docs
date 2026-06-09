---
title: Weaviate Autoscaler Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-overview
    name: Overview
    parent: weaviate-autoscaler
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Autoscaling Overview

This guide will give an overview of how KubeDB autoscales `Weaviate` database resources - both compute (CPU and memory) and storage.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateAutoscaler](/docs/guides/weaviate/concepts/autoscaler.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)

## How Autoscaling Works

KubeDB uses the `WeaviateAutoscaler` CR to configure automatic scaling of Weaviate resources. There are two types of autoscaling supported:

### Compute Autoscaling

KubeDB leverages the [Kubernetes Vertical Pod Autoscaler (VPA)](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler) to recommend compute resource adjustments. The process works as follows:

1. The user creates a `WeaviateAutoscaler` CR with `spec.compute` configured.
2. KubeDB creates a VPA resource for the `Weaviate` StatefulSet.
3. The VPA monitors resource usage and provides recommendations.
4. When the recommendation differs from the current resources by more than `resourceDiffPercentage`, KubeDB creates a `WeaviateOpsRequest` with `type: VerticalScaling` to apply the recommended resources.
5. After the OpsRequest completes, the pods are running with the updated resource requests and limits.

### Storage Autoscaling

KubeDB monitors PVC usage to automatically expand storage. The process works as follows:

1. The user creates a `WeaviateAutoscaler` CR with `spec.storage` configured.
2. KubeDB monitors the PVC storage usage of the Weaviate pods.
3. When the disk usage exceeds the `usageThreshold` percentage, KubeDB creates a `WeaviateOpsRequest` with `type: VolumeExpansion` to expand the storage by `scalingThreshold` percent.
4. After the OpsRequest completes, the PVCs are expanded to the new size.

In the next docs, we are going to show step-by-step guides on compute and storage autoscaling for Weaviate databases.
