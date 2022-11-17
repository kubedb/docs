---
title: ProxySQLAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-concepts-autoscaler
    name: ProxySQLAutoscaler
    parent: guides-proxysql-concepts
    weight: 26
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# ProxySQLAutoscaler

## What is ProxySQLAutoscaler

`ProxySQLAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [ProxySQL](https://www.proxysql.com/) compute resources and storage of database components in a Kubernetes native way.

## ProxySQLAutoscaler CRD Specifications

Like any official Kubernetes resource, a `ProxySQLAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `ProxySQLAutoscaler` CROs for autoscaling different components of database is given below:

**Sample `ProxySQLAutoscaler` for ProxySQL:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ProxySQLAutoscaler
metadata:
  name: psops-autoscale
  namespace: demo
spec:
  proxyRef:
    name: sample-proxysql
  compute:
    proxysql:
      trigger: "On"
      podLifeTimeThreshold: 5m
      minAllowed:
        cpu: 250m
        memory: 350Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
```

Here, we are going to describe the various sections of a `ProxySQLAutoscaler` crd.

A `ProxySQLAutoscaler` object has the following fields in the `spec` section.

### spec.proxyRef

`spec.proxyRef` is a required field that point to the [ProxySQL](/docs/guides/proxysql/concepts/proxysql) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.proxyRef.name :** specifies the name of the [ProxySQL](/docs/guides/proxysql/concepts/proxysql) object.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of the proxysql components. This field consists of the following sub-field:

- `spec.compute.proxysql` indicates the desired compute autoscaling configuration for a ProxySQL standalone or cluster.

All of them has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for this component of the database. If "On" then compute autoscaling is enabled. If "Off" then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are "cpu" and "memory".
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are "RequestsAndLimits" and "RequestsOnly".
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.
- `InMemoryScalingThreshold` the percentage of the Memory that will be passed as inMemorySizeGB for inmemory database engine, which is only available for the percona variant of the proxysql.