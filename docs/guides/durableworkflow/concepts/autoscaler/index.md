---
title: DurableWorkflowAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: dw-autoscaler-concepts
    name: DurableWorkflowAutoscaler
    parent: dw-concepts-durableworkflow
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DurableWorkflowAutoscaler

## What is DurableWorkflowAutoscaler

`DurableWorkflowAutoscaler` is a Kubernetes `Custom Resource Definition` (CRD). It lets you configure the compute autoscaling of the roles of a [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md). The short name is `dwscaler`.

Only CPU and memory can be autoscaled. The servers have no volumes, so there is no storage autoscaler.

{{< notice type="warning" message="The autoscaler is implemented but has not been verified on a cluster yet." >}}

## DurableWorkflowAutoscaler Spec

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DurableWorkflowAutoscaler
metadata:
  name: temporal-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: temporal
  opsRequestOptions:
    timeout: 30m
    apply: IfReady
  compute:
    frontend:
      trigger: "On"
      podLifeTimeThreshold: 10m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 250m
        memory: 256Mi
      maxAllowed:
        cpu: "2"
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
```

### spec.databaseRef

The `DurableWorkflow` to scale.

### spec.compute

One entry for each of `frontend`, `history`, `matching`, `worker` and `internalFrontend` (Temporal only). A role without an entry is not autoscaled. Each entry has:

- `trigger`: `"On"` or `"Off"` (the default).
- `minAllowed` and `maxAllowed`: the bounds of the recommendation.
- `controlledResources`: `cpu`, `memory` or both (the default).
- `containerControlledValues`: `RequestsAndLimits` (the default) or `RequestsOnly`.
- `resourceDiffPercentage`: the smallest difference, in percent, between the current and the recommended resources that triggers a change.
- `podLifeTimeThreshold`: the minimum age of a pod before it is changed, 15 minutes by default.

### spec.opsRequestOptions

`timeout`, `apply` and `maxRetries` of the `VerticalScaling` requests that the autoscaler creates.

## How it works

The autoscaler watches the resource usage of the pods of each role. When the recommendation differs enough from the current resources, it creates a `DurableWorkflowOpsRequest` of type `VerticalScaling` for the role. That request rolls the pods one at a time, so a highly available cluster stays available. See [Autoscaler](/docs/guides/durableworkflow/autoscaler/autoscaler.md).

## Next Steps

- Run the [autoscaler guide](/docs/guides/durableworkflow/autoscaler/autoscaler.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
