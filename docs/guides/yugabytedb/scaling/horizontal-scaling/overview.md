---
title: YugabyteDB Horizontal Scaling
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-horizontal-scaling-overview
    name: Overview
    parent: yugabytedb-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Horizontal Scaling

Only the tservers can be scaled. The number of masters is the replication factor, which is immutable.

**Scale out** (condition names in order): `ScaleTServers` raises the tserver PetSet replicas and waits for the new tservers to join, `WaitBalancerIdle` waits until the load balancer has moved tablets to them, and `UpdatePlacement` updates the cluster placement information.

**Scale in**: `DrainTServers` blacklists the tservers that will be removed and waits until their tablets have moved away, `ScaleTServers` lowers the PetSet replicas, `RemoveTServers` removes them from the master registry and clears the blacklist, and `UpdatePlacement` updates the placement. The PVCs of the removed pods are deleted.

## Zones

With a PlacementPolicy whose `zoneSpreadConstraint.whenUnsatisfiable` is `DoNotSchedule`, a role cannot have more pods than zones (see [Clustering](/docs/guides/yugabytedb/clustering/architecture-overview.md)). Use `ScheduleAnyway` on the zone constraint if you scale the tservers beyond the number of zones.

## Next Step

[Scale YugabyteDB horizontally](/docs/guides/yugabytedb/scaling/horizontal-scaling/scale-horizontally/).
