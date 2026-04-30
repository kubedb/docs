---
title: Updating Qdrant Version
menu:
  docs_{{ .version }}:
    identifier: qdrant-update-version-overview
    name: Overview
    parent: qdrant-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Qdrant Version

This guide shows how to update Qdrant to a supported target version.

## Before You Begin

- Ensure Qdrant is `Ready` before submitting the update request.
- Use the example files from `docs/examples/qdrant/quickstart/distributed.yaml` and `docs/examples/qdrant/update-version/ops-request.yaml`.

```bash
kubectl create ns demo
kubectl get qdrantversions
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply UpdateVersion OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/update-version/ops-request.yaml
kubectl get qdrantopsrequest -n demo qdrant-update-version
```

## Verify

```bash
kubectl describe qdrantopsrequest -n demo qdrant-update-version
```

## Cleaning up

```bash
kubectl delete qdrantopsrequest -n demo qdrant-update-version
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
