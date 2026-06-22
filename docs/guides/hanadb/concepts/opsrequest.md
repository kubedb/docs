---
title: HanaDBOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: hanadb-concepts-opsrequest
    name: HanaDBOpsRequest
    parent: hanadb-concepts
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDBOpsRequest

`HanaDBOpsRequest` is the KubeDB custom resource used to run day-2 operations on a `HanaDB` database.

## Supported Operations

The following operations are supported:

- `Restart`
- `VerticalScaling`
- `ReconfigureTLS`
- `RotateAuth`
- `VolumeExpansion`
- `StorageMigration`

Create one `HanaDBOpsRequest` for one operation. Wait for the request to become `Successful` before creating the next request for the same database.

## Spec

### spec.databaseRef

`spec.databaseRef.name` specifies the `HanaDB` object that the operation applies to. The `HanaDBOpsRequest` and the `HanaDB` must be in the same namespace.

### spec.type

`spec.type` specifies the operation type.

### spec.verticalScaling

`spec.verticalScaling.hanadb` updates CPU and memory resources for the main `hanadb` container.

```yaml
verticalScaling:
  hanadb:
    resources:
      requests:
        cpu: "2100m"
        memory: "8448Mi"
      limits:
        cpu: "4"
        memory: "14Gi"
```

### spec.tls

`spec.tls` configures TLS for `ReconfigureTLS` operations. You can add TLS using an issuer reference, rotate existing certificates, or remove TLS.

```yaml
tls:
  issuerRef:
    apiGroup: cert-manager.io
    kind: Issuer
    name: hdb-ca-issuer
```

### spec.authentication

`spec.authentication.secretRef` specifies a user-provided `kubernetes.io/basic-auth` Secret for `RotateAuth`.

### spec.volumeExpansion

`spec.volumeExpansion.hanadb` specifies the target PVC size. `spec.volumeExpansion.mode` can be `Online` or `Offline`.

```yaml
volumeExpansion:
  mode: Online
  hanadb: 65Gi
```

### spec.migration

`spec.migration.storageClassName` specifies the target `StorageClass` for `StorageMigration`.

```yaml
migration:
  storageClassName: longhorn-single-migrated
```

### spec.timeout

`spec.timeout` controls how long KubeDB waits for operation steps before marking the request failed.

### spec.apply

`spec.apply` controls when KubeDB applies the operation.

- `IfReady`: run only when the database is Ready.
- `Always`: run regardless of the current database phase.

## Status

The operation result is reported in `.status.phase`.

| Phase | Meaning |
|-------|---------|
| `Progressing` | KubeDB has started the operation. |
| `Successful` | KubeDB completed the operation. |
| `Failed` | KubeDB failed the operation. |
| `Denied` | KubeDB denied the operation during validation. |
| `Skipped` | KubeDB skipped the operation because a newer matching request superseded it. |
