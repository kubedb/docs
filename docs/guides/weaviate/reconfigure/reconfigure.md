---
title: Reconfigure Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-reconfigure-cluster
    name: Cluster
    parent: weaviate-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Weaviate

This guide shows how to use `WeaviateOpsRequest` with `type: Reconfigure` to update Weaviate runtime configuration.

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured.
- Install KubeDB and Ops Manager from [setup docs](/docs/setup/README.md).
- Review [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) and [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md).

To keep things isolated, use a namespace named `demo`:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy Weaviate

Apply the sample manifest:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
weaviate.kubedb.com/weaviate-sample created
```

Wait for readiness:

```bash
$ kubectl get weaviate -n demo weaviate-sample -w
NAME              VERSION   STATUS   AGE
weaviate-sample   1.33.1    Ready    2m
```

## Apply Reconfigure OpsRequest

Use the following request:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: weaviate-sample
  configuration:
    applyConfig:
      weaviate.yaml: |
        LOG_LEVEL: info
```

Apply from the example file:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-reconfigure created
```

## Verify Reconfiguration

Watch request status:

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-reconfigure
NAME                  TYPE          STATUS       AGE
weaviate-reconfigure  Reconfigure   Successful   2m
```

Inspect reconciliation details:

```bash
$ kubectl describe weaviateopsrequest -n demo weaviate-reconfigure
```

Confirm database is healthy after config rollout:

```bash
$ kubectl get weaviate -n demo weaviate-sample
$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-reconfigure
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```