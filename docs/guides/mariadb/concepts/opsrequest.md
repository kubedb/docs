---
title: MariaDBOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: my-opsrequest-concepts
    name: MariaDBOpsRequest
    parent: my-concepts-mariadb
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# MariaDBOpsRequest

## What is MariaDBOpsRequest

`MariaDBOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [MariaDB](https://www.mariadb.com/) administrative operations like database version upgrading, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## MariaDBOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MariaDBOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `MariaDBOpsRequest` CRs for different administrative operations is given below,

Sample `MariaDBOpsRequest` for upgrading database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
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
    message: The controller has scaled/upgraded the MariaDB successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `MariaDBOpsRequest` for horizontal scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
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
    message: The controller has scaled/upgraded the MariaDB successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Sample `MariaDBOpsRequest` for vertical scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: VerticalScaling  
  verticalScaling:
    mariadb:
      requests:
        memory: "200Mi"
        cpu: "0.1"
      limits:
        memory: "300Mi"
        cpu: "0.2"
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/upgraded the MariaDB successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Here, we are going to describe the various sections of a `MariaDBOpsRequest` cr.

### MariaDBOpsRequest `Spec`

A `MariaDBOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field that point to the [MariaDB](/docs/guides/mariadb/concepts/mariadb.md) object where the administrative operations will be applied. This field consists of the following sub-field:

- **spec.databaseRef.name :**  specifies the name of the [MariaDB](/docs/guides/mariadb/concepts/mariadb.md) object.

#### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `MariaDBOpsRequest`.

- `Upgrade` 
- `HorizontalScaling`
- `VerticalScaling`

>You can perform only one type of operation on a single `MariaDBOpsRequest` CR. For example, if you want to upgrade your database and scale up its replica then you have to create two separate `MariaDBOpsRequest`. At first, you have to create a `MariaDBOpsRequest` for upgrading. Once it is completed, then you can create another `MariaDBOpsRequest` for scaling. You should not create two `MariaDBOpsRequest` simultaneously.

#### spec.upgrade

If you want to upgrade your MariaDB version, you have to specify the `spec.upgrade`  section that specifies the desired version information. This field consists of the following sub-field:

- `spec.upgrade.targetVersion` refers to a [MariaDBVersion](/docs/guides/mariadb/concepts/catalog.md) CR that contains the MariaDB version information where you want to upgrade.

>You can only upgrade between MariaDB versions. KubeDB does not support downgrade for MariaDB.

#### spec.horizontalScaling

If you want to scale-up or scale-down your MariaDB cluster, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.member` indicates the desired number of members for your MariaDB cluster after scaling. For example, if your cluster currently has 4 members and you want to add additional 2 members then you have to specify 6 in `spec.horizontalScaling.member` field. Similarly, if you want to remove one member from the cluster, you have to specify 3  in `spec.horizontalScaling.member` field.

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MariaDB` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.mariadb` indicates the `MariaDB` server resources. It has the below structure:
  
```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request for `MariaDB` container, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for `MariaDB` container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. you can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.mariadb` and you can scale the resource the same way as `mariadb` container.

>You can increase/decrease resources for both `mariadb` container and `exporter` container on a single `MariaDBOpsRequest` CR.

### MariaDBOpsRequest `Status`

`.status` describes the current state and progress of the `MariaDBOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `MariaDBOpsRequest`. It can have the following three values:

|Phase          |Meaning                                                               |
|---------------|-----------------------------------------------------------------------|
|Successful     | KubeDB has successfully performed the operation requested in the MariaDBOpsRequest |
|Failed         | KubeDB has failed the operation requested in the MariaDBOpsRequest |
|Denied         | KubeDB has denied the operation requested in the MariaDBOpsRequest |

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `MariaDBOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `MariaDBOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. MariaDBOpsRequest has the following types of conditions:

| Type                | Meaning                                                                  |
| ------------------- | -------------------------------------------------------------------------|
| `Progressing`       | Specifies that the operation is now progressing |
| `Successful`        | Specifies such a state that the operation on the database has been successful. |
| `HaltDatabase`     | Specifies such a state that the database is halted by the operator   |
| `ResumeDatabase`    | Specifies such a state that the database is resumed by the operator    |
| `Failure`           | Specifies such a state that the operation on the database has been failed.  |
| `Scaling`           | Specifies such a state that the scaling operation on the database has stared |
| `VerticalScaling`   | Specifies such a state that vertical scaling has performed successfully on database  |
| `HorizontalScaling` | Specifies such a state that horizontal scaling has performed successfully on database |
| `Upgrading`         | Specifies such a state that database upgrading operation has stared  |
| `UpgradeVersion`    | Specifies such a state that version upgrading on the database have performed successfully  |

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
| `SuccessfullyHaltedDatabase`            | Database is successfully halted by the operator  |
| `FailedToHaltDatabase`                 | Database is failed to halt by the operator    |
| `SuccessfullyResumedDatabase`           | Database is successfully resumed to perform its usual operation  |
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
| `OpsRequestProcessedSuccessfully`       | Operator has completed the operation successfully requested by the OpeRequest cr  |

- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
