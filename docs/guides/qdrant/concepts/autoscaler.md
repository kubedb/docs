---
title: QdrantAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-concepts
    name: QdrantAutoscaler
    parent: qdrant-concepts-qdrant
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

`spec.compute` specifies the compute (CPU and memory) autoscaling configuration. It contains a `node` sub-section with the following fields:

- `spec.compute.node.trigger` — enables (`On`) or disables (`Off`) compute autoscaling.
- `spec.compute.node.podLifeTimeThreshold` — the minimum age of a pod before VPA can recommend resource updates.
- `spec.compute.node.resourceDiffPercentage` — the minimum percentage difference required before applying a recommendation.
- `spec.compute.node.minAllowed` — the minimum allowed CPU and memory resources.
- `spec.compute.node.maxAllowed` — the maximum allowed CPU and memory resources.
- `spec.compute.node.controlledResources` — the list of resources to be controlled (e.g., `["cpu", "memory"]`).
- `spec.compute.node.containerControlledValues` — specifies whether to control `RequestsAndLimits` or `RequestsOnly`.

#### spec.storage

`spec.storage` specifies the storage autoscaling configuration. It contains a `node` sub-section with the following fields:

- `spec.storage.node.trigger` — enables (`On`) or disables (`Off`) storage autoscaling.
- `spec.storage.node.usageThreshold` — the storage usage threshold (percentage) that triggers autoscaling.
- `spec.storage.node.scalingThreshold` — the percentage by which storage will be scaled when triggered.
- `spec.storage.node.expansionMode` — the volume expansion mode (`Online` or `Offline`).

#### spec.opsRequestOptions

`spec.opsRequestOptions` specifies the options for the `QdrantOpsRequest` created by the autoscaler. It contains:

- `spec.opsRequestOptions.timeout` — the timeout for the generated ops request.
- `spec.opsRequestOptions.apply` — when to apply the ops request. Can be `Always` or `IfReady`.

## Next Steps

- Read the [Qdrant autoscaler overview](/docs/guides/qdrant/autoscaler/overview.md).
- See the [compute autoscaler guide](/docs/guides/qdrant/autoscaler/compute/compute-autoscale.md) and [storage autoscaler guide](/docs/guides/qdrant/autoscaler/storage/storage-autoscale.md).