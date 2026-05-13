---
title: Qdrant Storage Autoscaler Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-storage-overview
    name: Overview
    parent: qdrant-autoscaler-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant Storage Autoscaler

This guide shows how to configure storage autoscaling for Qdrant.

## Before You Begin

- StorageClass with `allowVolumeExpansion: true`.
- KubeDB Autoscaler operator installed.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/autoscaler/storage/autoscaler.yaml`.

```bash
kubectl create ns demo
kubectl get storageclass
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply QdrantAutoscaler

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/autoscaler/storage/autoscaler.yaml
```

## Verify

```bash
kubectl get qdrantautoscaler -n demo qdrant-as-storage
kubectl describe qdrantautoscaler -n demo qdrant-as-storage
```

## Cleaning up

```bash
kubectl delete qdrantautoscaler -n demo qdrant-as-storage
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
