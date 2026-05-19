---
title: QdrantAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-concepts
    name: QdrantAutoscaler
    parent: qdrant-concepts
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# QdrantAutoscaler

## What is QdrantAutoscaler

`QdrantAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for automatic scaling of [Qdrant](https://qdrant.tech/) compute resources (CPU, memory) and storage in a Kubernetes native way.

## QdrantAutoscaler CRD Specifications

Like any official Kubernetes resource, a `QdrantAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

**Sample `QdrantAutoscaler` for compute autoscaling:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: QdrantAutoscaler
metadata:
  name: qdrant-as-compute
  namespace: demo
spec:
  databaseRef:
    name: qdrant-sample
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 10m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 400m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

**Sample `QdrantAutoscaler` for storage autoscaling:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: QdrantAutoscaler
metadata:
  name: qdrant-as-storage
  namespace: demo
spec:
  databaseRef:
    name: qdrant-sample
  storage:
    node:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

### QdrantAutoscaler `Spec`

A `QdrantAutoscaler` object has the following fields in the `spec` section:

#### spec.databaseRef

`spec.databaseRef` is a required field that points to the [Qdrant](/docs/guides/qdrant/concepts/) object for which autoscaling will be performed. It contains:

- `spec.databaseRef.name` — the name of the target Qdrant database (required).

#### spec.compute

`spec.compute` specifies the compute (CPU and memory) autoscaling configuration. It contains:

- `spec.compute.node` — the per-node compute autoscaling configuration:
  - `trigger` — enables (`On`) or disables (`Off`) compute autoscaling.
  - `podLifeTimeThreshold` — the minimum age of a pod before VPA can recommend resource updates.
  - `resourceDiffPercentage` — the minimum percentage difference required before applying a recommendation.
  - `minAllowed` — the minimum allowed CPU and memory resources.
  - `maxAllowed` — the maximum allowed CPU and memory resources.
  - `controlledResources` — the list of resources to be controlled (e.g., `["cpu", "memory"]`).
  - `containerControlledValues` — specifies whether to control `RequestsAndLimits` or `RequestsOnly`.
  - `inMemoryStorage` — configuration for in-memory storage autoscaling per node:
    - `scalingFactorPercentage` — the scaling factor percentage for in-memory storage.
    - `usageThresholdPercentage` — the usage threshold percentage that triggers scaling.
- `spec.compute.nodeTopology` — specifies per-node topology for compute autoscaling:
  - `name` — the name of the topology entry.
  - `scaleDownDiffPercentage` — the scale-down difference percentage for this topology.
  - `scaleUpDiffPercentage` — the scale-up difference percentage for this topology.

#### spec.storage

`spec.storage` specifies the storage autoscaling configuration. It contains a `node` sub-section with the following fields:

- `spec.storage.node.trigger` — enables (`On`) or disables (`Off`) storage autoscaling.
- `spec.storage.node.usageThreshold` — the storage usage threshold (percentage) that triggers autoscaling.
- `spec.storage.node.scalingThreshold` — the percentage by which storage will be scaled when triggered.
- `spec.storage.node.expansionMode` — the volume expansion mode (`Online` or `Offline`).
- `spec.storage.node.upperBound` — the upper bound for storage size.
- `spec.storage.node.scalingRules` — a list of scaling rule objects, each containing:
  - `appliesUpto` — the upper limit for which this rule applies.
  - `threshold` — the threshold for this scaling rule.

#### spec.opsRequestOptions

`spec.opsRequestOptions` specifies the options for the `QdrantOpsRequest` created by the autoscaler. It contains:

- `spec.opsRequestOptions.timeout` — the timeout for the generated ops request.
- `spec.opsRequestOptions.apply` — when to apply the ops request. Can be `Always` or `IfReady`.
- `spec.opsRequestOptions.maxRetries` — the maximum number of retries for the ops request.

## Next Steps

- Read the [Qdrant autoscaler overview](/docs/guides/qdrant/autoscaler/overview.md).
- See the [compute autoscaler guide](/docs/guides/qdrant/autoscaler/compute/compute-autoscale.md) and [storage autoscaler guide](/docs/guides/qdrant/autoscaler/storage/storage-autoscale.md).