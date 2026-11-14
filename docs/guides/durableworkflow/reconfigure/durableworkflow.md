---
title: Reconfigure a DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-reconfigure-guide
    name: Guide
    parent: dw-reconfigure-durableworkflow
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure a DurableWorkflow

## Before You Begin

- Have the `temporal-config` cluster of [Custom Configuration](/docs/guides/durableworkflow/configuration/custom-config.md) running in the `demo` namespace.
- You should be familiar with [how Reconfigure works](/docs/guides/durableworkflow/reconfigure/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Change the dynamic configuration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-reconfigure-dynamic
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: temporal-config
  configuration:
    applyConfig:
      dynamic_config.yaml: |
        history.persistenceMaxQPS:
        - value: 5000
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/reconfigure/dynamic.yaml
durableworkflowopsrequest.ops.kubedb.com/dw-reconfigure-dynamic created
```

The request becomes `Successful` without any pod being replaced:

```bash
$ kubectl get dwops -n demo dw-reconfigure-dynamic
NAME                     TYPE          STATUS       AGE
dw-reconfigure-dynamic   Reconfigure   Successful   1m

$ kubectl get dw -n demo temporal-config -o jsonpath='{.spec.configuration.inline.dynamic_config\.yaml}'
history.persistenceMaxQPS:
- value: 5000
```

## Change the static configuration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-reconfigure-static
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: temporal-config
  configuration:
    applyConfig:
      server.yaml: |
        persistence:
          datastores:
            default:
              sql:
                maxConns: 8
                maxIdleConns: 8
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/reconfigure/static.yaml
durableworkflowopsrequest.ops.kubedb.com/dw-reconfigure-static created
```

This time every pod is replaced, one at a time. The `server.yaml` entries from the request replace the ones in `spec.configuration.inline`, so repeat in the request every key of `server.yaml` that you want to keep.

## Cleaning up

```bash
kubectl delete dwops -n demo dw-reconfigure-dynamic dw-reconfigure-static
kubectl delete dw -n demo temporal-config
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
