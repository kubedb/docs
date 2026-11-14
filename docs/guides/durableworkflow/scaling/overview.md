---
title: Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: dw-scaling-overview
    name: Overview
    parent: dw-scaling-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Scaling Overview

The roles of a DurableWorkflow have no volumes, so they can be scaled in both directions. Neither horizontal nor vertical scaling causes downtime on a highly available cluster. Both were verified with 0 failed calls (horizontal up and down, and vertical).

## Horizontal scaling

A `HorizontalScaling` request sets the replicas of the roles you name. Each value is at least 1.

- **Scale up** is done at once for the role.
- **Scale down** removes **one pod at a time** and waits for the cluster to settle after each pod, because every pod that leaves moves history shards or task queues to the others. For `history`, the minimum interval between pods (`rollout.historyMinInterval`, 60 seconds by default) is respected.
- `internalFrontend` can be scaled for Temporal only.
- The number of history shards is fixed. A `history` role with more replicas than `spec.numHistoryShards` has pods that own no shard.

## Vertical scaling

A `VerticalScaling` request changes the resources of the roles you name. The pods are [rolled one at a time](/docs/guides/durableworkflow/update-version/overview.md#how-the-pods-are-rolled), with a surge pod for every role but `history`.

## Next Steps

- [Horizontal scaling](/docs/guides/durableworkflow/scaling/horizontal-scaling.md) and [vertical scaling](/docs/guides/durableworkflow/scaling/vertical-scaling.md).
- [Autoscale](/docs/guides/durableworkflow/autoscaler/autoscaler.md) the resources.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
