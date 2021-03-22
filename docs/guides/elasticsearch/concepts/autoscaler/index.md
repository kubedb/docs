---
title: ElasticsearchAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: es-autoscaler-concepts
    name: ElasticsearchAutoscaler
    parent: es-concepts-elasticsearch
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# ElasticsearchAutoscaler

## What is ElasticsearchAutoscaler

`ElasticsearchAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) compute resources and storage of database components in a Kubernetes native way.

## ElasticsearchAutoscaler CRD Specifications

Like any official Kubernetes resource, a `ElasticsearchAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `ElasticsearchAutoscaler` CROs for autoscaling different components of database is given below:

**Sample `ElasticsearchAutoscaler` YAML for the Elasticsearch combined cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ElasticsearchAutoscaler
metadata:
  name: es-as
  namespace: demo
spec:
  databaseRef:
    name: es-combined
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
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

**Sample `ElasticsearchAutoscaler` YAML for the Elasticsearch topology cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ElasticsearchAutoscaler
metadata:
  name: mg-as-topology
  namespace: demo
spec:
  databaseRef:
    name: es-topology
  compute:
    topology:
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
    topology:
      data:
        trigger: "On"
        usageThreshold: 60
        scalingThreshold: 50
```

Here, we are going to describe the various sections of a `ElasticsearchAutoscaler` crd.

A `ElasticsearchAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a `required` field that point to the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) object.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources i.e. cpu and memory of the database components. This field consists of the following sub-field:

- `spec.compute.node` indicates the desired compute autoscaling configuration for a combined Elasticsearch cluster.
- `spec.compute.topology` indicates the desired compute autoscaling configuration for different type of nodes running in the Elasticsearch topology cluster mode.
  - `topology.master` indicates the desired compute autoscaling configuration for master nodes.
  - `topology.data` indicates the desired compute autoscaling configuration for data nodes.
  - `topology.ingest` indicates the desired compute autoscaling configuration for ingest nodes.

All of them has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for this component of the database. If "On" then compute autoscaling is enabled. If "Off" then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are "cpu" and "memory".
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are "RequestsAndLimits" and "RequestsOnly".
- `resourceDiffPercentage` specifies the minimum resource difference between recommended value and the current value in percentage. If the difference percentage is greater than this value than autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.

### spec.storage

`spec.compute` specifies the autoscaling configuration for the storage resources of the database components. This field consists of the following sub-field:

- `spec.compute.node` indicates the desired storage autoscaling configuration for a combined Elasticsearch cluster.
- `spec.compute.topology` indicates the desired storage autoscaling configuration for different type of nodes running in the Elasticsearch topology cluster mode.
  - `topology.master` indicates the desired storage autoscaling configuration for the master nodes.
  - `topology.data` indicates the desired storage autoscaling configuration for the data nodes.
  - `topology.ingest` indicates the desired storage autoscaling configuration for the ingest nodes.

All of them has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If "On" then storage autoscaling is enabled. If "Off" then storage autoscaling is disabled.
- `usageThreshold` indicates usage percentage threshold, if the current storage usage exceeds then storage autoscaling will be triggered.
- `scalingThreshold` indicates the percentage of the current storage that will be scaled.
