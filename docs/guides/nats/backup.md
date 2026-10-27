---
title: Backup and Restore NATS
menu:
  docs_{{ .version }}:
    identifier: nats-backup
    name: Backup
    parent: nats-guides
    weight: 90
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Backup and Restore NATS

Core NATS holds no data, so backup means **JetStream streams**. NATS uses [KubeStash](https://kubestash.com) with the `nats-addon`. KubeDB does not add new backup kinds.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/backup](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/backup) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## What is backed up

- For each stream in the application account, the server's snapshot API writes the stream config and state plus a compressed archive in chunks. Consumer state is included by default.
- Publishers are not blocked while a snapshot runs, but `purge` and `delete` on that stream wait or fail until it ends.
- KV and object-store buckets are streams underneath, so they are covered.
- Accounts and users are part of the `NATS` spec and Secrets, which the `manifest-backup` task covers.
- **PVC volume snapshots are not a supported backup.** One node's disk is only one replica of a Raft group.

## Prerequisites

- Install KubeStash and the KubeDB addons (`nats-addon` is in the `kubedb-kubestash-catalog` chart).
- A `BackupStorage`, a `RetentionPolicy` and an encryption Secret in `demo`.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: http://minio.storage.svc:9000
      bucket: nats-backup
      region: us-east-1
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

## Backup

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: nats-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: NATS
    namespace: demo
    name: nats-cluster
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/30 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-nats-repo
          backend: s3-backend
          directory: /nats
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: nats-addon
        tasks:
          - name: logical-backup
            params:
              includeConsumers: "true"
          - name: manifest-backup
```

`logical-backup` parameters: `streams` (comma-separated, default all), `excludeStreams`, `includeConsumers` (default `true`), `chunkSize` and `windowSize` (lower them on a slow link), and `jsDomain` if the server has one.

### Operator mode

For `auth.mode: Operator` the key Secrets (`<name>-operator`, `<name>-sys-creds`, `<name>-auth`, every `<name>-account-<acct>`) are part of the backup and exist nowhere else. The `manifest-backup` task is therefore **mandatory**: the backup is refused without it. When the operator is externally managed (`operator.externalSecret`), those keys are yours to keep and KubeStash only warns. The repository then holds private keys, so treat its encryption password and storage access as key-grade secrets.

## Restore

A stream can only be restored if it does not exist, so restore into a fresh `NATS` object (or one whose streams are deleted).

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: nats-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: NATS
    namespace: demo
    name: nats-restored
  dataSource:
    repository: s3-nats-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: nats-addon
    tasks:
      - name: manifest-restore
      - name: logical-backup-restore
```

Streams are recreated with their sequence numbers and consumer positions. In Operator mode the key Secrets are recreated **before** the streams and the config, and the restore fails if they are missing, because a new operator key would reject every existing account JWT. The `logical-backup-restore` parameter `overwriteKeys` (default `true`) replaces the target's key Secrets with the snapshot's, `waitTimeout` (seconds, default `300`) waits for the target to be ready, and `jsDomain` is as above.

## Cleanup

```bash
$ kubectl delete -n demo backupconfiguration/nats-backup restoresession/nats-restore
```
