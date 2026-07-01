---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: cas-dr-overview-cassandra
    name: Overview
    parent: cas-dr-cassandra
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for Cassandra

KubeDB can run a single distributed `Cassandra` across multiple data centers (DCs) so
the database survives the loss of an entire data center. Cassandra is masterless: one
Dynamo-style ring spans the data centers, gossip carries membership, and every DC is a
full Cassandra datacenter that accepts reads and writes locally. So DR is not about
promoting a new primary and it is not about a fence that decides who may commit. It is
about two things: the per-query **consistency level** (`LOCAL_QUORUM`), which makes each
DC ack writes locally and tolerate the loss of another DC, and a single Lease-routed
write endpoint, which records and steers where a stable single-writer client sends
writes. When a data center is lost, the surviving DCs keep accepting writes at
`LOCAL_QUORUM` on their own, and the write endpoint follows to a surviving DC.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/cassandra/dr/guide/index.md) for every aspect of
  running in DC-DR mode (components, status, connecting, monitoring, consistency,
  switchover, failback, day-2 ops).
- [DC-DR Runbook](/docs/guides/cassandra/dr/runbook/index.md) for what to do in each
  operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Why Cassandra DC-DR is different

Most KubeDB engines (Postgres, MariaDB, MSSQL) keep their consensus quorum **inside** a
single DC, because a raft or cluster manager flaps or stalls when its quorum spans data
centers. Those engines run one independent group per DC and build a separate cross-DC
replication link, and DR means promoting a standby.

**Cassandra is the geo-native exception, further along the same axis as ClickHouse and
MongoDB.** Cassandra is fully masterless: there is no primary, no cross-DC election, and
multi-datacenter replication is first class in the engine. So for Cassandra:

- **One logical ring spans the DCs.** Each Member DC is a Cassandra datacenter, its
  KubeDB racks becoming racks within it. Gossip carries membership across DCs over the
  cross-DC overlay. Unlike Postgres raft or Galera wsrep, spreading membership across
  DCs is exactly what Cassandra is designed for.
- **Replication is native and continuous.** User keyspaces use `NetworkTopologyStrategy`
  with a replication factor per DC (for example `{dc-a: 3, dc-b: 3, dc-c: 3}`). Every
  write is sent once to each DC (the local coordinator forwards it to a remote
  coordinator that fans out to the local replicas), so the WAN one-copy-then-fan-out
  rule is native. There is **no second replication link to build**.
- **There is no failover in the engine, because there is no leader.** Cassandra has no
  primary and no cross-DC election, so nothing gets promoted and nothing gets fenced.
  Availability is preserved by consistency level, not by a quorum that decides who may
  commit. Writing in more than one DC at once (active-active) is legitimate for
  Cassandra.
- **Failback is native and clean.** A returned DC rejoins the ring by gossip, receives
  hinted handoff within the hint window, and a full cross-DC `nodetool repair`
  (anti-entropy) reconciles the rest. There is **no rewind**: Cassandra is AP and
  reconciles by last-write-wins on cell timestamps, so there is nothing to roll back.
  This is different from the Postgres `pg_rewind` path.

## How it works

DC-DR for Cassandra rests on five rules.

- **One ring, each DC a Cassandra datacenter.** The operator expands the distributed
  `Cassandra` CR into one Cassandra datacenter per Member DC, wiring
  `GossipingPropertyFileSnitch` (per-pod `cassandra-rackdc.properties` with `dc=<DC>`
  and `rack=<rack>`) and per-DC seeds so gossip and the storage port reach across DCs.
  The ring is one; the datacenters within it are the DR boundary.
- **Consistency, not a fence, is the correctness knob.** Use `LOCAL_QUORUM` for reads
  and writes so each DC acks locally (low latency, DC-loss tolerant) while data still
  replicates to the other DCs. A partitioned DC keeps serving `LOCAL_QUORUM` (Cassandra
  is AP), so there is **no split-brain fence in the strong sense**. On rejoin, hinted
  handoff plus anti-entropy `nodetool repair` reconcile. Use `EACH_QUORUM` only when a
  write must be durable in every DC before it acks (higher latency, not DC-loss
  tolerant).
- **The Lease routes the single write endpoint; it does not promote or fence anything.**
  A small control plane (`dr-controlplane`), backed by a three-site etcd quorum,
  publishes one `coordination.k8s.io` **Lease** per failover scope. For Cassandra the
  Lease is pure routing and observability: it records which DC the single user-facing
  write endpoint resolves to and steers clients there, giving a stable single-writer
  posture and one consistent cross-engine status. Because Cassandra is masterless, this
  is a write-routing choice, not an engine-enforced primary and not a data-plane fence.
  On an unplanned DC loss the orchestrator moves the Lease and the endpoint to a
  surviving DC. The Lease is routing, policy, and observability, **not** a failover
  mechanism (there is nothing to fail over, the ring keeps running).
- **Reads and writes can stay local.** Any DC serves `LOCAL_QUORUM` reads and writes, so
  traffic can stay in-DC for low latency while the ring replicates asynchronously to the
  other DCs. The single write endpoint is a convenience for a stable single-writer
  posture, not a requirement of the engine.
- **One cross-DC copy per write per DC, then fan out intra-DC.** With
  `NetworkTopologyStrategy`, a write is sent once across the WAN to each remote DC, whose
  local coordinator fans it out to that DC's replicas over the local network. This is
  native, so cross-DC write traffic is one copy per write per DC with no operator-side
  cascade to build (unlike the ClickHouse per-shard fetch source or the Postgres
  standby-DC cascade).

> **Why prefer an odd number of data DCs?** With an odd number of Member DCs (for
> example three), the `dr-controlplane` etcd quorum has its odd site count among the data
> DCs themselves, and Cassandra's own recommended geo shape is three or more full
> datacenters. No separate arbiter is needed. An arbiter DC appears only when the data-DC
> count is even, and it is engine-free (see below).

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds Cassandra data | Holds a coordination vote | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A full Cassandra datacenter with its racks and its own `NetworkTopologyStrategy` replication factor; a candidate for the single write endpoint. |
| **Arbiter** | no | yes (etcd only) | The arbiter DC. Holds only the `dr-controlplane` etcd vote to give the coordination quorum an odd site count. **No Cassandra runs here.** Present only when the data-DC count is even. |

> Unlike ClickHouse and MongoDB, whose arbiter DC holds a data-less engine voter (a
> Keeper voter or a voting mongod), the Cassandra arbiter DC holds **no engine member at
> all**. Cassandra's data plane needs no cross-DC voter because its correctness comes
> from per-DC quorum on each query, not from a cross-DC vote. The arbiter DC exists only
> to give the `dr-controlplane` failover service its odd etcd quorum.

## Consistency is the tradeoff (the one real decision)

Cassandra has no cross-DC quorum tax on membership: gossip is cheap and asynchronous.
The one real decision is the **consistency level** on your reads and writes, because it
sets the availability-versus-durability tradeoff directly.

### A. LOCAL_QUORUM (the documented default, DC-loss tolerant)

Each read and write reaches a quorum of replicas **within the local DC** and acks
there; the write still replicates asynchronously to the other DCs. Latency is local, and
losing an entire other DC does not block writes in the survivors. The bounded loss on a
hard DC loss is only writes that were acked at `LOCAL_QUORUM` locally but not yet
replicated to the survivors (reconciled by repair if the lost DC returns, otherwise
lost). This is the path documented in detail here and is Cassandra's own recommended
multi-DC default.

### B. EACH_QUORUM (durable in every DC before ack, not DC-loss tolerant)

A write must reach a quorum of replicas **in every DC** before it acks. This gives the
strongest cross-DC durability (a write that acked survives the loss of any one DC), but
it is **not** DC-loss tolerant: if a DC is down, writes at `EACH_QUORUM` fail. Use it
per keyspace or per statement only for data that cannot tolerate the `LOCAL_QUORUM` loss
window.

### At a glance

| Consistency | Ack scope | Write latency | Behavior on a DC loss |
| --- | --- | --- | --- |
| A. `LOCAL_QUORUM` (documented here) | quorum in the local DC | low (local) | writes continue in survivors; bounded loss of unreplicated recent writes |
| B. `EACH_QUORUM` | quorum in every DC | higher (cross-DC on every write) | writes fail while any DC is down (zero loss on the writes that did ack) |

The rest of these docs assume `LOCAL_QUORUM` normal operation with `EACH_QUORUM`
reserved for the keyspaces that need it. Choose per keyspace: `LOCAL_QUORUM` for
availability, `EACH_QUORUM` for the strictest cross-DC durability.

## The single-CR, single-endpoint model

The user creates **one** distributed `Cassandra` object (with `spec.distributed: true`
and a `podPlacementPolicy` referencing a `PlacementPolicy` that carries
`distributionRules` and a `failoverPolicy`) and gets **one** `AppBinding` and **one**
write endpoint. The operator expands the CR into one Cassandra datacenter per Member DC
(its racks placed within it), wires `NetworkTopologyStrategy`, cross-DC seeds, and the
snitch, and routes the single write endpoint to the active DC by following the Lease.

The single CR's `status.disasterRecovery` carries the whole cross-DC view: the active
(write-routed) DC, each DC's UN (up/normal) node count from `nodetool status`, the
cross-DC hint and repair backlog as the lag proxy, and the DR phase.

## Prerequisites

- A distributed Cassandra substrate: Open Cluster Management (OCM) hub and spoke
  clusters, KubeSlice connecting the spokes so the ring reaches across DCs, and a storage
  class on each data-bearing spoke. The Cassandra ports (CQL 9042, storage/gossip 7000
  or TLS 7001, and JMX where enabled) must be reachable across the DCs.
- The `dr-controlplane` service and its three-site etcd quorum installed across the data
  centers, with a `dr-controlplane` agent running in each spoke (DC). When the data-DC
  count is even, the third etcd member sits in the arbiter DC (which runs no Cassandra).
- The KubeDB Cassandra operator started with the DC-DR flags (coordination kubeconfig
  and the operator's local DC name).
- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster
  name, the agent `--dc-name`, the Lease `holderIdentity`, the Cassandra `dc=` in
  `cassandra-rackdc.properties`, the pod label
  `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR Cassandra

### 1. PlacementPolicy

Assign global replica indices to data centers and tag each DC with its role. Here three
Member DCs (`dc-a`, `dc-b`, `dc-c`) each become a full Cassandra datacenter with
replication factor 3, an odd layout that needs no arbiter:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: cas-dcdr
spec:
  clusterSpreadConstraint:
    slice:
      projectNamespace: kubeslice-demo
      sliceName: demo-slice
    failoverPolicy:
      trigger:
        scope: Global
      mode: ThreeDC
    distributionRules:
    - clusterName: dc-a
      role: Member
      replicaIndices: [0, 1, 2]     # Cassandra DC "dc-a" (racks within it), NTS RF 3
    - clusterName: dc-b
      role: Member
      replicaIndices: [3, 4, 5]     # Cassandra DC "dc-b", NTS RF 3
    - clusterName: dc-c
      role: Member
      replicaIndices: [6, 7, 8]     # Cassandra DC "dc-c", NTS RF 3 (odd layout, no arbiter)
```

- Each **Member** rule maps to one Cassandra datacenter; its `replicaIndices` become the
  nodes (racks) of that datacenter.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.
- This odd (`ThreeDC`) layout needs **no arbiter DC**. Use an `Arbiter` rule with an
  empty `replicaIndices` only when the data-DC count is even, and that arbiter runs no
  Cassandra.

### 2. Cassandra

Reference the `PlacementPolicy` and opt the Cassandra into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cas-dcdr
  namespace: demo
spec:
  version: "5.0.3"
  distributed: true
  topology:
    rack:
    - name: rack-a
      replicas: 3
      storage:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 1Gi
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: cas-dcdr
  deletionPolicy: WipeOut
```

The operator expands this into one Cassandra datacenter per Member DC (`dc-a`, `dc-b`,
`dc-c`), each a full ring member with `NetworkTopologyStrategy` replication, and routes
the single write endpoint to the active DC by following the Lease.

## Observe the DC-DR state

The single `Cassandra` object's `status.disasterRecovery` carries the whole cross-DC
view:

```bash
$ kubectl get cassandra -n demo cas-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-a",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    {
      "clusterName": "dc-a", "role": "Member", "replicationFactor": 3, "writable": true, "healthy": true,
      "upNodes": 3, "totalNodes": 3, "hintBacklogBytes": 0, "pendingRanges": 0
    },
    {
      "clusterName": "dc-b", "role": "Member", "replicationFactor": 3, "writable": false, "healthy": true,
      "upNodes": 3, "totalNodes": 3, "hintBacklogBytes": 12, "pendingRanges": 0
    },
    {
      "clusterName": "dc-c", "role": "Member", "replicationFactor": 3, "writable": false, "healthy": true,
      "upNodes": 3, "totalNodes": 3, "hintBacklogBytes": 8, "pendingRanges": 0
    }
  ]
}
```

- `activeDC` is the DC the single write endpoint currently resolves to (a routing
  choice, not a promoted primary, and not a claim that only this DC may write).
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports the DC role, whether it is the write-routed DC, its
  UN (up/normal) node count from `nodetool status`, its cross-DC hint backlog and repair
  backlog (the lag proxy), and its health.

## Unplanned failover

When the active DC is lost, the surviving DCs **keep accepting reads and writes at
`LOCAL_QUORUM` on their own**, because each DC acks locally and every DC is already
writable. There is no promotion and no fence. The orchestrator observes the Lease move
to a surviving DC and points the single write endpoint there.
`status.disasterRecovery.phase` moves to `FailingOver` and back to `Steady`. Bounded loss
is only writes that acked at `LOCAL_QUORUM` in the lost DC but had not yet replicated to
the survivors (recoverable by repair if that DC returns, otherwise lost). Use
`EACH_QUORUM` for keyspaces that cannot tolerate that window.

## Planned switchover (routing move)

To move the active (write-routed) DC on purpose, annotate the Cassandra with the target
DC:

```bash
$ kubectl annotate cassandra -n demo cas-dcdr dr.kubedb.com/switchover-to=dc-b
```

The orchestrator checks the target DC is healthy and within the hint/repair backlog
budget, then moves the Lease and the write endpoint to `dc-b`. Because Cassandra is
active-active and every DC already holds the data, there is no promotion and no catch-up
gate of the ClickHouse or Postgres kind; the move is a routing change. For the strictest
zero-loss handoff, drain hints and run a cross-DC `nodetool repair` first so the target
is fully converged.

## Cleanup

```bash
$ kubectl delete cassandra -n demo cas-dcdr
$ kubectl delete placementpolicy cas-dcdr
```

Deleting the `Cassandra` removes the per-DC datacenters and the generated per-DC
`PlacementPolicies`. The user-provided base `PlacementPolicy` is left for you to delete.
