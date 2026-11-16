---
title: DC-DR Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mssqlserver-dr-overview
    name: Overview
    parent: guides-mssqlserver-dr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cross Data Center Disaster Recovery (DC-DR) for MSSQLServer

KubeDB can run a single distributed `MSSQLServer` across multiple data centers so the
database survives the loss of an entire data center (DC). Exactly one DC is writable at any
instant; the other is a warm, read-only standby that streams from it across the DCs. When
the active DC is lost, KubeDB promotes the surviving DC, and the single connection endpoint
follows the new writable DC.

SQL Server is the closest native fit of all the engines, because Always On **Distributed
Availability Groups (DAG)** are already a cross-cluster geo-DR mechanism. DC-DR mostly
**automates and fences** what KubeDB already builds manually with a DAG: it makes the
`dr-controlplane` Lease the authority that drives the DAG role, adds a fail-closed fence and
a lag guard, and presents one CR over the native two-AG construct.

DC-DR targets `spec.topology.mode: DistributedAG`. The intra-DC HA is an Always On
**Availability Group (AG)** with `CLUSTER_TYPE = EXTERNAL`, driven by the raft-based
**mssql-coordinator**; the cross-DC link is the native **DAG** over the DB-mirroring
endpoint on port 5022.

This page is the conceptual overview and a quick start. See also:

- [DC-DR User Guide](/docs/guides/mssqlserver/dr/guide/index.md) for every aspect of
  running in DC-DR mode (components, monitoring, timing, scaling, day-2 ops).
- [DC-DR Runbook](/docs/guides/mssqlserver/dr/runbook/index.md) for scenario-by-scenario
  procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## How it works

DC-DR is built on one rule: **the Availability Group (and its coordinator raft) never
stretches across data centers; the DAG is the only cross-DC link.**

- **Each data center is a self-contained AG.** The operator expands the single
  `MSSQLServer` CR into one AG per data-bearing DC. On Kubernetes the AG has no WSFC: it
  runs `CLUSTER_TYPE = EXTERNAL` with the `mssql-coordinator` raft as the external cluster
  manager that provides quorum and drives intra-AG failover. The AG's `SYNCHRONOUS_COMMIT`
  replicas, the coordinator raft, and a local Arbiter (if the AG node count is even) are all
  intra-DC. The AG quorum never crosses the DC boundary, so cross-DC latency or a partition
  cannot flap the intra-AG primary election.
- **One cross-DC authority decides who is writable.** A small control plane
  (`dr-controlplane`), backed by a three-site etcd quorum, publishes one
  `coordination.k8s.io` **Lease** per failover scope. The DC that holds the Lease is the
  **active** (writable) DC, and its AG is the **DAG primary**. This is the single cross-DC
  failover decision.
- **Cross-DC replication is the native DAG (async).** The standby DC's AG is the DAG
  **secondary (forwarder)**, receiving the `ASYNCHRONOUS_COMMIT` DAG stream over 5022 and
  applying it synchronously within its own AG. A standby DC auto-seeds its databases from
  the primary AG with `SEEDING_MODE = AUTOMATIC`, so there is no backup or restore step. An
  intra-active-DC AG failover (the coordinator electing a new AG primary within the active
  DC) is transparent to the DAG, because the DAG links the two AGs by listener (the primary
  Service on 5022), which follows whichever replica is the AG primary, so the forwarder
  keeps streaming without reconfiguration.
- **Writability is fenced locally and fails closed.** A per-DC `dr-controlplane` agent
  projects the Lease holder onto its own spoke cluster as a small marker `ConfigMap`. The
  fence reads only that local marker: if it cannot confirm its DC holds the Lease (the DC
  lost it, or is partitioned from the coordination plane), it forces the local AG to the DAG
  **SECONDARY** role (`ALTER AVAILABILITY GROUP [dag] SET (ROLE = SECONDARY)`), whose
  databases are non-writable. Because the fence lives in the DC and fails closed, a cut-off
  old-active DC stops accepting writes on its own, before the hub even reacts. This local
  fence plus the etcd majority (only one DC can hold the Lease) is the split-brain
  guarantee.
- **Only the active DC's AG primary is labeled `primary`.** Each DC's coordinator elects its
  own AG primary, but a non-active DC's AG primary (the DAG forwarder) is labeled
  `kubedb.com/role: standby`, so the single primary `Service` and the `AppBinding` always
  resolve to the active DC's writable AG primary.
- **Two raft layers, cleanly separated.** The `mssql-coordinator` raft decides the
  **intra-AG** primary (intra-DC). The `dr-controlplane` Lease decides the **active AG**
  (cross-DC, the DAG role). They never mix: the coordinator never spans DCs; the Lease never
  picks an intra-AG replica.

### Data center roles

Each DC plays one role, set on the `PlacementPolicy` `distributionRule.role`:

| Role | Holds SQL Server data | Primary eligible | Purpose |
| --- | --- | --- | --- |
| **Member** | yes | yes | A full AG; a candidate for the active DC (the DAG primary). |
| **Arbiter** | no | no | Vote only, the `dr-controlplane` etcd tie-breaker; runs no SQL Server. |

> A native Distributed AG joins **exactly two** AGs, so DC-DR for SQL Server targets the
> even layout: two data DCs plus a data-less Arbiter DC. Spanning three or more data DCs
> needs chained DAGs and is a separate, larger design that is out of scope here. The arbiter
> DC holds only the `dr-controlplane` etcd member and no SQL Server.

A typical layout is two Member DCs plus one vote-only Arbiter DC: the three-site etcd quorum
lives across all three sites, but SQL Server data lives only in the two Member DCs.

## Deployment topologies (2 DCs vs 3 DCs)

The DR feature needs two things, in different quantities:

- **SQL Server data** lives in the **Member** data centers. A native DAG joins exactly two
  AGs, so you have exactly **two** Member DCs (one active, one warm standby).
- **The failover decision** is made by the `dr-controlplane` etcd **quorum**. A quorum makes
  progress only while a **majority of its three voting sites** is reachable. For
  single-fault tolerance *and* split-brain safety, those three votes should sit in **three
  independent failure domains**. The third domain is a tiny vote-only **Arbiter**
  (`role: Arbiter`) that holds no SQL Server data.

So "how many data centers" has two answers: how many hold **data** (two) and how many hold a
**quorum vote** (always three for automatic, split-brain-free failover).

### A. Two Member DCs + an Arbiter DC (recommended)

Three sites; two hold SQL Server data, the third is a vote-only Arbiter DC (`role: Arbiter`,
no SQL Server):

```yaml
failoverPolicy:
  mode: TwoDC
distributionRules:
- { clusterName: dc-east, role: Member, replicaIndices: [0, 1, 2] }
- { clusterName: dc-west, role: Member, replicaIndices: [3, 4, 5] }
- { clusterName: dc-arbiter, role: Arbiter }    # etcd vote only, no SQL Server
```

Any single site can be lost:

- **Lose a Member DC** then the surviving Member plus the Arbiter form a 2/3 majority, so the
  survivor acquires the Lease and is promoted automatically; the lost DC, if alive but
  partitioned, self-fences read-only.
- **Lose the Arbiter** then the two Members are still a 2/3 majority, so writes continue
  uninterrupted.

Because the Arbiter runs no SQL Server, it is small and cheap. **Run it in a third region or
cloud.** This is the lowest-cost way to get correct, automatic failover, and it is the
recommended topology whenever a third location is available.

### B. Two sites only (reduced resiliency)

If you genuinely have only two locations, you still need a third quorum vote, so you **place
it inside one of the two DCs** (run the third `dr-controlplane` etcd member there). There is
no separate Arbiter site, so that DC now holds **two of the three votes**:

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

- A working **distributed MSSQLServer** setup: Open Cluster Management (OCM) hub and spoke
  clusters, KubeSlice connecting the spokes (it exports each AG's port 5022 endpoint
  cross-cluster for the DAG), and a storage class on each spoke.
- The `dr-controlplane` service and its three-site etcd quorum installed across the data
  centers, with a `dr-controlplane` agent running in each spoke (DC).
- The KubeDB MSSQLServer operator started with the DC-DR flags:

  ```
  --dc-dr-enabled
  --dc-dr-coord-kubeconfig=<path to the coordination control plane kubeconfig>
  --dc-dr-local-dc=<this operator's data center name>
  ```

- One consistent **DC name** per data center, used everywhere: the OCM spoke cluster name,
  the agent `--dc-name`, the Lease `holderIdentity`, the marker `activeDC`, the pod label
  `open-cluster-management.io/cluster-name`, and the `PlacementPolicy`
  `distributionRule.clusterName`. Keep them identical.

## Deploy a DC-DR MSSQLServer

A DC-DR MSSQLServer is a distributed `MSSQLServer` whose `PlacementPolicy` carries a
`failoverPolicy` and per-DC roles. The user creates and edits a **single** `MSSQLServer`
object and gets one `AppBinding` and one connection endpoint; the operator expands it into
the per-DC AGs and wires the DAG between them.

### 1. PlacementPolicy

Assign the global pod ordinals to data centers and tag each DC with its role. Here two
Member DCs (`dc-east`, `dc-west`) each get three SQL Server pods, and `dc-arbiter` is the
tie-breaking vote:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: mssql-dcdr
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
  SQL Server) carries none.
- `failoverPolicy.trigger.scope: Global` makes this one cluster-wide failover scope. Use
  `Group` with a group name to put a database in its own scope.
- Give each Member DC an **odd** local node count so its AG keeps a clean coordinator-raft
  majority for intra-DC failover; an even AG gets an auto-injected local Arbiter PetSet.

### 2. MSSQLServer

Reference the `PlacementPolicy` and opt the MSSQLServer into DC-DR expansion:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-dcdr
  namespace: demo
  annotations:
    # Opt this distributed MSSQLServer into per-DC DC-DR expansion.
    dr.kubedb.com/enabled: "true"
spec:
  version: "2022-cu16"
  replicas: 6
  distributed: true
  topology:
    mode: DistributedAG
    availabilityGroup:
      databases:
        - agdb
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: mssql-dcdr
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

The operator then creates, per data-bearing DC:

- a per-DC AG (its own `mssql-coordinator` raft and `SYNCHRONOUS_COMMIT` replicas) backed by
  a `PetSet` named `<db>-<dc>` (for example `mssql-dcdr-dc-east`) with a DC-local governing
  `Service` exported over KubeSlice;
- the native DAG joining the two AGs over their KubeSlice-exported 5022 endpoints, with the
  active DC's AG as the DAG primary and the standby DC's AG as the DAG secondary (forwarder).

The Arbiter DC (`role: Arbiter`) runs no SQL Server pods.

## Observe the DC-DR state

The single `MSSQLServer` object's `status.disasterRecovery` carries the whole cross-DC view:

```bash
$ kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

```json
{
  "activeDC": "dc-east",
  "phase": "Steady",
  "dataCenters": [
    {
      "clusterName": "dc-east", "role": "Member",
      "agPrimary": "mssql-dcdr-dc-east-0", "dagRole": "Primary",
      "writable": true, "synchronizationHealth": "HEALTHY",
      "lastHardenedLSN": "0x00000029000000A0001", "healthy": true
    },
    {
      "clusterName": "dc-west", "role": "Member",
      "agPrimary": "mssql-dcdr-dc-west-0", "dagRole": "Secondary",
      "writable": false, "synchronizationHealth": "HEALTHY",
      "redoQueueBytes": 65536, "logSendQueueBytes": 4096,
      "lastHardenedLSN": "0x00000029000000A0001", "healthy": true
    }
  ]
}
```

- `activeDC` is the DC that currently holds the Lease and runs the writable DAG primary AG.
- `phase` is `Steady`, `FailingOver`, `FailingBack`, or `Degraded`.
- Each `dataCenters` entry reports that DC's AG primary pod, its DAG role, whether it is
  writable, its `synchronizationHealth`, its `redoQueueBytes` / `logSendQueueBytes` and
  `lastHardenedLSN` (from `sys.dm_hadr_database_replica_states`), and whether it is healthy.
  The in-DC coordinator computes these and surfaces them; the hub never opens cross-cluster
  SQL.

## Unplanned failover

When the active DC is lost, its agents stop renewing the primary-DC Lease. After the Lease
duration the surviving Member DC's agent acquires it; that DC becomes `activeDC`. The hub
observes the change and clears the survivor's fence: it runs
`ALTER AVAILABILITY GROUP [dag] FORCE_FAILOVER_ALLOW_DATA_LOSS` on the survivor AG, flips
`self.role` to `Primary` on it, and relabels its AG primary `primary`. The old DC, if
partially alive, has already self-fenced its AG to DAG secondary. The primary `Service` and
`AppBinding` then resolve to the new writable DC.

You do not trigger this; it is automatic. `status.disasterRecovery.phase` moves to
`FailingOver` during the transition and back to `Steady` once the survivor is primary. The
RPO is bounded by the survivor's cross-DC lag at the moment the active DC died (the
un-shipped redo).

## Planned switchover (zero-RPO)

To move the active DC on purpose (maintenance, rebalancing) without losing committed rows,
annotate the MSSQLServer with the target DC:

```bash
$ kubectl annotate mssqlserver -n demo mssql-dcdr dr.kubedb.com/switchover-to=dc-west
```

The switchover is coordinated for zero RPO:

1. The target must be a known, healthy DC within the lag budget.
2. The hub switches the DAG to `SYNCHRONOUS_COMMIT` on both AGs
   (`ALTER AVAILABILITY GROUP [dag] MODIFY AVAILABILITY GROUP ON ... AVAILABILITY_MODE = SYNCHRONOUS_COMMIT`).
3. The hub waits until the target AG's `last_hardened_lsn` equals the active AG's (LSN
   equality).
4. On the old primary AG it runs `ALTER AVAILABILITY GROUP [dag] SET (ROLE = SECONDARY)`, on
   the new primary AG it runs a graceful `FORCE_FAILOVER_ALLOW_DATA_LOSS` (safe once synced),
   flips `self.role` on both AGs, and hands off the Lease. The annotation is cleared
   automatically.

Because the DAG is synchronous and the target reaches LSN equality before the handoff, a
planned switchover loses no committed rows.

## Scale a data center

Each DC has its own intra-DC AG, so a single `spec.replicas` cannot describe a scale. Scale
a specific DC with a `MSSQLServerOpsRequest` that lists per-DC targets:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: mssql-dcdr-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mssql-dcdr
  horizontalScaling:
    dataCenters:
    - clusterName: dc-west
      replicas: 5
```

Each entry sets that data center's local AG node count; DCs not listed are unchanged. The
request resizes only `dc-west`'s AG, then updates the `PlacementPolicy` so the declarative
topology matches. No other DC's AG and no cross-DC writability is touched.

## Day-2 operations

The standard `MSSQLServerOpsRequest` operations work on a DC-DR cluster and act on every
per-DC AG: vertical scaling, volume expansion (online and offline), version update, and
storage migration each apply to all per-DC `PetSet`s. You issue them exactly as for a
non-distributed MSSQLServer.

## Cleanup

```bash
$ kubectl delete mssqlserver -n demo mssql-dcdr
$ kubectl delete placementpolicy mssql-dcdr
```

Deleting the `MSSQLServer` removes the per-DC `PetSet`s, governing `Service`s, and the
cluster-scoped per-DC `PlacementPolicies` the operator generated. The user-provided base
`PlacementPolicy` is left for you to delete.
