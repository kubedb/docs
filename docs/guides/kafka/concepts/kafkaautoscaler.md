---
title: KafkaAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: kf-autoscaler-concepts
    name: KafkaAutoscaler
    parent: kf-concepts-kafka
    weight: 26
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KafkaAutoscaler

## What is KafkaAutoscaler

`KafkaAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [Kafka](https://kafka.apache.org/) compute resources and storage of database components in a Kubernetes native way.

## KafkaAutoscaler CRD Specifications

Like any official Kubernetes resource, a `KafkaAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `KafkaAutoscaler` CROs for autoscaling different components of database is given below:

**Sample `KafkaAutoscaler` for combined cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kf-autoscaler-combined
  namespace: demo
spec:
  databaseRef:
    name: kafka-dev
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 24h
      minAllowed:
        cpu: 250m
        memory: 350Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
      resourceDiffPercentage: 10
  storage:
    node:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

**Sample `KafkaAutoscaler` for topology cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kf-autoscaler-topology
  namespace: demo
spec:
  databaseRef:
    name: kafka-prod
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    broker:
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
    controller:
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
    broker:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
    controller:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here, we are going to describe the various sections of a `KafkaAutoscaler` crd.

A `KafkaAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Kafka](/docs/guides/kafka/concepts/kafka.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Kafka](/docs/guides/kafka/concepts/kafka.md) object.

### spec.opsRequestOptions
These are the options to pass in the internally created opsRequest CRO. `opsRequestOptions` has two fields.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.node` indicates the desired compute autoscaling configuration for a combined Kafka cluster.
- `spec.compute.broker` indicates the desired compute autoscaling configuration for broker of a topology Kafka database.
- `spec.compute.controller` indicates the desired compute autoscaling configuration for controller of a topology Kafka database.


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

`spec.compute` specifies the autoscaling configuration for the storage resources of the database components. This field consists of the following sub-field:

- `spec.compute.node` indicates the desired storage autoscaling configuration for a combined Kafka cluster.
- `spec.compute.broker` indicates the desired storage autoscaling configuration for broker of a combined Kafka cluster.
- `spec.compute.controller` indicates the desired storage autoscaling configuration for controller of a topology Kafka cluster.


All of them has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
- `expansionMode` indicates the volume expansion mode.
