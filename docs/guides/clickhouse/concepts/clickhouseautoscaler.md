---
title: ClickHouseAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: ch-autoscaler-concepts
    name: ClickHouseAutoscaler
    parent: ch-concepts-clickhouse
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouseAutoscaler

## What is ClickHouseAutoscaler

`ClickHouseAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [ClickHouse](https://clickhouse.com/) compute resources and storage of database components in a Kubernetes native way.

## ClickHouseAutoscaler CRD Specifications

Like any official Kubernetes resource, a `ClickHouseAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `ClickHouseAutoscaler` CROs for autoscaling different components of database is given below:

**Sample `ClickHouseAutoscaler` for `clickhouse` cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ClickHouseAutoscaler
metadata:
  name: ch-compute-autoscale
  namespace: demo
spec:
  databaseRef:
    name: clickhouse-prod
  compute:
    clickhouse:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 1
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here, we are going to describe the various sections of a `ClickHouseAutoscaler` crd.

A `ClickHouseAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md) object.

### spec.opsRequestOptions
These are the options to pass in the internally created opsRequest CRO. `opsRequestOptions` has two fields.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.clickhouse` indicates the desired compute autoscaling configuration for ClickHouse database.


`spec.compute.clickhouse` has the following sub-fields:

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

- `spec.storage.clickhouse` indicates the desired storage autoscaling configuration for ClickHouse cluster.


All of them has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
- `expansionMode` indicates the volume expansion mode.
