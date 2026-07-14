---
title: ProxySQL Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-scaling-vertical-overview
    name: Overview
    parent: guides-proxysql-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ProxySQL Vertical Scaling

This guide will give an overview on how KubeDB Ops Manager vertically scales up `ProxySQL`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/)
  - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest/)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops Manager scales up or down `ProxySQL` instance resources. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of ProxySQL" src="/docs/guides/proxysql/scaling/vertical-scaling/overview/images/vertical-scaling.png">
<figcaption align="center">Fig: Vertical scaling process of ProxySQL</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `ProxySQL` Custom Resource (CR).

2. `KubeDB` Community operator watches the `ProxySQL` CR.

3. When the operator finds a `ProxySQL` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `ProxySQL` the user creates a `ProxySQLOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `ProxySQLOpsRequest` CR.

6. When it finds a `ProxySQLOpsRequest` CR, it halts the `ProxySQL` object which is referred from the `ProxySQLOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `ProxySQL` object during the vertical scaling process.  

7. Then the `KubeDB` Enterprise operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Enterprise operator updates the `ProxySQL` object to reflect the updated state.

9. After the successful update  of the `ProxySQL` resources, the `KubeDB` Enterprise operator resumes the `ProxySQL` object so that the `KubeDB` Community operator resumes its usual operations.

## Vertical Scaling Modes

KubeDB actuates vertical scaling in one of two modes, selected through the `spec.verticalScaling.mode`
field of the `ProxySQLOpsRequest`:

- **`Restart`** (default): The operator patches the `PetSet` with the new resources and restarts the
  Pods (one at a time, honoring the database's failover rules) so they come back with the updated CPU
  and Memory. This works on every Kubernetes cluster.
- **`InPlace`**: The operator resizes the running containers in place using the Kubernetes
  [in-place Pod resize](https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/)
  (`pods/resize` subresource) — no Pod restart, so scaling happens without downtime or failover. If a
  Node cannot accommodate the new resources (the resize is reported `Infeasible`), the operator
  automatically falls back to the `Restart` behavior for that Pod.

If `spec.verticalScaling.mode` is omitted, it defaults to `Restart`.

> **Note:** `InPlace` mode relies on the Kubernetes `InPlacePodVerticalScaling` feature gate, which is
> enabled by default from Kubernetes v1.33. On older clusters, or when the feature gate is disabled,
> use `Restart` mode.

In the next docs, we are going to show a step by step guide on updating resources of ProxySQL using `ProxySQLOpsRequest` CRD.