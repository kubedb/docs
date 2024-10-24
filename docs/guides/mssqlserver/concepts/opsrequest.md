---
title: MSSQLServerOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: ms-concepts-ops-request
    name: MSSQLServerOpsRequest
    parent: ms-concepts
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServerOpsRequest

## What is MSSQLServerOpsRequest

`MSSQLServerOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Microsoft SQL Server](https://learn.microsoft.com/en-us/sql/sql-server/) administrative operations like database version updating, horizontal scaling, vertical scaling, reconfigure, volume expansion, etc. in a Kubernetes native way.

## MSSQLServerOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MSSQLServerOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `MSSQLServerOpsRequest` CRs for different administrative operations is given below,

Sample `MSSQLServerOpsRequest` for updating database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: update-ms-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mssql-ag
  updateVersion:
    targetVersion: 2022-cu14
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the MSSQLServer successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `MSSQLServerOpsRequest` for horizontal scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: msops-horizontal-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mssql-ag
  horizontalScaling:
    replicas: 5
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the MSSQLServer successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `MSSQLServerOpsRequest` for vertical scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: mops-vscale-standalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mssql-standalone
  verticalScaling:
    mssqlserver:
      resources:
        requests:
          memory: "5Gi"
          cpu: "1000m"
        limits:
          memory: "5Gi"
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the MSSQLServer successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Here, we are going to describe the various sections of a `MSSQLServerOpsRequest` CR.

### MSSQLServerOpsRequest `Spec`

A `MSSQLServerOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field that point to the [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) object where the administrative operations will be applied. This field consists of the following sub-field:

- **spec.databaseRef.name :**  specifies the name of the [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) object.

#### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `MSSQLServerOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Restart`

>You can perform only one type of operation on a single `MSSQLServerOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `MSSQLServerOpsRequest`. At first, you have to create a `MSSQLServerOpsRequest` for updating. Once it is completed, then you can create another `MSSQLServerOpsRequest` for scaling. You should not create two `MSSQLServerOpsRequest` simultaneously.

#### spec.updateVersion

If you want to update your MSSQLServer version, you have to specify the `spec.updateVersion`  section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [MSSQLServerVersion](/docs/guides/mssqlserver/concepts/catalog.md) CR that contains the MSSQLServer version information where you want to update.

>You can only update between MSSQLServer versions. KubeDB does not support downgrade for MSSQLServer.

#### spec.horizontalScaling

If you want to scale-up or scale-down your MSSQLServer cluster, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.replicas` indicates the desired number of replicas for your MSSQLServer cluster after scaling. For example, if your cluster currently has 4 replicas, and you want to add additional 2 replicas then you have to specify 6 in `spec.horizontalScaling.replicas` field. Similarly, if you want to remove one member from the cluster, you have to specify 3  in `spec.horizontalScaling.replicas` field.

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MSSQLServer` resources like `cpu`, `memory` etc. that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.mssqlserver` indicates the `MSSQLServer` server resources. It has the below structure:

```yaml
resources:
  requests:
    memory: "5Gi"
    cpu: 1
  limits:
    memory: "5Gi"
    cpu: 2
```

Here, when you specify the resource request for `MSSQLServer` container, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for `MSSQLServer` container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. you can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.mssqlserver` and you can scale the resource the same way as `MSSQLServer` container.

>You can increase/decrease resources for both `MSSQLServer` container and `exporter` container on a single `MSSQLServerOpsRequest` CR.

### spec.volumeExpansion

> To use the volume expansion feature the storage class must support volume expansion

If you want to expand the volume of your MSSQLServer cluster, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.volumeExpansion.mode` specifies the volume expansion mode. Supported values are `Online` & `Offline`. The default is `Online`.
- `spec.volumeExpansion.mssqlserver` indicates the desired size for the persistent volumes of a MSSQLServer.

All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    mode: "Online"
    mssqlserver: 30Gi
```

This will expand the volume size of all the mssql server nodes to 30 GB.


#### spec.timeout

As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second). If a step doesn’t finish within the specified timeout, the ops request will result in failure.


#### spec.apply

This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`. Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.

### MSSQLServerOpsRequest `Status`

`.status` describes the current state and progress of the `MSSQLServerOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `MSSQLServerOpsRequest`. It can have the following values:

| Phase       | Meaning                                                                            |
|-------------|------------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the MSSQLServerOpsRequest |
| Progressing | KubeDB has started the execution of the applied MSSQLServerOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the MSSQLServerOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the MSSQLServerOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the MSSQLServerOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `MSSQLServerOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `MSSQLServerOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. MSSQLServerOpsRequest has the following types of conditions:

| Type                | Meaning                                                                                   |
|---------------------| ----------------------------------------------------------------------------------------- |
| `Progressing`       | Specifies that the operation is now progressing                                           |
| `Successful`        | Specifies such a state that the operation on the database has been successful.            |
| `HaltDatabase`      | Specifies such a state that the database is halted by the operator                        |
| `ResumeDatabase`    | Specifies such a state that the database is resumed by the operator                       |
| `Failed`           | Specifies such a state that the operation on the database has been failed.                |
| `Scaling`           | Specifies such a state that the scaling operation on the database has stared              |
| `VerticalScaling`   | Specifies such a state that vertical scaling has performed successfully on database       |
| `HorizontalScaling` | Specifies such a state that horizontal scaling has performed successfully on database     |
| `Updating`          | Specifies such a state that database updating operation has stared                       |
| `UpdateVersion`     | Specifies such a state that version updating on the database have performed successfully |

- The `status` field is a string, with possible values `"True"`, `"False"`, and `"Unknown"`.
    - `status` will be `"True"` if the current transition is succeeded.
    - `status` will be `"False"` if the current transition is failed.
    - `status` will be `"Unknown"` if the current transition is denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition. It has the following possible values:

| Reason                                   | Meaning                                                                          |
|------------------------------------------| -------------------------------------------------------------------------------- |
| `OpsRequestProgressingStarted`           | Operator has started the OpsRequest processing                                   |
| `OpsRequestFailedToProgressing`          | Operator has failed to start the OpsRequest processing                           |
| `SuccessfullyHaltedDatabase`             | Database is successfully halted by the operator                                  |
| `FailedToHaltDatabase`                   | Database is failed to halt by the operator                                       |
| `SuccessfullyResumedDatabase`            | Database is successfully resumed to perform its usual operation                  |
| `FailedToResumedDatabase`                | Database is failed to resume                                                     |
| `DatabaseVersionUpdatingStarted`         | Operator has started updating the database version                              |
| `SuccessfullyUpdatedDatabaseVersion`     | Operator has successfully updated the database version                          |
| `FailedToUpdateDatabaseVersion`          | Operator has failed to update the database version                              |
| `HorizontalScalingStarted`               | Operator has started the horizontal scaling                                      |
| `SuccessfullyPerformedHorizontalScaling` | Operator has successfully performed on horizontal scaling                        |
| `FailedToPerformHorizontalScaling`       | Operator has failed to perform on horizontal scaling                             |
| `VerticalScalingStarted`                 | Operator has started the vertical scaling                                        |
| `SuccessfullyPerformedVerticalScaling`   | Operator has successfully performed on vertical scaling                          |
| `FailedToPerformVerticalScaling`         | Operator has failed to perform on vertical scaling                               |
| `OpsRequestProcessedSuccessfully`        | Operator has completed the operation successfully requested by the OpeRequest cr |

- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
