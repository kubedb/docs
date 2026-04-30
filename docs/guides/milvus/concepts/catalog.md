---
title: MilvusVersion CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-catalog-concepts
    name: MilvusVersion
    parent: milvus-concepts-milvus
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MilvusVersion

## What is MilvusVersion

`MilvusVersion` is the catalog CRD that defines the Milvus engine image and related metadata for KubeDB-managed Milvus deployments.

KubeDB uses this CRD when resolving `Milvus.spec.version`.

## MilvusVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MilvusVersion
metadata:
  name: "2.6.11"
spec:
  version: "2.6.11"
  db:
    image: "kubedb/milvus:2.6.11"
  deprecated: false
```

## Key fields

- `metadata.name` is referenced from `Milvus.spec.version`.
- `spec.version` is the Milvus release version.
- `spec.db.image` is the image used for Milvus components.
- `spec.deprecated` indicates whether the entry is deprecated.

## Next Steps

- Read the [Milvus CRD concept](/docs/guides/milvus/concepts/milvus.md).
- Run the [Milvus quickstart](/docs/guides/milvus/quickstart/quickstart.md).