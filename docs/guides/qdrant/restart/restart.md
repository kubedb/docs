---
title: Restart Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-restart-overview
    name: Restart Qdrant
    parent: qdrant-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Qdrant

This guide shows how to restart Qdrant pods using `QdrantOpsRequest`.

## Before You Begin

- Ensure KubeDB and Ops-manager are installed.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/restart/ops-request.yaml`.
- Use a separate namespace:

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply Restart OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/restart/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-restart
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-restart
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-restart
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
