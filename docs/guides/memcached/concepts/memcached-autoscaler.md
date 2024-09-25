---
title: Autoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: mc-autoscaler-concepts
    name: Autoscaler
    parent: mc-concepts-memcached
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MemcachedAutoscaler

## What is MemcachedAutoscaler

`MemcachedAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling Memcached compute resources in a Kubernetes native way.

## MemcachedAutoscaler CRD Specifications

Like any official Kubernetes resource, a `MemcachedAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here is a sample `MemcachedAutoscaler` CRO for autoscaling different components of database is given below.
Sample `MemcachedAutoscaler`:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MemcachedAutoscaler
metadata:
  name: mc-as
  namespace: demo
spec:
  databaseRef:
    name: mc1
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    memcached:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 400m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here, we are going to describe the various sections of a `MemcachedAutoscaler` crd. A `MemcachedAutoscaler` object has the following fields in the spec section.

### spec.databaseRef
spec.databaseRef is a required field that point to the Memcached object for which the autoscaling will be performed. This field consists of the following sub-field:

##### 

### spec.opsRequestOptions
These are the options to pass in the internally created opsRequest CRO. opsRequestOptions has three fields. They have been described in details [here](/docs/guides/memcached/concepts/memcached-opsrequest.md).

### spec.compute
`spec.compute` specifies the autoscaling configuration for to compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.memcached` indicates the desired compute autoscaling 


`spec.compute.memcached` has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for this component of the database. If “On” then compute autoscaling is enabled. If “Off” then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are “cpu” and “memory”.
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are “RequestsAndLimits” and “RequestsOnly”.
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.

## Next Steps

- Learn about Memcached crd [here](/docs/guides/memcached/concepts/memcached.md).
- Deploy your first Memcached database with Memcached by following the guide [here](/docs/guides/memcached/quickstart/quickstart.md).