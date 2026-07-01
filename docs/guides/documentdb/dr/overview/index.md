---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: guides-documentdb-dr-overview
    name: Overview
    parent: guides-documentdb-dr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for DocumentDB

KubeDB can run a single distributed `DocumentDB` across multiple data centers so the
database survives the loss of an entire data center (DC). Exactly one DC is writable
at any instant; the others are warm, read-only standbys that stream from it across
the DCs. When the active DC is lost, KubeDB promotes a surviving DC, and the single
connection endpoint follows the new writable DC.

KubeDB `DocumentDB` is Microsoft DocumentDB (the `pg_documentdb` extension) running on
PostgreSQL under the hood, so DC-DR reuses the proven PostgreSQL machinery: WAL
streaming replication between data centers, the per-DC `documentdb-coordinator` raft,
and `pg_rewind` for failback. This guide builds on the same distributed substrate
(one CR, Open Cluster Management, KubeSlice, and a `PlacementPolicy`) and adds the
cross-DC failover machinery on top.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/documentdb/dr/guide/index.md), every
  aspect of running in DC-DR mode (components, monitoring, timing, scaling, day-2 ops).
- [DC-DR Runbook](/docs/guides/documentdb/dr/runbook/index.md), what to
  do in each operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## How it works

DC-DR is built on one rule: **the DocumentDB raft never stretches across data centers.**

- **Each data center is a self-contained DocumentDB group.** The operator expands the
  single `DocumentDB` CR into one group per data-bearing DC, each with its own
  `documentdb-coordinator` raft electing a **local** leader, its own local replicas,
  and (when its local replica count is even) its own local arbiter. The raft quorum
  never crosses the DC boundary, so cross-DC latency or a partition can never flap an
  election.
- **One cross-DC authority decides who is writable.** A small control plane
  (`dr-controlplane`), backed by a three-site etcd quorum, publishes one
  `coordination.k8s.io` **Lease** per failover scope. The DC that holds the Lease is
  the **active** (writable) DC. This is the single cross-DC failover decision.
- **Cross-DC replication is leader-to-leader WAL streaming.** The standby DC's local
  leader runs as an asynchronous streaming standby of the active DC's leader; that
  standby DC's own replicas cascade from its local leader. So a standby DC opens
  exactly one cross-DC replication link. Whether standbys stay Hot or Warm and whether
  streaming is Synchronous or Asynchronous follow the CR's `spec.standbyMode`
  (`Hot`/`Warm`) and `spec.streamingMode` (`Synchronous`/`Asynchronous`); cross-DC
  links are asynchronous by design.
- **Writability is fenced locally and fails closed.** A per-DC `dr-controlplane`
  agent projects the Lease holder onto its own spoke cluster as a small marker
  `ConfigMap`. The `documentdb-coordinator` reads only that local marker: if it cannot
  confirm its DC holds the Lease (the DC lost it, or is partitioned from the
  coordination plane), it demotes its leader to read-only. Because the fence lives in
  the DC and fails closed, a cut-off old-active DC stops accepting writes on its own,
  before the hub even reacts. This local fence plus the etcd majority (only one DC can
  hold the Lease) is the split-brain guarantee.
- **Only the active DC's leader is labeled `primary`.** Each DC's coordinator leads
  its own raft, but a non-active DC's leader is labeled `kubedb.com/role: standby`, so
  the single primary `Service` and the `AppBinding` always resolve to the active DC's
  writable leader.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds DocumentDB data | Primary eligible | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A full DocumentDB group; a candidate for the active DC. |
| **Arbiter** | no | no | Vote only, the `dr-controlplane` etcd tie-breaker; runs no DocumentDB. **This is the role a DocumentDB witness DC uses.** |
| **Witness** | yes | no | Data-bearing but never primary, for engines whose witness must carry data (e.g. MongoDB). **Not used by DocumentDB.** |

> For DocumentDB the third "witness" data center is **vote-only** (it holds only the
> `dr-controlplane` etcd member, no DocumentDB), so it is declared with `role: Arbiter`
> and empty `replicaIndices`. The petset `Witness` role is reserved for engines whose
> witness must carry data; DocumentDB does not use it.

A typical layout is two Member DCs plus one vote-only witness DC (`role: Arbiter`):
the three-site etcd quorum lives across all three, but DocumentDB data lives only in
the two Member DCs.

## Deployment topologies (2 DCs vs 3 DCs)

The DR feature needs two things, in different quantities:

- **DocumentDB data** lives in the **Member** data centers. You need at least **two**
  Member DCs for cross-DC redundancy (one active, one warm standby).
- **The failover decision** is made by the `dr-controlplane` etcd **quorum**. A quorum
  makes progress only while a **majority of its three voting sites** is reachable. For
  single-fault tolerance *and* split-brain safety, those three votes should sit in
  **three independent failure domains**. The third domain can be a tiny vote-only
  **witness** (`role: Arbiter`) that holds no DocumentDB data.

So "how many data centers" has two answers: how many hold **data** (two or three), and
how many hold a **quorum vote** (always three for automatic, split-brain-free
failover). The `failoverPolicy.mode` selects the data layout:

### A. Two Member DCs + a witness, `mode: TwoDC` (recommended)

Three sites; two hold DocumentDB data, the third is a vote-only witness DC
(`role: Arbiter`, no DocumentDB):

```yaml
failoverPolicy:
  mode: TwoDC
distributionRules:
- { clusterName: dc-east, role: Member, replicaIndices: [0, 1, 2] }
- { clusterName: dc-west, role: Member, replicaIndices: [3, 4, 5] }
- { clusterName: dc-witness, role: Arbiter }    # etcd vote only, no DocumentDB
```

Any single site can be lost:

- **Lose a Member DC** → the surviving Member plus the witness form a 2/3 majority, so
  the survivor acquires the Lease and is promoted automatically; the lost DC, if alive
  but partitioned, self-fences read-only.
- **Lose the witness** → the two Members are still a 2/3 majority, so writes continue
  uninterrupted.

Because the witness runs no DocumentDB, it is small and cheap. **Run it in a third
public cloud or region**, this is the lowest-cost way to get correct, automatic
failover, and it is the recommended topology whenever a third location is available.

### B. Three Member DCs, `mode: ThreeDC`

Three sites, all data-bearing and primary-eligible:

```yaml
failoverPolicy:
  mode: ThreeDC
distributionRules:
- { clusterName: dc-east,  role: Member, replicaIndices: [0, 1, 2] }
- { clusterName: dc-west,  role: Member, replicaIndices: [3, 4, 5] }
- { clusterName: dc-south, role: Member, replicaIndices: [6, 7, 8] }
```

More data copies and read capacity, and any DC can become primary. Tolerates the loss
of any single Member DC. The cost is three full DocumentDB groups instead of two, use
it when you want a data copy and primary capability in all three locations.

### C. Two sites only, reduced resiliency

If you genuinely have only two locations, you still need a third quorum vote, so you
**place it inside one of the two DCs** (run the third `dr-controlplane` etcd member
there). There is no separate witness site, so that DC now holds **two of the three
votes**:

- **Lose the other DC** (the one with one vote) → the two-vote DC keeps the majority →
  failover/continuity works automatically.
- **Lose the two-vote DC** → the survivor holds only one of three votes, cannot form a
  quorum, and therefore cannot safely become writable on its own. **Automatic failover
  does not happen**; recovery is a manual, operator-confirmed step, and you must be
  certain the failed DC is truly down to avoid split-brain.

This protects against losing one specific DC, not both symmetrically. Prefer adding a
cheap third witness site (topology A) whenever possible.

### At a glance

| Topology | Sites | Data DCs | Tolerates | Automatic failover |
| --- | --- | --- | --- | --- |
| Two Member + Arbiter witness (`TwoDC`) | 3 | 2 | any 1 site | yes |
| Three Member (`ThreeDC`) | 3 | 3 | any 1 site | yes |
| Two sites, co-located quorum | 2 | 2 | only the one-vote DC | only when the one-vote DC is lost |

## Prerequisites

- A working distributed DocumentDB substrate: Open Cluster Management (OCM) hub and
  spoke clusters, KubeSlice connecting the spokes, and a storage class on each spoke.
  DocumentDB reuses the same substrate as
  [Distributed Postgres](/docs/guides/postgres/distributed/overview/index.md), since it
  runs on PostgreSQL under the hood.
- The `dr-controlplane` service and its three-site etcd quorum installed across the
  data centers, with a `dr-controlplane` agent running in each spoke (DC).
- The KubeDB DocumentDB operator started with the DC-DR flags:

  ```
  --dc-dr-enabled
  --dc-dr-coord-kubeconfig=<path to the coordination control plane kubeconfig>
  --dc-dr-local-dc=<this operator's data center name>
  ```

- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster
  name, the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the
  pod label `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR DocumentDB

A DC-DR DocumentDB is a distributed `DocumentDB` whose `PlacementPolicy` carries a
`failoverPolicy` and per-DC roles. The user creates and edits a **single** `DocumentDB`
object and gets one `AppBinding` and one connection endpoint; the operator expands it
into the per-DC groups.

### 1. PlacementPolicy

Assign the global pod ordinals to data centers and tag each DC with its role. Here two
Member DCs (`dc-east`, `dc-west`) each get three DocumentDB pods, and `dc-arbiter` is
the tie-breaking witness:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: docdb-dcdr
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
    - clusterName: dc-east
      role: Member
      replicaIndices: [0, 1, 2]
    - clusterName: dc-west
      role: Member
      replicaIndices: [3, 4, 5]
    - clusterName: dc-arbiter
      role: Arbiter
```

- A data-bearing **Member** rule carries `replicaIndices`; the **Arbiter** witness DC
  (vote only, no DocumentDB) carries none.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.
  Use `Group` with a group name to put a database in its own scope.

### 2. DocumentDB

Reference the `PlacementPolicy` and opt the DocumentDB into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: docdb-dcdr
  namespace: demo
  annotations:
    # Opt this distributed DocumentDB into per-DC DC-DR expansion.
    dr.kubedb.com/enabled: "true"
spec:
  version: "pg17-0.109.0"
  replicas: 6
  distributed: true
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: docdb-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

The operator then creates, per data-bearing DC:

- a per-DC `PetSet` named `<db>-<dc>` (for example `docdb-dcdr-dc-east`) with its own
  intra-DC raft and DC-local governing `Service`;
- a per-DC arbiter `PetSet` `<db>-<dc>-arbiter` when that DC's local node count is
  even.

The witness DC (`role: Arbiter`) runs no DocumentDB pods.

## Observe the DC-DR state

The single `DocumentDB` object's `status.disasterRecovery` carries the whole cross-DC
view:

```bash
$ kubectl get documentdb -n demo docdb-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-east",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    { "clusterName": "dc-east", "role": "primary", "leader": "docdb-dcdr-dc-east-0", "writable": true,  "healthy": true },
    { "clusterName": "dc-west", "role": "standby", "leader": "docdb-dcdr-dc-west-0", "writable": false, "healthy": true, "lagBytes": 4096 }
  ]
}
```

- `activeDC` is the DC that currently holds the Lease and runs the writable primary.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports that DC's local leader, whether it is the writable
  primary, its health, and its cross-DC replication `lagBytes` (the in-DC coordinator
  computes this and surfaces it; the hub never opens cross-cluster SQL).

## Unplanned failover

When the active DC is lost, its agents stop renewing the primary-DC Lease. After the
Lease duration a surviving Member DC's agent acquires it; that DC becomes `activeDC`.
The hub observes the change and drives a bounded-loss promotion of the survivor
through a `ForceFailOver` `DocumentDBOpsRequest`, while the old DC self-fences
read-only locally (before the hub reacts, even under a partition). The primary
`Service` and `AppBinding` then resolve to the new writable DC.

You do not trigger this; it is automatic. `status.disasterRecovery.phase` moves to
`FailingOver` during the transition and back to `Steady` once the survivor is primary.

## Planned switchover (zero-RPO)

To move the active DC on purpose (maintenance, rebalancing) without losing committed
rows, annotate the DocumentDB with the target DC:

```bash
$ kubectl annotate documentdb -n demo docdb-dcdr dr.kubedb.com/switchover-to=dc-west
```

The switchover is coordinated for zero RPO:

1. The target must be a known, healthy DC within the lag budget.
2. The hub asks the active DC to **quiesce** (hold its primary read-only) via the
   primary-DC Lease, so the active primary's write position freezes.
3. The hub waits until the target has replayed to within one WAL page of the frozen
   position.
4. The Lease hands off to the target; it is promoted and the active DC resumes (now as
   a standby). The annotation is cleared automatically.

Because writes are frozen and the target fully catches up before the handoff, a
planned switchover loses no committed rows.

## Scale a data center

Each DC has its own intra-DC raft, so a single `spec.replicas` cannot describe a
scale. Scale a specific DC with a `DocumentDBOpsRequest` that lists per-DC targets:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: docdb-dcdr-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: docdb-dcdr
  horizontalScaling:
    dataCenters:
    - clusterName: dc-west
      replicas: 5
```

Each entry sets that data center's local node count; DCs not listed are unchanged.
The request resizes only `dc-west`'s raft (adding or removing nodes one at a time over
the DC-local network, managing that DC's arbiter parity), then updates the
`PlacementPolicy` so the declarative topology matches. No other DC's raft and no
cross-DC writability is touched. Scaling a DC to `1` makes it a single-node DC (no
in-DC HA, but still part of cross-DC DR); removing a whole DC is a topology change, not
a scale.

## Day-2 operations

The standard DocumentDB `DocumentDBOpsRequest` operations work on a DC-DR cluster and
act on every per-DC group: vertical scaling, volume expansion (online and offline),
version update, and storage migration each apply to all per-DC `PetSet`s and per-DC
arbiters. You issue them exactly as for a non-distributed DocumentDB.

## Cleanup

```bash
$ kubectl delete documentdb -n demo docdb-dcdr
$ kubectl delete placementpolicy docdb-dcdr
```

Deleting the `DocumentDB` removes the per-DC `PetSet`s, governing `Service`s, and the
cluster-scoped per-DC `PlacementPolicies` the operator generated. The user-provided
base `PlacementPolicy` is left for you to delete.
