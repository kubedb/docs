---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: sl-solrautoscaler-solr
    name: Autoscaler
    parent: sl-concepts-solr
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SolrAutoscaler

## What is SolrAutoscaler

`SolrAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [Solr](https://solr.apache.org/guide/solr/latest/index.html) compute resources and storage of database components in a Kubernetes native way.

## SolrAutoscaler CRD Specifications

Like any official Kubernetes resource, a `SolrAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `SolrAutoscaler` CROs for autoscaling different components of database is given below:

**Sample `SolrAutoscaler` YAML for an Solr combined cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SolrAutoscaler
metadata:
  name: sl-as
  namespace: demo
spec:
  databaseRef:
    name: solr-combined
  opsRequestOptions:
    timeout: 3m
    apply: IfReady
  compute:
    node:
      trigger: "On"
      podLifeTimeThreshold: 24h
      minAllowed:
        cpu: 1
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
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

**Sample `SolrAutoscaler` YAML for the Solr topology cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: SolrAutoscaler
metadata:
  name: sl-as-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  compute:
    master:
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
    data:
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
    ingest:
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
    data:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here, we are going to describe the various sections of a `SolrAutoscaler` crd.

A `SolrAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a `required` field that point to the [Solr](/docs/guides/solr/concepts/solr.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Solr](/docs/guides/solr/concepts/solr.md) object.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.node` indicates the desired compute autoscaling configuration for a combined Solr cluster.
- `spec.compute.overseer` indicates the desired compute autoscaling configuration for overseer nodes.
- `spec.compute.data` indicates the desired compute autoscaling configuration for data nodes.
- `spec.compute.coordinator` indicates the desired compute autoscaling configuration for coordinator nodes.

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

- `spec.storage.node` indicates the desired storage autoscaling configuration for a combined Solr cluster.
- `spec.storage.topology` indicates the desired storage autoscaling configuration for different type of nodes running in the Solr topology cluster mode.
- `spec.storage.overseer` indicates the desired storage autoscaling configuration for the overseer nodes.
- `spec.storage.data` indicates the desired storage autoscaling configuration for the data nodes.
- `spec.storage.coordinator` indicates the desired storage autoscaling configuration for the coordinator nodes.

All of them has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
