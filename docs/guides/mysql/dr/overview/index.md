---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-dr-overview
    name: Overview
    parent: guides-mysql-dr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for MySQL

KubeDB can run a single distributed `MySQL` across multiple data centers so the
database survives the loss of an entire data center (DC). Exactly one DC is writable
at any instant; the others are warm, read-only standbys that stream from it across the
DCs. When the active DC is lost, KubeDB promotes a surviving DC, and the single
connection endpoint follows the new writable DC.

DC-DR targets `spec.topology.mode: GroupReplication` (and largely `InnoDBCluster`,
which is Group Replication plus a MySQL Router). `SemiSync` mode is a different shape
(a single writable source with semi-sync replicas) and follows the
[Postgres DC-DR](/docs/guides/postgres/dr/overview/index.md) single-primary model
instead.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/mysql/dr/guide/index.md) — every aspect of running
  in DC-DR mode (components, monitoring, timing, scaling, day-2 ops).
- [DC-DR Runbook](/docs/guides/mysql/dr/runbook/index.md) — what to do in each
  operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## How it works

DC-DR is built on one rule: **the MySQL Group Replication (GR) group never stretches
across data centers.**

- **Each data center is a self-contained GR cluster.** The operator expands the single
  `MySQL` CR into one single-primary GR cluster per data-bearing DC, each with its own
  `group_replication_group_name` UUID, its own XCom/Paxos quorum, and its own **local**
  primary election driven by GR. The GR group never crosses the DC boundary, so cross-DC
  latency cannot stall the group and a partition cannot split it.
- **One cross-DC authority decides who is writable.** A small control plane
  (`dr-controlplane`), backed by a three-site etcd quorum, publishes one
  `coordination.k8s.io` **Lease** per failover scope. The DC that holds the Lease is the
  **active** (writable) DC. This is the single cross-DC failover decision.
- **Cross-DC replication is a normal async replication channel.** The standby DC's GR
  **primary** runs an asynchronous replication channel
  (`CHANGE REPLICATION SOURCE TO ... SOURCE_AUTO_POSITION = 1`) from the active DC's
  primary endpoint; GR then distributes those transactions synchronously to the rest of
  that DC's group. So a standby DC opens exactly one cross-DC link. This adapts the
  existing `RemoteReplica` wiring, but keeps GR and the coordinator running rather than
  RemoteReplica's stripped read-only single server. Because GTIDs are UUID-based and each
  DC's group has a distinct UUID, transactions flow across the channel without any domain
  ID collision; `SOURCE_AUTO_POSITION = 1` handles positioning.
- **Writability is fenced locally and fails closed.** A per-DC `dr-controlplane` agent
  projects the Lease holder onto its own spoke cluster as a small marker `ConfigMap`. The
  `mysql-coordinator` reads only that local marker: if it cannot confirm its DC holds the
  Lease (the DC lost it, or is partitioned from the coordination plane), it holds its GR
  primary `super_read_only = ON`. `super_read_only` blocks client writes but **not** the
  channel's applier, so a standby keeps applying the cross-DC stream while refusing direct
  writes. Because the fence lives in the DC and fails closed, a cut-off old-active DC stops
  accepting writes on its own, before the hub even reacts. This local fence plus the etcd
  majority (only one DC can hold the Lease) is the split-brain guarantee.
- **The fence is re-asserted after every GR election.** GR clears `super_read_only` on
  whichever member it elects primary, so the coordinator re-asserts the fence on every
  label loop. An intra standby-DC GR election (which moves the local primary and the
  channel) therefore cannot leave a writable leader behind.
- **Only the active DC's GR primary is labeled `primary`.** Each DC's GR elects its own
  primary, but a non-active DC's primary is labeled `kubedb.com/role: standby`, so the
  single primary `Service` and the `AppBinding` always resolve to the active DC's writable
  primary.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds MySQL data | Primary eligible | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A full GR cluster; a candidate for the active DC. |
| **Arbiter** | no | no | Vote only — the `dr-controlplane` etcd tie-breaker; runs no MySQL. **This is the role a MySQL witness DC uses.** |

> Group Replication has no data-less voter member (its quorum is intra-DC), so a MySQL
> witness DC holds only the `dr-controlplane` etcd member and no MySQL. It is declared
> with `role: Arbiter` and empty `replicaIndices`. The petset `Witness` role (a
> data-bearing witness) is for engines like MongoDB and is not used by MySQL.

A typical layout is two Member DCs plus one vote-only witness DC (`role: Arbiter`): the
three-site etcd quorum lives across all three, but MySQL data lives only in the two
Member DCs.

## Deployment topologies (2 DCs vs 3 DCs)

The DR feature needs two things, in different quantities:

- **MySQL data** lives in the **Member** data centers. You need at least **two** Member
  DCs for cross-DC redundancy (one active, one warm standby).
- **The failover decision** is made by the `dr-controlplane` etcd **quorum**. A quorum
  makes progress only while a **majority of its three voting sites** is reachable. For
  single-fault tolerance *and* split-brain safety, those three votes should sit in **three
  independent failure domains**. The third domain can be a tiny vote-only **witness**
  (`role: Arbiter`) that holds no MySQL data.

So "how many data centers" has two answers: how many hold **data** (two or three), and
how many hold a **quorum vote** (always three for automatic, split-brain-free failover).
The `failoverPolicy.mode` selects the data layout:

### A. Two Member DCs + a witness — `mode: TwoDC` (recommended)

Three sites; two hold MySQL data, the third is a vote-only witness DC (`role: Arbiter`,
no MySQL):

```yaml
failoverPolicy:
  mode: TwoDC
distributionRules:
- { clusterName: dc-east, role: Member, replicaIndices: [0, 1, 2] }
- { clusterName: dc-west, role: Member, replicaIndices: [3, 4, 5] }
- { clusterName: dc-witness, role: Arbiter }    # etcd vote only, no MySQL
```

Any single site can be lost:

- **Lose a Member DC** → the surviving Member plus the witness form a 2/3 majority, so
  the survivor acquires the Lease and is promoted automatically; the lost DC, if alive but
  partitioned, self-fences read-only.
- **Lose the witness** → the two Members are still a 2/3 majority, so writes continue
  uninterrupted.

Because the witness runs no MySQL, it is small and cheap. **Run it in a third public
cloud or region** — this is the lowest-cost way to get correct, automatic failover, and
it is the recommended topology whenever a third location is available.

### B. Three Member DCs — `mode: ThreeDC`

Three sites, all data-bearing and primary-eligible:

```yaml
failoverPolicy:
  mode: ThreeDC
distributionRules:
- { clusterName: dc-east,  role: Member, replicaIndices: [0, 1, 2] }
- { clusterName: dc-west,  role: Member, replicaIndices: [3, 4, 5] }
- { clusterName: dc-south, role: Member, replicaIndices: [6, 7, 8] }
```

More data copies and read capacity, and any DC can become primary. Tolerates the loss of
any single Member DC. The cost is three full GR clusters instead of two — use it when you
want a data copy and primary capability in all three locations. With three or more data
DCs, each standby DC runs its own async channel from the active, and an unplanned failover
promotes one survivor while every other standby re-points its channel at the new active.

### C. Two sites only — reduced resiliency

If you genuinely have only two locations, you still need a third quorum vote, so you
**place it inside one of the two DCs** (run the third `dr-controlplane` etcd member there).
There is no separate witness site, so that DC now holds **two of the three votes**:

- **Lose the other DC** (the one with one vote) → the two-vote DC keeps the majority →
  failover/continuity works automatically.
- **Lose the two-vote DC** → the survivor holds only one of three votes, cannot form a
  quorum, and therefore cannot safely become writable on its own. **Automatic failover does
  not happen**; recovery is a manual, operator-confirmed step, and you must be certain the
  failed DC is truly down to avoid split-brain.

This protects against losing one specific DC, not both symmetrically. Prefer adding a cheap
third witness site (topology A) whenever possible.

### At a glance

| Topology | Sites | Data DCs | Tolerates | Automatic failover |
| --- | --- | --- | --- | --- |
| Two Member + Arbiter witness (`TwoDC`) | 3 | 2 | any 1 site | yes |
| Three Member (`ThreeDC`) | 3 | 3 | any 1 site | yes |
| Two sites, co-located quorum | 2 | 2 | only the one-vote DC | only when the one-vote DC is lost |

## Prerequisites

- A working **distributed MySQL** setup: Open Cluster Management (OCM) hub and spoke
  clusters, KubeSlice connecting the spokes, and a storage class on each spoke.
- The `dr-controlplane` service and its three-site etcd quorum installed across the data
  centers, with a `dr-controlplane` agent running in each spoke (DC).
- The KubeDB MySQL operator started with the DC-DR flags:

  ```
  --dc-dr-enabled
  --dc-dr-coord-kubeconfig=<path to the coordination control plane kubeconfig>
  --dc-dr-local-dc=<this operator's data center name>
  ```

- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster name,
  the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the pod label
  `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR MySQL

A DC-DR MySQL is a distributed `MySQL` whose `PlacementPolicy` carries a `failoverPolicy`
and per-DC roles. The user creates and edits a **single** `MySQL` object and gets one
`AppBinding` and one connection endpoint; the operator expands it into the per-DC GR
clusters.

### 1. PlacementPolicy

Assign the global pod ordinals to data centers and tag each DC with its role. Here two
Member DCs (`dc-east`, `dc-west`) each get three MySQL pods, and `dc-arbiter` is the
tie-breaking witness:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: my-dcdr
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
  (vote only, no MySQL) carries none.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope. Use
  `Group` with a group name to put a database in its own scope.
- Give each Member DC an **odd** local node count so its GR group keeps a clean majority
  for intra-DC failover.

### 2. MySQL

Reference the `PlacementPolicy` and opt the MySQL into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: my-dcdr
  namespace: demo
  annotations:
    # Opt this distributed MySQL into per-DC DC-DR expansion.
    dr.kubedb.com/enabled: "true"
spec:
  version: "8.4.8"
  replicas: 6
  distributed: true
  topology:
    mode: GroupReplication
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: my-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

The operator then creates, per data-bearing DC:

- a per-DC `PetSet` named `<db>-<dc>` (for example `my-dcdr-dc-east`) with its own intra-DC
  GR group and DC-local governing `Service` (exported over KubeSlice);
- the cross-DC async channel on each standby DC's GR primary.

The witness DC (`role: Arbiter`) runs no MySQL pods.

## Observe the DC-DR state

The single `MySQL` object's `status.disasterRecovery` carries the whole cross-DC view:

```bash
$ kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-east",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    { "clusterName": "dc-east", "role": "Member", "primary": "my-dcdr-dc-east-0", "writable": true,  "healthy": true },
    { "clusterName": "dc-west", "role": "Member", "primary": "my-dcdr-dc-west-0", "writable": false, "healthy": true, "lagBytes": 4096, "secondsBehindSource": 1 }
  ]
}
```

- `activeDC` is the DC that currently holds the Lease and runs the writable GR primary.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports that DC's GR primary pod, whether it is the writable
  primary, its health, and its cross-DC lag as both a GTID gap (`lagBytes`) and
  `secondsBehindSource` (from `SHOW REPLICA STATUS`). The in-DC coordinator computes these
  and surfaces them; the hub never opens cross-cluster SQL.

## Unplanned failover

When the active DC is lost, its agents stop renewing the primary-DC Lease. After the Lease
duration a surviving Member DC's agent acquires it; that DC becomes `activeDC`. The hub
observes the change and clears the survivor's fence: it relabels the survivor's GR primary
`primary`, sets `super_read_only = OFF`, stops the survivor's inbound channel, and repoints
every other standby DC's channel at the new active. The old DC self-fences read-only
locally (before the hub reacts, even under a partition). The primary `Service` and
`AppBinding` then resolve to the new writable DC.

You do not trigger this; it is automatic. `status.disasterRecovery.phase` moves to
`FailingOver` during the transition and back to `Steady` once the survivor is primary. The
RPO is bounded by the survivor's cross-DC lag at the moment the active DC died (the
un-shipped GTID tail).

## Planned switchover (near-zero-RPO)

To move the active DC on purpose (maintenance, rebalancing) without losing committed rows,
annotate the MySQL with the target DC:

```bash
$ kubectl annotate my -n demo my-dcdr dr.kubedb.com/switchover-to=dc-west
```

The switchover is coordinated for near-zero RPO:

1. The target must be a known, healthy DC within the lag budget.
2. The hub quiesces writes on the active DC (holds its GR primary `super_read_only = ON`),
   so the active primary's `gtid_executed` freezes.
3. The hub waits until the target's channel has applied up to the active primary's frozen
   `gtid_executed`.
4. The Lease hands off to the target; its GR primary is promoted (relabeled `primary`,
   fence cleared) and the active DC resumes as a standby (its GR primary starts a channel
   from the new active). The annotation is cleared automatically.

Because writes are frozen and the target fully catches up by GTID before the handoff, a
planned switchover loses no committed rows.

## Scale a data center

Each DC has its own intra-DC GR group, so a single `spec.replicas` cannot describe a
scale. Scale a specific DC with a `MySQLOpsRequest` that lists per-DC targets:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-dcdr-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: my-dcdr
  horizontalScaling:
    dataCenters:
    - clusterName: dc-west
      replicas: 5
```

Each entry sets that data center's local GR node count; DCs not listed are unchanged. The
request resizes only `dc-west`'s GR group, then updates the `PlacementPolicy` so the
declarative topology matches. No other DC's group and no cross-DC writability is touched.

## Day-2 operations

The standard MySQL `MySQLOpsRequest` operations work on a DC-DR cluster and act on every
per-DC GR cluster: vertical scaling, volume expansion (online and offline), version update,
and storage migration each apply to all per-DC `PetSet`s. You issue them exactly as for a
non-distributed MySQL.

## Cleanup

```bash
$ kubectl delete my -n demo my-dcdr
$ kubectl delete placementpolicy my-dcdr
```

Deleting the `MySQL` removes the per-DC `PetSet`s, governing `Service`s, and the
cluster-scoped per-DC `PlacementPolicies` the operator generated. The user-provided base
`PlacementPolicy` is left for you to delete.
