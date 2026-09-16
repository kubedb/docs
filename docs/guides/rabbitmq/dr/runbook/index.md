---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: rm-dr-runbook-rabbitmq
    name: Runbook
    parent: rm-dr-rabbitmq
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# RabbitMQ DC-DR Runbook

Scenario-by-scenario procedures for operating a RabbitMQ workload in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB does
**automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/rabbitmq/dr/guide/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane
kubeconfig, and `rm-dcdr`/`demo` are the example database and namespace.

> **New to KubeDB?** Please start [here](/docs/README.md).

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the DC the coordination plane makes the active publish cluster):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC nodes, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=rm-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# Federation link status on the standby cluster:
kubectl exec -n demo rm-dcdr-dc-b-0 -- rabbitmqctl list_federation_links
```

Golden rules:

- **The Lease decides the active publish cluster.** Exactly one DC is `writable: true`
  in `status.disasterRecovery` at any instant.
- **The publish fence fails closed.** A cluster that cannot confirm it holds the Lease
  denies client publishes, so a partitioned old-active cluster stops taking publishes on
  its own.
- **Federation is asynchronous.** An unplanned failover loses the un-federated tail
  (bounded by federation lag). Only a drained planned switchover is zero message loss.
- **Never enable both federation directions at once** and never fence the federation
  user or the management user.

---

## 1. Intra-DC node loss (single node in a DC fails)

**Symptoms:** one RabbitMQ node pod in a DC is down; some quorum queues briefly
re-elect a leader within that cluster.

**Automatic:** the loss is handled entirely inside that DC's RabbitMQ cluster. Each
affected quorum queue's Raft group re-elects a leader among the surviving replicas, and
the pod reschedules. There is **no cross-DC effect**: the active DC stays active, the
standby stays a federation target, and the Lease does not move.

**Verify:**

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=rm-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # unchanged
```

**Action:** none required. Ensure the failed node rejoins its cluster. Use quorum queues
(not classic queues) so intra-DC HA tolerates one node loss; a three-node DC cluster
keeps quorum with one node down.

---

## 2. Full active-DC loss (zone/cluster failure)

**Symptoms:** the active DC's nodes are gone/unreachable; publishes to the AMQP endpoint
fail briefly.

**Automatic:** the standby is already a near-current Federation replica. The
`dr-controlplane` Lease moves to the standby, and the orchestrator flips the AMQP
endpoint to the standby's nodes, opens the standby's publish fence, and reverses the
federation direction (setting up the survivor-to-old-active upstream for when the old DC
returns). `phase` moves `FailingOver` to `Steady` and the survivor becomes
`writable: true`.

**Verify:**

```bash
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get pods -n demo -l app.kubernetes.io/instance=rm-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
```

**Action:** none required for availability. Note the RPO: messages the old active
accepted but had not yet federated are lost (bounded by the federation lag at the moment
of loss). There is no rewind. When the failed DC returns, see scenario 6 (failback).

---

## 3. Clean DC-vs-DR partition (both DCs up, cannot reach each other)

**Symptoms:** the two data DCs are up but the network between them is cut. Federation lag
on the standby climbs; the old active may still see local publishers.

**Automatic:** the publish fence is the split-brain guard. The old active loses its Lease
renewal across the partition and its fence **fails closed**, so it stops accepting client
publishes on its own **before** the hub reacts. The etcd majority (two sites plus the
Arbiter DC) keeps or grants the Lease to one side only, so exactly one cluster is ever
writable and the two clusters cannot diverge.

**Verify there is exactly one writable DC:**

```bash
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. Once the partition clears, the non-active cluster resumes
as the Federation target and catches up. If both sides took publishes before the fence
closed (should not happen with a fail-closed fence), treat the non-active side's forked
tail as scenario 6 (re-seed the affected queues).

---

## 4. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate rabbitmq -n demo rm-dcdr dr.kubedb.com/switchover-to=dc-b
```

**Automatic:** the hub gates on the target's health and federation lag budget, then
quiesces publishers by closing the active cluster's publish fence, waits for Federation
to drain to near-zero lag (so the target holds every message), flips the AMQP endpoint to
the target, opens the target's fence, reverses the federation direction, and moves the
Lease. Because Federation fully drained before the flip, **zero confirmed messages are
lost**.

**Verify:**

```bash
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-b
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 8 (federation lag high / switchover stuck).

---

## 5. Consumers not resuming after a flip

**Symptoms:** after a failover or switchover, a consumer re-reads messages it already
processed, or misses messages, on the new active cluster.

**Cause:** Federation is asynchronous and does not deduplicate across the flip. The
consumer either reconnected before the un-federated tail arrived, or is re-seeing the
redelivery window from the federated backlog.

**Diagnose:**

```bash
# Is the federation link up and draining on the (now active) side?
kubectl exec -n demo rm-dcdr-dc-b-0 -- rabbitmqctl list_federation_links
# Per-DC federation lag:
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.federationLagMessages}{"\n"}{end}'
```

**Action:** make consumers idempotent, or apply a dedup window across the flip so a
redelivered message is a no-op. Do not treat a small redelivery window as data loss: it
is the expected cost of asynchronous federation. For a zero-loss move, use a drained
planned switchover (scenario 4) rather than an unplanned failover.

---

## 6. Failback (return a recovered DC to active)

**Symptoms:** the previously lost DC is back but holds a forked tail (messages it
accepted before the failover that were never federated).

**Automatic:** the returned cluster becomes the Federation target of the new active. But
Federation only adds and never deletes, so a naive re-federation leaves the orphan forked
tail on top of the new active's data.

**Action:**

1. **Re-seed** the affected queues from the new active: purge the returned cluster's copy
   and re-federate from scratch, so the forked tail is removed. Or **accept and document**
   the orphan tail as bounded loss.
2. Once the returned DC is caught up (low federation lag), perform a **drained planned
   switchover** back to it:

   ```bash
   kubectl annotate rabbitmq -n demo rm-dcdr dr.kubedb.com/switchover-to=dc-a
   ```

There is no rewind; RabbitMQ cannot roll back a queue, so the re-seed is what restores
consistency.

---

## 7. A standby DC is lost

**Symptoms:** the non-active DC's nodes are gone; that DC shows `healthy: false` and
federation lag is unavailable.

**Impact:** none on publishes. The active DC keeps taking client publishes; you lose the
DR copy until the standby returns.

**Verify the active DC is still writable:**

```bash
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** recover the standby cluster's nodes. When it returns, Federation resumes
replicating from the active and the standby catches up. You are running without DR
protection until then.

---

## 8. Federation lag high (replica falling behind)

**Symptoms:** `federationLagMessages` on the standby climbs; a planned switchover stays in
`FailingOver` because Federation has not drained.

**Diagnose:**

```bash
# Per-DC federation lag and health:
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.federationLagMessages} healthy={.healthy}{"\n"}{end}'
# Federation link status:
kubectl exec -n demo rm-dcdr-dc-b-0 -- rabbitmqctl list_federation_links
```

**Causes & action:**

- **Cross-DC network bottleneck or publisher burst:** relieve the bottleneck (network,
  active-cluster publish load) so Federation drains. A planned switchover intentionally
  waits for the lag to reach near-zero after quiescing publishers.
- **A federation link is down or blocked:** check `list_federation_links` and the
  upstream parameter; restart the link on the standby cluster (or the standby nodes).
- **Abort a stuck switchover:** remove the annotation to cancel:
  `kubectl annotate rabbitmq -n demo rm-dcdr dr.kubedb.com/switchover-to-`. The active
  DC's publish fence reopens and it stays active.

---

## 9. Arbiter DC lost

**Symptoms:** the Arbiter DC is gone; its `dr-controlplane` etcd member is unreachable.

**Impact:** none on publishes. The two data DCs plus the lost arbiter leave the etcd
quorum with two of three members, still a majority, so the Lease can still be renewed and
the active DC keeps taking publishes. You lose the tie-break, so a subsequent **second**
failure can no longer keep an etcd majority.

**Verify the cluster is still writable:**

```bash
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** restore the Arbiter DC's etcd member to regain single-fault tolerance. The
Arbiter DC never runs RabbitMQ, so no node recovery is involved.

---

## 10. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read or renewed across the spokes.

**Automatic:** the current active cluster keeps taking publishes as long as its fence can
still confirm the Lease it last held; but if the fence cannot confirm the Lease it
**fails closed** and denies publishes, so a total coordination-plane outage can make the
active cluster go read-only. What you always lose is **failover and planned switchover**:
the hub cannot move the active DC until the etcd quorum returns.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc   # error / stale
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery.phase}'   # may be Degraded
```

**Action:** restore the `dr-controlplane` etcd quorum (its third member lives in the
Arbiter DC). Once the Lease is renewable, the fence reopens on the active cluster and
failover/switchover resume.

---

## 11. Which DC is active?

**Question:** confirm which cluster is taking publishes right now.

```bash
# The DR status view (authoritative for what the hub applied):
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'
# The Lease holder (what the coordination plane intends):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Where the AMQP endpoint resolves and which DC is writable:
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

In steady state the `activeDC`, the Lease `holderIdentity`, and the single
`writable: true` DC all name the same data center. A brief mismatch during
`FailingOver` is expected; it should converge back to `Steady`.

---

## 12. Suspected split writes (two clusters taking publishes)

This should be impossible: the etcd majority grants the Lease to one DC only, and the
publish fence on any non-Lease-holder fails closed. If `status.disasterRecovery` ever
shows two `writable: true` DCs, or publishes succeed against both endpoints:

**Diagnose immediately:**

```bash
kubectl get rabbitmq -n demo rm-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o yaml
kubectl exec -n demo rm-dcdr-dc-b-0 -- rabbitmqctl list_federation_links   # both directions must not be enabled
```

**Action:** confirm the publish fence is engaged on the non-Lease-holder (write
permission revoked for client users, or the AMQP listener gated) and that only one
federation direction is enabled. Never enable both directions at once: Federation has no
rename loop guard, so overlapping directions ping-pong a message between clusters. Tear
down the wrong-direction upstream, restore the fence on the non-active cluster, then treat
any forked tail as scenario 6 (re-seed).

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get rabbitmq -n demo rm-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl get pods -n demo -l app.kubernetes.io/instance=rm-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name -o wide
kubectl exec -n demo rm-dcdr-dc-b-0 -- rabbitmqctl list_federation_links
kubectl exec -n demo rm-dcdr-dc-b-0 -- rabbitmqctl list_parameters
```
