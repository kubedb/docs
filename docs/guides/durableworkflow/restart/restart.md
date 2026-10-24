---
title: DurableWorkflow Restart
menu:
  docs_{{ .version }}:
    identifier: dw-restart-durableworkflow
    name: Restart
    parent: dw-durableworkflow-guides
    weight: 115
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart DurableWorkflow

KubeDB supports restarting a DurableWorkflow with a `DurableWorkflowOpsRequest`. Restarting is useful if some pods are stuck, or to make the pods read a Secret or a ConfigMap again. On a highly available cluster (at least 2 replicas per role) a restart causes no downtime: it was verified with a load generator that started and signaled a workflow every 0.3 seconds and saw 0 failed calls.

## Before You Begin

- Have a highly available cluster, for example the one of the [Temporal quickstart](/docs/guides/durableworkflow/quickstart/temporal.md), named `temporal` in the `demo` namespace.
- You should be familiar with [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md) and [DurableWorkflowOpsRequest](/docs/guides/durableworkflow/concepts/durableworkflow-ops-request/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Restart all roles

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: temporal
  apply: Always
```

Here,

- `spec.type` is the operation, `Restart`.
- `spec.databaseRef` references the DurableWorkflow in the namespace of the request.
- `spec.apply` is `IfReady` (the default) or `Always`. `Always` runs even if the cluster is not `Ready`, which is what you want to repair a stuck pod.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/restart/restart.yaml
durableworkflowopsrequest.ops.kubedb.com/dw-restart created
```

The pods are replaced one at a time, role by role, as described in [how rolling works](/docs/guides/durableworkflow/update-version/overview.md#how-the-pods-are-rolled). Watch it with:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=temporal -w
$ kubectl get dwops -n demo dw-restart
NAME         TYPE      STATUS       AGE
dw-restart   Restart   Successful   9m
```

## Restart some roles

`spec.restart.roles` limits the restart to the roles you list, from `frontend`, `history`, `matching`, `worker` and `internalFrontend`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-restart-history
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: temporal
  restart:
    roles:
    - history
    - matching
  timeout: 30m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/restart/restart-roles.yaml
```

A cluster that is not highly available has a gap: a role with one replica is unavailable while its only pod restarts.

## Cleaning up

```bash
kubectl delete dwops -n demo dw-restart dw-restart-history
```

## Next Steps

- [Update the version](/docs/guides/durableworkflow/update-version/overview.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
