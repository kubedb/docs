---
title: OpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: mc-opsrequest-concepts
    name: OpsRequest
    parent: mc-concepts-memcached
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MemcachedOpsRequest

## What is MemcachedOpsRequest

`MemcachedOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for Memcached administrative operations like database version updating, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## MemcachedOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MemcachedOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `MemcachedOpsRequest` CRs for different administrative operations is given below.

Sample MemcachedOpsRequest for updating database version:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: update-memcd
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: memcd-quickstart
  updateVersion:
    targetVersion: 1.6.22
```

Sample `MemcachedOpsRequest` for horizontal scaling:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-horizontal-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: memcd-quickstart
  horizontalScaling:
    replicas: 2
```

Sample `MemcachedOpsRequest` for vertical scaling:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: memcd-quickstart
  verticalScaling:
    memcached:
      resources:
        requests:
          memory: "200Mi"
          cpu: "200m"
        limits:
          memory: "200Mi"
          cpu: "200m"
```

Sample `MemcachedOpsRequest` for reconfiguration:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-reconfig
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: memcd-quickstart
  configuration:
    applyConfig:
      memcached.conf: |
        -m 50
        -c 50
    restart: "true"
```

## MemcachedOpsRequest Spec
A `MemcachedOpsRequest` object has the following fields in the `spec` section:
### spec.databaseRef
`spec.databaseRef` is a required field that point to the Memcached object where the administrative operations will be applied. This field consists of the following sub-field:
- **spec.databaseRef.name :** specifies the name of the Memcached object.

### spec.type
`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `MemcachedOpsRequest`.

- UpdateVersion
- HorizontalScaling
- VerticalScaling
- Restart
- Reconfigure

### spec.updateVersion
If you want to update your Memcacheds version, you have to 
>You can only update between Memcached versions. KubeDB does not support downgrade for Memcached.

### spec.horizontalScaling
If you want to scale-up or scale-down your Memcached, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:


- `spec.horizontalScaling.replicas` indicates the desired number of replicas for your Memcahced instance after scaling. For example, if your cluster currently has 4 replicas, and you want to add additional 2 replicas then you have to specify 6 in spec.horizontalScaling.replicas field. Similarly, if you want to remove one replicas, you have to specify 3 in spec.horizontalScaling.replicas field.

### spec.verticalScaling
`spec.verticalScaling` is a required field specifying the information of Memcached resources like cpu, memory that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.memcached` indicates the `Memcached` server resources. It has the below structure:
```yaml
requests:
  memory: "200Mi"
  cpu: "200m"
limits:
  memory: "200Mi"
  cpu: "200m"
```

Here, when you specify the resource request for `Memcached` container, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for `Memcached` container, the kubelet enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

- `spec.verticalScaling.exporter` indicates the `exporter` container resources. It has the same structure as `spec.verticalScaling.memcached` and you can scale the resource the same way as `memcached` container.

>You can increase/decrease resources for both memcached  container and exporter container on a single MemcachedOpsRequest CR.

### spec.timeout
As we internally retry the ops request steps multiple times, this `timeout` field helps the users to specify the timeout for those steps of the ops request (in second). If a step doesn’t finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of opsRequest depending on the database state. It has two supported values: `Always` & `IfReady`. Use `IfReady`, if you want to process the opsRequest only when the database is Ready. And use `Always`, if you want to process the execution of opsReq irrespective of the Database state.

## MemcachedOpsRequest Status
After creating the Ops request status section is added in MemcachedOpsRequest CR. The yaml looks like following:

```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: MemcachedOpsRequest
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"MemcachedOpsRequest","metadata":{"annotations":{},"name":"memcached-mc","namespace":"demo"},"spec":{"databaseRef":{"name":"mc1"},"type":"VerticalScaling","verticalScaling":{"memcached":{"resources":{"limits":{"cpu":"200m","memory":"200Mi"},"requests":{"cpu":"200m","memory":"200Mi"}}}}}}
    creationTimestamp: "2024-08-26T10:41:05Z"
    generation: 1
    name: memcached-mc
    namespace: demo
    resourceVersion: "4572839"
    uid: f230978f-fc7b-4f8e-afad-6023910aba0e
  spec:
    apply: IfReady
    databaseRef:
      name: mc1
    type: VerticalScaling
    verticalScaling:
      memcached:
        resources:
          limits:
            cpu: 200m
            memory: 200Mi
          requests:
            cpu: 200m
            memory: 200Mi
  status:
    conditions:
    - lastTransitionTime: "2024-08-26T10:41:05Z"
      message: Memcached ops request is vertically scaling database
      observedGeneration: 1
      reason: VerticalScale
      status: "True"
      type: VerticalScale
    - lastTransitionTime: "2024-08-26T10:41:08Z"
      message: Successfully updated PetSets Resources
      observedGeneration: 1
      reason: UpdatePetSets
      status: "True"
      type: UpdatePetSets
    - lastTransitionTime: "2024-08-26T10:41:13Z"
      message: evict pod; ConditionStatus:True; PodName:mc1-0
      observedGeneration: 1
      status: "True"
      type: EvictPod--mc1-0
    - lastTransitionTime: "2024-08-26T10:41:13Z"
      message: is pod ready; ConditionStatus:False
      observedGeneration: 1
      status: "False"
      type: IsPodReady
    - lastTransitionTime: "2024-08-26T10:41:18Z"
      message: is pod ready; ConditionStatus:True; PodName:mc1-0
      observedGeneration: 1
      status: "True"
      type: IsPodReady--mc1-0
    - lastTransitionTime: "2024-08-26T10:41:18Z"
      message: is pod resources updated; ConditionStatus:True; PodName:mc1-0
      observedGeneration: 1
      status: "True"
      type: IsPodResourcesUpdated--mc1-0
    - lastTransitionTime: "2024-08-26T10:41:18Z"
      message: Successfully Restarted Pods With Resources
      observedGeneration: 1
      reason: RestartPods
      status: "True"
      type: RestartPods
    - lastTransitionTime: "2024-08-26T10:41:18Z"
      message: Successfully Vertically Scaled Database
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
    observedGeneration: 1
    phase: Successful
```
`.status` describes the current state of the `MemcachedOpsRequest` operation. It has the following fields:

### status.phase
`status.phase` indicates the overall phase of the operation for this `MemcachedOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                          |
| ---------- |----------------------------------------------------------------------------------|
| Successful | KubeDB has successfully performed the operation requested in the MemcachedOpsRequest |
| Failed     | KubeDB has failed the operation requested in the MemcachedOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the MemcachedOpsRequest                 |

### status.observedGeneration
`status.observedGeneration` shows the most recent generation observed by the `MemcachedOpsRequest` controller.

### status.conditions
`status.conditions` is an array that specifies the conditions of different steps of `MemcachedOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. `MemcachedOpsRequest` has the following types of conditions:

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

## Next Steps

- Learn about Memcached crd [here](/docs/guides/memcached/concepts/memcached.md).
- Deploy your first Memcached database with Memcached by following the guide [here](/docs/guides/memcached/quickstart/quickstart.md).