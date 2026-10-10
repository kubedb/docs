---
title: Neo4jAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: neo4j-autoscaler-concepts
    name: Neo4jAutoscaler
    parent: neo4j-concepts
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4jAutoscaler

## What is Neo4jAutoscaler?

`Neo4jAutoscaler` is a Kubernetes custom resource that declares how KubeDB should automatically scale the compute resources and persistent storage of a Neo4j cluster. The Autoscaler operator translates its recommendations into [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md) resources, which are executed by Ops Manager.

The following example enables both compute and storage autoscaling:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: Neo4jAutoscaler
metadata:
  name: neo4j-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: neo4j-autoscale
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
    maxRetries: 3
  compute:
    neo4j:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 1200Mi
      maxAllowed:
        cpu: "2"
        memory: 2Gi
      controlledResources:
        - cpu
        - memory
      containerControlledValues: RequestsAndLimits
  storage:
    neo4j:
      trigger: "On"
      usageThreshold: 80
      scalingThreshold: 50
      expansionMode: Online
```

## Specification

Like other Kubernetes resources, `Neo4jAutoscaler` contains `apiVersion`, `kind`, `metadata`, `spec`, and `status`. Users declare the policy in `spec`; the operator reports observed state in `status`.

### `spec.databaseRef`

`spec.databaseRef` is required and identifies the `Neo4j` resource in the same namespace.

| Field | Description |
|---|---|
| `name` | Name of the target KubeDB `Neo4j` resource. |

### `spec.opsRequestOptions`

These options are copied to every `Neo4jOpsRequest` created by the Autoscaler.

| Field | Description |
|---|---|
| `apply` | `IfReady` creates operations only while the database is ready; `Always` permits creation regardless of readiness. Defaults to `IfReady`. |
| `timeout` | Maximum duration allowed for each operation step. |
| `maxRetries` | Maximum retries for a failed operation. Defaults to `1`. |

### `spec.compute`

`spec.compute.neo4j` controls CPU and memory recommendations for the `neo4j` container.

| Field | Description |
|---|---|
| `trigger` | Enables autoscaling when set to `On`; use `Off` to disable it without deleting the resource. |
| `minAllowed` | Lower CPU and memory bounds for recommendations. |
| `maxAllowed` | Upper CPU and memory bounds for recommendations. |
| `controlledResources` | Resources controlled by the Autoscaler, normally `cpu` and `memory`. |
| `containerControlledValues` | `RequestsAndLimits` updates both values; `RequestsOnly` updates only requests. |
| `resourceDiffPercentage` | Minimum percentage difference between the current allocation and a recommendation before an update is applied. Defaults to `50`. |
| `podLifeTimeThreshold` | Minimum pod lifetime considered when deciding whether to apply a recommendation. Defaults to `15m`. |

`spec.compute.nodeTopology` is optional. When set, the Autoscaler selects resources from the named `NodeTopology` instead of applying an arbitrary recommendation. `scaleUpDiffPercentage` and `scaleDownDiffPercentage` control when it moves between topology entries; their defaults are `15` and `25`, respectively.

### `spec.storage`

`spec.storage.neo4j` controls expansion of the Neo4j data volumes.

| Field | Description |
|---|---|
| `trigger` | Enables storage autoscaling when set to `On`. |
| `usageThreshold` | Used-capacity percentage at which expansion is triggered. Defaults to `80`. |
| `scalingThreshold` | Percentage by which the current volume is increased. Defaults to `50`. |
| `scalingRules` | Optional size-dependent rules. Each rule has an `appliesUpto` capacity and a `threshold` percentage or absolute quantity. |
| `upperBound` | Optional maximum volume size. |
| `expansionMode` | Required expansion strategy: `Online` or `Offline`. |

For example, the following rules grow smaller volumes proportionally and larger volumes by a fixed amount:

```yaml
storage:
  neo4j:
    trigger: "On"
    usageThreshold: 80
    expansionMode: Online
    upperBound: 2Ti
    scalingRules:
      - appliesUpto: 500Gi
        threshold: 30pc
      - appliesUpto: 1Ti
        threshold: 20pc
      - appliesUpto: ""
        threshold: 100Gi
```

### `status`

The status is managed by KubeDB and should not be edited. Important fields include:

| Field | Description |
|---|---|
| `phase` | Current Autoscaler phase, such as `InProgress`, `Current`, or `Failed`. |
| `observedGeneration` | Most recent resource generation processed by the operator. |
| `conditions` | Events and outcomes reported by the Autoscaler controller. |
| `vpas` | Current compute recommendations and their conditions. |
| `checkpoints` | Historical CPU and memory samples used by the recommender. |

## Next Steps

- [Autoscale Neo4j compute resources](/docs/guides/neo4j/autoscaler/compute/autoscale.md)
- [Autoscale Neo4j storage](/docs/guides/neo4j/autoscaler/storage/autoscale.md)
