---
title: Neo4j Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: neo4j-storage-autoscaling-overview
    name: Overview
    parent: neo4j-storage-autoscaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Storage Autoscaling

KubeDB can automatically expand Neo4j persistent volumes as graph data grows. The Autoscaler watches volume usage and creates a volume-expansion operation before the disks become full.

## Before You Begin

You should be familiar with:

- [Neo4j](/docs/guides/neo4j/concepts/neo4j.md)
- [Neo4jAutoscaler](/docs/guides/neo4j/concepts/autoscaler.md)
- [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md)
- [Neo4j volume expansion](/docs/guides/neo4j/volume-expansion/overview.md)

## How Storage Autoscaling Works

<figure align="center">
  <img alt="Storage autoscaling process for Neo4j" src="/docs/images/neo4j/storage-autoscaling.png">
  <figcaption align="center">Fig: Neo4j storage autoscaling process</figcaption>
</figure>

1. The user creates a KubeDB `Neo4j` resource with durable storage.
2. The Provisioner creates a persistent volume for every Neo4j pod.
3. The user creates a `Neo4jAutoscaler` with a `spec.storage.neo4j` policy.
4. The Autoscaler reads PVC usage from the KubeDB storage metrics API.
5. When usage reaches `usageThreshold`, the Autoscaler calculates a larger size and creates a `Neo4jOpsRequest` of type `VolumeExpansion`.
6. Ops Manager expands the PVCs using the configured online or offline mode.

> Volume expansion requires a StorageClass with `allowVolumeExpansion: true`. Kubernetes does not support shrinking a PVC after it has been expanded.

The next guide demonstrates this workflow using actual Neo4j graph data.
