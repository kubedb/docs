---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: mg-dr-overview-mongodb
    name: Overview
    parent: mg-dr-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for MongoDB

KubeDB can run a single distributed `MongoDB` across multiple data centers (DCs) so
the database survives the loss of an entire data center. Exactly one DC runs the
writable primary at any instant; the others are warm secondaries that serve cross-DC
reads. When the active DC is lost, MongoDB's own majority election promotes a new
primary in a surviving DC automatically, and the single connection endpoint follows
the new writable DC.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/mongodb/dr/guide/index.md) for every aspect of
  running in DC-DR mode (components, status, connecting, monitoring, switchover,
  failback, day-2 ops).
- [DC-DR Runbook](/docs/guides/mongodb/dr/runbook/index.md) for what to do in each
  operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

> **Availability:** the distributed MongoDB substrate (`spec.distributed`, the
> `PlacementPolicy`, cross-cluster networking) and the DC-DR layer are net-new for
> MongoDB. Treat the field names and flows here as the intended user experience and
> confirm availability in your release before relying on them in production.

## Why MongoDB DC-DR is different

Most KubeDB engines (Postgres, MariaDB, MSSQL) keep their consensus quorum **inside**
a single DC, because a raft or cluster manager flaps or stalls when its quorum spans
data centers. Those engines therefore run one independent group per DC and build a
separate cross-DC replication link.

**MongoDB is the exception. A replica set is geo-aware by design.** Spreading voting
members across data centers with a tie-break voter is the documented, supported
MongoDB geo deployment. So for MongoDB:

- **One replica set spans the DCs.** There is one logical replica set whose members
  are pinned to DCs. The oplog is already the cross-DC link, replicated
  asynchronously to members in the other DCs. There is **no second replication link
  to build** and no remote-replica.
- **Failover is MongoDB's own election.** When the active DC is lost, the surviving
  voting members form a majority and elect a new primary automatically. KubeDB does
  **not** force or drive promotion.
- **Failback is native.** A returned old primary rolls back its un-replicated tail
  automatically when it rejoins as a secondary, or does a full initial resync if it
  fell outside the rollback/oplog window. There is **no `pg_rewind` equivalent**.

## How it works

DC-DR for MongoDB rests on four rules.

- **Votes are spread 3-site so no single data DC holds a majority.** With two data
  DCs the layout is `dc-a` data members `votes:1`, `dc-b` data members `votes:1`, and
  one data-less **MongoDB voting arbiter** in a third arbiter DC. With totals like
  2 + 2 + 1 the majority is 3, so `dc-a` plus the arbiter DC, or `dc-b` plus the
  arbiter DC, can elect, but neither data DC alone can. This removes split brain at
  its root: a partitioned data DC can never gather a majority by itself.
- **`w:majority` is the writable contract and the split-brain guarantee.** Because
  the safety comes from MongoDB's majority, the writable path defaults to
  `w:majority`. A partitioned minority DC then cannot commit, and a primary that
  loses its majority auto-steps-down to a secondary, so a cut-off DC goes read-only
  on its own. With `w:1` the bounded loss is the un-replicated oplog tail, which is
  rolled back natively on rejoin.
- **The Lease steers priority and follows the primary; it is not the failover
  mechanism.** A small control plane (`dr-controlplane`), backed by a three-site etcd
  quorum, publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease
  holder's data members get a higher MongoDB `priority`, so MongoDB keeps the primary
  in the DC the operator intends. Priority is a preference, not a pin: during a member
  bounce a standby member can briefly become primary and priority takeover returns
  it, so the observed primary can transiently differ from the Lease-intended DC. That
  is expected, not a bug; there is still exactly one primary at all times. On an
  unplanned active-DC loss MongoDB elects in a standby DC first and the Lease then
  **follows** the new primary (inverted from Postgres, where the Lease leads).
- **Only the active DC's primary carries `kubedb.com/role: primary`.** The existing
  `replication-mode-detector` sidecar labels the elected primary `primary` and the
  secondaries `secondary`. Because priority keeps the primary in the active DC, the
  single primary `Service` and the `AppBinding` resolve there. Standby DC members are
  non-hidden, electable `secondary` members that appear on a separate
  `<db>-secondary` read Service.

> **Why not confine the votes to one DC?** A tempting design is to give all votes to
> the active DC and force-reconfig them away on Lease loss. That is unsafe: in a
> partition both sides would issue `replSetReconfig {force:true}` at once (the config
> diverges), and because the confined DC holds a majority, a partitioned old primary
> could still commit `w:majority` writes, a true split brain. Spreading votes 3-site
> with `w:majority` removes both problems with no force reconfig.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds MongoDB data | Primary eligible | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A data-bearing replica-set member group; a candidate for the active DC. |
| **Arbiter** | no | no | The arbiter DC. Holds the `dr-controlplane` etcd vote **and** one MongoDB voting arbiter (data-less). Supplies the tie-break vote. |

> Unlike MariaDB and MSSQL, whose arbiter DC holds no engine member, the MongoDB
> arbiter DC holds **both** the `dr-controlplane` etcd member and a MongoDB voting
> arbiter, co-located so the coordination quorum and the replica-set quorum agree.
> The petset `Witness` role is **removed** for MongoDB; only `Member` and `Arbiter`
> are used.

## Deployment topologies

MongoDB DC-DR supports two shapes. The difference is the vote math after a full DC
loss.

### A. Two Member DCs plus an arbiter DC (the 2 + 2 + 1 even layout)

Three sites; two hold MongoDB data, the third holds the etcd vote plus one MongoDB
voting arbiter (no data):

```yaml
failoverPolicy:
  mode: TwoDC
distributionRules:
- { clusterName: dc-a, role: Member, replicaIndices: [0, 1] }   # votes:1 each
- { clusterName: dc-b, role: Member, replicaIndices: [2, 3] }   # votes:1 each
- { clusterName: dc-c, role: Arbiter }   # dr-controlplane etcd + 1 MongoDB arbiter
```

Five voting members, majority 3. Either data DC plus the arbiter DC can elect; no
single data DC can.

- **Lose a data DC** the survivor plus the arbiter DC still form a majority, so
  MongoDB elects a primary in the survivor automatically. But `w:majority` writes
  **stall**, because only two of the five data-bearing members are reachable and a
  majority of data acks is no longer possible (MongoDB's documented two-data-center
  limitation). The operator then issues a normal, majority-committed `replSetReconfig`
  that drops the lost members, so the majority recomputes to the survivors and
  `w:majority` writes resume.
- **Lose the arbiter DC alone** the two data DCs together hold 4 of 5 votes, still a
  majority, so a primary holds and writes continue.

### B. Odd number of Member DCs, no arbiter DC (recommended)

Three (or any odd number of) data-bearing `Member` DCs, every DC carrying data and
electable, no separate arbiter:

```yaml
failoverPolicy:
  mode: ThreeDC
distributionRules:
- { clusterName: dc-a, role: Member, replicaIndices: [0, 1] }
- { clusterName: dc-b, role: Member, replicaIndices: [2, 3] }
- { clusterName: dc-c, role: Member, replicaIndices: [4, 5] }
```

Cap the voting members to an odd total (typically one voting member per DC, extra
replicas at `votes:0`). A single DC loss then keeps a **data** majority, so MongoDB
elects in a surviving DC **and** `w:majority` never stalls, no reconfig-out step
needed. This is MongoDB's recommended geo shape; prefer it when a third data location
is available.

### At a glance

| Topology | Sites | Data DCs | Tolerates | `w:majority` after a data-DC loss |
| --- | --- | --- | --- | --- |
| 2 Member + Arbiter DC (`TwoDC`, 2 + 2 + 1) | 3 | 2 | any 1 site | stalls until the operator reconfigs the lost members out |
| Odd Member DCs (`ThreeDC`) | 3+ | 3+ | any 1 site | never stalls |

## The single-CR, single-endpoint model

The user creates **one** distributed `MongoDB` object (with `spec.distributed` and a
`PlacementPolicy` carrying `distributionRules` and a `failoverPolicy`) and gets
**one** `AppBinding` and **one** endpoint. The operator expands the CR into per-DC
member groups, all in one replica set, plus the MongoDB arbiter in the even layout,
with priority steered by the Lease.

The single CR's `status.disasterRecovery` carries the whole cross-DC view: the active
DC, each DC's members and primary, the cross-DC oplog lag in seconds, and the DR
phase.

## Prerequisites

- A distributed MongoDB substrate: Open Cluster Management (OCM) hub and spoke
  clusters, KubeSlice connecting the spokes (members reach each other over
  `*.slice.local` with split-horizon **Horizons** DNS), and a storage class on each
  data-bearing spoke.
- The `dr-controlplane` service and its three-site etcd quorum installed across the
  data centers, with a `dr-controlplane` agent running in each spoke (DC). In the even
  layout, the third etcd member sits in the arbiter DC alongside the MongoDB arbiter.
- The KubeDB MongoDB operator started with the DC-DR flags (coordination kubeconfig
  and the operator's local DC name).
- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster
  name, the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the
  pod label `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR MongoDB

### 1. PlacementPolicy

Assign global pod ordinals to data centers and tag each DC with its role. Here two
Member DCs (`dc-a`, `dc-b`) each get two MongoDB members, and `dc-c` is the arbiter
DC:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: mg-dcdr
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
      replicaIndices: [0, 1]
    - clusterName: dc-b
      role: Member
      replicaIndices: [2, 3]
    - clusterName: dc-c
      role: Arbiter
```

- A data-bearing **Member** rule carries `replicaIndices`; the **Arbiter** DC carries
  none (its single MongoDB arbiter is not ordinal-pinned, it is scheduled onto the
  arbiter spoke by the operator).
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.

### 2. MongoDB

Reference the `PlacementPolicy` and opt the MongoDB into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-dcdr
  namespace: demo
spec:
  version: "8.0.5"
  distributed: true
  replicaSet:
    name: rs0
  replicas: 4
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: mg-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

The operator expands this into one replica set whose members are pinned to `dc-a` and
`dc-b`, plus a single MongoDB voting arbiter in `dc-c`, and steers `priority` from the
Lease so the primary stays in the active DC.

## Observe the DC-DR state

The single `MongoDB` object's `status.disasterRecovery` carries the whole cross-DC
view:

```bash
$ kubectl get mongodb -n demo mg-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-a",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    { "clusterName": "dc-a", "role": "Member",  "primary": "mg-dcdr-0", "writable": true,  "healthy": true, "oplogLagSeconds": 0 },
    { "clusterName": "dc-b", "role": "Member",  "primary": "",          "writable": false, "healthy": true, "oplogLagSeconds": 2 },
    { "clusterName": "dc-c", "role": "Arbiter", "primary": "",          "writable": false, "healthy": true }
  ]
}
```

- `activeDC` is the DC whose members hold the higher priority and run the elected
  primary.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports the DC role, its elected primary pod (if any),
  whether it is the writable DC, its health, and its cross-DC `oplogLagSeconds` (the
  optime delta behind the active primary, computed in-DC; the hub never opens
  cross-cluster connections).

## Unplanned failover

When the active DC is lost, MongoDB's own election promotes a new primary in a
surviving DC using the survivor plus the arbiter DC (or the surviving data majority in
the odd layout). You do not trigger this. The orchestrator observes the new primary
and moves the Lease to match, then (in the 2 + 2 + 1 layout) issues a
majority-committed `replSetReconfig` dropping the lost members so `w:majority` writes
resume. `status.disasterRecovery.phase` moves to `FailingOver` and back to `Steady`.

## Planned switchover (near-zero RPO)

To move the active DC on purpose without losing committed writes, annotate the
MongoDB with the target DC:

```bash
$ kubectl annotate mongodb -n demo mg-dcdr dr.kubedb.com/switchover-to=dc-b
```

The orchestrator raises the target DC's `priority`, then issues a non-force
`replSetStepDown` on the current primary. A non-force stepDown only succeeds when an
electable secondary in the target is caught up within the catch-up window, which is
the near-zero-RPO gate. The Lease then follows to `dc-b`.

## Cleanup

```bash
$ kubectl delete mongodb -n demo mg-dcdr
$ kubectl delete placementpolicy mg-dcdr
```

Deleting the `MongoDB` removes the per-DC member groups, the arbiter, and the
generated per-DC `PlacementPolicies`. The user-provided base `PlacementPolicy` is left
for you to delete.
