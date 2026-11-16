---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: rd-dr-runbook
    name: Runbook
    parent: rd-dr
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Redis DC-DR Runbook

Scenario-by-scenario procedures for operating a Redis (or Valkey) cluster in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB does
**automatically**, how to **verify**, and the **action** to take.

Read the [DC-DR Overview](/docs/guides/redis/dr/overview/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane kubeconfig,
`redis-dcdr`/`demo` are the example database and namespace.

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get redis -n demo redis-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the source of truth for "who is active"):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC masters, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=redis-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# A spoke's marker (what its coordinators read):
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'
```

Golden rules:

- **The Lease decides who is writable.** Never run `REPLICAOF NO ONE` or a manual
  `SENTINEL FAILOVER` across DCs by hand.
- **The fence fails closed.** A DC that cannot confirm it holds the Lease keeps its master
  labeled `standby` and read-only by design; that is correct, not a bug.
- **Exactly one DC is `writable: true`** in `status.disasterRecovery` at any instant.

---

## 1. Active DC lost (zone/cluster failure)

**Symptoms:** the active DC's pods are gone/unreachable; writes fail briefly.

**Automatic:** the lost DC's agents stop renewing the Lease. After the Lease duration (~45s)
the surviving Member DC's agent acquires it and becomes `activeDC`. The hub records the change;
the survivor's fence clears, the operator runs `REPLICAOF NO ONE` on the survivor's master to
promote it, relabels it `primary`, and re-points any other standby DC at it. The old DC, if
partially alive, has already self-fenced (its master stays `standby`, read-only). The primary
`Service` and `AppBinding` follow to the new DC. `phase` moves `FailingOver` to `Steady`.

**Verify:**

```bash
kubectl get redis -n demo redis-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get redis -n demo redis-dcdr -o jsonpath='{.status.disasterRecovery.phase}'      # Steady
```

**Action:** none required for availability. Note the RPO: the replication tail not yet shipped
when the DC died (the offset gap) is lost, because Redis cross-DC replication is asynchronous.
When the failed DC returns, see scenario 9 (re-add a DC).

---

## 2. Network partition between data centers

**Symptoms:** DCs are up but cannot reach each other or the coordination plane.

**Automatic:** the side that loses the coordination plane stops getting Lease updates; its
marker `renewTime` freezes and, after the 30s fence TTL, its `rd-coordinator` keeps its master
labeled `standby` (read-only), **before** the Lease duration lets the other side acquire it
(this is the timing invariant). The side that keeps the etcd majority holds/acquires the Lease
and stays (or becomes) writable. There is no split-brain.

**Verify there is exactly one writable DC:**

```bash
kubectl get redis -n demo redis-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. The fenced side re-points its master `REPLICAOF` the active DC and
resumes its stream automatically (or full-resyncs if it diverged). If both sides show
`writable: false`, see scenario 6 (coordination plane down).

---

## 3. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate redis -n demo redis-dcdr dr.kubedb.com/switchover-to=dc-west
```

**Automatic:** the hub gates on the target's health and offset lag, quiesces writes on the
active master (holds it read-only through the Lease quiesce marker), waits until the target
master's replicated offset reaches the active master's `master_repl_offset`, then promotes the
target (`REPLICAOF NO ONE`), makes the old primary a `REPLICAOF` replica of the new active, and
hands off the Lease. No acknowledged writes are lost. The annotation is cleared on completion.

**Verify:**

```bash
kubectl get redis -n demo redis-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-west
kubectl get redis -n demo redis-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 7 (switchover stuck).

---

## 4. Planned failback to the original DC

After the original DC is healthy and caught up:

```bash
kubectl annotate redis -n demo redis-dcdr dr.kubedb.com/switchover-to=dc-east
```

Same near-zero-RPO flow as scenario 3. A DC that previously lost the Lease rejoins as a
`REPLICAOF` replica and catches up (partial resync, or a full RDB resync that drops the diverged
tail, see scenario 9) before it is eligible, so failback is safe even after an unplanned
failover.

---

## 5. A standby DC is lost

**Symptoms:** the non-active DC's pods are gone; that DC shows `healthy: false`.

**Impact:** none on writes, the active DC is unaffected. You lose your only DR copy and the
standby read capacity until it returns.

**Automatic:** nothing to fail over. The active master's replication backlog is bounded, so a
long-gone standby does not grow the active master's memory without limit.

**Action:** bring the DC back. When it returns it re-points `REPLICAOF` the active master and
resumes streaming (partial resync if the backlog still covers its offset, otherwise a full RDB
resync). Watch `linkStatus` return to `up` and `lagBytes` fall in
`status.disasterRecovery`.

---

## 6. Coordination plane (etcd quorum) down

**Symptoms:** no DC can renew or acquire the Lease; every DC's marker goes stale.

**Automatic:** every DC fails closed. After the 30s fence TTL, all masters are held `standby`
(read-only). Writes stop everywhere. This is the safe state: with no quorum, no DC can prove it
is the single writable DC.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc     # stale or missing renewTime
```

**Action:** restore the `dr-controlplane` etcd quorum (a majority of its three sites). As soon
as a DC's marker refreshes and names it active, its fence lifts and it resumes as the writable
primary. Do **not** force a master writable by hand while the quorum is down.

---

## 7. Switchover stuck (`phase: FailingOver` persists)

**Symptoms:** after annotating `dr.kubedb.com/switchover-to`, the active DC does not move.

**Likely causes and checks:**

- **Target lag over budget.** The hub will not hand off until the target is within
  `dr.kubedb.com/switchover-max-lag-bytes` (default 16Mi). Check the target's `lagBytes` and
  `linkStatus`:

  ```bash
  kubectl get redis -n demo redis-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}: link={.linkStatus} lag={.lagBytes}{"\n"}{end}'
  ```

  Wait for the standby to catch up, or raise the budget annotation if appropriate.
- **Target not healthy.** The target DC must show `healthy: true`.
- **Write pressure.** A very high write rate can keep the target from reaching offset equality;
  the quiesce freezes the active offset, so it should converge. If it does not, reduce load and
  retry.

**Action:** to abort, remove the annotation:

```bash
kubectl annotate redis -n demo redis-dcdr dr.kubedb.com/switchover-to-
```

The active DC stays where it is and the quiesce is cleared.

---

## 8. Marker stale on a healthy DC

**Symptoms:** a DC's master is unexpectedly held `standby`; its marker `renewTime` is old.

**Automatic:** this is the fence working. The `rd-coordinator` refuses to promote a master whose
`primary-dc` marker is missing, unparseable, stale (older than 30s), or names another DC.

**Verify:**

```bash
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'    # activeDC + renewTime
```

**Action:** find why the DC's `dr-controlplane` agent is not refreshing the marker (agent down,
RBAC on the `dc-failover` ConfigMap, or the DC genuinely lost the Lease). Fix the agent or the
connectivity; the marker refreshes and the fence lifts on its own.

---

## 9. Re-add a data center after it returns

**Symptoms:** a previously lost DC is back but was the old active DC, so its data may have
diverged from the new active.

**Automatic:** the operator makes the returned DC a `REPLICAOF` replica of the new active. Redis
reconciles it by partial resync from the backlog if possible, otherwise a full RDB resync that
**drops the diverged tail** and re-seeds from the active master. There is no manual rewind.

**Verify it caught up:**

```bash
kubectl get redis -n demo redis-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}: link={.linkStatus} lag={.lagBytes}{"\n"}{end}'
```

**Action:** once `linkStatus: up` and `lagBytes` is small, the DC is a healthy standby again. To
return the active role to it, run a planned failback (scenario 4).

---

## 10. Verifying the split-brain guarantee

At any instant there must be exactly one writable DC. To assert it:

```bash
# From status (what the hub sees):
kubectl get redis -n demo redis-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'

# From the pods (what the role labels say): exactly one primary across all DCs.
kubectl get pods -n demo -l app.kubernetes.io/instance=redis-dcdr,kubedb.com/role=primary -L open-cluster-management.io/cluster-name
```

If ever two DCs report `writable: true` or two pods across DCs carry `kubedb.com/role: primary`,
capture the Lease (`kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o yaml`)
and both spokes' `primary-dc` markers and open an issue; the fence and the etcd majority are
designed to make this impossible.
