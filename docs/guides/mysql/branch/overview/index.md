---
title: MySQL Branching Overview
menu:
  docs_{{ .version }}:
    identifier: mysql-branch-overview
    name: Overview
    parent: guides-mysql-branch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL Branching Overview

KubeDB Courier creates MySQL branches from CSI snapshots on CoW storage such as Ceph RBD. Each branch has its own database resource, Service, storage, and auth Secret. Use a branch for development, testing, or data preparation, and optionally refresh it from its source.

Read the common [branching overview](/docs/operatormanual/branching/overview.md) for the architecture and storage requirements. Start with [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation) or [Courier cross-cluster setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm). The tutorials demonstrate:

- [Same-cluster branching into another namespace](/docs/guides/mysql/branch/same-cluster/index.md).
- [Cross-cluster branching](/docs/guides/mysql/branch/cross-cluster/index.md) using the same network-reachable Ceph backend.

## The Branch resource

A Branch is a namespaced `courier.kubedb.com/v1alpha1` resource created on the source cluster. This example copies `demo/mysql-source` to `dev/mysql-dev`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mysql-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MySQL
      name: mysql-source
    namespace: demo
  target:
    namespace: dev
    name: mysql-dev
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
| `spec.source.databaseRef.apiGroup` | `kubedb.com` for a KubeDB MySQL database. |
| `spec.source.databaseRef.kind` | `MySQL`. |
| `spec.source.databaseRef.name` | Name of the source database. |
| `spec.source.namespace` | Source namespace; defaults to the Branch's namespace when omitted. |

The source must be a ready, durable KubeDB MySQL database with snapshot-capable Ceph storage. Put the Branch in the source namespace, as the tutorials do, to keep requests and source status together.

### Target

| Field | Meaning |
|---|---|
| `spec.target.name` | Name of the new MySQL database. Choose an unused name in the target namespace. |
| `spec.target.namespace` | Namespace of the branch database. |
| `spec.target.clusterName` | OCM managed-cluster name for cross-cluster branching. Omit for same-cluster branching. |
| `spec.target.storageClassName` | StorageClass on the target cluster; the examples use `ceph-rbd`. It must resolve compatible snapshot handles. |
| `spec.target.resources` | MySQL container CPU/memory requests and limits for the target. |
| `spec.target.issuerRef` | Target cert-manager Issuer or ClusterIssuer required when the source has TLS enabled. |

Courier prepares target configuration and copies applicable source configuration Secrets. An external ServiceAccount, Secret, or other dependency referenced by your configuration must also be available in the target environment. See [MySQL branch configuration](/docs/guides/mysql/branch/customization/index.md) for TLS and overrides.

### Optional behavior

Use the following independent tutorials to prepare and maintain the branch:

- [Refresh](/docs/guides/mysql/branch/refresh.md): replace branch data periodically using `spec.schedule.cron` and retain run history with `spec.historyLimit`.
- [Reset root password](/docs/guides/mysql/branch/reset-root-password.md): set `spec.resetRootPassword: true` to rotate the branch credential before readiness.
- [Post-actions](/docs/guides/mysql/branch/post-actions.md): use `spec.postActions` to mask or remove sensitive data before the Branch becomes ready.

See [configuration](/docs/guides/mysql/branch/customization/index.md) for snapshot selection, TLS, resource overrides, and deletion policies. The feature tutorials link the common references for field definitions and defaults.

## Status and readiness

Wait for the **Branch** to become `Ready`, even if the MySQL resource becomes ready earlier. Credential isolation and post-actions must finish first.

```shell
kubectl -n demo wait branch/mysql-dev \
  --for=jsonpath='{.status.phase}'=Ready --timeout=15m
kubectl -n demo get branch mysql-dev -o yaml
```

The status reports execution progress through `phase`, `conditions`, and managed `resources`. Same-cluster execution uses `mode: Local`; cross-cluster execution uses `Initiator` on the source and `Creator` on the target. See the common [branch lifecycle](/docs/operatormanual/branching/overview.md#branch-lifecycle) and [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) for details.

For cross-cluster branches, inspect the requested Branch on the source cluster and database resources on the target. The target also has a delivered Branch used by Courier; manage the lifecycle through the original source Branch.
