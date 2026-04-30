---
title: Rotate Auth of Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-rotate-auth-cluster
    name: Cluster
    parent: weaviate-rotate-auth
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Auth for Weaviate

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD.

There is no CRD-backed authentication rotation manifest for Weaviate to document yet.

## Current Recommendation

- Manage credential rotation using the auth secret referenced by the `Weaviate` resource.
- Roll out updates using your standard deployment process for the database.

## Check API Status

```bash
kubectl get crd | grep -i weaviateopsrequest
```

If no CRD is returned, do not apply any `kind: WeaviateOpsRequest` manifest from placeholders.