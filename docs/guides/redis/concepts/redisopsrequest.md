---
title: OpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: guides-redis-concepts-opsrequest
    name: OpsRequest
    parent: rd-concepts-redis
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RedisOpsRequest 

## What is RedisOpsRequest

`RedisOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Redis](https://www.redis.io/) administrative operations like database version updating, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## RedisOpsRequest CRD Specifications

Like any official Kubernetes resource, a `RedisOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `RedisOpsRequest` CRs for different administrative operations is given below,

Sample `RedisOpsRequest` for updating database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: standalone-redis
  updateVersion:
    targetVersion: 7.0.14
```

Sample `RedisOpsRequest` for horizontal scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: up-horizontal-redis-ops
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: redis-cluster
  horizontalScaling:
    shards: 5
    replicas: 2 
```


## What is RedisSentinelOpsRequest

`RedisSentinelOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Redis](https://www.redis.io/) administrative operations like database version updating, horizontal scaling, vertical scaling, reconfiguring TLS etc. in a Kubernetes native way.
The spec in `RedisOpsRequest` and `RedisSentinelOpsRequest` similar which will be described below.

Sample `RedisSentinelOpsRequest` for vertical scaling
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: redisops-vertical
  namespace: omed
spec:
  type: VerticalScaling
  databaseRef:
    name: sentinel-tls
  verticalScaling:
    redissentinel:
      resources:
        requests:
          memory: "300Mi"
          cpu: "200m"
        limits:
          memory: "800Mi"
          cpu: "500m"
```

Here, we are going to describe the various sections of `RedisOpsRequest` and `RedisSentinelOpsRequest` CR .

### RedisOpsRequest `Spec`

A `RedisOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Redis](/docs/guides/redis/concepts/redis.md) object where the administrative operations will be applied. This field consists of the following sub-field:

- **spec.databaseRef.name :**  specifies the name of the [Redis](/docs/guides/redis/concepts/redis.md) object.

#### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `RedisOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Restart`
- `Reconfigure`
- `ReconfigureTLS`
- `ReplaceSentinel` (Only in Sentinel Mode)

`Reconfigure` and `ReplaceSentinel` ops request can not be done in `RedisSentinelOpsRequest`

>You can perform only one type of operation on a single `RedisOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `RedisOpsRequest`. At first, you have to create a `RedisOpsRequest` for updating. Once it is completed, then you can create another `RedisOpsRequest` for scaling. You should not create two `RedisOpsRequest` simultaneously.

#### spec.updateVersion

If you want to update your Redis version, you have to specify the `spec.updateVersion`  section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [RedisVersion](/docs/guides/redis/concepts/catalog.md) CR that contains the Redis version information where you want to update.

>You can only update between Redis versions. KubeDB does not support downgrade for Redis.

#### spec.horizontalScaling

If you want to scale-up or scale-down your Redis cluster, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.replicas` indicates the desired number of replicas for your Redis instance after scaling. For example, if your cluster currently has 4 replicas, and you want to add additional 2 replicas then you have to specify 6 in `spec.horizontalScaling.replicas` field. Similarly, if you want to remove one replicas, you have to specify 3  in `spec.horizontalScaling.replicas` field.
- `spec.horizontalScaling.shards` indicates the desired number of shards for your Redis cluster. It is only applicable for Cluster Mode.

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `Redis` resources like `cpu`, `memory` that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.redis` indicates the `Redis` server resources. It has the below structure:
  
```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request for `Redis` container, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for `Redis` container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. you can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.redis` and you can scale the resource the same way as `redis` container.

>You can increase/decrease resources for both `redis` container and `exporter` container on a single `RedisOpsRequest` CR.

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.


### RedisOpsRequest `Status`

After creating the Ops request `status` section is added in RedisOpsRequest CR. The yaml looks like following : 
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"RedisOpsRequest","metadata":{"annotations":{},"name":"redisops-vertical","namespace":"demo"},"spec":{"databaseRef":{"name":"standalone-redis"},"type":"VerticalScaling","verticalScaling":{"redis":{"limits":{"cpu":"500m","memory":"800Mi"},"requests":{"cpu":"200m","memory":"300Mi"}}}}}
  creationTimestamp: "2023-02-02T09:14:01Z"
  generation: 1
  name: redisops-vertical
  namespace: demo
  resourceVersion: "483411"
  uid: 12c45d9c-daea-472d-be61-b88505cb755d
spec:
  apply: IfReady
  databaseRef:
    name: standalone-redis
  type: VerticalScaling
  verticalScaling:
    redis:
      resources:
        limits:
          cpu: 500m
          memory: 800Mi
        requests:
          cpu: 200m
          memory: 300Mi
status:
  conditions:
  - lastTransitionTime: "2023-02-02T09:14:01Z"
    message: Redis ops request is vertically scaling database
    observedGeneration: 1
    reason: VerticalScaling
    status: "True"
    type: VerticalScaling
  - lastTransitionTime: "2023-02-02T09:14:01Z"
    message: Successfully updated PetSets Resources
    observedGeneration: 1
    reason: UpdatePetSetResources
    status: "True"
    type: UpdatePetSetResources
  - lastTransitionTime: "2023-02-02T09:14:12Z"
    message: Successfully Restarted Pods With Resources
    observedGeneration: 1
    reason: RestartedPodsWithResources
    status: "True"
    type: RestartedPodsWithResources
  - lastTransitionTime: "2023-02-02T09:14:12Z"
    message: Successfully Vertically Scaled Database
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

```

`.status` describes the current state of the `RedisOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `RedisOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                          |
| ---------- |----------------------------------------------------------------------------------|
| Successful | KubeDB has successfully performed the operation requested in the RedisOpsRequest |
| Failed     | KubeDB has failed the operation requested in the RedisOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the RedisOpsRequest                 |

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `RedisOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `RedisOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. RedisOpsRequest has the following types of conditions:

| Type                | Meaning                                                                                  |
|---------------------|------------------------------------------------------------------------------------------|
| `Progressing`       | Specifies that the operation is now progressing                                          |
| `Successful`        | Specifies such a state that the operation on the database has been successful.           |
| `HaltDatabase`      | Specifies such a state that the database is halted by the operator                       |
| `ResumeDatabase`    | Specifies such a state that the database is resumed by the operator                      |
| `Failure`           | Specifies such a state that the operation on the database has been failed.               |
| `Scaling`           | Specifies such a state that the scaling operation on the database has stared             |
| `VerticalScaling`   | Specifies such a state that vertical scaling has performed successfully on database      |
| `HorizontalScaling` | Specifies such a state that horizontal scaling has performed successfully on database    |
| `UpdateVersion`     | Specifies such a state that version updating on the database have performed successfully |

- The `status` field is a string, with possible values `"True"`, `"False"`, and `"Unknown"`.
  - `status` will be `"True"` if the current transition is succeeded.
  - `status` will be `"False"` if the current transition is failed.
  - `status` will be `"Unknown"` if the current transition is denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition. It has the following possible values:

| Reason                                  | Meaning                                                                          |
|-----------------------------------------| -------------------------------------------------------------------------------- |
| `OpsRequestProgressingStarted`          | Operator has started the OpsRequest processing                                   |
| `OpsRequestFailedToProgressing`         | Operator has failed to start the OpsRequest processing                           |
| `SuccessfullyHaltedDatabase`            | Database is successfully halted by the operator                                  |
| `FailedToHaltDatabase`                  | Database is failed to halt by the operator                                       |
| `SuccessfullyResumedDatabase`           | Database is successfully resumed to perform its usual operation                  |
| `FailedToResumedDatabase`               | Database is failed to resume                                                     |
| `DatabaseVersionupdatingStarted`       | Operator has started updating the database version                              |
| `SuccessfullyUpdatedDatabaseVersion`    | Operator has successfully updated the database version                          |
| `FailedToUpdateDatabaseVersion`         | Operator has failed to update the database version                              |
| `HorizontalScalingStarted`              | Operator has started the horizontal scaling                                      |
| `SuccessfullyPerformedHorizontalScaling` | Operator has successfully performed on horizontal scaling                        |
| `FailedToPerformHorizontalScaling`      | Operator has failed to perform on horizontal scaling                             |
| `VerticalScalingStarted`                | Operator has started the vertical scaling                                        |
| `SuccessfullyPerformedVerticalScaling`  | Operator has successfully performed on vertical scaling                          |
| `FailedToPerformVerticalScaling`        | Operator has failed to perform on vertical scaling                               |
| `OpsRequestProcessedSuccessfully`       | Operator has completed the operation successfully requested by the OpeRequest cr |

- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
