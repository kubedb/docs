---
title: MSSQL Server Branching Overview
menu:
  docs_{{ .version }}:
    identifier: mssqlserver-branch-overview
    name: Overview
    parent: guides-mssqlserver-branch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQL Server Branching Overview

KubeDB Courier creates independent MSSQL Server deployments from CSI storage snapshots on copy-on-write storage such as Ceph RBD. Use a branch to develop against existing data, test changes, or prepare an isolated environment. Writes on either side do not propagate to the other; optional refresh replaces branch data with a fresh source snapshot.

The tutorials use a three-replica Availability Group. Courier clones the member volumes, and KubeDB recovers an independent Availability Group with its own endpoints.

Start with the common [branching architecture](/docs/operatormanual/branching/overview.md) and [Courier setup](/docs/operatormanual/branching/setup.md).

- [Branch into another namespace](/docs/guides/mssqlserver/branch/same-cluster/index.md).
- [Branch across clusters](/docs/guides/mssqlserver/branch/cross-cluster/index.md) sharing a network-reachable Ceph backend.

## The Branch resource

A Branch is a namespaced `courier.kubedb.com/v1alpha1` resource created on the source cluster. This example selects a source in `demo` and a target in `dev`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: mssql-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: MSSQLServer
      name: mssql-source
    namespace: demo
  target:
    namespace: dev
    name: mssql-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: '1'
        memory: 3Gi
      limits:
        memory: 3Gi
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssql-source-ca
  deletionPolicy: Delete
```

### Source and target

`source.databaseRef` selects the KubeDB resource, while `source.namespace` locates it. `target.name` and `target.namespace` identify the new deployment. Omit `target.clusterName` for the same cluster; for OCM, set it to the target ManagedCluster name.

`target.storageClassName` selects storage backed by the same Ceph system as the snapshots. `target.resources` sets database container resources for this example. The source requires durable storage.

Every example provides `target.issuerRef`, because SQL Server requires certificates for its supporting components even when client TLS is disabled. The issuer and its signing Secret must exist in `dev` on the target cluster.

### Optional behavior

- [Scheduled refresh](/docs/guides/mssqlserver/branch/refresh.md) replaces branch data periodically.
- [Root password reset](/docs/guides/mssqlserver/branch/reset-root-password.md) separates branch credentials from the source.
- [Post-actions](/docs/guides/mssqlserver/branch/post-actions.md) prepare the target before readiness.
- [Configuration](/docs/guides/mssqlserver/branch/customization/index.md) covers snapshot classes, TLS, and deletion policies.

## Status and readiness

The Branch progresses through snapshot creation, target provisioning, optional credential reset and actions, then `Ready`. A refresh temporarily leaves readiness while target storage is replaced. Inspect the original Branch for completion:

```shell
kubectl -n demo get branch mssql-dev
kubectl -n demo get branch mssql-dev -o yaml
```

`status.resources` references the cloned PVCs, auth Secret, and action Job when present. `status.snapshot` records the strategy and generation. Readiness includes configured preparation; the target database becoming ready alone does not establish that its post-actions have finished.
