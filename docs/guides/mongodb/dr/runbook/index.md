---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: mg-dr-runbook-mongodb
    name: Runbook
    parent: mg-dr-mongodb
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# MongoDB DC-DR Runbook

Scenario-by-scenario procedures for operating a MongoDB cluster in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB and
MongoDB do **automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/mongodb/dr/guide/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane
kubeconfig, `mg-dcdr`/`demo` are the example database and namespace.

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the DC the coordination plane intends as active):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC members, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=mg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# Replica-set members, priorities, and votes (against the primary):
kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.conf().members.map(m => ({host:m.host, priority:m.priority, votes:m.votes, arbiterOnly:m.arbiterOnly}))'
```

Golden rules:

- **MongoDB's majority election decides the primary.** Never force a member primary by
  hand and never run `replSetReconfig {force:true}` outside a documented double-failure
  DR action.
- **`w:majority` is the split-brain guarantee.** A minority DC cannot commit and a
  primary that loses its majority auto-steps-down.
- **Exactly one DC is `writable: true`** in `status.disasterRecovery`, and exactly one
  pod is labeled `kubedb.com/role: primary`, at any instant.
- **The Lease follows the primary.** A transient gap between the Lease-intended DC and
  the observed primary (during a member bounce and priority takeover) is expected.

---

## 1. Active DC lost (zone/cluster failure)

**Symptoms:** the active DC's members are gone/unreachable; writes fail briefly.

**Automatic:** the surviving voting members (a standby data DC plus the arbiter DC in
the even layout, or the surviving data majority in the odd layout) form a majority and
MongoDB **elects a new primary on its own**. The mode-detector relabels the new
primary `primary`. The orchestrator observes the new primary and moves the Lease to
match. In the 2 + 2 + 1 even layout it then issues a normal majority-committed
`replSetReconfig` dropping the lost DC's members, so `w:majority` writes (otherwise
stalled on two of five data members) resume. `phase` moves `FailingOver` to `Steady`.

**Verify:**

```bash
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get pods -n demo -l app.kubernetes.io/instance=mg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
```

**Action:** none required for availability. Note the RPO: `w:1` writes not yet
replicated when the DC died are lost (`w:majority` writes are not). When the failed DC
returns, see scenario 9 (re-add a DC).

---

## 2. `w:majority` writes stall after a data-DC loss (even layout only)

**Symptoms:** in the 2 + 2 + 1 layout, after losing a whole data DC, a primary is
elected but `w:majority` writes time out.

**Cause:** only two of the five data-bearing members are reachable, so a majority of
data acks is impossible (MongoDB's documented two-data-center limitation).

**Automatic:** the orchestrator issues a normal majority-committed `replSetReconfig`
that drops the lost members, so the majority recomputes to the survivors and
`w:majority` resumes. This is **not** a force reconfig (the surviving data DC plus the
arbiter hold a majority of the original config).

**Verify:**

```bash
kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.conf().members.length'   # dropped to the survivors
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{.status.disasterRecovery.phase}'    # Steady
```

**Action:** none if the reconfig completed. To avoid this stall entirely, run an
**odd** number of Member DCs with no Arbiter DC; a single DC loss then keeps a data
majority and `w:majority` never stalls.

---

## 3. Network partition between data centers

**Symptoms:** DCs are up but cannot reach each other.

**Automatic:** a primary on the minority side loses its majority and
**auto-steps-down** to a secondary, so the cut-off DC goes read-only on its own. With
`w:majority` a minority side cannot commit, so there is no split brain and the fence
needs no action (it could not act anyway: lowering priority is a normal reconfig that
needs a majority the isolated DC does not have, and force reconfig is forbidden). The
majority side keeps or elects the primary and stays writable.

**Verify there is exactly one writable DC:**

```bash
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. The minority side rejoins, rolls back any un-replicated
tail natively, and resumes as secondaries automatically.

---

## 4. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate mongodb -n demo mg-dcdr dr.kubedb.com/switchover-to=dc-b
```

**Automatic:** the hub gates on the target's health and oplog lag, raises the target
DC's `priority` by a normal `replSetReconfig`, then issues a **non-force**
`replSetStepDown` on the current primary. The non-force stepDown only proceeds once an
electable target secondary is caught up, so near-zero committed writes are lost. The
Lease follows to the new primary.

**Verify:**

```bash
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-b
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 7 (switchover stuck).

---

## 5. Planned failback to the original DC

After the original DC is healthy and caught up (failback is native: rollback of the
un-replicated tail, or a full initial resync if it fell outside the rollback/oplog
window), steer the primary back:

```bash
kubectl annotate mongodb -n demo mg-dcdr dr.kubedb.com/switchover-to=dc-a
```

Same near-zero-RPO flow as scenario 4. There is no `pg_rewind` step; MongoDB rejoins
the returned members on its own before they become electable.

---

## 6. Arbiter DC lost (even layout)

**Symptoms:** the Arbiter DC is gone; its etcd member and the MongoDB voting arbiter
are unreachable.

**Impact:** none on writes. The two data DCs together hold 4 of the 5 votes, still a
majority, so a primary holds and `w:majority` writes continue. You lose the tie-break
vote, so a subsequent **second** failure (a data DC) can no longer auto-elect.

**Verify the cluster is still writable:**

```bash
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** restore the Arbiter DC (the etcd member and the MongoDB arbiter) to regain
single-fault tolerance.

---

## 7. Planned switchover stuck (target not catching up)

**Symptoms:** after annotating `switchover-to`, `phase` stays `FailingOver` and the
primary does not move.

**Diagnose:**

```bash
# Target oplog lag and health:
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.oplogLagSeconds} healthy={.healthy}{"\n"}{end}'
# Replication state from the primary:
kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.printSecondaryReplicationInfo()'
```

**Causes & action:**

- **Target lag not converging** the non-force `replSetStepDown` refuses until an
  electable target secondary is caught up. Relieve the cross-DC bottleneck (network,
  primary write load) so the target drains its lag.
- **Target unhealthy** ensure the target DC has a ready, electable secondary.
- **Abort** remove the annotation to cancel:
  `kubectl annotate mongodb -n demo mg-dcdr dr.kubedb.com/switchover-to-`.

---

## 8. A standby DC is lost

**Symptoms:** a non-active DC's members are gone; that DC shows `healthy: false`.

**Impact:** none on writes in the odd layout (the active DC is unaffected). In the even
layout, losing a data DC is scenario 1/2. You lose that DC's redundancy and its
secondary read capacity until it returns.

**Verify the active DC is still writable:**

```bash
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** recover the DC's members; they reschedule and resync from the active
primary over the oplog automatically.

---

## 9. Re-add / recover a previously lost data center

After a DC returns from a failure:

**Automatic:** its members rejoin the replica set and catch up over the native oplog. A
member that was previously primary rolls back its un-replicated tail automatically, or
does a full initial resync if it fell outside the rollback/oplog window. In the even
layout, if the lost members were reconfigured out (scenario 2), the operator
reconfigs the returned members back in as low-priority secondaries.

**Verify:**

```bash
kubectl get mongodb -n demo mg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} healthy={.healthy} lag={.oplogLagSeconds}{"\n"}{end}'
```

**Action:** to make it active again, perform a planned failback (scenario 5) once its
oplog lag is small.

---

## 10. A DC is unexpectedly read-only

**Symptoms:** a DC you expect to run the primary has only secondaries.

**Diagnose:**

```bash
# Where is the primary, and what are the member priorities?
kubectl get pods -n demo -l app.kubernetes.io/instance=mg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.conf().members.map(m => ({host:m.host, priority:m.priority}))'
# What DC does the Lease intend?
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
```

**Causes & action:**

- **Transient priority takeover** during a member bounce a standby can briefly hold the
  primary and MongoDB priority takeover returns it. Wait a few seconds and recheck;
  this is expected.
- **Lease intends another DC** the priority is intentionally lower here; this DC is not
  the active one (correct).
- **Lost majority** the DC is partitioned or short of votes, so MongoDB cannot elect a
  primary here. See scenario 3.

Never run `replSetReconfig {force:true}` or patch `kubedb.com/role` by hand to force a
primary; the next reconcile reverts it and you risk a diverged config.

---

## 11. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read/renewed across the spokes.

**Automatic:** MongoDB keeps running on its own majority, so the cluster stays writable
in whichever DC holds the primary; the Lease is policy, not the failover mechanism, so
its loss does not by itself make MongoDB read-only. What you lose is **priority
steering and planned switchover**: the operator cannot reconfig priorities or move the
active DC until the Lease quorum returns.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc   # error / stale
kubectl get pods -n demo -l app.kubernetes.io/instance=mg-dcdr -L kubedb.com/role  # a primary still exists
```

**Action:** restore the `dr-controlplane` etcd quorum (in the even layout it shares the
Arbiter DC with the MongoDB arbiter). Once the Lease is renewable, priority steering and
switchover resume.

---

## 12. Suspected split-brain (two primaries)

This should be impossible with `w:majority` and 3-site votes (no single DC holds a
majority, and a primary that loses its majority auto-steps-down). If
`status.disasterRecovery` ever shows two `writable: true` DCs, or two pods labeled
`kubedb.com/role: primary`:

**Diagnose immediately:**

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=mg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.status().members.map(m => ({name:m.name, state:m.stateStr}))'
```

**Action:** confirm clients write with `w:majority` (a minority primary cannot commit
majority writes, so it cannot diverge committed data). The minority primary should
auto-step-down within its election timeout. Verify the vote layout still spreads votes
3-site (no single data DC was given a majority of votes by a bad reconfig). Do not
force-reconfig; restore connectivity and let MongoDB settle to a single primary.

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get mongodb -n demo mg-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl get pods -n demo -l app.kubernetes.io/instance=mg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name -o wide
kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.status()'
kubectl exec -n demo mg-dcdr-0 -- mongosh --quiet --eval 'rs.conf()'
```
