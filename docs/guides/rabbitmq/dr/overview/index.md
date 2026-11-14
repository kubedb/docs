---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: rm-dr-overview-rabbitmq
    name: Overview
    parent: rm-dr-rabbitmq
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for RabbitMQ

KubeDB can run a single distributed `RabbitMQ` across two data centers (DCs) so a
RabbitMQ workload survives the loss of an entire data center. Exactly one DC is the
active publish cluster at any instant. The other DC runs a self-contained standby
cluster that receives an asynchronous replica of the active cluster's messages. When
the active DC is lost, the single AMQP publish endpoint is flipped to the standby, the
standby is allowed to take client publishes, and clients continue against identical
queue and exchange names.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/rabbitmq/dr/guide/index.md) for every aspect of
  running in DC-DR mode (components, the naming contract, connecting, monitoring, the
  publish fence, switchover, failback, day-2 ops).
- [DC-DR Runbook](/docs/guides/rabbitmq/dr/runbook/index.md) for what to do in each
  operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Why RabbitMQ DC-DR is its own camp

Most KubeDB engines have a single writable primary and a leader-to-leader replication
stream. Postgres promotes a survivor, MongoDB elects a new primary, and the endpoint
follows the writable node. **RabbitMQ has none of that.**

RabbitMQ is not a single-writer database. A quorum queue has its own Raft group with a
per-queue leader among that queue's replicas, inside one cluster. There is no
cluster-wide primary, and the quorum queue's own Raft already handles queue leadership
intra-cluster (majority via the pod disruption budget). So the single-primary DR
pattern (one writable leader, a leader-to-leader stream, promote the survivor) does not
map. For RabbitMQ:

- **Cross-DC replication is asynchronous message replication, not a leader stream.**
  RabbitMQ's built-in **Federation** plugin (or the **Shovel** plugin) copies messages
  from exchanges and queues on a source cluster to a target cluster. Both are open
  source (not Enterprise) and are pre-enabled in the KubeDB RabbitMQ image. There is no
  new replication engine to build.
- **The active DC is a publish-endpoint routing decision, not an engine state.**
  Publishers write to whichever cluster the clients are pointed at. DR is
  active/passive: one cluster takes publishes, Federation replicates them to the
  standby, and on failover the clients are redirected. The `dr-controlplane` Lease
  decides which cluster is the publish target; a local publish fence stops clients
  writing to a non-active cluster.
- **There is no rewind and no zero-RPO.** Federation is asynchronous, so an unplanned
  failover loses the un-federated tail (bounded by federation lag), and a returned
  old-active cluster may hold messages that were never federated. RabbitMQ cannot
  rewind a queue. Failback reverses the federation direction and reconciles or accepts
  the un-federated tail as bounded loss.

So RabbitMQ DC-DR is two independent RabbitMQ clusters (one per Member DC), each with
its own intra-DC quorum-queue Raft, joined by asynchronous Federation, with the Lease
choosing the publish cluster and a publish fence preventing split writes.

## How it works

DC-DR for RabbitMQ rests on five rules.

- **Quorum-queue Raft stays intra-DC.** Each Member DC runs its own self-contained
  RabbitMQ cluster: its own nodes, its own quorum queues, and each queue's own Raft
  group. The Raft group never crosses the DC boundary, so inter-DC latency or a
  partition can never flap queue leadership or stall a queue. There is no cross-DC
  RabbitMQ voter. Use quorum queues (not classic queues) so intra-DC HA survives node
  loss; classic queues are non-replicated.
- **The active cluster is chosen only by the `dr-controlplane` primary-DC Lease.** A
  small control plane, backed by a three-site etcd quorum, publishes one Lease per
  failover scope. The Lease holder is the active publish DC. Exactly one cluster is the
  publish target at any instant.
- **Cross-DC replication is Federation, active to standby.** A federation upstream on
  the standby cluster pulls messages from the active cluster's exchanges and queues
  asynchronously. One cross-DC link carries each federated resource, and the intra-DC
  quorum-queue replicas fan out locally (the WAN one-copy rule). The upstream points at
  the active cluster's Lease-routed endpoint, so an intra-active-DC queue-leader change
  is transparent to the standby. Because a federation direction is one way (upstream on
  the standby, pulling from the active), the operator enables only the active-to-standby
  direction and never both directions at once. A failover swaps which DC holds the
  active upstream.
- **Writability is gated by the Lease and fenced locally, fail closed.** A non-active
  cluster must refuse client publishes. The fence, driven by the primary-dc marker,
  denies the publish operation on the non-active cluster (revokes the publish
  permission or gates the AMQP listener) and fails closed: a cluster that cannot confirm
  it holds the Lease denies publishes. This local fence, plus the etcd majority, is the
  split-brain guarantee. Without it a partitioned old-active cluster that still sees
  publishers would keep accepting messages that never federate, diverging the two
  clusters.
- **One AMQP endpoint follows the active cluster.** The single user-facing AMQP Service
  resolves to the active cluster's nodes (selected by the Lease), so publishers and
  consumers always reach the publish cluster. Because Federation preserves queue and
  exchange names on both clusters, after the endpoint flips clients keep using the same
  queues and resume from the federated state.

> **Why never both directions at once?** Federation moves messages between two clusters
> without a rename loop guard. If both federation directions overlap, the same message
> can ping-pong between the two clusters. A failover therefore disables the old
> direction's upstream before (or atomically with) enabling the new direction's.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds RabbitMQ | Purpose |
| --- | --- | --- |
| **Member** | yes | A self-contained RabbitMQ cluster with its own quorum-queue Raft. One Member is the active publish cluster; the other is the Federation target while standby. |
| **Arbiter** | no | The arbiter DC. Holds only the `dr-controlplane` etcd member and never RabbitMQ, because RabbitMQ has no cross-DC voter. Supplies the tie-break etcd vote. |

## The single-CR, single-endpoint model

The user creates **one** distributed `RabbitMQ` object (with `spec.distributed` and a
`PlacementPolicy` carrying `distributionRules` and a `failoverPolicy`) and gets **one**
AMQP endpoint. The operator expands the CR across the data centers:

- one self-contained **`RabbitMQ`** cluster per Member DC, each with its own intra-DC
  quorum-queue Raft;
- operator-managed **Federation upstreams and policies** wiring the active cluster to
  the standby (active-to-standby direction only);
- the Lease-gated AMQP publish endpoint plus the local publish fence.

The single CR's `status.disasterRecovery` carries the whole cross-DC view: the active
DC, each cluster's node health, the federation lag, and the DR phase.

> **Scope.** This spec targets the even two-data-DC layout (two Member DCs plus an
> Arbiter DC). Active/passive Federation is inherently two-cluster, so spanning three or
> more data DCs (fan-out federation and a three-way failover) is a separate, larger
> design and is out of scope here.

## Prerequisites

- A distributed RabbitMQ substrate: an Open Cluster Management (OCM) hub and spoke
  clusters, and **flat cross-DC pod networking (KubeSlice) or external listeners**.
  RabbitMQ nodes advertise in-cluster `.svc` endpoints, so Federation's cross-cluster
  reach and the cross-DC publish endpoint need routable connectivity between the
  clusters. Wiring the endpoints for cross-DC reach is part of the DC-DR setup.
- The `dr-controlplane` service and its three-site etcd quorum installed across the
  data centers, with a `dr-controlplane` agent in each spoke (DC). The third etcd
  member sits in the Arbiter DC.
- The KubeDB RabbitMQ operator started with the DC-DR flags (coordination kubeconfig and
  the operator's local DC name).
- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster
  name, the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the
  pod label `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR RabbitMQ

### 1. PlacementPolicy

Assign global pod ordinals to data centers and tag each DC with its role. Here two
Member DCs (`dc-a`, `dc-b`) each hold a three-node RabbitMQ cluster, and `dc-c` is the
Arbiter DC:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: rm-dcdr
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
  self-contained RabbitMQ cluster. The **Arbiter** DC carries an empty `replicaIndices`
  and holds no RabbitMQ, only the `dr-controlplane` etcd member.
- `failoverPolicy.mode: TwoDC` expects two Member DCs plus the Arbiter DC.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.

### 2. RabbitMQ

Reference the `PlacementPolicy` and opt the RabbitMQ into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm-dcdr
  namespace: demo
spec:
  version: "3.13.2"
  distributed: true
  replicas: 6
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: rm-dcdr
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
self-contained RabbitMQ cluster in `dc-a` and one in `dc-b`, and wires the Federation
upstreams and policies on the standby cluster to replicate the active cluster into the
standby.

## Observe the DC-DR state

The single `RabbitMQ` object's `status.disasterRecovery` carries the whole cross-DC
view:

```bash
$ kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-a",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    { "clusterName": "dc-a", "role": "Member",  "writable": true,  "nodesReady": 3, "federationLagMessages": 0,    "healthy": true },
    { "clusterName": "dc-b", "role": "Member",  "writable": false, "nodesReady": 3, "federationLagMessages": 1200, "healthy": true },
    { "clusterName": "dc-c", "role": "Arbiter", "writable": false, "nodesReady": 0, "federationLagMessages": 0,    "healthy": true }
  ]
}
```

- `activeDC` is the DC that holds the Lease and takes client publishes.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports the DC role, whether it is the writable cluster, how
  many nodes are ready, its federation lag in messages (the standby's backlog behind the
  active), and its health.

## Unplanned failover

When the active DC is lost, the standby is already a near-current Federation replica.
The orchestrator observes the Lease move to the standby, flips the AMQP endpoint to the
standby's nodes, opens the standby's publish fence, and reverses the federation
direction (tearing down the old active's upstream if it is reachable, and setting up the
survivor's upstream for when the old DC returns).
`status.disasterRecovery.phase` moves to `FailingOver` and back to `Steady`.

The RPO is the un-federated tail: messages the active cluster accepted but had not yet
federated when it died are lost. There is no rewind.

## Planned switchover (drained, zero message loss)

To move the active DC on purpose without losing messages, annotate the RabbitMQ with the
target DC:

```bash
$ kubectl annotate rabbitmq -n demo rm-dcdr dr.kubedb.com/switchover-to=dc-b
```

The orchestrator quiesces publishers by closing the active cluster's publish fence,
waits for Federation to drain to near-zero lag (so the target has every message), then
flips the AMQP endpoint, opens the target's fence, and reverses the federation
direction. Because Federation has fully drained before the flip, no confirmed message is
lost. The Lease then follows to `dc-b`.

## Failback

Failback is not a rewind. A returned old-active cluster becomes the Federation target of
the new active. Messages it accepted but never federated before the failover are a
forked tail RabbitMQ cannot rewind, and because Federation only adds and never deletes,
a naive re-federation leaves those orphan messages on top of the new active's data. For
correctness, re-seed the affected queues from the new active (purge and re-federate) or
accept and document the orphan tail as bounded loss, and make consumers idempotent (or
apply a dedup window) across the flip. Once the returned DC is caught up, a drained
planned switchover returns the active DC.

## Cleanup

```bash
$ kubectl delete rabbitmq -n demo rm-dcdr
$ kubectl delete placementpolicy rm-dcdr
```

Deleting the `RabbitMQ` removes the per-DC RabbitMQ clusters, the operator-managed
Federation upstreams and policies, and the generated cluster-scoped per-DC
`PlacementPolicies` (which carry no owner reference, so the operator deletes them
explicitly). The user-provided base `PlacementPolicy` is left for you to delete.
