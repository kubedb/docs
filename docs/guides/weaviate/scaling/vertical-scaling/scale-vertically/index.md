---
title: Scale Weaviate Vertically
menu:
  docs_{{ .version }}:
    identifier: weaviate-scale-vertically
    name: Scale Vertically
    parent: weaviate-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling for Weaviate

This guide shows how to increase or decrease Weaviate node resources using `WeaviateOpsRequest`.

## Before You Begin

- Ensure your cluster has enough allocatable resources for the target CPU/memory.
- Install KubeDB and Ops Manager from [setup docs](/docs/setup/README.md).
- Review [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md).

Create a test namespace:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy Weaviate

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
weaviate.kubedb.com/weaviate-sample created

$ kubectl get weaviate -n demo weaviate-sample -w
NAME              VERSION   STATUS   AGE
weaviate-sample   1.33.1    Ready    2m
```

## Create VerticalScaling OpsRequest

Use this manifest:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: weaviate-sample
  verticalScaling:
    node:
      resources:
        requests:
          cpu: "500m"
          memory: 1Gi
        limits:
          cpu: "1"
          memory: 2Gi
```

Apply the sample file:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/scaling/vertical-scaling/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-vertical-scale created
```

## Verify Vertical Scaling

Monitor request:

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-vertical-scale
NAME                     TYPE              STATUS       AGE
weaviate-vertical-scale  VerticalScaling   Successful   3m
```

Inspect details:

```bash
$ kubectl describe weaviateopsrequest -n demo weaviate-vertical-scale
```

Verify pod resources after reconciliation:

```bash
$ kubectl get pod -n demo weaviate-sample-0 -o json | jq '.spec.containers[].resources'
```

Confirm database remains healthy:

```bash
$ kubectl get weaviate -n demo weaviate-sample
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-vertical-scale
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```