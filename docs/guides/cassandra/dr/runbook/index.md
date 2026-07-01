---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: cas-dr-runbook-cassandra
    name: Runbook
    parent: cas-dr-cassandra
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Cassandra DC-DR Runbook

Scenario-by-scenario procedures for operating a Cassandra cluster in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB and
Cassandra do **automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/cassandra/dr/guide/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane
kubeconfig, and `cas-dcdr`/`demo` are the example database and namespace. The example pod
`cas-dcdr-rack-a-0` is the first node of rack `rack-a`.

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the DC the coordination plane routes writes to):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC nodes and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=cas-dcdr -L open-cluster-management.io/cluster-name

# Ring status per DC (UN = up/normal), from any node:
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status

# Cross-DC streaming, pending hints, pending ranges (the lag signal):
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool netstats
```

Golden rules:

- **Consistency level, not a fence, is the correctness knob.** Cassandra is AP and
  masterless. A partitioned DC keeps serving `LOCAL_QUORUM`; there is no cross-DC quorum
  that stops it committing. Reconciliation happens after the fact via hinted handoff and
  `nodetool repair`.
- **There is no promotion and no failover in the engine.** Every DC is a full writable
  Cassandra datacenter. DR is a routing change (the write endpoint follows the Lease to a
  surviving DC), not an election.
- **Exactly one DC is `writable: true`** in `status.disasterRecovery` at any instant (the
  write-routed active DC), even though the engine lets every DC accept writes. This is a
  routing marker, not an engine fence.
- **The Lease is routing, not safety.** If the Lease is stale, Cassandra keeps running on
  its own; what you lose is switchover and endpoint steering.
- **Use `EACH_QUORUM`** only for keyspaces that must be durable in every DC before ack; it
  fails while any DC is down.

---

## 1. Active DC lost (zone/cluster failure)

**Symptoms:** the active DC's nodes are gone/unreachable; writes to the endpoint fail
briefly until it re-points.

**Automatic:** the surviving DCs **keep accepting reads and writes at `LOCAL_QUORUM` on
their own**, because each DC acks locally and every DC is already a full writable
datacenter. There is no promotion and no fence. The orchestrator observes the Lease move
to a surviving DC and points the single write endpoint there. `phase` moves `FailingOver`
to `Steady`.

**Verify:**

```bash
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status                                   # lost DC's nodes DN/absent
```

**Action:** none required for availability. Note the RPO: only writes that acked at
`LOCAL_QUORUM` in the lost DC but had not yet replicated to survivors are at risk
(recoverable by repair if the DC returns). Keyspaces written at `EACH_QUORUM` lose zero on
writes that acked. When the failed DC returns, see scenario 8 (re-add a DC).

---

## 2. Network partition between data centers

**Symptoms:** DCs are up but cannot reach each other.

**Automatic:** **both sides keep serving `LOCAL_QUORUM` locally**, because Cassandra is AP
and each DC acks within itself. There is **no strong split-brain fence**: unlike
ClickHouse Keeper or Postgres raft, a minority DC is not stopped from committing. Both
sides may take writes during the partition. On heal, hinted handoff (within the hint
window) plus anti-entropy `nodetool repair` reconcile divergent writes by last-write-wins
on cell timestamps. The write endpoint stays on (or the orchestrator moves it to) the DC
the Lease resolves to.

**Verify both sides are serving and check hint backlog:**

```bash
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} writable:{.writable} hints:{.pendingHints}{"\n"}{end}'
```

**Action:** heal the network. The DCs re-gossip and reconcile automatically via hinted
handoff, then run a full cross-DC `nodetool repair` to converge anything beyond the hint
window. There is no rewind. If the divergence window matters for a keyspace, write it at
`EACH_QUORUM` so it does not ack during a partition.

---

## 3. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate cassandra -n demo cas-dcdr dr.kubedb.com/switchover-to=dc-b
```

**Automatic:** the hub gates on the target's health and hint/repair backlog, then moves
the Lease and the write endpoint to `dc-b`. Because every DC is already a full writable
datacenter, this is a routing move: there is no promotion and no engine catch-up gate. For
a strict zero-loss handoff, drain hints and run a cross-DC `nodetool repair` toward the
target first so it is fully converged.

**Verify:**

```bash
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-b
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 6 (switchover stuck).

---

## 4. Planned failback to the original DC

After the original DC is healthy and caught up (failback is native: it rejoins by gossip,
takes hinted handoff, and a cross-DC repair reconciles the rest, with no rewind), steer
the active DC back:

```bash
kubectl annotate cassandra -n demo cas-dcdr dr.kubedb.com/switchover-to=dc-a
```

Same routing-move flow as scenario 3. There is no `pg_rewind` step and no rollback;
Cassandra reconciles by last-write-wins on cell timestamps. Run a full cross-DC
`nodetool repair` first if the outage outlasted the hint window.

---

## 5. Arbiter DC lost (even layout only)

**Symptoms:** in an even (`TwoDC`) layout, the Arbiter DC is gone; its `dr-controlplane`
etcd member is unreachable. (The Arbiter DC runs **no Cassandra**, so no data or ring node
is affected.)

**Impact:** none on the Cassandra ring; both data DCs keep serving. You lose the tie-break
etcd vote, so the coordination plane may not be able to move the Lease until the Arbiter
DC (or its etcd member) returns. In the preferred odd (`ThreeDC`) layout there is no
Arbiter DC and this scenario does not apply.

**Verify the ring is unaffected and check the coordination plane:**

```bash
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status                                # both data DCs UN
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc                         # may be stale
```

**Action:** restore the Arbiter DC's etcd member to regain the coordination quorum and
Lease steering. The Cassandra ring needs nothing.

---

## 6. Planned switchover stuck (endpoint not moving)

**Symptoms:** after annotating `switchover-to`, `phase` stays `FailingOver` and the
endpoint does not move.

**Diagnose:**

```bash
# Target health and backlog from status:
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} healthy={.healthy} hints={.pendingHints} repair={.repairBacklog}{"\n"}{end}'
# Ring and streaming from the target DC:
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool netstats
```

**Causes & action:**

- **Target over the backlog budget** the switchover gates on the target's hint and repair
  backlog. Drain hints and run a cross-DC `nodetool repair` toward the target so it is
  converged, then retry.
- **Target unhealthy** ensure the target DC shows its expected UN nodes in `nodetool
  status`.
- **Coordination plane down** the Lease cannot move (see scenario 11); restore it.
- **Abort** remove the annotation to cancel:
  `kubectl annotate cassandra -n demo cas-dcdr dr.kubedb.com/switchover-to-`.

---

## 7. A standby DC is lost

**Symptoms:** a non-active DC's nodes are gone; that DC shows `healthy: false` and DN/absent
nodes in `nodetool status`.

**Impact:** none on writes. The active DC keeps serving `LOCAL_QUORUM` on its own. You lose
that DC's redundancy and its local read capacity until it returns, plus any keyspace
written at `EACH_QUORUM` will fail while that DC is down.

**Verify the active DC is serving:**

```bash
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName} up:{.upNormalNodes}/{.totalNodes}{end}'
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status
```

**Action:** recover the DC's nodes; they rejoin the ring by gossip and catch up through
hinted handoff. Run a cross-DC `nodetool repair` after they return if the outage outlasted
the hint window.

---

## 8. Re-add / recover a previously lost data center

After a DC returns from a failure:

**Automatic:** its nodes rejoin the ring by gossip and take hinted handoff for writes that
happened within the hint window. There is no rewind and no rollback; Cassandra reconciles
by last-write-wins on cell timestamps.

**Verify:**

```bash
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} healthy={.healthy} hints={.pendingHints} repair={.repairBacklog}{"\n"}{end}'
```

**Action:** run a **full cross-DC `nodetool repair`** to reconcile everything beyond the
hint window:

```bash
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool repair --full
```

To make the returned DC the active DC again, perform a planned failback (scenario 4) once
its hints are drained and the repair is complete.

---

## 9. Hint backlog growing (a DC falling behind)

**Symptoms:** `pendingHints` for a DC keeps rising in `status.disasterRecovery`; that DC is
lagging.

**Impact:** cross-DC replication is behind for that DC, so the RPO window on an unplanned
loss widens. `LOCAL_QUORUM` reads in the lagging DC may not see the newest writes from
other DCs yet.

**Verify:**

```bash
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool netstats
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} hints={.pendingHints} repair={.repairBacklog}{"\n"}{end}'
```

**Action:** relieve the cross-DC bottleneck (WAN bandwidth, node load in the lagging DC).
If hints are being dropped past `max_hint_window_in_ms`, run a cross-DC `nodetool repair`
to reconcile. Tune `max_hint_window_in_ms` for the outages you expect to cover with hinted
handoff.

---

## 10. A DC is unexpectedly rejecting writes

**Symptoms:** a DC you expect to serve writes is rejecting them.

**Diagnose:**

```bash
# Is this DC the write-routed one, and is it healthy?
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} writable:{.writable} up:{.upNormalNodes}/{.totalNodes}{"\n"}{end}'
# What DC does the Lease route to?
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Ring health from a node:
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status
```

**Causes & action:**

- **Writing at `EACH_QUORUM` while another DC is down** `EACH_QUORUM` needs a quorum in
  every DC, so it fails when any DC is down. This is by design. Use `LOCAL_QUORUM` for
  DC-loss tolerance, or restore the down DC.
- **Not enough local replicas for `LOCAL_QUORUM`** if too many nodes in the local DC are
  down, `LOCAL_QUORUM` cannot be met. Recover the DC's nodes (`nodetool status` shows
  DN).
- **Not the write-routed DC** applications should use the endpoint `<db>`, which routes to
  the active DC; writing directly to a standby DC is legitimate for Cassandra but changes
  your single-writer posture.

Cassandra will not block a write for lack of a cross-DC vote; a write failure here is a
local-replica or consistency-level issue, not a fence.

---

## 11. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read/renewed across the spokes.

**Automatic:** Cassandra keeps running on its own, so the ring stays writable in every DC
and the endpoint stays where it last resolved; the Lease is routing, not the failover
mechanism, so its loss does not stop writes. What you lose is **endpoint steering and
planned switchover**: the orchestrator cannot move the active DC until the Lease quorum
returns.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc   # error / stale
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status          # ring still serving
```

**Action:** restore the `dr-controlplane` etcd quorum (in the even layout its third member
is in the Arbiter DC). Once the Lease is renewable, endpoint steering and switchover
resume.

---

## 12. Suspected data divergence after a partition

Cassandra is AP: during a partition, both sides can accept `LOCAL_QUORUM` writes, so
divergence is expected and reconciled after the fact, not prevented. This is not a bug and
not the same as ClickHouse or Postgres split-brain (there is no committed-log fork to
repair, only cell values reconciled by timestamp).

**Diagnose:**

```bash
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool netstats
kubectl get cassandra -n demo cas-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} hints={.pendingHints} repair={.repairBacklog}{"\n"}{end}'
```

**Action:** let hinted handoff drain, then run a **full cross-DC `nodetool repair`** to
reconcile all replicas; the newest cell wins by timestamp. For data that must never
diverge across a partition, write it at `EACH_QUORUM` (it will not ack during a partition)
and read it at a consistency level that spans DCs.

```bash
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool repair --full
```

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get cassandra -n demo cas-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl get pods -n demo -l app.kubernetes.io/instance=cas-dcdr -L open-cluster-management.io/cluster-name -o wide
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool status
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool netstats
kubectl exec -n demo cas-dcdr-rack-a-0 -- nodetool gossipinfo
```
