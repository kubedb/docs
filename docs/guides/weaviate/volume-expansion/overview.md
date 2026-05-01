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

Expand Weaviate persistent volume size using `WeaviateOpsRequest` with `type: VolumeExpansion`.

## Before You Begin

- Ensure the selected StorageClass supports volume expansion.
- Ensure the database is healthy before applying the request.
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

Continue with [Expand Weaviate Volume](/docs/guides/weaviate/volume-expansion/volume-expansion.md).

## Verify

```bash
kubectl describe weaviateopsrequest -n demo weaviate-volume-expand
kubectl get pvc -n demo
kubectl get weaviate -n demo weaviate-sample
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-volume-expand
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
