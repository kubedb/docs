---
title: HanaDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-catalog-concepts
    name: HanaDBVersion
    parent: hanadb-concepts-hanadb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDBVersion

## What is HanaDBVersion

`HanaDBVersion` is the catalog CRD that maps a HanaDB version string to the container images and metadata used by KubeDB.

KubeDB resolves `HanaDB.spec.version` through this catalog.

## HanaDBVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: HanaDBVersion
metadata:
  name: "2.0.82"
spec:
  version: "2.0.82"
  db:
    image: "kubedb/hanadb:2.0.82"
  deprecated: false
```

## Key fields

- `metadata.name` is the value used in `HanaDB.spec.version`.
- `spec.version` is the HanaDB engine version.
- `spec.db.image` points to the image used for database pods.
- `spec.deprecated` marks versions that are not recommended for new use.

## Next Steps

- Read the [HanaDB CRD concept](/docs/guides/hanadb/concepts/hanadb.md).
- Run the [HanaDB quickstart](/docs/guides/hanadb/quickstart/quickstart.md).