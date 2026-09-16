---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: ch-dr-guide-clickhouse
    name: User Guide
    parent: ch-dr-clickhouse
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running ClickHouse in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed ClickHouse in cross data
center disaster recovery (DC-DR) mode: the components, the naming contract, deployment,
connecting through the single write endpoint, reading locally, monitoring, lag and RPO,
Keeper placement, switchover and failback, scaling, and day-2 operations.

Read the [DC-DR Overview](/docs/guides/clickhouse/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/clickhouse/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the DC the single write endpoint resolves to. The Lease is routing, policy, and observability, **not** the failover mechanism. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease for its DC and projects the Lease decision into the local spoke as the `primary-dc` marker. |
| **KubeDB ClickHouse operator (hub)** | the OCM hub | Expands the `ClickHouse` CR into per-DC `ReplicatedMergeTree` replica groups, spreads the 3-site Keeper ensemble, routes the single write endpoint by following the Lease, drives planned switchover, and writes `status.disasterRecovery`. |
| **ClickHouse Keeper ensemble** | a voter in each Member DC plus a data-less voter in the Arbiter DC | The Raft service that registers parts and orders replication. **Its Raft majority is the failover authority and the split-brain guarantee.** No ZooKeeper. |
| **KubeSlice** | each spoke | Provides the cross-DC pod network so the one logical cluster spans clusters and `ReplicatedMergeTree` replicates over port 9009, coordinated by Keeper on 9181/9234. |

## The DC-name contract

One string identifies a data center everywhere. **Keep these identical:**

- the OCM spoke cluster name
- the agent `--dc-name`
- the primary-DC Lease `holderIdentity`
- the marker `data.activeDC`
- the pod label `open-cluster-management.io/cluster-name`
- the `PlacementPolicy` `distributionRule.clusterName`

## Deploying

### PlacementPolicy

Map the global replica indices to data centers and tag each DC with its role:

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
        scope: Global       # one cluster-wide failover scope (or Group + a group name)
      mode: TwoDC           # TwoDC: 2 Member DCs + an Arbiter DC; ThreeDC: 3 Member DCs
    distributionRules:
    - clusterName: dc-a
      role: Member
      replicaIndices: [0, 1]
    - clusterName: dc-b
      role: Member
      replicaIndices: [2, 3]
    - clusterName: dc-c
      role: Arbiter
      replicaIndices: []
```

- A data-bearing **Member** rule carries `replicaIndices`; the **Arbiter** DC carries an
  empty list. Its single data-less Keeper voter is scheduled onto the arbiter spoke and
  co-located with the third etcd member (it is not ordinal-pinned; the operator reads the
  Arbiter DC's `clusterName` and schedules the voter onto that spoke via OCM
  ManifestWork).
- `mode: TwoDC` expects two Member DCs plus the Arbiter DC (the even layout); `ThreeDC`
  expects an odd number of Member DCs and no separate Arbiter DC (each data DC then holds
  its own Keeper voter, keeping Keeper quorum among the data DCs).
- Roles are `Member` and `Arbiter` only.

### ClickHouse

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

### What the operator creates

- **One logical cluster** whose shards each have a `ReplicatedMergeTree` replica in each
  Member DC, all sharing one Keeper ensemble and the same `default_replica_path` macros.
  Replication is the native ClickHouse link over port 9009; there is no second
  replication link.
- A **3-site Keeper ensemble**: a voter in each Member DC and, in the even (`TwoDC`)
  layout, one data-less Keeper voter scheduled onto the Arbiter DC. This is what lets the
  engine's own quorum survive a DC loss and fence a minority.
- A single **write endpoint** (Service plus `AppBinding`) that the orchestrator points at
  the active DC by following the Lease. Reads can go to any replica.

All data-bearing pods carry the offshoot selectors plus the
`open-cluster-management.io/cluster-name` label, so the single write endpoint and the
single `AppBinding` keep working as the active DC moves.

> The macros and `default_replica_path` are consistent across DCs so a shard's replicas
> in different DCs share the same Keeper path and replicate to each other. Do not change
> them per DC.

## Connecting

A DC-DR ClickHouse exposes a single write endpoint, the same shape as any KubeDB
ClickHouse:

- the **write endpoint** `<db>` resolves to the active DC's replicas (native port 9000);
  the Lease-driven fence keeps it off a non-active DC, fail-closed;
- one **`AppBinding`** `<db>` for applications and KubeDB integrations.

Because ClickHouse is multi-master, every replica can technically accept writes, but the
single endpoint gives a stable single-writer posture: applications keep using `<db>` and,
after a failover, reconnect and land on the new active DC.

### Writes and the Keeper-quorum contract

There is no `w:majority` knob to set: the split-brain guarantee is built into the engine.
Every insert registers a part in Keeper, which needs Keeper quorum. A partitioned
minority DC that has lost Keeper quorum simply cannot register parts, so its inserts
fail. You do not have to opt into this; it is how `ReplicatedMergeTree` plus a spread
3-site Keeper behaves.

```sql
-- Against the write endpoint <db>:9000:
INSERT INTO orders (item, qty) VALUES ('widget', 1);
```

- On a spread 3-site Keeper (topology A), an insert commits only when the writing DC
  holds Keeper quorum, so a cut-off minority DC cannot commit and there is no split
  brain.
- The bounded loss on an unplanned active-DC loss is only committed-but-unfetched parts
  (registered in Keeper, data still on the lost DC's disk), which are recoverable when
  that DC returns.

### Read locally

Any replica serves consistent-enough reads. Point read traffic at an in-DC replica for
low latency; reads are eventually consistent, bounded by that replica's
`absoluteDelaySeconds`.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC the single write endpoint currently resolves to (a routing choice, not a promoted primary). |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | `Member` or `Arbiter`. |
| `dataCenters[].keeperVoter` | Whether the DC holds a Keeper voter. |
| `dataCenters[].keeperQuorum` | Whether the DC currently sees Keeper quorum (the safety signal). |
| `dataCenters[].writable` | True only for the active (write-routed) DC. |
| `dataCenters[].shards[]` | Per shard: `shard`, `totalReplicas`, `activeReplicas`. |
| `dataCenters[].absoluteDelaySeconds` | The DC's cross-DC replication delay behind the active DC, in seconds (from `system.replicas.absolute_delay`). |
| `dataCenters[].queueSize` | The DC's pending replication queue length (from `system.replicas.queue_size`). |
| `dataCenters[].healthy` | Whether the DC has ready replicas. |

### Useful checks

```bash
# Which DC the Lease intends as the write-routed active DC:
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# Per-DC replicas and DCs:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=ch-dcdr \
    -L open-cluster-management.io/cluster-name

# Cluster hosts and per-replica health (from any replica):
$ kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client \
    --query "SELECT cluster, host_name, replica_num FROM system.clusters"

# Replication delay and queue per replica (the lag signal):
$ kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client \
    --query "SELECT database, table, absolute_delay, queue_size, log_pointer, log_max_index, total_replicas, active_replicas FROM system.replicas"
```

## Replication, lag, and RPO

- Cross-DC replication is **native ClickHouse `ReplicatedMergeTree`** over port 9009,
  asynchronous, coordinated by the shared Keeper ensemble. There is exactly one logical
  cluster, so there is no extra replication link to manage.
- The lag signals come from `system.replicas`: `absolute_delay` (seconds a replica is
  behind), `queue_size`, and `log_pointer` versus `log_max_index` (how many log entries
  the replica still has to fetch). These are surfaced into `status.disasterRecovery` as
  `absoluteDelaySeconds` and `queueSize`.
- A **planned switchover loses near-zero committed writes**, because the orchestrator
  waits until the target DC's replicas show near-zero `absolute_delay` and an empty queue
  before it flips the endpoint. An **unplanned failover** may lose only
  committed-but-unfetched parts (bounded by the standby lag when the active DC died); a
  clean partition that put the lost DC in the Keeper minority loses zero committed data,
  because that DC could not commit without quorum.

## Keeper placement and the arbiter

- **Keeper is spread 3-site so no single data DC holds a Keeper majority.** With two data
  DCs, each holds a voter and the Arbiter DC holds one data-less voter; the majority is 2
  of 3, so either data DC plus the Arbiter DC keeps quorum, but neither data DC alone
  does. This removes split brain at its root: a partitioned data DC cannot register parts
  by itself.
- The requirement is an **odd Keeper voter total**, not an odd DC count:
  - **Even layout** (two data DCs plus the Arbiter DC): one Keeper voter per data DC and
    one data-less voter in the Arbiter DC (total 3). Do not add extra voters that would
    let one data DC hold a majority.
  - **Odd layout** (three or more Member DCs, no Arbiter DC): one Keeper voter per data
    DC, for an odd total, keeping Keeper quorum among the data DCs. This is the layout the
    DC-count rule prefers.
- The **Arbiter DC** holds the `dr-controlplane` etcd vote **and** one data-less Keeper
  voter, co-located so the two quorums agree on which DCs are alive. This is the same
  arbiter trick as MongoDB.

> For write-heavy ingest, weigh the two alternatives from the overview: two clusters with
> a per-region Keeper cross-replicated (topology B) avoids the cross-DC Keeper round trip,
> and single-DC Keeper (topology C) is the lowest-latency manual-failover floor.

## Planned switchover (near-zero RPO)

Move the active (write-routed) DC on purpose by annotating the ClickHouse:

```bash
$ kubectl annotate clickhouse -n demo ch-dcdr dr.kubedb.com/switchover-to=dc-b
```

The hub then:

1. checks the target is a known, healthy DC within the lag budget;
2. sets `phase: FailingOver` and quiesces writes on the current active DC (routes clients
   away from the endpoint);
3. waits until the target DC's replicas report `absolute_delay` near zero and an empty
   replication queue (`queue_size: 0`), the catch-up gate that makes this near-zero RPO;
4. moves the Lease and the single write endpoint to `dc-b`.

Watch `status.disasterRecovery` for `phase` returning to `Steady` with the new
`activeDC`. There is no promotion step, because every replica is already writable.

## Failback

Failback is native and clean. A returned DC's replicas rejoin the Keeper ensemble and
catch up through `ReplicatedMergeTree` (they fetch the missing parts). There is **no
rewind**: a partitioned DC that lacked Keeper quorum committed nothing to diverge, so
there is nothing to roll back.

Once the returned DC is caught up (near-zero `absoluteDelaySeconds`, empty queue), steer
the active DC back with a planned switchover:

```bash
$ kubectl annotate clickhouse -n demo ch-dcdr dr.kubedb.com/switchover-to=dc-a
```

## Scaling and day-2 operations

The standard `ClickHouseOpsRequest` operations (`VerticalScaling`, `HorizontalScaling`,
`VolumeExpansion`, `UpdateVersion`, `Reconfigure`, `ReconfigureTLS`, `Restart`,
`RotateAuth`, `StorageMigration`) apply to a DC-DR cluster. They act on the distributed
per-DC replica groups across the DCs and are issued exactly as for a single-cluster
ClickHouse. There is no failover ops type: failover is the engine's Keeper quorum, and
the planned switchover is the `dr.kubedb.com/switchover-to` annotation, not an ops
request.

`HorizontalScaling` gains a per-DC form so you can scale each DC's per-shard replica
count independently:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-dcdr-hscale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ch-dcdr
  horizontalScaling:
    dataCenters:
    - clusterName: dc-a
      replicas: 3
    - clusterName: dc-b
      replicas: 2
```

> **Note:** the distributed ClickHouse substrate and the DC-DR layer are net-new for
> ClickHouse. Treat the field names and flows in this guide as the intended user
> experience; confirm availability in your release before relying on them in production.

## Deletion and cleanup

```bash
$ kubectl delete clickhouse -n demo ch-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC replica groups, the arbiter-DC
Keeper voter, and the cluster-scoped per-DC `PlacementPolicies` it generated (these carry
no owner reference, so the operator deletes them explicitly). The user-provided base
`PlacementPolicy` is left for you to delete.

## Limitations

- **Adding or removing a whole data center** is a topology change (a replica-group and
  Keeper-ensemble change), performed by editing the `PlacementPolicy` topology, not by a
  scaling request.
- Cross-DC `ReplicatedMergeTree` replication is asynchronous; an unplanned failover has a
  non-zero RPO bounded by the standby lag (only committed-but-unfetched parts). A clean
  partition loses zero committed data because the minority DC cannot commit without Keeper
  quorum; use a planned switchover for a near-zero-RPO move.
- On a spread 3-site Keeper (topology A), every insert pays a cross-DC Keeper round trip.
  For write-heavy ingest, consider the two-cluster per-region-Keeper topology (B) or the
  single-DC Keeper topology (C) described in the overview.
