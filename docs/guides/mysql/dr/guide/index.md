---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-dr-guide
    name: User Guide
    parent: guides-mysql-dr
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running MySQL in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed MySQL in cross data center
disaster recovery (DC-DR) mode: the components, the naming contract, deployment,
connecting, monitoring, replication and lag, timing and tuning, quorum and roles,
switchover and failback, scaling, day-2 operations, backup, and deletion.

Read the [DC-DR Overview](/docs/guides/mysql/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/mysql/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the active (writable) DC. This is the single cross-DC failover authority. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease on behalf of its DC and projects the Lease decision into the local spoke as a marker `ConfigMap`. |
| **KubeDB MySQL operator (hub)** | the OCM hub | Expands the `MySQL` CR into per-DC GR clusters, watches the Lease, drives failover/switchover, and writes `status.disasterRecovery`. |
| **`mysql-coordinator`** | every MySQL pod | Manages GR membership and the `kubedb.com/role` label, reads the local marker, and fences its GR primary `super_read_only` when its DC is not active; runs the cross-DC channel on a standby DC's primary. |
| **KubeSlice** | each spoke | Provides the cross-DC pod network so a standby DC's GR primary can stream from the active DC's primary endpoint. |

The marker `ConfigMap` is the contract between the agent (producer) and the coordinator
(consumer):

```
ConfigMap primary-dc  (namespace: dc-failover, on each spoke)
  data.activeDC  = the DC the quorum currently trusts as primary
  data.renewTime = RFC3339, the observed primary-DC Lease renewTime
```

The coordinator trusts the marker for 30s (the fence TTL); absent, stale, unparseable, or
naming another DC all mean *not active* and the GR primary stays `super_read_only`. This is
the fail-closed fence.

## The DC-name contract

One string identifies a data center everywhere. **Keep these identical:**

- the OCM spoke cluster name
- the agent `--dc-name`
- the primary-DC Lease `holderIdentity`
- the marker `data.activeDC`
- the pod label `open-cluster-management.io/cluster-name`
- the `PlacementPolicy` `distributionRule.clusterName`

## Operator configuration

Start the MySQL operator with:

```
--dc-dr-enabled
--dc-dr-coord-kubeconfig=<kubeconfig of the coordination control plane>
--dc-dr-local-dc=<the data center this operator instance runs in>
```

The per-DC pod coordinators automatically receive `DC_DR_ENABLED`, `DC_NAME`,
`DC_DR_NAMESPACE` (default `dc-failover`), `DC_DR_MARKER` (default `primary-dc`), and
`DC_DR_SOURCE_HOST` (the active DC's primary endpoint the standby channels from) through
their PetSet template, so the fence and the cross-DC channel work without extra wiring.

## Deploying

### PlacementPolicy

Map the global pod ordinals to data centers and tag each DC with its role:

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
        scope: Global       # one cluster-wide failover scope (or Group + a group name)
      mode: TwoDC           # TwoDC: 2 Members + a tie-breaker; ThreeDC: 3 Members
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
  (vote only, no MySQL) carries none. Group Replication has no data-less voter member, so a
  MySQL witness DC is always `role: Arbiter` (the petset `Witness` role, a data-bearing
  witness, is for engines like MongoDB and is not used by MySQL).
- `mode: TwoDC` expects exactly two Member DCs plus the Arbiter witness DC; `ThreeDC`
  expects at least three Member DCs.
- Give each Member DC an **odd** local node count so its GR group keeps a clean majority.

### MySQL

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: my-dcdr
  namespace: demo
  annotations:
    dr.kubedb.com/enabled: "true"          # opt into per-DC DC-DR expansion
    # dr.kubedb.com/failover-group: payments  # optional: a Group failover scope
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

### What the operator creates

Per data-bearing DC `<dc>`:

- a per-DC `PetSet` `<db>-<dc>` (e.g. `my-dcdr-dc-east`) with its own intra-DC GR group
  (its own `group_replication_group_name` UUID);
- a DC-local headless governing `Service`, exported over KubeSlice, so the DC's pods
  discover only each other;
- a cluster-scoped per-DC `PlacementPolicy` `<base>-<dc>` pinning that group to the DC;
- the cross-DC async channel on each standby DC's GR primary.

The witness DC (`role: Arbiter`) runs no MySQL pods. All per-DC pods carry the offshoot
selectors plus the `open-cluster-management.io/cluster-name` label, so the global
primary/standby Services and the single `AppBinding` keep working.

## Connecting

A DC-DR MySQL exposes the same single endpoint as any KubeDB MySQL:

- the **primary Service** `<db>` resolves to the active DC's writable GR primary (only that
  primary is labeled `kubedb.com/role: primary`);
- the **standby Service** `<db>-standby` resolves to the read-only members;
- one **`AppBinding`** `<db>` for applications and KubeDB integrations.

Because only the active DC's primary carries the `primary` label, the endpoint follows
failover automatically — applications keep using `<db>` and reconnect after a failover,
landing on the new active DC.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC that holds the Lease and runs the writable GR primary. |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | The DC role: `Member` or `Arbiter`. |
| `dataCenters[].primary` | That DC's local GR primary pod. |
| `dataCenters[].writable` | True only for the active DC. |
| `dataCenters[].lagBytes` | The DC's cross-DC GTID gap behind the active primary, in bytes. |
| `dataCenters[].secondsBehindSource` | The DC's `Seconds_Behind_Source` on its cross-DC channel. |
| `dataCenters[].healthy` | Whether the DC's GR group is healthy. |

### Useful checks

```bash
# Which DC is active (from the coordination plane):
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# The marker each spoke reads (run against a spoke):
$ kubectl -n dc-failover get configmap primary-dc -o yaml

# Per-DC GR primaries and roles:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=my-dcdr \
    -L kubedb.com/role,open-cluster-management.io/cluster-name

# On a standby DC's GR primary, the cross-DC channel:
#   SHOW REPLICA STATUS FOR CHANNEL 'dcdr'\G   (Replica_IO_Running, Replica_SQL_Running, Seconds_Behind_Source)
```

## Replication, lag, and RPO

- Cross-DC replication is an **asynchronous** replication channel on the standby DC's GR
  primary; GR then distributes intra-DC, so each standby DC opens exactly one cross-DC
  link. On a GR member the channel is a **named** channel (`dcdr`).
- Lag is reported two ways: `lagBytes` (the GTID gap, `gtid_executed` on the active primary
  vs the standby) and `secondsBehindSource` (`Seconds_Behind_Source` from
  `SHOW REPLICA STATUS`). The DC's coordinator computes both (the hub never opens
  cross-cluster SQL). They are the basis for the RPO of an unplanned failover.
- A **planned switchover loses no committed rows** because writes are frozen and the target
  catches up by GTID before the handoff. An **unplanned failover** may lose the last
  unreplicated transactions (bounded by the standby's lag at the moment the active DC died).

## Timing and tuning (RTO vs safety)

DC-DR has one timing invariant that must hold for correctness:

> **fence TTL + cross-DC clock skew < primary-DC Lease duration**

The marker `renewTime` tracks the Lease's renewTime. A partitioned active DC self-fences at
`lastRenew + fence TTL`; a survivor can only acquire the expired Lease at
`lastRenew + LeaseDuration`. Keeping the fence TTL inside the Lease duration guarantees the
old active DC goes `super_read_only` **before** any new DC becomes writable — no split-brain
window.

Default values:

| Parameter | Where | Default |
| --- | --- | --- |
| Fence TTL | mysql-coordinator | 30s |
| Marker refresh interval | dr-controlplane agent | 5s |
| Primary-DC Lease duration | dr-controlplane agent (`--election-lease-duration`) | 45s |
| Lease renew deadline | dr-controlplane agent (`--election-renew-deadline`) | 30s |
| Lease retry period | dr-controlplane agent (`--election-retry-period`) | 2s |

The failover **RTO floor** is roughly the Lease duration (the time a survivor waits to
acquire). To lower RTO, lower the Lease duration **and** the fence TTL together, always
preserving `fence TTL + skew < LeaseDuration`. The retry period must stay well under the
fence TTL so the holder restamps `renewTime` and the marker reads fresh in normal
operation.

## Quorum, roles, and arbiters

- Each DC's GR group needs its own majority. Give each Member DC an **odd** local node count
  so an intra-DC primary failure keeps quorum. Unlike a raft-based engine, KubeDB does not
  add a data-less arbiter member inside a GR group, so avoid even local group sizes.
- The witness DC (`role: Arbiter`) holds only the `dr-controlplane` vote, never MySQL data.
- One GR subtlety the fence handles: GR sets its own elected primary `super_read_only = OFF`,
  so the coordinator re-asserts the fence on every label loop. The fail-closed split-brain
  guarantee depends on that re-assertion winning the race after each intra standby-DC GR
  election.

Separately from a DC's *intra-DC* quorum, the **cross-DC** failover quorum needs a majority
of three voting sites. For how to lay this out across two or three data centers (and why a
third witness site is preferred), see
[Deployment topologies](/docs/guides/mysql/dr/overview/index.md#deployment-topologies-2-dcs-vs-3-dcs).

## Planned switchover (near-zero-RPO)

Move the active DC on purpose by annotating the MySQL:

```bash
$ kubectl annotate my -n demo my-dcdr dr.kubedb.com/switchover-to=dc-west
```

The hub then:

1. checks the target is a known, healthy DC within the lag budget;
2. sets `phase: FailingOver` and quiesces the active DC (holds its GR primary
   `super_read_only = ON`), freezing the active `gtid_executed`;
3. waits until the target's channel has applied up to that frozen GTID set;
4. hands the Lease to the target, whose GR primary is promoted; the old DC resumes as a
   standby and starts a channel from the new active.

The annotation is cleared automatically once the target is active. Watch
`status.disasterRecovery` for `phase` returning to `Steady` with the new `activeDC`.

## Failback

Failback is just a switchover back to the original DC once it is healthy again:

```bash
$ kubectl annotate my -n demo my-dcdr dr.kubedb.com/switchover-to=dc-east
```

A DC that lost the Lease and rejoins starts a channel from the new active and catches up by
GTID auto-positioning. If it diverged beyond the source's purged GTIDs (its forked tail
from a bounded-loss promotion), it is re-seeded (the MySQL clone plugin, or
`RESET REPLICA ALL` plus re-provision) before it is eligible.

## Per-DC horizontal scaling

Each DC has its own GR group, so scale a specific DC with a `MySQLOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-dcdr-scale-west
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

- Each entry sets that DC's local GR node count; DCs not listed are unchanged.
- Nodes are added or removed over the DC-local network; new members join the DC's GR group
  and recover from its primary.
- The base `PlacementPolicy` is renumbered so the declarative topology matches.
- Keep each Member DC's count **odd** for a clean GR majority. Removing a whole DC is a
  topology change, not horizontal scaling.

## Day-2 operations

The standard `MySQLOpsRequest` operations apply to every per-DC GR cluster on a DC-DR
cluster; issue them exactly as for a non-distributed MySQL:

| Operation | DC-DR behavior |
| --- | --- |
| **Vertical scaling** | Patches every per-DC PetSet, restarts per-DC pods. |
| **Volume expansion** (online/offline) | Expands every per-DC data PVC, waits on all per-DC PetSets. |
| **Version update** | Updates every per-DC PetSet. |
| **Storage migration** | Orphan-deletes and waits on every per-DC PetSet. |
| **Reconfigure / Restart / Rotate-Auth / Reconfigure-TLS** | Apply across the per-DC pods. |

## Backup

Back up a DC-DR MySQL the same way as any KubeDB MySQL (KubeStash / the MySQL archiver).
Backups run against the writable endpoint, so they read from the active DC; the AppBinding
follows failover, so a scheduled backup continues against the new active DC after a
failover. Point-in-time recovery works as usual.

## DC-aware health

On a DC-DR cluster the operator's health check is DC-aware:

- the **active** DC is expected to have a writable primary (the write-check runs there);
- a **standby** DC is expected to be `super_read_only` with a healthy inbound channel
  (`Replica_IO_Running = Yes`, `Replica_SQL_Running = Yes`), so the passive DC does not flap
  to `NotReady`. Set `spec.healthChecker.disableWriteCheck: true` semantics apply on standby
  DCs;
- the GR `ONLINE`-member count is checked against the **per-DC** group size, not the global
  `spec.replicas`.

## Deletion and cleanup

```bash
$ kubectl delete my -n demo my-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC PetSets, governing Services, and the
cluster-scoped per-DC `PlacementPolicies` it generated (these carry no owner reference, so
the operator deletes them explicitly). The user-provided base `PlacementPolicy` is left for
you to delete.

## Limitations

- **Adding or removing a whole data center** is a topology change (a new GR group and a
  cross-DC seed), distinct from horizontal scaling, and is performed by editing the
  `PlacementPolicy` topology, not a `HorizontalScaling` request.
- Cross-DC replication is asynchronous; an unplanned failover has a non-zero RPO bounded by
  the standby lag. Use a **planned switchover** for zero-RPO moves.
- A standby DC that diverges beyond the active's purged GTIDs must be re-seeded (clone) on
  failback rather than GTID catching up.
- All correctness depends on the timing invariant above; do not set a fence TTL that meets
  or exceeds the Lease duration.
