---
title: Restart Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-restart-overview
    name: Restart Weaviate
    parent: weaviate-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Weaviate

KubeDB supports restarting Weaviate through a `WeaviateOpsRequest` with `type: Restart`.

## Before You Begin

- You need a Kubernetes cluster and configured `kubectl`.
- Install KubeDB and Ops Manager following [setup docs](/docs/setup/README.md).
- Review [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md).
- Use `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/restart/ops-request.yaml`.

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

## Apply Restart OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: weaviate-sample
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/restart/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-restart created
```

During restart, Ops Manager restarts Weaviate pods in a controlled order and reconciles the database back to `Ready` state.

## Verify

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-restart
$ kubectl describe weaviateopsrequest -n demo weaviate-restart
$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
$ kubectl get weaviate -n demo weaviate-sample
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-restart
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
