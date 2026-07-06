---
title: HanaDBOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-concepts-opsrequest
    name: HanaDBOpsRequest
    parent: guides-hanadb-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDBOpsRequest

## What is HanaDBOpsRequest

`HanaDBOpsRequest` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to
run **day-2 operations** — such as restart, reconfiguration, scaling, volume expansion, storage
migration, TLS management, and credential rotation — against an existing `HanaDB` database. The KubeDB
Ops-manager operator watches `HanaDBOpsRequest` objects and orchestrates the change while keeping the
database available.

## Supported operation types

Every operation type below is implemented by the KubeDB Ops-manager operator for HanaDB:

| `spec.type`          | What it does                                                             | Guide |
|----------------------|-------------------------------------------------------------------------|-------|
| `Restart`            | Rolling restart of the database pods.                                   | [Restart](/docs/guides/hanadb/restart/restart.md) |
| `Reconfigure`        | Apply / change / remove custom `global.ini` configuration.              | [Reconfigure](/docs/guides/hanadb/reconfigure/reconfigure.md) |
| `ReconfigureTLS`     | Add, rotate, or remove TLS certificates.                                | [Reconfigure TLS](/docs/guides/hanadb/tls/overview.md) |
| `VerticalScaling`    | Change CPU/memory of the database (and sidecar) containers.             | [Vertical Scaling](/docs/guides/hanadb/scaling/vertical-scaling/vertical-scaling.md) |
| `VolumeExpansion`    | Grow the data PVCs (requires an expandable StorageClass).               | [Volume Expansion](/docs/guides/hanadb/volume-expansion/volume-expansion.md) |
| `HorizontalScaling`  | Add/remove nodes of a System Replication cluster.                       | — |
| `StorageMigration`   | Move the data volumes to a different StorageClass.                      | [Storage Migration](/docs/guides/hanadb/storage-migration/storage-migration.md) |
| `RotateAuth`         | Rotate the `SYSTEM` password (auto-generated or user-provided).         | [Rotate Authentication](/docs/guides/hanadb/rotate-authentication/rotate-authentication.md) |

## HanaDBOpsRequest Spec

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: hanadb-cluster
  timeout: 30m
  apply: IfReady
```

### spec.databaseRef

`spec.databaseRef.name` is **required** and references the target `HanaDB` object in the same namespace.

### spec.type

`spec.type` is **required** and is one of the supported operation types listed above.

### spec.timeout

`spec.timeout` bounds how long the operation may run before it is marked failed (for example `30m`). It
is **required** for `StorageMigration` and recommended for every storage-heavy or restart-heavy
operation.

### spec.apply

`spec.apply` controls when the operation starts:

- `IfReady` (default) — wait until the database is `Ready` before applying.
- `Always` — apply even if the database is not `Ready` (used by `Restart` to recover an unhealthy
  database).

### Operation-specific fields

Each type reads its own sub-spec:

- `spec.restart` — empty marker for `Restart`.
- `spec.configuration` — `{ configSecret, applyConfig, removeCustomConfig, restart }` for `Reconfigure`.
  `applyConfig` keys must be `global.ini`; the values are **merged** with the existing inline
  configuration. `restart` is `auto` (default), `true`, or `false`.
- `spec.tls` — `{ issuerRef, certificates, rotateCertificates, remove }` for `ReconfigureTLS`.
- `spec.verticalScaling` — `{ hanadb, coordinator, exporter }` resources for `VerticalScaling`.
- `spec.volumeExpansion` — `{ hanadb, mode }` where `mode` is `Online` or `Offline`, for `VolumeExpansion`.
- `spec.horizontalScaling` — `{ replicas }` for `HorizontalScaling` (System Replication only, `>= 2`).
- `spec.migration` — `{ storageClassName, oldPVReclaimPolicy }` for `StorageMigration`.
- `spec.authentication` — `{ secretRef }` for `RotateAuth` (omit to auto-generate a new password).

## HanaDBOpsRequest Status

`status.phase` of a `HanaDBOpsRequest` is one of `Pending`, `Progressing`, `Successful`, `Failed`,
`Skipped`, `WaitingForApproval`, `Approved`, or `Denied`. The `status.conditions` array records each
step of the operation (for example `UpdatePetSets`, `RestartPods`, `Successful`).

## Next Steps

- Run a [Restart](/docs/guides/hanadb/restart/restart.md) or [Reconfigure](/docs/guides/hanadb/reconfigure/reconfigure.md).
- Review the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
