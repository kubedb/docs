---
title: Backup & Restore Neo4j | KubeStash
description: Backup and Restore Neo4j database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-neo4j-logical-backup-stashv2
    name: Logical Backup
    parent: guides-neo4j-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore Neo4j database using KubeStash

KubeStash allows you to backup and restore `Neo4j` databases. It supports backups for `Neo4j` instances running in Standalone and HA cluster configurations. KubeStash makes managing your `Neo4j` backups and restorations more straightforward and efficient.

This guide will give you an overview how you can take backup and restore your `Neo4j` databases using `KubeStash`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore Neo4j databases, please check the following guide [here](/docs/guides/neo4j/backup/kubestash/overview/index.md).

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)
- [Task](https://kubestash.com/docs/latest/concepts/crds/addon/#task-specification)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/neo4j/backup/kubestash/logical/examples](/docs/guides/neo4j/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Backup Neo4j

KubeStash supports backups for `Neo4j` instances across different configurations, including Standalone and HA Cluster setups. In this demonstration, we'll focus on a `Neo4j` database using HA cluster configuration. The backup and restore process is similar for Standalone configuration.

This section will demonstrate how to backup a `Neo4j` database. Here, we are going to deploy a `Neo4j` database using KubeDB. Then, we are going to backup this database into an `S3` bucket. Finally, we are going to restore the backed up data into another `Neo4j` database.

### Deploy Sample Neo4j Database

Let's deploy a sample `Neo4j` database and insert some data into it.

**Create Neo4j CR:**

Below is the YAML of a sample `Neo4j` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: sample-neo4j
  namespace: demo
spec:
  version: 2025.11.2
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Create the above `Neo4j` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/examples/sample-neo4j.yaml
neo4j.kubedb.com/sample-neo4j created
```

KubeDB will deploy a `Neo4j` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get neo4j -n demo sample-neo4j
NAME           VERSION     STATUS   AGE
sample-neo4j   2025.11.2   Ready    5m1s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-neo4j
NAME                 TYPE     DATA   AGE
sample-neo4j-auth    Opaque   2      5m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-neo4j
NAME             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                 AGE
sample-neo4j     ClusterIP   10.43.214.193   <none>        6362/TCP,7687/TCP,7474/TCP                              5m55s
sample-neo4j-0   ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   5m55s
sample-neo4j-1   ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   5m55s
sample-neo4j-2   ClusterIP   None            <none>        6362/TCP,7687/TCP,7474/TCP,7688/TCP,7000/TCP,6000/TCP   5m55s
```

Here, we have to use service `sample-neo4j` and secret `sample-neo4j-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/neo4j/concepts/appbinding.md) CR that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME           TYPE               VERSION                AGE
sample-neo4j   kubedb.com/Neo4j   2025.11.2-enterprise   86s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-neo4j -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Neo4j","metadata":{"annotations":{},"name":"sample-neo4j","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"2Gi"}}},"storageType":"Durable","version":"2025.11.2"}}
  creationTimestamp: "2026-06-22T05:50:48Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-neo4j
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
  name: sample-neo4j
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Neo4j
      name: sample-neo4j
      uid: a2ab1ada-ebd5-4673-b04f-eba5ec189007
  resourceVersion: "223034"
  uid: f207d8bc-e4d1-427f-84cc-3199e684619e
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Neo4j
    name: sample-neo4j
    namespace: demo
  clientConfig:
    service:
      name: sample-neo4j
      port: 7687
      scheme: noe4j
  secret:
    name: sample-neo4j-auth
  type: kubedb.com/Neo4j
  version: 2025.11.2-enterprise
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following fields to be set in the AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into one of the database pods and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-neo4j"
NAME             READY   STATUS    RESTARTS   AGE
sample-neo4j-0   1/1     Running   0          118s
sample-neo4j-1   1/1     Running   0          112s
sample-neo4j-2   1/1     Running   0          106s
```

Retrieve the auth credentials so we can connect using `cypher-shell`,

```bash
$ kubectl get secret -n demo sample-neo4j-auth -o jsonpath='{.data.username}' | base64 -d
neo4j

$ kubectl get secret -n demo sample-neo4j-auth -o jsonpath='{.data.password}' | base64 -d
UEke.988YxVdJbGq
```

Now, let's exec into the pod and create some nodes,

```bash
$ export PASS=$(kubectl get secret -n demo sample-neo4j-auth -o jsonpath='{.data.password}' | base64 -d)

# create a few Person nodes and a relationship in the default "neo4j" database
$ kubectl exec -it -n demo sample-neo4j-0 -- cypher-shell -u neo4j -p "$PASS" \
    "CREATE (alice:Person {name: 'Alice', age: 30})
     CREATE (bob:Person {name: 'Bob', age: 25})
     CREATE (alice)-[:KNOWS]->(bob);"
0 rows
ready to start consuming query after 25 ms, results consumed after another 0 ms
Added 2 nodes, Created 1 relationships, Set 4 properties, Added 2 labels

# verify that the data has been inserted
$ kubectl exec -it -n demo sample-neo4j-0 -- cypher-shell -u neo4j -p "$PASS" \
    "MATCH (p:Person) RETURN p.name AS name, p.age AS age ORDER BY name;"
+---------------+
| name    | age |
+---------------+
| "Alice" | 30  |
| "Bob"   | 25  |
+---------------+

2 rows
```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into an `S3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired S3 bucket,

```bash
$ echo -n '<your-access-key-id>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-secret-access-key>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of `BackupStorage` CR we are going to create,

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
      endpoint: http://minio.demo.svc.cluster.local:80
      bucket: kubestash-qa
      prefix: demo
      region: us-east-1
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

Now, we are ready to backup our database to our desired backend.

**Create RetentionPolicy:**

Now, let's create a `RetentionPolicy` to specify how the old Snapshots should be cleaned up.

Below is the YAML of the `RetentionPolicy` object that we are going to create,

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
      from: All
```

Let's create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting the respective `sample-neo4j` Neo4j database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to backup the `sample-neo4j` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-neo4j-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: sample-neo4j
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
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-neo4j-repo
          backend: s3-backend
          directory: /backup
      addon:
        name: neo4j-addon
        tasks:
          - name: logical-backup
```

- `.spec.target` refers to the targeted `sample-neo4j` Neo4j database that we created earlier.
- `.spec.backends[*].storageRef` refers to the `BackupStorage` we created earlier where the backup data will be stored.
- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.sessions[*].addon` refers to the `neo4j-addon` that performs the backup. The `logical-backup` task uses the `neo4j-admin database backup` command under the hood.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-neo4j-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                  PHASE   PAUSED   AGE
sample-neo4j-backup   Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME            INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-neo4j-repo                0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the S3 bucket, we will see the `Repository` YAML stored in the `demo/backup` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                          SCHEDULE      TIMEZONE   SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-neo4j-backup-frequent-backup   */5 * * * *   <none>     False     0        <none>          2m18s

```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                             INVOKER-TYPE          INVOKER-NAME          PHASE       DURATION   AGE
sample-neo4j-backup-frequent-backup-1782108669   BackupConfiguration   sample-neo4j-backup   Succeeded   44s        119s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `s3-neo4j-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo s3-neo4j-repo
NAME            INTEGRITY   SNAPSHOT-COUNT   SIZE   PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-neo4j-repo               1                0 B    Ready   2m46s                    2m57s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-neo4j-repo
NAME                                                           REPOSITORY      SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1782108669   s3-neo4j-repo   frequent-backup   2026-06-22T06:11:20Z   Delete            Succeeded   3m2s
```

> Note: KubeStash creates a `Snapshot` with the following labels:
> - `kubedb.com/db-version: <db-version>`
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the Database.

```bash
$ kubectl get snapshots -n demo s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1782108669 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  annotations:
    kubedb.com/db-version: 2025.11.2-enterprise
  creationTimestamp: "2026-06-22T06:11:20Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Neo4j
    kubestash.com/app-ref-name: sample-neo4j
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: s3-neo4j-repo
  name: s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1782108669
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: s3-neo4j-repo
      uid: 8d9ccc3f-189b-44ac-bc40-a3b15a8282a9
  resourceVersion: "224107"
  uid: 1ec4abef-305d-48f6-85aa-42b9f766a6ca
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Neo4j
    name: sample-neo4j
    namespace: demo
  backupSession: sample-neo4j-backup-frequent-backup-1782108669
  deletionPolicy: Delete
  repository: s3-neo4j-repo
  session: frequent-backup
  snapshotID: 01KVPZ8WNXRWKSPN23RMFP6MNC
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Neo4jAdmin
      duration: 14.932375451s
      neo4jStats:
        - compressed: true
          database: system
          databaseID: 00000000-0000-0000-0000-000000000001
          file: s3://kubestash-qa/demo/backup/repository/v1/frequent-backup/dump/system-2026-06-22T06-11-41.backup
          full: true
          highestTX: 145
          lowestTX: 1
          recovered: true
          storeIDHash: "-297679445"
          time: 2026-06-22T06:11:41
        - compressed: true
          database: neo4j
          databaseID: 3efaaefe-1e16-4501-998f-a7f66f0d1ebe
          file: s3://kubestash-qa/demo/backup/repository/v1/frequent-backup/dump/neo4j-2026-06-22T06-11-47.backup
          full: true
          highestTX: 247
          lowestTX: 1
          recovered: true
          storeIDHash: "-506103840"
          time: 2026-06-22T06:11:47
      path: s3://kubestash-qa/demo/backup/repository/v1/frequent-backup/dump/
      phase: Succeeded
  conditions:
    - lastTransitionTime: "2026-06-22T06:11:20Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2026-06-22T06:12:01Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  phase: Succeeded
  snapshotTime: "2026-06-22T06:11:20Z"
  totalComponents: 1
  verificationStatus: NotVerified
```

> KubeStash uses the `neo4j-admin database backup` command to perform backups of the target `Neo4j` databases. It backs up every database of the instance (including the `system` database). Therefore, the component name for logical backups is set as `dump`, and the `neo4jStats` field lists each backed up database.

Now, if we navigate to the S3 bucket, we will see the backed up data stored in the `demo/backup/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/backup/snapshots` directory.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and once it becomes ready, we are going to restore it from the backup.

Now, we have to deploy the restored database similarly as we have deployed the original `sample-neo4j` database.

Below is the YAML for `Neo4j` CR we are going to deploy,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: restored-neo4j
  namespace: demo
spec:
  version: 2025.11.2
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/examples/restored-neo4j.yaml
neo4j.kubedb.com/restored-neo4j created
```

Let's wait for the database to be ready to use,

```bash
$ kubectl get neo4j -n demo restored-neo4j
NAME            VERSION     STATUS   AGE
restored-neo4j   2025.11.2   Ready    5m1s
```

The database is `Ready`. Now, we are going to restore the backed up data into this database.

#### Create RestoreSession:

Now, we need to create a `RestoreSession` CR pointing to the targeted `Neo4j` database.

Below, is the contents of the YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the `Neo4j` database named `restored-neo4j`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-neo4j-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: restored-neo4j
  dataSource:
    repository: s3-neo4j-repo
    snapshot: latest
  addon:
    name: neo4j-addon
    tasks:
      - name: logical-backup-restore
        params:
          seedServerName: "restored-neo4j-0" ## Neo4j Pod Name
    jobTemplate:
      spec:
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: data-restored-neo4j-0 # PVC Name
        volumeMounts:
          - mountPath: /data
            name: data
            subPath: data
        securityContext:
          runAsNonRoot: true
          runAsUser: 7474
```

Here,

- `.spec.target` refers to the newly created `restored-neo4j` Neo4j object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from the latest `Snapshot`.
- `.spec.addon.tasks[*].params.seedServerName` specifies the `Neo4j` pod that will be used to seed the restored data into the cluster. The other replicas are then synced from this seed server.
- `.spec.addon.jobTemplate` mounts the data PVC of the seed pod (`data-restored-neo4j-0`) into the restore `Job` at `/data` and runs the `Job` as the `neo4j` user (`runAsUser: 7474`), so the restored store files have the correct ownership.

Let's create the RestoreSession CR object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/neo4j/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-neo4j-restore created
```

Once you have created the `RestoreSession` object, KubeStash will create a restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restoresession -n demo
NAME                   REPOSITORY      FAILURE-POLICY   PHASE       DURATION   AGE
sample-neo4j-restore   s3-neo4j-repo                    Succeeded   18s        116s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the nodes we created earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get neo4j -n demo restored-neo4j
NAME            VERSION     STATUS   AGE
restored-neo4j   2025.11.2   Ready    6m31s
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-neo4j"
NAME              READY   STATUS    RESTARTS   AGE
restored-neo4j-0   1/1     Running   0          6m7s
restored-neo4j-1   1/1     Running   0          6m1s
restored-neo4j-2   1/1     Running   0          5m55s
```

Now, let's exec into one of the `Pod` and verify the restored data.

```bash
$ export PASS=$(kubectl get secret -n demo restored-neo4j-auth -o jsonpath='{.data.password}' | base64 -d)

# verify that the Person nodes have been restored
$ kubectl exec -it -n demo restored-neo4j-0 -- cypher-shell -u neo4j -p "$PASS" \
    "MATCH (p:Person) RETURN p.name AS name, p.age AS age ORDER BY name;"
+---------------+
| name    | age |
+---------------+
| "Alice" | 30  |
| "Bob"   | 25  |
+---------------+

2 rows
```

So, from the above output, we can see the nodes we had created in the original database `sample-neo4j` have been restored in the `restored-neo4j` database.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com -n demo sample-neo4j-backup
kubectl delete restoresessions.core.kubestash.com -n demo sample-neo4j-restore
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete neo4j -n demo restored-neo4j
kubectl delete neo4j -n demo sample-neo4j
```
