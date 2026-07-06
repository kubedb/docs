---
title: Weaviate Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-vertical-scaling-overview
    name: Overview
    parent: weaviate-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scaling Weaviate

This guide will give you an overview of how KubeDB Ops Manager updates the resources (for example Memory, CPU etc.) of a `Weaviate` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Weaviate Quickstart](/docs/guides/weaviate/quickstart/quickstart.md)

## How Vertical Scaling Process Works

The vertical scaling process consists of the following steps:

1. At first, a user creates a `Weaviate` CR.

2. `KubeDB` provisioner operator watches for the `Weaviate` CR.

3. When the operator finds a `Weaviate` CR, it creates a `PetSet` and related necessary resources like secret, service, etc.

4. Then, in order to update the resources (for example `CPU`, `Memory` etc.) of the `Weaviate` cluster, the user creates a `WeaviateOpsRequest` CR with the desired resources.

5. `KubeDB` Ops Manager watches for the `WeaviateOpsRequest` CR.

6. When it finds one, it halts the `Weaviate` object so that the `KubeDB` provisioner operator doesn't perform any operation on the `Weaviate` during the scaling process.

7. Then the KubeDB Ops-manager operator updates the resources of the PetSet's Pods to reach the desired state, restarting the pods one at a time.

8. After successfully updating the resources of the PetSet's Pods, the `KubeDB` Ops Manager updates the `Weaviate` object resources to reflect the updated state.

9. After successfully updating the `Weaviate` resources, the `KubeDB` Ops Manager resumes the `Weaviate` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on updating the resources of a Weaviate database using the vertical scaling operation.
