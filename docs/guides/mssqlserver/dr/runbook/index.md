---
title: DC-DR Runbook
menu:
  docs_{{ .version }}:
    identifier: guides-mssqlserver-dr-runbook
    name: Runbook
    parent: guides-mssqlserver-dr
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# MSSQLServer DC-DR Runbook

Scenario-by-scenario procedures for operating an MSSQLServer cluster in cross data center
disaster recovery (DC-DR) mode. Each scenario lists the **symptoms**, what KubeDB does
**automatically**, how to **verify**, and the **action** to take.

Read the [User Guide](/docs/guides/mssqlserver/dr/guide/index.md) for the concepts and
commands referenced here. Throughout, `<coord>` is the coordination control plane kubeconfig,
`mssql-dcdr`/`demo` are the example database and namespace.

## Quick reference

```bash
# Active DC, phase, and per-DC view:
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{.status.disasterRecovery}' | jq

# Lease holder (the source of truth for "who is active"):
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'

# Per-DC AG primaries, roles, and DCs:
kubectl get pods -n demo -l app.kubernetes.io/instance=mssql-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name

# A spoke's marker (what its coordinators read):
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'
```

Golden rules:

- **The Lease decides who is writable.** Never run `SET (ROLE = PRIMARY)` or a manual DAG
  failover by hand.
- **The fence fails closed.** A DC that cannot confirm it holds the Lease holds its AG as the
  DAG secondary by design, that is correct, not a bug.
- **Exactly one DC is `writable: true`** in `status.disasterRecovery` at any instant.

---

## 1. Active DC lost (zone/cluster failure)

**Symptoms:** the active DC's pods are gone/unreachable; writes fail briefly.

**Automatic:** the lost DC's agents stop renewing the Lease. After the Lease duration (~45s)
the surviving Member DC's agent acquires it and becomes `activeDC`. The hub clears the
survivor's fence: it runs `ALTER AVAILABILITY GROUP [dag] FORCE_FAILOVER_ALLOW_DATA_LOSS` on
the survivor AG, flips `self.role` to `Primary`, and relabels its AG primary `primary`. The
old DC, if partially alive, has already self-fenced its AG to DAG secondary. The primary
`Service` and `AppBinding` follow to the new DC. `phase` moves `FailingOver` to `Steady`.

**Verify:**

```bash
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'   # the survivor
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{.status.disasterRecovery.phase}'      # Steady
```

**Action:** none required for availability. Note the RPO: redo not yet shipped when the DC
died (the un-shipped tail) is lost. When the failed DC returns, see scenario 11 (re-add a
DC).

---

## 2. Network partition between data centers

**Symptoms:** DCs are up but cannot reach each other or the coordination plane.

**Automatic:** the side that loses the coordination plane stops getting Lease updates; its
marker `renewTime` freezes and, after the 30s fence TTL, its coordinator forces its AG to the
DAG `SECONDARY` role, **before** the Lease duration lets the other side acquire (this is the
timing invariant). The side that keeps the etcd majority holds/acquires the Lease and stays
(or becomes) writable. There is no split-brain.

**Verify there is exactly one writable DC:**

```bash
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'
```

**Action:** heal the network. The fenced side rejoins the DAG as secondary and resumes its
stream automatically (or re-seeds if it diverged). If both sides show `writable: false`, see
scenario 6 (coordination plane down).

---

## 3. Planned switchover (maintenance on the active DC)

**Action:**

```bash
kubectl annotate mssqlserver -n demo mssql-dcdr dr.kubedb.com/switchover-to=dc-west
```

**Automatic:** the hub gates on the target's health and lag, switches the DAG to
`SYNCHRONOUS_COMMIT` on both AGs, waits until the target AG's `last_hardened_lsn` equals the
active AG's, then runs `SET (ROLE = SECONDARY)` on the old AG and a graceful
`FORCE_FAILOVER_ALLOW_DATA_LOSS` on the new AG and hands off the Lease. Zero committed rows
are lost. The annotation is cleared on completion.

**Verify:**

```bash
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{.status.disasterRecovery.activeDC}'  # dc-west
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{.status.disasterRecovery.phase}'     # Steady
```

**If it does not complete:** see scenario 8 (switchover stuck).

---

## 4. Planned failback to the original DC

After the original DC is healthy and caught up:

```bash
kubectl annotate mssqlserver -n demo mssql-dcdr dr.kubedb.com/switchover-to=dc-east
```

Same zero-RPO flow as scenario 3. A DC that previously lost the Lease rejoins the DAG as
secondary and catches up (or re-seeds the diverged databases over 5022, see scenario 11)
before it is eligible, so failback is safe even after an unplanned failover.

---

## 5. A standby DC is lost

**Symptoms:** the non-active DC's pods are gone; that DC shows `healthy: false`.

**Impact:** none on writes, the active DC is unaffected. You lose your only DR copy and the
standby read capacity until it returns.

**Verify the active DC is still writable:**

```bash
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[?(@.writable==true)]}{.clusterName}{end}'
```

**Action:** recover the DC's nodes; the per-DC AG reschedules and its primary re-establishes
the DAG stream from the active automatically.

---

## 6. Coordination plane (dr-controlplane / etcd) unavailable

**Symptoms:** the Lease cannot be read/renewed; markers go stale across all spokes; every DC
eventually fences to the DAG secondary role.

**Automatic:** this is fail-closed, with no trustworthy Lease, **no** DC is allowed to be
writable. The database is read-only globally rather than risk split-brain.

**Verify:**

```bash
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc        # error / stale renewTime
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName}={.writable} {end}'  # all false
```

**Action:** restore the `dr-controlplane` etcd quorum. Once the Lease is renewable, the
holder's marker refreshes, its coordinator returns its AG to the DAG primary role, and writes
resume. Do **not** run a manual DAG failover to work around this.

---

## 7. Failover not promoting a survivor

**Symptoms:** the active DC is gone but `activeDC` does not move, or no writable DC appears.

**Diagnose:**

```bash
# Did the Lease move?
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
# Is the survivor AG primary up?
kubectl get pods -n demo -l app.kubernetes.io/instance=mssql-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
```

**Common causes & action:**

- **Lease did not move** then only Member DCs are eligible; confirm the survivor is a `Member`
  in the `PlacementPolicy` and its agent can reach the coordination plane.
- **No quorum in the survivor** then the survivor DC's AG lost its own intra-DC
  coordinator-raft majority (for example an even local size that lost a node); recover its
  pods so the AG re-forms.
- **Fence not clearing** then confirm the marker on the survivor's spoke names the survivor
  and has a fresh `renewTime`; the coordinator returns the AG to the DAG primary role only
  once it does.

---

## 8. Planned switchover stuck (target not catching up)

**Symptoms:** after annotating `switchover-to`, `phase` stays `FailingOver` and the Lease does
not hand off.

**Diagnose:**

```bash
# Target lag and health:
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} redo={.redoQueueBytes} send={.logSendQueueBytes} sync={.synchronizationHealth} healthy={.healthy}{"\n"}{end}'
```

**Causes & action:**

- **LSN equality never reached** then the DAG must be `SYNCHRONOUS_COMMIT` for the target to
  reach the active AG's `last_hardened_lsn`. Confirm the switch to synchronous took effect
  and the redo/send queues are draining.
- **Target unhealthy / no lag report** then the switchover refuses a target whose
  `synchronization_health` is not `HEALTHY`; ensure the target AG primary is up and the DAG
  forwarder is streaming.
- **Target legitimately too far behind** then let the redo queue drain, or relieve the
  cross-DC bottleneck (scenario 9) before retrying.
- **Abort** then remove the annotation to cancel:
  `kubectl annotate mssqlserver -n demo mssql-dcdr dr.kubedb.com/switchover-to-`.

---

## 9. Lag growing on a standby DC

**Symptoms:** a DC's `redoQueueBytes` / `logSendQueueBytes` climbs steadily, or
`synchronizationHealth` degrades.

**Diagnose:** cross-DC network throughput/latency, write volume on the active primary, and the
DAG forwarder on the standby DC's AG primary.

```bash
# On any AG replica:
#   SELECT synchronization_health_desc, redo_queue_size, log_send_queue_size, last_hardened_lsn
#   FROM sys.dm_hadr_database_replica_states;
```

**Action:** relieve the bottleneck (network, primary load). High lag widens the RPO of an
unplanned failover and can block a planned switchover until the redo queue drains.

---

## 10. A DC is unexpectedly read-only (fence tripped)

**Symptoms:** a DC you expect to be active holds its AG as the DAG secondary; its AG primary
is labeled `standby`.

**Diagnose the fence chain:**

```bash
# Does this spoke's marker name this DC and is renewTime fresh?
kubectl -n dc-failover get configmap primary-dc -o jsonpath='{.data}'
# Is the dr-controlplane agent running in this DC and renewing?
kubectl get pods -n <agent-namespace> -l app=dr-controlplane-agent
```

**Causes & action:**

- **Marker stale** (`renewTime` old) then the agent cannot reach the coordination plane, or
  the projector is failing; restore agent connectivity.
- **Marker names another DC** then this DC simply is not the active one (correct).
- **Clock skew** then large cross-DC clock skew can trip the TTL early; verify NTP. The timing
  budget assumes skew is well under (LeaseDuration minus fence TTL).

Never run a manual DAG failover or patch `kubedb.com/role` by hand to force writability, the
next fence loop re-asserts the DAG secondary role, and you risk split-brain.

---

## 11. Re-add / recover a previously lost data center

After a DC returns from a failure:

**Automatic:** its per-DC AG re-forms, and its AG rejoins the DAG as the secondary
(forwarder). If its databases did not diverge it resumes the DAG stream directly. If they
forked after an unplanned `FORCE_FAILOVER_ALLOW_DATA_LOSS`, SQL Server cannot stream over the
diverged databases, so the operator removes them from the AG and lets DAG automatic seeding
re-seed them over 5022. Once caught up, the DC's AG is a healthy DAG secondary and its lag
appears in status.

**Verify:**

```bash
kubectl get mssqlserver -n demo mssql-dcdr -o jsonpath='{range .status.disasterRecovery.dataCenters[*]}{.clusterName} healthy={.healthy} sync={.synchronizationHealth} redo={.redoQueueBytes}{"\n"}{end}'
```

**Action:** to make it active again, perform a planned failback (scenario 4) once its redo
queue is small.

---

## 12. Scale a data center up or down

```bash
# Scale dc-west to 5 nodes:
kubectl apply -f - <<'YAML'
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata: { name: mssql-dcdr-scale-west, namespace: demo }
spec:
  type: HorizontalScaling
  databaseRef: { name: mssql-dcdr }
  horizontalScaling:
    dataCenters:
    - { clusterName: dc-west, replicas: 5 }
YAML
```

**Verify:** the per-DC PetSet reaches the new size and only that DC changed.

```bash
kubectl get petset -n demo -l app.kubernetes.io/instance=mssql-dcdr
kubectl get mssqlserveropsrequest -n demo mssql-dcdr-scale-west -o jsonpath='{.status.phase}'
```

**Notes:** keep each Member DC's count **odd** for a clean coordinator-raft majority (an even
AG gets an auto-injected local arbiter PetSet); removing a whole DC is a topology change, not
horizontal scaling.

---

## 13. Version upgrade in DC-DR

Issue a normal `UpdateVersion` `MSSQLServerOpsRequest`; the operator updates every per-DC
PetSet. Plan it during a low-traffic window and confirm each DC returns healthy in
`status.disasterRecovery` before relying on failover again.

---

## 14. Suspected split-brain (two writable DCs)

This should be impossible by design (etcd majority + fail-closed fence + the timing
invariant). If `status.disasterRecovery` ever shows two `writable: true` DCs, or two pods
labeled `kubedb.com/role: primary`:

**Diagnose immediately:**

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=mssql-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name
kubectl --kubeconfig <coord> -n dc-failover get lease primary-dc -o jsonpath='{.spec.holderIdentity}'
```

**Action:** the Lease holder is the true active DC. The other DC's fence should force its AG
to the DAG secondary role within the TTL; if it does not, its marker is wrong (check its
agent/projector) or the timing invariant is misconfigured (verify fence TTL < Lease
duration). Stop writes to the non-Lease-holder DC at the application layer until the fence
reasserts, then reconcile (the non-holder rejoins the DAG as secondary, re-seeding the
diverged databases if needed).

---

## Escalation checklist

When unsure, collect:

```bash
kubectl get mssqlserver -n demo mssql-dcdr -o yaml
kubectl --kubeconfig <coord> -n dc-failover get lease -o yaml
kubectl -n dc-failover get configmap primary-dc -o yaml   # on each spoke
kubectl get pods -n demo -l app.kubernetes.io/instance=mssql-dcdr -L kubedb.com/role,open-cluster-management.io/cluster-name -o wide
kubectl logs -n demo <ag-primary-pod> -c mssql-coordinator
```
