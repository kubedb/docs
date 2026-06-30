---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-dr-guide
    name: User Guide
    parent: guides-postgres-dr
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running Postgres in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed Postgres in cross data
center disaster recovery (DC-DR) mode: the components, the naming contract,
deployment, connecting, monitoring, replication and lag, timing and tuning, quorum
and roles, switchover and failback, scaling, day-2 operations, backup, and deletion.

Read the [DC-DR Overview](/docs/guides/postgres/dr/overview/index.md)
first for the architecture, and the
[DC-DR Runbook](/docs/guides/postgres/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the active (writable) DC. This is the single cross-DC failover authority. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease on behalf of its DC and projects the Lease decision into the local spoke as a marker `ConfigMap`. |
| **KubeDB Postgres operator (hub)** | the OCM hub | Expands the `Postgres` CR into per-DC groups, watches the Lease, drives failover/switchover, and writes `status.disasterRecovery`. |
| **`pg-coordinator`** | every Postgres pod | Runs the per-DC raft, reads the local marker, and fences its leader read-only when its DC is not active. |
| **KubeSlice** | each spoke | Provides the cross-DC pod network so a standby DC's leader can stream from the active DC's leader. |

The marker `ConfigMap` is the contract between the agent (producer) and the
coordinator (consumer):

```
ConfigMap primary-dc  (namespace: dc-failover, on each spoke)
  data.activeDC  = the DC the quorum currently trusts as primary
  data.renewTime = RFC3339, the observed primary-DC Lease renewTime
  data.quiesce   = the DC asked to hold read-only for a planned switchover (else empty)
```

The coordinator trusts the marker for 30s (the fence TTL); absent, stale,
unparseable, or naming another DC all mean *not active* and the leader stays
read-only. This is the fail-closed fence.

## The DC-name contract

One string identifies a data center everywhere. **Keep these identical:**

- the OCM spoke cluster name
- the agent `--dc-name`
- the primary-DC Lease `holderIdentity`
- the marker `data.activeDC`
- the pod label `open-cluster-management.io/cluster-name`
- the `PlacementPolicy` `distributionRule.clusterName`

## Operator configuration

Start the Postgres operator with:

```
--dc-dr-enabled
--dc-dr-coord-kubeconfig=<kubeconfig of the coordination control plane>
--dc-dr-local-dc=<the data center this operator instance runs in>
```

The per-DC pod coordinators automatically receive `DC_DR_ENABLED`, `DC_NAME`,
`DC_DR_NAMESPACE` (default `dc-failover`), and `DC_DR_MARKER` (default `primary-dc`)
through their PetSet template, so the fence works without extra wiring.

## Deploying

### PlacementPolicy

Map the global pod ordinals to data centers and tag each DC with its role:

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
  (vote only, no Postgres) carries none. (The petset `Witness` role, a data-bearing
  witness, is for engines like MongoDB and is not used by Postgres.)
- `mode: TwoDC` expects exactly two Member DCs plus the Arbiter witness DC;
  `ThreeDC` expects at least three Member DCs.

### Postgres

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-dcdr
  namespace: demo
  annotations:
    dr.kubedb.com/enabled: "true"          # opt into per-DC DC-DR expansion
    # dr.kubedb.com/failover-group: payments  # optional: a Group failover scope
    # dr.kubedb.com/switchover-max-lag-bytes: "16777216"  # optional lag budget override
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

### What the operator creates

Per data-bearing DC `<dc>`:

- a per-DC `PetSet` `<db>-<dc>` (e.g. `pg-dcdr-dc-east`) with its own intra-DC raft;
- a DC-local headless governing `Service` so the DC's pods discover only each other;
- a cluster-scoped per-DC `PlacementPolicy` `<base>-<dc>` pinning that group to the DC;
- a per-DC arbiter `PetSet` `<db>-<dc>-arbiter` when that DC's local node count is even.

The witness DC (`role: Arbiter`) runs no Postgres pods. All per-DC pods carry the offshoot selectors
plus the `open-cluster-management.io/cluster-name` label, so the global primary/standby
Services and the single `AppBinding` keep working.

## Connecting

A DC-DR Postgres exposes the same single endpoint as any KubeDB Postgres:

- the **primary Service** `<db>` resolves to the active DC's writable leader (only
  that leader is labeled `kubedb.com/role: primary`);
- the **standby Service** `<db>-standby` resolves to the read-only leaders;
- one **`AppBinding`** `<db>` for applications and KubeDB integrations.

Because only the active DC's leader carries the `primary` label, the endpoint follows
failover automatically — applications keep using `<db>` and reconnect after a
failover, landing on the new active DC.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get pg -n demo pg-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC that holds the Lease and runs the writable primary. |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | `primary` for the active DC's leader, else `standby`. |
| `dataCenters[].leader` | That DC's local raft leader pod. |
| `dataCenters[].writable` | True only for the active DC. |
| `dataCenters[].lagBytes` | The DC's cross-DC replication lag behind the active primary. |
| `dataCenters[].healthy` | Whether the DC has a ready pod. |

### Useful checks

```bash
# Which DC is active (from the coordination plane):
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# The marker each spoke reads (run against a spoke):
$ kubectl -n dc-failover get configmap primary-dc -o yaml

# Per-DC leaders and roles:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=pg-dcdr \
    -L kubedb.com/role,open-cluster-management.io/cluster-name

# A standby DC leader stamps its lag here:
$ kubectl get pod -n demo <leader-pod> -o jsonpath='{.metadata.annotations.kubedb\.com/dc-lag-bytes}'
```

## Replication, lag, and RPO

- Cross-DC replication is **asynchronous** leader-to-leader streaming. Within a
  standby DC, the local followers **cascade** from their DC's leader, so each standby
  DC opens exactly one cross-DC link.
- `lagBytes` is how far a DC's leader is behind the active primary, computed by that
  DC's coordinator (the hub never opens cross-cluster SQL). It is the basis for the
  RPO of an unplanned failover.
- A **planned switchover loses no committed rows** (zero RPO) because writes are
  frozen and the target fully catches up before the handoff. An **unplanned failover**
  may lose the last unreplicated bytes (bounded by the standby's lag at the moment the
  active DC died).

## Timing and tuning (RTO vs safety)

DC-DR has one timing invariant that must hold for correctness:

> **fence TTL + cross-DC clock skew < primary-DC Lease duration**

The marker `renewTime` tracks the Lease's renewTime. A partitioned active DC
self-fences at `lastRenew + fence TTL`; a survivor can only acquire the expired Lease
at `lastRenew + LeaseDuration`. Keeping the fence TTL inside the Lease duration
guarantees the old active DC goes read-only **before** any new DC becomes writable —
no split-brain window.

Default values:

| Parameter | Where | Default |
| --- | --- | --- |
| Fence TTL | pg-coordinator | 30s |
| Marker refresh interval | dr-controlplane agent | 5s |
| Primary-DC Lease duration | dr-controlplane agent (`--election-lease-duration`) | 45s |
| Lease renew deadline | dr-controlplane agent (`--election-renew-deadline`) | 30s |
| Lease retry period | dr-controlplane agent (`--election-retry-period`) | 2s |

The failover **RTO floor** is roughly the Lease duration (the time a survivor waits to
acquire). To lower RTO, lower the Lease duration **and** the fence TTL together,
always preserving `fence TTL + skew < LeaseDuration`. The retry period must stay well
under the fence TTL so the holder restamps `renewTime` and the marker reads fresh in
normal operation.

## Quorum, roles, and arbiters

- Each DC's raft needs its own quorum. A DC with an **even** local node count gets its
  own in-DC arbiter (`<db>-<dc>-arbiter`) so intra-DC failover keeps quorum; an odd
  count needs none.
- The witness DC (`role: Arbiter`) holds only the `dr-controlplane` vote, never Postgres data.
- Scaling a DC re-evaluates its parity automatically: the arbiter is created or removed
  (and de-registered from the DC raft) as the local count crosses even/odd.

Separately from a DC's *intra-DC* quorum, the **cross-DC** failover quorum needs a
majority of three voting sites. For how to lay this out across two or three data
centers (and why a third witness site is preferred), see
[Deployment topologies](/docs/guides/postgres/dr/overview/index.md#deployment-topologies-2-dcs-vs-3-dcs).

## Planned switchover (zero-RPO)

Move the active DC on purpose by annotating the Postgres:

```bash
$ kubectl annotate pg -n demo pg-dcdr dr.kubedb.com/switchover-to=dc-west
```

The hub then:

1. checks the target is a known, healthy DC within the lag budget
   (`dr.kubedb.com/switchover-max-lag-bytes`, default 16 MiB);
2. sets `phase: FailingOver` and asks the active DC to **quiesce** (hold its primary
   read-only) via the primary-DC Lease, freezing the active write position;
3. waits until the target has replayed to within one WAL page of that frozen position;
4. hands the Lease to the target, which is promoted; the old DC resumes as a standby.

The annotation is cleared automatically once the target is active. Watch
`status.disasterRecovery` for `phase` returning to `Steady` with the new `activeDC`.

## Failback

Failback is just a switchover back to the original DC once it is healthy again:

```bash
$ kubectl annotate pg -n demo pg-dcdr dr.kubedb.com/switchover-to=dc-east
```

A DC that lost the Lease and rejoins automatically rewinds any divergent WAL tail
(`pg_rewind`, with a base-backup reseed fallback) and resumes streaming from the
current active primary before it is eligible.

## Per-DC horizontal scaling

Each DC has its own raft, so scale a specific DC with a `PostgresOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-dcdr-scale-west
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pg-dcdr
  horizontalScaling:
    dataCenters:
    - clusterName: dc-west
      replicas: 5
```

- Each entry sets that DC's local node count; DCs not listed are unchanged.
- Nodes are added or removed one at a time over the DC-local network; the DC's arbiter
  is created/removed as parity changes; on scale-down the removed node's replication
  slot is dropped.
- The base `PlacementPolicy` is renumbered so the declarative topology matches.
- Scaling a Member DC to `1` makes it a single-node DC (no in-DC HA, still part of
  cross-DC DR). Scaling to `0` is rejected — removing a whole DC is a topology change,
  not horizontal scaling.

## Day-2 operations

The standard `PostgresOpsRequest` operations apply to every per-DC group on a DC-DR
cluster; issue them exactly as for a non-distributed Postgres:

| Operation | DC-DR behavior |
| --- | --- |
| **Vertical scaling** | Patches every per-DC PetSet and per-DC arbiter, restarts per-DC pods. |
| **Volume expansion** (online/offline) | Expands every per-DC data PVC and per-DC arbiter PVC, waits on all per-DC PetSets. |
| **Version update** | Updates every per-DC PetSet. |
| **Storage migration** | Orphan-deletes and waits on every per-DC PetSet. |
| **Reconfigure / Restart / Rotate-Auth** | Apply across the per-DC pods. |

## Backup

Back up a DC-DR Postgres the same way as any KubeDB Postgres (KubeStash / the Postgres
archiver). Logical and base backups run against the writable endpoint, so they read
from the active DC; the AppBinding follows failover, so a scheduled backup continues
against the new active DC after a failover. Point-in-time recovery works as usual.

## Deletion and cleanup

```bash
$ kubectl delete pg -n demo pg-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC PetSets, governing Services, and
the cluster-scoped per-DC `PlacementPolicies` it generated (these carry no owner
reference, so the operator deletes them explicitly). The user-provided base
`PlacementPolicy` is left for you to delete.

## Limitations

- **Adding or removing a whole data center** is a topology change (a new group, raft,
  and cross-DC seed), distinct from horizontal scaling, and is performed by editing the
  `PlacementPolicy` topology, not a `HorizontalScaling` request.
- Cross-DC replication is asynchronous; an unplanned failover has a non-zero RPO
  bounded by the standby lag. Use a **planned switchover** for zero-RPO moves.
- All correctness depends on the timing invariant above; do not set a fence TTL that
  meets or exceeds the Lease duration.
