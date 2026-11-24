---
title: Update the Version of a DC-DR Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-update-version
    name: Update Version
    parent: guides-postgres-distributed
    weight: 72
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update the Version of a DC-DR Postgres

A DC-DR Postgres is updated with an `UpdateVersion` PostgresOpsRequest, as a single-cluster Postgres is. What happens depends on the kind of update:

- A **minor** update (for example `17.2` to `17.9`) restarts every pod on the new image, one at a time. You do nothing but apply the request.
- A **major** update (for example `17` to `18`) upgrades the active primary with `pg_upgrade` and re-seeds every other pod from it. It needs **manual steps on the spokes**, and while it runs the standby data center holds no usable copy of the database. Read the whole [major update](#major-update) section before you start one.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/distributed/update-version/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/update-version/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Before You Begin

- You need a DC-DR Postgres that is `Ready`, `Steady` and `protected`, created as in [Provision a DC-DR Postgres](/docs/guides/postgres/distributed/provision/index.md). Check the old data center as well, as described in [Switchover and failover](/docs/guides/postgres/distributed/switchover-failover/index.md#check-before-you-start).
- You should be familiar with the single-cluster [update version](/docs/guides/postgres/update-version/overview/index.md) flow.
- Pick the target `PostgresVersion` from the KubeDB catalog. It must be of the same distribution as the current one, and not deprecated.

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
```

## How the update interacts with failover

- A minor update steps aside when a failover, a failback or a planned switchover to another data center starts: the request fails with a message such as `DC-DR phase is FailingOver; this DC-DR operation steps aside so the cross-DC move is not delayed (re-run it once the phase is Steady)`. Apply it again once the database is `Steady`.
- A major update does not step aside. A planned switchover requested while it runs waits until the update has finished, which can be long, because the update itself waits for you (see below). Do not request a switchover during a major update.
- Nothing stops the hub from failing over during a major update. A failover, or a switchover, before the update has re-seeded every pod fails the request (condition `DCDRMajorActiveDCMoved`) and leaves the database paused. Use the optional [standby-hold](#step-1-optional-hold-the-standby-data-center) to prevent it.

## Minor update

Apply an `UpdateVersion` request with the new minor version:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-docs-update-minor
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: pg-docs
  updateVersion:
    targetVersion: "17.9"
```

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/update-version/yamls/update-version-minor.yaml
postgresopsrequest.ops.kubedb.com/pg-docs-update-minor created
```

KubeDB updates the per data center PetSets and restarts the pods on the new image: the standby data center first, then the active one; inside a data center the replicas first, one at a time, and the Raft leader last, after a leadership transfer. Each pod must be back and streaming before the next one is restarted. The primary data center does not change. The arbiters are restarted at the end if their `pg-coordinator` image changed.

## Major update

### What the request does

Only the active data center's primary is upgraded in place, with `pg_upgrade`. Every other pod is restarted on the new image and re-seeded (a full base backup) from the upgraded primary. In order:

1. It records the active data center and its primary, pins that pod as the Raft leader of the active data center and pauses that data center's `pg-coordinator` failover. The other data center's Raft is not touched.
2. It copies the old binaries on the primary, updates the per data center PetSets to the target version, and restarts the primary on the new image.
3. On the primary, it initializes a new data directory, runs `pg_upgrade --check`, runs `pg_upgrade`, and swaps the data directories. The primary then accepts writes again, on the new major version.
4. It restarts the other pods of the active data center on the new image, one at a time, and each is re-seeded from the upgraded primary.
5. It restarts **all** pods of the standby data center on the new image together, and they are re-seeded from the upgraded primary over the slice.
6. It resumes the coordinators, restarts the arbiters whose `pg-coordinator` image changed, and checks that the standby data center streams from the active primary.

Each phase is checkpointed in a condition of the request, named `DCDRMajor*` (for example `DCDRMajorPlan`, `DCDRMajorLeaderPinned`, `DCDRMajorStep--UPGRADE_STEP_PG_UPGRADE`, `DCDRMajorPrimaryResumed`, `DCDRMajorPod--<pod>`), so you can see how far it got.

**`pg_upgrade` gives the upgraded database a new identity**, a new database system identifier. Within DC-DR, `pg-coordinator` refuses to replace a pod's data with a copy of a database with another identity: this lineage guard protects a data center's data from being overwritten by an unrelated, for example empty, database. Each re-seed in steps 4 and 5 therefore waits until you allow it on that pod's spoke, with an accept ConfigMap. Until you allow it, the pods keep their old data.

> **Warning:** From step 5 until its pods have been re-seeded, **the standby data center runs no Postgres**. Its pods were restarted on the new image and cannot start their old-version data. During that window there is no second copy of the database: if the active data center is lost, nothing can take over. The database reports `protected: false`. Keep the window short: create the accept ConfigMap on the standby spoke as soon as the request asks for it, and make sure the slice bandwidth suffices for a full base backup (see [Set up the spoke clusters](/docs/guides/postgres/distributed/spoke-clusters/index.md)).

Every re-seeded pod keeps its old data directory as a copy until the re-seed has finished, so each data volume, except the primary's, needs free space at least as large as the primary's used space. The request checks this, and the coordinators' support for the upgrade, before anything is paused (condition `DCDRMajorPreflight`), and fails at once otherwise.

### Prerequisites

- The **current** `pg-coordinator` must support the DC-DR major upgrade calls. If it does not, the request fails at once with `the coordinator of pod <pod> does not support the major upgrade calls`. Update the coordinator first, with a minor update to a PostgresVersion of the same major version that ships a newer coordinator.
- The **target** PostgresVersion's `pg-coordinator` image must not be older than the current one, and must support the same calls. A target coordinator that lacks them is detected after the primary's restart; the request then rolls back by itself (the data directory is still untouched), and fails.
- Enough free space on every data volume, as above. Expand the volumes first with an Online volume expansion (`spec.volumeExpansion.mode: Online`) if needed.
- Take a backup.

### Step 1 (optional): hold the standby data center

To make sure the hub cannot move the primary while the update runs, place a [standby-hold](/docs/guides/postgres/distributed/switchover-failover/index.md#keep-a-data-center-from-becoming-primary-standby-hold) on the **standby** data center's spoke. Here `spoke-1` is active, so hold `spoke-2`:

```bash
$ kubectl --kubeconfig $S2 apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/switchover-failover/yamls/standby-hold.yaml
configmap/primary-dc-pg-docs-standby-hold created
```

The trade-off: while the hold exists, the loss of the active data center is not failed over at all. Without the hold, such a loss is failed over, but the update fails and the database stays paused for a manual recovery.

### Step 2: apply the request

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-docs-update-major
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: pg-docs
  updateVersion:
    targetVersion: "18.3"
  apply: IfReady
```

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/update-version/yamls/update-version-major.yaml
postgresopsrequest.ops.kubedb.com/pg-docs-update-major created
```

Watch its conditions:

```bash
$ kubectl --kubeconfig $HUB -n demo get postgresopsrequest pg-docs-update-major -o jsonpath='{range .status.conditions[*]}{.type}={.status} {.message}{"\n"}{end}'
```

### Step 3: allow the re-seeds on each spoke

After the primary has been upgraded (`DCDRMajorPrimaryResumed=True`), the request restarts the next pod and waits for it to be re-seeded. The request does **not** time out here: it waits for as long as it takes you to create the accept ConfigMap. While it waits, it prints the exact command to run, with the cluster to run it on, in the ops-manager log on the hub:

```bash
$ kubectl --kubeconfig $HUB -n kubedb logs kubedb-kubedb-ops-manager-0 -c operator | grep 'kubectl create configmap' | tail -1
... After a major upgrade the coordinators re-seed these pods only once the accept ConfigMap exists: on cluster spoke-1 run: kubectl create configmap accept-foreign-reseed-demo-pg-docs -n dc-failover --from-literal=sourceSystemIdentifier=7711652864947908888
```

The `pg-coordinator` of each waiting pod logs the same command, with `LINEAGE GUARD REFUSED BASE BACKUP; local data untouched`.

Run the command on the spoke it names, in the DC-DR namespace (`dc-failover` in this guide):

```bash
$ kubectl --kubeconfig $S1 create configmap accept-foreign-reseed-demo-pg-docs -n dc-failover \
    --from-literal=sourceSystemIdentifier=<id>
```

- The ConfigMap is named `accept-foreign-reseed-<namespace>-<db>`. It allows the re-seed of this one database on this one spoke, and only from the database whose system identifier is the value of `sourceSystemIdentifier`: the upgraded primary. A ConfigMap without that key, or with another identifier, is ignored (the coordinator logs `accept ConfigMap ignored: it does not name this source lineage`).
- You can read the identifier yourself on the upgraded primary: `SELECT system_identifier FROM pg_control_system();`.
- The request asks for it once per spoke: first on the active data center's spoke, for its other pods, then on the standby data center's spoke. Create it on each spoke the request names.
- The re-seed starts at the coordinator's next attempt. Each re-seed is a full base backup, so its duration grows with the database size; across data centers it is bounded by the slice bandwidth.

Until you create the ConfigMap on the standby data center's spoke, that data center keeps the old version's data on its volumes: it is the fallback copy. If the upgraded primary is not what you expect, do not create the ConfigMap there; see [If the update stops](#if-the-update-stops).

### Step 4: clean up

When every pod has been re-seeded, the request sets its last condition, `DCDRMajorAcceptConfigMapCleanup`, with a reminder, and the request becomes `Successful`:

```
delete ConfigMap accept-foreign-reseed-demo-pg-docs in namespace dc-failover on clusters spoke-1,spoke-2 if it exists: while it exists the coordinators accept a re-seed of this database from the upgraded lineage
```

**Delete the accept ConfigMaps**, and the standby-hold if you placed one:

```bash
$ for k in $S1 $S2; do kubectl --kubeconfig $k -n dc-failover delete cm accept-foreign-reseed-demo-pg-docs --ignore-not-found; done
$ kubectl --kubeconfig $S2 -n dc-failover delete cm primary-dc-pg-docs-standby-hold
```

A leftover accept ConfigMap keeps allowing re-seeds of this database from the upgraded lineage, which weakens the lineage guard. Then check that the database is `Ready`, `Steady` and `protected` again.

### If the update stops

- **Before the primary is restarted on the new image**, a failure (including the preflight checks) undoes the pin, puts the PetSets back on the current version, resumes the database and fails the request. Nothing else changed; fix the cause and apply the request again.
- **If `INIT_NEW_DATA_DIR` or `PG_UPGRADE_CHECK` fails, or the target coordinator lacks the upgrade calls**, the data directory is still untouched: the request rolls back by itself (the primary is restarted on the old image and resumed, condition `DCDRMajorRollback`), then fails.
- **Any later failure** fails the request with the condition `DCDRMajorUpgradeFailed` (or `DCDRMajorActiveDCMoved` after a failover or switchover), and **the database stays paused** for a manual recovery: KubeDB does not reconcile a half-upgraded database. The `DCDRMajor*` conditions show how far the update got: which steps ran on the primary, and which pods (`DCDRMajorPod--<pod>=True`) were re-seeded. Do not delete or edit the Postgres object. Keep the standby data center's old data (do not create the accept ConfigMap on its spoke) and contact AppsCode support with the request's conditions and the ops-manager and `pg-coordinator` logs.

## Cleaning up

```bash
kubectl --kubeconfig $HUB -n demo delete postgresopsrequest pg-docs-update-minor pg-docs-update-major
```

## Next Steps

- [Teardown](/docs/guides/postgres/distributed/teardown/index.md).
