---
title: Weaviate Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-volume-expansion-overview
    name: Overview
    parent: weaviate-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Volume Expansion

This guide will give you an overview of how KubeDB Ops Manager expands the volume of a `Weaviate` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Weaviate Quickstart](/docs/guides/weaviate/quickstart/quickstart.md)

## How Volume Expansion Process Works

The volume expansion process consists of the following steps:

1. At first, a user creates a `Weaviate` CR.

2. `KubeDB` provisioner operator watches for the `Weaviate` CR.

3. When the operator finds a `Weaviate` CR, it creates a `PetSet` and related necessary resources, and provisions a `PersistentVolumeClaim` (PVC) for each node.

4. Then, in order to expand the volume of the `Weaviate` cluster, the user creates a `WeaviateOpsRequest` CR with the desired volume size.

5. `KubeDB` Ops Manager watches for the `WeaviateOpsRequest` CR.

6. When it finds one, it halts the `Weaviate` object so that the `KubeDB` provisioner operator doesn't perform any operation on the `Weaviate` during the volume expansion process.

7. Then the `KubeDB` Ops Manager expands the PVCs to reach the desired size. Volume expansion requires a `StorageClass` that supports volume expansion (`allowVolumeExpansion: true`).

8. After successfully expanding the PVCs, the `KubeDB` Ops Manager updates the `Weaviate` object's storage to reflect the updated state.

9. After successfully updating the storage, the `KubeDB` Ops Manager resumes the `Weaviate` object so that the `KubeDB` Provisioner operator resumes its usual operations.

Volume expansion can be performed in two modes:

- **Online** — the volume is expanded without restarting the pods (requires a CSI driver that supports online expansion).
- **Offline** — the pods are recreated to apply the expanded volume.

In the next doc, we are going to show a step-by-step guide on expanding the volume of a Weaviate cluster using the volume expansion operation.
