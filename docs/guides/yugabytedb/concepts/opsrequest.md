---
title: YugabyteDBOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-opsrequest-concepts
    name: YugabyteDBOpsRequest
    parent: yugabytedb-concepts
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDBOpsRequest

## What is YugabyteDBOpsRequest

`YugabyteDBOpsRequest` (`ops.kubedb.com/v1alpha1`, short name `ybops`) is the CRD for day-2 operations on a KubeDB-managed YugabyteDB. KubeDB Ops-manager performs the operation, and every operation that restarts pods uses one rolling engine: masters first (the current leader last), then tservers one at a time, with a health check between pods.

## Spec

Like any Kubernetes resource a `YugabyteDBOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status`.

> Note: YAML files used in the guides are stored in [docs/examples/yugabytedb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### spec.databaseRef

The name of the `YugabyteDB` in the same namespace.

### spec.type

One of `Restart`, `ReconfigureTLS`, `RotateAuth`, `Reconfigure`, `HorizontalScaling`, `VerticalScaling`, `VolumeExpansion`, `UpdateVersion`, and the phase 2 types `FinalizeUpgrade`, `RollbackUpgrade`, `StorageMigration`, `Rewind`, `Clone` (not verified on a cluster). The matching field below must be set.

### spec.apply

`IfReady` (default) runs the operation only while the database is `Ready`; `Always` runs it regardless.

### spec.timeout and spec.maxRetries

`timeout` is the timeout of each step of the operation. If a step does not finish in time the request fails. `maxRetries` defaults to 1.

### spec.rolling

Tunes operations that restart pods one at a time:

- `masterWait`: pause after a master is back and caught up (default 60s).
- `tserverWait`: pause after a tserver is back and caught up (default 60s).
- `leaderDrainTimeout`: how long to wait for tablet leaders to leave a tserver before it is stopped (default 120s). The operation does not fail after the timeout; it records a `LeaderDrainTimedOut` condition and goes on, and the remaining leaders fail over by Raft election.

### Per-type fields

| Type | Field | Notes |
|------|-------|-------|
| Restart | `restart` | no fields needed |
| VerticalScaling | `verticalScaling.master`, `verticalScaling.tserver`, `verticalScaling.readReplica` (resources), `mode` (`Restart` default, or `InPlace`: CPU only) | [in-place](/docs/guides/yugabytedb/scaling/vertical-scaling/in-place.md) is not verified |
| HorizontalScaling | `horizontalScaling.tserver` (number of tservers, min 1), `horizontalScaling.readReplica` (min 0) | masters cannot be scaled, their number is the replication factor |
| VolumeExpansion | `volumeExpansion.mode` (`Online` or `Offline`), `master`, `tserver`, `readReplica` (new size) | not verified end to end |
| Reconfigure | `configuration.configSecret`, `configuration.applyConfig`, `configuration.removeCustomConfig`, `configuration.restart` (`auto`, `true`, `false`) | keys: `master.flags`, `tserver.flags`, `ysql_pg.conf`, `ysql_hba.conf`. `connectionManager` ([guide](/docs/guides/yugabytedb/connection-manager/connection-manager.md), not verified) |
| ReconfigureTLS | `tls.issuerRef`, `tls.nodeToNode`, `tls.clientToNode`, `tls.webserver`, `tls.rotateCertificates`, `tls.remove` | [webserver](/docs/guides/yugabytedb/reconfigure-tls/webserver-tls.md) is not verified |
| RotateAuth | `authentication.secretRef` | the username must stay `yugabyte` |
| UpdateVersion | `updateVersion.targetVersion`, `updateVersion.soakDuration` (default 10m), `updateVersion.finalize` (`Auto` default, or `Manual`) | |
| FinalizeUpgrade | `finalizeUpgrade: {}` | [guide](/docs/guides/yugabytedb/update-version/manual-finalize/finalize-upgrade.md), not verified |
| RollbackUpgrade | `rollbackUpgrade: {}` | [guide](/docs/guides/yugabytedb/update-version/manual-finalize/rollback-upgrade.md), not verified |
| StorageMigration | `migration.storageClassName`, `migration.oldPVReclaimPolicy`; `spec.timeout` is required | [guide](/docs/guides/yugabytedb/storage-migration/storage-migration.md), not verified |
| Rewind | `rewind.database`, one of `rewind.restoreTime` and `rewind.minus` | [guide](/docs/guides/yugabytedb/pitr/pitr.md), not verified |
| Clone | `clone.sourceDatabase`, `clone.targetDatabase`, `clone.asOf` | [guide](/docs/guides/yugabytedb/pitr/pitr.md), not verified |

## Sample manifests

**Restart:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: yb-demo
  apply: IfReady
  rolling:
    masterWait: 20s
    tserverWait: 20s
```

**VerticalScaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef: {name: yb-demo}
  verticalScaling:
    master:
      resources:
        requests: {cpu: 500m, memory: 1280Mi}
        limits: {memory: 1280Mi}
    tserver:
      resources:
        requests: {cpu: "1", memory: 2560Mi}
        limits: {memory: 2560Mi}
  rolling: {masterWait: 20s, tserverWait: 20s}
```

**HorizontalScaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-hscale-out
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef: {name: yb-demo}
  horizontalScaling:
    tserver: 6
```

**UpdateVersion:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-up-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef: {name: yb-up}
  updateVersion:
    targetVersion: 2025.2.8
    soakDuration: 2m
  rolling: {masterWait: 20s, tserverWait: 20s}
```

**VolumeExpansion (not verified, see [Volume Expansion](/docs/guides/yugabytedb/volume-expansion/volume-expansion.md)):**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: yb-demo
  volumeExpansion:
    mode: Online
    tserver: 8Gi
```

The manifests for Reconfigure, ReconfigureTLS and RotateAuth are in their own guides.

## Status

`status.phase` is `Pending`, `Progressing`, `Successful` or `Failed` (also `WaitingForApproval`, `Approved`, `Denied`, `Skipped`). `status.conditions` has one condition per step, for example a Restart shows `Restart`, `RollMasters`, `RollTServers` and `Successful`. Progress of the rolling engine is checkpointed in the OpsRequest, so an Ops-manager restart resumes where it stopped.

```bash
$ kubectl get ybops -n demo
```

## Next Steps

- [Restart](/docs/guides/yugabytedb/restart/restart.md), [Vertical Scaling](/docs/guides/yugabytedb/scaling/vertical-scaling/scale-vertically/), [Horizontal Scaling](/docs/guides/yugabytedb/scaling/horizontal-scaling/scale-horizontally/)
- [Reconfigure](/docs/guides/yugabytedb/reconfigure/reconfigure.md), [Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md), [Rotate Auth](/docs/guides/yugabytedb/rotate-auth/rotateauth.md), [Update Version](/docs/guides/yugabytedb/update-version/versionupgrading/)
