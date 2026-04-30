---
title: Updating Weaviate Version
menu:
  docs_{{ .version }}:
    identifier: weaviate-update-version-overview
    name: Overview
    parent: weaviate-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Weaviate Version

This guide tracks version update documentation for Weaviate.

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD, so there is no validated update-version manifest to publish yet.

## Before You Begin

- Ensure Weaviate is `Ready` before submitting the update request.
- Use the example files from `docs/examples/weaviate/quickstart/weaviate.yaml` and `docs/examples/weaviate/update-version/ops-request.yaml`.

```bash
kubectl create ns demo
kubectl get weaviateversions
```

## Deploy Weaviate

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate.yaml
kubectl get weaviate -n demo weaviate-sample -w
```

See the detailed note in [Upgrade Weaviate Version](/docs/guides/weaviate/update-version/versionupgrading/).

## Verify

```bash
kubectl describe weaviateopsrequest -n demo weaviate-update-version
```

## Cleaning up

```bash
kubectl delete weaviateopsrequest -n demo weaviate-update-version
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
