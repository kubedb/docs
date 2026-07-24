---
title: MySQL Replication Mode Transform Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-replication-mode-transform-overview
    name: Overview
    parent: guides-mysql-mode-transform
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL Replication Mode Transform

This guide will give an overview on how KubeDB Ops Manager transforms the replication mode of a `MySQL` database — including **promoting a standalone MySQL into a clustered topology** and switching an existing cluster from one topology to another.

Two step-by-step guides build on this overview:

- [MySQL Topology Mode Change](/docs/guides/mysql/replication-mode-transform/topology-mode-change/index.md) — change the mode of an
  existing database: standalone → `GroupReplication` (Single-Primary or Multi-Primary) /
  `InnoDBCluster` / `SemiSync`, and changes between clustered topologies.
- [MySQL Remote/Read Only Replica Mode Transfer](/docs/guides/mysql/replication-mode-transform/remote-replica-mode-transfer/index.md) —
  transform a Remote Replica into a standalone or clustered database.

## Supported Transformations

The target topology is selected with `spec.replicationModeTransformation.targetMode`, which accepts
`GroupReplication` (default), `InnoDBCluster` or `SemiSync`.

| From (source) | To `GroupReplication` | To `InnoDBCluster` | To `SemiSync` |
|---------------|:---------------------:|:------------------:|:-------------:|
| **Standalone** (no `spec.topology`) | ✅ | ✅ | ✅ |
| **RemoteReplica** | ✅ | ✅ | ✅ |
| **GroupReplication** | — | ✅ | ✅ |
| **InnoDBCluster** | ✅ | — | ❌ not supported yet |
| **SemiSync** | ❌ not supported yet | ❌ not supported yet | — |

Notes:

- **Your data is preserved.** Promotions and transformations never delete a volume. When a new
  replica has to be seeded, it is seeded in place with MySQL's `CLONE INSTANCE`, which overwrites
  the data directory while the `PersistentVolumeClaim` is retained.
- **Transformations between clustered topologies happen in place.** `GroupReplication` ⇄
  `InnoDBCluster` keeps the running group and simply hands over management (adopting the group into
  an InnoDB Cluster, or releasing it back to plain Group Replication) — no teardown and no re-clone.
- A standalone database is scaled up to at least 3 members when it is promoted, since a clustered
  topology needs a quorum.
- `spec.replicationModeTransformation.mode` selects the Group Replication primary mode —
  **`Single-Primary`** (default) or **`Multi-Primary`** (multi-master, every member accepts writes).
  It applies to the group-based targets; it is ignored for `SemiSync`.
- Replication Mode Transformation requires MySQL **8.4.2 or newer**.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
    - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)

## How Replication Mode Transform Process Works

The following diagram shows how KubeDB Ops Manager transform replication mode of `MySQL` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
<img alt="Replication Mode Transform process of MySQL" src="/docs/guides/mysql/replication-mode-transform/overview/images/replication-mode-transform.svg">
<figcaption align="center">Fig: Replication Mode Transform process of MySQL</figcaption>
</figure>

The Replication Mode Transform process consists of the following steps:

1. At first, a user creates a `MySQL` Custom Resource (CR).

2. `KubeDB` provisioner operator watches the `MySQL` CR.

3. When the operator finds a `MySQL` CR, it creates required `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to transform replication mode of the `MySQL` database the user creates a `MySQLOpsRequest` CR with desired information.

5. `KubeDB` ops-manager operator watches the `MySQLOpsRequest` CR.

6. When it finds a `MySQLOpsRequest` CR, it pauses the `MySQL` object which is referred from the `MySQLOpsRequest`. So, the `KubeDB` provisioner operator doesn't perform any operations on the `MySQL` object during the mode transform process.

7. Then the `KubeDB` ops-request operator will transform replication mode to reach the expected replication mode defined in the `MySQLOpsRequest` CR.

8. After the successful transformation of replication mode of the MySQL database, the `KubeDB` ops-request operator updates the new replication mode in the `MySQL` object to reflect the updated state. After that, the `KubeDB` ops-request operator resumes the `MySQL` object so that the `KubeDB` provisioner operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on transform replication mode of various MySQL database using `MySQLOpsRequest` CRD.
