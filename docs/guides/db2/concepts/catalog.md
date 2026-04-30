---
title: DB2Version CRD
menu:
  docs_{{ .version }}:
    identifier: db2-catalog-concepts
    name: DB2Version
    parent: db2-concepts-db2
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DB2Version

## What is DB2Version

`DB2Version` is the catalog CRD that defines which DB2 engine image and related components KubeDB should use when creating a `DB2` database.

When you set `spec.version` in a `DB2` resource, KubeDB resolves that value against a `DB2Version` object.

## DB2Version Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DB2Version
metadata:
  name: "11.5.8.0"
spec:
  version: "11.5.8.0"
  db:
    image: "kubedb/db2:11.5.8.0"
  deprecated: false
```

## Key fields

- `metadata.name` is the version label you reference from `DB2.spec.version`.
- `spec.version` is the upstream DB2 version represented by this entry.
- `spec.db.image` is the container image KubeDB uses for DB2 pods.
- `spec.deprecated` marks whether this entry should be avoided for new deployments.

## Next Steps

- Read the [DB2 CRD concept](/docs/guides/db2/concepts/db2.md).
- Run the [DB2 quickstart](/docs/guides/db2/quickstart/quickstart.md).