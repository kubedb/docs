---
title: DocumentDBOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: dc-concepts-documentdb-opsrequest
    name: DocumentDBOpsRequest
    parent: dc-concepts-documentdb
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DocumentDBOpsRequest

## What is DocumentDBOpsRequest

`DocumentDBOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for `DocumentDB` administrative operations like database version updating, horizontal scaling, vertical scaling, volume expansion, reconfiguration, TLS reconfiguration, auth rotation, restart, and failover in a Kubernetes native way.

## DocumentDBOpsRequest CRD Specifications

Like any official Kubernetes resource, a `DocumentDBOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `DocumentDBOpsRequest` CRs for different administrative operations is given below.

Sample `DocumentDBOpsRequest` for updating database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-ops-update
  namespace: demo
spec:
  databaseRef:
    name: documentdb-group
  type: UpdateVersion
  updateVersion:
    targetVersion: "16.0"
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the DocumentDB successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `DocumentDBOpsRequest` for horizontal scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-hscale
  namespace: demo
spec:
  databaseRef:
    name: documentdb-group
  type: HorizontalScaling
  horizontalScaling:
    replicas: 3
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the DocumentDB successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `DocumentDBOpsRequest` for vertical scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-vscale
  namespace: demo
spec:
  databaseRef:
    name: documentdb-group
  type: VerticalScaling
  verticalScaling:
    documentdb:
      resources:
        requests:
          memory: "1200Mi"
          cpu: "0.7"
        limits:
          memory: "1200Mi"
          cpu: "0.7"
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the DocumentDB successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Here, we are going to describe the various sections of a `DocumentDBOpsRequest` cr.

### DocumentDBOpsRequest `Spec`

A `DocumentDBOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field that points to the `DocumentDB` object where the administrative operations will be applied. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the `DocumentDB` object.

#### spec.type

`spec.type` is a required field that specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `DocumentDBOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Restart`
- `Reconfigure`
- `ReconfigureTLS`
- `RotateAuth`
- `ReconnectStandby`
- `ForceFailOver`
- `SetRaftKeyPair`
- `StorageMigration`

>You can perform only one type of operation on a single `DocumentDBOpsRequest` CR. For example, if you want to update your database and scale up its replicas then you have to create two separate `DocumentDBOpsRequest`. At first, you have to create a `DocumentDBOpsRequest` for updating. Once it is completed, then you can create another `DocumentDBOpsRequest` for scaling. You should not create two `DocumentDBOpsRequest` simultaneously.

#### spec.updateVersion

If you want to update your DocumentDB version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a `DocumentDBVersion` CR that contains the DocumentDB version information where you want to update.

>You can only update between DocumentDB versions. KubeDB does not support downgrade for DocumentDB.

#### spec.horizontalScaling

If you want to scale-up or scale-down your DocumentDB cluster, you have to specify the `spec.horizontalScaling` section. This field consists of the following sub-fields:

- `spec.horizontalScaling.replicas` indicates the desired number of replicas for your DocumentDB cluster after scaling. For example, if your cluster currently has 3 replicas and you want to add 2 more, then you have to specify 5 in this field. Similarly, to remove one replica you specify 2.
- `spec.horizontalScaling.standbyMode` specifies the standby mode of the newly added standbys. It can be either `Hot` or `Warm`. Optional; defaults to `Hot`.
- `spec.horizontalScaling.streamingMode` specifies the replication streaming mode used for the standbys. It can be either `Synchronous` or `Asynchronous`. Optional; defaults to `Asynchronous`.
- `spec.horizontalScaling.readReplicas` is a list used to scale the read-replica members of the cluster.

#### spec.verticalScaling

`spec.verticalScaling` is a field specifying the information of `DocumentDB` resources like `cpu`, `memory` etc. that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.documentdb` indicates the `DocumentDB` server (main container) resources. It has the below structure:

```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request for the `DocumentDB` container, the scheduler uses this information to decide which node to place the container of the Pod on, and when you specify a resource limit for the `DocumentDB` container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can find more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.documentdb` and you can scale the resource the same way.
- `spec.verticalScaling.coordinator` indicates the `coordinator` container resources. It has the same structure as `spec.verticalScaling.documentdb`.
- `spec.verticalScaling.arbiter` indicates the resources of the `arbiter` Pods.
- `spec.verticalScaling.readReplicas` is a list used to scale the resources of individual read replicas. Each entry has a `documentdb` field (the container resources) and a `name` field identifying the read replica.
- `spec.verticalScaling.mode` specifies how the scaling is actuated. `Restart` (the default) applies the new resources by restarting the Pods, while `InPlace` resizes the running Pods in place via the Kubernetes `pods/resize` subresource (no restart), automatically falling back to `Restart` for any Pod whose Node cannot fit the new resources. Optional; defaults to `Restart`.

>You can increase/decrease resources for the `documentdb`, `exporter`, `coordinator`, and `arbiter` containers on a single `DocumentDBOpsRequest` CR.

#### spec.volumeExpansion

If you want to expand the volume of your DocumentDB cluster, you have to specify the `spec.volumeExpansion` section. This field consists of the following sub-fields:

- `spec.volumeExpansion.mode` specifies the volume expansion mode. It can be either `Online` or `Offline`, depending on whether the underlying storage class supports online volume expansion.
- `spec.volumeExpansion.documentdb` indicates the desired size for the persistent volume of the `DocumentDB` container.

>Volume expansion is only supported when the underlying `StorageClass` supports volume expansion (`allowVolumeExpansion: true`).

#### spec.configuration

If you want to reconfigure your running DocumentDB cluster with a new custom configuration, you have to specify the `spec.configuration` section (used with the `Reconfigure` type). It lets you apply, change, or remove custom configuration by referencing a Secret or providing inline configuration data.

#### spec.tls

If you want to reconfigure the TLS configuration of your DocumentDB cluster (issuer, certificates, or turning TLS on/off), you have to specify the `spec.tls` section (used with the `ReconfigureTLS` type).

#### spec.authentication

`spec.authentication` is used with the `RotateAuth` type to rotate the database credentials. You can let the operator generate a new random password, or supply your own credentials by referencing a Secret.

#### spec.restart

`spec.restart` is used with the `Restart` type to restart every Pod of the database in a controlled, rolling fashion.

#### spec.reconnectStandby

`spec.reconnectStandby` is used with the `ReconnectStandby` type to re-establish replication for standbys that have fallen out of sync with the primary.

#### spec.forceFailOver

`spec.forceFailOver` is used with the `ForceFailOver` type to force a failover, promoting a standby to become the new primary.

#### spec.setRaftKeyPair

`spec.setRaftKeyPair` is used with the `SetRaftKeyPair` type to set the Raft key pair used for secure intra-cluster communication.

#### spec.migration

`spec.migration` is used with the `StorageMigration` type to migrate the database storage between different storage backends.

#### spec.timeout

`spec.timeout` is the timeout for each step of the ops request. If a step doesn't finish within the specified timeout, the ops request will result in failure.

#### spec.apply

`spec.apply` controls the execution of the OpsRequest depending on the database state. It can be either `IfReady` or `Always`. When set to `IfReady` (the default), the operator only proceeds if the database is in a ready state. When set to `Always`, the operator applies the OpsRequest regardless of the database state.

#### spec.maxRetries

`spec.maxRetries` specifies how many times a failed step of the OpsRequest should be retried before the OpsRequest is marked as failed. Optional; defaults to 1.

### DocumentDBOpsRequest `Status`

`.status` describes the current state and progress of the `DocumentDBOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `DocumentDBOpsRequest`. It can have the following values:

| Phase      | Meaning                                                                               |
| ---------- | ------------------------------------------------------------------------------------- |
| Successful | KubeDB has successfully performed the operation requested in the DocumentDBOpsRequest |
| Failed     | KubeDB has failed the operation requested in the DocumentDBOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the DocumentDBOpsRequest                 |

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `DocumentDBOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `DocumentDBOpsRequest` processing. Each condition entry has the following fields:

- `type` specifies the type of the condition. DocumentDBOpsRequest has the following types of conditions:

| Type                | Meaning                                                                               |
|---------------------| ------------------------------------------------------------------------------------- |
| `Progressing`       | Specifies that the operation is now progressing                                       |
| `Successful`        | Specifies such a state that the operation on the database has been successful.        |
| `HaltDatabase`      | Specifies such a state that the database is halted by the operator                    |
| `ResumeDatabase`    | Specifies such a state that the database is resumed by the operator                   |
| `Failure`           | Specifies such a state that the operation on the database has been failed.            |
| `Scaling`           | Specifies such a state that the scaling operation on the database has started         |
| `VerticalScaling`   | Specifies such a state that vertical scaling has performed successfully on database   |
| `HorizontalScaling` | Specifies such a state that horizontal scaling has performed successfully on database |
| `Updating`          | Specifies such a state that database updating operation has started                   |
| `UpdateVersion`     | Specifies such a state that version updating on the database has performed successfully |

- The `status` field is a string, with possible values `"True"`, `"False"`, and `"Unknown"`.
  - `status` will be `"True"` if the current transition succeeded.
  - `status` will be `"False"` if the current transition failed.
  - `status` will be `"Unknown"` if the current transition is denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
