---
title: QdrantOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-opsrequest-concepts
    name: QdrantOpsRequest
    parent: qdrant-concepts-qdrant
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# QdrantOpsRequest

## What is QdrantOpsRequest

`QdrantOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Qdrant](https://qdrant.tech/) administrative operations like database version updating, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## QdrantOpsRequest CRD Specifications

Like any official Kubernetes resource, a `QdrantOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here are some sample `QdrantOpsRequest` CRs for different administrative operations:

**Sample `QdrantOpsRequest` for updating database version:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: qdrant-sample
  updateVersion:
    targetVersion: "1.18.0"
status:
  conditions:
  - lastTransitionTime: "2024-10-01T10:00:00Z"
    message: The controller has updated the Qdrant successfully
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  phase: Successful
```

**Sample `QdrantOpsRequest` for horizontal scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: qdrant-sample
  horizontalScaling:
    node: 5
status:
  conditions:
  - lastTransitionTime: "2024-10-01T10:00:00Z"
    message: The controller has scaled/updated the Qdrant successfully
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  phase: Successful
```

**Sample `QdrantOpsRequest` for vertical scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: qdrant-sample
  verticalScaling:
    node:
      resources:
        requests:
          memory: "1Gi"
          cpu: "500m"
        limits:
          memory: "1Gi"
          cpu: "500m"
status:
  conditions:
  - lastTransitionTime: "2024-10-01T10:00:00Z"
    message: The controller has scaled/updated the Qdrant successfully
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  phase: Successful
```

### QdrantOpsRequest `Spec`

A `QdrantOpsRequest` object has the following fields in the `spec` section:

#### spec.databaseRef

`spec.databaseRef` is a required field that points to the [Qdrant](/docs/guides/qdrant/concepts/qdrant.md) object where the administrative operations will be applied. It contains:

- `spec.databaseRef.name` — the name of the target Qdrant database (required).

#### spec.type

`spec.type` specifies the type of operation that will be applied to the database. Supported operations are:

- `Reconfigure` — reconfigure a running Qdrant database with new configuration.
- `ReconfigureTLS` — reconfigure TLS configuration for a running Qdrant database.
- `Restart` — restart the database pods in a rolling fashion.
- `RotateAuth` — rotate the authentication credentials of a running Qdrant database.
- `UpdateVersion` — update the version of a running Qdrant database.
- `HorizontalScaling` — scale the number of nodes up or down.
- `VerticalScaling` — vertically scale the resources (CPU and memory) of database pods.
- `VolumeExpansion` — expand the persistent volume claim size of a running Qdrant database.

#### spec.updateVersion

`spec.updateVersion` is used when `spec.type` is `UpdateVersion`. It contains:

- `spec.updateVersion.targetVersion` — the target `QdrantVersion` to update to.

#### spec.horizontalScaling

`spec.horizontalScaling` is used when `spec.type` is `HorizontalScaling`. It contains:

- `spec.horizontalScaling.node` — the desired number of Qdrant nodes.

#### spec.verticalScaling

`spec.verticalScaling` is used when `spec.type` is `VerticalScaling`. It contains:

- `spec.verticalScaling.node.resources` — the CPU and memory resource requests and limits for Qdrant nodes.

#### spec.volumeExpansion

`spec.volumeExpansion` is used when `spec.type` is `VolumeExpansion`. It contains:

- `spec.volumeExpansion.node` — the new desired storage size for Qdrant nodes.
- `spec.volumeExpansion.mode` — the volume expansion mode. Can be `Online` or `Offline`.

#### spec.timeout

`spec.timeout` is an optional field that specifies the timeout duration for the OpsRequest to complete. If the OpsRequest does not complete within the specified timeout, it will be marked as failed. The value is in the form of a Kubernetes duration (e.g., `5m`, `1h`).

#### spec.apply

`spec.apply` is an optional field that specifies when the OpsRequest will be applied. Possible values are `Always` and `IfReady`. The default is `IfReady`, which means the OpsRequest will only be applied when the target database is in `Ready` state.

## Next Steps

- See [Qdrant ops request overview](/docs/guides/qdrant/ops-request/overview.md) for operation links.
- Follow operation tutorials like [Restart](/docs/guides/qdrant/restart/restart.md) and [Volume Expansion](/docs/guides/qdrant/volume-expansion/volume-expansion.md).