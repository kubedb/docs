---
title: DruidAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: guides-druid-concepts-druidautoscaler
    name: DruidAutoscaler
    parent: guides-druid-concepts
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DruidAutoscaler

## What is DruidAutoscaler

`DruidAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [Druid](https://druid.apache.org/) compute resources and storage of database components in a Kubernetes native way.

## DruidAutoscaler CRD Specifications

Like any official Kubernetes resource, a `DruidAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `DruidAutoscaler` CROs for autoscaling different components of database is given below:

**Sample `DruidAutoscaler` for `druid` cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DruidAutoscaler
metadata:
  name: dr-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: druid-prod
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    coordinators:
      trigger: "On"
      podLifeTimeThreshold: 24h
      minAllowed:
        cpu: 200m
        memory: 300Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
      resourceDiffPercentage: 10
    brokers:
      trigger: "On"
      podLifeTimeThreshold: 24h
      minAllowed:
        cpu: 200m
        memory: 300Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
      resourceDiffPercentage: 10
  storage:
    historicals:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
    middleManagers:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here, we are going to describe the various sections of a `DruidAutoscaler` crd.

A `DruidAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Druid](/docs/guides/druid/concepts/druid.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Druid](/docs/guides/druid/concepts/druid.md) object.

### spec.opsRequestOptions
These are the options to pass in the internally created opsRequest CRO. `opsRequestOptions` has two fields.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.coordinators` indicates the desired compute autoscaling configuration for coordinators of a topology Druid database.
- `spec.compute.overlords` indicates the desired compute autoscaling configuration for overlords of a topology Druid database.
- `spec.compute.brokers` indicates the desired compute autoscaling configuration for brokers of a topology Druid database.
- `spec.compute.routers` indicates the desired compute autoscaling configuration for routers of a topology Druid database.
- `spec.compute.historicals` indicates the desired compute autoscaling configuration for historicals of a topology Druid database.
- `spec.compute.middleManagers` indicates the desired compute autoscaling configuration for middleManagers of a topology Druid database.


All of them has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for this component of the database. If "On" then compute autoscaling is enabled. If "Off" then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are "cpu" and "memory".
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are "RequestsAndLimits" and "RequestsOnly".
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.

There are two more fields, those are only specifiable for the percona variant inMemory databases.
- `inMemoryStorage.UsageThresholdPercentage` If db uses more than usageThresholdPercentage of the total memory, memoryStorage should be increased.
- `inMemoryStorage.ScalingFactorPercentage` If db uses more than usageThresholdPercentage of the total memory, memoryStorage should be increased by this given scaling percentage.

### spec.storage

`spec.storage` specifies the autoscaling configuration for the storage resources of the database components. This field consists of the following sub-field:

- `spec.storage.historicals` indicates the desired storage autoscaling configuration for historicals of a topology Druid cluster.
- `spec.storage.middleManagers` indicates the desired storage autoscaling configuration for middleManagers of a topology Druid cluster.

> `spec.storage` is only supported for druid data nodes i.e. `historicals` and `middleManagers` as they are the only nodes containing volumes.

All of them has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
- `expansionMode` indicates the volume expansion mode.
