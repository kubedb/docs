---
title: HanaDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-catalog-concepts
    name: HanaDBVersion
    parent: hanadb-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# HanaDBVersion

## What is HanaDBVersion?

`HanaDBVersion` is the catalog custom resource that maps a HanaDB version string to the container images and metadata used by KubeDB.

KubeDB resolves `HanaDB.spec.version` through this catalog.

## HanaDBVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: HanaDBVersion
metadata:
  name: "2.0.82"
spec:
  coordinator:
    image: ghcr.io/kubedb/hanadb-coordinator:v0.4.0
  db:
    image: docker.io/saplabs/hanaexpress:2.00.082.00.20250528.1
  exporter:
    image: ghcr.io/kubedb/hanadb-exporter:1.0.0
  securityContext:
    runAsGroup: 79
    runAsUser: 12000
  updateConstraints:
    allowlist:
    - 2.0.82
  version: "2.0.82"
```

## Key fields

- `metadata.name` is the value used in `HanaDB.spec.version`.
- `spec.version` is the HanaDB engine version.
- `spec.coordinator.image` points to the coordinator sidecar image.
- `spec.db.image` points to the image used for database pods.
- `spec.exporter.image` points to the metrics exporter image.
- `spec.securityContext` provides the default user and group used by database containers.
- `spec.updateConstraints.allowlist` lists the versions this catalog entry can update to.
- `spec.deprecated` marks versions that are not recommended for new use.

## Next Steps

- Read the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
- Run the [HanaDB quickstart](/docs/guides/hanadb/quickstart/quickstart.md).
