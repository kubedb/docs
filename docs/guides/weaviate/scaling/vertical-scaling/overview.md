---
title: Weaviate Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-vertical-scaling-overview
    name: Overview
    parent: weaviate-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Vertical Scaling

This guide tracks vertical scaling documentation for Weaviate.

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD, so there is no validated vertical scaling manifest to publish yet.

## Before You Begin

- Ensure database is healthy and all pods are running.
- Use the example files from `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/scaling/vertical-scaling/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Weaviate

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

See the detailed note in [Vertical Scaling for Weaviate](/docs/guides/weaviate/scaling/vertical-scaling/scale-vertically/).

## Verify

```bash
kubectl describe weaviateopsrequest -n demo weaviate-vertical-scale
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-vertical-scale
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
