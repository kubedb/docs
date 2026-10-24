---
title: Vertical Scaling a DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-vertical-scaling-guide
    name: Vertical Scaling
    parent: dw-scaling-durableworkflow
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling a DurableWorkflow

## Before You Begin

- Have the `temporal` cluster of the [quickstart](/docs/guides/durableworkflow/quickstart/temporal.md), with 2 replicas for each role, in the `demo` namespace.
- You should be familiar with [how scaling works](/docs/guides/durableworkflow/scaling/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Change the resources

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: temporal
  verticalScaling:
    history:
      resources:
        requests:
          cpu: "1"
          memory: 2Gi
        limits:
          memory: 2Gi
    frontend:
      resources:
        requests:
          cpu: 500m
          memory: 512Mi
        limits:
          memory: 1Gi
```

Here, `spec.verticalScaling` has one optional entry for each role, with the `resources` of the container, and optionally `nodeSelectionPolicy` and `topology`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/scaling/vertical.yaml
durableworkflowopsrequest.ops.kubedb.com/dw-vscale created
```

The roles you named are rolled one pod at a time. When it is `Successful`, check the pods:

```bash
$ kubectl get pod -n demo temporal-history-0 -o jsonpath='{.spec.containers[0].resources}'
```

## Cleaning up

```bash
kubectl delete dwops -n demo dw-vscale
```

## Next Steps

- Let the [autoscaler](/docs/guides/durableworkflow/autoscaler/autoscaler.md) do it.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
