---
title: WeaviateAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-concepts
    name: WeaviateAutoscaler
    parent: weaviate-concepts-weaviate
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# WeaviateAutoscaler

## What is WeaviateAutoscaler

`WeaviateAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for automatic scaling of [Weaviate](https://weaviate.tech/) compute resources (CPU, memory) and storage in a Kubernetes native way.

## WeaviateAutoscaler CRD Specifications

Like any official Kubernetes resource, a `WeaviateAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

**Sample `WeaviateAutoscaler` for compute autoscaling:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: WeaviateAutoscaler
metadata:
  name: weaviate-as-compute
  namespace: demo
spec:
  databaseRef:
    name: weaviate-sample
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

**Sample `WeaviateAutoscaler` for storage autoscaling:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: WeaviateAutoscaler
metadata:
  name: weaviate-as-storage
  namespace: demo
spec:
  databaseRef:
    name: weaviate-sample
  storage:
    node:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

### WeaviateAutoscaler `Spec`

A `WeaviateAutoscaler` object has the following fields in the `spec` section:

#### spec.databaseRef

`spec.databaseRef` is a required field that points to the [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) object for which autoscaling will be performed. It contains:

- `spec.databaseRef.name` - the name of the target Weaviate database (required).

#### spec.compute

`spec.compute` specifies the compute (CPU and memory) autoscaling configuration. It contains a `node` sub-section with the following fields:

- `spec.compute.node.trigger` - enables (`On`) or disables (`Off`) compute autoscaling.
- `spec.compute.node.podLifeTimeThreshold` - the minimum age of a pod before VPA can recommend resource updates.
- `spec.compute.node.resourceDiffPercentage` - the minimum percentage difference required before applying a recommendation.
- `spec.compute.node.minAllowed` - the minimum allowed CPU and memory resources.
- `spec.compute.node.maxAllowed` - the maximum allowed CPU and memory resources.
- `spec.compute.node.controlledResources` - the list of resources to be controlled (e.g., `["cpu", "memory"]`).
- `spec.compute.node.containerControlledValues` - specifies whether to control `RequestsAndLimits` or `RequestsOnly`.

#### spec.storage

`spec.storage` specifies the storage autoscaling configuration. It contains a `node` sub-section with the following fields:

- `spec.storage.node.trigger` - enables (`On`) or disables (`Off`) storage autoscaling.
- `spec.storage.node.usageThreshold` - the storage usage threshold (percentage) that triggers autoscaling.
- `spec.storage.node.scalingThreshold` - the percentage by which storage will be scaled when triggered.
- `spec.storage.node.expansionMode` - the volume expansion mode (`Online` or `Offline`).

#### spec.opsRequestOptions

`spec.opsRequestOptions` specifies the options for the `WeaviateOpsRequest` created by the autoscaler. It contains:

- `spec.opsRequestOptions.timeout` - the timeout for the generated ops request.
- `spec.opsRequestOptions.apply` - when to apply the ops request. Can be `Always` or `IfReady`.

## Next Steps

- Read the [Weaviate autoscaler overview](/docs/guides/weaviate/autoscaler/overview.md).
- See the [compute autoscaler guide](/docs/guides/weaviate/autoscaler/compute/cluster.md) and [storage autoscaler guide](/docs/guides/weaviate/autoscaler/storage/cluster.md).
