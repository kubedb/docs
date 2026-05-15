---
title: OracleVersion CRD
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-oracleversion
    name: OracleVersion
    parent: guides-oracle-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# OracleVersion

## What is OracleVersion

`OracleVersion` is the catalog CRD that defines Oracle engine version metadata and runtime images used by KubeDB.

When you set `spec.version` in an `Oracle` resource, KubeDB resolves that value using an `OracleVersion` object.

## OracleVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: OracleVersion
metadata:
  name: "21.3.0"
spec:
  version: "21.3.0"
  db:
    image: ghcr.io/kubedb/oracle-ee:21.3.0
  deprecated: false
```

## Key fields

- `metadata.name` is the version key referenced from `Oracle.spec.version`.
- `spec.version` is the Oracle engine version.
- `spec.db.image` is the container image KubeDB uses for Oracle database pods.
- `spec.deprecated` marks whether the version is deprecated for new use.

## Next Steps

- Read the [Oracle CRD concept](/docs/guides/oracle/concepts/oracle.md).
- Follow the [Oracle quickstart](/docs/guides/oracle/quickstart/guide.md).