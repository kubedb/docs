---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: kf-dr-overview-kafka
    name: Overview
    parent: kf-dr-kafka
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for Kafka

KubeDB can run a single distributed `Kafka` across two data centers (DCs) so a Kafka
workload survives the loss of an entire data center. Exactly one DC is the active
write cluster at any instant. The other DC runs a self-contained standby cluster that
receives an asynchronous mirror of the active cluster's topics. When the active DC is
lost, the single bootstrap endpoint is flipped to the standby, the standby is allowed
to take producer writes, and clients continue against identical topic names.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/kafka/dr/guide/index.md) for every aspect of running
  in DC-DR mode (components, the naming contract, connecting, monitoring, the produce
  fence, switchover, failback, day-2 ops).
- [DC-DR Runbook](/docs/guides/kafka/dr/runbook/index.md) for what to do in each
  operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Why Kafka DC-DR is its own camp

Most KubeDB engines have a single writable primary and a leader-to-leader replication
stream. Postgres promotes a survivor, MongoDB elects a new primary, and the endpoint
follows the writable node. **Kafka has none of that.**

Kafka is not a single-writer database. It is a partitioned log: each topic-partition
has its own leader among that partition's in-sync replicas (ISR), inside one cluster.
There is no cluster-wide primary, and KRaft's own Raft quorum already handles
controller and ISR leadership intra-cluster. So the single-primary DR pattern (one
writable leader, a leader-to-leader stream, promote the survivor) does not map. For
Kafka:

- **Cross-DC replication is asynchronous log mirroring, not a leader stream.** Kafka's
  built-in MirrorMaker 2 (MM2) copies topics (data plus configs), consumer-group
  offsets, and heartbeats from a source cluster to a target cluster. KubeDB expresses
  MM2 as ordinary `ConnectCluster` and `Connector` objects. There is no new
  replication engine to build.
- **The active DC is a write-endpoint routing decision, not an engine state.**
  Producers write to whichever cluster the clients are pointed at. DR is
  active/passive: one cluster takes writes, MM2 mirrors them to the standby, and on
  failover the clients are redirected. The `dr-controlplane` Lease decides which
  cluster is the write target; a local produce fence stops producers writing to a
  non-active cluster.
- **There is no rewind and no zero-RPO.** MM2 is asynchronous, so an unplanned
  failover loses the un-mirrored tail (bounded by MM2 lag), and a returned old-active
  cluster may hold records that were never mirrored. Kafka cannot rewind a log.
  Failback reverses the mirror direction and reconciles or accepts the un-mirrored
  tail as bounded loss.

So Kafka DC-DR is two independent Kafka clusters (one per Member DC), each with its
own intra-DC KRaft quorum, joined by asynchronous MM2 mirroring, with the Lease
choosing the write cluster and a produce fence preventing split writes.

## How it works

DC-DR for Kafka rests on five rules.

- **KRaft stays intra-DC.** Each Member DC runs its own self-contained Kafka cluster:
  its own KRaft controller quorum, its own brokers, its own per-partition ISR. The
  KRaft quorum never crosses the DC boundary, so inter-DC latency or a partition can
  never flap controller election or stall ISR. There is no cross-DC Kafka voter.
- **The active cluster is chosen only by the `dr-controlplane` primary-DC Lease.** A
  small control plane, backed by a three-site etcd quorum, publishes one Lease per
  failover scope. The Lease holder is the active write DC. Exactly one cluster is the
  write target at any instant.
- **Cross-DC replication is MM2, active to standby.** Following the "consume from
  remote, produce to local" best practice, the `ConnectCluster` that does the
  mirroring sits with the target (standby) cluster: its `kafkaRef` points at the
  standby Kafka, which also holds Connect's internal config, offset, and status
  topics. It runs a `MirrorSourceConnector` (topic data and configs), a
  `MirrorCheckpointConnector` (consumer-group offsets, so consumers resume after a
  flip), and a `MirrorHeartbeatConnector` (liveness and lag). MM2 uses
  `IdentityReplicationPolicy` so topic names are identical on both clusters and a
  failover is transparent to clients. Because a `ConnectCluster`'s `kafkaRef` is fixed
  and its internal topics live on its local cluster, a single `ConnectCluster` cannot
  reverse direction: KubeDB pre-provisions one `ConnectCluster` per DC and enables the
  mirror connectors only on the current standby's `ConnectCluster`. A failover swaps
  which DC's `ConnectCluster` has the connectors enabled, and never enables both
  directions at once.
- **Writability is gated by the Lease and fenced locally, fail closed.** A non-active
  cluster must refuse producer writes. The fence, driven by the primary-dc marker,
  denies the produce operation on the non-active cluster and fails closed: a cluster
  that cannot confirm it holds the Lease denies produce. This local fence, plus the
  etcd majority, is the split-brain guarantee. Without it a partitioned old-active
  cluster that still sees producers would keep accepting writes that never mirror,
  diverging the two logs.
- **One bootstrap endpoint follows the active cluster.** The single user-facing
  bootstrap Service resolves to the active cluster's brokers (the per-DC `<db>-pods`
  bootstrap, selected by the Lease), so producers and consumers always reach the write
  cluster. Because MM2 uses `IdentityReplicationPolicy`, the same topic names exist on
  the standby, so after the endpoint flips clients keep working and consumers resume
  from the offsets the `MirrorCheckpointConnector` already translated.

> **Why never both directions at once?** `IdentityReplicationPolicy` keeps topic names
> identical and so loses the topic-rename loop guard the default `{source}.` policy
> relies on. If both mirror directions overlap, the same topic can ping-pong between
> the two clusters. A failover therefore disables the old direction's connectors
> before (or atomically with) enabling the new direction's.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds Kafka | Purpose |
| --- | --- | --- |
| **Member** | yes | A self-contained Kafka cluster with its own KRaft quorum. One Member is the active write cluster; the other is the MM2 mirror target while standby. |
| **Arbiter** | no | The arbiter DC. Holds only the `dr-controlplane` etcd member and never Kafka, because Kafka has no cross-DC voter. Supplies the tie-break etcd vote. |

## The single-CR, single-endpoint model

The user creates **one** distributed `Kafka` object (with `spec.distributed` and a
`PlacementPolicy` carrying `distributionRules` and a `failoverPolicy`) and gets **one**
bootstrap endpoint. The operator expands the CR across three CRD kinds:

- one **`Kafka`** cluster per Member DC, each with its own intra-DC KRaft quorum;
- one **`ConnectCluster`** per DC (each `kafkaRef` pointing at its local Kafka);
- the three **`Connector`** objects (`MirrorSourceConnector`,
  `MirrorCheckpointConnector`, `MirrorHeartbeatConnector`) enabled on the current
  standby's `ConnectCluster` for the active-to-standby mirror.

The single CR's `status.disasterRecovery` carries the whole cross-DC view: the active
DC, each cluster's broker health, the MM2 mirror lag, and the DR phase.

> **Scope.** This spec targets the even two-data-DC layout (two Member DCs plus an
> Arbiter DC). Active/passive MM2 is inherently two-cluster, so spanning three or more
> data DCs (fan-out mirroring and a three-way failover) is a separate, larger design
> and is out of scope here.

## Prerequisites

- A distributed Kafka substrate: an Open Cluster Management (OCM) hub and spoke
  clusters, and **flat cross-DC pod networking (KubeSlice) or external listeners**.
  Kafka brokers advertise in-cluster `.svc` listeners, so MM2's cross-cluster reach
  and the cross-DC bootstrap need routable connectivity between the clusters. Wiring
  the advertised listeners for cross-DC reach is part of the DC-DR setup.
- The `dr-controlplane` service and its three-site etcd quorum installed across the
  data centers, with a `dr-controlplane` agent in each spoke (DC). The third etcd
  member sits in the Arbiter DC.
- The KubeDB Kafka operator started with the DC-DR flags (coordination kubeconfig and
  the operator's local DC name).
- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster
  name, the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the
  pod label `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR Kafka

### 1. PlacementPolicy

Assign global pod ordinals to data centers and tag each DC with its role. Here two
Member DCs (`dc-a`, `dc-b`) each hold a three-node Kafka cluster, and `dc-c` is the
Arbiter DC:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: kf-dcdr
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
  self-contained Kafka cluster. The **Arbiter** DC carries an empty `replicaIndices`
  and holds no Kafka, only the `dr-controlplane` etcd member.
- `failoverPolicy.mode: TwoDC` expects two Member DCs plus the Arbiter DC.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.

### 2. Kafka

Reference the `PlacementPolicy` and opt the Kafka into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kf-dcdr
  namespace: demo
spec:
  version: 4.0.0
  distributed: true
  replicas: 6
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: kf-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

`spec.replicas: 6` is the total broker count across both Member DCs. The
`PlacementPolicy` `replicaIndices` split it into a three-node cluster in `dc-a`
(ordinals 0, 1, 2) and a three-node cluster in `dc-b` (ordinals 3, 4, 5). The operator
expands this into one self-contained Kafka cluster in `dc-a` and one in `dc-b`, a
`ConnectCluster` in each DC (each `kafkaRef` pointing at its local Kafka), and the three
MM2 `Connector` objects on the standby DC's `ConnectCluster` mirroring the active
cluster into the standby.

## Observe the DC-DR state

The single `Kafka` object's `status.disasterRecovery` carries the whole cross-DC view:

```bash
$ kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-a",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    { "clusterName": "dc-a", "role": "Member",  "writable": true,  "brokersReady": 3, "mirrorLagMillis": 0,   "healthy": true },
    { "clusterName": "dc-b", "role": "Member",  "writable": false, "brokersReady": 3, "mirrorLagMillis": 850, "healthy": true },
    { "clusterName": "dc-c", "role": "Arbiter", "writable": false, "brokersReady": 0, "mirrorLagMillis": 0,   "healthy": true }
  ]
}
```

- `activeDC` is the DC that holds the Lease and takes producer writes.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports the DC role, whether it is the writable cluster, how
  many brokers are ready, its MM2 mirror lag in milliseconds (the standby's replication
  latency behind the active), and its health.

## Unplanned failover

When the active DC is lost, the standby is already a near-current MM2 mirror. The
orchestrator observes the Lease move to the standby, flips the bootstrap endpoint to
the standby's brokers, opens the standby's produce fence, and reverses the MM2
direction (disabling the connectors on the old active's `ConnectCluster` if it is
reachable, and enabling them on the survivor's for when the old DC returns).
`status.disasterRecovery.phase` moves to `FailingOver` and back to `Steady`.

The RPO is the un-mirrored MM2 tail: records the active cluster accepted but had not
yet mirrored when it died are lost. There is no rewind.

## Planned switchover (drained, zero record loss)

To move the active DC on purpose without losing records, annotate the Kafka with the
target DC:

```bash
$ kubectl annotate kafka -n demo kf-dcdr dr.kubedb.com/switchover-to=dc-b
```

The orchestrator quiesces producers by closing the active cluster's produce fence,
waits for MM2 to drain to near-zero lag (so the target has every record), then flips
the bootstrap endpoint, opens the target's fence, and reverses the mirror direction.
Because MM2 has fully drained before the flip, no committed record is lost. The Lease
then follows to `dc-b`.

## Failback

Failback is not a rewind. A returned old-active cluster becomes the MM2 target of the
new active. Records it accepted but never mirrored before the failover are a forked
tail Kafka cannot rewind, and because MM2 only adds and never deletes, a naive
re-mirror leaves those orphan records on top of the new active's data. For
correctness, re-seed the affected topics from the new active (wipe and re-mirror) or
accept and document the orphan tail as bounded loss. Once the returned DC is caught up,
a drained planned switchover returns the active DC.

## Cleanup

```bash
$ kubectl delete kafka -n demo kf-dcdr
$ kubectl delete placementpolicy kf-dcdr
```

Deleting the `Kafka` removes the per-DC Kafka clusters, the per-DC `ConnectCluster`
objects, the MM2 `Connector` objects, and the generated cluster-scoped per-DC
`PlacementPolicies` (which carry no owner reference, so the operator deletes them
explicitly). The user-provided base `PlacementPolicy` is left for you to delete.
