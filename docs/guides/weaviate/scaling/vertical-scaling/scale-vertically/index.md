---
title: Scale Weaviate Vertically
menu:
  docs_{{ .version }}:
    identifier: weaviate-scale-vertically
    name: Scale Vertically
    parent: weaviate-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling for Weaviate

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD.

That means there is no validated vertical scaling manifest to publish for Weaviate at this time.

## Current Recommendation

Use direct `Weaviate` spec updates for resource requests and limits in `spec.podTemplate` while following maintenance best practices.

## Check API Status

```bash
kubectl get crd | grep -i weaviateopsrequest
```

If this CRD becomes available in a newer release, use that release documentation for vertical scaling request examples.