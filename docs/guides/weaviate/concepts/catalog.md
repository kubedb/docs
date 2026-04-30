---
title: WeaviateVersion CRD
menu:
  docs_{{ .version }}:
    identifier: weaviate-catalog-concepts
    name: WeaviateVersion
    parent: weaviate-concepts-weaviate
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# WeaviateVersion

## What is WeaviateVersion

`WeaviateVersion` is the catalog CRD that defines image and release metadata for KubeDB-managed Weaviate clusters.

The value in `Weaviate.spec.version` is resolved against `WeaviateVersion` objects.

## WeaviateVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: WeaviateVersion
metadata:
  name: "1.33.1"
spec:
  version: "1.33.1"
  db:
    image: "kubedb/weaviate:1.33.1"
  deprecated: false
```

## Key fields

- `metadata.name` is used from `Weaviate.spec.version`.
- `spec.version` identifies the Weaviate release.
- `spec.db.image` is the runtime image for Weaviate pods.
- `spec.deprecated` indicates deprecation status.

## Next Steps

- Read the [Weaviate CRD concept](/docs/guides/weaviate/concepts/weaviate.md).
- Run the [Weaviate quickstart](/docs/guides/weaviate/quickstart/quickstart.md).