---
title: Continuous Archiving and Point-in-time Recovery for Neo4j
description: Archive Neo4j full and differential backups with KubeDB and recover a graph to a timestamp before an accidental change.
menu:
  docs_{{ .version }}:
    identifier: pitr-neo4j-archiver
    name: Overview
    parent: pitr-neo4j
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [Neo4j quickstart](/docs/guides/neo4j/quickstart/quickstart.md).

# Continuous Archiving and Point-in-time Recovery for Neo4j

An accidental write can leave a database healthy while its data is wrong. Point-in-time recovery (PITR) lets you recover the graph as it existed before that write, using a full backup and the transaction logs in subsequent differential backups.

In this tutorial, you will deploy a three-member Neo4j source, enable continuous archiving with `Neo4jArchiver`, and restore into a separate standalone Neo4j instance. The recovery exercise changes a person's properties, deletes another person and their relationship, and adds a new person. You will then verify that the restored graph contains only the state committed before your chosen timestamp.

The examples use namespace `demo`. Kubernetes resources have the prefix `neo4j-pitr-`, and backup data uses an isolated object-storage prefix, `neo4j-pitr-demo`.

## Before You Begin

You need:

- A Kubernetes cluster and `kubectl` configured to use it.
- [KubeDB](/docs/setup/README.md), KubeStash, and the Sidekick controller, with support for `Neo4jArchiver` and `Neo4j.spec.init.archiver`.
- A Neo4j Enterprise version and matching backup addon that support full and differential backups. This example uses `2026.06.0`.
- An S3-compatible bucket reachable from the database and backup pods. This example uses an existing MinIO service, `minio.demo.svc.cluster.local:80`, and bucket `kubestash`.
- A provisioner for persistent volumes. The example uses `local-path` and requests `2Gi` per database pod; choose a storage class appropriate for your environment.
- Bash, `jq`, and OpenSSL for the commands below.

> **Tested environment:** This walkthrough was exercised on September 10, 2026, using Neo4j `2026.06.0-enterprise`, Sidekick `v0.0.15`, and development builds of KubeDB and KubeStash. This records the validation environment; it is not a minimum supported release declaration. Verify that your installed controller, CRDs, catalog, and addons support these fields together.

Check the installed API and addon before continuing:

```bash
kubectl explain neo4jarchiver.spec
kubectl explain neo4j.spec.init.archiver
kubectl get neo4jversions.catalog.kubedb.com 2026.06.0
kubectl get addons.addons.kubestash.com neo4j-addon
kubectl get deployments -n kubedb
kubectl get deployments -n kubestash
```

Create the namespace if it does not already exist:

```bash
kubectl get namespace demo >/dev/null 2>&1 || kubectl create namespace demo
```

The manifests are in [docs/guides/neo4j/pitr/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/neo4j/pitr/yamls). Download that directory and run the commands from it. Review the bucket, endpoint, storage class, and credentials before applying the files. Apply each manifest at the step shown; the restore manifest is used only after choosing a recoverable timestamp.

## How the Archive Chain Works

`Neo4jArchiver` describes the backup policy for selected KubeDB `Neo4j` resources. The database label and explicit archiver reference in this example connect `neo4j-pitr-source` to `neo4j-pitr-archiver`.

| Component | Responsibility in this example |
| --- | --- |
| KubeDB | Creates the backup configuration and Sidekick for the selected source; coordinates initialization of the restore target. |
| KubeStash full-backup session | Runs `neo4j-admin` through the `Neo4jAdmin` driver and backs up database manifests. |
| Sidekick | Runs differential backups every five minutes, adding transaction-log artifacts to the full backup's archive chain. |
| Manifest-backup session | Backs up Kubernetes manifests and referenced resources on a separate schedule. |
| Full and manifest repositories | Record the locations and snapshots used during recovery. |
| Restore sessions | Restore requested manifests and recover database files into the target's seed pod volume. |

Neo4j stores its native `.backup` artifacts in S3-compatible storage. The data-backup driver is `Neo4jAdmin`; this workflow does not require CSI `VolumeSnapshot` resources. Keep the encryption Secret available for the KubeStash repository data that uses it; its presence is not a claim that native Neo4j artifacts are encrypted with the Restic password. Configure object-storage encryption separately as needed.

A five-minute interval schedules backup attempts; it does not guarantee a five-minute maximum data loss. Backup duration, failures, and connectivity affect the latest recoverable transaction. Monitor successful differential backups as well as the full-backup schedule.

## Prepare Backup Storage

### Storage Credentials

Edit `storage-secret.yaml` with your bucket credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-secret
  namespace: demo
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: "<your-access-key-id>"
  AWS_SECRET_ACCESS_KEY: "<your-secret-access-key>"
```

```bash
kubectl apply -f storage-secret.yaml
```

If `minio-secret` already contains the correct credentials, reuse it and skip this apply. Do not replace an existing Secret with placeholder values or commit populated credentials.

### BackupStorage

The storage object uses a dedicated prefix, so this tutorial's archives are separate from other backups in the bucket.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: neo4j-pitr-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      bucket: kubestash
      endpoint: http://minio.demo.svc.cluster.local:80
      region: us-east-1
      prefix: neo4j-pitr-demo
      secretName: minio-secret
  usagePolicy:
    allowedNamespaces:
      from: Same
  default: false
  deletionPolicy: Delete
```

```bash
kubectl apply -f backupstorage.yaml
kubectl get backupstorage neo4j-pitr-storage -n demo
```

Wait for `PHASE` to become `Ready`. The example's HTTP endpoint is specific to its internal MinIO service; use the appropriate endpoint and transport security for your storage backend.

### Retention and Encryption Secret

`retention-policy.yaml` keeps up to five successful snapshots and two failed snapshots, subject to a maximum retention period of two months:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: neo4j-pitr-retention
  namespace: demo
spec:
  maxRetentionPeriod: 2mo
  successfulSnapshots:
    last: 5
  failedSnapshots:
    last: 2
  usagePolicy:
    allowedNamespaces:
      from: Same
```

```bash
kubectl apply -f retention-policy.yaml
kubectl create secret generic neo4j-pitr-encryption -n demo \
  --from-literal=RESTIC_PASSWORD="$(openssl rand -hex 24)"
```

The generated password is not printed. Store it securely for recovery. Alternatively, replace the placeholder in `encryption-secret.yaml` and apply that file instead of running `create secret`.

The snapshot count is also a retention constraint: `2mo` does not mean that every point within two months is guaranteed recoverable. A recovery point needs its full backup and the continuous differential chain that covers it. The archiver's successful/failed log history limits control monitoring history, not the desired recovery window.

## Enable Archiving

### Create the Neo4jArchiver

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: Neo4jArchiver
metadata:
  name: neo4j-pitr-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Same
    selector:
      matchLabels:
        archiver: neo4j-pitr
  backupStorage:
    ref:
      name: neo4j-pitr-storage
      namespace: demo
    subDir: /neo4j-backup
  retentionPolicy:
    name: neo4j-pitr-retention
    namespace: demo
  encryptionSecret:
    name: neo4j-pitr-encryption
    namespace: demo
  fullBackup:
    driver: Neo4jAdmin
    scheduler:
      schedule: "0 1 * * *"
    sessionHistoryLimit: 3
    timeout: 1h
  differentialBackup:
    backupInterval: 5m
    successfulLogHistoryLimit: 5
    failedLogHistoryLimit: 5
  manifestBackup:
    scheduler:
      schedule: "*/30 * * * *"
    sessionHistoryLimit: 3
    timeout: 15m
  deletionPolicy: Delete
```

```bash
kubectl apply -f neo4jarchiver.yaml
```

The full-backup schedule runs daily at `01:00` in the scheduler's timezone; the manifest schedule runs every thirty minutes. KubeStash also triggers an initial backup when the generated backup configuration becomes ready. The test cluster uses UTC for its schedules. The differential interval is independent of those CronJob schedules.

### Deploy the Source

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-pitr-source
  namespace: demo
  labels:
    archiver: neo4j-pitr
spec:
  version: "2026.06.0"
  replicas: 3
  archiver:
    ref:
      name: neo4j-pitr-archiver
      namespace: demo
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f neo4j.yaml
kubectl wait neo4j/neo4j-pitr-source -n demo \
  --for=condition=Ready --timeout=10m
kubectl get pods -n demo -l app.kubernetes.io/instance=neo4j-pitr-source
```

The live run produced three ready source pods:

```text
NAME                  READY   STATUS    RESTARTS   AGE
neo4j-pitr-source-0   1/1     Running   0          47s
neo4j-pitr-source-1   1/1     Running   0          42s
neo4j-pitr-source-2   1/1     Running   0          37s
```

> **Deletion policy:** `WipeOut` makes deleting these disposable Neo4j resources destructive to their database storage. Do not use the cleanup commands against a database you intend to keep.

### Verify the First Full Backup

KubeDB creates `neo4j-pitr-source-archiver`; you do not create a separate `BackupConfiguration` for this workflow.

```bash
kubectl get backupconfiguration neo4j-pitr-source-archiver -n demo
kubectl get backupsessions -n demo
kubectl get repositories neo4j-pitr-source-full neo4j-pitr-source-manifest -n demo
kubectl get snapshots.storage.kubestash.com -n demo \
  -l kubestash.com/app-ref-name=neo4j-pitr-source
```

Wait for the configuration and repositories to be `Ready` and the initial full and manifest backup sessions to be `Succeeded`. Use the fully qualified Snapshot resource name: a cluster may also have Longhorn resources named `snapshots`.

This is the initial full-backup artifact recorded in the test:

```text
BackupSession: neo4j-pitr-source-archiver-full-backup-1789046588
Phase:         Succeeded
Database:      neo4j
Artifact:      neo4j-2026-09-10T13-23-42.backup
Artifact time: 2026-09-10T13:23:42 UTC
```

The graph used below is created after this full backup. Recovering it therefore requires the differential chain, not just restoring the full backup.

## Recover a Graph to a Middle Timestamp

The sequence is:

```text
Full backup → Create Alice and Bob → Differential backup
                                      ↓
                              Record recovery cutoff
                                      ↓
                   Change Alice, delete Bob, add Charlie
                                      ↓
                         Later differential backup
                                      ↓
                    Restore only transactions before cutoff
```

### Connect Without Printing Credentials

Define a helper in your Bash session. It reads the credentials from the Secret already mounted in the database pod and passes them to `cypher-shell` through environment variables:

```bash
source_cypher() {
  kubectl exec -n demo neo4j-pitr-source-0 -- sh -c '
    export NEO4J_USERNAME="$(cat /config/neo4j-auth/username)"
    export NEO4J_PASSWORD="$(cat /config/neo4j-auth/password)"
    exec cypher-shell -a neo4j://neo4j-pitr-source.demo.svc:7687 \
      -d neo4j --format plain "$1"
  ' sh "$1"
}
```

The `neo4j://` address allows the client to route writes to the database leader. The source pod must be running and the mounted credentials must be available.

### Create the Baseline Graph

Run this once against the fresh source:

```bash
source_cypher "CREATE (a:PITRPerson {name: 'Alice', age: 30}),
                     (b:PITRPerson {name: 'Bob', age: 25}),
                     (a)-[:KNOWS]->(b);"

source_cypher "MATCH (p:PITRPerson)
               RETURN p.name AS name, p.age AS age ORDER BY name;"
source_cypher "MATCH (:PITRPerson)-[r:KNOWS]->(:PITRPerson)
               RETURN count(r) AS relationships;"
```

```text
name, age
"Alice", 30
"Bob", 25
relationships
1
```

Wait for a successful differential cycle after these writes. Inspect the Sidekick logs and monitoring history:

```bash
kubectl logs -n demo neo4j-pitr-source-sidekick --tail=60
kubectl get snapshots.storage.kubestash.com \
  neo4j-pitr-source-differential-snapshot -n demo -o json |
  jq '.status.components.log.logStats'
```

A completed cycle should report `Differential backup cycle completed`, and `lastSucceededStats` should contain a completion time later than the baseline writes. A running pod alone does not prove that the writes have been archived.

### Record the Cutoff, Then Change the Data

After the baseline differential backup succeeds, capture a whole-second UTC timestamp from Neo4j:

```bash
RECOVERY_TIMESTAMP=$(source_cypher \
  "RETURN toString(datetime.truncate('second', datetime({timezone: '+00:00'}))) AS recoveryTimestamp;" |
  tail -n 1 | tr -d '"\r')
printf '%s\n' "$RECOVERY_TIMESTAMP"
```

The live run returned:

```text
2026-09-10T13:32:30Z
```

Keep this value for the restore manifest. Allow at least two seconds before the next writes so that the cutoff is clearly between the transactions:

```bash
sleep 2
source_cypher "MATCH (a:PITRPerson {name: 'Alice'}) SET a.age = 99;
               MATCH (b:PITRPerson {name: 'Bob'}) DETACH DELETE b;
               CREATE (:PITRPerson {name: 'Charlie', age: 40});"

source_cypher "MATCH (p:PITRPerson)
               RETURN p.name AS name, p.age AS age ORDER BY name;"
source_cypher "MATCH (:PITRPerson)-[r:KNOWS]->(:PITRPerson)
               RETURN count(r) AS relationships;"
```

The live source then returned:

```text
name, age
"Alice", 99
"Charlie", 40
relationships
0
```

The changes had committed by `2026-09-10T13:32:54.085Z`, after the cutoff. These mutations are deliberately limited to the tutorial's `PITRPerson` nodes in its disposable source database.

### Wait for a Backup After the Changes

Wait for another successful differential cycle. Check the history and logs again, making sure the cycle started after the destructive writes and completed successfully.

For a recovery target between backups, the archive chain must include transactions beyond that target. Neo4j's restore command replays differential transaction logs and stops before the specified UTC time. See [Neo4j's timestamp recovery semantics](https://neo4j.com/docs/operations-manual/current/backup-restore/restore-backup/#restore-data-up-to-a-specific-date).

Do not use a future timestamp to demonstrate PITR. In the tested plugin, a requested time newer than every available artifact emits a warning and restores the latest available state. That can succeed without proving historical recovery.

## Restore into a New Neo4j Instance

The restore uses `spec.init.archiver` with the source's full and manifest repositories and the same encryption Secret. KubeDB creates the required `RestoreSession` resources, selects backup data, and supplies the target seed pod and PVC to the data-restore job. You do not need to hand-create those sessions in a matching release.

Use the same Neo4j version for this recovery exercise. The target intentionally has one replica so the example validates recovery separately from subsequent cluster expansion. It has no source archiver selector label.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-pitr-restored
  namespace: demo
spec:
  version: "2026.06.0"
  replicas: 1
  init:
    archiver:
      fullDBRepository:
        name: neo4j-pitr-source-full
        namespace: demo
      manifestRepository:
        name: neo4j-pitr-source-manifest
        namespace: demo
      encryptionSecret:
        name: neo4j-pitr-encryption
        namespace: demo
      # Replace with the UTC cutoff recorded in your own run.
      recoveryTimestamp: "2026-09-10T13:32:30Z"
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Edit `restored-neo4j.yaml` and set `spec.init.archiver.recoveryTimestamp` to the value of `$RECOVERY_TIMESTAMP` from your run. The checked-in timestamp records this tutorial's test; it is not valid for a newly created archive chain.

```bash
kubectl apply -f restored-neo4j.yaml
kubectl get restoresessions -n demo
```

Wait for every generated restore session to report `Succeeded`, then verify the database-specific restore condition. KubeDB selects a successful full-backup Snapshot with a `dump` component; the restore addon follows its archive metadata to the differential artifact whose chain covers the cutoff.

```bash
kubectl wait neo4j/neo4j-pitr-restored -n demo \
  --for=condition=SuccessfullyDataRestored --timeout=10m
kubectl wait neo4j/neo4j-pitr-restored -n demo \
  --for=condition=Ready --timeout=10m
kubectl get neo4j neo4j-pitr-source neo4j-pitr-restored -n demo
```

`Ready` by itself is insufficient: a target can accept connections during provisioning before data recovery finishes. Check the successful restore sessions and `SuccessfullyDataRestored` before validating the graph.

The live resources reached these final states:

```text
NAME                  VERSION     STATUS   AGE
neo4j-pitr-source     2026.06.0   Ready    15h
neo4j-pitr-restored   2026.06.0   Ready    14h

NAME                                       REPOSITORY               PHASE       DURATION
neo4j-pitr-restored-data-backup-restorer   neo4j-pitr-source-full   Succeeded   25s
```

The verified restore used this chain:

```text
Full artifact          neo4j-2026-09-10T13-23-42.backup   transactions 1-3
First differential    neo4j-2026-09-10T13-30-51.backup   transactions 4-8
Second differential   neo4j-2026-09-10T13-35-50.backup   transactions 9-11
Recovery cutoff       2026-09-10T13:32:30Z
Recovered checkpoint  transaction 8
RestoreSession         neo4j-pitr-restored-data-backup-restorer: Succeeded (25s)
```

The restore command retained `--restore-until=2026-09-10 13:32:30`, merged all three artifacts, and checkpointed at transaction 8. Transactions 9-11 contained the later changes and were not applied.

### Verify the Restored Graph

Define the corresponding helper for the target:

```bash
restored_cypher() {
  kubectl exec -n demo neo4j-pitr-restored-0 -- sh -c '
    export NEO4J_USERNAME="$(cat /config/neo4j-auth/username)"
    export NEO4J_PASSWORD="$(cat /config/neo4j-auth/password)"
    exec cypher-shell -a neo4j://neo4j-pitr-restored.demo.svc:7687 \
      -d neo4j --format plain "$1"
  ' sh "$1"
}

restored_cypher "MATCH (p:PITRPerson)
                 RETURN p.name AS name, p.age AS age ORDER BY name;"
restored_cypher "MATCH (a:PITRPerson)-[:KNOWS]->(b:PITRPerson)
                 RETURN a.name AS from, b.name AS to;"
restored_cypher "MATCH (p:PITRPerson {name: 'Charlie'})
                 RETURN count(p) AS charlieCount;"
```

The live restored database returned:

```text
name, age
"Alice", 30
"Bob", 25
source, target
"Alice", "Bob"
charlieCount
0
```

This proves that recovery returned the graph to the middle timestamp: Alice has her original age, Bob and the `KNOWS` relationship exist, and Charlie does not.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| No backup configuration appears | Check the source's selector label, `spec.archiver.ref`, namespace selection, installed CRDs, and controller logs. |
| BackupStorage is not `Ready` | Check that the bucket exists, the credential Secret is in the expected namespace, and the endpoint is reachable from backup pods. |
| Sidekick restarts with a missing required argument | Verify that the KubeDB controller and Neo4j backup-plugin images come from a compatible release. |
| Sidekick runs but archives are stale | Inspect its logs and the latest successful differential history; do not infer archive health from pod readiness. |
| Snapshot list is empty | Use `snapshots.storage.kubestash.com` explicitly to avoid querying another API group's Snapshot resource. |
| RestoreSession is `Invalid` with `Component dump not exist` | Verify that the controller selected a successful full-backup Snapshot containing `status.components.dump`, and install matching controller and addon builds. |
| Restore reports no continuous chain covering the target | Check that a full backup predates the cutoff and differential artifacts cover it; verify that retention or manual object deletion has not removed required artifacts. |
| Restore warns that the target is newer than the latest backup | Wait for an archive that covers the desired time and restore into a new target; a latest-state fallback is not evidence of PITR. |
| Database is ready but expected data is absent | Check both restore sessions and `SuccessfullyDataRestored`, then query the correct database and target service. |

Inspect the restore job's output when a recovery fails or its result is unexpected:

```bash
kubectl describe restoresession neo4j-pitr-restored-data-backup-restorer -n demo
kubectl logs -n demo job/neo4j-pitr-restored-data-backup-restorer
kubectl get events -n demo --sort-by=.lastTimestamp
```

Use `status.components.log.logStats` and the archive/restore logs as evidence of current differential-backup health; do not confuse the monitoring object with a full-backup Snapshot containing the `dump` component.

## Cleanup

Keep the source, restored database, repositories, and encryption Secret until you have finished verifying recovery. The live tutorial resources were left in place for inspection.

When you explicitly want to remove the disposable databases:

```bash
kubectl delete neo4j neo4j-pitr-source neo4j-pitr-restored -n demo
kubectl delete neo4jarchiver neo4j-pitr-archiver -n demo
```

The database manifests use `WipeOut`, so this removes their database storage. Inspect the remaining repositories and snapshots before deleting backup-related resources. The example uses `Delete` rather than `WipeOut` for archive storage; do not assume that deleting Kubernetes objects purges the archived objects from the bucket. Preserve the encryption Secret while any retained repository still needs it. Do not delete the shared MinIO credential Secret or the entire `demo` namespace.

## Next Steps

- [Backup and restore standalone and HA Neo4j](/docs/guides/neo4j/backup/kubestash/logical/standalone-and-ha/) for the explicit `BackupConfiguration`/`RestoreSession` workflow.
- [Customize Neo4j backup and restore](/docs/guides/neo4j/backup/kubestash/customization/index.md) for database selection, resource settings, and restore parameters.
- [Back up composite databases and aliases](/docs/guides/neo4j/backup/kubestash/logical/composite-database/) for catalog and alias-specific considerations beyond this single-database recovery exercise.
