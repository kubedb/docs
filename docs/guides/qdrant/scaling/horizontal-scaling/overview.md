---
title: Qdrant Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-horizontal-scaling-overview
    name: Overview
    parent: qdrant-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Horizontal Scaling

This guide shows how to scale Qdrant nodes horizontally.

## Before You Begin

- Ensure database is healthy (`status.phase=Ready`).
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/scaling/horizontal-scaling/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply HorizontalScaling OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/scaling/horizontal-scaling/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-horizontal-scale
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-horizontal-scale
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-horizontal-scale
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
