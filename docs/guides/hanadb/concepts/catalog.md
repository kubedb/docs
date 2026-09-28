---
title: HanaDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-concepts-catalog
    name: HanaDBVersion
    parent: guides-hanadb-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDBVersion

## What is HanaDBVersion

`HanaDBVersion` is a Kubernetes `Custom Resource Definition` (CRD). It is a cluster-scoped catalog object
that holds the container images KubeDB uses for a particular SAP HANA release. When you set
`spec.version` on a `HanaDB` object, KubeDB looks up the matching `HanaDBVersion` to pick the database,
coordinator, and exporter images.

A separate catalog object per version lets cluster administrators control exactly which images are used,
deprecate old versions, and constrain version upgrades.

## HanaDBVersion Spec

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: HanaDBVersion
metadata:
  name: 2.0.82
spec:
  version: 2.0.82
  db:
    image: docker.io/saplabs/hanaexpress:2.00.082.00.20250528.1
  coordinator:
    image: ghcr.io/kubedb/hanadb-coordinator:v0.5.0
  exporter:
    image: ghcr.io/kubedb/hanadb-exporter:1.0.0
  securityContext:
    runAsUser: 12000
    runAsGroup: 79
  updateConstraints:
    allowlist:
    - '>= 2.0.82, <= 2.0.88'
```

### spec.version

`spec.version` is the SAP HANA database version string. It is the value users set in `HanaDB.spec.version`.

### spec.db.image

`spec.db.image` is the SAP HANA, express edition container image. This is a **required** field. KubeDB
does not ship HANA images; the catalog points at images you are licensed to use.

### spec.coordinator.image

`spec.coordinator.image` is the image for the `hanadb-coordinator` raft sidecar that runs alongside each
node of a `SystemReplication` cluster. It elects the primary and maintains the `kubedb.com/role` labels.

### spec.exporter.image

`spec.exporter.image` is the [hanadb_exporter](https://github.com/kubedb/hanadb-exporter) image used for
Prometheus metrics. This is a **required** field.

### spec.securityContext

`spec.securityContext.runAsUser` and `spec.securityContext.runAsGroup` are the UID/GID the database
container runs as (the HANA `hxeadm` user, `12000:79`). KubeDB uses these to default the pod security
context and the data volume `fsGroup`.

### spec.deprecated

`spec.deprecated` is an optional boolean. When `true`, the operator rejects new `HanaDB` objects that
reference this version and surfaces a warning. Existing databases keep running.

### spec.updateConstraints

`spec.updateConstraints.allowlist` / `denylist` constrain which versions a database on this version may
be upgraded to.

## List available versions

```bash
kubectl get hanadbversions
```

## Next Steps

- Learn about the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
- Deploy your first database with the [Quickstart](/docs/guides/hanadb/quickstart/quickstart.md).
