---
title: OracleOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-opsrequest
    name: OracleOpsRequest
    parent: guides-oracle-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# OracleOpsRequest

## What is OracleOpsRequest

`OracleOpsRequest` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to perform **day-2 operations** (such as restart, reconfigure, scaling, volume expansion and authentication rotation) on a KubeDB managed [Oracle](/docs/guides/oracle/concepts/oracle.md) database in a Kubernetes native way.

## OracleOpsRequest CRD Specification

Like any official Kubernetes resource, an `OracleOpsRequest` has `apiVersion`, `kind`, `metadata`, `spec` and `status` sections. Below are example `OracleOpsRequest` CRs for different operations:

**Restart:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: oracle-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: oracle-sa-sample
  timeout: 30m
  apply: IfReady
```

**Reconfigure:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: oracle-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: oracle-sa-sample
  configuration:
    configSecret:
      name: oracle-custom
```

**Vertical Scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: oracle-vertical-scaling
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: oracle-sa-sample
  verticalScaling:
    node:
      resources:
        requests:
          cpu: "3"
          memory: "10Gi"
        limits:
          cpu: "5"
          memory: "10Gi"
```

**Volume Expansion:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: oracle-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: oracle-sa-sample
  volumeExpansion:
    mode: "Offline"
    node: 12Gi
```

**Rotate Authentication:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: oracle-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: oracle-sa-sample
  apply: IfReady
```

Here, we are going to describe the various sections of an `OracleOpsRequest` CR.

### spec.databaseRef

`spec.databaseRef` is a required field that points to the [Oracle](/docs/guides/oracle/concepts/oracle.md) object on which this operation will be applied. It has the following field:

- `spec.databaseRef.name` — the name of the referenced `Oracle` object. The `OracleOpsRequest` must be created in the same namespace as the referenced `Oracle`.

### spec.type

`spec.type` is a required field that specifies the kind of operation to perform. The supported values are:

- `Restart` — restart all of the database pods (a reconciliation-safe rolling restart).
- `Reconfigure` — apply a new custom configuration to the database.
- `VerticalScaling` — change the CPU and memory resources of the database.
- `VolumeExpansion` — increase the size of the database's persistent volume.
- `RotateAuth` — rotate the database authentication credentials.
- `StorageMigration` — migrate the database's storage to a different storage class.

### spec.configuration

For a `Reconfigure` operation, `spec.configuration` describes the new configuration. It has the fields:

- `spec.configuration.configSecret.name` — a Secret (with key `oracle.cnf`) holding the new configuration.
- `spec.configuration.applyConfig` — a map to provide the configuration inline.
- `spec.configuration.removeCustomConfig` — set to `true` to remove a previously applied custom configuration.

### spec.verticalScaling

For a `VerticalScaling` operation, `spec.verticalScaling.node.resources` specifies the desired `requests`/`limits` of CPU and memory for the database node.

### spec.volumeExpansion

For a `VolumeExpansion` operation:

- `spec.volumeExpansion.mode` — `Online` or `Offline`.
- `spec.volumeExpansion.node` — the desired size of the database node's PVC.
- `spec.volumeExpansion.observer` — (DataGuard) the desired size of the observer's PVC.

### spec.authentication

For a `RotateAuth` operation, you can optionally set `spec.authentication.secretRef.name` to a `kubernetes.io/basic-auth` Secret with your own credentials. If omitted, the operator generates a new random password. (The Oracle `SYS` user cannot be renamed, so only the password is rotated.)

### spec.timeout and spec.apply

- `spec.timeout` — how long the operator waits for each step before timing out.
- `spec.apply` — `IfReady` (default; apply only when the database is ready) or `Always`.

## OracleOpsRequest Status

`status` describes the current state and progress of the operation. The important fields are:

- `status.phase` — the current phase: `Pending`, `Progressing`, `Successful`, `Failed`, `WaitingForApproval`, `Approved`, `Denied`, or `Skipped`.
- `status.observedGeneration` — the most recently observed generation of the request.
- `status.conditions` — an array of conditions describing each step of the operation.

## Next Steps

- Learn about the [Oracle CRD](/docs/guides/oracle/concepts/oracle.md).
- Restart an Oracle database using an [OracleOpsRequest](/docs/guides/oracle/restart/restart.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
