---
title: PostgreSQL Branching Overview
menu:
  docs_{{ .version }}:
    identifier: pg-branch-overview
    name: Overview
    parent: pg-branch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL Branching Overview

KubeDB Courier creates PostgreSQL branches from CSI snapshots on CoW storage such as Ceph RBD. Each branch has its own database resource, Service, storage, and auth Secret. Use a branch for development, testing, or data preparation, and optionally refresh it from its source.

Start with [Courier setup](/docs/operatormanual/branching/setup.md) or [Courier cross-cluster setup](/docs/operatormanual/branching/setup-ocm.md). The tutorials demonstrate:

- [Same-cluster branching into another namespace](/docs/guides/postgres/branch/same-cluster.md).
- [Cross-cluster branching](/docs/guides/postgres/branch/cross-cluster.md) using the same network-reachable Ceph backend.

## The Branch resource

A Branch is a namespaced `courier.kubedb.com/v1alpha1` resource created on the source cluster. This example copies `demo/prod-pg` to `dev/pg-dev`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
    namespace: demo
  target:
    namespace: dev
    name: pg-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  deletionPolicy: Delete
```

### Source

| Field | Meaning |
|---|---|
| `spec.source.databaseRef.apiGroup` | `kubedb.com` for a KubeDB PostgreSQL database. |
| `spec.source.databaseRef.kind` | `Postgres`. |
| `spec.source.databaseRef.name` | Name of the source database. |
| `spec.source.namespace` | Source namespace; defaults to the Branch's namespace when omitted. |

The source must be a ready, durable KubeDB PostgreSQL database with snapshot-capable Ceph storage. Put the Branch in the source namespace, as the tutorials do, to keep requests and source status together.

### Target

| Field | Meaning |
|---|---|
| `spec.target.name` | Name of the new PostgreSQL database. Choose an unused name in the target namespace. |
| `spec.target.namespace` | Namespace of the branch database. |
| `spec.target.clusterName` | OCM managed-cluster name for cross-cluster branching. Omit for same-cluster branching. |
| `spec.target.storageClassName` | StorageClass on the target cluster; the examples use `ceph-rbd`. It must resolve compatible snapshot handles. |
| `spec.target.resources` | PostgreSQL container CPU/memory requests and limits for the target. |
| `spec.target.issuerRef` | Target cert-manager Issuer or ClusterIssuer required when the source has TLS enabled. |

Courier prepares target configuration and copies applicable source configuration Secrets. An external ServiceAccount, Secret, or other dependency referenced by your configuration must also be available in the target environment. See [PostgreSQL branch configuration](/docs/guides/postgres/branch/configuration.md) for TLS and overrides.

### Optional behavior

| Field | Meaning |
|---|---|
| `spec.resetRootPassword` | Generate an independent branch root credential; defaults to `false`. |
| `spec.postActions` | Ordered list of action images and Job settings to run against the target before the Branch becomes `Ready`. |
| `spec.schedule.cron` | Standard five-field cron expression for periodic refresh. Omit for a one-time branch. |
| `spec.historyLimit.success` | Successful status-history entries to retain; defaults to `3`. |
| `spec.historyLimit.failed` | Failed status-history entries to retain; defaults to `2`. |
| `spec.volumeSnapshotClassName` | Explicit VolumeSnapshotClass selection for source and importing snapshots. |
| `spec.deletionPolicy` | `Delete` removes the branch database and managed resources; `Orphan` retains the database, its PVCs, and required Secrets. Defaults to `Delete`. |

Post-action entries support `image`, `jobDefaults`, and `jobTemplate`. The [post-action tutorial](/docs/guides/postgres/branch/post-actions.md) explains this shape with runnable SQL. Deletion policies and TLS examples are covered in [configuration](/docs/guides/postgres/branch/configuration.md).

## Status and readiness

Wait for the **Branch** to become `Ready`, even if the PostgreSQL resource becomes ready earlier. Credential isolation and post-actions must finish first.

```bash
kubectl -n demo wait branch/pg-dev \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo get branch pg-dev -o yaml
```

| Status field | Meaning |
|---|---|
| `phase` | Current step, including `Snapshotting`, `Cloning`, `Provisioning`, `ActionsRunning`, `Ready`, `Refreshing`, `Failed`, or `Deleting`. |
| `mode` | `Local` for same-cluster execution, `Initiator` on a cross-cluster source, or `Creator` on its target. |
| `conditions` | Readiness and failure details for individual steps. |
| `resources` | Managed object names, such as cloned PVCs, auth Secret, and post-action Job. |
| `snapshot` | Snapshot strategy, readiness, and source-volume provenance. |
| `refreshGeneration` | Current refresh generation. |
| `lastRefreshTime`, `lastSuccessfulRefreshTime`, `nextRefreshTime` | Attempt, completion, and scheduling timestamps. |
| `history` | Bounded run history. |

For cross-cluster branches, inspect the requested Branch on the source cluster and database resources on the target. The target also has a delivered Branch used by Courier; manage the lifecycle through the original source Branch.
