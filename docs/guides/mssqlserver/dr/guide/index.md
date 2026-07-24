---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mssqlserver-dr-guide
    name: User Guide
    parent: guides-mssqlserver-dr
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running MSSQLServer in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed MSSQLServer in cross data center
disaster recovery (DC-DR) mode: the components, the naming contract, deployment, connecting,
monitoring, replication and lag, timing and tuning, quorum and roles, switchover and
failback, scaling, day-2 operations, backup, and deletion.

Read the [DC-DR Overview](/docs/guides/mssqlserver/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/mssqlserver/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the active (writable) DC. This is the single cross-DC failover authority. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease on behalf of its DC and projects the Lease decision into the local spoke as a marker `ConfigMap`. |
| **KubeDB MSSQLServer operator (hub)** | the OCM hub | Expands the `MSSQLServer` CR into per-DC AGs, wires the DAG between them, watches the Lease, drives the DAG failover and switchover, and writes `status.disasterRecovery`. |
| **`mssql-coordinator`** | every SQL Server pod | Runs the intra-AG raft, manages AG membership and the `kubedb.com/role` label, reads the local marker, and forces its AG to the DAG `SECONDARY` role when its DC is not active. |
| **KubeSlice** | each spoke | Provides the cross-DC overlay so the DAG can connect the two AGs over the DB-mirroring endpoint on port 5022. |

The `mssql-coordinator` reuses the pg-coordinator HTTP/raft wire protocol (client port 2379,
peer 2380, `/current-primary`, `/transfer`, `/add-node`, `/remove-node`), so the Postgres
fence and orchestrator patterns apply directly.

The marker `ConfigMap` is the contract between the agent (producer) and the coordinator
(consumer):

```
ConfigMap primary-dc  (namespace: dc-failover, on each spoke)
  data.activeDC  = the DC the quorum currently trusts as primary
  data.renewTime = RFC3339, the observed primary-DC Lease renewTime
```

The coordinator trusts the marker for 30s (the fence TTL); absent, stale, unparseable, or
naming another DC all mean *not active* and the AG is held as the DAG secondary. This is the
fail-closed fence.

## The DC-name contract

One string identifies a data center everywhere. **Keep these identical:**

- the OCM spoke cluster name
- the agent `--dc-name`
- the primary-DC Lease `holderIdentity`
- the marker `data.activeDC`
- the pod label `open-cluster-management.io/cluster-name`
- the `PlacementPolicy` `distributionRule.clusterName`

## Operator configuration

Start the MSSQLServer operator with:

```
--dc-dr-enabled
--dc-dr-coord-kubeconfig=<kubeconfig of the coordination control plane>
--dc-dr-local-dc=<the data center this operator instance runs in>
```

The per-DC pod coordinators automatically receive `DC_DR_ENABLED`, `DC_NAME`,
`DC_DR_NAMESPACE` (default `dc-failover`), and `DC_DR_MARKER` (default `primary-dc`) through
their PetSet template, so the fence works without extra wiring. The cross-DC DAG endpoints
are derived by the operator from the KubeSlice-exported 5022 Services, not from
hand-supplied LoadBalancer URLs.

## Deploying

### PlacementPolicy

Map the global pod ordinals to data centers and tag each DC with its role:

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
        scope: Global       # one cluster-wide failover scope (or Group + a group name)
      mode: TwoDC           # exactly two Member DCs plus an Arbiter DC
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
- A native Distributed AG joins **exactly two** AGs, so `mode: TwoDC` expects exactly two
  Member DCs plus the Arbiter DC. Three or more data DCs need chained DAGs and are out of
  scope.
- Give each Member DC an **odd** local node count so its AG keeps a clean coordinator-raft
  majority; an even AG gets an auto-injected local Arbiter PetSet.

### MSSQLServer

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssql-dcdr
  namespace: demo
  annotations:
    dr.kubedb.com/enabled: "true"          # opt into per-DC DC-DR expansion
    # dr.kubedb.com/failover-group: payments  # optional: a Group failover scope
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

### What the operator creates

Per data-bearing DC `<dc>`:

- a per-DC AG with its own `mssql-coordinator` raft and `SYNCHRONOUS_COMMIT` replicas, backed
  by a `PetSet` `<db>-<dc>` (for example `mssql-dcdr-dc-east`);
- a DC-local headless governing `Service`, exported over KubeSlice (it carries port 5022), so
  the DC's pods discover only each other and the DAG can reach the AG listener;
- a cluster-scoped per-DC `PlacementPolicy` `<base>-<dc>` pinning that AG to the DC;
- the native DAG joining the two AGs over their 5022 endpoints, with the active DC's AG as
  the DAG primary and the standby DC's AG auto-seeding as the DAG secondary (forwarder).

The Arbiter DC (`role: Arbiter`) runs no SQL Server pods. All per-DC pods carry the offshoot
selectors plus the `open-cluster-management.io/cluster-name` label, so the global
primary/standby Services and the single `AppBinding` keep working.

## Connecting

A DC-DR MSSQLServer exposes the same single endpoint as any KubeDB MSSQLServer:

- the **primary Service** `<db>` resolves to the active DC's writable AG primary (only that
  AG primary is labeled `kubedb.com/role: primary`);
- the **secondary Service** `<db>-secondary` resolves to the read-only replicas;
- one **`AppBinding`** `<db>` for applications and KubeDB integrations.

Because only the active DC's AG primary carries the `primary` label, the endpoint follows
failover automatically. Applications keep using `<db>` and reconnect after a failover,
landing on the new active DC.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC that holds the Lease and runs the writable DAG primary AG. |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | The DC role: `Member` or `Arbiter`. |
| `dataCenters[].agPrimary` | That DC's local AG primary pod. |
| `dataCenters[].dagRole` | The AG's DAG role: `Primary` or `Secondary`. |
| `dataCenters[].writable` | True only for the active DC. |
| `dataCenters[].synchronizationHealth` | The DAG sync health (`HEALTHY`, `PARTIALLY_HEALTHY`, `NOT_HEALTHY`). |
| `dataCenters[].redoQueueBytes` | The DAG forwarder's redo queue size (un-applied redo). |
| `dataCenters[].logSendQueueBytes` | The DAG send queue size (un-shipped log). |
| `dataCenters[].lastHardenedLSN` | The DC's `last_hardened_lsn`, used for the LSN equality gate. |
| `dataCenters[].healthy` | Whether the DC's AG is healthy. |

### Useful checks

```bash
# Which DC is active (from the coordination plane):
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# The marker each spoke reads (run against a spoke):
$ kubectl -n dc-failover get configmap primary-dc -o yaml

# Per-DC AG primaries and roles:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=mssql-dcdr \
    -L kubedb.com/role,open-cluster-management.io/cluster-name

# On any AG replica, the DAG sync state, redo and send queues, and LSN:
#   SELECT synchronization_health_desc, redo_queue_size, log_send_queue_size, last_hardened_lsn
#   FROM sys.dm_hadr_database_replica_states;
```

## Replication, lag, and RPO

- Cross-DC replication is the native **Distributed Availability Group**, an
  `ASYNCHRONOUS_COMMIT` link over the DB-mirroring endpoint on port 5022. The standby DC's AG
  is the DAG secondary (forwarder); it applies the DAG stream synchronously within its own
  AG.
- Lag is read from `sys.dm_hadr_database_replica_states` on the DAG forwarder:
  `synchronization_health`, `redo_queue_size` (redo not yet applied), `log_send_queue_size`
  (log not yet shipped), and `last_hardened_lsn` versus the primary AG's. The DC's
  coordinator computes these and surfaces them (the hub never opens cross-cluster SQL). They
  are the basis for the RPO of an unplanned failover.
- A **planned switchover loses no committed rows** because the DAG is switched to
  `SYNCHRONOUS_COMMIT` and the target reaches `last_hardened_lsn` equality before the
  handoff. An **unplanned failover** may lose the last un-shipped redo (bounded by the
  standby's lag at the moment the active DC died).

## Timing and tuning (RTO vs safety)

DC-DR has one timing invariant that must hold for correctness:

> **fence TTL + cross-DC clock skew < primary-DC Lease duration**

The marker `renewTime` tracks the Lease's renewTime. A partitioned active DC self-fences (it
forces its AG to DAG secondary) at `lastRenew + fence TTL`; a survivor can only acquire the
expired Lease at `lastRenew + LeaseDuration`. Keeping the fence TTL inside the Lease duration
guarantees the old active DC goes read-only **before** any new DC becomes writable, with no
split-brain window.

Default values:

| Parameter | Where | Default |
| --- | --- | --- |
| Fence TTL | mssql-coordinator | 30s |
| Marker refresh interval | dr-controlplane agent | 5s |
| Primary-DC Lease duration | dr-controlplane agent (`--election-lease-duration`) | 45s |
| Lease renew deadline | dr-controlplane agent (`--election-renew-deadline`) | 30s |
| Lease retry period | dr-controlplane agent (`--election-retry-period`) | 2s |

The failover **RTO floor** is roughly the Lease duration (the time a survivor waits to
acquire). To lower RTO, lower the Lease duration **and** the fence TTL together, always
preserving `fence TTL + skew < LeaseDuration`. The retry period must stay well under the
fence TTL so the holder restamps `renewTime` and the marker reads fresh in normal operation.

## Quorum, roles, and arbiters

- Each DC's AG needs its own coordinator-raft majority. Give each Member DC an **odd** local
  node count so an intra-DC primary failure keeps quorum. An even AG node count gets an
  auto-injected voting **Arbiter** PetSet (`SetArbiterDefault`, label
  `kubedb.com/role=arbiter`), evaluated per DC group.
- The Arbiter **DC** (`role: Arbiter`, empty `replicaIndices`) is different: it holds only
  the `dr-controlplane` etcd vote, never SQL Server data. Do not confuse the per-AG arbiter
  PetSet (intra-DC even-node quorum) with the Arbiter DC (the cross-DC tie-breaker).
- The fence forces a non-active DC's AG to the DAG `SECONDARY` role and keeps its AG primary
  labeled `standby`, so the single primary Service and AppBinding never resolve to it. An
  intra standby-DC AG election (the coordinator electing a new AG primary) is transparent to
  the DAG, which follows the listener.

Separately from a DC's *intra-DC* quorum, the **cross-DC** failover quorum needs a majority
of three voting sites. For how to lay this out across two or three data centers (and why a
third Arbiter site is preferred), see
[Deployment topologies](/docs/guides/mssqlserver/dr/overview/index.md#deployment-topologies-2-dcs-vs-3-dcs).

## Planned switchover (zero-RPO)

Move the active DC on purpose by annotating the MSSQLServer (there is no Switchover
OpsRequest type; the engine-aware quiesce and catch-up run in the hub):

```bash
$ kubectl annotate mssqlserver -n demo mssql-dcdr dr.kubedb.com/switchover-to=dc-west
```

The hub then:

1. checks the target is a known, healthy DC within the lag budget;
2. sets `phase: FailingOver` and switches the DAG to `SYNCHRONOUS_COMMIT` on both AGs
   (`ALTER AVAILABILITY GROUP [dag] MODIFY AVAILABILITY GROUP ON ... AVAILABILITY_MODE = SYNCHRONOUS_COMMIT`);
3. waits until the target AG's `last_hardened_lsn` equals the active AG's;
4. runs `SET (ROLE = SECONDARY)` on the old primary AG, a graceful
   `FORCE_FAILOVER_ALLOW_DATA_LOSS` (safe once synced) on the new primary AG, flips
   `self.role` on both AGs, and hands the Lease to the target.

The annotation is cleared automatically once the target is active. Watch
`status.disasterRecovery` for `phase` returning to `Steady` with the new `activeDC`.

## Failback

Failback is a switchover back to the original DC once it is healthy again:

```bash
$ kubectl annotate mssqlserver -n demo mssql-dcdr dr.kubedb.com/switchover-to=dc-east
```

A returned old-primary AG rejoins the DAG as the secondary (forwarder). If its databases did
not diverge it resumes the DAG stream directly. If it accepted writes that were never shipped
(a forked tail after an unplanned `FORCE_FAILOVER_ALLOW_DATA_LOSS`), SQL Server cannot stream
over the diverged databases, so the operator removes them from the AG and lets DAG automatic
seeding re-seed them over 5022 (the SQL Server analog of a re-seed, not a transparent
rejoin). After catch-up, the zero-RPO switchover above returns the active DC.

## Per-DC horizontal scaling

Each DC has its own AG, so scale a specific DC with a `MSSQLServerOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: mssql-dcdr-scale-west
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

- Each entry sets that DC's local AG node count; DCs not listed are unchanged.
- Nodes are added or removed over the DC-local network; new replicas join the AG with
  `SEEDING_MODE = AUTOMATIC` and recover from the local AG primary.
- The base `PlacementPolicy` is renumbered so the declarative topology matches.
- Keep each Member DC's count **odd** for a clean coordinator-raft majority. Removing a whole
  DC is a topology change, not horizontal scaling.

## Day-2 operations

The standard `MSSQLServerOpsRequest` operations apply to every per-DC AG on a DC-DR cluster;
issue them exactly as for a non-distributed MSSQLServer:

| Operation | DC-DR behavior |
| --- | --- |
| **Vertical scaling** | Patches every per-DC PetSet, restarts per-DC pods. |
| **Volume expansion** (online/offline) | Expands every per-DC data PVC, waits on all per-DC PetSets. |
| **Version update** | Updates every per-DC PetSet. |
| **Storage migration** | Orphan-deletes and waits on every per-DC PetSet. |
| **Reconfigure / Restart / Rotate-Auth / Reconfigure-TLS** | Apply across the per-DC pods. |

## Backup

Back up a DC-DR MSSQLServer the same way as any KubeDB MSSQLServer (KubeStash). Backups run
against the writable endpoint, so they read from the active DC; the AppBinding follows
failover, so a scheduled backup continues against the new active DC after a failover.
Point-in-time recovery works as usual.

## DC-aware health

On a DC-DR cluster the operator's health check is DC-aware:

- the **active** DC is expected to have a writable AG primary (the write-check runs there);
- a **standby** DC is expected to be a DAG secondary (forwarder) with a healthy inbound DAG
  (`synchronization_health = HEALTHY`), so the passive DC does not flap to `NotReady`;
- the ONLINE/CONNECTED replica count is checked against the **per-DC** AG size, not the
  global `spec.replicas`.

## Deletion and cleanup

```bash
$ kubectl delete mssqlserver -n demo mssql-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC PetSets, governing Services, and the
cluster-scoped per-DC `PlacementPolicies` it generated (these carry no owner reference, so
the operator deletes them explicitly). The user-provided base `PlacementPolicy` is left for
you to delete.

## Limitations

- **A native DAG joins exactly two AGs**, so DC-DR for SQL Server is a two-data-DC design.
  Three or more data DCs need chained DAGs and are out of scope.
- **Adding or removing a whole data center** is a topology change (a new AG and a DAG
  re-wire), distinct from horizontal scaling, and is performed by editing the
  `PlacementPolicy` topology, not a `HorizontalScaling` request.
- Cross-DC replication is asynchronous; an unplanned failover has a non-zero RPO bounded by
  the standby lag (un-shipped redo). Use a **planned switchover** for zero-RPO moves.
- A standby DC that forked after an unplanned `FORCE_FAILOVER_ALLOW_DATA_LOSS` must have its
  diverged databases removed and re-seeded over the DAG on failback rather than streaming
  directly.
- All correctness depends on the timing invariant above; do not set a fence TTL that meets or
  exceeds the Lease duration.
