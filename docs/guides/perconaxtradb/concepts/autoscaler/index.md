---
title: PerconaXtraDBAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-concepts-autoscaler
    name: PerconaXtraDBAutoscaler
    parent: guides-perconaxtradb-concepts
    weight: 26
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# PerconaXtraDBAutoscaler

## What is PerconaXtraDBAutoscaler

`PerconaXtraDBAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [PerconaXtraDB](https://www.perconaxtradb.com/) compute resources and storage of database components in a Kubernetes native way.

## PerconaXtraDBAutoscaler CRD Specifications

Like any official Kubernetes resource, a `PerconaXtraDBAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `PerconaXtraDBAutoscaler` CROs for autoscaling different components of database is given below:

**Sample `PerconaXtraDBAutoscaler` for PerconaXtraDB:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PerconaXtraDBAutoscaler
metadata:
  name: md-as
  namespace: demo
spec:
  databaseRef:
    name: sample-pxc
  compute:
    perconaxtradb:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 250m
        memory: 350Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
  storage:
    perconaxtradb:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
      expansionMode: "Online"
```

Here, we are going to describe the various sections of a `PerconaXtraDBAutoscaler` crd.

A `PerconaXtraDBAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [PerconaXtraDB](/docs/guides/perconaxtradb/concepts/perconaxtradb) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [PerconaXtraDB](/docs/guides/perconaxtradb/concepts/perconaxtradb) object.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.perconaxtradb` indicates the desired compute autoscaling configuration for a PerconaXtraDB standalone or cluster.

All of them has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for this component of the database. If "On" then compute autoscaling is enabled. If "Off" then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are "cpu" and "memory".
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are "RequestsAndLimits" and "RequestsOnly".
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.
- `InMemoryScalingThreshold` the percentage of the Memory that will be passed as inMemorySizeGB for inmemory database engine, which is only available for the percona variant of the perconaxtradb.

### spec.storage

`spec.compute` specifies the autoscaling configuration for the storage resources of the database components. This field consists of the following sub-field:

- `spec.compute.mairadb` indicates the desired storage autoscaling configuration for a PerconaXtraDB standalone or cluster.

All of them has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
- `expansionMode` specifies the mode of volume expansion when storage autoscaler performs volume expansion OpsRequest. Default value is `Online`.

