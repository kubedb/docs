---
title: Reconfiguring Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-reconfigure-overview
    name: Overview
    parent: qdrant-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Qdrant

This guide shows how to update runtime configuration of Qdrant using `QdrantOpsRequest`.

## Before You Begin

- Ensure KubeDB and Ops-manager are installed.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/reconfigure/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply Reconfigure OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-reconfigure
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-reconfigure
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-reconfigure
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
