---
title: MilvusOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-concepts-milvusopsrequest
    name: MilvusOpsRequest
    parent: milvus-concepts
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MilvusOpsRequest

## What is MilvusOpsRequest

`MilvusOpsRequest` is a Kubernetes `CustomResourceDefinition` (CRD). It provides declarative configuration for administrative operations on a [Milvus](/docs/guides/milvus/concepts/milvus.md) database such as version updates, scaling, restarts, TLS changes, authentication rotation, and storage operations.

## Sample MilvusOpsRequest Objects

### UpdateVersion

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: milvus-cluster
  updateVersion:
    targetVersion: 2.6.11
  timeout: 5m
  apply: IfReady
```

### VerticalScaling

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: vertical-scaling
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: milvus-cluster
  verticalScaling:
    mixcoord:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
    proxy:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

### HorizontalScaling

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: milvus-cluster
  horizontalScaling:
    topology:
      proxy: 2
      streamingnode: 2
```

### VolumeExpansion

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: volume-expansion-online
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: milvus-cluster
  volumeExpansion:
    streamingnode: 4Gi
    mode: Online
```

### Restart

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: milvus-cluster
  timeout: 5m
  apply: Always
```

### Reconfigure

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: reconfigure-1
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: milvus-cluster
  configuration:
    removeCustomConfig: true
    configSecret:
      name: mv-configuration
    applyConfig:
      milvus.yaml: |
        log:
          level: info
          file:
            maxAge: 30
    restart: "false"
  timeout: 5m
  apply: IfReady
```

### ReconfigureTLS

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: mvops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: milvus-cluster
  tls:
    rotateCertificates: true
```

### RotateAuth

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: milvus-rotate-auth-user-secret
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: milvus-cluster
  authentication:
    secretRef:
      kind: Secret
      name: milvus-new-auth1
```

### StorageMigration

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: milvus-cluster
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 10m
```

## MilvusOpsRequest Spec

### spec.databaseRef

`spec.databaseRef` is required. It points to the [Milvus](/docs/guides/milvus/concepts/milvus.md) object the operation will run against.

### spec.type

`spec.type` selects the operation KubeDB should perform.

The Milvus guides in this repository use these values:

- `UpdateVersion`
- `VerticalScaling`
- `HorizontalScaling`
- `VolumeExpansion`
- `StorageMigration`
- `Restart`
- `Reconfigure`
- `ReconfigureTLS`
- `RotateAuth`

Use only one operation type per `MilvusOpsRequest`.

### spec.updateVersion

Used when `spec.type: UpdateVersion`.

```yaml
spec:
  updateVersion:
    targetVersion: 2.6.11
```

`targetVersion` must name an existing `MilvusVersion`.

### spec.verticalScaling

Used when `spec.type: VerticalScaling`.

The keys depend on topology:

- Standalone: `node`
- Distributed: `proxy`, `mixcoord`, `datanode`, `querynode`, `streamingnode`

Each block carries the new Kubernetes container resources:

```yaml
resources:
  requests:
    cpu: "1"
    memory: "2Gi"
  limits:
    cpu: "1"
    memory: "2Gi"
```

- `spec.verticalScaling.mode` specifies how the scaling is actuated. `Restart` (the default) applies the new resources by restarting the Pods, while `InPlace` resizes the running Pods in place via the Kubernetes `pods/resize` subresource (no restart), automatically falling back to `Restart` for any Pod whose Node cannot fit the new resources. Optional; defaults to `Restart`.

### spec.horizontalScaling

Used when `spec.type: HorizontalScaling`.

Horizontal scaling is distributed-only. The requested replica counts are set under `spec.horizontalScaling.topology`:

```yaml
spec:
  horizontalScaling:
    topology:
      proxy: 2
      mixcoord: 2
      datanode: 2
      querynode: 2
      streamingnode: 2
```

### spec.volumeExpansion

Used when `spec.type: VolumeExpansion`.

```yaml
spec:
  volumeExpansion:
    mode: Online
    node: 4Gi
```

The target key depends on topology:

- Standalone: `node`
- Distributed: `streamingnode`

Only `streamingnode` is valid for distributed Milvus because it is the only distributed role with persistent Milvus storage.

### spec.migration

Used when `spec.type: StorageMigration`.

```yaml
spec:
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
```

This operation migrates the Milvus PVCs to a different `StorageClass`.

### spec.configuration

Used when `spec.type: Reconfigure`.

The Milvus reconfigure flow supports:

- `configSecret` - reference to a secret whose `milvus.yaml` key holds configuration.
- `applyConfig` - inline configuration to merge under the `milvus.yaml` key.
- `removeCustomConfig` - remove previously applied custom configuration before applying the new one.
- `restart` - whether pods should be restarted after the config change.

### spec.tls

Used when `spec.type: ReconfigureTLS`.

Depending on the fields you set, this can:

- add TLS,
- rotate certificates,
- switch to a new issuer,
- remove TLS.

The current guides use these patterns:

- `rotateCertificates: true`
- `remove: true`
- `issuerRef` plus `external` and `internal`

See [Reconfigure TLS Overview](/docs/guides/milvus/reconfigure-tls/overview.md).

### spec.authentication

Used when `spec.type: RotateAuth`.

If you omit `spec.authentication`, KubeDB generates a new password for the existing auth secret. If you set `spec.authentication.secretRef`, KubeDB switches the database to the user-provided secret.

The referenced secret should contain `username` and `password`.

### spec.timeout

`spec.timeout` sets how long the operator should allow each step of the operation to run before timing out.

### spec.apply

`spec.apply` controls when the ops request should be executed, for example `IfReady` or `Always`.

## Status

KubeDB updates `.status` as the operation progresses.

The fields you will commonly observe are:

- `status.phase` - overall state such as `Progressing`, `Successful`, or `Failed`.
- `status.conditions` - step-by-step condition entries recorded by the operator.
- `status.observedGeneration` - latest generation seen by the controller.

## Related Concepts

- [Milvus](/docs/guides/milvus/concepts/milvus.md)
- [MilvusAutoscaler](/docs/guides/milvus/concepts/milvusautoscaler.md)
