---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-dr-runbook
    name: Runbook
    parent: guides-postgres-dr
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Postgres DC-DR Runbook

Scenario-by-scenario procedures for operating a Postgres cluster in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB does
**automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/postgres/dr/guide/index.md) for the
concepts and commands referenced here. Throughout, `<coord>` is the coordination
control plane kubeconfig, `pg-dcdr`/`demo` are the example database and namespace.

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get pg -n demo pg-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the source of truth for "who is active"):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC leaders, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=pg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# A spoke's marker (what its coordinators read):
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'
```

Golden rules:

- **The Lease decides who is writable.** Never make a pod writable by hand.
- **The fence fails closed.** A DC that cannot confirm it holds the Lease is read-only
  by design — that is correct, not a bug.
- **Exactly one DC is `writable: true`** in `status.disasterRecovery` at any instant.

---

## 1. Active DC lost (zone/cluster failure)

**Symptoms:** the active DC's pods are gone/unreachable; writes fail briefly.

**Automatic:** the lost DC's agents stop renewing the Lease. After the Lease duration
(~45s) a surviving Member DC's agent acquires it and becomes `activeDC`. The hub drives
a bounded-loss promotion (`ForceFailOver` `PostgresOpsRequest`) of the survivor; the
old DC, if partially alive, self-fences read-only. The primary `Service` and
`AppBinding` follow to the new DC. `phase` moves `FailingOver` → `Steady`.

**Verify:**

```bash
kubectl get pg -n demo pg-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get postgresopsrequest -n demo -l app.kubernetes.io/managed-by=kubedb-dcdr  # the failover ops
```

**Action:** none required for availability. Note the RPO: writes not yet replicated
when the DC died are lost. When the failed DC returns, see scenario 11 (re-add a DC).

---

## 2. Network partition between data centers

**Symptoms:** DCs are up but cannot reach each other or the coordination plane.

**Automatic:** the side that loses the coordination plane stops getting Lease updates;
its marker `renewTime` freezes and, after the 30s fence TTL, its coordinator demotes
its leader to read-only — **before** the Lease duration lets the other side acquire
(this is the timing invariant). The side that keeps the etcd majority holds/acquires
the Lease and stays (or becomes) writable. There is no split-brain.

**Verify there is exactly one writable DC:**

```bash
kubectl get pg -n demo pg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. The fenced side rejoins, rewinds any divergent tail, and
resumes streaming automatically. If both sides show `writable: false`, see scenario 6
(coordination plane down).

---

## 3. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate pg -n demo pg-dcdr dr.kubedb.com/switchover-to=dc-west
```

**Automatic:** the hub gates on the target's health and lag, quiesces the active DC
(holds its primary read-only via the Lease), waits until the target catches up to
within one WAL page, then hands off. Zero committed rows are lost. The annotation is
cleared on completion.

**Verify:**

```bash
kubectl get pg -n demo pg-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-west
kubectl get pg -n demo pg-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 8 (switchover stuck).

---

## 4. Planned failback to the original DC

After the original DC is healthy and caught up:

```bash
kubectl annotate pg -n demo pg-dcdr dr.kubedb.com/switchover-to=dc-east
```

Same zero-RPO flow as scenario 3. A DC that previously lost the Lease rewinds its
divergent tail (`pg_rewind`, base-backup reseed fallback) before it is eligible, so
failback is safe even after an unplanned failover.

---

## 5. A standby DC is lost

**Symptoms:** a non-active DC's pods are gone; that DC shows `healthy: false`.

**Impact:** none on writes — the active DC is unaffected. You lose that DC's
redundancy and its standby read capacity until it returns.

**Verify the active DC is still writable:**

```bash
kubectl get pg -n demo pg-dcdr -o jsonpath='{.status.disasterRecovery.dataCenters[?(@.writable==true)].clusterName}'
```

**Action:** recover the DC's nodes; the per-DC group reschedules and re-seeds from the
active primary automatically.

---

## 6. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read/renewed; markers go stale across all spokes;
every DC eventually fences read-only.

**Automatic:** this is fail-closed — with no trustworthy Lease, **no** DC is allowed to
be writable. The database is read-only globally rather than risk split-brain.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc        # error / stale renewTime
kubectl get pg -n demo pg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'  # all false
```

**Action:** restore the `dr-controlplane` etcd quorum. Once the Lease is renewable, the
holder's marker refreshes, its coordinator un-fences, and writes resume. Do **not**
force a pod writable to work around this.

---

## 7. Failover not promoting a survivor

**Symptoms:** the active DC is gone but `activeDC` does not move, or no writable DC
appears.

**Diagnose:**

```bash
# Did the Lease move?
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Are there candidate pods in the survivor DC?
kubectl get pods -n demo -l app.kubernetes.io/instance=pg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
# Did the hub create a failover ops request, and what is its phase?
kubectl get postgresopsrequest -n demo -l app.kubernetes.io/managed-by=kubedb-dcdr -o wide
```

**Common causes & action:**

- **Lease did not move** — only Member DCs are eligible; confirm the survivor is a
  `Member` in the `PlacementPolicy` and its agent can reach the coordination plane.
- **No candidates** — the survivor DC has no ready data pods; recover its pods.
- **Ops request failed** — inspect its conditions; the hub does not create a duplicate
  while one is open, so resolve or delete the stuck request and let reconcile retry.

---

## 8. Planned switchover stuck (target not catching up)

**Symptoms:** after annotating `switchover-to`, `phase` stays `FailingOver` and the
Lease does not hand off.

**Diagnose:**

```bash
# Target lag and health:
kubectl get pg -n demo pg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lag={.lagBytes} healthy={.healthy}{"\n"}{end}'
```

**Causes & action:**

- **Target lag not converging** — the active DC must be quiesced for the target to
  reach the frozen LSN. Confirm the active DC's coordinator honored the quiesce (its
  primary should be read-only); check the marker's `data.quiesce` names the active DC.
- **Target unhealthy / no lag report** — the switchover refuses a target with no
  `lagBytes` yet; ensure the target DC's leader is up and publishing lag.
- **Target legitimately too far behind** — raise the budget only if you accept the
  catch-up time: `kubectl annotate pg -n demo pg-dcdr dr.kubedb.com/switchover-max-lag-bytes=<bytes>`.
- **Abort** — remove the annotation to cancel:
  `kubectl annotate pg -n demo pg-dcdr dr.kubedb.com/switchover-to-`.

---

## 9. Lag growing on a standby DC

**Symptoms:** a DC's `lagBytes` climbs steadily.

**Diagnose:** cross-DC network throughput/latency, write volume on the active primary,
and replication health on the standby DC's leader.

```bash
kubectl get pod -n demo <standby-dc-leader> -o jsonpath='{.metadata.annotations.kubedb\.com/dc-lag-bytes}'
# On the active primary, check the cross-DC replica:
#   SELECT client_addr, state, sent_lsn, replay_lsn FROM pg_stat_replication;
```

**Action:** relieve the bottleneck (network, primary load). High lag widens the RPO of
an unplanned failover and can block a planned switchover until it drains.

---

## 10. A DC is unexpectedly read-only (fence tripped)

**Symptoms:** a DC you expect to be active is read-only; its leader is labeled
`standby`.

**Diagnose the fence chain:**

```bash
# Does this spoke's marker name this DC and is renewTime fresh?
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'
# Is the dr-controlplane agent running in this DC and renewing?
kubectl get pods -n <agent-namespace> -l app=dr-controlplane-agent
```

**Causes & action:**

- **Marker stale** (`renewTime` old) — the agent cannot reach the coordination plane,
  or the projector is failing; restore agent connectivity.
- **Marker names another DC** — this DC simply is not the active one (correct).
- **Clock skew** — large cross-DC clock skew can trip the TTL early; verify NTP. The
  timing budget assumes skew is well under (LeaseDuration − fence TTL).

Never patch `kubedb.com/role` by hand to force writability — the next reconcile and the
fence will revert it, and you risk split-brain.

---

## 11. Re-add / recover a previously lost data center

After a DC returns from a failure:

**Automatic:** its per-DC group reschedules, and each pod seeds from the active DC
primary (a node that was previously active rewinds its divergent tail first). Once
caught up, the DC's leader becomes a healthy read-only standby and its `lagBytes`
appears in status.

**Verify:**

```bash
kubectl get pg -n demo pg-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} healthy={.healthy} lag={.lagBytes}{"\n"}{end}'
```

**Action:** to make it active again, perform a planned failback (scenario 4) once its
lag is small.

---

## 12. Scale a data center up or down

```bash
# Scale dc-west to 5 nodes:
kubectl apply -f - <<'YAML'
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata: { name: pg-dcdr-scale-west, namespace: demo }
spec:
  type: HorizontalScaling
  databaseRef: { name: pg-dcdr }
  horizontalScaling:
    dataCenters:
    - { clusterName: dc-west, replicas: 5 }
YAML
```

**Verify:** the per-DC PetSet reaches the new size, the arbiter appears/disappears with
parity, and only that DC changed.

```bash
kubectl get petset -n demo -l app.kubernetes.io/instance=pg-dcdr
kubectl get postgresopsrequest -n demo pg-dcdr-scale-west -o jsonpath='{.status.phase}'
```

**Notes:** scaling to `1` is allowed (single-node DC, no in-DC HA); scaling to `0` is
rejected — removing a DC is a topology change.

---

## 13. Version upgrade in DC-DR

Issue a normal `UpdateVersion` `PostgresOpsRequest`; the operator updates every per-DC
PetSet. Plan it during a low-traffic window and confirm each DC returns healthy in
`status.disasterRecovery` before relying on failover again.

---

## 14. Suspected split-brain (two writable DCs)

This should be impossible by design (etcd majority + fail-closed fence + the timing
invariant). If `status.disasterRecovery` ever shows two `writable: true` DCs, or two
pods labeled `kubedb.com/role: primary`:

**Diagnose immediately:**

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=pg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
```

**Action:** the Lease holder is the true active DC. The other DC's fence should trip
within the TTL; if it does not, its marker is wrong (check its agent/projector) or the
timing invariant is misconfigured (verify fence TTL < Lease duration). Stop writes to
the non-Lease-holder DC at the application layer until the fence reasserts, then
reconcile (the non-holder rewinds and rejoins as a standby).

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get pg -n demo pg-dcdr -o yaml
kubectl get postgresopsrequest -n demo -l app.kubernetes.io/managed-by=kubedb-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl -n dc-failover get configmap primary-dc -o yaml   # on each spoke
kubectl get pods -n demo -l app.kubernetes.io/instance=pg-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name -o wide
kubectl logs -n demo <leader-pod> -c pg-coordinator
```
