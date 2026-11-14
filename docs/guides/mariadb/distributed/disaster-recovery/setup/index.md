---
title: Setup MariaDB DC-DR
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-disaster-recovery-setup
    name: Setup
    parent: guides-mariadb-distributed-disaster-recovery
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Setup MariaDB Cross Data Center Disaster Recovery (DC-DR)

> **New to KubeDB?** Please start [here](/docs/README.md).

This guide walks through deploying a DC-DR enabled distributed MariaDB across two
Member data centers (DCs) plus one Arbiter DC, and verifying that exactly one DC
is writable. Read the
[DC-DR Overview](/docs/guides/mariadb/distributed/disaster-recovery/overview/index.md)
first for the architecture and the concepts referenced below (the `primary-dc`
Lease, the marker fence, role labeling, and the cross-DC asynchronous link).

## Before you begin

DC-DR builds directly on the distributed MariaDB substrate. Complete the
following from the
[Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview/index.md)
before you start here:

- An **OCM** hub with the three participating spoke clusters joined and accepted.
  In this guide they are `dc-a`, `dc-b`, and `dc-c`. The OCM spoke cluster name is
  the DC name and must match the PlacementPolicy `clusterName` exactly.
- The OCM **WorkConfiguration** patch (`RawFeedbackJsonString`) applied on every
  spoke.
- **KubeSlice** installed, a project and `SliceConfig` covering all three
  clusters, and CoreDNS forwarding `*.slice.local` on every cluster.
- The **KubeDB operator** installed on the hub with
  `--set petset.features.ocm.enabled=true`.

In addition, DC-DR requires the cross-DC failover authority:

- The **`dr-controlplane`** three site etcd quorum running behind the OCM control
  plane, with one etcd member in each of `dc-a`, `dc-b`, and `dc-c` (the Arbiter
  DC contributes its vote here).
- The per-DC `dr-controlplane` agent running on each spoke, projecting the
  `primary-dc` marker ConfigMap into the `dc-failover` namespace.
- The KubeDB operator started with the DC-DR flags so its hub orchestrator watches
  the Lease: `--dc-dr-enabled`, `--dc-dr-coord-kubeconfig`, and
  `--dc-dr-local-dc`.

> **Note:** The `dr-controlplane` agent needs write access to ConfigMaps in each
> spoke's `dc-failover` namespace, and the MariaDB coordinator needs read access
> to that ConfigMap from the database namespace. These RBAC rules ship with the
> DC-DR Helm values.

## Step 1: Define the DC-DR PlacementPolicy

The PlacementPolicy is what turns a plain distributed MariaDB into a DC-DR
cluster. Two things matter here:

- `clusterSpreadConstraint.failoverPolicy` with `mode: TwoDC` and
  `trigger.scope: Global`. This declares the two Member DC plus Arbiter DC layout
  and that a single `primary-dc` Lease decides the writable DC for the whole
  cluster.
- A `role` on each `distributionRule`. The two data centers are `role: Member`
  (each becomes a self contained Galera cluster), and the third is `role: Arbiter`
  with an empty `replicaIndices` (no MariaDB data, only the `dr-controlplane` etcd
  vote).

Create `placement-policy.yaml`:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-mariadb-dcdr
spec:
  clusterSpreadConstraint:
    slice:
      projectNamespace: kubeslice-demo-distributed-mariadb
      sliceName: demo-slice
    failoverPolicy:
      mode: TwoDC
      trigger:
        scope: Global
    distributionRules:
      - clusterName: dc-a
        role: Member
        storageClassName: local-path   # optional; omit to use the cluster default
        replicaIndices:
          - 0
          - 1
          - 2
      - clusterName: dc-b
        role: Member
        storageClassName: local-path   # optional; omit to use the cluster default
        replicaIndices:
          - 3
          - 4
          - 5
      - clusterName: dc-c
        role: Arbiter
        replicaIndices: []
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

> **Note:** Each Member DC's `replicaIndices` set becomes one independent Galera
> cluster with its own gcomm peer set and its own quorum. Use an odd count per
> Member DC (3 here) so each local Galera cluster keeps odd quorum without a
> per-DC garbd. A Member DC with an even local node count gets its own intra-DC
> garbd automatically.

Apply the policy on the hub:

```bash
$ kubectl apply -f placement-policy.yaml --context dc-a --kubeconfig $HOME/.kube/config
```

## Step 2: Create the DC-DR MariaDB

Create the `demo` namespace if it does not exist:

```bash
$ kubectl create namespace demo
```

Define the distributed MariaDB and reference the PlacementPolicy. The interim
annotation `dr.kubedb.com/enabled: "true"` enables the DC-DR behavior (this is
transitioning to the PlacementPolicy `failoverPolicy` as the single source of
truth). Create `mariadb.yaml`:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-dcdr
  namespace: demo
  annotations:
    dr.kubedb.com/enabled: "true"
spec:
  distributed: true
  deletionPolicy: WipeOut
  replicas: 6
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  storageType: Durable
  version: 12.1.2
  podTemplate:
    spec:
      podPlacementPolicy:
        name: distributed-mariadb-dcdr
```

`spec.replicas: 6` is partitioned across the Member DCs by the PlacementPolicy:
3 nodes in `dc-a` and 3 in `dc-b`. The Arbiter DC (`dc-c`) carries no MariaDB
data.

Apply the resource on the hub:

```bash
$ kubectl apply -f mariadb.yaml --context dc-a --kubeconfig $HOME/.kube/config
```

The operator expands this one CR into one Galera cluster per Member DC, each with
its own governing ServiceExport, and configures the standby DC's node 0 as a GTID
asynchronous replica of the active DC's primary ServiceExport. The DC that first
acquires the `primary-dc` Lease bootstraps writable; the other Member DC seeds
from it and follows.

## Step 3: Verify exactly one writable DC

### 1. Check which DC holds the Lease

The active DC is whichever spoke holds the `primary-dc` Lease. Inspect the
projected marker ConfigMap on each spoke:

```bash
$ kubectl get configmap primary-dc -n dc-failover -o yaml --context dc-a
$ kubectl get configmap primary-dc -n dc-failover -o yaml --context dc-b
```

The `data.activeDC` value names the active DC and is the same on every spoke. In
this example assume it is `dc-a`.

### 2. Confirm the DR status on the CR

```bash
$ kubectl get mariadb mariadb-dcdr -n demo -o jsonpath='{.status.disasterRecovery}' --context dc-a | jq
```

**Output (abridged):**

```json
{
  "activeDC": "dc-a",
  "phase": "Steady",
  "dataCenters": [
    { "clusterName": "dc-a", "role": "Member", "writable": true,  "healthy": true },
    { "clusterName": "dc-b", "role": "Member", "writable": false, "healthy": true, "lagBytes": 0 },
    { "clusterName": "dc-c", "role": "Arbiter", "healthy": true }
  ]
}
```

Exactly one `dataCenters[]` entry has `writable: true`.

### 3. Confirm role labels resolve only to the active DC

Only the active DC's nodes carry `kubedb.com/role: Primary`; the standby DC's
nodes are `standby`:

```bash
# Active DC nodes are Primary
$ kubectl get pods -n demo -l 'kubedb.com/role=Primary' --context dc-a

# Standby DC nodes are standby
$ kubectl get pods -n demo -l 'kubedb.com/role=standby' --context dc-b
```

Because the single `<db>` primary Service resolves only to the `Primary` labeled
nodes, every client write lands on the active DC.

### 4. Confirm the standby DC is read only and following

Connect to the standby DC's node 0 and confirm the fence and the asynchronous
replica:

```bash
$ kubectl exec -it -n demo pod/mariadb-dcdr-3 --context dc-b -- bash
mariadb -uroot -p$MYSQL_ROOT_PASSWORD
```

```sql
SHOW VARIABLES LIKE 'super_read_only';
SHOW SLAVE STATUS\G
```

`super_read_only` is `ON`, and `SHOW SLAVE STATUS` shows the GTID asynchronous
replica streaming from the active DC's primary endpoint
(`mariadb-dcdr.demo.svc.slice.local`) with both threads running and a small
`Seconds_Behind_Master`.

### 5. Confirm writes are refused on the standby DC

A direct write attempt against the standby DC is rejected by the fence:

```sql
CREATE DATABASE should_fail;
-- ERROR 1290 (HY000): The MariaDB server is running with the --super-read-only option
```

This confirms the fail-closed guarantee: only the Lease holder accepts writes.

## Triggering a planned switchover

To move the active DC on purpose (for example to drain a DC for maintenance) with
zero data loss, set the switchover annotation on the CR. The hub quiesces writes
on the current active DC, waits for the target's GTID to catch up, then moves the
Lease:

```bash
$ kubectl annotate mariadb mariadb-dcdr -n demo \
    dr.kubedb.com/switchover-to=dc-b --overwrite --context dc-a
```

Watch the DR status transition through `FailingOver` back to `Steady` with
`activeDC: dc-b`:

```bash
$ kubectl get mariadb mariadb-dcdr -n demo \
    -o jsonpath='{.status.disasterRecovery.phase} {.status.disasterRecovery.activeDC}{"\n"}' \
    --context dc-a --watch
```

## Cleanup

```bash
$ kubectl delete mariadb mariadb-dcdr -n demo --context dc-a
$ kubectl delete placementpolicy distributed-mariadb-dcdr --context dc-a
```

> **Note:** Per-DC PlacementPolicies and ServiceExports created by the operator
> are cleaned up with the MariaDB. The Arbiter DC's `dr-controlplane` etcd member
> is part of the control plane, not the database, and is not removed by deleting
> the MariaDB.

## Next Steps

- Review the [DC-DR Overview](/docs/guides/mariadb/distributed/disaster-recovery/overview/index.md)
  for failover, failback, and the lag guard semantics.
- See the [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview/index.md)
  for the OCM, KubeSlice, and operator install that DC-DR depends on.
