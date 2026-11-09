---
title: Database Branching Overview
menu:
  docs_{{ .version }}:
    identifier: branching-overview
    name: Overview
    parent: branching
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Database Branching Overview

A database branch is a separate database created from the schema and data captured in a source database's storage snapshot. Developers, tests, and automation can change the branch's tables and records while the source continues to serve its own workload. Each KubeDB branch has its own database resource, Service, and auth Secret.

KubeDB Courier manages this process through a `Branch` resource in the source cluster. It snapshots the source storage, creates target volumes from those snapshots, and asks KubeDB to provision the branch database. Courier completes any requested password reset and post-actions before marking the Branch `Ready`.

## Why branching matters

- **Validate changes against realistic data.** Test migrations and application behavior against existing rows, constraints, and relationships before deploying them.
- **Keep development and CI independent.** Give each developer or test run its own database so concurrent work does not modify shared test data or schemas.
- **Make experiments disposable.** Recreate a branch after an unsuccessful backfill or destructive test instead of repairing a shared development database.
- **Give AI agents an isolated workspace.** Agents can execute generated SQL and evaluate schema changes on a branch. With production credentials, an incorrect query could delete records, alter tables, or read confidential information.

## Storage requirements

Use CSI storage that supports snapshots and copy-on-write (CoW) clones, such as **Ceph RBD**. A CoW clone initially references the snapshot's existing blocks, avoiding an upfront copy of the entire data volume. As source and branch writes diverge, the backend keeps the block versions needed by the snapshot and allocates space for changes. The branch continues to read its snapshot baseline until its own writes or a refresh replace that data. The PostgreSQL examples use Ceph RBD for both same-cluster and cross-cluster branching.

For **cross-cluster branching**, both Kubernetes clusters must use the **same storage backend, reachable over the network from both clusters**. Installing Ceph independently in each cluster does not provide a shared snapshot backend. Both CSI installations must resolve the source snapshot handle against the same backend and have the credentials needed to use it. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

Courier transfers Kubernetes resource definitions and snapshot references between clusters. Storage access continues through the shared backend; Courier does not perform a database dump and upload.

## Same-cluster flow

1. Create a Branch referencing a source KubeDB database.
2. Courier takes source volume snapshots and waits for them to become usable.
3. Courier creates the target volumes. For another namespace, it imports snapshot references into the target namespace first.
4. KubeDB starts the target database using the cloned storage.
5. Courier completes optional credential isolation and post-actions, then marks the Branch `Ready`.

Source and target can be in different namespaces. The [PostgreSQL same-cluster tutorial](/docs/guides/postgres/branch/same-cluster/index.md) demonstrates this arrangement.

## Cross-cluster flow

Cross-cluster branching uses an existing OCM hub and two registered managed clusters. Courier runs an addon manager on the hub and an agent on each participating managed cluster.

1. Create the Branch on the source managed cluster, naming the target managed cluster.
2. The source agent snapshots the data and sends a `BranchWork` to its namespace on the hub.
3. The Courier addon manager delivers the target resources through OCM `ManifestWork`.
4. The target agent imports the snapshot references and creates the branch against the shared storage backend.
5. Target status returns through the hub to the source Branch.

The hub coordinates resource delivery. KubeDB and the database workloads run on the managed clusters. See [Courier cross-cluster setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm) and the [PostgreSQL cross-cluster tutorial](/docs/guides/postgres/branch/cross-cluster/index.md).

## Branch lifecycle

A Branch progresses through `Pending`, `Snapshotting`, `Cloning`, and `Provisioning`. If post-actions are configured, it enters `ActionsRunning` before `Ready`. A scheduled branch later enters `Refreshing` while its data is replaced.

Source and branch accept independent writes. Ongoing source changes reach the branch only through a [refresh](/docs/operatormanual/branching/refresh.md). Refresh replaces the branch's data and interrupts its availability during volume replacement and restart.
