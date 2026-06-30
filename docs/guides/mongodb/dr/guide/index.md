---
title: DC-DR User Guide
menu:
  docs_{{ .version }}:
    identifier: mg-dr-guide-mongodb
    name: User Guide
    parent: mg-dr-mongodb
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Running MongoDB in DC-DR Mode: User Guide

This guide covers every aspect of operating a distributed MongoDB in cross data
center disaster recovery (DC-DR) mode: the components, the naming contract,
deployment, connecting with `w:majority`, reading from a secondary DC, monitoring,
lag and RPO, votes and roles, switchover and failback, scaling, and day-2 operations.

Read the [DC-DR Overview](/docs/guides/mongodb/dr/overview/index.md) first for the
architecture, and the [DC-DR Runbook](/docs/guides/mongodb/dr/runbook/index.md) for
scenario-by-scenario procedures.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Components and where they run

| Component | Runs in | Responsibility |
| --- | --- | --- |
| **`dr-controlplane`** + 3-site etcd quorum | across the data centers (an OCM control plane) | Publishes one `coordination.k8s.io` **Lease** per failover scope. The Lease holder is the DC whose members get the higher MongoDB `priority`. The Lease is policy and observability, not the failover mechanism. |
| **`dr-controlplane` agent** | each spoke (DC) | Contends for the primary-DC Lease for its DC and projects the Lease decision into the local spoke. |
| **KubeDB MongoDB operator (hub)** | the OCM hub | Expands the `MongoDB` CR into per-DC member groups in one replica set, steers `priority` by `replSetReconfig`, drives planned switchover, follows MongoDB's election with the Lease, and writes `status.disasterRecovery`. |
| **`replication-mode-detector`** | every data-bearing MongoDB pod | Polls `isMaster` and labels the elected primary `kubedb.com/role: primary` and secondaries `secondary`. Election is native; the operator never forces it. |
| **MongoDB voting arbiter** | the arbiter DC (even layout only) | A data-less voting member that supplies the tie-break vote, co-located with the third etcd member. |
| **KubeSlice** | each spoke | Provides the cross-DC pod network (`*.slice.local`) so the one replica set spans clusters and the oplog replicates across DCs. |

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

Map the global pod ordinals to data centers and tag each DC with its role:

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
```

- A data-bearing **Member** rule carries `replicaIndices`; the **Arbiter** DC carries
  none. Its single MongoDB voting arbiter is scheduled onto the arbiter spoke and
  co-located with the third etcd member.
- `mode: TwoDC` expects two Member DCs plus the Arbiter DC (the 2 + 2 + 1 even
  layout); `ThreeDC` expects an odd number of Member DCs and no Arbiter DC.
- Roles are `Member` and `Arbiter` only. The `Witness` role used by other engines is
  removed for MongoDB.

### MongoDB

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

### What the operator creates

- **One replica set** (`rs0`) whose data members are pinned to the Member DCs by the
  `PlacementPolicy` `distributionRules`. The oplog replicates across DCs natively;
  there is no second replication link.
- In the even (`TwoDC`) layout, **one data-less MongoDB voting arbiter** scheduled
  onto the Arbiter DC, so the vote total is odd (for example 2 + 2 + 1).
- A `<db>-secondary` read `Service` selecting `kubedb.com/role: secondary`, so clients
  can target standby-DC secondaries for cross-DC reads.
- Split-horizon **Horizons** DNS (`members[*].horizons.external`) so external clients
  reach each member by a routable name.

All data-bearing pods carry the offshoot selectors plus the
`open-cluster-management.io/cluster-name` label, so the single primary and secondary
Services and the single `AppBinding` keep working as the primary moves.

> The standby DC members are **non-hidden**, `votes:1`, low-`priority` electable
> secondaries. Do **not** make them `hidden:true`: hidden members serve no reads, are
> never election candidates, and get no role label, so a secondary Service would
> select nothing.

## Connecting

A DC-DR MongoDB exposes the same single endpoint as any KubeDB MongoDB:

- the **primary Service** `<db>` resolves to the active DC's writable primary (only
  that pod is labeled `kubedb.com/role: primary`);
- the **secondary Service** `<db>-secondary` resolves to the read-only secondaries
  across the standby DCs;
- one **`AppBinding`** `<db>` for applications and KubeDB integrations.

Because priority keeps the primary in the active DC and only that pod is labeled
`primary`, the endpoint follows failover automatically. Applications keep using `<db>`
and reconnect after a failover, landing on the new active DC.

### Write with `w:majority`

`w:majority` is the writable contract **and** the split-brain guarantee. Always write
with majority concern so a partitioned minority DC cannot commit:

```javascript
db.orders.insertOne(
  { item: "widget", qty: 1 },
  { writeConcern: { w: "majority", wtimeout: 10000 } }
)
```

- With `w:majority`, a primary that loses its majority auto-steps-down and a cut-off
  DC self-fences read-only, so no committed write is lost to split brain.
- With `w:1`, a write acknowledges before it replicates cross-DC. On an unplanned
  active-DC loss the bounded loss is the un-replicated oplog tail, which MongoDB rolls
  back natively when the old primary rejoins.

### Read from a secondary DC

Target the secondary Service and a non-primary read preference to serve reads from a
nearer standby DC:

```javascript
// Connect to the <db>-secondary Service, then:
db.getMongo().setReadPref("secondaryPreferred")
db.orders.find({ item: "widget" })
```

Secondary reads are eventually consistent, bounded by `oplogLagSeconds`.

## Monitoring and observability

### status.disasterRecovery

The single CR carries the whole cross-DC view:

```bash
$ kubectl get mongodb -n demo mg-dcdr -o jsonpath='{.status.disasterRecovery}' | jq
```

| Field | Meaning |
| --- | --- |
| `activeDC` | The DC whose members hold the higher priority and run the elected primary. |
| `phase` | `Steady`, `FailingOver`, `FailingBack`, or `Degraded`. |
| `lastTransitionTime` | When `activeDC` last changed. |
| `dataCenters[].clusterName` | The data center, by its OCM managed cluster name. |
| `dataCenters[].role` | `Member` or `Arbiter`. |
| `dataCenters[].primary` | That DC's elected primary pod, empty if the DC holds no primary. |
| `dataCenters[].writable` | True only for the active DC. |
| `dataCenters[].oplogLagSeconds` | The DC's cross-DC oplog lag behind the active primary, in seconds. |
| `dataCenters[].healthy` | Whether the DC has a ready member. |

### Useful checks

```bash
# Which DC the Lease intends as active (from the coordination plane):
$ kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc \
    -o jsonpath='{.spec.holderIdentity}'

# Per-DC members and roles:
$ kubectl get pods -n demo -l app.kubernetes.io/instance=mg-dcdr \
    -L kubedb.com/role,open-cluster-management.io/cluster-name

# The replica-set config and members (against the primary):
$ kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.conf().members.map(m => ({host:m.host, priority:m.priority, votes:m.votes, arbiterOnly:m.arbiterOnly}))'

# Replication lag from the primary's view:
$ kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.printSecondaryReplicationInfo()'
```

## Replication, lag, and RPO

- Cross-DC replication is the **native MongoDB oplog**, asynchronous. There is exactly
  one logical replica set, so there is no extra replication link to manage.
- `oplogLagSeconds` is how far a DC's members are behind the active primary's optime,
  computed in-DC and surfaced into status. It is the basis for the RPO of an unplanned
  failover.
- A **planned switchover loses near-zero committed writes**, because the non-force
  `replSetStepDown` only proceeds when an electable target secondary is caught up. An
  **unplanned failover** may lose the last un-replicated `w:1` oplog tail (bounded by
  the standby lag when the active DC died); `w:majority` writes are never lost.

## Votes, roles, and the arbiter

- **Votes are spread 3-site so no single data DC holds a majority.** The hub keeps the
  votes balanced and steers only `priority`, never `votes`, in the steady and failover
  paths.
- The requirement is an **odd total of voting members**, not an odd DC count:
  - **Even layout** (two data DCs plus the Arbiter DC): keep equal voting members per
    data DC and let the single data-less arbiter supply the odd vote (the 2 + 2 + 1
    shape). Do not add per-DC arbiters; an extra vote in one data DC would break the
    symmetry that lets either data DC plus the arbiter elect.
  - **Odd layout** (three or more Member DCs, no Arbiter DC): cap the data DCs to an
    odd voting total, typically one voting member per DC with extra replicas at
    `votes:0`.
  - In both layouts, use `votes:0` members for extra read redundancy without changing
    the vote balance.
- The **Arbiter DC** holds the `dr-controlplane` etcd vote **and** one MongoDB voting
  arbiter, co-located so the two quorums agree.

## Planned switchover (near-zero RPO)

Move the active DC on purpose by annotating the MongoDB:

```bash
$ kubectl annotate mongodb -n demo mg-dcdr dr.kubedb.com/switchover-to=dc-b
```

The hub then:

1. checks the target is a known, healthy DC within the oplog lag budget;
2. sets `phase: FailingOver` and raises the target DC's member `priority` by a normal
   majority-committed `replSetReconfig`;
3. issues a **non-force** `replSetStepDown` on the current primary, which only
   succeeds once an electable target secondary is caught up (the catch-up gate);
4. once MongoDB elects the new primary in the target DC, moves the Lease to match.

Watch `status.disasterRecovery` for `phase` returning to `Steady` with the new
`activeDC`.

## Failback

Failback is native. A DC that lost the primary and rejoins rolls back any
un-replicated tail automatically (rollback files), or does a full initial resync if it
fell outside the rollback/oplog window. There is no `pg_rewind` step to run.

Once the returned DC is caught up, steer the primary back with a planned switchover:

```bash
$ kubectl annotate mongodb -n demo mg-dcdr dr.kubedb.com/switchover-to=dc-a
```

## Scaling and day-2 operations

The standard `MongoDBOpsRequest` operations (`VerticalScaling`, `VolumeExpansion`,
`UpdateVersion`, `Reconfigure`, `ReconfigureTLS`, `Restart`, `Reprovision`,
`RotateAuth`, `Horizons`, `StorageMigration`) apply to a DC-DR cluster. They act on the
distributed member groups across the DCs and are issued exactly as for a
single-cluster MongoDB. There is no failover ops type: failover is MongoDB's native
election, and the planned switchover is the `dr.kubedb.com/switchover-to` annotation,
not an ops request.

> **Note:** the distributed MongoDB substrate and the DC-DR layer are net-new for
> MongoDB. Treat the field names and flows in this guide as the intended user
> experience; confirm availability in your release before relying on them in
> production.

## Deletion and cleanup

```bash
$ kubectl delete mongodb -n demo mg-dcdr
```

Per `deletionPolicy`, the operator removes the per-DC member groups, the MongoDB
arbiter, and the cluster-scoped per-DC `PlacementPolicies` it generated (these carry
no owner reference, so the operator deletes them explicitly). The user-provided base
`PlacementPolicy` is left for you to delete.

## Limitations

- **Adding or removing a whole data center** is a topology change (a member-group and
  cross-DC seed change), performed by editing the `PlacementPolicy` topology, not by a
  scaling request.
- Cross-DC oplog replication is asynchronous; an unplanned failover has a non-zero RPO
  bounded by the standby lag with `w:1`. Use `w:majority` for the split-brain
  guarantee and a planned switchover for a near-zero-RPO move.
- In the 2 + 2 + 1 even layout, a full data-DC loss stalls `w:majority` writes until
  the operator reconfigs the lost members out. Prefer an odd number of Member DCs to
  avoid the stall.
