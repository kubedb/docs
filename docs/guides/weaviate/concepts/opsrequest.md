---
title: WeaviateOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: weaviate-concepts-opsrequest
    name: WeaviateOpsRequest
    parent: weaviate-concepts
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# WeaviateOpsRequest

## What is WeaviateOpsRequest

`WeaviateOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Weaviate](https://weaviate.io/) administrative operations like horizontal scaling, vertical scaling, volume expansion, reconfiguration, etc. in a Kubernetes native way.

## WeaviateOpsRequest CRD Specifications

Like any official Kubernetes resource, a `WeaviateOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here are some sample `WeaviateOpsRequest` CRs for different administrative operations:

**Sample `WeaviateOpsRequest` for horizontal scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: weaviate-sample
  horizontalScaling:
    node: 5
status:
  conditions:
  - lastTransitionTime: "2024-10-01T10:00:00Z"
    message: The controller has scaled/updated the Weaviate successfully
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  phase: Successful
```

**Sample `WeaviateOpsRequest` for vertical scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: weaviate-sample
  verticalScaling:
    mode: InPlace
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
    message: The controller has scaled/updated the Weaviate successfully
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  phase: Successful
```

**Sample `WeaviateOpsRequest` for volume expansion:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: weaviate-sample
  volumeExpansion:
    mode: Online
    node: 10Gi
status:
  conditions:
  - lastTransitionTime: "2024-10-01T10:00:00Z"
    message: The controller has expanded the volume of Weaviate successfully
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  phase: Successful
```

### WeaviateOpsRequest `Spec`

A `WeaviateOpsRequest` object has the following fields in the `spec` section:

#### spec.databaseRef

`spec.databaseRef` is a required field that points to the [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) object where the administrative operations will be applied. It contains:

- `spec.databaseRef.name` — the name of the target Weaviate database (required).

#### spec.type

`spec.type` is a required field that specifies the type of operation that will be applied to the database. Supported operations are:

- `HorizontalScaling` — scale the number of Weaviate nodes up or down.
- `VerticalScaling` — vertically scale the resources (CPU and memory) of the Weaviate node Pods.
- `VolumeExpansion` — expand the persistent volume claim size of a running Weaviate database.
- `Restart` — restart the database Pods in a rolling fashion.
- `Reconfigure` — reconfigure a running Weaviate database with new configuration.
- `ReconfigureTLS` — reconfigure the TLS configuration for a running Weaviate database.
- `RotateAuth` — rotate the authentication credentials of a running Weaviate database.
- `StorageMigration` — migrate the storage class or data of a running Weaviate database.

#### spec.horizontalScaling

`spec.horizontalScaling` is used when `spec.type` is `HorizontalScaling`. It contains:

- `spec.horizontalScaling.node` — the desired number of Weaviate nodes.

#### spec.verticalScaling

`spec.verticalScaling` is used when `spec.type` is `VerticalScaling`. It contains:

- `spec.verticalScaling.node` — the resource requirements (`PodResources`) for the Weaviate node Pods, i.e. the CPU and memory resource requests and limits.
- `spec.verticalScaling.mode` specifies how the scaling is actuated. `Restart` (the default) applies the new resources by restarting the Pods, while `InPlace` resizes the running Pods in place via the Kubernetes `pods/resize` subresource (no restart), automatically falling back to `Restart` for any Pod whose Node cannot fit the new resources. Optional; defaults to `Restart`.

#### spec.volumeExpansion

`spec.volumeExpansion` is used when `spec.type` is `VolumeExpansion`. It contains:

- `spec.volumeExpansion.mode` — the volume expansion mode. Can be `Online` or `Offline`.
- `spec.volumeExpansion.node` — the desired size (a Kubernetes resource quantity, e.g. `10Gi`) of the persistent volume for the Weaviate node Pods.

#### spec.restart

`spec.restart` is used when `spec.type` is `Restart`. It is an empty object (`{}`). No further configuration is needed for a restart operation.

#### spec.configuration

`spec.configuration` is used when `spec.type` is `Reconfigure`. It contains:

- `spec.configuration.applyConfig` — a map of key-value pairs for inline configuration changes.
- `spec.configuration.configSecret` — the secret containing the new configuration.
- `spec.configuration.removeCustomConfig` — specifies whether to remove the custom configuration.
- `spec.configuration.backupConfigSecret` — an optional reference to a Kubernetes Secret providing environment variables for the database container.

#### spec.tls

`spec.tls` is used when `spec.type` is `ReconfigureTLS`. It contains:

- `spec.tls.issuerRef` — a reference to the `Issuer` or `ClusterIssuer` used to generate the TLS certificates.
- `spec.tls.certificates` — a list of certificate specifications for configuring TLS.
- `spec.tls.rotateCertificates` — specifies whether to rotate the TLS certificates.
- `spec.tls.remove` — specifies whether to remove the TLS configuration.

#### spec.authentication

`spec.authentication` is used when `spec.type` is `RotateAuth`. It contains:

- `spec.authentication.secretRef` — a reference to the secret containing the new authentication credentials:
  - `apiGroup` — the API group of the referenced secret.
  - `kind` — the kind of the referenced secret.
  - `name` — the name of the secret (required).

#### spec.migration

`spec.migration` is used when `spec.type` is `StorageMigration`. It contains:

- `spec.migration.storageClassName` — the target storage class name for migration.
- `spec.migration.oldPVReclaimPolicy` — the reclaim policy for the old PersistentVolume.

#### spec.timeout

`spec.timeout` is an optional field that specifies the timeout duration for each step of the OpsRequest to complete. If a step does not complete within the specified timeout, the OpsRequest will be marked as failed. The value is in the form of a Kubernetes duration (e.g., `5m`, `1h`).

#### spec.apply

`spec.apply` is an optional field that specifies when the OpsRequest will be applied. Possible values are `Always` and `IfReady`. The default is `IfReady`, which means the OpsRequest will only be applied when the target database is in `Ready` state.

#### spec.maxRetries

`spec.maxRetries` is an optional `<integer>` field that specifies the maximum number of times the OpsRequest should be retried if it fails. It defaults to `1`.

### WeaviateOpsRequest `Status`

`.status` describes the current state and progress of the `WeaviateOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `WeaviateOpsRequest`. It can have the following values:

| Phase              | Meaning                                                                            |
|--------------------|------------------------------------------------------------------------------------|
| Pending            | The WeaviateOpsRequest has been created but execution has not started yet           |
| Progressing        | KubeDB has started the execution of the applied WeaviateOpsRequest                  |
| Successful         | KubeDB has successfully performed the operation requested in the WeaviateOpsRequest |
| Failed             | KubeDB has failed the operation requested in the WeaviateOpsRequest                 |
| Denied             | KubeDB has denied the operation requested in the WeaviateOpsRequest                 |
| Skipped            | KubeDB has skipped the operation requested in the WeaviateOpsRequest                |
| WaitingForApproval | The WeaviateOpsRequest is waiting for approval before execution                     |

Ops-manager Operator can skip an opsRequest only if its execution has not been started yet and there is a newer opsRequest applied in the cluster. `spec.type` has to be the same as the skipped one, in this case.

#### status.pausedBackups

`status.pausedBackups` is a list of references to backup objects that were paused during the operation. Each entry has:

- `apiGroup` — the API group of the paused backup.
- `kind` — the kind of the paused backup.
- `name` — the name of the paused backup (required).
- `namespace` — the namespace of the paused backup.

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `WeaviateOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `WeaviateOpsRequest` processing. Each condition entry has the following fields:

- `type` specifies the type of the condition. WeaviateOpsRequest has the following types of conditions:

| Type                | Meaning                                                                          |
|---------------------|----------------------------------------------------------------------------------|
| `Progressing`       | Specifies that the operation is now progressing                                  |
| `Successful`        | Specifies that the operation on the database has been successful                 |
| `HaltDatabase`      | Specifies that the database is halted by the operator                            |
| `ResumeDatabase`    | Specifies that the database is resumed by the operator                           |
| `Failed`            | Specifies that the operation on the database has failed                          |
| `Scaling`           | Specifies that the scaling operation on the database has started                 |
| `VerticalScaling`   | Specifies that vertical scaling has performed successfully on the database       |
| `HorizontalScaling` | Specifies that horizontal scaling has performed successfully on the database     |

- The `status` field is a string, with possible values `"True"`, `"False"`, and `"Unknown"`.
    - `status` will be `"True"` if the current transition is succeeded.
    - `status` will be `"False"` if the current transition is failed.
    - `status` will be `"Unknown"` if the current transition is denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.

## Next Steps

- Learn about the [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) CRD.
- Deploy your first Weaviate database by following the guide [here](/docs/guides/weaviate/quickstart/quickstart.md).
