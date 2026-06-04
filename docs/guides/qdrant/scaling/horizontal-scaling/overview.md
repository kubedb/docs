---
title: Qdrant Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-horizontal-scaling-overview
    name: Overview
    parent: qdrant-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling Overview

This guide will give you an overview of how `KubeDB` Ops Manager scales up/down the number of members of a `Qdrant`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)

## How Horizontal Scaling Process Works

The following diagram shows how `KubeDB` Ops Manager used to scale up the number of members of a `Qdrant` cluster. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Horizontal scaling Flow" src="/docs/guides/qdrant/images/qdrant-horizontal-scaling.png">
<figcaption align="center">Fig: Horizontal scaling process of Qdrant</figcaption>
</figure>

The horizontal scaling process consists of the following steps:

1. At first, a user creates a `Qdrant` CR.

2. `KubeDB` provisioner operator watches for the `Qdrant` CR.

3. When it finds one, it creates a `PetSet` and related necessary stuff like secret, service, etc.

4. Then, in order to scale the cluster up or down, the user creates a `QdrantOpsRequest` CR with the desired number of replicas after scaling.

5. `KubeDB` Ops Manager watches for `QdrantOpsRequest`.

6. When it finds one, it halts the `Qdrant` object so that the `KubeDB` provisioner operator doesn't perform any operation on the `Qdrant` during the scaling process.

7. Then `KubeDB` Ops Manager will add nodes in case of scale up or remove nodes in case of scale down.

8. Then the `KubeDB` Ops Manager will scale the PetSet replicas to reach the expected number of replicas for the cluster.

9. After successful scaling of the PetSet's replica, the `KubeDB` Ops Manager updates the `spec.replicas` field of `Qdrant` object to reflect the updated cluster state.

10. After successful scaling of the `Qdrant` replicas, the `KubeDB` Ops Manager resumes the `Qdrant` object so that the `KubeDB` provisioner operator can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on scaling of a Qdrant cluster using Horizontal Scaling.
