---
title: Rotate Authentication of a DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-rotate-auth-guide
    name: Guide
    parent: dw-rotateauth-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of a DurableWorkflow

## Before You Begin

- Have the `temporal` cluster of the [quickstart](/docs/guides/durableworkflow/quickstart/temporal.md), with `createDatabase` left at its default, in the `demo` namespace.
- You should be familiar with [how Rotate Authentication works](/docs/guides/durableworkflow/rotateauth/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Check the current credentials

```bash
$ kubectl get secret -n demo temporal-default-auth -o jsonpath='{.data.username}' | base64 -d
dw_temporal
```

## Rotate with a generated password

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: temporal
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/rotateauth/rotate-auth.yaml
durableworkflowopsrequest.ops.kubedb.com/dw-rotate-auth created
```

When the request is `Successful`, the Secret has the next user and a new password:

```bash
$ kubectl get secret -n demo temporal-default-auth -o jsonpath='{.data.username}' | base64 -d
dw_temporal_b
```

## Rotate with a password you choose

Put the new password in a Secret, under the key `password`, and refer to it:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dw-new-store-password
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: dw_temporal
  password: change-me-too
---
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-rotate-auth-secret
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: temporal
  authentication:
    secretRef:
      kind: Secret
      name: dw-new-store-password
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/rotateauth/rotate-auth-secret.yaml
```

Only the `password` key is read. The user name is chosen by the operator.

## Cleaning up

```bash
kubectl delete dwops -n demo dw-rotate-auth dw-rotate-auth-secret
kubectl delete secret -n demo dw-new-store-password
```

## Next Steps

- Let the [recommendation engine](/docs/guides/durableworkflow/recommendation/recommendation.md) tell you when to rotate.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
