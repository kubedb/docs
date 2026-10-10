---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: ch-dr-runbook-clickhouse
    name: Runbook
    parent: ch-dr-clickhouse
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# ClickHouse DC-DR Runbook

Scenario-by-scenario procedures for operating a ClickHouse cluster in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB and
ClickHouse do **automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/clickhouse/dr/guide/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane
kubeconfig, and `ch-dcdr`/`demo` are the example database and namespace. The example pod
`ch-dcdr-appscode-cluster-shard-0-0` is the first replica of shard 0.

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the DC the coordination plane routes writes to):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC replicas and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=ch-dcdr -L open-cluster-management.io/cluster-name

# Replication delay, queue, and log pointers (from any replica):
kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client \
  --query "SELECT database, table, absolute_delay, queue_size, log_pointer, log_max_index, total_replicas, active_replicas FROM system.replicas"
```

Golden rules:

- **ClickHouse Keeper quorum decides who can commit.** Never try to force a write into a
  DC that has lost Keeper quorum; it cannot register parts and its inserts fail by design.
  That is the split-brain guarantee, not a bug.
- **There is no promotion.** Every replica is writable. DR is a routing change (the write
  endpoint follows the Lease to a DC that holds Keeper quorum), not an election.
- **Exactly one DC is `writable: true`** in `status.disasterRecovery` at any instant (the
  write-routed active DC), even though the engine would let more than one accept writes.
- **The Lease is routing, not safety.** If the Lease is stale, ClickHouse keeps running on
  its own Keeper quorum; what you lose is switchover and endpoint steering.

---

## 1. Active DC lost (zone/cluster failure)

**Symptoms:** the active DC's replicas are gone/unreachable; writes to the endpoint fail
briefly until it re-points.

**Automatic:** the surviving DCs that still hold Keeper quorum (a standby data DC plus the
Arbiter DC in the even layout, or the surviving data majority in the odd layout) **keep
accepting writes on their own**, because Keeper quorum survives and every replica is
already writable. There is no promotion. The orchestrator observes the Lease move to a
surviving DC and points the single write endpoint there. `phase` moves `FailingOver` to
`Steady`.

**Verify:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get pods -n demo -l app.kubernetes.io/instance=ch-dcdr -L open-cluster-management.io/cluster-name
```

**Action:** none required for availability. Note the RPO: only committed-but-unfetched
parts on the lost DC's disk are at risk (a clean partition that put the lost DC in the
Keeper minority loses zero committed data). When the failed DC returns, see scenario 8
(re-add a DC).

---

## 2. Network partition between data centers

**Symptoms:** DCs are up but cannot reach each other.

**Automatic:** the side that keeps Keeper quorum keeps registering parts and stays
writable. The minority side **loses Keeper quorum and cannot register parts, so its
inserts fail on their own**, at the engine level. There is no split brain and the fence
needs no action: a minority DC simply cannot commit. The write endpoint stays on (or moves
to) the majority side.

**Verify there is exactly one writable DC and check who holds Keeper quorum:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}=writable:{.writable},quorum:{.keeperQuorum} {end}'
```

**Action:** heal the network. The minority side rejoins the Keeper ensemble and catches up
through `ReplicatedMergeTree` automatically. There is no rewind, because the minority
committed nothing.

---

## 3. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate clickhouse -n demo ch-dcdr dr.kubedb.com/switchover-to=dc-b
```

**Automatic:** the hub gates on the target's health and lag, quiesces writes on the current
active DC (routes clients away), waits until the target's replicas show `absolute_delay`
near zero and an empty queue (`queue_size: 0`), then moves the Lease and the write endpoint
to `dc-b`. Because it waits for the target to catch up before flipping, near-zero committed
writes are lost. There is no promotion step.

**Verify:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-b
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 6 (switchover stuck).

---

## 4. Planned failback to the original DC

After the original DC is healthy and caught up (failback is native: its replicas rejoin the
Keeper ensemble and fetch missing parts, with no rewind), steer the active DC back:

```bash
kubectl annotate clickhouse -n demo ch-dcdr dr.kubedb.com/switchover-to=dc-a
```

Same near-zero-RPO flow as scenario 3. There is no `pg_rewind` step and no rollback; a DC
that lacked Keeper quorum committed nothing to diverge.

---

## 5. Arbiter DC lost (even layout)

**Symptoms:** the Arbiter DC is gone; its etcd member and the data-less Keeper voter are
unreachable.

**Impact:** none on writes. The two data DCs together still hold 2 of the 3 Keeper voters,
a majority, so Keeper quorum holds and writes continue. You lose the tie-break voter, so a
subsequent **second** failure (a data DC) can no longer keep Keeper quorum automatically.

**Verify the cluster is still writable and holds quorum:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} quorum:{.keeperQuorum} writable:{.writable}{"\n"}{end}'
```

**Action:** restore the Arbiter DC (the etcd member and the data-less Keeper voter) to
regain single-fault tolerance.

---

## 6. Planned switchover stuck (target not catching up)

**Symptoms:** after annotating `switchover-to`, `phase` stays `FailingOver` and the endpoint
does not move.

**Diagnose:**

```bash
# Target lag and health from status:
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} delay={.absoluteDelaySeconds} queue={.queueSize} healthy={.healthy}{"\n"}{end}'
# Replication state from a target replica:
kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client \
  --query "SELECT table, absolute_delay, queue_size, log_pointer, log_max_index FROM system.replicas"
```

**Causes & action:**

- **Target lag not converging** the switchover waits for the target's `absolute_delay` near
  zero and an empty queue before flipping. Relieve the cross-DC bottleneck (network, insert
  load, Keeper round-trip latency) so the target drains its replication queue.
- **Target unhealthy** ensure the target DC has ready replicas of every shard.
- **Abort** remove the annotation to cancel:
  `kubectl annotate clickhouse -n demo ch-dcdr dr.kubedb.com/switchover-to-`.

---

## 7. A standby DC is lost

**Symptoms:** a non-active DC's replicas are gone; that DC shows `healthy: false`.

**Impact:** none on writes as long as the remaining DCs keep Keeper quorum (the active DC
plus the Arbiter DC in the even layout). You lose that DC's redundancy and its local read
capacity until it returns.

**Verify the active DC still holds quorum and is writable:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName} quorum:{.keeperQuorum}{end}'
```

**Action:** recover the DC's replicas; they reschedule and catch up through
`ReplicatedMergeTree` over Keeper automatically.

---

## 8. Re-add / recover a previously lost data center

After a DC returns from a failure:

**Automatic:** its replicas rejoin the Keeper ensemble and catch up over native
`ReplicatedMergeTree` (they fetch the missing parts). There is no rewind and no rollback; a
DC that lacked Keeper quorum committed nothing to diverge.

**Verify:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} healthy={.healthy} delay={.absoluteDelaySeconds} queue={.queueSize}{"\n"}{end}'
```

**Action:** to make it the active DC again, perform a planned failback (scenario 4) once its
`absoluteDelaySeconds` is near zero and its queue is empty.

---

## 9. Keeper quorum lost across the ensemble (double failure)

**Symptoms:** more than one Keeper voter is unreachable at once, so the ensemble has no Raft
majority; inserts fail cluster-wide.

**Impact:** with the ensemble below majority, **no DC can register parts**, so writes stop
everywhere. This is the engine protecting against split brain, not a KubeDB fault.

**Verify:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} keeperVoter:{.keeperVoter} keeperQuorum:{.keeperQuorum}{"\n"}{end}'
```

**Action:** restore enough Keeper voters to regain a majority (bring back a failed Member
DC's voter or the Arbiter DC's voter). Writes resume automatically once the ensemble has
quorum again. Do not try to force a single surviving voter to act alone.

---

## 10. A DC is unexpectedly read-only / rejecting writes

**Symptoms:** a DC you expect to serve the endpoint is rejecting inserts.

**Diagnose:**

```bash
# Does this DC hold Keeper quorum, and is it the write-routed DC?
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} writable:{.writable} quorum:{.keeperQuorum}{"\n"}{end}'
# What DC does the Lease route to?
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Replication and Keeper reachability from a replica:
kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client \
  --query "SELECT table, is_readonly, is_session_expired, absolute_delay FROM system.replicas"
```

**Causes & action:**

- **Not the write-routed DC** the endpoint intentionally routes writes elsewhere; this DC
  is a standby (correct). Point writes at the endpoint `<db>`, not at this DC directly.
- **Lost Keeper quorum** `keeperQuorum:false` or `is_readonly:1` means this DC cannot reach
  a Keeper majority, so it is read-only by design. See scenario 2 or 9.
- **Lease routes here but writes still fail** check Keeper reachability and the fence
  marker; the endpoint fails closed if the marker is stale.

Never try to bypass the fence or write directly to a DC that lacks Keeper quorum; it cannot
commit and you risk confusing clients.

---

## 11. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read/renewed across the spokes.

**Automatic:** ClickHouse keeps running on its own Keeper quorum, so the cluster stays
writable in whichever DC the endpoint last resolved to; the Lease is routing, not the
failover mechanism, so its loss does not by itself stop writes. What you lose is **endpoint
steering and planned switchover**: the orchestrator cannot move the active DC until the
Lease quorum returns.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc   # error / stale
kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client --query "SELECT 1"  # DB still serving
```

**Action:** restore the `dr-controlplane` etcd quorum (it shares the Arbiter DC with the
data-less Keeper voter). Once the Lease is renewable, endpoint steering and switchover
resume.

---

## 12. Suspected split-brain (two DCs taking committed writes)

This should be impossible with a spread 3-site Keeper: no single data DC holds a Keeper
majority, and a minority DC cannot register parts, so it cannot commit. If
`status.disasterRecovery` ever shows two `writable: true` DCs:

**Diagnose immediately:**

```bash
kubectl get clickhouse -n demo ch-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} writable:{.writable} quorum:{.keeperQuorum}{"\n"}{end}'
kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client \
  --query "SELECT table, is_readonly, total_replicas, active_replicas FROM system.replicas"
```

**Action:** confirm the Keeper ensemble still spreads voters 3-site (no single data DC was
given a Keeper majority by a bad topology change). Because a minority DC cannot register
parts, it cannot diverge committed data even if the routing briefly shows two writable DCs.
Restore connectivity, let the endpoint settle on one active DC, and correct the Keeper
placement if it drifted.

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get clickhouse -n demo ch-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl get pods -n demo -l app.kubernetes.io/instance=ch-dcdr -L open-cluster-management.io/cluster-name -o wide
kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client --query "SELECT * FROM system.replicas FORMAT Vertical"
kubectl exec -n demo ch-dcdr-appscode-cluster-shard-0-0 -- clickhouse-client --query "SELECT * FROM system.clusters FORMAT Vertical"
```
