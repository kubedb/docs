---
title: QdrantVersion CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-catalog-concepts
    name: QdrantVersion
    parent: qdrant-concepts-qdrant
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# QdrantVersion

## What is QdrantVersion

`QdrantVersion` is the catalog CRD that defines image and release metadata for KubeDB-managed Qdrant clusters.

KubeDB resolves `Qdrant.spec.version` using this catalog entry.

## QdrantVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: QdrantVersion
metadata:
  name: "1.17.0"
spec:
  version: "1.17.0"
  db:
    image: "kubedb/qdrant:1.17.0"
  deprecated: false
```

## Key fields

- `metadata.name` is referenced by `Qdrant.spec.version`.
- `spec.version` identifies the engine release.
- `spec.db.image` is the image used by Qdrant pods.
- `spec.deprecated` marks unsupported or legacy versions.

## Next Steps

- Read the [Qdrant CRD concept](/docs/guides/qdrant/concepts/qdrant.md).
- Run the [Qdrant quickstart](/docs/guides/qdrant/quickstart/quickstart.md).