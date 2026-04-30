---
title: Expanding Weaviate Storage
menu:
  docs_{{ .version }}:
    identifier: weaviate-volume-expansion-overview
    name: Overview
    parent: weaviate-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Volume Expansion for Weaviate

This guide tracks volume expansion documentation for Weaviate.

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD, so there is no validated volume expansion manifest to publish yet.

## Before You Begin

- Ensure StorageClass supports volume expansion.
- Use the example files from `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/volume-expansion/ops-request.yaml`.

```bash
kubectl create ns demo
kubectl get storageclass
```

## Deploy Weaviate

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

See the detailed note in [Expand Weaviate Volume](/docs/guides/weaviate/volume-expansion/volume-expansion.md).

## Verify

```bash
kubectl describe weaviateopsrequest -n demo weaviate-volume-expand
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-volume-expand
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
