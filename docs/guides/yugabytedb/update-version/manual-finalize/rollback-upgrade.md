---
title: Roll Back a YugabyteDB Upgrade
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-update-version-rollback
    name: Rollback Upgrade
    parent: yugabytedb-update-version-manual
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Roll Back a YugabyteDB Upgrade

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

A `RollbackUpgrade` request returns a database to the version it ran before an `UpdateVersion` request, as long as that upgrade has **not been finalized**. It needs an upgrade started with `spec.updateVersion.finalize: Manual` that is waiting with the `UpgradePendingFinalize` condition (see [Finalize a YugabyteDB Upgrade](/docs/guides/yugabytedb/update-version/manual-finalize/finalize-upgrade.md)). After a `FinalizeUpgrade` there is no rollback.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-rollback-upgrade
  namespace: demo
spec:
  type: RollbackUpgrade
  databaseRef:
    name: yb-demo
  rollbackUpgrade: {}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/update-version/manual-finalize/rollback-upgrade.yaml
```

## What it does

The order is the opposite of an upgrade. These are the step names of the API:

1. `RollbackAutoFlags`: `rollback_auto_flags` returns the AutoFlags to the configuration version recorded by `RecordAutoFlagsConfig` at the start of the upgrade.
2. `RollbackTServers`: the tservers return to the previous image, one at a time.
3. `RollbackMasters`: the masters return to the previous image, one at a time, after the tservers.
4. `RecreateSnapshotSchedules`: only when `spec.pitr` has schedules.
5. `RestoreVersion`: `spec.version` is set back to the previous version.

Like the other rolling operations, the pods restart one at a time.

## Rules checked by the webhook

- The database must have a pending upgrade, otherwise the request is rejected.
- Only one upgrade related request (`UpdateVersion`, `FinalizeUpgrade`, `RollbackUpgrade`) can run at a time.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-rollback-upgrade
```
