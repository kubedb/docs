---
title: Snapshot Backup & Restore Etcd | KubeStash
description: Take scheduled snapshot backups of an Etcd cluster with KubeStash and bootstrap a new cluster from one
menu:
  docs_{{ .version }}:
    identifier: guides-etcd-backup-snapshot-stashv2
    name: Snapshot Backup & Restore
    parent: guides-etcd-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Snapshot Backup & Restore of Etcd using KubeStash

This guide walks through the full lifecycle: deploying an `Etcd` cluster, attaching an
`EtcdArchiver` backup policy to it, watching a snapshot land in an S3 bucket, and then bootstrapping
a brand new `Etcd` cluster from that snapshot.

{{< notice type="warning" message="Backup of an `Etcd` cluster is **snapshot-only**. There is no continuous archiving and therefore no point-in-time recovery — a restore rolls the data back to the state captured by one completed snapshot. Read the [overview](/docs/guides/etcd/backup/kubestash/overview/index.md) before you design a backup schedule." >}}

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured to talk to it. If you do not have one, you
  can create one with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md). etcd support is
  an alpha feature gate, so it must be enabled explicitly — pass `--set featureGates.Etcd=true` to
  the KubeDB Helm chart (equivalently `--feature-gates=Etcd=true` on the operator binaries).
- Install `KubeStash` in your cluster following the steps
  [here](https://kubestash.com/docs/latest/setup/install/kubestash/).
- Install the KubeStash `kubectl` plugin following the steps
  [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- Read the [Etcd Backup & Restore Overview](/docs/guides/etcd/backup/kubestash/overview/index.md)
  first — especially the part about restore being gated at bootstrap time.

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)

To keep everything isolated, this tutorial uses a separate namespace called `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** the YAML files used in this tutorial are stored in
> [docs/examples/etcd/backup](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/backup)
> of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Prepare the Backend

We are going to store the backed up data in an `S3` bucket. Any other KubeStash backend works the
same way — see the [backend configuration docs](https://kubestash.com/docs/latest/guides/backends/overview/).

### Create the storage Secret

```bash
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

### Create the BackupStorage

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
      endpoint: ap-south-1.linodeobjects.com
      bucket: kubestash-etcd
      region: ap-south-1
      prefix: etcd-demo
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/backup/s3-storage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

### Create a RetentionPolicy

The `RetentionPolicy` decides how many old snapshots survive.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: demo-retention
  namespace: demo
spec:
  default: true
  failedSnapshots:
    last: 2
  maxRetentionPeriod: 2mo
  successfulSnapshots:
    last: 5
  usagePolicy:
    allowedNamespaces:
      from: Same
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/backup/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Create the encryption Secret

KubeStash encrypts everything it writes to the backend with a restic password.

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret/encrypt-secret created
```

## Deploy a Sample Etcd

Below is the `Etcd` object we are going to back up. Two things are worth pointing out: the
`archiver: "true"` label, which is what the archiver's `spec.databases.selector` will match, and
`spec.archiver.ref`, which is the database side of the opt-in.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: sample-etcd
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  archiver:
    ref:
      name: etcd-archiver
      namespace: demo
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/backup/sample-etcd.yaml
etcd.kubedb.com/sample-etcd created
```

> If you omit `spec.archiver` entirely, the provisioner will find a matching `EtcdArchiver` on its
> own and patch the reference in for you. Setting it explicitly just makes the intent obvious.

Wait for the cluster to become `Ready`:

```bash
$ kubectl get etcd -n demo sample-etcd
NAME          VERSION   STATUS   AGE
sample-etcd   3.6.4     Ready    4m12s
```

KubeDB also created the auth `Secret`, the Services and the `AppBinding` that the backup job will use
to reach the cluster:

```bash
$ kubectl get secret,svc,appbinding -n demo -l app.kubernetes.io/instance=sample-etcd
NAME                      TYPE                       DATA   AGE
secret/sample-etcd-auth   kubernetes.io/basic-auth   2      4m30s

NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/sample-etcd        ClusterIP   10.96.121.40    <none>        2379/TCP                     4m30s
service/sample-etcd-pods   ClusterIP   None            <none>        2379/TCP,2380/TCP,2381/TCP   4m30s

NAME                                             TYPE              VERSION   AGE
appbinding.appcatalog.appscode.com/sample-etcd   kubedb.com/etcd   3.6.4     4m28s
```

The [AppBinding](/docs/guides/etcd/concepts/appbinding.md) is what the backup job reads to learn the client endpoint (`spec.clientConfig.service`
on port `2379`), the credentials (`spec.secret`) and, when TLS is enabled, the client certificate
(`spec.tlsSecret`).

### Put some data in the cluster

Read the root credentials out of the auth `Secret`:

```bash
$ kubectl get secret -n demo sample-etcd-auth -o jsonpath='{.data.username}' | base64 -d
root
$ kubectl get secret -n demo sample-etcd-auth -o jsonpath='{.data.password}' | base64 -d
```

Then exec into a member and write a key so we have something to verify after the restore:

```bash
$ kubectl exec -it -n demo sample-etcd-0 -c etcd -- sh

$ export ETCDCTL_API=3
$ export ETCDCTL_ENDPOINTS=http://127.0.0.1:2379
$ export ETCDCTL_USER=root:<root-password>

$ etcdctl put /demo/hello "hello from the original cluster"
OK

$ etcdctl get /demo/hello
/demo/hello
hello from the original cluster

$ exit
```

## Configure Backup

Instead of writing a `BackupConfiguration` by hand, you create an `EtcdArchiver` — the backup
*policy* — and the KubeDB provisioner derives the `BackupConfiguration` from it.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: EtcdArchiver
metadata:
  name: etcd-archiver
  namespace: demo
spec:
  databases:
    namespaces:
      from: Same
    selector:
      matchLabels:
        archiver: "true"
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
    subDir: etcd-archiver
  retentionPolicy:
    name: demo-retention
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: Restic
    scheduler:
      successfulJobsHistoryLimit: 3
      failedJobsHistoryLimit: 3
      schedule: "0 */2 * * *"
      jobTemplate:
        backoffLimit: 1
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 3
      failedJobsHistoryLimit: 3
      schedule: "0 */2 * * *"
      jobTemplate:
        backoffLimit: 1
  deletionPolicy: WipeOut
```

Here,

- `spec.databases` is the archiver side of the double opt-in. `namespaces.from: Same` restricts it to
  the archiver's own namespace, and `selector` narrows it to `Etcd` objects labelled `archiver: "true"`.
- `spec.backupStorage.ref` points at the `BackupStorage`. `subDir` is prefixed to the repository
  directory, so the data ends up under `etcd-archiver/demo/sample-etcd/<session>` in the bucket.
- `spec.fullBackup` generates the `full-backup` session, which takes the etcd snapshot. Its
  `driver` field is required by the shared schema but is **ignored for etcd** — an etcd snapshot is
  always streamed through the client API into a restic repository.
- `spec.manifestBackup` generates the `manifest-backup` session, which captures the Kubernetes
  objects (the `Etcd` manifest, the auth `Secret`, and so on).
- `spec.deletionPolicy` decides what happens to the repositories when the backup is torn down. It
  defaults to `Retain`.

> **There is no `spec.logBackup`.** etcd has no WAL-shipping primitive, so the field does not exist.
> Your recovery point is whatever the last completed snapshot captured.

Create it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/backup/etcd-archiver.yaml
etcdarchiver.archiver.kubedb.com/etcd-archiver created
```

### Verify the generated BackupConfiguration

The provisioner creates a `BackupConfiguration` named `<db-name>-archiver`, owned by the `Etcd`
object:

```bash
$ kubectl get backupconfiguration -n demo
NAME                   PHASE   PAUSED   AGE
sample-etcd-archiver   Ready            1m20s
```

Its two sessions and two repositories match the two sections of the archiver:

```bash
$ kubectl get backupconfiguration -n demo sample-etcd-archiver \
    -o jsonpath='{range .spec.sessions[*]}{.name}{"\t"}{.repositories[0].name}{"\t"}{range .addon.tasks[*]}{.name}{" "}{end}{"\n"}{end}'
full-backup       sample-etcd-full       etcd-backup manifest-backup
manifest-backup   sample-etcd-manifest   manifest-backup
```

The `Repository` objects are initialized as soon as the `BackupConfiguration` becomes `Ready`:

```bash
$ kubectl get repository -n demo
NAME                   INTEGRITY   SNAPSHOT-COUNT   SIZE   PHASE   LAST-SUCCESSFUL-BACKUP   AGE
sample-etcd-full                   0                0 B    Ready                            1m35s
sample-etcd-manifest               0                0 B    Ready                            1m35s
```

And a `CronJob` per session fires on the schedule you set:

```bash
$ kubectl get cronjob -n demo
NAME                                           SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-etcd-archiver-full-backup       0 */2 * * *   False     0        <none>          1m45s
trigger-sample-etcd-archiver-manifest-backup   0 */2 * * *   False     0        <none>          1m45s
```

### Watch the first BackupSession

KubeStash triggers an instant backup as soon as the `BackupConfiguration` becomes `Ready`; after
that the `CronJob`s take over.

```bash
$ kubectl get backupsession -n demo -w
NAME                                          INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
sample-etcd-archiver-full-backup-1755248400   BackupConfiguration   sample-etcd-archiver   Running                12s
sample-etcd-archiver-full-backup-1755248400   BackupConfiguration   sample-etcd-archiver   Succeeded   31s        41s
```

If you want to trigger one yourself instead of waiting for the schedule, create a `BackupSession`
naming the session you want:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
metadata:
  name: manual-full-backup
  namespace: demo
spec:
  invoker:
    apiGroup: core.kubestash.com
    kind: BackupConfiguration
    name: sample-etcd-archiver
  session: full-backup
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/backup/trigger-backupsession.yaml
backupsession.core.kubestash.com/manual-full-backup created
```

### Verify the Snapshot

Once a session succeeds, its `Repository` is updated and a `Snapshot` object records the run:

```bash
$ kubectl get repository -n demo sample-etcd-full
NAME               INTEGRITY   SNAPSHOT-COUNT   SIZE      PHASE   LAST-SUCCESSFUL-BACKUP   AGE
sample-etcd-full   true        1                1.482 MiB Ready   45s                      6m10s

$ kubectl get snapshot -n demo -l kubestash.com/repo-name=sample-etcd-full
NAME                                                           REPOSITORY         SESSION       SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
sample-etcd-full-sample-etcd-archiver-full-backup-1755248400    sample-etcd-full   full-backup   2026-08-15T10:00:00Z   Delete            Succeeded   50s
```

Looking inside the `Snapshot`, the etcd data lands under the `dump` component — `etcdctl snapshot
save` produces a single file, so the backup is a logical dump rather than a physical volume copy:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  labels:
    kubestash.com/app-ref-kind: Etcd
    kubestash.com/app-ref-name: sample-etcd
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: sample-etcd-full
  name: sample-etcd-full-sample-etcd-archiver-full-backup-1755248400
  namespace: demo
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Etcd
    name: sample-etcd
    namespace: demo
  backupSession: sample-etcd-archiver-full-backup-1755248400
  repository: sample-etcd-full
  session: full-backup
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      integrity: true
      path: repository/v1/full-backup/dump
      phase: Succeeded
    manifest:
      driver: Restic
      integrity: true
      path: repository/v1/full-backup/manifest
      phase: Succeeded
  integrity: true
  phase: Succeeded
  snapshotTime: "2026-08-15T10:00:00Z"
```

> Only `Snapshot`s whose `status.phase` is `Succeeded` and whose `spec.type` is `FullBackup` are
> eligible for restore. Anything else is skipped when the provisioner picks a snapshot.

Navigating the bucket, the data is under
`etcd-demo/etcd-archiver/demo/sample-etcd/full-backup/repository/v1/full-backup/dump`, and the
`Snapshot` YAMLs are alongside it under `snapshots`. Everything is encrypted with the restic
password from `encrypt-secret`, so it is unreadable until decrypted.

### Pausing backup

To stop the schedules without dropping the repositories, set `spec.pause` on the archiver:

```bash
$ kubectl patch etcdarchiver -n demo etcd-archiver --type merge -p '{"spec":{"pause":true}}'
etcdarchiver.archiver.kubedb.com/etcd-archiver patched

$ kubectl get backupconfiguration -n demo sample-etcd-archiver
NAME                   PHASE   PAUSED   AGE
sample-etcd-archiver   Ready   true     12m
```

Set it back to `false` to resume.

## Restore

This is where etcd differs from most KubeDB databases, so read this before you try anything.

{{< notice type="warning" message="**You cannot restore into a running `Etcd` cluster.** etcd requires a member's data directory to be either genuinely empty or a fully rebuilt snapshot directory *before* the member starts. There is no restore task you can point at a live cluster, and KubeDB will not create one. Restore happens only at bootstrap time, on a new `Etcd` object." >}}

So we deploy a **new** `Etcd` object with `spec.init.archiver` filled in, and the provisioner
performs the restore before it creates the PetSet.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: restored-etcd
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    archiver:
      recoveryTimestamp: "0001-01-01T00:00:00Z"
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: sample-etcd-full
        namespace: demo
      manifestRepository:
        name: sample-etcd-manifest
        namespace: demo
  deletionPolicy: WipeOut
```

Here,

- `spec.init.archiver.fullDBRepository` names the `Repository` holding the etcd snapshots. This is
  the field that turns on data restore.
- `spec.init.archiver.manifestRepository` names the `Repository` holding the Kubernetes objects.
  Restoring it recreates the auth `Secret` before any pod could reference it. It is optional; without
  `fullDBRepository` it would restore the manifests only and let the cluster come up empty.
- `spec.init.archiver.encryptionSecret` must be the same restic password the backup was written with.
- `spec.init.archiver.recoveryTimestamp` is required by the schema, but for etcd it is **not** a
  point-in-time target — there is nothing to replay forward. It selects the newest successful full
  snapshot completed **at or before** that instant. The zero time shown above means "no constraint",
  i.e. restore the newest snapshot; put a real RFC 3339 timestamp there to pin an older one. If no
  snapshot qualifies, the restore fails loudly instead of silently picking a later one.
- `spec.storageType` must be `Durable` with `spec.storage` set. There is no PVC to restore into
  ahead of the pods with `Ephemeral` storage, and the provisioner rejects that combination.

{{< notice type="note" message="KubeStash's manifest-restore options do not yet carry an etcd section, so the restored manifests cannot currently be renamed or filtered the way some other KubeDB databases allow. If you only want the data and would rather let KubeDB generate a fresh auth `Secret` for the new cluster, omit `manifestRepository` and keep `fullDBRepository` alone." >}}

Create it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/backup/restored-etcd.yaml
etcd.kubedb.com/restored-etcd created
```

### Watch the restore happen before the cluster boots

The new `Etcd` reports `DataRestoring` while the seed volume is being rebuilt. Note that no pods
exist yet — the PetSet is deliberately withheld:

```bash
$ kubectl get etcd -n demo restored-etcd
NAME            VERSION   STATUS          AGE
restored-etcd   3.6.4     DataRestoring   35s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=restored-etcd
No resources found in demo namespace.
```

The provisioner creates the manifest restore first, then the snapshot restore:

```bash
$ kubectl get restoresession -n demo
NAME                              REPOSITORY             FAILURE-POLICY   PHASE       DURATION   AGE
restored-etcd-manifest-restorer   sample-etcd-manifest                    Succeeded   9s         40s
restored-etcd-0-snapshot-restorer sample-etcd-full                        Running                12s
```

The snapshot restore targets the seed member's `PersistentVolumeClaim` — created ahead of the PetSet
so the data directory can be rebuilt while no etcd process is running:

```bash
$ kubectl get pvc -n demo
NAME                   STATUS   VOLUME                                     CAPACITY   ACCESS MODES   AGE
data-restored-etcd-0   Bound    pvc-7b3d1f0c-9f1e-4a0a-9a3f-6d2b0c8a1e44   1Gi        RWO            45s

$ kubectl get restoresession -n demo restored-etcd-0-snapshot-restorer \
    -o jsonpath='{.spec.target.kind}{"/"}{.spec.target.name}{"\n"}'
PersistentVolumeClaim/data-restored-etcd-0
```

Only ordinal 0 is ever restored into. Once it is up, members 1 and 2 join through the normal etcd
membership path and stream their copy of the data from the leader.

When the snapshot `RestoreSession` succeeds, the provisioner records the restore on the `Etcd`
object and finally creates the PetSet:

```bash
$ kubectl get etcd -n demo restored-etcd -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\t"}{.reason}{"\n"}{end}'
DatabaseSuccessfullyRestored   True    DatabaseSuccessfullyRestored
ProvisioningStarted            True    DatabaseProvisioningStartedSuccessfully
ReplicaReady                   True    AllReplicasReady
AcceptingConnection            True    DatabaseAcceptingConnectionRequest
Ready                          True    ReadinessCheckSucceeded
Provisioned                    True    DatabaseSuccessfullyProvisioned
```

```bash
$ kubectl get etcd -n demo restored-etcd
NAME            VERSION   STATUS   AGE
restored-etcd   3.6.4     Ready    3m50s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=restored-etcd
NAME              READY   STATUS    RESTARTS   AGE
restored-etcd-0   1/1     Running   0          2m40s
restored-etcd-1   1/1     Running   0          2m18s
restored-etcd-2   1/1     Running   0          1m56s
```

If the restore fails instead, the `DatabaseSuccessfullyRestored` condition is set to `False` with
reason `FailedToRestoreSuccessfully`, and the PetSet is still not created — the cluster never boots on
a half-restored data directory.

### Verify the restored data

```bash
$ kubectl exec -it -n demo restored-etcd-0 -c etcd -- sh

$ export ETCDCTL_API=3
$ export ETCDCTL_ENDPOINTS=http://127.0.0.1:2379
$ export ETCDCTL_USER=root:<root-password>

$ etcdctl get /demo/hello
/demo/hello
hello from the original cluster

$ exit
```

The key we wrote into `sample-etcd` before the backup is present in `restored-etcd`. Anything written
to `sample-etcd` *after* that snapshot completed is not — that is the snapshot-only trade-off.

## Cleanup

```bash
$ kubectl delete etcdarchiver -n demo etcd-archiver
$ kubectl delete backupsession -n demo --all
$ kubectl delete etcd -n demo restored-etcd
$ kubectl delete etcd -n demo sample-etcd
$ kubectl delete retentionpolicy -n demo demo-retention
$ kubectl delete backupstorage -n demo s3-storage
$ kubectl delete secret -n demo s3-secret
$ kubectl delete secret -n demo encrypt-secret
$ kubectl delete ns demo
```

Deleting the `EtcdArchiver` makes the provisioner delete the generated `BackupConfiguration`; what
happens to the repositories then depends on the archiver's `spec.deletionPolicy`.

## Next Steps

- Read the [Etcd Backup & Restore Overview](/docs/guides/etcd/backup/kubestash/overview/index.md)
  for the architecture behind what you just did.
- Detail concepts of the [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md) object.
- Detail concepts of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object.
- Detail concepts of the [AppBinding](/docs/guides/etcd/concepts/appbinding.md) object.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
