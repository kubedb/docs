---
title: YugabyteDB Cluster Architecture
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-clustering-architecture
    name: Architecture Overview
    parent: yugabytedb-clustering
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Cluster Architecture

## Roles

A YugabyteDB universe made by KubeDB has two roles. Each role runs in its own PetSet (`<name>-master` and `<name>-tserver`) with the `OnDelete` update strategy, so only KubeDB decides when a pod is replaced.

| Role | Pods | What it does | Ports |
|------|------|--------------|-------|
| `yb-master` | exactly the replication factor (3 for RF3) | system catalog (one Raft group), load balancer, cluster configuration (placement, blacklists) | RPC 7100, UI 7000 |
| `yb-tserver` | N, at least the replication factor | stores the tablet replicas, runs the YSQL (PostgreSQL) and YCQL query layers | RPC 9100, UI/metrics 9000, YSQL 5433, YCQL 9042 |

Restarting a tserver also restarts the PostgreSQL process that runs inside it. YSQL client connections to that pod are dropped, which is why the guides define "zero downtime" as "no failed calls except connections dropped to the restarted pod".

## Services

| Service | Selects | Purpose |
|---------|---------|---------|
| `<name>` | tserver pods with the label `kubedb.com/auth-bootstrapped=true` | client access on 5433 (and 9042 with YCQL) |
| `<name>-pods` | all pods, `publishNotReadyAddresses: true` | headless governing Service, gives every pod the FQDN `<pod>.<name>-pods.<ns>.svc` |
| `<name>-stats` | all pods | metrics, only when monitoring is enabled |

The client Service only routes to a tserver after the first-reconcile auth bootstrap changed the default `yugabyte` password.

## Replication factor

`spec.replicationFactor` (1, 3, 5 or 7) is both the number of masters and the number of replicas of each tablet. It is immutable. With RF 3 the universe tolerates the loss of one master and one tserver (one zone). With RF 1 there is nothing to fail over to, so no operation can be zero downtime.

## Placement across zones

KubeDB reads the zone of each pod from its node label `topology.kubernetes.io/zone` and writes it into a pod annotation (`yugabytedb.kubedb.com/zone`). After all tservers are alive it sets the cluster placement from the observed zones: one placement block per (cloud, region, zone), with `min_num_replicas: 1` per block when the number of zones equals the replication factor. `spec.placement.preferredZones` orders the zones that should hold the leaders.

To place pods on different zones, create a `PlacementPolicy` and reference it from each role: `spec.topology.<role>.podTemplate.spec.podPlacementPolicy.name`. See the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md) for an example.

### Constraint: DoNotSchedule limits a role to one pod per zone

When the policy's `zoneSpreadConstraint.whenUnsatisfiable` is `DoNotSchedule`, the PetSet controller also adds a required one-pod-per-zone anti-affinity. A role then cannot have more pods than zones. In the test, scaling tservers from 3 to 6 over three zones left the new pods `Pending`.

If the number of tservers may exceed the number of zones, set `whenUnsatisfiable: ScheduleAnyway` on the zone constraint. Pods that already exist keep the old affinity until they are recreated, so run a [Restart](/docs/guides/yugabytedb/restart/restart.md) after changing the policy.

## Rolling operations

Every operation that restarts pods (Restart, VerticalScaling, Reconfigure when a flag needs a restart, ReconfigureTLS, UpdateVersion) uses one rolling engine:

1. Pre-flight: all masters alive with one leader, all tservers alive, no under-replicated tablets.
2. Masters first, highest ordinal first, the current leader last (after a leader step-down). After each pod, wait for it to be back as a follower and caught up, then wait `rolling.masterWait`.
3. Then tservers one at a time. For each: check that it is safe to take down, move its tablet leaders away (blacklist the node, wait up to `rolling.leaderDrainTimeout`), delete the pod, wait until it is ready and caught up, remove the blacklist entry, wait `rolling.tserverWait`.

If the leaders do not drain in time, the operation does not fail. It records a `LeaderDrainTimedOut` condition and the remaining leaders fail over by Raft election.

## Disk space

YugabyteDB rejects writes when the data disk has less than 5% free space. Plan the volume size with that margin, and use [Volume Expansion](/docs/guides/yugabytedb/volume-expansion/volume-expansion.md) (not verified end to end) to grow it.

## Next Steps

- [Quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md)
- [Horizontal Scaling](/docs/guides/yugabytedb/scaling/horizontal-scaling/scale-horizontally/)
