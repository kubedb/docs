---
title: Etcd Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-autoscaling-compute-overview
    name: Overview
    parent: etcd-autoscaling-compute
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Compute Resource Autoscaling

This guide gives an overview of how the KubeDB Autoscaler operator autoscales the compute
resources (cpu and memory) of an etcd cluster using the `EtcdAutoscaler` CRD.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdAutoscaler](/docs/guides/etcd/concepts/etcdautoscaler.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## How Compute Autoscaling Works

The autoscaling process consists of the following steps:

1. A user creates an `Etcd` Custom Resource Object (CRO).

2. The `KubeDB` Provisioner operator watches the `Etcd` CRO.

3. When the operator finds an `Etcd` CRO, it creates the `PetSet` and the related resources
   (Secrets, Services, RBAC, and so on) that back the etcd members.

4. To set up autoscaling for that cluster, the user creates an `EtcdAutoscaler` CRO with the
   desired configuration.

5. The `KubeDB` Autoscaler operator watches the `EtcdAutoscaler` CRO.

6. The `KubeDB` Autoscaler operator generates a recommendation for the `etcd` container using a
   modified version of the Kubernetes
   [official VPA recommender](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/recommender).
   The recommender keeps a decaying CPU and memory histogram per container and reports the
   recommendation under `status.vpas[].recommendation` on the `EtcdAutoscaler`.

7. If the generated recommendation does not match the current resources of the etcd container,
   the `KubeDB` Autoscaler operator creates an `EtcdOpsRequest` of type `VerticalScaling` to
   move the cluster to the recommended resources.

8. The `KubeDB` Ops-manager operator watches that `EtcdOpsRequest` CRO.

9. The `KubeDB` Ops-manager operator then scales the etcd container vertically, as specified in
   the `EtcdOpsRequest` CRO.

## The generated EtcdOpsRequest

This generation relationship is the important part to understand: **the autoscaler never edits
the `Etcd` object directly.** Every change it wants to make is expressed as an
`EtcdOpsRequest`, which is then executed by the Ops-manager operator exactly as if you had
written it by hand. Concretely, the autoscaler:

- Names the request with the `etcdops-` prefix plus a random suffix, for example
  `etcdops-etcd-autoscale-vft8xm`.
- Sets an owner reference back to the `EtcdAutoscaler`, so `kubectl describe` on the request
  shows which autoscaler produced it, and deleting the autoscaler garbage-collects it.
- Sets `spec.type: VerticalScaling` and fills in `spec.verticalScaling.etcd.resources` with the
  recommendation, merged with the resources already set on the `etcd` container of the database.
- Copies `spec.timeout`, `spec.apply` and `spec.maxRetries` from
  `EtcdAutoscaler.spec.opsRequestOptions`, when that block is present.

Two behavioral details follow from the implementation:

- Only `spec.verticalScaling.etcd` is ever populated. `EtcdVerticalScalingSpec` also has an
  `exporter` field for structural parity with the other KubeDB databases, but etcd exposes its
  Prometheus metrics natively on its own metrics listener and KubeDB deploys **no exporter
  sidecar** for it — so there is no exporter container for the autoscaler to size.
- A new request is only created when its `spec` differs from the last request the same
  autoscaler created. An identical recommendation does not produce a stream of duplicate
  requests.

## A note on node topology

`EtcdAutoscaler.spec.compute.nodeTopology` lets the recommendation be snapped to the node groups
(or machine profiles) of a `NodeTopology` object, so the autoscaler will not recommend a size no
node in the pool can actually accommodate.

Be aware of the limitation here, though: `EtcdVerticalScalingSpec.Etcd` is an
`ContainerResources` — unlike the `PodResources` used by some other KubeDB databases, it carries
no node-selection or topology fields. So **the generated `EtcdOpsRequest` does not carry
placement hints in its spec.** When `nodeTopology` is configured, the chosen node group is only
communicated through the machine-profile annotation the autoscaler stamps onto the ops request's
metadata. Do not expect node-affinity-aware placement of the etcd pods to be driven from the
generated request itself; this is a known apimachinery-level gap.

In the next doc, we show a step-by-step guide for autoscaling the compute resources of an etcd
cluster using the `EtcdAutoscaler` CRD.

## Next Steps

- [Autoscale the compute resources of an Etcd cluster](/docs/guides/etcd/autoscaler/compute/compute-autoscale.md).
- [Storage Autoscaling Overview](/docs/guides/etcd/autoscaler/storage/overview.md).
