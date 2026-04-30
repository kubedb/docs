---
title: Qdrant Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-vertical-scaling-overview
    name: Overview
    parent: qdrant-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Vertical Scaling

This guide shows how to update CPU and memory resources of Qdrant nodes.

## Before You Begin

- Ensure database is healthy and all pods are running.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/scaling/vertical-scaling/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply VerticalScaling OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/scaling/vertical-scaling/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-vertical-scale
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-vertical-scale
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-vertical-scale
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
