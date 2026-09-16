---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: kf-dr-runbook-kafka
    name: Runbook
    parent: kf-dr-kafka
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Kafka DC-DR Runbook

Scenario-by-scenario procedures for operating a Kafka workload in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB does
**automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/kafka/dr/guide/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane
kubeconfig, and `kf-dcdr`/`demo` are the example database and namespace.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the DC the coordination plane makes the active write cluster):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC brokers, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=kf-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# MM2 connector status on the standby's ConnectCluster:
kubectl get connector -n demo -l app.kubernetes.io/instance=kf-dcdr
```

Golden rules:

- **The Lease decides the active write cluster.** Exactly one DC is `writable: true`
  in `status.disasterRecovery` at any instant.
- **The produce fence fails closed.** A cluster that cannot confirm it holds the Lease
  denies producer writes, so a partitioned old-active cluster stops taking writes on
  its own.
- **MM2 is asynchronous.** An unplanned failover loses the un-mirrored tail (bounded by
  MM2 lag). Only a drained planned switchover is zero record loss.
- **Never enable both MM2 directions at once** and never fence the MM2 connector
  principal or `super.users`.

---

## 1. Intra-DC broker loss (single broker in a DC fails)

**Symptoms:** one broker pod in a DC is down; some partitions briefly re-elect a
leader within that cluster.

**Automatic:** the loss is handled entirely inside that DC's Kafka cluster. KRaft
re-elects partition leaders among the surviving in-sync replicas (ISR), and the pod
reschedules. There is **no cross-DC effect**: the active DC stays active, the standby
stays a mirror, and the Lease does not move.

**Verify:**

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=kf-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # unchanged
```

**Action:** none required. Ensure the failed broker rejoins its cluster's ISR. Provision
user topics with a replication factor and `min.insync.replicas` that tolerate one broker
loss inside a DC.

---

## 2. Full active-DC loss (zone/cluster failure)

**Symptoms:** the active DC's brokers are gone/unreachable; producers to the bootstrap
endpoint fail briefly.

**Automatic:** the standby is already a near-current MM2 mirror. The `dr-controlplane`
Lease moves to the standby, and the orchestrator flips the bootstrap endpoint to the
standby's brokers, opens the standby's produce fence, and reverses the MM2 direction
(enabling the survivor-to-old-active connectors for when the old DC returns). `phase`
moves `FailingOver` to `Steady` and the survivor becomes `writable: true`.

**Verify:**

```bash
kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get pods -n demo -l app.kubernetes.io/instance=kf-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
```

**Action:** none required for availability. Note the RPO: records the old active
accepted but had not yet mirrored are lost (bounded by the MM2 lag at the moment of
loss). There is no rewind. When the failed DC returns, see scenario 6 (failback).

---

## 3. Clean DC-vs-DR partition (both DCs up, cannot reach each other)

**Symptoms:** the two data DCs are up but the network between them is cut. MM2 lag on
the standby climbs; the old active may still see local producers.

**Automatic:** the produce fence is the split-brain guard. The old active loses its
Lease renewal across the partition and its fence **fails closed**, so it stops
accepting producer writes on its own **before** the hub reacts. The etcd majority (two
sites plus the Arbiter DC) keeps or grants the Lease to one side only, so exactly one
cluster is ever writable and the two logs cannot diverge.

**Verify there is exactly one writable DC:**

```bash
kubectl get kafka -n demo kf-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. Once the partition clears, the non-active cluster resumes
as the MM2 target and catches up. If both sides took writes before the fence closed
(should not happen with a fail-closed fence), treat the non-active side's forked tail
as scenario 6 (re-seed the affected topics).

---

## 4. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate kafka -n demo kf-dcdr dr.kubedb.com/switchover-to=dc-b
```

**Automatic:** the hub gates on the target's health and MM2 lag budget, then quiesces
producers by closing the active cluster's produce fence, waits for MM2 to drain to
near-zero lag (so the target holds every record), flips the bootstrap endpoint to the
target, opens the target's fence, reverses the mirror direction, and moves the Lease.
Because MM2 fully drained before the flip, **zero committed records are lost**.

**Verify:**

```bash
kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-b
kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 8 (MM2 lag high / switchover stuck).

---

## 5. Consumers not resuming after a flip

**Symptoms:** after a failover or switchover, a consumer group re-reads from the
beginning or skips records on the new active cluster.

**Cause:** consumer offsets were not translated into the new active cluster's offset
space, or the group reconnected before the `MirrorCheckpointConnector` emitted a
checkpoint.

**Diagnose:**

```bash
# Is the checkpoint connector running on the (now active) side and syncing group offsets?
kubectl get connector -n demo -l app.kubernetes.io/instance=kf-dcdr
# Confirm sync.group.offsets.enabled=true in the checkpoint connector config secret.
```

**Action:** ensure the `MirrorCheckpointConnector` runs with
`sync.group.offsets.enabled=true` so it translates group offsets. If a group flipped
before a checkpoint was available, reset it to the translated offset. Do not disable
the checkpoint connector: it is what lets consumers resume across a flip.

---

## 6. Failback (return a recovered DC to active)

**Symptoms:** the previously lost DC is back but holds a forked tail (records it
accepted before the failover that were never mirrored).

**Automatic:** the returned cluster becomes the MM2 target of the new active. But MM2
only adds and never deletes, so a naive re-mirror leaves the orphan forked tail on top
of the new active's data.

**Action:**

1. **Re-seed** the affected topics from the new active: wipe the returned cluster's
   copy and re-mirror from scratch, so the forked tail is removed. Or **accept and
   document** the orphan tail as bounded loss.
2. Once the returned DC is caught up (low MM2 lag), perform a **drained planned
   switchover** back to it:

   ```bash
   kubectl annotate kafka -n demo kf-dcdr dr.kubedb.com/switchover-to=dc-a
   ```

There is no rewind; Kafka cannot roll back a log, so the re-seed is what restores
consistency.

---

## 7. A standby DC is lost

**Symptoms:** the non-active DC's brokers are gone; that DC shows `healthy: false` and
MM2 lag is unavailable.

**Impact:** none on writes. The active DC keeps taking producer writes; you lose the
DR copy until the standby returns.

**Verify the active DC is still writable:**

```bash
kubectl get kafka -n demo kf-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** recover the standby cluster's brokers. When it returns, MM2 resumes
mirroring from the active and the standby catches up. You are running without DR
protection until then.

---

## 8. MM2 lag high (mirror falling behind)

**Symptoms:** `mirrorLagMillis` on the standby climbs; a planned switchover stays in
`FailingOver` because MM2 has not drained.

**Diagnose:**

```bash
# Per-DC MM2 lag and health:
kubectl get kafka -n demo kf-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.mirrorLagMillis}ms healthy={.healthy}{"\n"}{end}'
# MM2 connector status:
kubectl get connector -n demo -l app.kubernetes.io/instance=kf-dcdr
```

**Causes & action:**

- **Cross-DC network bottleneck or producer burst:** relieve the bottleneck (network,
  active-cluster write load) so MM2 drains. A planned switchover intentionally waits for
  the lag to reach near-zero after quiescing producers.
- **A mirror connector is failed/paused:** check the `Connector` status and its config
  secret; restart the failed connector on the standby's `ConnectCluster`.
- **Abort a stuck switchover:** remove the annotation to cancel:
  `kubectl annotate kafka -n demo kf-dcdr dr.kubedb.com/switchover-to-`. The active DC's
  produce fence reopens and it stays active.

---

## 9. Arbiter DC lost

**Symptoms:** the Arbiter DC is gone; its `dr-controlplane` etcd member is unreachable.

**Impact:** none on writes. The two data DCs plus the lost arbiter leave the etcd
quorum with two of three members, still a majority, so the Lease can still be
renewed and the active DC keeps taking writes. You lose the tie-break, so a subsequent
**second** failure can no longer keep an etcd majority.

**Verify the cluster is still writable:**

```bash
kubectl get kafka -n demo kf-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** restore the Arbiter DC's etcd member to regain single-fault tolerance. The
Arbiter DC never runs Kafka, so no broker recovery is involved.

---

## 10. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read or renewed across the spokes.

**Automatic:** the current active cluster keeps taking writes as long as its fence can
still confirm the Lease it last held; but if the fence cannot confirm the Lease it
**fails closed** and denies produce, so a total coordination-plane outage can make the
active cluster go read-only. What you always lose is **failover and planned
switchover**: the hub cannot move the active DC until the etcd quorum returns.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc   # error / stale
kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery.phase}'   # may be Degraded
```

**Action:** restore the `dr-controlplane` etcd quorum (its third member lives in the
Arbiter DC). Once the Lease is renewable, the fence reopens on the active cluster and
failover/switchover resume.

---

## 11. Which DC is active?

**Question:** confirm which cluster is taking writes right now.

```bash
# The DR status view (authoritative for what the hub applied):
kubectl get kafka -n demo kf-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'
# The Lease holder (what the coordination plane intends):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Where the bootstrap endpoint resolves and which DC is writable:
kubectl get kafka -n demo kf-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

In steady state the `activeDC`, the Lease `holderIdentity`, and the single
`writable: true` DC all name the same data center. A brief mismatch during
`FailingOver` is expected; it should converge back to `Steady`.

---

## 12. Suspected split writes (two clusters taking writes)

This should be impossible: the etcd majority grants the Lease to one DC only, and the
produce fence on any non-Lease-holder fails closed. If `status.disasterRecovery` ever
shows two `writable: true` DCs, or producers succeed against both bootstraps:

**Diagnose immediately:**

```bash
kubectl get kafka -n demo kf-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o yaml
kubectl get connector -n demo -l app.kubernetes.io/instance=kf-dcdr   # both directions must not be enabled
```

**Action:** confirm the produce fence is engaged on the non-Lease-holder (ACL revoked
for client principals, or the client listener gated) and that only one MM2 direction's
connectors are enabled. Never enable both directions at once: with
`IdentityReplicationPolicy` there is no topic-rename loop guard, so overlapping mirrors
ping-pong a topic between clusters. Disable the wrong-direction connectors, restore the
fence on the non-active cluster, then treat any forked tail as scenario 6 (re-seed).

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get kafka -n demo kf-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl get pods -n demo -l app.kubernetes.io/instance=kf-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name -o wide
kubectl get connectcluster,connector -n demo -l app.kubernetes.io/instance=kf-dcdr -o yaml
```
