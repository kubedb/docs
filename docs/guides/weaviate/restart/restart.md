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

This guide tracks restart documentation for Weaviate.

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD, so there is no validated restart manifest to publish yet.

## Before You Begin

- Install KubeDB and Ops-manager.
- Use the example files from `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/restart/ops-request.yaml`.
- Create namespace:

```bash
kubectl create ns demo
```

## Deploy Weaviate

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

No CRD-backed restart manifest can be generated from this repository today.

## Verify

```bash
kubectl describe weaviateopsrequest -n demo weaviate-restart
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-restart
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
