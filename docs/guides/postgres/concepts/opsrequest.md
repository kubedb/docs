---
title: PostgresOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-concepts-opsrequest
    name: PostgresOpsRequest
    parent: pg-concepts-postgres
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgresOpsRequest

## What is PostgresOpsRequest

`PostgresOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Postgres](https://www.postgresql.org/) administrative operations like database version updating, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## PostgresOpsRequest CRD Specifications

Like any official Kubernetes resource, a `PostgresOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `PostgresOpsRequest` CRs for different administrative operations is given below,

Sample `PostgresOpsRequest` for updating database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-ops-update
  namespace: demo
spec:
  databaseRef:
    name: pg-group
  type: UpdateVersion
  updateVersion:
    targetVersion: 8.0.35
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the Postgres successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `PostgresOpsRequest` for horizontal scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  databaseRef:
    name: pg-group
  type: HorizontalScaling  
  horizontalScaling:
    replicas: 3
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the Postgres successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `PostgresOpsRequest` for vertical scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  databaseRef:
    name: pg-group
  type: VerticalScaling  
  verticalScaling:
    postgres:
      requests:
        memory: "1200Mi"
        cpu: "0.7"
      limits:
        memory: "1200Mi"
        cpu: "0.7"
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/updated the Postgres successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Here, we are going to describe the various sections of a `PostgresOpsRequest` cr.

### PostgresOpsRequest `Spec`

A `PostgresOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Postgres](/docs/guides/postgres/concepts/postgres.md) object where the administrative operations will be applied. This field consists of the following sub-field:

- **spec.databaseRef.name :**  specifies the name of the [Postgres](/docs/guides/postgres/concepts/postgres.md) object.

#### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `PostgresOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `volumeExpansion`
- `Restart`
- `Reconfigure`
- `ReconfigureTLS`

>You can perform only one type of operation on a single `PostgresOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `PostgresOpsRequest`. At first, you have to create a `PostgresOpsRequest` for updating. Once it is completed, then you can create another `PostgresOpsRequest` for scaling. You should not create two `PostgresOpsRequest` simultaneously.

#### spec.updateVersion

If you want to update your Postgres version, you have to specify the `spec.updateVersion`  section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [PostgresVersion](/docs/guides/postgres/concepts/catalog.md) CR that contains the Postgres version information where you want to update.

>You can only update between Postgres versions. KubeDB does not support downgrade for Postgres.

#### spec.horizontalScaling

If you want to scale-up or scale-down your Postgres cluster, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.member` indicates the desired number of members for your Postgres cluster after scaling. For example, if your cluster currently has 4 members and you want to add additional 2 members then you have to specify 6 in `spec.horizontalScaling.member` field. Similarly, if you want to remove one member from the cluster, you have to specify 3  in `spec.horizontalScaling.member` field.

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `Postgres` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.postgres` indicates the `Postgres` server resources. It has the below structure:
  
```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request for `Postgres` container, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for `Postgres` container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. you can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.postgres` and you can scale the resource the same way as `postgres` container.

>You can increase/decrease resources for both `postgres` container and `exporter` container on a single `PostgresOpsRequest` CR.

### PostgresOpsRequest `Status`

`.status` describes the current state and progress of the `PostgresOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `PostgresOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                             |
| ---------- | ----------------------------------------------------------------------------------- |
| Successful | KubeDB has successfully performed the operation requested in the PostgresOpsRequest |
| Failed     | KubeDB has failed the operation requested in the PostgresOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the PostgresOpsRequest                 |

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `PostgresOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `PostgresOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. PostgresOpsRequest has the following types of conditions:

| Type               | Meaning                                                                                   |
|--------------------| ----------------------------------------------------------------------------------------- |
| `Progressing`      | Specifies that the operation is now progressing                                           |
| `Successful`       | Specifies such a state that the operation on the database has been successful.            |
| `HaltDatabase`     | Specifies such a state that the database is halted by the operator                        |
| `ResumeDatabase`   | Specifies such a state that the database is resumed by the operator                       |
| `Failure`          | Specifies such a state that the operation on the database has been failed.                |
| `Scaling`          | Specifies such a state that the scaling operation on the database has stared              |
| `VerticalScaling`  | Specifies such a state that vertical scaling has performed successfully on database       |
| `HorizontalScaling` | Specifies such a state that horizontal scaling has performed successfully on database     |
| `updating`        | Specifies such a state that database updating operation has stared                       |
| `UpdateVersion`    | Specifies such a state that version updating on the database have performed successfully |

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
