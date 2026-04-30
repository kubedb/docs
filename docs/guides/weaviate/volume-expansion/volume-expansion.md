---
title: Expand Weaviate Volume
menu:
  docs_{{ .version }}:
    identifier: weaviate-volume-expansion-cluster
    name: Cluster
    parent: weaviate-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand Weaviate Volume

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD.

Until that API exists, a correct `VolumeExpansion` manifest for Weaviate cannot be generated from this repo.

## Current Recommendation

- Expand PVC size through the storage workflow supported by your `StorageClass`.
- Update the `Weaviate` resource storage specification as needed.

## Check API Status

```bash
kubectl get crd | grep -i weaviateopsrequest
```

If this CRD appears in your installed release, use that release documentation for official volume expansion request manifests.