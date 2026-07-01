---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: cas-dr-guide-cassandra
    name: User Guide
    parent: cas-dr-cassandra
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running Cassandra in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed Cassandra in cross data center
disaster recovery (DC-DR) mode: the components, the naming contract, deployment,
connecting through the single write endpoint, reading and writing locally, consistency
levels, monitoring, hint and repair backlog as the lag proxy, switchover and failback,
scaling, and day-2 operations.

Read the [DC-DR Overview](/docs/guides/cassandra/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/cassandra/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the DC the single write endpoint resolves to. The Lease is routing, policy, and observability, **not** a failover or fence mechanism. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease for its DC and projects the Lease decision into the local spoke as the `primary-dc` marker. |
| **KubeDB Cassandra operator (hub)** | the OCM hub | Expands the `Cassandra` CR into one Cassandra datacenter per Member DC, wires `NetworkTopologyStrategy`, cross-DC seeds, and the snitch, routes the single write endpoint by following the Lease, drives planned switchover, and writes `status.disasterRecovery`. |
| **Cassandra ring** | every Member DC (a full Cassandra datacenter) | The masterless Dynamo-style ring. Gossip carries membership across DCs; `NetworkTopologyStrategy` replicates continuously. There is no leader and no cross-DC election. |
| **KubeSlice** | each spoke | Provides the cross-DC pod network so the one ring spans clusters: gossip and the storage port (7000/7001) reach across DCs and CQL (9042) is reachable for the endpoint. |

## The DC-name contract

One string identifies a data center everywhere. **Keep these identical:**

- the OCM spoke cluster name
- the agent `--dc-name`
- the primary-DC Lease `holderIdentity`
- the marker `data.activeDC`
- the Cassandra `dc=` in each pod's `cassandra-rackdc.properties`
- the pod label `open-cluster-management.io/cluster-name`
- the `PlacementPolicy` `distributionRule.clusterName`

## Deploying

### PlacementPolicy

Map the global replica indices to data centers and tag each DC with its role:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  name: cas-dcdr
spec:
  clusterSpreadConstraint:
    slice:
      projectNamespace: kubeslice-demo
      sliceName: demo-slice
    failoverPolicy:
      trigger:
        scope: Global       # one cluster-wide failover scope (or Group + a group name)
      mode: ThreeDC         # ThreeDC: 3 Member DCs, no arbiter; TwoDC: 2 Member DCs + an engine-free Arbiter DC
    distributionRules:
    - clusterName: dc-a
      role: Member
      replicaIndices: [0, 1, 2]
    - clusterName: dc-b
      role: Member
      replicaIndices: [3, 4, 5]
    - clusterName: dc-c
      role: Member
      replicaIndices: [6, 7, 8]
```

- Each **Member** rule maps to one Cassandra datacenter; its `replicaIndices` are the
  nodes (racks) of that datacenter, replicated with `NetworkTopologyStrategy`.
- `mode: ThreeDC` expects an odd number of Member DCs and **no** separate Arbiter DC (the
  preferred layout, and Cassandra's own recommended geo shape). `mode: TwoDC` expects two
  Member DCs plus an engine-free Arbiter DC that runs only the `dr-controlplane` etcd
  member (no Cassandra) so the coordination quorum keeps an odd site count.
- Roles are `Member` and `Arbiter` only. An `Arbiter` rule carries an empty
  `replicaIndices` and no Cassandra is scheduled onto it.

### Cassandra

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cas-dcdr
  namespace: demo
spec:
  version: "5.0.3"
  distributed: true
  topology:
    rack:
    - name: rack-a
      replicas: 3
      storage:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 1Gi
  storageType: Durable
  podTemplate:
    spec:
      podPlacementPolicy:
        name: cas-dcdr
  deletionPolicy: WipeOut
```

### What the operator creates

- **One logical ring** whose members are spread across the Member DCs, each Member DC a
  full Cassandra datacenter with its own racks. Every pod is configured with
  `GossipingPropertyFileSnitch` and a per-pod `cassandra-rackdc.properties`
  (`dc=<clusterName>`, `rack=<rack>`), and the seed list includes a few seeds per remote
  DC so gossip forms one ring across the DCs.
- **`NetworkTopologyStrategy` replication** with a replication factor per DC. User
  keyspaces must use `NetworkTopologyStrategy`; a keyspace on `SimpleStrategy` is **not**
  DC-DR safe (it ignores datacenters). See "Keyspaces and replication" below.
- A single **write endpoint** (Service plus `AppBinding`) that the orchestrator points at
  the active DC by following the Lease. Reads and writes at `LOCAL_QUORUM` can also go to
  any DC directly.

All data-bearing pods carry the offshoot selectors plus the
`open-cluster-management.io/cluster-name` label, so the single write endpoint and the
single `AppBinding` keep working as the active DC moves.

> The snitch (`GossipingPropertyFileSnitch`) and the per-pod `dc=`/`rack=` values are the
> topology contract. Do not change a pod's `dc=` after it has joined the ring; that is a
> datacenter move, not a config edit.

## Keyspaces and replication

Create every user keyspace with `NetworkTopologyStrategy` and a replication factor per
Member DC. For the three-DC layout above:

```sql
CREATE KEYSPACE app
  WITH replication = {
    'class': 'NetworkTopologyStrategy',
    'dc-a': 3,
    'dc-b': 3,
    'dc-c': 3
  };
```

- `NetworkTopologyStrategy` places replicas per datacenter, which is what makes the ring
  DC-DR safe: each DC holds a full copy and can serve `LOCAL_QUORUM` on its own.
- A keyspace using `SimpleStrategy` is not datacenter-aware and is **not** DC-DR safe.
  Convert it with `ALTER KEYSPACE ... WITH replication = {'class':
  'NetworkTopologyStrategy', ...}` and then run `nodetool repair` so the new replicas are
  populated.
- The system keyspaces `system_auth`, `system_distributed`, and `system_traces` should
  also use `NetworkTopologyStrategy` across the DCs so authentication and system state
  survive a DC loss.

## Connecting

A DC-DR Cassandra exposes a single write endpoint, the same shape as any KubeDB
Cassandra:

- the **write endpoint** `<db>` resolves to the active DC's nodes (CQL port 9042); the
  Lease-driven routing keeps a stable single-writer posture;
- one **`AppBinding`** `<db>` for applications and KubeDB integrations.

Because Cassandra is masterless, every DC can accept reads and writes, but the single
endpoint gives a stable single-writer posture: applications keep using `<db>` and, after a
failover, reconnect and land on the new active DC. If your driver is datacenter-aware, you
can also point it at a local DC with a `DCAwareRoundRobinPolicy` and `LOCAL_QUORUM` for
lowest latency.

### Writes and consistency (the correctness knob)

There is no fence and no quorum that decides who may commit. Correctness comes from the
**consistency level** you choose per statement or per keyspace:

```sql
-- Local-quorum write against the write endpoint <db>:9042 (the default, DC-loss tolerant):
CONSISTENCY LOCAL_QUORUM;
INSERT INTO app.orders (id, item, qty) VALUES (uuid(), 'widget', 1);
```

- **`LOCAL_QUORUM`** acks when a quorum of replicas in the local DC has the write; it
  keeps working when another DC is down and replicates asynchronously to the other DCs.
  This is the documented default.
- **`EACH_QUORUM`** acks only when a quorum in **every** DC has the write; it is the
  strongest cross-DC durability but fails while any DC is down. Reserve it for keyspaces
  or statements that cannot tolerate the `LOCAL_QUORUM` loss window.
- The bounded loss on an unplanned active-DC loss is only writes that acked at
  `LOCAL_QUORUM` in the lost DC but had not yet replicated to survivors (bounded by hint
  and replication backlog when the DC died, recoverable by repair if the DC returns).

### Read locally

Any DC serves `LOCAL_QUORUM` reads. Point read traffic at an in-DC coordinator for low
latency; reads are eventually consistent across DCs, bounded by hint and repair backlog.
Use `LOCAL_QUORUM` reads plus `LOCAL_QUORUM` writes on a keyspace with RF >= 3 per DC for
read-your-writes within a DC.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get cassandra -n demo cas-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC the single write endpoint currently resolves to (a routing choice, not a promoted primary). |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name (also the Cassandra `dc=`). |
| `dataCenters[].role` | `Member` or `Arbiter`. |
| `dataCenters[].writable` | True only for the active (write-routed) DC (a routing marker, not an engine fence). |
| `dataCenters[].upNormalNodes` | The DC's UN (up/normal) node count from `nodetool status`. |
| `dataCenters[].totalNodes` | The DC's total node count. |
| `dataCenters[].pendingHints` | Cross-DC hinted-handoff backlog for this DC (a lag proxy). |
| `dataCenters[].repairBacklog` | Anti-entropy repair staleness for this DC (a lag proxy). |
| `dataCenters[].healthy` | Whether the DC has its expected UN nodes. |

### Useful checks

```bash
# Which DC the Lease intends as the write-routed active DC:
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# Per-DC nodes and DCs:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=cas-dcdr \
    -L open-cluster-management.io/cluster-name

# Ring and per-DC node status (UN = up/normal), from any node:
$ kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status

# Cross-DC streaming, pending hints, and pending ranges (the lag signal):
$ kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool netstats

# Gossip view of every node across DCs:
$ kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool gossipinfo
```

## Replication, lag, and RPO

- Cross-DC replication is **native Cassandra `NetworkTopologyStrategy`** over the storage
  port (7000, or 7001 with TLS), asynchronous, one copy per write per remote DC (the
  local coordinator forwards to a remote coordinator that fans out to local replicas).
  There is exactly one ring, so there is no extra replication link to manage.
- The lag signals are the **hinted-handoff backlog** (`nodetool netstats`, hints stored
  for a temporarily down node) and **repair staleness** (how long since the last
  cross-DC `nodetool repair`). These are surfaced into `status.disasterRecovery` as
  `pendingHints` and `repairBacklog`.
- A **planned switchover is a routing move**, so it loses no committed data on its own;
  for the strictest handoff, drain hints and run a cross-DC repair first so the target is
  fully converged. An **unplanned failover** may lose only writes that acked at
  `LOCAL_QUORUM` in the lost DC but had not yet replicated (bounded by the hint and
  replication backlog when the DC died). Keyspaces written at `EACH_QUORUM` lose zero on
  the writes that acked, at the cost of not acking while any DC is down.

## Arbiter and DC-count guidance

- **Prefer an odd number of Member DCs.** Three full Cassandra datacenters, each with its
  own replication factor, is the preferred layout and needs **no arbiter**. It is also
  Cassandra's own recommended geo shape.
- **An Arbiter DC appears only for an even data-DC count** (two Member DCs). It is
  **engine-free**: it runs only the `dr-controlplane` etcd member so the coordination
  quorum keeps an odd site count. No Cassandra runs in the Arbiter DC, because
  Cassandra's data plane needs no cross-DC voter (its correctness is per-DC quorum on each
  query).
- This is different from ClickHouse and MongoDB, whose arbiter DC holds a data-less engine
  voter. Cassandra's arbiter holds none.

## Planned switchover (routing move)

Move the active (write-routed) DC on purpose by annotating the Cassandra:

```bash
$ kubectl annotate cassandra -n demo cas-dcdr dr.kubedb.com/switchover-to=dc-b
```

The hub then:

1. checks the target is a known, healthy DC within the hint and repair backlog budget;
2. sets `phase: FailingOver`;
3. moves the Lease and the single write endpoint to `dc-b`.

Watch `status.disasterRecovery` for `phase` returning to `Steady` with the new
`activeDC`. There is no promotion step and no engine catch-up gate, because every DC is
already a full writable datacenter. For a strict zero-loss handoff, drain hints and run a
cross-DC `nodetool repair` toward the target before switching so it is fully converged.

## Failback

Failback is native and clean. A returned DC rejoins the ring by gossip, receives hinted
handoff within the hint window, and a full cross-DC `nodetool repair` (anti-entropy)
reconciles the rest. There is **no rewind**: Cassandra is AP and reconciles by
last-write-wins on cell timestamps, so there is nothing to roll back (unlike the Postgres
`pg_rewind` path).

After the returned DC is caught up (its `pendingHints` drained and a cross-DC repair
complete), steer the active DC back with a planned switchover:

```bash
$ kubectl annotate cassandra -n demo cas-dcdr dr.kubedb.com/switchover-to=dc-a
```

> Tune `max_hint_window_in_ms` for the outage you want hinted handoff to cover. Beyond
> the hint window, hints are dropped and only `nodetool repair` reconciles the gap, so run
> a full cross-DC repair as part of any failback that outlasted the hint window.

## Scaling and day-2 operations

The standard `CassandraOpsRequest` operations (`VerticalScaling`, `HorizontalScaling`,
`VolumeExpansion`, `UpdateVersion`, `Reconfigure`, `ReconfigureTLS`, `Restart`,
`RotateAuth`, `StorageMigration`) apply to a DC-DR cluster. They act on the distributed
per-DC datacenters across the DCs and are issued exactly as for a single-cluster
Cassandra. There is no failover ops type: there is nothing to fail over in a masterless
ring, and the planned switchover is the `dr.kubedb.com/switchover-to` annotation, not an
ops request.

`HorizontalScaling` gains a per-DC form so you can scale each DC's node count
independently:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cas-dcdr-hscale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: cas-dcdr
  horizontalScaling:
    dataCenters:
    - clusterName: dc-a
      replicas: 4
    - clusterName: dc-b
      replicas: 3
```

> After changing a DC's node count, run `nodetool cleanup` on the DC (removes data no
> longer owned after token ranges move) and a `nodetool repair` if you also changed a
> keyspace's replication factor.

> **Note:** the distributed Cassandra substrate and the DC-DR layer are net-new for
> Cassandra (KubeDB models Cassandra as a single logical DC today). Treat the field names
> and flows in this guide as the intended user experience; confirm availability in your
> release before relying on them in production.

## Deletion and cleanup

```bash
$ kubectl delete cassandra -n demo cas-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC datacenters and the cluster-scoped
per-DC `PlacementPolicies` it generated (these carry no owner reference, so the operator
deletes them explicitly). The user-provided base `PlacementPolicy` is left for you to
delete.

## Limitations

- **Adding or removing a whole data center** is a topology change (a datacenter and
  replication-factor change), performed by editing the `PlacementPolicy` topology and the
  keyspaces' `NetworkTopologyStrategy`, then running `nodetool repair`, not by a scaling
  request.
- Cross-DC replication is asynchronous; an unplanned failover has a non-zero RPO bounded
  by the hint and replication backlog (only recent `LOCAL_QUORUM` writes not yet
  replicated). Use `EACH_QUORUM` for keyspaces that cannot tolerate that window, and a
  planned switchover with a preceding repair for a strict handoff.
- User keyspaces must use `NetworkTopologyStrategy`; a `SimpleStrategy` keyspace is not
  datacenter-aware and is not DC-DR safe.
- There is no strong split-brain fence: a partitioned DC keeps serving `LOCAL_QUORUM`
  (Cassandra is AP). Divergent writes across a partition are reconciled by
  last-write-wins on cell timestamps during hinted handoff and repair on heal, not
  prevented.
