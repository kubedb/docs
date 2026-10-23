---
title: Switchover and Failover of a DC-DR Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-switchover-failover
    name: Switchover and Failover
    parent: guides-postgres-distributed
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Switchover and Failover of a DC-DR Postgres

This page moves the primary of a DC-DR Postgres between data centers:

| Operation | When to use it | Data loss | How |
|---|---|---|---|
| [Planned switchover](#planned-switchover) | Both data centers are healthy; you want the primary elsewhere (maintenance, failback) | None | Annotate the Postgres |
| [Planned failover](#planned-failover-lease-handoff) | The active data center is alive but its database cannot serve, and you decide to move | Bounded by the replication lag | Annotate the failover Lease |
| [Unplanned failover](#unplanned-failover-loss-of-the-active-data-center) | The active data center is lost | Writes not yet replicated | Automatic |
| [Failback](#failback) | After a failover, the old data center has rejoined | None | Planned switchover |

It also covers how to keep a data center from becoming primary (standby-hold), the emergency break-glass override, and how to bound the data loss of an unplanned failover.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/distributed/switchover-failover/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/switchover-failover/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Before You Begin

- Provision `pg-docs` as in [Provision a DC-DR Postgres](/docs/guides/postgres/distributed/provision/index.md).
- Read [Concepts](/docs/guides/postgres/distributed/concepts/index.md), in particular the failover scope.

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
```

### Every operation moves the whole failover scope

All databases that follow the same primary-DC Lease move together: a switchover or failover requested on one of them moves all of them. Before any operation, check that the database is alone in its scope, or that you mean to move every database in it:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{range .status.conditions[?(@.type=="DCDRFailoverScopeShared")]}{.status} {.message}{"\n"}{end}'
False This database is the only member of failover scope "primary-dc-pg-docs", so it fails over independently.
```

When the condition is `True`, its message names the other databases. To give a database its own scope, set `failoverPolicy.failoverGroupRef` on its PlacementPolicy to a group name no other database uses, as `pg-docs` does.

### Check before you start

Start a planned operation only when the database is `Ready`, `Steady` and `protected`:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{.status.phase} {.status.disasterRecovery.phase} active={.status.disasterRecovery.activeDC} protected={.status.disasterRecovery.protected}{"\n"}'
Ready Steady active=spoke-2 protected=true
```

The status fields are not enough on their own. Also check the pods: exactly one pod with `kubedb.com/role=primary` in all clusters, and every standby streaming (see [Provision, Step 5](/docs/guides/postgres/distributed/provision/index.md#step-5-write-data-and-see-it-in-the-other-data-center)).

The example outputs on this page were recorded with a test client in `psql-client` on spoke-1 that inserted a row through `pg-docs.demo.svc.slice.local` every half second and recorded every acknowledged id. After each operation, every acknowledged id was checked against the table. Your timings will differ with your network and write load.

## Planned switchover

A planned switchover moves the primary with no loss of acknowledged writes. Set the target data center on the Postgres object:

```bash
$ kubectl --kubeconfig $HUB -n demo annotate pg pg-docs dr.kubedb.com/switchover-to=spoke-1 --overwrite
postgres.kubedb.com/pg-docs annotated
```

The orchestrator stamps the start time and requests a quiesce of the active data center:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{.metadata.annotations}'
{"dr.kubedb.com/enabled":"true","dr.kubedb.com/quiesce-active":"true","dr.kubedb.com/switchover-started":"2026-10-23T04:09:02Z","dr.kubedb.com/switchover-to":"spoke-1", ...}
```

The quiesce reaches the active data center through its marker ConfigMap, and `pg-coordinator` write-locks the primary:

```bash
$ kubectl --kubeconfig $S2 -n dc-failover get cm primary-dc-pg-docs -o jsonpath='{.data}'
{"activeDC":"spoke-2","quiesce":"spoke-2","renewTime":"2026-10-23T04:09:03Z"}

$ kubectl --kubeconfig $S2 -n demo logs pg-docs-spoke-2-1 -c pg-coordinator | grep write-lock
I1023 04:09:09.622327       1 dcdr_quiesce.go:134] "DC-DR: primary write-lock changed" dcdr.db="demo/pg-docs" dcdr.dc="spoke-2" dcdr.raftLeader=2 dcdr.myId=2 pod="pg-docs-spoke-2-1" writeLock=true reason="planned-switchover quiesce"
```

Then the target catches up to the frozen primary position (within 8 KiB), the Lease is handed to the target, the target's leader is promoted, and the old primary becomes a standby. Watch it on the hub:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -w
```

The observed sequence (Postgres phase, DR phase, active data center, `protected`, holder of `primary-dc-pg-docs`):

| Time | Status | Lease |
|---|---|---|
| 04:09:01 | annotation set | spoke-2 |
| 04:09:08 | `Ready FailingOver active=spoke-2 protected=true` | spoke-2 |
| 04:09:15 | `Ready FailingOver active=spoke-2 protected=true` | spoke-1 |
| 04:09:21 | `Ready Steady active=spoke-1 protected=true` | spoke-1 |
| 04:10:05 | `Critical Degraded active=spoke-1 protected=false` (old primary rejoining as a standby) | spoke-1 |
| 04:11:27 | `Critical Steady active=spoke-1 protected=true` | spoke-1 |
| 04:12:43 | `Ready Steady active=spoke-1 protected=true` | spoke-1 |

The new data center was `Steady` 20 s after the request. While the old primary restarts as a standby, the database briefly reports `Critical` and `Degraded`; it is writable on the new data center throughout. When the switchover completes, the orchestrator removes the `switchover-to`, `switchover-started` and `quiesce-active` annotations.

```bash
$ kubectl --kubeconfig $S1 -n demo get pods -l app.kubernetes.io/instance=pg-docs -L kubedb.com/role
NAME                        READY   STATUS    RESTARTS   AGE   ROLE
pg-docs-spoke-1-0           3/3     Running   0          14m   primary
pg-docs-spoke-1-1           3/3     Running   0          13m   standby
pg-docs-spoke-1-arbiter-0   2/2     Running   0          11m   arbiter

$ kubectl --kubeconfig $S2 -n demo get pods -l app.kubernetes.io/instance=pg-docs -L kubedb.com/role
NAME                        READY   STATUS    RESTARTS        AGE   ROLE
pg-docs-spoke-2-0           3/3     Running   3 (8m45s ago)   16m   standby
pg-docs-spoke-2-1           3/3     Running   3 (8m45s ago)   15m   standby
pg-docs-spoke-2-arbiter-0   2/2     Running   2 (8m45s ago)   14m   arbiter
```

What the writer saw: the last write before the switchover was acknowledged at 04:09:09, writes failed from 04:09:09 to 04:09:52, and succeeded again from 04:09:53. **All 412 acknowledged writes were present** afterwards. The database was writable on spoke-1 at about 04:09:21; most of the 44 s gap seen by the client is the time `pg-docs.demo.svc.slice.local` takes to point to the new primary (about 30 s). Configure applications to retry.

In repeated tests, planned switchovers reached `Steady` in 11-15 s, and no acknowledged write was lost.

### Abort a switchover

A switchover waits until the target has caught up. If the target cannot catch up, the active primary stays write-locked until the switchover completes or is aborted. Abort it with:

```bash
$ kubectl --kubeconfig $HUB -n demo annotate pg pg-docs dr.kubedb.com/switchover-abort=true --overwrite
```

Presence of the annotation aborts, whatever its value. Do not just remove `switchover-to`. The orchestrator lifts the quiesce, writes resume on the original data center, the Lease does not move, and all switchover annotations are removed. In an example with the target's replay paused, the switchover held at the catch-up gate with the active data center write-locked, and the abort restored writes 6 s later, with no acknowledged write lost:

```
05:07:48 request switchover-to=spoke-2
05:07:53 Ready/FailingOver/spoke-1/true quiesce=true   lease=spoke-1
...      (held at the catch-up gate)
05:09:20 set switchover-abort=true
05:09:30 Ready/Steady/spoke-1/true                     lease=spoke-1
```

A switchover that has not completed 10 minutes after the quiesce aborts by itself. To change the deadline for one database, set `dr.kubedb.com/switchover-timeout` (a Go duration such as `5m`) on it.

### Switchover settings

All are annotations on the Postgres object:

| Annotation | Meaning | Default |
|---|---|---|
| `dr.kubedb.com/switchover-to` | Target data center; setting it starts the switchover | - |
| `dr.kubedb.com/switchover-abort` | Presence aborts the switchover | - |
| `dr.kubedb.com/switchover-timeout` | Auto-abort deadline after the quiesce | `10m` |
| `dr.kubedb.com/switchover-zero-rpo-lag-bytes` | Remaining lag allowed at the handoff; `0` demands an exact catch-up | `8192` |
| `dr.kubedb.com/switchover-max-lag-bytes` | Lag within which a standby data center counts as protected | the RPO budget if set, else 16 MiB |

`dr.kubedb.com/switchover-started` and `dr.kubedb.com/quiesce-active` are set by the orchestrator. Do not set or edit them.

A switchover requires the active primary to be up: the orchestrator measures the target's lag and confirms the write-lock by connecting to it. When the active primary is down but its data center is alive, use a planned failover instead.

Name only a Member data center as the target. A switchover to the Arbiter is not rejected up front; it stays `FailingOver` until you abort it.

## Planned failover (Lease handoff)

A planned failover moves the failover authority to the other data center **without** the quiesce and catch-up of a switchover. Use it when the active data center is alive but its database cannot serve, for example when its primary is down and the data center has no healthy local standby to take over. In that state nothing fails over by itself: the active data center's agent keeps renewing the Lease, because the Lease tracks the health of the data center, not of the database.

> **Warning:** A planned failover moves **every database in the scope**, and acknowledged writes that have not reached the target data center are lost. When the active primary is healthy, use a [planned switchover](#planned-switchover). Do not stop a data center's agent to force a failover: one agent serves every scope of its data center, and stopping it fails all of them over.

Set the target data center on the scope's Lease, on the hub:

```bash
$ kubectl --kubeconfig $HUB -n dc-failover annotate lease primary-dc-pg-docs dr.open-cluster-management.io/handoff-to=spoke-2
lease.coordination.k8s.io/primary-dc-pg-docs annotated
```

The current holder's agent releases the Lease, the target's agent acquires it within one 5 s retry, and the annotation clears itself:

```bash
$ kubectl --kubeconfig $HUB -n dc-failover get lease primary-dc-pg-docs -o jsonpath='holder={.spec.holderIdentity} t={.spec.leaseTransitions} ann={.metadata.annotations.dr\.open-cluster-management\.io/handoff-to}{"\n"}'
holder=spoke-2 t=3 ann=
```

The old active data center sees from its marker that another data center is now active, write-locks its primary at once and rejoins as a standby:

```bash
$ kubectl --kubeconfig $S1 -n demo logs pg-docs-spoke-1-0 -c pg-coordinator | grep -E 'fence holding|write-lock'
I1023 04:13:48.663123       1 dcdr_fence.go:433] "DC-DR fence holding this leader read only" marker="primary-dc-pg-docs" reason="another DC is active" dcdr.dc="spoke-1" dcdr.activeDC="" dcdr.markerRenew="2026-10-23T04:13:46Z"
I1023 04:13:48.676620       1 dcdr_quiesce.go:134] "DC-DR: primary write-lock changed" dcdr.db="demo/pg-docs" dcdr.dc="spoke-1" dcdr.raftLeader=1 dcdr.myId=1 pod="pg-docs-spoke-1-0" writeLock=true reason="another DC is the confirmed active DC"
```

Observed sequence:

| Time | Status | Lease |
|---|---|---|
| 04:13:45 | annotation set | spoke-1 |
| 04:13:49 | - | spoke-2 |
| 04:14:02 | `Ready Degraded active=spoke-2` | spoke-2 |
| 04:14:32 | `Critical Degraded active=spoke-2 protected=false` (spoke-1 rejoining) | spoke-2 |
| 04:15:34 | `Critical Steady active=spoke-2 protected=true` | spoke-2 |
| 04:17:21 | `Ready Steady active=spoke-2 protected=true` | spoke-2 |

The writer's writes failed from 04:13:48 to 04:14:14. The standby data center was streaming with no lag, so **all 882 acknowledged writes were present** afterwards. With replication lag, the writes acknowledged by the old primary that had not reached the target are lost.

If the promotion of the new data center does not complete, the orchestrator drives it with a `ForceFailOver` PostgresOpsRequest named `<db>-dcdr-failover-<suffix>`.

## Unplanned failover: loss of the active data center

Nothing needs to be done: when the active data center stops renewing its Lease, the other Member data center takes over.

1. The active data center's agent stops renewing `primary-dc-<group>`.
2. After the 45 s lease duration, the other Member's agent acquires the Lease.
3. The markers flip, and the orchestrator promotes the surviving data center's cross-DC streamer (if needed through a `<db>-dcdr-failover-<suffix>` `ForceFailOver` PostgresOpsRequest).
4. When the lost data center comes back, its pods rejoin as standbys of the new primary; a diverged old primary is rewound with `pg_rewind`.

The following status sequence was recorded in an example where the VM of the active data center, `spoke-2`, was rebooted, with a writer in `spoke-1` (Postgres phase, DR phase, active data center, `protected`, Lease holder and transition count):

```
10:02:51 Ready     Steady      active=spoke-2 protected=true   lease=spoke-2/t4
10:03:32 reboot of the spoke-2 VM
10:04:03                                                       lease=spoke-1/t5
10:04:42 Critical  FailingOver active=spoke-1                  lease=spoke-1/t5
10:04:47 Critical  Degraded    active=spoke-1                  lease=spoke-1/t5
10:04:58 NotReady  Degraded    active=spoke-1                  lease=spoke-1/t5
10:06:53 Ready     Degraded    active=spoke-1                  lease=spoke-1/t5
10:07:10 Ready     Steady      active=spoke-1 protected=true   lease=spoke-1/t5
```

- The Lease moved 31 s after the reboot started, and the new primary in `spoke-1` accepted writes at 10:04:57, about 85 s after the reboot.
- The old primary's last write was at 10:03:35. No two pods were ever writable at the same time.
- All 564 acknowledged writes were present. Because replication is asynchronous, writes that the lost primary acknowledged but had not yet sent are lost; the standby was streaming with no lag in this test.
- `spoke-2` came back and rejoined as a standby by itself; the database was `protected` again 3.5 minutes after the reboot.

While the lost data center is down there is no second copy of the data: the status is `Degraded` and `protected` is not `true`. A second failure in that window cannot be recovered from the other data center.

### A cut-off active data center

If the active data center keeps running but loses its connection to the hub and the other data center, it fences itself before the other data center can take over (see [Split-brain protection](/docs/guides/postgres/distributed/concepts/index.md#split-brain-protection)). In an example where the active data center `spoke-1` was isolated by a firewall rule for 360 s:

```
09:54:26  network cut
09:54:55  spoke-1 marker stale: "DC-DR fence holding this leader read only" reason="marker stale"
09:55:06  spoke-1 primary write-locked: reason="fence uncertain (marker stale, missing or unreadable) past the grace"
09:55:23  Lease moved to spoke-2
09:56:13  spoke-2 primary writable
10:01:17  Ready Steady active=spoke-2 protected=true (after the cut was lifted)
```

There were never two writable primaries. A writer in `spoke-2` lost none of its 1248 acknowledged writes. A writer **inside** the isolated data center had 185 writes acknowledged after the cut; all of them were acknowledged before the write-lock at 09:55:06, and 184 of them were lost when `spoke-1` rejoined. This window, about 40 s from the cut to the write-lock, is the designed data loss of an isolated active data center.

> **Note:** When you compare writes after such a split, compare by content (for example id and timestamp), not by id alone. Both timelines reuse the same sequence values after the fork.

### Bound the data loss of an unplanned failover

`spec.replication.bestEffortCrossDCLagBytesForFailover` sets how many bytes of WAL an unplanned failover may lose. When the surviving data center is further behind the lost primary than this budget, or its lag cannot be established, it refuses to promote, and the database stays down until you decide:

```yaml
spec:
  replication:
    walLimitPolicy: WALKeepSize
    walKeepSize: 1536
    maxSlotWALKeepSize: 8192
    bestEffortCrossDCLagBytesForFailover: 1048576
```

- Unset (the default), nothing is enforced. `0` demands a fully caught-up survivor.
- The figure is close to exact when the active data center is really lost. During a network partition where the old primary keeps running, it can be understated, so treat it as a budget, not a guarantee. Only synchronous replication bounds data loss.
- While a promotion is held, the database shows `protected: false` with a `protectionMessage`, the condition `DCDRPromotionStalled`, and the coordinator in the surviving data center logs how far behind it is, for example `10453120 bytes behind the last known primary position, over the 1048576 byte budget`.

To accept the loss and let the failover proceed:

```bash
$ kubectl --kubeconfig $HUB -n demo annotate pg pg-docs dr.kubedb.com/accept-failover-data-loss=true
```

The orchestrator removes the annotation once the failover it allowed has completed, so it never authorizes a later, unrelated loss.

## Failback

After an unplanned or planned failover, the old data center rejoins as a standby by itself. When the database is `Ready`, `Steady` and `protected` again, move the primary back, if you want to, with a [planned switchover](#planned-switchover):

```bash
$ kubectl --kubeconfig $HUB -n demo annotate pg pg-docs dr.kubedb.com/switchover-to=<original-dc> --overwrite
```

Wait for `protected: true` first: right after a failover the returning data center may still be rewinding and catching up.

## Keep a data center from becoming primary (standby-hold)

To keep a data center a standby, so that it never takes the Lease or promotes, for example while it is being repaired or while you do not trust its data, create a **standby-hold** ConfigMap on **that data center's spoke**, in `dc-failover`. Its name is the scope's Lease name plus `-standby-hold`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: primary-dc-pg-docs-standby-hold
  namespace: dc-failover
```

```bash
$ kubectl --kubeconfig $S1 apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/switchover-failover/yamls/standby-hold.yaml
configmap/primary-dc-pg-docs-standby-hold created
```

- The data center's agent stops contending for the Lease and logs `DC-DR standby-hold ConfigMap present, this DC will not contend for the primary DC Lease while it exists`. Its `pg-coordinator` keeps the data center's leader a read-only standby. The data center still replicates and still rejoins after a failover.
- Only you create and delete this ConfigMap; no component writes it. If `pg-coordinator` cannot read it (an error other than not-found), it assumes the hold is present.
- It has **no effect on the active data center**. There, the agent logs `standby-hold ignored on the active DC; to move the primary use a planned switchover (dr.kubedb.com/switchover-to)` every few seconds. Move the primary away first, then place the hold.
- While the only other Member data center is held, the scope cannot fail over at all: if the active data center is lost, nothing takes over.

Remove it to let the data center contend again; this takes effect immediately:

```bash
$ kubectl --kubeconfig $S1 -n dc-failover delete cm primary-dc-pg-docs-standby-hold
```

## Emergency: break-glass override

If the hub, or the connection to it, is lost for longer than the fence tolerates, the active data center fences itself read-only even though its database is healthy. This is the intended trade against split-brain. If you must keep that data center writable anyway, create the **override** ConfigMap (`<Lease name>-override`) on the spoke of the **current active** data center:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: primary-dc-pg-docs-override
  namespace: dc-failover
```

```bash
$ kubectl --kubeconfig <active-dc-spoke-kubeconfig> apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/switchover-failover/yamls/break-glass-override.yaml
configmap/primary-dc-pg-docs-override created
```

> **Warning:** While the override exists, split-brain protection for the scope is bypassed. Use it only when you know the state of the other data center, and never on two data centers.

- `pg-coordinator` reads it from its own spoke, so it works without the hub: the local leader stays writable regardless of the marker, and both `pg-coordinator` and the agent log `DC-DR BREAK GLASS OVERRIDE ACTIVE` continuously.
- The data center's agent marks the Lease with `dr.open-cluster-management.io/override-hold=<dc>`, and every other Member stops contending. The scope cannot fail over while the override stands, even if the pinned data center dies.
- It does not promote anyone. The agent pins the Lease only for the data center that last held it; on any other spoke the agent ignores it and logs that it refuses to pin the Lease to a data center that never held it.
- Remove it (`kubectl --kubeconfig <active-dc-spoke-kubeconfig> -n dc-failover delete cm primary-dc-pg-docs-override`) as soon as the hub is back and the Lease is held by the right data center. Before a data center that carried an override returns after an outage, delete its override ConfigMap, or its old primary comes back writable while the other data center holds the Lease.

## Troubleshooting

Read these in order; they agree within seconds on a healthy system, and the first one that disagrees shows the broken layer:

1. The Lease on the hub: `kubectl --kubeconfig $HUB -n dc-failover get lease primary-dc-pg-docs`.
2. The marker on each cluster: `kubectl --kubeconfig <cluster> -n dc-failover get cm primary-dc-pg-docs -o jsonpath='{.data}'`.
3. The Postgres status: `status.disasterRecovery.activeDC` and each data center's `writable`.
4. The pods: exactly one pod in all clusters with `kubedb.com/role=primary`.

Conditions on the Postgres object:

| Condition | Meaning |
|---|---|
| `DCDRFailoverScopeShared` | `True` when other databases share the failover scope; the message names them. |
| `DCDRPromotionStalled` | A cross-DC promotion is held, for example by the data-loss budget. |
| `ForceFailOverRetryCapReached` | Repeated `ForceFailOver` attempts failed and the orchestrator stopped creating new ones. Fix the cause shown in the failed `<db>-dcdr-failover-*` PostgresOpsRequests, then delete them to reset the count. |

## Cleaning up

```bash
kubectl --kubeconfig $S1 -n demo delete pod psql-client
```

To delete `pg-docs`, see [Teardown](/docs/guides/postgres/distributed/teardown/index.md#delete-a-dc-dr-database).

## Next Steps

- [Horizontal scaling](/docs/guides/postgres/distributed/scaling/index.md).
