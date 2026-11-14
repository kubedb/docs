---
title: Horizontal Scaling a DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-horizontal-scaling-guide
    name: Horizontal Scaling
    parent: dw-scaling-durableworkflow
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling a DurableWorkflow

## Before You Begin

- Have the `temporal` cluster of the [quickstart](/docs/guides/durableworkflow/quickstart/temporal.md), with 2 replicas for each role, in the `demo` namespace.
- You should be familiar with [how scaling works](/docs/guides/durableworkflow/scaling/overview.md).
- Keep the [connection budget](/docs/guides/durableworkflow/quickstart/prerequisites.md#the-connection-budget) in mind: every added pod opens up to 10 connections.

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Scale up

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-hscale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: temporal
  horizontalScaling:
    frontend: 3
    history: 3
    matching: 3
```

Here, `spec.horizontalScaling` has one optional field for each role: `frontend`, `history`, `matching`, `worker` and `internalFrontend`. A role that is not named is not changed.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/scaling/horizontal.yaml
durableworkflowopsrequest.ops.kubedb.com/dw-hscale created

$ kubectl get pods -n demo -l app.kubernetes.io/instance=temporal
```

## Scale down

Apply a request with lower numbers. The pods leave one at a time:

```bash
$ kubectl apply -f - <<'YAML'
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: temporal
  horizontalScaling:
    frontend: 2
    history: 2
    matching: 2
YAML
```

## Cleaning up

```bash
kubectl delete dwops -n demo dw-hscale dw-hscale-down
```

## Next Steps

- [Vertical scaling](/docs/guides/durableworkflow/scaling/vertical-scaling.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
