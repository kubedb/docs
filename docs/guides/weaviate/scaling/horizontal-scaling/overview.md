---
title: Weaviate Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-horizontal-scaling-overview
    name: Overview
    parent: weaviate-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling Weaviate

This guide will give you an overview of how KubeDB Ops Manager scales the number of nodes (replicas) of a `Weaviate` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Weaviate Quickstart](/docs/guides/weaviate/quickstart/quickstart.md)

## How Horizontal Scaling Process Works

The horizontal scaling process consists of the following steps:

1. At first, a user creates a `Weaviate` CR.

2. `KubeDB` provisioner operator watches for the `Weaviate` CR and creates a `PetSet` and related necessary resources.

3. Then, in order to scale the number of nodes of the `Weaviate` cluster, the user creates a `WeaviateOpsRequest` CR with the desired node count.

4. `KubeDB` Ops Manager watches for the `WeaviateOpsRequest` CR.

5. When it finds one, it halts the `Weaviate` object so that the `KubeDB` provisioner operator doesn't perform any operation on the `Weaviate` during the scaling process.

6. To **scale up**, the Ops Manager increases the replica count of the PetSet, waits for the new nodes to join the cluster and sync the schema, and then rebalances the shard replicas across the enlarged cluster.

7. To **scale down**, the Ops Manager rebalances shard replicas off the nodes that are going away, then decreases the replica count of the PetSet and removes the surplus nodes.

8. After successfully scaling, the `KubeDB` Ops Manager updates the `Weaviate` object's replica count to reflect the updated state and resumes the `Weaviate` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on scaling a Weaviate cluster horizontally.
