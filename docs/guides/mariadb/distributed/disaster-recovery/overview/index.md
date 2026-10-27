---
title: MariaDB Cross Data Center Disaster Recovery Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-disaster-recovery-overview
    name: Overview
    parent: guides-mariadb-distributed-disaster-recovery
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# MariaDB Cross Data Center Disaster Recovery (DC-DR) Overview

> **New to KubeDB?** Please start [here](/docs/README.md).

## Introduction

The [Distributed MariaDB](/docs/guides/mariadb/distributed/overview/index.md) guide deploys a single Galera
cluster whose pod ordinals are stretched across multiple Kubernetes clusters
over KubeSlice. Every node is a synchronous wsrep writer and peers resolve each
other over `*.slice.local` ServiceExports. That layout maximizes write
availability inside a single failure domain, but a synchronous Galera primary
component that spans data centers (DCs) is fragile: inter-DC network latency
slows down every commit, and an inter-DC partition can stall the cluster or
split the primary component.

**Cross Data Center Disaster Recovery (DC-DR)** changes the shape of the
deployment to survive a full data center loss. Instead of one Galera cluster
stretched across DCs, each Member DC runs its own self contained Galera cluster,
and the DCs are linked by asynchronous, leader to leader replication. A single
cross-DC failover authority decides which DC is writable at any instant, so
there is exactly one active (writable) DC and one or more read-only standby DCs.

This guide explains the DC-DR architecture and concepts. For a hands on setup,
see [Setup DC-DR](/docs/guides/mariadb/distributed/disaster-recovery/setup/index.md).

## Why not a stretched Galera cluster

Galera certification is synchronous: a commit is acknowledged only after the
write set has been ordered across the whole primary component. When that primary
component spans DCs:

- Every write pays the inter-DC round trip latency.
- A network partition between DCs can drop the cluster below quorum, stalling
  writes in both DCs.
- A flapping inter-DC link repeatedly evicts and rejoins nodes, triggering
  expensive state transfers (SST).

DC-DR removes the cross-DC link from the synchronous write path. Galera quorum
becomes strictly intra-DC, and the only thing that crosses the DC boundary is an
asynchronous replication stream plus a single failover decision.

## Core architecture rule: Galera quorum is strictly intra-DC

- **Each Member DC runs its own self contained Galera cluster.** It has its own
  wsrep primary component, its own SST and IST, and (only when its local node
  count is even) its own local **garbd** (Galera Arbitrator) for intra-DC
  even-node quorum. The wsrep certification quorum never crosses the DC boundary,
  so inter-DC latency cannot stall commits and an inter-DC partition cannot split
  the primary component.
- **One MariaDB CR, expanded by the operator.** You still manage a single
  distributed `MariaDB`. The operator partitions `spec.replicas` by the
  PlacementPolicy `distributionRules` and materializes one Galera cluster per
  Member DC, each with its own governing ServiceExport and its own gcomm peer set
  scoped to that DC, plus the cross-DC asynchronous link. The single CR's
  `status.disasterRecovery` carries the per-DC view.

## The cross-DC failover authority

The cross-DC decision is made by the `dr-controlplane`, a three site etcd quorum
running behind an OCM control plane. It publishes one
`coordination.k8s.io` Lease named `primary-dc` for the global failover scope. The
Lease holder is the active DC. This is the single cross-DC failover authority,
and exactly one DC is writable at a time.

Everything keys off one string, the **OCM spoke cluster name**, which is the DC
name. It is the same value used as the Lease `holderIdentity`, the marker
`activeDC`, the pod label `open-cluster-management.io/cluster-name`, and the
PlacementPolicy `distributionRule.clusterName`. Keep them identical.

The per-DC `dr-controlplane` agent projects the Lease holder onto its spoke as a
marker ConfigMap so the data plane never has to reach across DCs to decide
writability:

```
ConfigMap  primary-dc   (namespace dc-failover, on each spoke)
  data.activeDC   = the DC the quorum currently trusts
  data.renewTime  = RFC3339, the observed primary-dc Lease renewTime
  TTL 30s, fail closed: absent, stale, unparseable, or another DC => not active
```

## Cross-DC replication: leader to leader, asynchronous

MariaDB has no asynchronous replication between two Galera clusters out of the
box, so the cross-DC link is net new. Under DC-DR the standby DC's node 0 is a
GTID asynchronous replica of the active DC's writer endpoint:

```sql
CHANGE MASTER TO
  MASTER_HOST = '<db>.<ns>.svc.slice.local',   -- the active DC primary ServiceExport
  MASTER_USE_GTID = slave_pos;
```

The replica uses GTID auto-positioning against the active DC's primary Service
(the load balanced active endpoint), not a fixed node, so an intra-active-DC
writer change is transparent and any active-DC node can serve the stream. Inside
the standby DC, Galera then certifies those applied writes synchronously to the
rest of that DC's nodes. So a DC is internally a normal KubeDB Galera cluster,
and externally either the writable head or a single asynchronous follower,
decided by the Lease. With more than two data DCs, each standby DC runs its own
asynchronous link from the active DC, and an unplanned failover promotes one
survivor while every other standby re-points its `CHANGE MASTER` at the new
active.

### Galera to Galera GTID needs explicit configuration

Linking two Galera clusters by GTID requires settings that the stretched layout
never needed:

- `wsrep_gtid_mode = ON`.
- A `gtid_domain_id` for the asynchronous stream that is distinct from the wsrep
  domain, so the two GTID sources do not collide.
- `log_slave_updates = ON` on every active-DC node, so any of them can serve the
  binlog from a given GTID.
- A dedicated replication user for the cross-DC stream.
- Bounded binlog retention on the source (`expire_logs_days` /
  `binlog_expire_logs_seconds`), so a slow or dead DR DC cannot make the active DC
  retain binlog until its disk fills.

## Fail-closed fence and split-brain safety

Writability is gated by the Lease and fenced locally, and the fence fails closed.

- A non-active DC's Galera cluster is held `read_only = ON` and
  `super_read_only = ON`. These block client writes but not the replication SQL
  thread or the wsrep applier, so the standby DC's asynchronous replica still
  applies the incoming stream and Galera still certifies it to the rest of that
  DC's nodes. Only client connections are refused.
- The in-DC fence reads the projected `primary-dc` marker ConfigMap. If the DC
  cannot confirm it holds the Lease (marker absent, stale past the 30s TTL,
  unparseable, or naming another DC), it forces `super_read_only`.
- This local fence plus the etcd majority is the split-brain guarantee. A
  partitioned old-active DC that still sees clients cannot accept writes, because
  it can no longer confirm it holds the Lease.

## Role labeling and the primary Service

In plain distributed mode the `md-coordinator` labels every Galera node
`kubedb.com/role: Primary` (multi-writer), so the `<db>` primary Service load
balances writes across all nodes. Under DC-DR:

- Only the active DC's nodes carry `kubedb.com/role: Primary`. Within the active
  DC, Galera remains multi-writer, so all active-DC nodes are `Primary`.
- A standby DC's nodes are labeled `standby`, even though Galera considers them
  part of a (separate) primary component.
- As a result the single primary Service and the AppBinding resolve only to the
  active DC. The fence sets this label from the Lease.

The Galera health check is DC-aware: it requires `Primary` only for active-DC
nodes, expects standby-DC nodes to be `standby` with `super_read_only = ON` and a
healthy asynchronous replica state, and scopes the
`wsrep_cluster_state_uuid` / `wsrep_cluster_conf_id` split-brain comparison per
DC, since the two DCs are now separate Galera clusters with different uuids.

## Arbiter DC and per-DC garbd

- The **Arbiter DC** (`role: Arbiter`, empty `replicaIndices`) holds only the
  `dr-controlplane` etcd member and no MariaDB data. It contributes the third
  vote to the cross-DC etcd quorum so a two data center deployment can still reach
  a majority when one data DC is lost.
- A Member DC whose local node count is even gets its own intra-DC **garbd**
  (Galera Arbitrator) so the local Galera cluster keeps odd quorum. Parity is
  evaluated per DC group, not on the global `spec.replicas`. Prefer odd local
  group sizes to avoid needing a per-DC garbd.

## Failover, switchover, and failback

### Planned switchover (zero RPO)

Quiesce writes on the active DC (set `read_only = ON` on its nodes), wait until
the standby DC's asynchronous replica's GTID reaches the active DC's binlog GTID,
then move the Lease and swap source and follower. Because writes are quiesced and
the standby is fully caught up before the handoff, no rows are lost.

Trigger a planned switchover with the CR annotation:

```
dr.kubedb.com/switchover-to: <dc>
```

This is hub driven. There is no `Switchover` OpsRequest type, because the
engine-aware quiesce and catch-up must run in the hub, not in the
engine-agnostic `dr-controlplane`.

### Unplanned failover (DC loss)

If the active DC is lost, the survivor stops its slave thread and becomes
writable without the catch-up wait. The bounded loss is the GTID tail that the
active DC committed but had not yet shipped to the standby. Every other standby
re-points its `CHANGE MASTER` at the new active DC.

### Failback via SST re-seed

When a failed DC returns, it re-attaches as the asynchronous follower of the new
active DC. Because Galera cannot rewind a multi-writer node, the safe and simple
failback is a full SST re-seed of the returned cluster from the new active DC
(dropping its forked tail), then a GTID asynchronous catch-up. Only when the GTID
histories are provably non-divergent can it skip the re-seed. After catch-up, a
coordinated zero RPO Lease handoff returns the active DC.

## Cross-DC lag guard

Plain MariaDB health is the binary wsrep `Synced` signal. DC-DR adds a cross-DC
lag metric, measured on the standby DC's asynchronous replica as
`Seconds_Behind_Master` and the GTID gap (`@@gtid_slave_pos` versus the source's
`@@gtid_binlog_pos`). The lag budget is checked before a planned switchover so a
switchover never moves the Lease to a lagging standby.

## Status: `status.disasterRecovery`

The single distributed `MariaDB` CR exposes the cross-DC view in
`status.disasterRecovery`:

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC that currently holds the `primary-dc` Lease (the writable DC). |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `dataCenters[]` | Per-DC view: `clusterName`, `role`, `leader`, `writable`, `lagBytes`, `healthy`. |
| `lastTransitionTime` | When the DR phase last changed. |

## Architecture at a glance

The example below uses two Member DCs (`dc-a`, `dc-b`) plus one Arbiter DC
(`dc-c`), with `dc-a` holding the Lease.

```
                         dr-controlplane (3 site etcd quorum)
                         publishes Lease  primary-dc => holder: dc-a
                                  |
           +----------------------+----------------------+
           |                      |                      |
   project marker          project marker          etcd member only
           v                      v                      v
  +------------------+   +------------------+   +------------------+
  |      dc-a        |   |      dc-b        |   |      dc-c        |
  |  (active DC)     |   |  (standby DC)    |   |  (Arbiter DC)    |
  |                  |   |                  |   |                  |
  | Galera cluster   |   | Galera cluster   |   | no MariaDB data  |
  | nodes 0,1,2      |   | nodes 3,4,5      |   | etcd member only |
  | role=Primary     |   | role=standby     |   |                  |
  | read_only=OFF    |   | super_read_only  |   |                  |
  |                  |   |   = ON           |   |                  |
  | serves binlog    |   | node 3 is the    |   |                  |
  | via primary Svc  |   | GTID async       |   |                  |
  |                  |   | replica of dc-a  |   |                  |
  +--------+---------+   +---------+--------+   +------------------+
           |                       ^
           |  async GTID stream    |
           +-----------------------+
   <db>.<ns>.svc.slice.local (active DC primary ServiceExport)
```

- Clients reach the single `<db>` primary Service, which resolves only to the
  active DC's nodes (the `Primary` labeled nodes), so writes always land on the
  Lease holder.
- The standby DC stays read only and catches up asynchronously.
- The Arbiter DC carries no MariaDB data and only contributes its etcd vote.

## Enabling DC-DR

DC-DR is currently enabled with an interim annotation on the MariaDB CR:

```
dr.kubedb.com/enabled: "true"
```

This is transitioning to the PlacementPolicy `clusterSpreadConstraint.failoverPolicy`
as the single source of truth. The PlacementPolicy already carries the
`failoverPolicy` and the per-DC `role` (`Member` / `Arbiter`) on its
`distributionRules`.

## Next Steps

- Follow [Setup DC-DR](/docs/guides/mariadb/distributed/disaster-recovery/setup/index.md)
  to deploy a two Member DC plus Arbiter DC MariaDB and verify exactly one
  writable DC.
- Review the [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview/index.md)
  for the OCM, KubeSlice, and PlacementPolicy substrate that DC-DR builds on.
