---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-dr-overview
    name: Overview
    parent: guides-postgres-dr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for Postgres

KubeDB can run a single distributed `Postgres` across multiple data centers so the
database survives the loss of an entire data center (DC). Exactly one DC is writable
at any instant; the others are warm, read-only standbys that stream from it across
the DCs. When the active DC is lost, KubeDB promotes a surviving DC, and the single
connection endpoint follows the new writable DC.

This guide builds on [Distributed Postgres](/docs/guides/postgres/distributed/overview/index.md).
Read that first: DC-DR reuses the same substrate (one CR, Open Cluster Management,
KubeSlice, and a `PlacementPolicy`) and adds the cross-DC failover machinery on top.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/postgres/dr/guide/index.md) — every
  aspect of running in DC-DR mode (components, monitoring, timing, scaling, day-2 ops).
- [DC-DR Runbook](/docs/guides/postgres/dr/runbook/index.md) — what to
  do in each operational scenario.

> **New to KubeDB?** Please start [here](/docs/README.md).

## How it works

DC-DR is built on one rule: **the Postgres raft never stretches across data centers.**

- **Each data center is a self-contained Postgres group.** The operator expands the
  single `Postgres` CR into one group per data-bearing DC, each with its own
  `pg-coordinator` raft electing a **local** leader, its own local replicas, and (when
  its local replica count is even) its own local arbiter. The raft quorum never
  crosses the DC boundary, so cross-DC latency or a partition can never flap an
  election.
- **One cross-DC authority decides who is writable.** A small control plane
  (`dr-controlplane`), backed by a three-site etcd quorum, publishes one
  `coordination.k8s.io` **Lease** per failover scope. The DC that holds the Lease is
  the **active** (writable) DC. This is the single cross-DC failover decision.
- **Cross-DC replication is leader-to-leader streaming.** The standby DC's local
  leader runs as an asynchronous streaming standby of the active DC's leader; that
  standby DC's own replicas cascade from its local leader. So a standby DC opens
  exactly one cross-DC replication link.
- **Writability is fenced locally and fails closed.** A per-DC `dr-controlplane`
  agent projects the Lease holder onto its own spoke cluster as a small marker
  `ConfigMap`. The `pg-coordinator` reads only that local marker: if it cannot
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

Each DC plays one role, set on the `PlacementPolicy` `distributionRule`:

| Role | Holds Postgres data | Primary eligible | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A full Postgres group; a candidate for the active DC. |
| **Arbiter** | no | no | Votes only, to break ties; runs no Postgres. |
| **Witness** | no (for Postgres) | no | Holds only the `dr-controlplane` etcd member / control-plane vote. |

A typical layout is two Member DCs plus one Witness DC (the three-site etcd quorum
lives across all three, but Postgres data lives only in the two Member DCs).

## Prerequisites

- A working [Distributed Postgres](/docs/guides/postgres/distributed/overview/index.md)
  setup: Open Cluster Management (OCM) hub and spoke clusters, KubeSlice connecting
  the spokes, and a storage class on each spoke.
- The `dr-controlplane` service and its three-site etcd quorum installed across the
  data centers, with a `dr-controlplane` agent running in each spoke (DC).
- The KubeDB Postgres operator started with the DC-DR flags:

  ```
  --dc-dr-enabled
  --dc-dr-coord-kubeconfig=<path to the coordination control plane kubeconfig>
  --dc-dr-local-dc=<this operator's data center name>
  ```

- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster
  name, the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the
  pod label `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR Postgres

A DC-DR Postgres is a distributed `Postgres` whose `PlacementPolicy` carries a
`failoverPolicy` and per-DC roles. The user creates and edits a **single** `Postgres`
object and gets one `AppBinding` and one connection endpoint; the operator expands it
into the per-DC groups.

### 1. PlacementPolicy

Assign the global pod ordinals to data centers and tag each DC with its role. Here two
Member DCs (`dc-east`, `dc-west`) each get three Postgres pods, and `dc-arbiter` is the
tie-breaking witness:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: pg-dcdr
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
      role: Witness
```

- A **Member** or **Witness** rule that bears data carries `replicaIndices`; an
  **Arbiter** / Postgres **Witness** rule carries none.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope.
  Use `Group` with a group name to put a database in its own scope.

### 2. Postgres

Reference the `PlacementPolicy` and opt the Postgres into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-dcdr
  namespace: demo
  annotations:
    # Opt this distributed Postgres into per-DC DC-DR expansion.
    dr.kubedb.com/enabled: "true"
spec:
  version: "17.2"
  replicas: 6
  distributed: true
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: pg-dcdr
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

The operator then creates, per data-bearing DC:

- a per-DC `PetSet` named `<db>-<dc>` (for example `pg-dcdr-dc-east`) with its own
  intra-DC raft and DC-local governing `Service`;
- a per-DC arbiter `PetSet` `<db>-<dc>-arbiter` when that DC's local node count is
  even.

The Witness DC runs no Postgres pods.

## Observe the DC-DR state

The single `Postgres` object's `status.disasterRecovery` carries the whole cross-DC
view:

```bash
$ kubectl get pg -n demo pg-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-east",
  "phase": "Steady",
  "lastTransitionTime": "2026-06-30T10:00:00Z",
  "dataCenters": [
    { "name": "dc-east", "role": "primary", "leader": "pg-dcdr-dc-east-0", "writable": true,  "healthy": true },
    { "name": "dc-west", "role": "standby", "leader": "pg-dcdr-dc-west-0", "writable": false, "healthy": true, "lagBytes": 4096 }
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
through a `ForceFailOver` `PostgresOpsRequest`, while the old DC self-fences read-only
locally (before the hub reacts, even under a partition). The primary `Service` and
`AppBinding` then resolve to the new writable DC.

You do not trigger this; it is automatic. `status.disasterRecovery.phase` moves to
`FailingOver` during the transition and back to `Steady` once the survivor is primary.

## Planned switchover (zero-RPO)

To move the active DC on purpose (maintenance, rebalancing) without losing committed
rows, annotate the Postgres with the target DC:

```bash
$ kubectl annotate pg -n demo pg-dcdr dr.kubedb.com/switchover-to=dc-west
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
scale. Scale a specific DC with a `PostgresOpsRequest` that lists per-DC targets:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-dcdr-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pg-dcdr
  horizontalScaling:
    dataCenters:
    - name: dc-west
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

The standard Postgres `PostgresOpsRequest` operations work on a DC-DR cluster and act
on every per-DC group: vertical scaling, volume expansion (online and offline),
version update, and storage migration each apply to all per-DC `PetSet`s and per-DC
arbiters. You issue them exactly as for a non-distributed Postgres.

## Cleanup

```bash
$ kubectl delete pg -n demo pg-dcdr
$ kubectl delete placementpolicy pg-dcdr
```

Deleting the `Postgres` removes the per-DC `PetSet`s, governing `Service`s, and the
cluster-scoped per-DC `PlacementPolicies` the operator generated. The user-provided
base `PlacementPolicy` is left for you to delete.
