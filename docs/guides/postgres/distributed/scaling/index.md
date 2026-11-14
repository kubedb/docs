---
title: Horizontal Scaling of a DC-DR Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-scaling
    name: Horizontal Scaling
    parent: guides-postgres-distributed
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scaling of a DC-DR Postgres

A DC-DR Postgres has one Raft group per data center, so "scale to 5" is ambiguous: 5 where? Horizontal scaling of a DC-DR database is therefore **per data center**. You give the new local node count of one or more data centers; data centers you do not list are left unchanged.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/distributed/scaling/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/scaling/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Before You Begin

- You need a DC-DR Postgres that is `Ready`, `Steady` and `protected`, created as in [Provision a DC-DR Postgres](/docs/guides/postgres/distributed/provision/index.md).
- You should be familiar with [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md) and with the plain [horizontal scaling](/docs/guides/postgres/scaling/horizontal-scaling/overview/index.md) flow.

This page scales `pg-docs` from [Provision a DC-DR Postgres](/docs/guides/postgres/distributed/provision/index.md): two nodes in `spoke-1` and two in `spoke-2`, with `hub` as the Arbiter.

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
```

Find the active data center first; you scale the other one:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{.status.phase} {.status.disasterRecovery.phase} active={.status.disasterRecovery.activeDC} protected={.status.disasterRecovery.protected}{"\n"}'
Ready Steady active=spoke-1 protected=true
```

The examples below have `spoke-1` active and scale `spoke-2`. If `spoke-2` is active in your installation, either move the primary to `spoke-1` with a [planned switchover](/docs/guides/postgres/distributed/switchover-failover/index.md#planned-switchover) first, or change `clusterName` in the YAML files to `spoke-1`.

## How per data center scaling works

You set `spec.horizontalScaling.dataCenters` on a `HorizontalScaling` PostgresOpsRequest. Each entry names a data center by its `clusterName` (as in the PlacementPolicy) and its desired local node count:

```yaml
spec:
  type: HorizontalScaling
  databaseRef:
    name: pg-docs
  horizontalScaling:
    dataCenters:
    - clusterName: spoke-2
      replicas: 3
```

For each listed data center whose live count differs, KubeDB:

- adds or removes **one node at a time** against that data center's own Raft group, so the data center never loses its Raft quorum;
- when the node to remove is the data center's Raft leader, transfers leadership to another node first;
- keeps the per-DC voter count odd: a data center with an odd node count has its arbiter removed, one with an even count gets an arbiter `<db>-<dc>-arbiter`;
- renumbers the `replicaIndices` of the PlacementPolicy so that the declared topology matches, and updates `spec.replicas` to the new total.

Scaling one data center never touches the other data center's Raft group or the cross-DC writability decision.

> **Important:** Scale the **standby** data center. Scaling the **active** data center works and loses no data, but in this release it recreates every pod of that data center, the primary included, without a handover first. In an example run the database had no writable primary for about 82 s. To grow or shrink the active data center, first move the primary away with a [planned switchover](/docs/guides/postgres/distributed/switchover-failover/index.md#planned-switchover), then scale it while it is the standby.

## Scale up the standby data center

The database starts with two nodes in each data center. `spoke-1` is active, so scale `spoke-2` from 2 to 3 nodes:

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/scaling/yamls/hscale-spoke-2-up.yaml
postgresopsrequest.ops.kubedb.com/pg-docs-hscale-spoke-2-up created
```

Watch the request:

```bash
$ kubectl --kubeconfig $HUB -n demo get postgresopsrequest pg-docs-hscale-spoke-2-up -w
```

In this example, the request went through these conditions, and the database through these states:

| Time | OpsRequest (last condition) | Postgres / DR phase / active DC / protected |
|---|---|---|
| 10:18:31 | `Running=True` | `Ready` / `Steady` / spoke-1 / true |
| 10:18:46 | `ScaleDataCentersStep=True` | `Ready` / `Steady` / spoke-1 / true |
| 10:19:35 | `ScaleDataCentersStep--spoke-2=False` | `Critical` / `Degraded` / spoke-1 / false |
| 10:20:32 | `ScaleDataCentersStep--pg-docs-spoke-2-2=False` | `Critical` / `Degraded` / spoke-1 / false |
| 10:21:34 | `ScaleDataCentersStep--pg-docs-spoke-2-2=False` | `Critical` / `Steady` / spoke-1 / true |
| 10:22:42 | `Successful=True`, phase `Successful` after 253 s | `Ready` / `Steady` / spoke-1 / true |

While the standby data center is reshaped it is briefly not streaming, so the database reports `Degraded` and `protected: false`. The active data center keeps its primary throughout: a test client writing throughout had all 1525 acknowledged writes present afterwards, and only `pg-docs-spoke-1-0` was writable at any time.

`spoke-2` now runs three data pods, and its arbiter is gone because three nodes already form an odd Raft group:

```bash
$ kubectl --kubeconfig $S2 -n demo get pods -l app.kubernetes.io/instance=pg-docs -L kubedb.com/role
NAME                READY   STATUS    RESTARTS   AGE     ROLE
pg-docs-spoke-2-0   3/3     Running   0          3m39s   standby
pg-docs-spoke-2-1   3/3     Running   0          3m5s    standby
pg-docs-spoke-2-2   3/3     Running   0          110s    standby

$ kubectl --kubeconfig $HUB get placementpolicy pg-docs -o jsonpath='{.spec.clusterSpreadConstraint.distributionRules}'
[{"clusterName":"spoke-1","replicaIndices":[0,1],"role":"Member"},{"clusterName":"spoke-2","replicaIndices":[2,3,4],"role":"Member"},{"clusterName":"hub","replicaIndices":[],"role":"Arbiter"}]

$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{.spec.replicas}'
5
```

The other data center is untouched:

```bash
$ kubectl --kubeconfig $S1 -n demo get pods -l app.kubernetes.io/instance=pg-docs -L kubedb.com/role
NAME                        READY   STATUS    RESTARTS        AGE    ROLE
pg-docs-spoke-1-0           3/3     Running   6 (5h51m ago)   2d5h   primary
pg-docs-spoke-1-1           3/3     Running   6 (5h51m ago)   2d5h   standby
pg-docs-spoke-1-arbiter-0   2/2     Running   4 (5h51m ago)   2d5h   arbiter
```

## Scale down the standby data center

Scale `spoke-2` back to 2 nodes:

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/scaling/yamls/hscale-spoke-2-down.yaml
postgresopsrequest.ops.kubedb.com/pg-docs-hscale-spoke-2-down created
```

The request was `Successful` after 93 s in this example, and the database stayed `Steady` and `protected` throughout. With an even count again, `spoke-2` gets its arbiter back:

```bash
$ kubectl --kubeconfig $S2 -n demo get pods -l app.kubernetes.io/instance=pg-docs -L kubedb.com/role
NAME                        READY   STATUS    RESTARTS   AGE     ROLE
pg-docs-spoke-2-0           3/3     Running   0          7m11s   standby
pg-docs-spoke-2-1           3/3     Running   0          6m37s   standby
pg-docs-spoke-2-arbiter-0   2/2     Running   0          35s     arbiter

$ kubectl --kubeconfig $HUB get placementpolicy pg-docs -o jsonpath='{.spec.clusterSpreadConstraint.distributionRules}'
[{"clusterName":"spoke-1","replicaIndices":[0,1],"role":"Member"},{"clusterName":"spoke-2","replicaIndices":[2,3],"role":"Member"},{"clusterName":"hub","replicaIndices":[],"role":"Arbiter"}]
```

## Rules

- A data center cannot be scaled to 0. Adding or removing a whole data center is a topology change, not horizontal scaling.
- Use `dataCenters`, not `horizontalScaling.replicas`, on a DC-DR database.
- Run one scaling request at a time, and start it only when the database is `Ready`, `Steady` and `protected`.

## Cleaning up

```bash
kubectl --kubeconfig $HUB -n demo delete postgresopsrequest pg-docs-hscale-spoke-2-up pg-docs-hscale-spoke-2-down
```

## Next Steps

- [Teardown](/docs/guides/postgres/distributed/teardown/index.md).
