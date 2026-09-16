---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: es-dr-runbook-elasticsearch
    name: Runbook
    parent: es-dr-elasticsearch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Elasticsearch DC-DR Runbook

Scenario-by-scenario procedures for operating an Elasticsearch workload in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB does
**automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/elasticsearch/dr/guide/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane kubeconfig,
and `es-dcdr`/`demo` are the example database and namespace.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the DC the coordination plane makes the active write cluster):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC nodes, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=es-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# CCR follow lag per DC from status:
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.followLagOps}{"\n"}{end}'
```

Golden rules:

- **The Lease decides the active write cluster.** Exactly one DC is `writable: true` in
  `status.disasterRecovery` at any instant.
- **The follower-read-only fence fails closed.** A cluster that cannot confirm it holds the
  Lease keeps its indices as read-only followers and never self-promotes, so a partitioned
  old-active cluster stops taking writes on its own.
- **CCR is asynchronous.** An unplanned failover loses the un-followed tail (bounded by CCR
  follow lag). Only a drained planned switchover is zero document loss.
- **Never run CCR both directions for the same index at once**, and never promote followers
  without the Lease.

---

## 1. Intra-DC node loss (single node in a DC fails)

**Symptoms:** one Elasticsearch node pod in a DC is down; some shards briefly re-allocate or
re-elect a primary within that cluster.

**Automatic:** the loss is handled entirely inside that DC's Elasticsearch cluster. The
master quorum re-allocates shards and promotes surviving replica shards to primary, and the
pod reschedules. There is **no cross-DC effect**: the active DC stays active, the standby
stays a follower, and the Lease does not move.

**Verify:**

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=es-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # unchanged
```

**Action:** none required. Ensure the failed node rejoins its cluster. Provision user indices
with a replica count that tolerates one node loss inside a DC (so shards stay green).

---

## 2. Full active-DC loss (zone/cluster failure)

**Symptoms:** the active DC's nodes are gone/unreachable; writes to the search/index endpoint
fail briefly.

**Automatic:** the standby is already a near-current CCR follower. The `dr-controlplane` Lease
moves to the standby, and the orchestrator pauses and promotes the standby's follower indices
(`pause_follow`, `unfollow`, convert to writable), flips the search/index endpoint to the
standby's nodes, and starts CCR in the reverse direction (auto-follow on the old active for
when it returns). `phase` moves `FailingOver` to `Steady` and the survivor becomes
`writable: true`.

**Verify:**

```bash
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get pods -n demo -l app.kubernetes.io/instance=es-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
```

**Action:** none required for availability. Note the RPO: operations the old active accepted
but had not yet followed are lost (bounded by the CCR follow lag at the moment of loss). There
is no rewind. When the failed DC returns, see scenario 6 (failback).

---

## 3. Clean DC-vs-DR partition (both DCs up, cannot reach each other)

**Symptoms:** the two data DCs are up but the network between them is cut. CCR follow lag on
the standby climbs; the old active may still see local clients.

**Automatic:** the follower-read-only fence is the split-brain guard. The old active loses its
Lease renewal across the partition; because promotion requires the Lease, the standby cluster
**stays read-only-follower** and never self-promotes, and the old active keeps its role only
while it can confirm the Lease. The etcd majority (two sites plus the Arbiter DC) keeps or
grants the Lease to one side only, so exactly one cluster is ever writable and the two
clusters cannot diverge.

**Verify there is exactly one writable DC:**

```bash
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. Once the partition clears, the non-active cluster resumes as the
CCR follower and catches up. If the follower fell past the leader's soft-delete retention it
forces a full re-follow (scenario 8). If both sides somehow took writes (should not happen with
a fail-closed fence), treat the non-active side's forked tail as scenario 6 (re-seed the
affected indices).

---

## 4. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate elasticsearch -n demo es-dcdr dr.kubedb.com/switchover-to=dc-b
```

**Automatic:** the hub gates on the target's health and CCR follow-lag budget, then quiesces
indexing on the active cluster, waits for CCR to drain to zero follow lag (so the target holds
every operation), pauses and promotes the target's follower indices, flips the search/index
endpoint to the target, starts CCR in the reverse direction, and moves the Lease. Because CCR
fully drained before the flip, **zero acknowledged documents are lost**.

**Verify:**

```bash
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-b
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 8 (CCR follow lag high / switchover stuck).

---

## 5. Indices not writable after a flip

**Symptoms:** after a failover or switchover, writes to the new active cluster are rejected
because target indices are still follower (read-only) indices.

**Cause:** the follower indices were not promoted (`pause_follow`, `unfollow`, convert to
writable), or promotion partially completed.

**Diagnose:**

```bash
# Are the target's indices still followers?
curl -k -u "admin:$PASSWORD" "https://es-dcdr-dc-b.demo.svc:9200/_ccr/stats" | jq
# Which DC does status report as writable?
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** confirm the Lease actually moved to the target (`activeDC` matches the Lease
`holderIdentity`). If the Lease moved but promotion did not finish, the hub retries on the next
reconcile; if it is stuck, ensure the target's follower indices are paused, unfollowed, and
converted to regular writable indices. Never promote a cluster that does not hold the Lease.

---

## 6. Failback (return a recovered DC to active)

**Symptoms:** the previously lost DC is back but holds a forked tail (operations it accepted
before the failover that were never followed).

**Automatic:** the returned cluster becomes the CCR follower of the new active via auto-follow
and catches up. But the forked tail (operations that were never followed) sits on top of the
returned cluster's copy and Elasticsearch cannot rewind it.

**Action:**

1. **Reconcile or re-seed** the affected indices from the new active: delete the returned
   cluster's copy of an affected index so auto-follow re-seeds it from scratch, removing the
   forked tail. Or **accept and document** the forked tail as bounded loss.
2. Once the returned DC is caught up (low follow lag), perform a **drained planned switchover**
   back to it:

   ```bash
   kubectl annotate elasticsearch -n demo es-dcdr dr.kubedb.com/switchover-to=dc-a
   ```

There is no rewind; Elasticsearch cannot roll back an index, so the re-seed is what restores
consistency.

---

## 7. A standby DC is lost

**Symptoms:** the non-active DC's nodes are gone; that DC shows `healthy: false` and CCR follow
lag is unavailable.

**Impact:** none on writes. The active DC keeps taking client writes; you lose the DR copy
until the standby returns.

**Verify the active DC is still writable:**

```bash
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** recover the standby cluster's nodes. When it returns, auto-follow resumes following
from the active and the standby catches up. If it fell past the leader's soft-delete retention,
the affected follower indices force a full re-follow. You are running without DR protection
until then.

---

## 8. CCR follow lag high (follower falling behind)

**Symptoms:** `followLagOps` on the standby climbs; a planned switchover stays in `FailingOver`
because CCR has not drained.

**Diagnose:**

```bash
# Per-DC follow lag and health:
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.followLagOps} healthy={.healthy}{"\n"}{end}'
# Raw CCR follow-stats on the standby cluster:
curl -k -u "admin:$PASSWORD" "https://es-dcdr-dc-b.demo.svc:9200/_ccr/stats" | jq
```

**Causes & action:**

- **Cross-DC network bottleneck or indexing burst:** relieve the bottleneck (network,
  active-cluster write load) so CCR drains. A planned switchover intentionally waits for the
  lag to reach zero after quiescing indexing.
- **A follower past the leader's soft-delete retention:** the follower can no longer resume
  from the retained history and forces a full re-follow. Raise
  `index.soft_deletes.retention_lease.period` on the leader indices to bound catch-up, or
  re-seed the affected follower indices.
- **Abort a stuck switchover:** remove the annotation to cancel:
  `kubectl annotate elasticsearch -n demo es-dcdr dr.kubedb.com/switchover-to-`. Indexing on
  the active DC resumes and it stays active.

---

## 9. Arbiter DC lost

**Symptoms:** the Arbiter DC is gone; its `dr-controlplane` etcd member is unreachable.

**Impact:** none on writes. The two data DCs plus the lost arbiter leave the etcd quorum with
two of three members, still a majority, so the Lease can still be renewed and the active DC
keeps taking writes. You lose the tie-break, so a subsequent **second** failure can no longer
keep an etcd majority.

**Verify the cluster is still writable:**

```bash
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** restore the Arbiter DC's etcd member to regain single-fault tolerance. The Arbiter
DC never runs Elasticsearch, so no node recovery is involved.

---

## 10. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read or renewed across the spokes.

**Automatic:** the current active cluster keeps taking writes as long as it can still confirm
the Lease it last held; but a cluster that cannot confirm the Lease **fails closed** and keeps
its indices read-only, so a total coordination-plane outage can make the active cluster go
read-only. What you always lose is **failover and planned switchover**: the hub cannot move the
active DC until the etcd quorum returns.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc   # error / stale
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery.phase}'   # may be Degraded
```

**Action:** restore the `dr-controlplane` etcd quorum (its third member lives in the Arbiter
DC). Once the Lease is renewable, the active cluster keeps its writable indices and
failover/switchover resume.

---

## 11. Which DC is active?

**Question:** confirm which cluster is taking writes right now.

```bash
# The DR status view (authoritative for what the hub applied):
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'
# The Lease holder (what the coordination plane intends):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Where the search/index endpoint resolves and which DC is writable:
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

In steady state the `activeDC`, the Lease `holderIdentity`, and the single `writable: true` DC
all name the same data center. A brief mismatch during `FailingOver` is expected; it should
converge back to `Steady`.

---

## 12. Suspected split writes (two clusters taking writes)

This should be impossible: the etcd majority grants the Lease to one DC only, and the
follower-read-only fence keeps any non-Lease-holder's indices read-only. If
`status.disasterRecovery` ever shows two `writable: true` DCs, or writes succeed against both
endpoints:

**Diagnose immediately:**

```bash
kubectl get elasticsearch -n demo es-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o yaml
# Neither cluster should have the other's indices as leaders while its own are being followed:
curl -k -u "admin:$PASSWORD" "https://es-dcdr-dc-a.demo.svc:9200/_ccr/stats" | jq
curl -k -u "admin:$PASSWORD" "https://es-dcdr-dc-b.demo.svc:9200/_ccr/stats" | jq
```

**Action:** confirm the non-Lease-holder's indices are follower (read-only) indices and that
only one CCR direction is enabled. Never run CCR both directions for the same index at once: an
index that is both a leader and a follower would ping-pong operations between clusters. Demote
the wrong side back to followers (re-follow from the Lease holder), restore the fence, then
treat any forked tail as scenario 6 (re-seed).

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get elasticsearch -n demo es-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl get pods -n demo -l app.kubernetes.io/instance=es-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name -o wide
curl -k -u "admin:$PASSWORD" "https://es-dcdr-dc-a.demo.svc:9200/_ccr/stats" | jq
curl -k -u "admin:$PASSWORD" "https://es-dcdr-dc-b.demo.svc:9200/_ccr/stats" | jq
```
