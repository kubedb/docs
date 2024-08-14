---
title: PgpoolAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: pp-autoscaler-concepts
    name: PgpoolAutoscaler
    parent: pp-concepts-pgpool
    weight: 26
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PgpoolAutoscaler

## What is PgpoolAutoscaler

`PgpoolAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [Pgpool](https://pgpool.net/mediawiki/index.php/Main_Page) compute resources of Pgpool components in a Kubernetes native way.

## PgpoolAutoscaler CRD Specifications

Like any official Kubernetes resource, a `PgpoolAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `PgpoolAutoscaler` CROs for autoscaling different components of pgpool is given below:

**Sample `PgpoolAutoscaler` for pgpool:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PgpoolAutoscaler
metadata:
  name: pgpool-auto-scale
  namespace: demo
spec:
  databaseRef:
    name: pgpool
  compute:
    pgpool:
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
```

Here, we are going to describe the various sections of a `PgpoolAutoscaler` crd.

A `PgpoolAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Pgpool](/docs/guides/pgpool/concepts/pgpool.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Pgpool](/docs/guides/pgpool/concepts/pgpool.md) object.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of Pgpool components. This field consists of the following sub-field:

- `trigger` indicates if compute autoscaling is enabled for this component of the pgpool. If "On" then compute autoscaling is enabled. If "Off" then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are "cpu" and "memory".
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are "RequestsAndLimits" and "RequestsOnly".
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.