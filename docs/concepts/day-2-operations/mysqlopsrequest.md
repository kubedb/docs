---
title: MySQLOpsRequests
menu:
  docs_{{ .version }}:
    identifier: concepts-opsrequests-mysqlopsrequests
    name: MySQLOpsRequests
    parent: concepts-opsrequests
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

{{< notice type="warning" message="This doc has described only the KubeDB enterprise feature. If you are a KubeDB enterprise user then you have to explore it" >}}

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# MySQLOpsRequest

## What is MySQLOpsRequest

`MySQLOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [MySQL](https://www.mysql.com/) administrative operations like database version upgrading, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## MySQLOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MySQLOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `MySQLOpsRequest` CRs for different administrative operations is given below,

Sample `MySQLOpsRequest` for upgrading database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-ops-upgrade
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: Upgrade
  upgrade:
    targetVersion: 8.0.20
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
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/upgraded the MySQL successfully
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

Here, we are going to describe the various sections of a `MySQLOpsRequest` crd.

### MySQLOpsRequest `Spec`

A `MySQLOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field that point to the [MySQL](/docs/concepts/databases/mysql.md) object where the administrative operations will be applied. This field consists of the following sub-field:

- **spec.databaseRef.name :**  specifies the name of the [MySQL](/docs/concepts/databases/mysql.md) object.

#### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `MySQLOpsRequest`.

- `Upgrade` 
- `HorizontalScaling`
- `VerticalScaling`

>You can perform only one type of operation on a single `MySQLOpsRequest` CR. For example, if you want to upgrade your database and scale up its replica then you have to create two separate `MySQLOpsRequest`. At first, you have to create a `MySQLOpsRequest` for upgrading. Once it is completed, then you can create another `MySQLOpsRequest` for scaling. You should not create two `MySQLOpsRequest` simultaneously.

#### spec.upgrade

If you want to upgrade you MySQL version, you have to specify the `spec.upgrade`  section that specifies the desired version information. This field consists of the following sub-field:

- `spec.upgrade.targetVersion` refers to a [MySQLVersion](/docs/concepts/catalog/mysql.md) CR that contains the MySQL version information where you want to upgrade.

>You can only upgrade between MySQL versions. KubeDB does not support downgrade for MySQL.

#### spec.horizontalScaling

If you want to scale-up or scale-down your MySQL cluster, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.member` indicates the desired number of members for your MySQL cluster after scaling. For example, if your cluster currently has 4 members and you want to add additional 2 members then you have to specify 6 in `spec.horizontalScaling.member` field. Similarly, if you want to remove one member from the cluster, you have specify 3  in `spec.horizontalScaling.member` field.

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

Here, when you specify the resource request for `MySQL` container, the scheduler uses thisinformation to decide which node to place the container of the Pod on and when you specify a resourcelimit for `MySQL` container, the `kubelet` enforces those limits so that the running container is notallowed to use more of that resource than the limit you set. you can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

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

| Type                | Meaning                                                                  |
| ------------------- | -------------------------------------------------------------------------|
| `Progressing`       | Specifies that the operation is now in the progressing state  |
| `Successful`        | Specifies such a state that the operation on the database has beensuccessful. |
| `PauseDatabase`     | Specifies such a state that the database is paused by the operator   |
| `ResumeDatabase`    | Specifies such a state that the database is resumed by the operator    |
| `Failure`           | Specifies such a state that the operation on the database has been failed.  |
| `Scaling`           | Specifies such a state that the scaling operation on the database has stared |
| `VerticalScaling`   | Specifies such a state that vertical scaling have performed successfully on database  |
| `HorizontalScaling` | Specifies such a state that horizontal scaling have performed successfully on database |
| `Upgrading`         | Specifies such a state that database upgrading operation has stared  |
| `UpgradeVersion`    | Specifies such a state that version upgrading on database have performed successfully  |

- The `status` field is a string, with possible values `"True"`, `"False"`, and `"Unknown"`.
  - `status` will be `"True"` if the current transition is succeeded.
  - `status` will be `"False"` if the current transition is failed.
  - `status` will be `"Unknown"` if the current transition is denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition. It has the following possible values:

| Reason                                  | Meaning                                       |
| --------------------------------------- | -----------------------------------------------|
| `OpsRequestProgressingStarted`          | Operator has started the OpsRequest processing    |
| `OpsRequestFailedToProgressing`         | Operator has failed to start the OpsRequest processing    |
| `SuccessfullyPausedDatabase`            | Database is successfully paused by the operator  |
| `FailedToPauseDatabase`                 | Database is failed to pause by the operator    |
| `SuccessfullyResumedDatabase`           | Database is successfully resumed to perform it's usual operation  |
| `FailedToResumedDatabase`               | Database is failed to resume                   |
| `DatabaseVersionUpgradingStarted`       | Operator has started upgrading the database version    |
| `SuccessfullyUpgradedDatabaseVersion`   | Operator has successfully upgraded the database version |
| `FailedToUpgradeDatabaseVersion`        | Operator has failed to upgrade the database version   |
| `HorizontalScalingStarted`              | Operator has started the horizontal scaling          |
| `SuccessfullyPerformedHorizontalScaling` | Operator has successfully performed on horizontal scaling     |
| `FailedToPerformHorizontalScaling`      | Operator has failed to perform on horizontal scaling     |
| `VerticalScalingStarted`                | Operator has started the vertical scaling    |
| `SuccessfullyPerformedVerticalScaling`  | Operator has successfully performed on vertical scaling   |
| `FailedToPerformVerticalScaling`        | Operator has failed to perform on vertical scaling   |
| `OpsRequestProcessedSuccessfully`       | Operator has successfully completed the operator requested by the OpeRequest cr |

- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
