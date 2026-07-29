---
title: Qdrant Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-vertical-scaling-overview
    name: Overview
    parent: qdrant-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling Qdrant

This guide will give you an overview of how KubeDB Ops Manager updates the resources(for example Memory, CPU etc.) of the `Qdrant`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how the `KubeDB` Ops Manager used to update the resources of the `Qdrant`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of Qdrant" src="/docs/guides/qdrant/images/qdrant-vertical-scaling.png">
<figcaption align="center">Fig: Vertical scaling process of Qdrant</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `Qdrant` CR.

2. `KubeDB` provisioner operator watches for the `Qdrant` CR.

3. When the operator finds a `Qdrant` CR, it creates a `PetSet` and related necessary stuff like secret, service, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `Qdrant` cluster the user creates a `QdrantOpsRequest` CR with desired information.

5. `KubeDB` Ops Manager watches for `QdrantOpsRequest`.

6. When it finds one, it halts the `Qdrant` object so that the `KubeDB` provisioner operator doesn't perform any operation on the `Qdrant` during the scaling process.

7. Then the KubeDB Ops-manager operator will update resources of the PetSet's Pods to reach desired state.

8. After successful updating of the resources of the PetSet's Pods, the `KubeDB` Ops Manager updates the `Qdrant` object resources to reflect the updated state.

9. After successful updating of the `Qdrant` resources, the `KubeDB` Ops Manager resumes the `Qdrant` object so that the `KubeDB` Provisioner operator resumes its usual operations.

## Vertical Scaling Modes

KubeDB actuates vertical scaling in one of two modes, selected through the `spec.verticalScaling.mode`
field of the `QdrantOpsRequest`:

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

In the next doc, we are going to show a step-by-step guide on updating resources of Qdrant database using vertical scaling operation.
