---
title: MilvusAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-concepts-milvusautoscaler
    name: MilvusAutoscaler
    parent: milvus-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MilvusAutoscaler

## What is MilvusAutoscaler

`MilvusAutoscaler` is a Kubernetes `CustomResourceDefinition` (CRD). It provides declarative configuration for autoscaling the compute resources and persistent storage of a [Milvus](/docs/guides/milvus/concepts/milvus.md) database.

Depending on which section you configure, the autoscaler creates one of these ops requests automatically:

- `VerticalScaling` `MilvusOpsRequest` for compute autoscaling.
- `VolumeExpansion` `MilvusOpsRequest` for storage autoscaling.

## Sample MilvusAutoscaler Objects

### Standalone

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MilvusAutoscaler
metadata:
  name: milvus-standalone-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: milvus-standalone
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
      controlledResources: ["cpu", "memory"]
  storage:
    node:
      trigger: "On"
      usageThreshold: 30
      expansionMode: "Offline"
      scalingRules:
        - appliesUpto: "100Ti"
          threshold: "50%"
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

### Distributed

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MilvusAutoscaler
metadata:
  name: milvus-cluster-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: milvus-cluster
  compute:
    proxy:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
      controlledResources: ["cpu", "memory"]
    mixcoord:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
      controlledResources: ["cpu", "memory"]
    datanode:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
      controlledResources: ["cpu", "memory"]
    querynode:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
      controlledResources: ["cpu", "memory"]
    streamingnode:
      trigger: "On"
      podLifeTimeThreshold: 1m
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 1000m
        memory: 2Gi
      resourceDiffPercentage: 10
      controlledResources: ["cpu", "memory"]
  storage:
    streamingnode:
      trigger: "On"
      usageThreshold: 34
      expansionMode: "Online"
      scalingRules:
        - appliesUpto: "100Ti"
          threshold: "50%"
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

## MilvusAutoscaler Spec

### spec.databaseRef

`spec.databaseRef` is required. It points to the [Milvus](/docs/guides/milvus/concepts/milvus.md) object that will be autoscaled.

```yaml
spec:
  databaseRef:
    name: milvus-cluster
```

### spec.compute

`spec.compute` configures CPU and memory autoscaling.

The valid keys depend on the Milvus topology:

- Standalone: `node`
- Distributed: `proxy`, `mixcoord`, `datanode`, `querynode`, `streamingnode`

Each compute block supports:

- `trigger` - `On` or `Off`.
- `podLifeTimeThreshold` - minimum pod age before autoscaling is considered.
- `minAllowed` - lower resource bound.
- `maxAllowed` - upper resource bound.
- `resourceDiffPercentage` - minimum percentage difference between current resources and the recommendation before a scaling operation is created.
- `controlledResources` - which resources may be autoscaled, typically `cpu` and `memory`.

For compute autoscaling, the autoscaler creates a `VerticalPodAutoscaler` recommendation first and then generates a `VerticalScaling` [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md) when the recommendation passes the configured thresholds.

See [Compute Autoscaling Overview](/docs/guides/milvus/autoscaler/compute/overview.md).

### spec.storage

`spec.storage` configures storage autoscaling.

The valid keys depend on the Milvus topology:

- Standalone: `node`
- Distributed: `streamingnode`

Only `streamingnode` is supported for distributed Milvus because it is the only distributed role with persistent Milvus storage.

Each storage block supports:

- `trigger` - `On` or `Off`.
- `usageThreshold` - disk usage percentage that triggers expansion.
- `expansionMode` - `Online` or `Offline`.
- `scalingRules` - how much storage should be added when usage crosses the threshold.

For storage autoscaling, the autoscaler watches PVC usage from Prometheus and creates a `VolumeExpansion` [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md) when the configured threshold is crossed.

See [Storage Autoscaling Overview](/docs/guides/milvus/autoscaler/storage/overview.md).

### spec.opsRequestOptions

`spec.opsRequestOptions` controls how the internally generated `MilvusOpsRequest` should be created and executed.

The Milvus autoscaling guides currently use:

- `apply` - for example `IfReady`.
- `timeout` - the maximum time each generated ops request should get.

Example:

```yaml
spec:
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

## How It Works

### Compute autoscaling

1. A user creates a `MilvusAutoscaler` with `spec.compute`.
2. The autoscaler creates one VPA recommendation source per targeted component.
3. When the recommendation differs from current resources by more than `resourceDiffPercentage`, the autoscaler creates a `VerticalScaling` `MilvusOpsRequest`.
4. The Ops-manager applies the scaling operation.

### Storage autoscaling

1. A user creates a `MilvusAutoscaler` with `spec.storage`.
2. The autoscaler watches persistent-volume usage from Prometheus.
3. When usage exceeds `usageThreshold`, the autoscaler creates a `VolumeExpansion` `MilvusOpsRequest`.
4. The Ops-manager applies the volume expansion.

## Prerequisites

- Compute autoscaling requires a metrics server and VPA recommender.
- Storage autoscaling requires Prometheus volume metrics.
- Storage autoscaling also requires an expansion-capable `StorageClass`.

## Related Concepts

- [Milvus](/docs/guides/milvus/concepts/milvus.md)
- [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
