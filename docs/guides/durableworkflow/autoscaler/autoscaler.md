---
title: Autoscaling a DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-autoscaler-guide
    name: Compute Autoscaler
    parent: dw-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resources of a DurableWorkflow

This guide shows how to let KubeDB adjust the CPU and memory of the roles of a DurableWorkflow with a `DurableWorkflowAutoscaler`. There is no storage autoscaling, because the servers have no volumes.

{{< notice type="warning" message="The autoscaler is implemented but has not been verified on a cluster yet. The behavior below is taken from its code and from the other KubeDB autoscalers." >}}

## Before You Begin

- Install the KubeDB autoscaler together with the operator, and have a `temporal` cluster as in the [quickstart](/docs/guides/durableworkflow/quickstart/temporal.md). Give the roles resource requests, for example with the `podTemplate` of [Roles and High Availability](/docs/guides/durableworkflow/clustering/roles-and-ha.md), so that there is something to compare to.
- You should be familiar with [DurableWorkflowAutoscaler](/docs/guides/durableworkflow/concepts/autoscaler/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create the autoscaler

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
    history:
      trigger: "On"
      podLifeTimeThreshold: 10m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 500m
        memory: 512Mi
      maxAllowed:
        cpu: "4"
        memory: 8Gi
      controlledResources: ["cpu", "memory"]
```

Here, only `frontend` and `history` are autoscaled; the other roles keep their resources. The keys of `compute` are the roles. See [DurableWorkflowAutoscaler](/docs/guides/durableworkflow/concepts/autoscaler/index.md) for all fields.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/autoscaler/autoscaler.yaml
durableworkflowautoscaler.autoscaling.kubedb.com/temporal-autoscaler created
```

## What happens

When the recommendation for a role leaves the range of the current resources by more than `resourceDiffPercentage`, and the pods are older than `podLifeTimeThreshold`, the autoscaler creates a `DurableWorkflowOpsRequest` of type `VerticalScaling` for that role, within `minAllowed` and `maxAllowed`:

```bash
$ kubectl get dwops -n demo
```

The request is processed like any [vertical scaling](/docs/guides/durableworkflow/scaling/vertical-scaling.md), one pod at a time. Autoscaling changes the resources only. The number of replicas is not autoscaled.

## Cleaning up

```bash
kubectl delete dwscaler -n demo temporal-autoscaler
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
