---
title: Finalize a YugabyteDB Upgrade
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-update-version-finalize
    name: Finalize Upgrade
    parent: yugabytedb-update-version-manual
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Finalize a YugabyteDB Upgrade

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

By default an `UpdateVersion` request finalizes the upgrade itself, after the soak (see the [overview](/docs/guides/yugabytedb/update-version/overview.md)). Finalizing is the point of no return: it promotes the AutoFlags of the new version for good (`promote_auto_flags kExternal`) and then upgrades the YSQL catalog (`upgrade_ysql`). After it, the database cannot go back to the previous version. If you want a long monitoring window on the new version, ask for a manual finalize.

## Start an upgrade that waits

Set `spec.updateVersion.finalize: Manual` (the default is `Auto`). The `soakDuration` still applies, and the request then stops with the `UpgradePendingFinalize` condition on the database.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-update-version-manual
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: yb-demo
  updateVersion:
    targetVersion: 2025.2.8
    soakDuration: 30m
    # Manual stops after the soak with the UpgradePendingFinalize condition on the database.
    # A FinalizeUpgrade or RollbackUpgrade request completes the upgrade.
    finalize: Manual
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/update-version/manual-finalize/update-version-manual.yaml
```

Conditions of such a request, in order (names from the API constants): `Preflight`, `DeleteSnapshotSchedules` (only when `spec.pitr` has schedules, because YugabyteDB wants them removed during an upgrade), `RecordAutoFlagsConfig`, the PetSet update, `RollMasters`, `RollTServers`, `PromoteAutoFlags`, `Soak` and `UpgradePendingFinalize`.

While the upgrade is pending:

- The database runs the new binaries, but the new AutoFlags are only promoted in the volatile form, so a rollback is still possible.
- A second `UpdateVersion` request is rejected by the webhook ("an upgrade of this database is pending finalization, send a FinalizeUpgrade or RollbackUpgrade request first"). Only one upgrade related request can run at a time.

The example assumes that `yb-demo` runs `2025.1.4`; the webhook rejects a target version that is the current version.

## Finalize

`FinalizeUpgrade` has no fields; it acts on the upgrade that is pending on the database.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-finalize-upgrade
  namespace: demo
spec:
  type: FinalizeUpgrade
  databaseRef:
    name: yb-demo
  finalizeUpgrade: {}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/update-version/manual-finalize/finalize-upgrade.yaml
```

The steps are `FinalizeUpgrade`, `UpgradeYSQL` and, if the database has `spec.pitr`, `RecreateSnapshotSchedules`. The new version is then recorded in `spec.version`.

The webhook rejects a `FinalizeUpgrade` when the database has no pending upgrade ("the database has no `UpgradePendingFinalize` condition, run an UpdateVersion with `spec.updateVersion.finalize: Manual` first").

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-update-version-manual yb-finalize-upgrade
```
