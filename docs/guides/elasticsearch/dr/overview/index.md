---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: es-dr-overview-elasticsearch
    name: Overview
    parent: es-dr-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for Elasticsearch

KubeDB can run a single distributed `Elasticsearch` across two data centers (DCs) so an
Elasticsearch workload survives the loss of an entire data center. Exactly one DC is the
active write cluster at any instant. The other DC runs a self-contained standby cluster
whose indices are asynchronous Cross-Cluster Replication (CCR) followers of the active
cluster's leader indices. When the active DC is lost, the follower indices on the standby
are promoted to writable, the single search/index endpoint is flipped to the standby, and
clients continue against identical index names.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/elasticsearch/dr/guide/index.md) for every aspect of
  running in DC-DR mode (components, the naming contract, connecting, monitoring, the
  follower-read-only fence, switchover, failback, day-2 ops).
- [DC-DR Runbook](/docs/guides/elasticsearch/dr/runbook/index.md) for what to do in each
  operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Why Elasticsearch DC-DR is its own camp

Most KubeDB engines have a single writable primary and a leader-to-leader replication
stream. Postgres promotes a survivor, MongoDB elects a new primary, and the endpoint
follows the writable node. **Elasticsearch has none of that.**

Elasticsearch is not a single-writer database. It is a cluster of per-shard primaries with
a master-eligible voting quorum: each index is split into shards, every shard has its own
primary and replicas, and a master-eligible quorum handles cluster state and shard
allocation intra-cluster. There is no cluster-wide write primary, and the master quorum
already handles leadership inside one cluster. So the single-primary DR pattern (one
writable leader, a leader-to-leader stream, promote the survivor) does not map. For
Elasticsearch:

- **Cross-DC replication is asynchronous Cross-Cluster Replication (CCR), not a leader
  stream.** The standby cluster registers the active cluster as a remote cluster and
  creates **follower indices** (via **auto-follow patterns** for new indices) that pull
  operations from the active cluster's **leader indices** asynchronously. CCR is
  configured through the Elasticsearch CCR REST API. There is no new replication engine to
  build.
- **The active DC is a write-endpoint routing decision, not an engine state.** Clients
  index into whichever cluster the endpoint is pointed at. DR is active/passive: one
  cluster takes writes, CCR follows them onto the standby, and on failover the endpoint is
  redirected. The `dr-controlplane` Lease decides which cluster is the write target; the
  follower-read-only fence stops writes to a non-active cluster.
- **There is no rewind and no zero-RPO.** CCR is asynchronous, so an unplanned failover
  loses the un-followed tail (bounded by CCR follow lag), and a returned old-active cluster
  may hold documents that were never followed. Elasticsearch cannot rewind an index.
  Failback re-follows the returned cluster from the new active and reconciles or accepts
  the un-followed tail as bounded loss.

So Elasticsearch DC-DR is two independent Elasticsearch clusters (one per Member DC), each
with its own intra-DC master quorum, joined by asynchronous CCR, with the Lease choosing
the write cluster and the follower-read-only fence preventing split writes.

## How it works

DC-DR for Elasticsearch rests on five rules.

- **The master quorum stays intra-DC.** Each Member DC runs its own self-contained
  Elasticsearch cluster: its own master-eligible voting quorum, its own data nodes, its own
  per-shard primaries and replicas. The master quorum never crosses the DC boundary, so
  inter-DC latency or a partition can never flap master election or stall shard allocation.
  There is no cross-DC Elasticsearch voter.
- **The active cluster is chosen only by the `dr-controlplane` primary-DC Lease.** A small
  control plane, backed by a three-site etcd quorum, publishes one Lease per failover scope.
  The Lease holder is the active write DC. Exactly one cluster is the write target at any
  instant.
- **Cross-DC replication is CCR, active to standby.** The standby cluster registers the
  active cluster as a remote cluster (`cluster.remote.<alias>.seeds` at the active's
  transport endpoint, port 9300) and runs auto-follow patterns that create follower indices
  for the active's leader indices. Each follower index pulls its operations once across the
  WAN from the active; within the standby cluster the follower's own replica shards fan out
  intra-DC (the WAN one-copy rule). CCR is one-directional, active to standby, and the
  operator owns the follow direction so the two directions never overlap. A failover pauses
  and promotes the followers on the new active and starts CCR in the reverse direction, and
  never runs CCR both directions for the same index at once.
- **Writability is gated by the Lease and fenced locally, fail closed.** A follower index
  is inherently read-only, which is the fence. A non-active cluster refuses client writes to
  would-be leader indices unless it holds the Lease; a cluster that cannot confirm it holds
  the Lease stays read-only-follower. This local fence, plus the etcd majority, is the
  split-brain guarantee. Without it a partitioned old-active cluster that still sees clients
  would keep accepting writes that never follow, diverging the two clusters.
- **One search/index endpoint follows the active cluster.** The single user-facing endpoint
  (client port 9200) resolves to the active cluster's nodes (selected by the Lease), so
  clients always reach the write cluster. Because CCR keeps index names identical on both
  clusters, after the endpoint flips clients keep working against the same indices.

> **Why never both directions at once?** CCR is one-directional per index: a leader index
> is followed by a follower index. If both directions were enabled for the same index, the
> two clusters would each try to follow the other and the data would ping-pong. A failover
> therefore pauses and promotes the old followers before (or atomically with) starting CCR
> in the new direction.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds Elasticsearch | Purpose |
| --- | --- | --- |
| **Member** | yes | A self-contained Elasticsearch cluster with its own master quorum. One Member is the active write cluster; the other is the CCR follower while standby. |
| **Arbiter** | no | The arbiter DC. Holds only the `dr-controlplane` etcd member and never Elasticsearch, because Elasticsearch has no cross-DC voter. Supplies the tie-break etcd vote. |

## The single-CR, single-endpoint model

The user creates **one** distributed `Elasticsearch` object (with `spec.distributed` and a
`PlacementPolicy` carrying `distributionRules` and a `failoverPolicy`) and gets **one**
search/index endpoint. The operator expands the CR across the Member DCs:

- one self-contained **Elasticsearch cluster per Member DC**, each with its own intra-DC
  master quorum;
- a **remote-cluster registration plus auto-follow patterns** on the standby cluster, so
  the standby's follower indices track the active's leader indices;
- the **Lease-gated search/index endpoint** that resolves to the active cluster.

The single CR's `status.disasterRecovery` carries the whole cross-DC view: the active DC,
each cluster's node health, the CCR follow lag, and the DR phase.

> **Scope.** This spec targets the even two-data-DC layout (two Member DCs plus an Arbiter
> DC). Active/passive CCR is inherently two-cluster, so spanning three or more data DCs
> (fan-out following and a three-way failover) is a separate, larger design and is out of
> scope here.

## Prerequisites

- A distributed Elasticsearch substrate: an Open Cluster Management (OCM) hub and spoke
  clusters, and **flat cross-DC pod networking (KubeSlice) or external listeners**.
  Elasticsearch nodes need cross-DC reachability for the **transport port (9300)** (for the
  remote-cluster seeds CCR uses) and the **client endpoint (9200)**, so routable
  connectivity between the clusters is part of the DC-DR setup.
- The `dr-controlplane` service and its three-site etcd quorum installed across the data
  centers, with a `dr-controlplane` agent in each spoke (DC). The third etcd member sits in
  the Arbiter DC.
- The KubeDB Elasticsearch operator started with the DC-DR flags (coordination kubeconfig
  and the operator's local DC name).
- A **CCR-capable image**. CCR is an Elastic **Platinum/Enterprise** feature, so confirm the
  licensed Elasticsearch image supports it. For OpenSearch, use OpenSearch's own
  cross-cluster replication plugin (OSS), which provides the equivalent leader/follower model.
- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster name,
  the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the pod label
  `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR Elasticsearch

### 1. PlacementPolicy

Assign global pod ordinals to data centers and tag each DC with its role. Here two Member
DCs (`dc-a`, `dc-b`) each hold a three-node Elasticsearch cluster, and `dc-c` is the
Arbiter DC:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: es-dcdr
spec:
  clusterSpreadConstraint:
    slice:
      projectNamespace: kubeslice-demo
      sliceName: demo-slice
    failoverPolicy:
      mode: TwoDC
      trigger:
        scope: Global
    distributionRules:
    - clusterName: dc-a
      role: Member
      replicaIndices: [0, 1, 2]
    - clusterName: dc-b
      role: Member
      replicaIndices: [3, 4, 5]
    - clusterName: dc-c
      role: Arbiter
      replicaIndices: []
```

- A data-bearing **Member** rule carries `replicaIndices` mapping its ordinals to a
  self-contained Elasticsearch cluster with its own master quorum. The **Arbiter** DC
  carries an empty `replicaIndices` and holds no Elasticsearch, only the `dr-controlplane`
  etcd member.
- `failoverPolicy.mode: TwoDC` expects two Member DCs plus the Arbiter DC.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.

### 2. Elasticsearch

Reference the `PlacementPolicy` and opt the Elasticsearch into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-dcdr
  namespace: demo
spec:
  version: xpack-8.19.9
  distributed: true
  enableSSL: true
  replicas: 6
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: es-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

`spec.replicas: 6` is the total node count across both Member DCs. The `PlacementPolicy`
`replicaIndices` split it into a three-node cluster in `dc-a` (ordinals 0, 1, 2) and a
three-node cluster in `dc-b` (ordinals 3, 4, 5). The operator expands this into one
self-contained Elasticsearch cluster in `dc-a` and one in `dc-b`, registers each as a
remote of the other, and enables auto-follow patterns on the standby DC's cluster so its
follower indices track the active cluster's leader indices.

## Observe the DC-DR state

The single `Elasticsearch` object's `status.disasterRecovery` carries the whole cross-DC
view:

```bash
$ kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-a",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    { "clusterName": "dc-a", "role": "Member",  "writable": true,  "nodesReady": 3, "followLagOps": 0,   "healthy": true },
    { "clusterName": "dc-b", "role": "Member",  "writable": false, "nodesReady": 3, "followLagOps": 128, "healthy": true },
    { "clusterName": "dc-c", "role": "Arbiter", "writable": false, "nodesReady": 0, "followLagOps": 0,   "healthy": true }
  ]
}
```

- `activeDC` is the DC that holds the Lease and takes client writes.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports the DC role, whether it is the writable cluster, how many
  nodes are ready, its CCR follow lag in operations (the standby's replication backlog
  behind the active), and its health.

## Unplanned failover

When the active DC is lost, the standby is already a near-current CCR follower. The
orchestrator observes the Lease move to the standby, pauses and promotes the standby's
follower indices (`pause_follow`, `unfollow`, convert to regular writable indices), flips
the search/index endpoint to the standby, and starts CCR in the reverse direction (auto-follow
on the old active's cluster once it returns). `status.disasterRecovery.phase` moves to
`FailingOver` and back to `Steady`.

The RPO is the un-followed CCR tail: operations the active cluster accepted but had not yet
followed onto the standby when it died are lost. There is no rewind.

## Planned switchover (drained, zero document loss)

To move the active DC on purpose without losing documents, annotate the Elasticsearch with
the target DC:

```bash
$ kubectl annotate elasticsearch -n demo es-dcdr dr.kubedb.com/switchover-to=dc-b
```

The orchestrator quiesces indexing on the active cluster, waits for CCR to drain to zero
follow lag (so the target has every operation), then pauses and promotes the target's
follower indices, flips the search/index endpoint, and starts CCR in the reverse direction.
Because CCR has fully drained before the flip, no acknowledged document is lost. The Lease
then follows to `dc-b`.

## Failback

Failback is not a rewind. A returned old-active cluster becomes the CCR follower of the new
active (auto-follow). Operations it accepted but never followed before the failover are a
forked tail Elasticsearch cannot rewind. For correctness, re-follow the returned cluster
from the new active and reconcile the forked tail out of band, or re-seed the affected
indices, or accept and document the forked tail as bounded loss. Once the returned DC is
caught up, a drained planned switchover returns the active DC.

## Cleanup

```bash
$ kubectl delete elasticsearch -n demo es-dcdr
$ kubectl delete placementpolicy es-dcdr
```

Deleting the `Elasticsearch` removes the per-DC Elasticsearch clusters, the remote-cluster
registrations and auto-follow patterns, and the generated cluster-scoped per-DC
`PlacementPolicies` (which carry no owner reference, so the operator deletes them
explicitly). The user-provided base `PlacementPolicy` is left for you to delete.
