---
title: Expanding Qdrant Storage
menu:
  docs_{{ .version }}:
    identifier: qdrant-volume-expansion-overview
    name: Overview
    parent: qdrant-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Volume Expansion for Qdrant

This guide shows how to increase PVC size of Qdrant data volumes.

## Before You Begin

- Ensure your StorageClass supports volume expansion.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/volume-expansion/ops-request.yaml`.

```bash
kubectl create ns demo
kubectl get storageclass
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply VolumeExpansion OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/volume-expansion/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-volume-expand
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-volume-expand
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-volume-expand
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
