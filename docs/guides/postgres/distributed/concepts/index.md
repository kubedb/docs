---
title: Postgres DC-DR Concepts
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-concepts
    name: Concepts
    parent: guides-postgres-distributed
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Postgres Cross Data Center Disaster Recovery (DC-DR)

KubeDB can run one PostgreSQL database across two data centers (DCs) and move it between them, either on request (planned switchover) or automatically when the active data center is lost (unplanned failover). This feature is called **DC-DR**.

This page explains how DC-DR is built, what each component does, how split-brain is prevented, what to expect for data loss (RPO) and recovery time (RTO), and the current limitations. The other pages in this section walk through a complete installation:

1. [Concepts](/docs/guides/postgres/distributed/concepts/index.md) (this page)
2. [Build the stretched hub](/docs/guides/postgres/distributed/stretched-hub/index.md)
3. [Set up the spoke clusters](/docs/guides/postgres/distributed/spoke-clusters/index.md)
4. [Install KubeDB and dr-controlplane](/docs/guides/postgres/distributed/install/index.md)
5. [Provision a DC-DR Postgres](/docs/guides/postgres/distributed/provision/index.md)
6. [Switchover and failover](/docs/guides/postgres/distributed/switchover-failover/index.md)
7. [Horizontal scaling](/docs/guides/postgres/distributed/scaling/index.md)
8. [Teardown](/docs/guides/postgres/distributed/teardown/index.md)

## Topology

```
          DC A                         DC B                       DC W
   +----------------+           +----------------+          +------------+
   | hub-a          |           | hub-b          |          | etcd-w     |
   |  k8s control   |           |  k8s control   |          |  etcd only |
   |  plane + etcd  |<--------->|  plane + etcd  |<-------->|  (tie-     |
   |  hub workloads |   etcd    |  hub workloads |   etcd   |  breaker)  |
   +----------------+           +----------------+          +------------+
           |  one Kubernetes cluster ("hub") over hub-a + hub-b
           |  OCM hub, KubeDB, dr-controlplane, KubeSlice controller
           |
   +----------------+           +----------------+
   | spoke-1        |  slice    | spoke-2        |
   |  Postgres x2   |<--------->|  Postgres x2   |
   +----------------+ WireGuard +----------------+
```

- **Stretched hub.** One Kubernetes cluster, the *hub*, whose control-plane nodes sit in DC A and DC B. Its etcd has one member in each of three sites (DC A, DC B and a third site, DC W, that runs etcd only). Losing any one site keeps etcd quorum, so the hub keeps working. The hub runs the management components (OCM hub, KubeDB, dr-controlplane, KubeSlice controller) and holds the failover authority (see [The failover authority](#the-failover-authority-the-primary-dc-lease)). It runs no database pods.
- **Spokes.** One Kubernetes cluster per data center (`spoke-1` in DC A, `spoke-2` in DC B). The database pods run here. Each spoke is registered with the hub as an [Open Cluster Management (OCM)](https://open-cluster-management.io/) *ManagedCluster*.
- **Arbiter data center.** A DC-DR Postgres needs two data-bearing data centers plus a third, data-less one that only breaks ties. In this guide the hub itself is registered as a ManagedCluster named `hub` and used as the **Arbiter** data center. It holds no data and can never become primary. The real third-site vote is the etcd member in DC W.
- **Pod delivery.** The PetSet controller on the hub places every database pod (and its Secrets, ConfigMaps and RBAC) on a spoke through OCM **ManifestWorks**, following a **PlacementPolicy**.
- **Networking.** [KubeSlice](https://kubeslice.io/) connects the pod networks of the hub and the spokes with a WireGuard *slice*. Services exported on the slice are resolvable from every cluster as `<name>.<namespace>.svc.slice.local`. Replication between data centers, and the hub operator's connections to the databases, use this network. Applications connect to `<db>.<namespace>.svc.slice.local`, which always points to the current primary.

A data center is identified everywhere by **one name: its OCM ManagedCluster name** (for example `spoke-1`). The same string is the `clusterName` in the PlacementPolicy, the holder of the failover Lease, the `--dc-name` of that data center's agent, and the value of the pod label `open-cluster-management.io/cluster-name`. Choose the names before you start, and keep them identical everywhere.

## How a DC-DR Postgres is laid out

A Postgres becomes a DC-DR database when it has `spec.distributed: true`, the annotation `dr.kubedb.com/enabled: "true"`, and a PlacementPolicy with a `failoverPolicy`. KubeDB expands it on the hub into:

- one PetSet per data-bearing data center, named `<db>-<dc>` (for example `pg-docs-spoke-1`), with its own PlacementPolicy `<policy>-<dc>`, derived from the PlacementPolicy `<policy>` that you create;
- an arbiter PetSet `<db>-<dc>-arbiter` in a data center whose replica count is even, so that the data center's own Raft group always has an odd number of voters.

Each data center runs its **own Raft group** in the `pg-coordinator` sidecar of its pods. Raft never stretches across data centers:

- In the **active** data center, the Raft leader is the only writable primary. It carries the pod label `kubedb.com/role: primary`.
- In the **standby** data center, the Raft leader is a read-only *standby leader*. It streams WAL across data centers from the active primary through a replication slot named after its data center (for `spoke-1`, `dcdr_spoke1`), and the other pods of that data center cascade from it.
- If the primary pod fails, its own data center's Raft elects a new local primary within seconds. The active data center does not change.

## The failover authority: the primary-DC Lease

Which data center is active is decided by one Kubernetes **Lease** on the hub, in the namespace `dc-failover`:

- `primary-dc` for databases that do not name a failover group, or
- `primary-dc-<group>` for databases whose PlacementPolicy sets `failoverPolicy.failoverGroupRef.name: <group>`.

The Lease's `holderIdentity` is the active data center. The Lease is stored in the hub's three-site etcd, so a partitioned minority of sites cannot change it. All databases that share a Lease (a *failover scope*) always fail over together. Give every database its own `failoverGroupRef` unless you want them to move together; KubeDB reports a shared scope through the condition `DCDRFailoverScopeShared` on the Postgres object.

The Lease is maintained by the **dr-controlplane** agents, one per data center (the Arbiter data center included):

- Each agent renews a health Lease `dc-health-<dc>` (15 s duration, renewed every 5 s).
- Each agent of a Member data center contends for the scope's `primary-dc*` Lease with a 45 s lease duration, a 40 s renew deadline and a 5 s retry period. Arbiter and Witness data centers never contend.
- Each agent copies the current holder into a local **marker** ConfigMap, `dc-failover/primary-dc` (or `primary-dc-<group>`), on its own cluster. The marker carries `activeDC`, `renewTime` and, during a planned switchover, `quiesce`.

The database pods never look across data centers to decide whether they may write. `pg-coordinator` reads only the local marker, and it **fails closed**: if the marker is missing, unreadable, names another data center, or is older than 30 s, the data center is treated as not active.

## Components

| Component | Runs on | Role |
|---|---|---|
| etcd (three sites) | hub-a, hub-b, etcd-w | Stores the hub's API data, including the failover Leases. |
| OCM (cluster-manager, klusterlet, work agent, addon framework) | hub and every cluster | Registers clusters, delivers database pods and objects with ManifestWorks, installs add-ons. |
| KubeSlice (controller on the hub; worker, gateway and slice DNS in each cluster) | hub and every cluster | Cross-cluster pod network and `*.slice.local` DNS. |
| PetSet controller | hub (part of the KubeDB chart) | Places pods on spokes per the PlacementPolicy; ships the PetSet, PlacementPolicy and FailoverGroup APIs. |
| KubeDB provisioner | hub | Creates the per-DC PetSets, arbiters and RBAC, and runs the **DC-DR orchestrator**: status, planned switchover, promotion after a failover. |
| KubeDB ops-manager and webhook server | hub | PostgresOpsRequests and validation. |
| `pg-coordinator` | sidecar in every Postgres pod | Intra-DC Raft and failover, fencing, write-lock, rejoin (`pg_rewind`). |
| dr-controlplane topology controller | hub | Creates one primary-DC Lease per failover scope from the PlacementPolicies; runs the FailoverGroup controller. |
| dr-controlplane addon manager | hub | Installs the agent on every labelled ManagedCluster as an OCM add-on, with a hub credential. |
| dr-controlplane agent | every cluster, the hub included | Renews `dc-health-<dc>`, contends for `primary-dc*`, writes the local marker ConfigMaps. |

## Planned switchover

A planned switchover moves the active data center with **zero acknowledged-write loss**. You request it with the annotation `dr.kubedb.com/switchover-to=<dc>` on the Postgres object. The orchestrator then:

1. asks the active data center to quiesce; `pg-coordinator` write-locks the active primary (`default_transaction_read_only`) and ends write transactions;
2. waits until the target data center has replayed to within 8 KiB of the frozen primary position;
3. hands the primary-DC Lease to the target data center;
4. promotes the target data center's leader and rewinds the old primary into a standby.

A switchover that cannot finish is aborted automatically after 10 minutes, and you can abort it at any time with `dr.kubedb.com/switchover-abort`. Writes then resume on the original data center.

## Unplanned failover

When the active data center is lost:

1. Its agent stops renewing the primary-DC Lease.
2. After the 45 s lease duration, the other Member data center's agent acquires the Lease.
3. The markers flip, and the orchestrator on the hub promotes the surviving data center's cross-DC streamer. It may do this through a `ForceFailOver` PostgresOpsRequest named `<db>-dcdr-failover-<suffix>`.
4. When the lost data center comes back, its pods rejoin as standbys of the new active data center. An old primary whose history diverged is brought back with `pg_rewind`.

## Split-brain protection

DC-DR guarantees that two data centers are never writable at the same time. Several mechanisms work together.

**Fence before promote.** If the active data center is cut off from the hub but keeps running, its marker stops being refreshed:

- 30 s after the last Lease renewal it observed, the marker is stale and the data center is no longer considered active;
- 10 s later, `pg-coordinator` write-locks the primary and terminates open write transactions;
- the Lease cannot be acquired by the other data center until its 45 s duration has passed, plus up to one 5 s retry.

So the isolated primary stops accepting writes before the other data center can be promoted. If the uncertainty lasts longer (60 s for a stale marker, 90 s for a missing one), `pg-coordinator` also removes the primary label from the pod.

**Raft inside each data center.** Within each data center, `pg-coordinator` runs Raft with the Pre-Vote phase enabled. A member that wants to start an election first runs a pre-vote round without increasing its term, and a real election follows only if a majority of its peers would vote for it. A pod that was cut off and has fallen behind cannot win that round, so when it reconnects it does not disrupt the healthy cluster with a needless election. Raft's CheckQuorum option is not enabled. Protection against a stale leader across data centers comes from a lease instead: only the data center that holds the time-bound `primary-dc` Lease may be writable, and a primary that can no longer confirm that Lease write-locks itself before the Lease can expire and be acquired by another data center.

**Only Members can be primary.** Only data centers with `role: Member` in the PlacementPolicy contend for the Lease. An Arbiter never holds data and never becomes primary.

**Nothing at the database level moves the Lease.** A crashed Postgres, replication lag or failed health checks do not move the Lease; only a data center that stops renewing (or a planned switchover or handoff) does. A failed primary inside a healthy data center is replaced by that data center's own Raft.

**A third site breaks ties.** The Lease lives in the three-site etcd. During a network partition, only the side that still has an etcd majority can renew or acquire it. The etcd members use Pre-Vote (`--pre-vote=true`): before disrupting the cluster with a real election, a member asks the others whether they can still see the leader, and only an unopposed candidate starts the election. Raft prevents a stale or isolated leader with CheckQuorum, which the etcd server always enables: a leader that loses the majority steps down, and members ignore vote requests while their leader lease is still valid. See [etcd and Raft settings](/docs/guides/postgres/distributed/stretched-hub/index.md#step-3-run-etcd-on-the-three-sites).

## Data loss (RPO) and recovery time (RTO)

Cross data center replication is **asynchronous**. The numbers below were measured on a test installation of this guide's topology (KubeDB `v2026.10.12`, three single-site VMs with sub-millisecond latency between them) with a writer that records every acknowledged write.

| Scenario | Data loss | Recovery |
|---|---|---|
| Planned switchover | None (0 acknowledged writes lost in every run) | `Steady` on the new data center in 11-15 s; writes paused for about 15 s |
| Primary pod failure inside the active data center | None | New local primary writable in 4-6 s; the active data center does not change |
| Loss of the active data center (VM reboot) | Writes not yet streamed to the other data center (0 lost in the test) | Other data center writable after about 80 s; `protected` again after about 3.5 minutes |
| Active data center isolated by a network cut, still running | Writes acknowledged **inside the isolated data center** between the cut and the write-lock, about 40 s (184 rows in the test). Writes made through the other data center are not affected | Write-lock about 40 s after the cut; the other data center writable about 67 s after the lock |

Clients connect through `<db>.<namespace>.svc.slice.local`. After every primary move, this name takes about 30 s to resolve to the new primary, even when the database itself is writable sooner. Configure your applications to retry connections.

You can bound the loss of an unplanned failover with `spec.replication.bestEffortCrossDCLagBytesForFailover`; see [Switchover and failover](/docs/guides/postgres/distributed/switchover-failover/index.md#bound-the-data-loss-of-an-unplanned-failover).

## Status

The Postgres object reports the DC-DR state in `status.disasterRecovery`:

| Field | Meaning |
|---|---|
| `activeDC` | The data center that holds the primary-DC Lease. |
| `phase` | `Steady`, `FailingOver`, `FailingBack` or `Degraded`. `Steady` requires confirmed protection. |
| `protected` | `true` when another Member data center is confirmed streaming within the lag budget. Treat a missing value as `false` before any planned operation. |
| `protectionMessage` | Why `protected` has its current value. |
| `dataCenters[]` | One entry per data center: `clusterName`, `role`, `leader`, `writable`, `crossDCStreamer`, `lagBytes`, `healthy` (the freshness of its `dc-health-<dc>` Lease), `notReadyPods`. |

## Limitations

- **Two data-bearing data centers.** This guide covers `failoverPolicy.mode: TwoDC`: exactly two Member data centers and one Arbiter.
- **Day-2 operations.** Of the PostgresOpsRequest types, only per data center horizontal scaling is documented for DC-DR databases in this release. See [Horizontal scaling](/docs/guides/postgres/distributed/scaling/index.md).
- **Hub components are single-replica.** Restarting any of them does not move the primary or lose writes, but the loss of a hub VM that runs them was not tested.
- **OCM uses one hub API address.** The klusterlets register against a single hub API server URL. Use a VIP, a load balancer or a DNS name for production so that the loss of one hub node does not cut the klusterlets off.
- **Switchover target.** A switchover to a data center that is not a Member (for example the Arbiter) is not rejected up front; it stays `FailingOver` until you abort it.
- **Leftover group Leases.** When you delete the last database of a failover group, its `primary-dc-<group>` Lease and marker ConfigMaps stay behind. See [Teardown](/docs/guides/postgres/distributed/teardown/index.md).
- **Backup.** Backup and restore of a DC-DR database with KubeStash is not covered by this guide.
- **Monitoring is not supported.** Do not set `spec.monitor` on a DC-DR Postgres in `v2026.10.12`. The provisioner in this release has no RBAC for `discovery.k8s.io` EndpointSlices, which it needs to build the stats endpoints of a DC-DR database, so a DC-DR Postgres with `spec.monitor` set stays in `Provisioning`. The stats endpoints are also derived from non-DC-DR pod names (`<db>-<ordinal>` instead of `<db>-<dc>-<ordinal>`), so they would not point at the DC-DR pods even with that permission.

## Next Steps

- [Build the stretched hub](/docs/guides/postgres/distributed/stretched-hub/index.md).
