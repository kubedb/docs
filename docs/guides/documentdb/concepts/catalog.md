---
title: DocumentDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: documentdb-catalog-concepts
    name: DocumentDBVersion
    parent: documentdb-concepts-documentdb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DocumentDBVersion

## What is DocumentDBVersion

`DocumentDBVersion` is the catalog CRD that defines which DocumentDB engine image and related components KubeDB should use.

The value in `DocumentDB.spec.version` must match an available `DocumentDBVersion` resource.

## DocumentDBVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DocumentDBVersion
metadata:
  name: "pg17-0.109.0"
spec:
  version: "pg17-0.109.0"
  db:
    image: "kubedb/documentdb:pg17-0.109.0"
  deprecated: false
```

## Key fields

- `metadata.name` is referenced from `DocumentDB.spec.version`.
- `spec.version` is the released database engine version.
- `spec.db.image` is the runtime image used by KubeDB.
- `spec.deprecated` indicates if this catalog entry is deprecated.

## Next Steps

- Read the [DocumentDB CRD concept](/docs/guides/documentdb/concepts/documentdb.md).
- Run the [DocumentDB quickstart](/docs/guides/documentdb/quickstart/quickstart.md).