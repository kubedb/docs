---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: rd-dr-overview
    name: Overview
    parent: rd-dr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for Redis

KubeDB can run a single distributed `Redis` (or `Valkey`, a drop-in Redis fork with the
identical replication mechanics) across multiple data centers so the database survives the
loss of an entire data center (DC). Exactly one DC is writable at any instant; the other is a
warm, read-only standby that streams from it across the DCs. When the active DC is lost,
KubeDB promotes the surviving DC, and the single connection endpoint follows the new writable
DC.

DC-DR adapts the Postgres/MariaDB DC-DR design to Redis. It makes the `dr-controlplane` Lease
the authority that decides which DC is writable, adds a fail-closed fence in the
`rd-coordinator`, wires the cross-DC async link with plain `REPLICAOF`, and presents one CR
over the per-DC Redis deployments.

> **Native scope: Sentinel and Standalone.** Only Sentinel and Standalone modes have a native
> cross-cluster replication primitive (plain `REPLICAOF` between two independent
> deployments), so those are the initial DC-DR scope. **Cluster mode** has no native
> cross-cluster replication (a `cluster-enabled` node rejects `REPLICAOF` to a node outside
> its own gossip ring, and stretching one gossip ring across DCs is the forbidden
> anti-pattern), so cross-DC DR for Cluster mode needs an external logical-sync tool
> (RedisShake-style) and is out of the initial native scope.

This page is the conceptual overview and a quick start. See also:

- [DC-DR Runbook](/docs/guides/redis/dr/runbook/index.md) for scenario-by-scenario
  procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## How it works

DC-DR is built on one rule: **the gossip ring or Sentinel quorum never stretches across data
centers; a plain cross-DC `REPLICAOF` link is the only thing that crosses the DC boundary.**

- **Each data center is a self-contained Redis.** The operator expands the single `Redis` CR
  into one Redis per data-bearing DC. In Sentinel mode each DC runs its own master, replicas,
  and Sentinel quorum for intra-DC HA; in Standalone mode each DC runs its own master (with
  optional replicas). The Sentinel quorum (or the intra-DC replication topology) never crosses
  the DC boundary, so cross-DC latency or a partition cannot flap an intra-DC election.
- **One cross-DC authority decides who is writable.** A small control plane
  (`dr-controlplane`), backed by a three-site etcd quorum, publishes one `coordination.k8s.io`
  **Lease** per failover scope. The DC that holds the Lease is the **active** (writable) DC,
  and its master is the one writable master. This is the single cross-DC failover decision.
- **Cross-DC replication is a plain async `REPLICAOF` (net-new).** The standby DC's master is
  `REPLICAOF` the active DC's master, a chained async replica that streams the replication
  backlog cross-DC. With more than two data DCs, the active master feeds one `REPLICAOF`
  replica per standby DC. An intra-active-DC failover (Sentinel electing a new master within
  the active DC) is transparent to the standby, because the standby points at the active DC's
  role-pinned primary `Service`, which follows whichever pod is the active master.
- **Writability is fenced locally and fails closed.** A per-DC `dr-controlplane` agent projects
  the Lease holder onto its own spoke cluster as a small marker `ConfigMap` (`primary-dc` in
  the `dc-failover` namespace, `data.activeDC` + `data.renewTime`, 30s TTL). The
  `rd-coordinator` fence reads only that local marker: if it cannot confirm its DC holds the
  Lease (the DC lost it, is partitioned, or the marker is missing or stale), it keeps the
  local master labeled `standby` and never promotes it. Because the fence lives in the DC and
  fails closed, a cut-off old-active DC stops advertising as primary on its own, before the
  hub even reacts. This local fence plus the etcd majority (only one DC can hold the Lease) is
  the split-brain guarantee.
- **Only the active DC's master is labeled `primary`.** Each DC's Sentinel elects its own
  local master, but a non-active DC's master is held `kubedb.com/role: standby` by the fence,
  so the single primary `Service` and the `AppBinding` always resolve to the active DC's
  writable master. Standby DCs keep `replica-read-only yes`.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds Redis data | Primary eligible | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A self-contained Redis; a candidate for the active (writable) DC. |
| **Arbiter** | no | no | Vote only, the `dr-controlplane` etcd tie-breaker; runs no Redis. |

> Redis needs no cross-DC voter of its own: its quorum (gossip or Sentinel) is intra-DC. The
> Arbiter DC holds only the `dr-controlplane` etcd member and no Redis.

A typical layout is two Member DCs plus one vote-only Arbiter DC: the three-site etcd quorum
lives across all three sites, but Redis data lives only in the two Member DCs.

## Deployment topologies (2 DCs vs 3 DCs)

The DR feature needs two things, in different quantities:

- **Redis data** lives in the **Member** data centers (one active, one warm standby).
- **The failover decision** is made by the `dr-controlplane` etcd **quorum**. A quorum makes
  progress only while a **majority of its three voting sites** is reachable. For single-fault
  tolerance *and* split-brain safety, those three votes should sit in **three independent
  failure domains**. The third domain is a tiny vote-only **Arbiter** (`role: Arbiter`) that
  holds no Redis data.

So "how many data centers" has two answers: how many hold **data** (two) and how many hold a
**quorum vote** (always three for automatic, split-brain-free failover).

### A. Two Member DCs + an Arbiter DC (recommended)

Three sites; two hold Redis data, the third is a vote-only Arbiter DC (`role: Arbiter`, no
Redis):

```yaml
failoverPolicy:
  mode: TwoDC
distributionRules:
- { clusterName: dc-east, role: Member, replicaIndices: [0, 1, 2] }
- { clusterName: dc-west, role: Member, replicaIndices: [3, 4, 5] }
- { clusterName: dc-arbiter, role: Arbiter }    # etcd vote only, no Redis
```

Any single site can be lost:

- **Lose a Member DC** then the surviving Member plus the Arbiter form a 2/3 majority, so the
  survivor acquires the Lease and is promoted automatically; the lost DC, if alive but
  partitioned, self-fences read-only.
- **Lose the Arbiter** then the two Members are still a 2/3 majority, so writes continue
  uninterrupted.

Because the Arbiter runs no Redis, it is small and cheap. **Run it in a third region or
cloud.** This is the lowest-cost way to get correct, automatic failover, and it is the
recommended topology whenever a third location is available.

### B. Two sites only (reduced resiliency)

If you genuinely have only two locations, you still need a third quorum vote, so you **place it
inside one of the two DCs** (run the third `dr-controlplane` etcd member there). There is no
separate Arbiter site, so that DC now holds **two of the three votes**:

- **Lose the other DC** (the one with one vote) then the two-vote DC keeps the majority, so
  failover and continuity work automatically.
- **Lose the two-vote DC** then the survivor holds only one of three votes, cannot form a
  quorum, and therefore cannot safely become writable on its own. **Automatic failover does
  not happen**; recovery is a manual, operator-confirmed step, and you must be certain the
  failed DC is truly down to avoid split-brain.

This protects against losing one specific DC, not both symmetrically. Prefer adding a cheap
third Arbiter site (topology A) whenever possible.

### At a glance

| Topology | Sites | Data DCs | Tolerates | Automatic failover |
| --- | --- | --- | --- | --- |
| Two Member + Arbiter (`TwoDC`) | 3 | 2 | any 1 site | yes |
| Two sites, co-located quorum | 2 | 2 | only the one-vote DC | only when the one-vote DC is lost |

## Prerequisites

- A working **distributed Redis** setup: Open Cluster Management (OCM) hub and spoke clusters,
  KubeSlice connecting the spokes (it exports each DC's Redis endpoint cross-cluster for the
  `REPLICAOF` link), and a storage class on each spoke.
- The `dr-controlplane` service and its three-site etcd quorum installed across the data
  centers, with a `dr-controlplane` agent running in each spoke (DC).
- The KubeDB Redis operator started with the DC-DR flags:

  ```
  --dc-dr-enabled
  --dc-dr-coord-kubeconfig=<path to the coordination control plane kubeconfig>
  --dc-dr-local-dc=<this operator's data center name>
  ```

- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster name, the
  agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the pod label
  `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR Redis

A DC-DR Redis is a distributed `Redis` whose `PlacementPolicy` carries a `failoverPolicy` and
per-DC roles. The user creates and edits a **single** `Redis` object and gets one `AppBinding`
and one connection endpoint; the operator expands it into the per-DC Redis deployments and
wires the `REPLICAOF` link between them.

### 1. PlacementPolicy

Assign the global pod ordinals to data centers and tag each DC with its role. Here two Member
DCs (`dc-east`, `dc-west`) each get three Redis pods, and `dc-arbiter` is the tie-breaking
vote:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: redis-dcdr
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

- A data-bearing **Member** rule carries `replicaIndices`; the **Arbiter** DC (vote only, no
  Redis) carries none.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope. Use
  `Group` with a group name to put a database in its own scope.

### 2. Redis

Reference the `PlacementPolicy` and opt the Redis into DC-DR expansion. This example uses
Sentinel mode:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-dcdr
  namespace: demo
  annotations:
    # Opt this distributed Redis into per-DC DC-DR expansion.
    dr.kubedb.com/enabled: "true"
spec:
  version: "7.4.0"
  replicas: 3
  mode: Sentinel
  sentinelRef:
    name: sen-dcdr
    namespace: demo
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: redis-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

The operator then creates, per data-bearing DC:

- a self-contained Redis (its own master, replicas, and Sentinel quorum) backed by a `PetSet`
  named `<db>-<dc>` (for example `redis-dcdr-dc-east`) with a DC-local governing `Service`
  exported over KubeSlice, so peer discovery stays intra-DC;
- the cross-DC `REPLICAOF` link from each standby DC's master to the active DC's master (its
  role-pinned primary `Service`, exported cross-DC over KubeSlice), with the active DC's master
  `REPLICAOF NO ONE`.

The Arbiter DC (`role: Arbiter`) runs no Redis pods.

## Observe the DC-DR state

The single `Redis` object's `status.disasterRecovery` carries the whole cross-DC view:

```bash
$ kubectl get redis -n demo redis-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-east",
  "phase": "Steady",
  "dataCenters": [
    {
      "clusterName": "dc-east", "role": "Member",
      "master": "redis-dcdr-dc-east-0",
      "writable": true, "healthy": true
    },
    {
      "clusterName": "dc-west", "role": "Member",
      "master": "redis-dcdr-dc-west-0",
      "writable": false, "linkStatus": "up",
      "lagBytes": 4096, "healthy": true
    }
  ]
}
```

- `activeDC` is the DC that currently holds the Lease and runs the one writable master.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports that DC's master pod, whether it is writable, its cross-DC
  `linkStatus` (`master_link_status` from `INFO replication`, empty on the active DC), its
  `lagBytes` (the active master's `master_repl_offset` minus this DC master's replicated
  offset), and whether it is healthy. The in-DC coordinator computes these from
  `INFO replication` and publishes them as pod annotations; the hub never opens a cross-DC
  Redis connection.

## Unplanned failover

When the active DC is lost, its agents stop renewing the primary-DC Lease. After the Lease
duration the surviving Member DC's agent acquires it; that DC becomes `activeDC`. The hub
observes the change and records it in status; the survivor's fence then clears, the operator
runs `REPLICAOF NO ONE` on the survivor's master to promote it, relabels it `primary`, and
re-points any other standby DC at the new active master. The old DC, if partially alive, has
already self-fenced (its master stays `standby`, read-only). The primary `Service` and
`AppBinding` then resolve to the new writable DC.

You do not trigger this; it is automatic. `status.disasterRecovery.phase` moves to
`FailingOver` during the transition and back to `Steady` once the survivor is primary. Because
Redis replication is asynchronous (acked by offset), the RPO is bounded by the survivor's
cross-DC offset lag at the moment the active DC died (the un-shipped replication tail).

## Planned switchover (near-zero-RPO)

To move the active DC on purpose (maintenance, rebalancing) without losing writes, annotate the
Redis with the target DC:

```bash
$ kubectl annotate redis -n demo redis-dcdr dr.kubedb.com/switchover-to=dc-west
```

The switchover is coordinated for near-zero RPO:

1. The target must be a known, healthy DC within the lag budget
   (`dr.kubedb.com/switchover-max-lag-bytes`, default 16Mi).
2. The hub quiesces writes on the active DC's master (it is held read-only through the Lease
   quiesce marker), so its `master_repl_offset` stops advancing.
3. The hub waits until the target master's replicated offset reaches the quiesced active
   master's `master_repl_offset` (within a small near-zero-RPO window).
4. The hub hands off the Lease to the target DC. The target is promoted (`REPLICAOF NO ONE`),
   the old primary becomes a `REPLICAOF` replica of the new active, and the annotation is
   cleared automatically.

Because the active master is quiesced and the target reaches offset equality before the handoff,
a planned switchover loses no acknowledged writes.

## Failback

A returned old-active DC becomes a `REPLICAOF` replica of the new active. Redis reconciles a
diverged replica by partial resync from the replication backlog when possible, otherwise a full
RDB resync that discards the replica's local data and re-seeds it, so there is no rewind to
implement; the diverged tail is simply dropped by the full resync. After catch-up, a
coordinated planned switchover returns the active DC.

## Scale a data center

Each DC has its own intra-DC Redis, so a single `spec.replicas` cannot describe a scale. Scale a
specific DC with a `RedisOpsRequest` that lists per-DC targets (each entry sets that DC's local
node count; DCs not listed are unchanged), then the operator updates the `PlacementPolicy` so
the declarative topology matches. No other DC and no cross-DC writability is touched.

## Cleanup

```bash
$ kubectl delete redis -n demo redis-dcdr
$ kubectl delete placementpolicy redis-dcdr
```

Deleting the `Redis` removes the per-DC `PetSet`s, governing `Service`s, and the cluster-scoped
per-DC `PlacementPolicies` the operator generated. The user-provided base `PlacementPolicy` is
left for you to delete.
