---
title: Redis Branching Overview
menu:
  docs_{{ .version }}:
    identifier: redis-branch-overview
    name: Overview
    parent: guides-redis-branch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis Branching Overview

KubeDB Courier creates independent Redis deployments from storage snapshots on copy-on-write storage such as Ceph RBD. Use branches for development, testing, and isolated data preparation. Changes on either side do not propagate to the other; optional refresh replaces branch data with a fresh source snapshot.

The tutorials use a three-node Redis source managed by three Sentinels. The source uses Sentinel for failover. The walkthrough persists the member selected for the snapshot before branching.

Start with the common [branching architecture](/docs/operatormanual/branching/overview.md) and [Courier setup](/docs/operatormanual/branching/setup.md).

- [Branch into another namespace](/docs/guides/redis/branch/same-cluster/index.md).
- [Branch across clusters](/docs/guides/redis/branch/cross-cluster/index.md) sharing a network-reachable Ceph backend.

## The Branch resource

Create the namespaced `courier.kubedb.com/v1alpha1` Branch on the source cluster. This example selects a source in `demo` and a target in `dev`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: redis-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Redis
      name: redis-source
    namespace: demo
  target:
    namespace: dev
    name: redis-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        memory: 1Gi
  deletionPolicy: Delete
```

### Source and target

`source.databaseRef` selects the KubeDB resource; `source.namespace` locates it. `target.name` and `target.namespace` identify the new deployment. Omit `target.clusterName` for the same cluster; for OCM, use the target ManagedCluster name.

`target.storageClassName` must resolve the source snapshots on the same backend. `target.resources` controls the target database containers. Sources require durable storage.

### Optional behavior

- [Scheduled refresh](/docs/guides/redis/branch/refresh.md) periodically replaces branch data.
- [Root password reset](/docs/guides/redis/branch/reset-root-password.md) separates branch credentials from the source.
- [Post-actions](/docs/guides/redis/branch/post-actions.md) prepare branch data before readiness.
- [Configuration](/docs/guides/redis/branch/customization/index.md) covers snapshot classes, TLS, and deletion policies.

## Status and readiness

Snapshot creation, target provisioning, optional credential preparation, and post-actions must finish before Branch readiness. Inspect the original Branch:

```shell
kubectl -n demo get branch redis-dev
kubectl -n demo get branch redis-dev -o yaml
```

`status.snapshot` records the snapshot strategy and generation; `status.resources` references the cloned PVCs, auth Secret, and action Job when present. During refresh the Branch temporarily leaves readiness while its target storage is replaced. The database becoming ready alone does not establish that its post-actions are complete.
