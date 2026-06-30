---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-dr-runbook
    name: Runbook
    parent: guides-mysql-dr
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# MySQL DC-DR Runbook

Scenario-by-scenario procedures for operating a MySQL cluster in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB does
**automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/mysql/dr/guide/index.md) for the concepts and commands
referenced here. Throughout, `<coord>` is the coordination control plane kubeconfig,
`my-dcdr`/`demo` are the example database and namespace.

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the source of truth for "who is active"):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC GR primaries, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=my-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# A spoke's marker (what its coordinators read):
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'
```

Golden rules:

- **The Lease decides who is writable.** Never clear `super_read_only` on a pod by hand.
- **The fence fails closed.** A DC that cannot confirm it holds the Lease is
  `super_read_only` by design — that is correct, not a bug.
- **Exactly one DC is `writable: true`** in `status.disasterRecovery` at any instant.

---

## 1. Active DC lost (zone/cluster failure)

**Symptoms:** the active DC's pods are gone/unreachable; writes fail briefly.

**Automatic:** the lost DC's agents stop renewing the Lease. After the Lease duration
(~45s) a surviving Member DC's agent acquires it and becomes `activeDC`. The hub clears the
survivor's fence (relabels its GR primary `primary`, `super_read_only = OFF`, stops its
inbound channel) and repoints every other standby's channel at the new active; the old DC,
if partially alive, self-fences read-only. The primary `Service` and `AppBinding` follow to
the new DC. `phase` moves `FailingOver` → `Steady`.

**Verify:**

```bash
kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery.phase}'      # Steady
```

**Action:** none required for availability. Note the RPO: transactions not yet replicated
when the DC died (the un-shipped GTID tail) are lost. When the failed DC returns, see
scenario 11 (re-add a DC).

---

## 2. Network partition between data centers

**Symptoms:** DCs are up but cannot reach each other or the coordination plane.

**Automatic:** the side that loses the coordination plane stops getting Lease updates; its
marker `renewTime` freezes and, after the 30s fence TTL, its coordinator holds its GR
primary `super_read_only = ON` — **before** the Lease duration lets the other side acquire
(this is the timing invariant). The side that keeps the etcd majority holds/acquires the
Lease and stays (or becomes) writable. There is no split-brain.

**Verify there is exactly one writable DC:**

```bash
kubectl get my -n demo my-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. The fenced side rejoins, GTID auto-positions (or re-seeds if
diverged beyond the purged GTIDs), and resumes its channel automatically. If both sides
show `writable: false`, see scenario 6 (coordination plane down).

---

## 3. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate my -n demo my-dcdr dr.kubedb.com/switchover-to=dc-west
```

**Automatic:** the hub gates on the target's health and lag, quiesces the active DC (holds
its GR primary `super_read_only = ON`), waits until the target's channel applies up to the
frozen `gtid_executed`, then hands off. Zero committed rows are lost. The annotation is
cleared on completion.

**Verify:**

```bash
kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-west
kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 8 (switchover stuck).

---

## 4. Planned failback to the original DC

After the original DC is healthy and caught up:

```bash
kubectl annotate my -n demo my-dcdr dr.kubedb.com/switchover-to=dc-east
```

Same zero-RPO flow as scenario 3. A DC that previously lost the Lease catches up by GTID
(or re-seeds via the clone plugin if it diverged beyond the purged GTIDs) before it is
eligible, so failback is safe even after an unplanned failover.

---

## 5. A standby DC is lost

**Symptoms:** a non-active DC's pods are gone; that DC shows `healthy: false`.

**Impact:** none on writes — the active DC is unaffected. You lose that DC's redundancy and
its standby read capacity until it returns.

**Verify the active DC is still writable:**

```bash
kubectl get my -n demo my-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'
# Full per-DC writability:
kubectl get my -n demo my-dcdr -o json | jq -r '.status.disasterRecovery.dataCenters[] | "\(.clusterName)=\(.writable)"'
```

**Action:** recover the DC's nodes; the per-DC GR group reschedules and its primary
re-establishes the cross-DC channel from the active automatically.

---

## 6. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read/renewed; markers go stale across all spokes; every
DC eventually fences `super_read_only`.

**Automatic:** this is fail-closed — with no trustworthy Lease, **no** DC is allowed to be
writable. The database is read-only globally rather than risk split-brain.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc        # error / stale renewTime
kubectl get my -n demo my-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'  # all false
```

**Action:** restore the `dr-controlplane` etcd quorum. Once the Lease is renewable, the
holder's marker refreshes, its coordinator clears `super_read_only`, and writes resume. Do
**not** clear `super_read_only` on a pod to work around this.

---

## 7. Failover not promoting a survivor

**Symptoms:** the active DC is gone but `activeDC` does not move, or no writable DC appears.

**Diagnose:**

```bash
# Did the Lease move?
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Are there ONLINE GR members in the survivor DC?
kubectl get pods -n demo -l app.kubernetes.io/instance=my-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
```

**Common causes & action:**

- **Lease did not move** — only Member DCs are eligible; confirm the survivor is a `Member`
  in the `PlacementPolicy` and its agent can reach the coordination plane.
- **No quorum in the survivor** — the survivor DC's GR group lost its own intra-DC majority
  (for example an even local size that lost a node); recover its pods so GR re-forms.
- **Fence not clearing** — confirm the marker on the survivor's spoke names the survivor and
  has a fresh `renewTime`; the coordinator clears `super_read_only` only once it does.

---

## 8. Planned switchover stuck (target not catching up)

**Symptoms:** after annotating `switchover-to`, `phase` stays `FailingOver` and the Lease
does not hand off.

**Diagnose:**

```bash
# Target lag and health:
kubectl get my -n demo my-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} lagBytes={.lagBytes} secs={.secondsBehindSource} healthy={.healthy}{"\n"}{end}'
```

**Causes & action:**

- **Target lag not converging** — the active DC must be quiesced for the target to reach the
  frozen GTID set. Confirm the active DC's GR primary is `super_read_only` (its coordinator
  honored the quiesce).
- **Target unhealthy / no lag report** — the switchover refuses a target with no lag yet;
  ensure the target DC's GR primary is up and its channel is running
  (`Replica_IO_Running`/`Replica_SQL_Running = Yes`).
- **Target legitimately too far behind** — let it drain, or relieve the cross-DC bottleneck
  (scenario 9) before retrying.
- **Abort** — remove the annotation to cancel:
  `kubectl annotate my -n demo my-dcdr dr.kubedb.com/switchover-to-`.

---

## 9. Lag growing on a standby DC

**Symptoms:** a DC's `lagBytes` / `secondsBehindSource` climbs steadily.

**Diagnose:** cross-DC network throughput/latency, write volume on the active primary, and
the channel on the standby DC's GR primary.

```bash
# On the standby DC's GR primary:
#   SHOW REPLICA STATUS FOR CHANNEL 'dcdr'\G   (Seconds_Behind_Source, Replica_SQL_Running)
# On the active primary, the cross-DC client:
#   SELECT * FROM performance_schema.replication_connection_status\G
```

**Action:** relieve the bottleneck (network, primary load). High lag widens the RPO of an
unplanned failover and can block a planned switchover until it drains. Ensure the active
DC's binlog retention (`binlog_expire_logs_seconds`) outlasts the lag, or a slow DR DC's
channel can break when needed binlogs are purged.

---

## 10. A DC is unexpectedly read-only (fence tripped)

**Symptoms:** a DC you expect to be active is `super_read_only`; its GR primary is labeled
`standby`.

**Diagnose the fence chain:**

```bash
# Does this spoke's marker name this DC and is renewTime fresh?
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'
# Is the dr-controlplane agent running in this DC and renewing?
kubectl get pods -n <agent-namespace> -l app=dr-controlplane-agent
```

**Causes & action:**

- **Marker stale** (`renewTime` old) — the agent cannot reach the coordination plane, or the
  projector is failing; restore agent connectivity.
- **Marker names another DC** — this DC simply is not the active one (correct).
- **Clock skew** — large cross-DC clock skew can trip the TTL early; verify NTP. The timing
  budget assumes skew is well under (LeaseDuration − fence TTL).

Never clear `super_read_only` or patch `kubedb.com/role` by hand to force writability — the
next label loop re-asserts the fence, and you risk split-brain.

---

## 11. Re-add / recover a previously lost data center

After a DC returns from a failure:

**Automatic:** its per-DC GR group re-forms, and its primary starts a channel from the
active DC and catches up by GTID auto-positioning (a DC that diverged beyond the active's
purged GTIDs is re-seeded via the clone plugin first). Once caught up, the DC's primary
becomes a healthy `super_read_only` standby and its lag appears in status.

**Verify:**

```bash
kubectl get my -n demo my-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} healthy={.healthy} lagBytes={.lagBytes}{"\n"}{end}'
```

**Action:** to make it active again, perform a planned failback (scenario 4) once its lag is
small.

---

## 12. Scale a data center up or down

```bash
# Scale dc-west to 5 nodes:
kubectl apply -f - <<'YAML'
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata: { name: my-dcdr-scale-west, namespace: demo }
spec:
  type: HorizontalScaling
  databaseRef: { name: my-dcdr }
  horizontalScaling:
    dataCenters:
    - { clusterName: dc-west, replicas: 5 }
YAML
```

**Verify:** the per-DC PetSet reaches the new size and only that DC changed.

```bash
kubectl get petset -n demo -l app.kubernetes.io/instance=my-dcdr
kubectl get mysqlopsrequest -n demo my-dcdr-scale-west -o jsonpath='{.status.phase}'
```

**Notes:** keep each Member DC's count **odd** for a clean GR majority; removing a whole DC
is a topology change, not horizontal scaling.

---

## 13. Version upgrade in DC-DR

Issue a normal `UpdateVersion` `MySQLOpsRequest`; the operator updates every per-DC PetSet.
Plan it during a low-traffic window and confirm each DC returns healthy in
`status.disasterRecovery` before relying on failover again.

---

## 14. Suspected split-brain (two writable DCs)

This should be impossible by design (etcd majority + fail-closed fence + the timing
invariant + fence re-assertion after every GR election). If `status.disasterRecovery` ever
shows two `writable: true` DCs, or two pods labeled `kubedb.com/role: primary`:

**Diagnose immediately:**

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=my-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
```

**Action:** the Lease holder is the true active DC. The other DC's fence should trip within
the TTL; if it does not, its marker is wrong (check its agent/projector) or the timing
invariant is misconfigured (verify fence TTL < Lease duration). Stop writes to the
non-Lease-holder DC at the application layer until the fence reasserts, then reconcile (the
non-holder GTID catches up or re-seeds and rejoins as a standby).

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get my -n demo my-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl -n dc-failover get configmap primary-dc -o yaml   # on each spoke
kubectl get pods -n demo -l app.kubernetes.io/instance=my-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name -o wide
kubectl logs -n demo <gr-primary-pod> -c mysql-coordinator
```
