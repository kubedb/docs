---
title: MySQLOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-concepts-opsrequest
    name: MySQLOpsRequest
    parent: guides-mysql-concepts
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQLOpsRequest

## What is MySQLOpsRequest

`MySQLOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [MySQL](https://www.mysql.com/) administrative operations like database version updating, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## MySQLOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MySQLOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `MySQLOpsRequest` CRs for different administrative operations is given below,

Sample `MySQLOpsRequest` for updating database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-ops-update
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: UpdateVersion
  updateVersion:
    targetVersion: 8.0.32
status:
  conditions:
  - lastTransitionTime: "2022-06-16T13:52:58Z"
    message: The controller has scaled/updated the MySQL successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `MySQLOpsRequest` for horizontal scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: HorizontalScaling  
  horizontalScaling:
    member: 3
status:
  conditions:
  - lastTransitionTime: "2022-06-16T13:52:58Z"
    message: The controller has scaled/updated the MySQL successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `MySQLOpsRequest` for vertical scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: VerticalScaling  
  verticalScaling:
    mysql:
      requests:
        memory: "1200Mi"
        cpu: "0.7"
      limits:
        memory: "1200Mi"
        cpu: "0.7"
status:
  conditions:
  - lastTransitionTime: "2022-06-11T09:59:05Z"
    message: The controller has scaled/updated the MySQL successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Here, we are going to describe the various sections of a `MySQLOpsRequest` cr.

### MySQLOpsRequest `Spec`

A `MySQLOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field that point to the [MySQL](/docs/guides/mysql/concepts/database/index.md) object where the administrative operations will be applied. This field consists of the following sub-field:

- **spec.databaseRef.name :**  specifies the name of the [MySQL](/docs/guides/mysql/concepts/database/index.md) object.

#### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `MySQLOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `volumeExpansion`
- `Restart`
- `Reconfigure`
- `ReconfigureTLS`

>You can perform only one type of operation on a single `MySQLOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `MySQLOpsRequest`. At first, you have to create a `MySQLOpsRequest` for updating. Once it is completed, then you can create another `MySQLOpsRequest` for scaling. You should not create two `MySQLOpsRequest` simultaneously.

#### spec.updateVersion

If you want to update your MySQL version, you have to specify the `spec.updateVersion`  section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [MySQLVersion](/docs/guides/mysql/concepts/catalog/index.md) CR that contains the MySQL version information where you want to update.

>You can only update between MySQL versions. KubeDB does not support downgrade for MySQL.

#### spec.horizontalScaling

If you want to scale-up or scale-down your MySQL cluster, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.member` indicates the desired number of members for your MySQL cluster after scaling. For example, if your cluster currently has 4 members and you want to add additional 2 members then you have to specify 6 in `spec.horizontalScaling.member` field. Similarly, if you want to remove one member from the cluster, you have to specify 3  in `spec.horizontalScaling.member` field.

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MySQL` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.mysql` indicates the `MySQL` server resources. It has the below structure:
  
```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request for `MySQL` container, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for `MySQL` container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. you can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.mysql` and you can scale the resource the same way as `mysql` container.

>You can increase/decrease resources for both `mysql` container and `exporter` container on a single `MySQLOpsRequest` CR.

### MySQLOpsRequest `Status`

`.status` describes the current state and progress of the `MySQLOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `MySQLOpsRequest`. It can have the following three values:

|Phase          |Meaning                                                               |
|---------------|-----------------------------------------------------------------------|
|Successful     | KubeDB has successfully performed the operation requested in the MySQLOpsRequest |
|Failed         | KubeDB has failed the operation requested in the MySQLOpsRequest |
|Denied         | KubeDB has denied the operation requested in the MySQLOpsRequest |

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `MySQLOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `MySQLOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. MySQLOpsRequest has the following types of conditions:

| Type               | Meaning                                                                  |
|--------------------| -------------------------------------------------------------------------|
| `Progressing`      | Specifies that the operation is now progressing |
| `Successful`       | Specifies such a state that the operation on the database has been successful. |
| `HaltDatabase`     | Specifies such a state that the database is halted by the operator   |
| `ResumeDatabase`   | Specifies such a state that the database is resumed by the operator    |
| `Failure`          | Specifies such a state that the operation on the database has been failed.  |
| `Scaling`          | Specifies such a state that the scaling operation on the database has stared |
| `VerticalScaling`  | Specifies such a state that vertical scaling has performed successfully on database  |
| `HorizontalScaling` | Specifies such a state that horizontal scaling has performed successfully on database |
| `Updating`         | Specifies such a state that database updating operation has stared  |
| `UpdateVersion`    | Specifies such a state that version updating on the database have performed successfully  |

- The `status` field is a string, with possible values `"True"`, `"False"`, and `"Unknown"`.
  - `status` will be `"True"` if the current transition is succeeded.
  - `status` will be `"False"` if the current transition is failed.
  - `status` will be `"Unknown"` if the current transition is denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition. It has the following possible values:

| Reason                                  | Meaning                                       |
|-----------------------------------------| -----------------------------------------------|
| `OpsRequestProgressingStarted`          | Operator has started the OpsRequest processing    |
| `OpsRequestFailedToProgressing`         | Operator has failed to start the OpsRequest processing    |
| `SuccessfullyHaltedDatabase`            | Database is successfully halted by the operator  |
| `FailedToHaltDatabase`                  | Database is failed to halt by the operator    |
| `SuccessfullyResumedDatabase`           | Database is successfully resumed to perform its usual operation  |
| `FailedToResumedDatabase`               | Database is failed to resume                   |
| `DatabaseVersionUpdatingStarted`        | Operator has started updating the database version    |
| `SuccessfullyUpdatedDatabaseVersion`    | Operator has successfully updated the database version |
| `FailedToUpdateDatabaseVersion`         | Operator has failed to update the database version   |
| `HorizontalScalingStarted`              | Operator has started the horizontal scaling          |
| `SuccessfullyPerformedHorizontalScaling` | Operator has successfully performed on horizontal scaling     |
| `FailedToPerformHorizontalScaling`      | Operator has failed to perform on horizontal scaling     |
| `VerticalScalingStarted`                | Operator has started the vertical scaling    |
| `SuccessfullyPerformedVerticalScaling`  | Operator has successfully performed on vertical scaling   |
| `FailedToPerformVerticalScaling`        | Operator has failed to perform on vertical scaling   |
| `OpsRequestProcessedSuccessfully`       | Operator has completed the operation successfully requested by the OpeRequest cr  |

- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
