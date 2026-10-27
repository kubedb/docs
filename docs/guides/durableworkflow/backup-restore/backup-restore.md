---
title: Backup and Restore a DurableWorkflow
description: Back up and restore a DurableWorkflow with KubeStash
menu:
  docs_{{ .version }}:
    identifier: dw-backup-restore-guide
    name: Guide
    parent: dw-backup-restore-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Backup and Restore a DurableWorkflow

All the state of a `DurableWorkflow` (the workflow histories, the queues, the namespaces and the visibility records) is in its persistence stores. The `durableworkflow-addon` of [KubeStash](https://kubestash.com) backs the stores up as one unit and restores them as one unit. You write one `BackupConfiguration` for the `DurableWorkflow`, and the addon runs the KubeStash addon of each store (`postgres-addon`, `mysql-addon`, `cassandra-addon`) for you.

{{< notice type="info" message="Verified on a kind cluster: Temporal and Cadence on a managed Postgres, backed up, deleted together with their database, restored, and provisioned again with every workflow history intact. MySQL and Cassandra stores are not verified yet." >}}

## Before You Begin

- Install KubeDB with the `DurableWorkflow` feature gate, and [KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/) with the `kubedb-kubestash-catalog`, which has the `durableworkflow-addon`.
- A `BackupStorage`, a `RetentionPolicy` and an encryption `Secret` in the `demo` namespace, as in the [KubeStash guides](https://kubestash.com/docs/latest/guides/). This guide calls them `s3-storage`, `keep-last-5` and `encrypt-secret`.
- A `DurableWorkflow` with a store that has a KubeStash addon: PostgreSQL, MySQL or Cassandra. The visibility store may be Elasticsearch or OpenSearch; it is not backed up, it is rebuilt after a restore.

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow/backup-restore](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow/backup-restore) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## How it works

For each backup session the addon reads the stores of the `DurableWorkflow` from its status, where the operator publishes the database each store uses, and starts a backup of each store server with the addon of its engine. Stores that live on the same server, as they do with [managed persistence](/docs/guides/durableworkflow/configuration/managed-persistence.md), are one backup. The backups of the stores are tied together in a manifest that is kept in the snapshot of the `DurableWorkflow`, so a restore finds them all from that one snapshot.

The addon creates a `BackupConfiguration` for each store server, named `<name>-<store>` (or `<name>-stores` when one server holds both stores), with its own repository next to the one you named, in the same backend and with the same retention and encryption secret. You do not need to create or edit them.

## Back up

Create a cluster. This guide uses a [managed Postgres](/docs/guides/durableworkflow/configuration/managed-persistence.md):

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: dw
  namespace: demo
spec:
  version: temporal-1.32.1
  numHistoryShards: 16
  persistence:
    managed:
      engine: Postgres
      replicas: 1
      version: "18.3-bookworm"
      storage:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 2Gi
  adminTools:
    enabled: true
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/backup-restore/dw.yaml
durableworkflow.kubedb.com/dw created
```

Once it is `Ready`, create the `BackupConfiguration`. It targets the `DurableWorkflow` and uses the `coordinated-backup` task of the `durableworkflow-addon`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: dw-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: DurableWorkflow
    name: dw
    namespace: demo
  backends:
  - name: s3
    storageRef:
      name: s3-storage
      namespace: demo
    retentionPolicy:
      name: keep-last-5
      namespace: demo
  sessions:
  - name: daily
    scheduler:
      schedule: "0 2 * * *"
      jobTemplate:
        backoffLimit: 1
    repositories:
    - name: dw-repo
      backend: s3
      directory: /dw
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
    addon:
      name: durableworkflow-addon
      tasks:
      - name: coordinated-backup
    retryConfig:
      maxRetry: 1
      delay: 30s
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/backup-restore/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/dw-backup created
```

KubeStash takes a first backup as soon as the configuration is `Ready`, and then follows the schedule. To take one at once, create a `BackupSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
metadata:
  name: dw-backup-manual-1
  namespace: demo
spec:
  invoker:
    apiGroup: core.kubestash.com
    kind: BackupConfiguration
    name: dw-backup
  session: daily
```

{{< notice type="warning" message="The name of a BackupSession must end in a dash and a number, like dw-backup-manual-1. The KubeStash operator fails on any other name and the session stays Pending." >}}

```bash
$ kubectl get backupsession,snapshot -n demo
NAME                                                          INVOKER-TYPE          INVOKER-NAME   PHASE       DURATION
backupsession.core.kubestash.com/dw-backup-manual-1           BackupConfiguration   dw-backup      Succeeded   55s
backupsession.core.kubestash.com/dw-stores-logical-1792926589 BackupConfiguration   dw-stores      Succeeded   32s

NAME                                                      REPOSITORY      SESSION   SNAPSHOT-TIME          DELETION-POLICY   PHASE
snapshot.storage.kubestash.com/dw-repo-dw-backup-manual-1 dw-repo         daily     2026-10-25T11:10:36Z   Delete            Succeeded
```

The snapshot of the `DurableWorkflow` (`dw-repo-dw-backup-manual-1`) is the one to restore from.

## Restore

A restore never writes into a running cluster. You restore into a `DurableWorkflow` that waits for it: set `spec.init.waitForInitialRestore` to `true`. While it waits, the operator creates no database, schema or pod for it, and only records where its stores are.

For a disaster recovery, delete the cluster and create it again with the same name:

```bash
$ kubectl delete durableworkflow -n demo dw
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: dw
  namespace: demo
spec:
  version: temporal-1.32.1
  numHistoryShards: 16
  init:
    waitForInitialRestore: true
  persistence:
    managed:
      engine: Postgres
      replicas: 1
      version: "18.3-bookworm"
      storage:
        accessModes: [ReadWriteOnce]
        resources:
          requests:
            storage: 2Gi
  adminTools:
    enabled: true
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/backup-restore/dw-restored.yaml
durableworkflow.kubedb.com/dw created
```

The new `Postgres` becomes `Ready`, and the `DurableWorkflow` stays in `Provisioning` with no server pods. Now create the `RestoreSession`, with the snapshot you took:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: dw-restore-1
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: DurableWorkflow
    name: dw
    namespace: demo
  dataSource:
    repository: dw-repo
    snapshot: dw-repo-dw-backup-manual-1
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: durableworkflow-addon
    tasks:
    - name: coordinated-restore
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/backup-restore/restoresession.yaml
restoresession.core.kubestash.com/dw-restore-1 created

$ kubectl get restoresession -n demo
NAME           REPOSITORY   FAILURE-POLICY   PHASE       DURATION   AGE
dw-restore-1   dw-repo                       Succeeded   10s        2m
```

When every store is restored, the addon sets the `DataRestored` condition of the `DurableWorkflow`. The operator then provisions the cluster on top of the restored data: it resets the application user to the password in the cluster's secret, runs the schema job (which finds the restored schema and applies a newer step if the server is newer), and starts the servers.

```bash
$ kubectl get durableworkflow -n demo dw
NAME   VERSION           STATUS   AGE
dw     temporal-1.32.1   Ready    6m
```

The namespaces, the workflow histories (open and closed) and the visibility list are back, and open workflows can be signalled again.

## What a restore checks

The restore is refused, with the reason in the `RestoreSession`, when

- the target is not waiting for a restore (it is running, or `waitForInitialRestore` is not set);
- the target is of another distribution, or has another `numHistoryShards`, which cannot change;
- the target runs an older server version than the backup. A newer one is fine: the schema moves forward;
- a store of the target has another engine, or the two stores sit on servers differently than they did;
- a store of the target uses another database than the backup did. The restored data is under the old name. To restore into a cluster with another name, set `spec.persistence.<store>.databaseName` and `authSecret` to the old names.

## Things to know

- A store server is backed up whole. With [managed persistence](/docs/guides/durableworkflow/configuration/managed-persistence.md) that is only this cluster. If you point several clusters at one shared PostgreSQL, each backup contains all of them.
- Use one repository per session.
- A search visibility store (Elasticsearch, OpenSearch) is not backed up. It is rebuilt after the restore.
- Archived histories in a bucket are not copied, they are already outside the stores.

## Cleaning up

```bash
kubectl delete backupconfiguration -n demo dw-backup
kubectl delete restoresession -n demo dw-restore-1
kubectl delete durableworkflow -n demo dw
```
