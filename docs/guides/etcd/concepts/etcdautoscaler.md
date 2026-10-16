---
title: EtcdAutoscaler CRD
menu:
  docs_{{ .version }}:
    identifier: etcd-autoscaler-concepts
    name: EtcdAutoscaler
    parent: etcd-concepts-etcd
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# EtcdAutoscaler

## What is EtcdAutoscaler

`EtcdAutoscaler` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for autoscaling the compute resources and the storage of a KubeDB managed [etcd](https://etcd.io/) cluster in a Kubernetes native way.

An `EtcdAutoscaler` never mutates the `Etcd` object directly. It watches the resource usage of the members, computes a recommendation, and when that recommendation differs enough from the current setting it creates an [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) — a `VerticalScaling` one for compute, a `VolumeExpansion` one for storage — which then goes through the normal, quorum-aware ops-manager flow.

> `EtcdAutoscaler` scales members **vertically** (bigger members) and grows their **volumes**. It does not change the number of etcd members: quorum size is a deliberate design decision for a consensus cluster, not something to be adjusted automatically off a usage metric.

## EtcdAutoscaler CRD Specifications

Like any official Kubernetes resource, an `EtcdAutoscaler` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

**Sample `EtcdAutoscaler` for an etcd cluster:**

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: EtcdAutoscaler
metadata:
  name: etcd-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: etcd-quickstart
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    etcd:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 400m
        memory: 1Gi
      maxAllowed:
        cpu: 2
        memory: 4Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
  storage:
    etcd:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
      expansionMode: "Online"
      upperBound: 100Gi
```

Here, we are going to describe the various sections of an `EtcdAutoscaler` crd.

An `EtcdAutoscaler` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that points to the [Etcd](/docs/guides/etcd/concepts/etcd.md) object for which the autoscaling will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object.

### spec.opsRequestOptions

These are the options passed into the internally created [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md). It has the following fields:

- `timeout` specifies the timeout for each step of the generated ops request.
- `apply` controls whether the generated ops request runs only when the database is Ready (`IfReady`, the default) or regardless of the database state (`Always`).
- `maxRetries` is the number of times a failed generated ops request is retried. Defaults to `1`.

### spec.compute

`spec.compute` specifies the autoscaling configuration for the compute resources, i.e. cpu and memory, of the etcd members. This field consists of the following sub-fields:

- `spec.compute.etcd` indicates the desired compute autoscaling configuration for the `etcd` container.
- `spec.compute.nodeTopology` optionally references a `NodeTopology` object so the recommender can take the shape of the node pool into account when deciding whether to scale up.

`spec.compute.etcd` has the following sub-fields:

- `trigger` indicates if compute autoscaling is enabled for this component of the database. If `"On"` then compute autoscaling is enabled. If `"Off"` then compute autoscaling is disabled.
- `minAllowed` specifies the minimal amount of resources that will be recommended, default is no minimum.
- `maxAllowed` specifies the maximum amount of resources that will be recommended, default is no maximum.
- `controlledResources` specifies which type of compute resources (cpu and memory) are allowed for autoscaling. Allowed values are `"cpu"` and `"memory"`.
- `containerControlledValues` specifies which resource values should be controlled. Allowed values are `"RequestsAndLimits"` and `"RequestsOnly"`.
- `resourceDiffPercentage` specifies the minimum resource difference between the recommended value and the current value in percentage. If the difference percentage is greater than this value then autoscaling will be triggered.
- `podLifeTimeThreshold` specifies the minimum pod lifetime of at least one of the pods before triggering autoscaling.

> The generated `VerticalScaling` ops request carries only the container resources (`spec.verticalScaling.etcd.resources`) — the etcd vertical scaling API has no node-selection or topology sub-fields — so the autoscaler does not emit node-affinity or placement hints today.

### spec.storage

`spec.storage` specifies the autoscaling configuration for the storage resources of the etcd members. This field consists of the following sub-field:

- `spec.storage.etcd` indicates the desired storage autoscaling configuration for the etcd data volume.

It has the following sub-fields:

- `trigger` indicates if storage autoscaling is enabled for this component of the database. If `"On"` then storage autoscaling is enabled. If `"Off"` then storage autoscaling is disabled.
- `usageThreshold` indicates the usage percentage threshold; if the current storage usage exceeds it, storage autoscaling will be triggered. The default is 80%.
- `scalingThreshold` indicates the percentage of the current storage that will be added. The default is 50%.
- `scalingRules` allows a more dynamic scaling threshold — for example, grow by a percentage up to a certain size and by an absolute amount after that.
- `upperBound` sets a maximum size beyond which the volume will not be grown any further.
- `expansionMode` indicates the volume expansion mode, `Online` or `Offline`.

> Keep an eye on `spec.configuration.tuning.quotaBackendBytes` when you let the volume grow: etcd will not use more backend space than the quota allows, no matter how large the underlying volume becomes. See [Etcd CRD](/docs/guides/etcd/concepts/etcd.md#specconfiguration).

## Next Steps

- Learn about the [Etcd](/docs/guides/etcd/concepts/etcd.md) crd.
- Learn about the [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md) crd.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
