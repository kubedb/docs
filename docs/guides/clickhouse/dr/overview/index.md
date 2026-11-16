---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: ch-dr-overview-clickhouse
    name: Overview
    parent: ch-dr-clickhouse
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for ClickHouse

KubeDB can run a single distributed `ClickHouse` across multiple data centers (DCs) so
the database survives the loss of an entire data center. Every replica is writable
(ClickHouse `ReplicatedMergeTree` is multi-master), so DR is not about promoting a new
primary. It is about two things: the ClickHouse Keeper Raft quorum, which decides which
DCs can commit at all, and a single Lease-routed write endpoint, which records and
steers where clients send writes. When a data center is lost, the surviving DCs that
still hold Keeper quorum keep accepting writes, and the write endpoint follows to a DC
that holds quorum.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/clickhouse/dr/guide/index.md) for every aspect of
  running in DC-DR mode (components, status, connecting, monitoring, switchover,
  failback, day-2 ops).
- [DC-DR Runbook](/docs/guides/clickhouse/dr/runbook/index.md) for what to do in each
  operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Why ClickHouse DC-DR is different

Most KubeDB engines (Postgres, MariaDB, MSSQL) keep their consensus quorum **inside**
a single DC, because a raft or cluster manager flaps or stalls when its quorum spans
data centers. Those engines run one independent group per DC and build a separate
cross-DC replication link, and DR means promoting a standby.

**ClickHouse is the exception, the same way MongoDB is.** `ReplicatedMergeTree` is
multi-master and geo-aware by design: replicas of a table in different DCs replicate
asynchronously through a shared ClickHouse Keeper ensemble. So for ClickHouse:

- **One logical cluster spans the DCs.** The same shards, with a `ReplicatedMergeTree`
  replica of each shard in each DC, all share one Keeper ensemble and the same
  `default_replica_path` macros. Replication is the native ClickHouse replication link
  over port 9009. There is **no second replication link to build** and no remote
  replica to manage.
- **Failover is the engine's own quorum, not a promotion.** ClickHouse Keeper is a Raft
  service. The DCs that hold the Keeper majority keep registering parts and serving
  writes. A partitioned minority DC loses Keeper quorum, cannot register parts, and so
  its inserts fail. That quorum, not any promotion step, is the failover and the
  split-brain guarantee. KubeDB never promotes a replica, because every replica is
  already writable.
- **Failback is native and clean.** A returned DC's replicas rejoin the Keeper ensemble
  and catch up through `ReplicatedMergeTree` (they fetch the missing parts). There is
  **no rewind**: a partitioned minority DC that lacked Keeper quorum committed nothing
  to diverge, so there is nothing to roll back. This is cleaner than the Postgres
  `pg_rewind` path and even than MongoDB rollback.

## How it works

DC-DR for ClickHouse rests on five rules.

- **ClickHouse Keeper is spread 3-site and is the failover authority.** With two data
  DCs the layout is a Keeper voter in `dc-a`, a Keeper voter in `dc-b`, and one
  data-less **Keeper voter** in a third arbiter DC. That arbiter-DC voter is co-located
  with the `dr-controlplane` etcd member, so the Keeper Raft quorum and the Lease
  quorum share the same 3-site topology (exactly MongoDB's arbiter trick). A single DC
  loss then leaves a surviving Keeper majority, so the survivors keep committing with no
  manual step.
- **Keeper quorum is the writable contract and the split-brain guarantee.** Because the
  safety comes from Keeper's Raft majority, a partitioned minority DC cannot register
  parts and so cannot commit any insert. A cut-off DC goes non-writable on its own, at
  the engine level, with no operator action. This is the same shape as MongoDB majority
  plus `w:majority`, and it is the hard guarantee. The Lease-driven endpoint fence is
  only a routing layer on top of it.
- **The Lease routes the single write endpoint; it does not promote anything.** A small
  control plane (`dr-controlplane`), backed by a three-site etcd quorum, publishes one
  `coordination.k8s.io` **Lease** per failover scope. The Lease records which DC the
  single write endpoint resolves to and steers clients there, giving a stable
  single-writer posture and one consistent cross-engine status. Because ClickHouse is
  multi-master, this is a write-routing choice, not an engine-enforced primary. On an
  unplanned active-DC loss the orchestrator moves the Lease and the endpoint to a
  surviving DC that still holds Keeper quorum. The Lease is routing, policy, and
  observability, **not** the failover mechanism (Keeper quorum is).
- **Reads can stay local.** Any replica serves consistent-enough reads, so read traffic
  can stay in-DC for low latency while writes route to the active DC through the single
  endpoint.
- **One cross-DC part copy per shard, then fetch intra-DC.** `ReplicatedMergeTree`
  fetches are not DC-aware: a replica pulls a new part from whichever replica advertises
  it in Keeper, which can be the cross-DC one. With a single replica of a shard per DC
  (the minimal DR shape) that is already one part copy per DC. But when a DC runs two or
  more replicas of the same shard for intra-DC HA, each can independently pull the same
  part across the WAN. ClickHouse has no native same-DC fetch affinity, so the operator
  designates one in-DC replica per shard as the cross-DC fetch source and points the
  others at it, so they fetch that part intra-DC. This holds cross-DC part traffic to one
  copy per shard per DC, the `ReplicatedMergeTree` analog of the Postgres standby-DC
  intra-DC cascade.

> **Why not confine Keeper to the active DC?** You can (see the topologies below), and
> it gives the lowest write latency. But then losing the active DC also loses its Keeper
> quorum, and the surviving replicas have no quorum to commit against until you bring up
> a new Keeper and re-point them, an explicit manual recovery with RTO impact. Spreading
> Keeper 3-site removes that manual step at the cost of a cross-DC Keeper round trip on
> every insert.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds ClickHouse data | Holds a Keeper voter | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A data-bearing DC holding a `ReplicatedMergeTree` replica of every shard and a Keeper voter; a candidate for the active (write-routed) DC. |
| **Arbiter** | no | yes (data-less) | The arbiter DC. Holds the `dr-controlplane` etcd vote **and** one data-less ClickHouse Keeper voter. Supplies the tie-break vote for both quorums. No ClickHouse data. |

> Unlike MariaDB and MSSQL, whose arbiter DC holds no engine member, the ClickHouse
> arbiter DC holds **both** the `dr-controlplane` etcd member and a data-less Keeper
> voter, co-located so the coordination quorum and the Keeper quorum agree on which DCs
> are alive. This is the same pattern as MongoDB's voting arbiter.

## The Keeper placement decision (the one real tradeoff)

ClickHouse Keeper Raft is latency-sensitive: every insert registers a part, which is a
Keeper write that needs a quorum round trip. ClickHouse is a write-heavy ingest engine,
so where you place Keeper is the one real tradeoff. There are three placements, and the
right one depends on your write-latency tolerance. **3-site spread Keeper is the
documented automatic-DR path here, but it is not automatically the best choice for every
workload.**

### A. Spread Keeper 3-site, single cluster (automatic DR, higher write latency)

One logical ClickHouse cluster with Keeper voters in `dc-a`, `dc-b`, and a data-less
voter in the arbiter DC (co-located with the `dr-controlplane` etcd member). A DC loss
leaves a surviving Keeper majority, so writes continue in the survivor with **no manual
step**. The cost is a cross-DC Keeper round trip on every insert. Batched inserts
amortize it well; high-frequency small inserts may find it prohibitive. This is the path
documented in detail here, because it is the closest analog to the MongoDB design and
gives hands-off failover.

### B. Two clusters, per-region Keeper, cross-replicated (often the better fit)

Each DC runs its own cluster with its own in-DC Keeper (low local write latency); the
two cross-replicate the same `ReplicatedMergeTree` tables and a `Distributed` table
fronts both. Each region writes locally, and on a region loss the other already holds a
full replica. This matches ClickHouse's write-local nature and avoids the cross-DC
Keeper tax, at the cost of a more complex topology and looser cross-region ordering.
This is Altinity's "multi-region writes" pattern and is **frequently the better
production choice for write-heavy ingest**. Failover is still a routing change and the
Lease still picks the write-routed DC.

### C. Single-DC Keeper (lowest latency, manual failover)

Keeper lives only in the active DC; standby replicas use it cross-DC. Inserts are lowest
latency, but losing the active DC (and its Keeper) leaves the standby with no quorum, so
a new Keeper must be brought up and the replicas re-pointed (an explicit recovery with
RTO impact). This is the simplest Altinity default and is acceptable when low write
latency outweighs automatic DR.

### At a glance

| Topology | Keeper | Write latency | Failover on active-DC loss |
| --- | --- | --- | --- |
| A. 3-site spread (documented here) | voters in dc-a, dc-b, arbiter DC | higher (cross-DC round trip per insert) | automatic, survivor keeps quorum |
| B. Two clusters, per-region Keeper | one Keeper per region | low (local) | routing change, other region already replicated |
| C. Single-DC Keeper | only in the active DC | lowest | manual: rebuild Keeper, re-point replicas |

The rest of these docs describe topology **A**, the 3-site single-cluster spread. Its
safety claim (a minority DC cannot register parts) is specific to it. Pick per workload:
for write-heavy ClickHouse, topology B is frequently better, and topology C is the
low-latency manual-failover floor.

## The single-CR, single-endpoint model

The user creates **one** distributed `ClickHouse` object (with `spec.distributed: true`
and a `PlacementPolicy` carrying `distributionRules` and a `failoverPolicy`) and gets
**one** `AppBinding` and **one** write endpoint. The operator expands the CR into per-DC
`ReplicatedMergeTree` replicas of every shard, a 3-site Keeper ensemble (arbiter-DC
voter included), and the Lease-routed write endpoint.

The single CR's `status.disasterRecovery` carries the whole cross-DC view: the active
(write-routed) DC, each DC's per-shard replica health, whether the DC holds Keeper
quorum, the cross-DC replication delay, and the DR phase.

## Prerequisites

- A distributed ClickHouse substrate: Open Cluster Management (OCM) hub and spoke
  clusters, KubeSlice connecting the spokes so replicas reach each other, and a storage
  class on each data-bearing spoke. The ClickHouse and Keeper ports (native 9000,
  replication 9009, Keeper client 9181, Keeper Raft 9234) must be reachable across the
  DCs.
- The `dr-controlplane` service and its three-site etcd quorum installed across the data
  centers, with a `dr-controlplane` agent running in each spoke (DC). The third etcd
  member sits in the arbiter DC alongside the data-less Keeper voter.
- The KubeDB ClickHouse operator started with the DC-DR flags (coordination kubeconfig
  and the operator's local DC name).
- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster
  name, the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the
  pod label `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR ClickHouse

### 1. PlacementPolicy

Assign global replica indices to data centers and tag each DC with its role. Here two
Member DCs (`dc-a`, `dc-b`) each hold a replica of every shard plus a Keeper voter, and
`dc-c` is the arbiter DC holding only a data-less Keeper voter and the `dr-controlplane`
etcd member:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: ch-dcdr
spec:
  clusterSpreadConstraint:
    slice:
      projectNamespace: kubeslice-demo
      sliceName: demo-slice
    failoverPolicy:
      trigger:
        scope: Global
      mode: TwoDC
    distributionRules:
    - clusterName: dc-a
      role: Member
      replicaIndices: [0, 1]      # dc-a: a replica of each shard + a Keeper voter
    - clusterName: dc-b
      role: Member
      replicaIndices: [2, 3]      # dc-b: a replica of each shard + a Keeper voter
    - clusterName: dc-c
      role: Arbiter
      replicaIndices: []          # arbiter DC: dr-controlplane etcd + a data-less Keeper voter
```

- A data-bearing **Member** rule carries `replicaIndices`; the **Arbiter** DC carries
  an empty list (its single data-less Keeper voter is not ordinal-pinned, it is
  scheduled onto the arbiter spoke by the operator).
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.

### 2. ClickHouse

Reference the `PlacementPolicy` and opt the ClickHouse into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: ch-dcdr
  namespace: demo
spec:
  version: "25.7.1"
  distributed: true
  clusterTopology:
    cluster:
    - name: appscode-cluster
      shards: 2
      replicas: 2
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: ch-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

The operator expands this into per-DC `ReplicatedMergeTree` replicas of every shard,
pinned to `dc-a` and `dc-b`, plus a data-less Keeper voter in `dc-c`, and routes the
single write endpoint to the active DC by following the Lease.

## Observe the DC-DR state

The single `ClickHouse` object's `status.disasterRecovery` carries the whole cross-DC
view:

```bash
$ kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-a",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    {
      "clusterName": "dc-a", "role": "Member", "keeperVoter": true, "keeperQuorum": true,
      "writable": true, "healthy": true, "absoluteDelaySeconds": 0, "queueSize": 0,
      "shards": [
        { "shard": 0, "totalReplicas": 2, "activeReplicas": 2 },
        { "shard": 1, "totalReplicas": 2, "activeReplicas": 2 }
      ]
    },
    {
      "clusterName": "dc-b", "role": "Member", "keeperVoter": true, "keeperQuorum": true,
      "writable": false, "healthy": true, "absoluteDelaySeconds": 2, "queueSize": 5,
      "shards": [
        { "shard": 0, "totalReplicas": 2, "activeReplicas": 2 },
        { "shard": 1, "totalReplicas": 2, "activeReplicas": 2 }
      ]
    },
    {
      "clusterName": "dc-c", "role": "Arbiter", "keeperVoter": true, "keeperQuorum": true,
      "writable": false, "healthy": true
    }
  ]
}
```

- `activeDC` is the DC the write endpoint currently resolves to (a routing choice, not
  a promoted primary).
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports the DC role, whether it holds a Keeper voter and
  whether it currently has Keeper quorum, whether it is the write-routed DC, its
  per-shard replica health, its cross-DC `absoluteDelaySeconds` and `queueSize`, and its
  health.

## Unplanned failover

When the active DC is lost, the surviving DCs that still hold Keeper quorum (a standby
data DC plus the arbiter DC in the even layout) **keep accepting writes on their own**,
because Keeper quorum survives and every replica is already writable. There is no
promotion. The orchestrator observes the Lease move to a surviving DC and points the
single write endpoint there. `status.disasterRecovery.phase` moves to `FailingOver` and
back to `Steady`. Bounded loss is only committed-but-unfetched parts on the lost DC's
disk (a clean partition that put the lost DC in the Keeper minority loses zero committed
data, because it could not commit without quorum).

## Planned switchover (near-zero RPO)

To move the active (write-routed) DC on purpose without losing committed writes,
annotate the ClickHouse with the target DC:

```bash
$ kubectl annotate clickhouse -n demo ch-dcdr dr.kubedb.com/switchover-to=dc-b
```

The orchestrator quiesces writes on the current active DC (routes clients away), waits
until the target DC's replicas show `absoluteDelaySeconds` near zero and an empty
replication queue (`queueSize: 0`), then moves the Lease and the write endpoint to
`dc-b`. Because it waits for the target to catch up before flipping, near-zero committed
writes are lost.

## Cleanup

```bash
$ kubectl delete clickhouse -n demo ch-dcdr
$ kubectl delete placementpolicy ch-dcdr
```

Deleting the `ClickHouse` removes the per-DC replica groups, the arbiter-DC Keeper
voter, and the generated per-DC `PlacementPolicies`. The user-provided base
`PlacementPolicy` is left for you to delete.
