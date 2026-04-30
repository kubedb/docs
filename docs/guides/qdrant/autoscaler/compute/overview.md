---
title: Qdrant Compute Autoscaler Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-compute-overview
    name: Overview
    parent: qdrant-autoscaler-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Qdrant Compute Autoscaler

This guide shows how to configure compute autoscaling for Qdrant.

## Before You Begin

- Install KubeDB Autoscaler operator.
- Install metrics-server in your cluster.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/autoscaler/compute/autoscaler.yaml`.

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply QdrantAutoscaler

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/autoscaler/compute/autoscaler.yaml
```

## Verify

```bash
kubectl get qdrantautoscaler -n demo qdrant-as-compute
kubectl describe qdrantautoscaler -n demo qdrant-as-compute
```

## Cleaning up

```bash
kubectl delete qdrantautoscaler -n demo qdrant-as-compute
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
