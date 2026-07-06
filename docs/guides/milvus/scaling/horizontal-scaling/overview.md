---
title: Milvus Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-scaling-horizontal-scaling-overview
    name: Overview
    parent: milvus-scaling-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling Milvus

This guide will give an overview on how the KubeDB Ops-manager operator horizontally scales a `Milvus` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

## How Horizontal Scaling Process Works

Horizontal scaling changes the **number of replicas** of the Milvus distributed roles.

> **Horizontal scaling is distributed-only.** A `Standalone` Milvus is a single all-in-one workload and cannot be horizontally scaled — there is exactly one PetSet with one replica. To distribute load horizontally, deploy Milvus in `Distributed` mode.

A `MilvusOpsRequest` of type `HorizontalScaling` carries the desired replica counts under `spec.horizontalScaling.topology`, keyed by role:

```yaml
spec:
  type: HorizontalScaling
  horizontalScaling:
    topology:
      proxy: 1
      streamingnode: 1
      # mixcoord / querynode / datanode are also supported by the API
```

The flow is:

1. A user creates a `MilvusOpsRequest` of type `HorizontalScaling`.
2. The operator validates the request and pauses the `Milvus` database.
3. The operator updates the replica counts on the `Milvus` object and the per-role PetSets, then adds or removes pods to match.
4. The operator waits for the affected roles to become healthy.
5. The operator resumes the database and marks the `MilvusOpsRequest` as `Successful`.

The `spec.horizontalScaling.topology` API accepts `proxy`, `mixcoord`, `querynode`, `streamingnode` and `dataNode`. The sample used in the guide only scales `proxy` and `streamingnode`; the other roles are scaled the same way.

In the next doc, we will see a step-by-step guide on horizontally scaling a distributed Milvus database.
