---
title: RedisAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: rd-autoscaler-concepts
    name: RedisAutoscaler
    parent: rd-concepts-redis
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# RedisAutoscaler

## What is RedisAutoscaler

`RedisAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [Redis](https://www.redis.io/) compute resources and storage of database components in a Kubernetes native way.

## RedisAutoscaler CRD Specifications

Like any official Kubernetes resource, a `RedisAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here is a sample `RedisAutoscaler` CRDs for autoscaling different components of database is given below:

**Sample `RedisAutoscaler` for standalone database:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RedisAutoscaler
metadata:
  name: standalone-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: redis-standalone
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
  compute:
    standalone: 
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 600m
        memory: 600Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
      resourceDiffPercentage: 10
  storage:
    standalone:
      trigger: "On"
      usageThreshold: 25
      scalingThreshold: 20
```

Here is a sample `RedisSentinelAutoscaler` CRDs for autoscaling different components of database is given below:

**Sample `RedisSentinelAutoscaler` for standalone database:**
```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RedisSentinelAutoscaler
metadata:
  name: sentinel-autoscalar
  namespace: demo
spec:
  databaseRef:
    name: sentinel
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
  compute:
    sentinel: 
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 600m
        memory: 600Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
      resourceDiffPercentage: 10
```

Here, we are going to describe the various sections of a `RedisAutoscaler` and `RedisSentinelAutoscaler`  crd.

A `RedisAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Redis](/docs/guides/redis/concepts/redis.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Redis](/docs/guides/redis/concepts/redis.md) object.

### spec.opsRequestOptions
These are the options to pass in the internally created opsRequest CRD. `opsRequestOptions` has three fields. They have been described in details [here](/docs/guides/redis/concepts/redisopsrequest.md#specreadinesscriteria).

### spec.compute

`spec.compute` specifies the autoscaling configuration for to compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.standalone` indicates the desired compute autoscaling configuration for a standalone mode in Redis database.
- `spec.compute.cluster` indicates the desired compute autoscaling configuration for cluster mode in Redis database.
- `spec.compute.sentinel` indicates the desired compute autoscaling configuration for sentinel mode in Redis database.

`RedisSentinelAutoscaler` on has only `spec.compute.sentinel` field.

All of them has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for this component of the database. If "On" then compute autoscaling is enabled. If "Off" then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are "cpu" and "memory".
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are "RequestsAndLimits" and "RequestsOnly".
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.

### spec.storage

`spec.storage` specifies the autoscaling configuration for the storage resources of the database components. This field consists of the following sub-field:

- `spec.storage.standalone` indicates the desired storage autoscaling configuration for a standalone mode in Redis database.
- `spec.storage.cluster` indicates the desired storage autoscaling configuration for cluster mode in Redis database.
- `spec.storage.sentinel` indicates the desired storage autoscaling configuration for sentinel mode in Redis database.

`RedisSentinelAutoscaler` does not have `spec.stoage` section. 

All of them has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
- `expansionMode` indicates the volume expansion mode.

## Next Steps

- Learn about Redis crd [here](/docs/guides/redis/concepts/redis.md).
- Deploy your first Redis database with Redis by following the guide [here](/docs/guides/redis/quickstart/quickstart.md).
