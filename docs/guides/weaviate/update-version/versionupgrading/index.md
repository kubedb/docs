---
title: Upgrade Weaviate Version
menu:
  docs_{{ .version }}:
    identifier: weaviate-version-upgrading
    name: Version Upgrading
    parent: weaviate-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Upgrade Weaviate Version

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD.

Until that API exists, a correct `UpdateVersion` manifest for Weaviate cannot be generated from this repo.

## Current Recommendation

Upgrade Weaviate by updating `spec.version` in the `Weaviate` custom resource to a supported `WeaviateVersion`.

```bash
kubectl get weaviateversions
```

## Check API Status

```bash
kubectl get crd | grep -i weaviateopsrequest
```