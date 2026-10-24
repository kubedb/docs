---
title: Update Version of a DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-updateversion-guide
    name: Guide
    parent: dw-updateversion-durableworkflow
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update Version of a DurableWorkflow

This guide shows how to update the version of a DurableWorkflow with a `DurableWorkflowOpsRequest`.

## Before You Begin

- Have a highly available cluster that runs a version which has a newer version in its `updateConstraints.allowlist`. List the versions with `kubectl get dwversion`. The names in the samples (`temporal-1.32.2`, `cadence-1.4.3`) are examples of versions that **a cluster administrator adds to the catalog when they are released**. They are not in the catalog today.
- You should be familiar with [how the update works](/docs/guides/durableworkflow/update-version/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Update Temporal

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-update-temporal
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: temporal
  updateVersion:
    targetVersion: temporal-1.32.2
```

Here,

- `spec.updateVersion.targetVersion` is the `DurableWorkflowVersion` to move to.
- `spec.updateVersion.allowDowntime` (optional) is needed only for a hop whose schema update cannot run online.
- `spec.updateVersion.surge` (optional) turns the surge pod on or off per role.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/update-version/temporal.yaml
durableworkflowopsrequest.ops.kubedb.com/dw-update-temporal created
```

## Update Cadence

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-update-cadence
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: cadence
  updateVersion:
    targetVersion: cadence-1.4.3
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/update-version/cadence.yaml
```

## Watch it

```bash
$ kubectl get dwops -n demo dw-update-temporal -w
NAME                 TYPE            STATUS        AGE
dw-update-temporal   UpdateVersion   Progressing   2m

$ kubectl get dw -n demo temporal
NAME       VERSION           STATUS   AGE
temporal   temporal-1.32.2   Ready    40m
```

If the new pods cannot become ready, the request ends `Failed` and `spec.version` is still the old one. Read `kubectl describe dwops` for the step that failed, fix the cause (for example the image), and create a new request.

## Cleaning up

```bash
kubectl delete dwops -n demo dw-update-temporal dw-update-cadence
```

## Next Steps

- Learn about [recommendations](/docs/guides/durableworkflow/recommendation/recommendation.md) that propose a version update.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
