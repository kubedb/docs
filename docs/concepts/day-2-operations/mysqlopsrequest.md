---
title: MySQLOpsRequest
menu:
  docs_{{ .version }}:
    identifier: mysql-ops-request
    name: MySQLOpsRequest
    parent: day-2-operations
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: day-2-operations
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# MySQLOpsRequest

## What is MySQLOpsRequest

`MySQLOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [MySQL](https://www.mysql.com/) administrative operations like database version upgrading, resources scaling in a Kubernetes native way. You have to configure and create a `MySQLOpsRequest` object for specific operations.

## MySQLOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MySQLOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample common structure of `MySQLOpsRequest`  for the operations of `MySQL`  is shown below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: Upgrade
  upgrade:
    targetVersion: 8.0.20
  type: HorizontalScaling  
  horizontalScaling:
    member: 3
  type: VerticalScaling  
  verticalScaling:
    mysql:
      requests:
        memory: "200Mi"
        cpu: "0.1"
      limits:
        memory: "300Mi"
        cpu: "0.2"
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/upgraded the MySQL successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Here, we are going to describe the various sections of `MySQLOpsRequest` crd.

### MySQLOpsRequest `Spec`

A `MySQLOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field specifying the reference of the [MySQL](/docs/concepts/databases/mysql.md) object where the administrative operations will be applied. This field consists of the following sub-fields:

- **spec.databaseRef.name :**  specifies the name of the [MySQL](/docs/concepts/databases/mysql.md) object.

#### spec.type

`spec.type` specifies what kind of operation will be applied to the database. Currently, there are three types of operation are allowed in `MySQLOpsRequest` is shown below,

- `Upgrade`, `HorizontalScaling` and `VerticalScaling`.

#### spec.upgrade

`spec.upgrade` is a required field specifying the information of `MySQL` version upgrading. This field consists of the following sub-fields:

- `spec.upgrade.targetVersion` refers to a `MySQL` version name that is upgrade from the current version to this targeted version.

#### spec.horizontalScaling

`spec.horizontalScaling` is a required field specifying the information of `MySQL` server node scaling. This field consists of the following sub-fields:

- `spec.horizontalScaling.member` indicates the number of server nodes of `MySQL` to be operated on.

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MySQL` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.mysql` indicates the `MySQL` server resources. It has the below structure:
  
    ```
    requests:
      memory: "200Mi"
      cpu: "0.1"
    limits:
      memory: "300Mi"
      cpu: "0.2"
    ```

  Here, when you specify the resource request for `MySQL` Container, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for `MySQL` Container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. you can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.mysql` and you can scale the resource the same way as `mysql` container.

### MySQLOpsRequest `Status`

`.status` describes the current state and progress of the `MySQLOpsRequest` operation and updated by the `MySQLOpsRequest` controller. The controller continually and actively manages every object's actual state to match the desired state you supplied. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `MySQLOpsRequest`.

- `status.phase` will be `Successful` only if the overall condition is Succeeded.

- `status.phase` will be `Failed` if any transition of the condition is failed.

- `status.phase` will be `Denied` if the operation type is not supported.

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `MySQLOpsRequest` controller.

#### status.conditions

`status.conditions` has an array of `MySQLOpsRequet` operations condition and this field describes the information of different conditions of the operation. Each element of the conditions array has six possible fields:

- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one status to another.

- The `message` field is a human-readable message indicating details about the transition.

- The `status` field is a string, with possible values `"True"`, `"False"`, and `"Unknown"`.
  - `status` will be `"True"` if the current transition is succeeded.
  - `status` will be `"False"` if the current transition is failed.
  - `status` will be `"Unknown"` if the current transition is denied.

- The `observedGeneration` shows the most recent transition generation observed by the `MySQLOpsRequest` controller.

- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition. It has the following possible values:

  | Reason                              | Usage                                            |
  | ----------------------------------- | ------------------------------------------------ |
  | `OpsRequestReconcileFailed`         | Last reconcile get failed                        |
  | `OpsRequestObserveGenerationFailed` | Last observe generation get failed               |
  | `OpsRequestDenied`                  | Ops request type denied                          |
  | `OpsRequestProgressing`             | Ops request is progressing                       |
  | `PausingDatabase`                   | Database get pausing                             |
  | `PausedDatabase`                    | Database is paused                               |
  | `ResumingDatabase`                  | Database get resuming                            |
  | `ResumedDatabase`                   | Database is resumed                              |
  | `OpsRequestUpgradingVersion`        | Ops request for upgrading db version             |
  | `OpsRequestUpgradedVersion`         | Ops request is successfully upgraded  db version |
  | `OpsRequestUpgradedVersionFailed`   | Ops request get failed in upgrading db version   |
  | `OpsRequestScalingDatabase`         | Ops request for scaling initialization           |
  | `OpsRequestHorizontalScaling`       | Ops request for horizontal scaling               |
  | `OpsRequestHorizontalScalingFailed` | Ops request get failed in horizontal scaling     |
  | `OpsRequestVerticalScaling`         | Ops request for vertical scaling                 |
  | `OpsRequestVerticalScalingFailed`   | Ops request get failed in vertical scaling       |
  | `OpsRequestSuccessful`              | Ops request is successful in the last transition |
  
- The `type` indicates the condition transition where the `MySQLOpsRequest` is in and it has the following possible values:

  | Type                | Usage                                                |
  | ------------------- | ---------------------------------------------------- |
  | `Progressing`       | Ops request is in progressing transition             |
  | `successful`        | Ops request is succeeded in the last transition      |
  | `PausingDatabase`   | Ops request is in pausing database transition        |
  | `PausedDatabase`    | Ops request is in paused database transition         |
  | `ResumingDatabase`  | Ops request is in resuming database transition       |
  | `ResumedDatabase`   | Ops request is in resumed database transition        |
  | `Failure`           | Ops request is failed in the last transition         |
  | `Denied`            | Ops request is denied for unsupported operation type |
  | `Scaling`           | Ops request is in the scaling transition             |
  | `VerticalScaling`   | Ops request is in the vertical scaling transition    |
  | `HorizaontalScaling`| Ops request is in the horizontal scaling transition  |
  | `UpgradingVersion`  | Ops request is in the upgrading transition           |
  | `UpgradedVersion`   | Ops request is in the vertical scaling transition    |