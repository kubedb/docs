---
title: YugabyteDB Point in Time Recovery (Rewind and Clone)
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-pitr-guide
    name: Rewind and Clone
    parent: yugabytedb-pitr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Point in Time Recovery

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

YugabyteDB keeps the history of a database itself, in snapshot schedules, so this method needs **no object storage**. It is separate from the [KubeStash backup](/docs/guides/yugabytedb/backup/kubestash/).

## Declare a schedule

`spec.pitr` lists one schedule per database. Each entry becomes a YugabyteDB snapshot schedule (`create_snapshot_schedule`):

- `database`: the YSQL database or YCQL keyspace to protect.
- `kind`: `YSQL` (default) or `YCQL`. The webhook rejects a kind whose API is not enabled.
- `intervalMinutes`: the time between two snapshots (minimum 1).
- `retentionMinutes`: how far back the database can be rewound; at least `intervalMinutes`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: YugabyteDB
metadata:
  name: yb-pitr
  namespace: demo
spec:
  version: 2025.2.8
  # One snapshot schedule per database.
  pitr:
    - database: yugabyte
      kind: YSQL
      intervalMinutes: 60
      retentionMinutes: 1440
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/pitr/yugabytedb-pitr.yaml
```

`status.pitr` shows for every database the `scheduleID` and the `lastSnapshotTime`.

## Limits (from YugabyteDB)

- **At most one schedule per database.** `spec.pitr` is keyed by `database`.
- **No rewind past an `upgrade_ysql`.** A YSQL catalog upgrade (the end of an upgrade finalization) is a barrier for the history. During an upgrade KubeDB deletes the schedules (`DeleteSnapshotSchedules`) and recreates them after finalize or rollback, so do not count on rewinding to a time before the upgrade.
- **Roles and tablespaces are not rewound.**
- **A clone needs YugabyteDB 2024.2 or later.** Both catalog versions in the KubeDB catalog qualify.
- While a database is restored, it is unavailable to writes (the webhook warns about this).

## Rewind

A `Rewind` request restores a database in place with `restore_snapshot_schedule`. Set exactly one of `restoreTime` and `minus`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-rewind
  namespace: demo
spec:
  type: Rewind
  databaseRef:
    name: yb-demo
  rewind:
    database: yugabyte
    # Set exactly one of minus and restoreTime.
    minus: 15m
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/pitr/rewind.yaml
```

`minus` is relative to the start of the request. Steps: `RewindPreflight`, `RewindResolveTime` (turns `minus` into a time and records it, so a retry restores to the same time), `RewindRestore` and `RewindWaitRestored`.

The webhook requires: `spec.rewind.database` has an entry in `spec.pitr`; `minus` is positive and not longer than the retention; `restoreTime` is neither in the future nor older than the retention.

## Clone

A `Clone` request copies a YSQL database as of a time into a new database, with `CREATE DATABASE ... TEMPLATE ... AS OF`. The source stays untouched.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-clone
  namespace: demo
spec:
  type: Clone
  databaseRef:
    name: yb-demo
  clone:
    sourceDatabase: yugabyte
    targetDatabase: yugabyte_clone
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/pitr/clone.yaml
```

- `asOf` defaults to now. When `asOf` is set, the source database needs a schedule in `spec.pitr` and the time must be inside its retention and not in the future.
- `targetDatabase` must not exist and must differ from `sourceDatabase`.
- YSQL must be enabled and the version must be 2024.2 or later.
- Steps: `ClonePreflight`, `CloneCreateDatabase`, `CloneWaitReady`.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-rewind yb-clone
```
