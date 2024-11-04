---
title: MSSQLServerAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: ms-concepts-autoscaler
    name: MSSQLServerAutoscaler
    parent: ms-concepts
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServerAutoscaler

## What is MSSQLServerAutoscaler

`MSSQLServerAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [Microsoft SQL Server](https://learn.microsoft.com/en-us/sql/sql-server/) compute resources and storage of database in a Kubernetes native way.

## MSSQLServerAutoscaler CRD Specifications

Like any official Kubernetes resource, a `MSSQLServerAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here is a sample `MSSQLServerAutoscaler` CRO for autoscaling is given below:

**Sample `MSSQLServerAutoscaler` for mssqlserver database:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MSSQLServerAutoscaler
metadata:
  name: standalone-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: mssqlserver-standalone
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
  compute:
    mssqlserver:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 800m
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 4Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
      resourceDiffPercentage: 10
  storage:
    mssqlserver:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here, we are going to describe the various sections of a `MSSQLServerAutoscaler` CRD.

A `MSSQLServerAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) object.

### spec.opsRequestOptions
These are the options to pass in the internally created opsRequest CRO. `opsRequestOptions` has two fields. They have been described in details [here](/docs/guides/mssqlserver/concepts/opsrequest.md#spectimeout).

### spec.compute

`spec.compute` specifies the autoscaling configuration for to compute resources i.e. cpu and memory of the database. This field consists of the following sub-field:

- `spec.compute.mssqlserver` indicates the desired compute autoscaling configuration for a MSSQLServer database.

This has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for the database. If "On" then compute autoscaling is enabled. If "Off" then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are "cpu" and "memory".
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are "RequestsAndLimits" and "RequestsOnly".
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.

### spec.storage

`spec.storage` specifies the autoscaling configuration for the storage resources of the database. This field consists of the following sub-field:

- `spec.storage.mssqlserver` indicates the desired storage autoscaling configuration for a MSSQLServer database.

 It has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
- `expansionMode` indicates the volume expansion mode.

## Next Steps

- Learn about [backup and restore](/docs/guides/mssqlserver/backup/overview/index.md) SQL Server using KubeStash.
- Learn about MSSQLServer CRD [here](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Deploy your first MSSQLServer database with MSSQLServer by following the guide [here](/docs/guides/mssqlserver/quickstart/quickstart.md).
