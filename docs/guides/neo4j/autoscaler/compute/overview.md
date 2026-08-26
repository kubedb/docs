---
title: Neo4j Compute Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-compute-autoscaling-overview
    name: Overview
    parent: neo4j-compute-autoscaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Compute Autoscaling

KubeDB can automatically adjust the CPU and memory assigned to Neo4j pods. A `Neo4jAutoscaler` observes real resource usage, generates recommendations, and creates a `Neo4jOpsRequest` when the recommended resources differ sufficiently from the current allocation.

## Before You Begin

You should be familiar with:

- [Neo4j](/docs/guides/neo4j/concepts/neo4j.md)
- [Neo4jAutoscaler](/docs/guides/neo4j/concepts/autoscaler.md)
- [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md)
- [Neo4j vertical scaling](/docs/guides/neo4j/scaling/vertical-scaling/overview.md)

## How Compute Autoscaling Works

<figure align="center">
  <img alt="Compute autoscaling process for Neo4j" src="/docs/images/neo4j/compute-autoscaling.png">
  <figcaption align="center">Fig: Neo4j compute autoscaling process</figcaption>
</figure>

1. The user creates a KubeDB `Neo4j` resource.
2. The Provisioner creates the Neo4j cluster and its supporting Kubernetes resources.
3. The user creates a `Neo4jAutoscaler` with a `spec.compute.neo4j` policy.
4. The Autoscaler creates and watches a Vertical Pod Autoscaler recommendation for the Neo4j container.
5. After the pod lifetime and resource-difference thresholds are satisfied, the Autoscaler creates a `Neo4jOpsRequest` of type `VerticalScaling`.
6. Ops Manager applies the recommendation and updates the Neo4j pods within the configured minimum and maximum bounds.

The next guide demonstrates this workflow end to end.
