---
title: Backup & Restore Elasticsearch | KubeStash
description: Backup ans Restore Elasticsearch database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-es-logical-backup-stashv2
    name: Logical Backup
    parent: guides-es-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore Elasticsearch database using KubeStash

KubeStash allows you to backup and restore `Elasticsearch` databases. It supports backups for `Elasticsearch` instances running in Standalone,  and HA cluster configurations. KubeStash makes managing your `Elasticsearch` backups and restorations more straightforward and efficient.

This guide will give you an overview how you can take backup and restore your `Elasticsearch` databases using `Kubestash`.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore Elasticsearch databases, please check the following guide [here](/docs/guides/postgres/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/backup/kubestash/logical/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


## Backup Elasticsearch

KubeStash supports backups for `Elasticsearch` instances across different configurations, including Standalone and HA Cluster setups. In this demonstration, we'll focus on a `Elasticsearch` database using HA cluster configuration. The backup and restore process is similar for Standalone configuration.

This section will demonstrate how to take logical backup of a `Elasticsearch` database. Here, we are going to deploy a `Elasticsearch` database using KubeDB. Then, we are going to back up the database at the application level to a `GCS` bucket. Finally, we will restore the entire `Elasticsearch` database.

### Deploy Sample Elasticsearch Database

Let's deploy a sample `Elasticsearch` database and insert some data into it.

**Create Elasticsearch CR:**

Below is the YAML of a sample `Elasticsearch` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-quickstart
  namespace: demo
spec:
  version: xpack-8.15.0
  enableSSL: true
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Create the above `Elasticsearch` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/logical/examples/es-quickstart.yaml
elasticsearch.kubedb.com/es-quickstart created
```

KubeDB will deploy a `Elasticsearch` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get es -n demo es-quickstart
NAME              VERSION        STATUS   AGE
es-quickstart     xpack-8.15.0   Ready    3h
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo
NAME                                        TYPE                       DATA   AGE
es-quickstart-apm-system-cred               kubernetes.io/basic-auth   2      3h35m
es-quickstart-beats-system-cred             kubernetes.io/basic-auth   2      3h35m
es-quickstart-ca-cert                       kubernetes.io/tls          2      3h1m
es-quickstart-client-cert                   kubernetes.io/tls          3      3h1m
es-quickstart-config                        Opaque                     1      3h1m
es-quickstart-elastic-cred                  kubernetes.io/basic-auth   2      3h35m
es-quickstart-http-cert                     kubernetes.io/tls          3      3h1m
es-quickstart-kibana-system-cred            kubernetes.io/basic-auth   2      3h35m
es-quickstart-logstash-system-cred          kubernetes.io/basic-auth   2      3h35m
es-quickstart-remote-monitoring-user-cred   kubernetes.io/basic-auth   2      3h35m
es-quickstart-transport-cert                kubernetes.io/tls          3      3h1m

$ kubectl get service -n demo -l=app.kubernetes.io/instance=es-quickstart
NAME                   TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
es-quickstart          ClusterIP   10.128.185.239   <none>        9200/TCP   3h2m
es-quickstart-master   ClusterIP   None             <none>        9300/TCP   3h2m
es-quickstart-pods     ClusterIP   None             <none>        9200/TCP   3h2m
```

Here, we have to use service `es-quickstart` and secret `es-quickstart-elastic-cred` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/elasticsearch/concepts/appbinding/index.md) CR that holds the necessary information to connect with the database.


**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
 $ kubectl get appbindings -n demo
NAME              TYPE                       VERSION   AGE
es-quickstart     kubedb.com/elasticsearch   8.15.0    3h6m
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo es-quickstart -o yaml
```

```yaml
apiVersion: v1
items:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    kind: AppBinding
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: |
          {"apiVersion":"kubedb.com/v1alpha2","kind":"Elasticsearch","metadata":{"annotations":{},"name":"es-quickstart","namespace":"demo"},"spec":{"enableSSL":true,"storageType":"Durable","topology":{"data":{"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"linode-block-storage"}},"ingest":{"replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"linode-block-storage"}},"master":{"replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"linode-block-storage"}}},"version":"xpack-8.15.0"}}
      creationTimestamp: "2024-09-18T09:46:17Z"
      generation: 1
      labels:
        app.kubernetes.io/component: database
        app.kubernetes.io/instance: es-quickstart
        app.kubernetes.io/managed-by: kubedb.com
        app.kubernetes.io/name: elasticsearches.kubedb.com
      name: es-quickstart
      namespace: demo
      ownerReferences:
        - apiVersion: kubedb.com/v1
          blockOwnerDeletion: true
          controller: true
          kind: Elasticsearch
          name: es-quickstart
          uid: 3dba7eba-9e83-49dd-bd8d-7917f89b1b43
      resourceVersion: "18128"
      uid: af0e9f28-9ab6-4a72-a31a-bb80439f6d8b
    spec:
      appRef:
        apiGroup: kubedb.com
        kind: Elasticsearch
        name: es-quickstart
        namespace: demo
      clientConfig:
        caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURIVENDQWdXZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREF3TVJNd0VRWURWUVFLRXdwcmRXSmwKWkdJdVkyOXRNUmt3RndZRFZRUURFeEJsY3kxeGRXbGphM04wWVhKMExXTmhNQjRYRFRJME1Ea3hPREE1TkRZeApNVm9YRFRNME1Ea3hOakE1TkRZeE1Wb3dNREVUTUJFR0ExVUVDaE1LYTNWaVpXUmlMbU52YlRFWk1CY0dBMVVFCkF4TVFaWE10Y1hWcFkydHpkR0Z5ZEMxallUQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0MKZ2dFQkFPQkJTWTIzVnIxSlI5UjVFSXdjeWpEdDVBdXd6bDV1eld5QU5yb0NXZng3V2IwZEFENGVJOHpBNTdUNQoxaEtMWlFsT1FlSWpjcTB2dHpCZkJaSDgwSWZQUzd6M0dRNzdNbzZlS1pwQVdpeC9iVm1BbGpvVGsrOG5DUUY1CjgrUjZXczU2NllpR3J6bmhidUl6SFNNVkdiS0IzWDlTRGs4bWNOcEZCaVFqaU5INHdnMmxmNGNNUUQ3cXpydzcKakdxMmEva082KzZJUHVwQ0R0YUgwdXFhbHkzcmhUK2g2Ukx4ZHRiMStBUDloUUtmbXFLS0YwOWdsQTgxSlg0ZApLQ2ZmMXl2SkY4WW5FalE0M1VHUmFVZ1RtNkNCRkI4RWJKdnJBRkhMU1AvOHVTY1hKMmxZVFlJUzV1MldJYzdtCmdVNWE2MTV5UmZRaWRTak00MjY3clFrNmR2Y0NBd0VBQWFOQ01FQXdEZ1lEVlIwUEFRSC9CQVFEQWdLa01BOEcKQTFVZEV3RUIvd1FGTUFNQkFmOHdIUVlEVlIwT0JCWUVGRHpJK2ZBMkZLTXliNkQvTCtNdEY3WVdHMTlOTUEwRwpDU3FHU0liM0RRRUJDd1VBQTRJQkFRQy81a1VBRjZlSDJhYmxSYzFQYURyWUx4ZlJYL2o5eTJwcmNxejlrcjBOCk9pT1ZEakdlSHRZN3Rjc0l4b2JhZDNDMUdybFVVMXlxUVd2dGVYNE8vMXpLTmJTeHZBTk4zbnJuMDd0a3ZpejQKOUVXVktzVFpKVlJRMEtVTkNjcml4WXcyV2RBVzhvTkJlWll6RGNnTmZWa1NRWHZDcjdhaEZaaWNmZWYzOVNHVApIKzBlVU52MlJZTkJHMlNGdTYweWRLNDdUeHNLS1NkTHFMWktOMHpsSDJCT2huSjducm4vK3FITi9TZDJoLzJQCnlibjN4ckdTeGxWakRNLzJ1U1VQd1F4azJ2VlluSkVoOEFuWFB6WlJ2c1ZjS1Vyc25OdVNQa1N2OVh1YnBpYkUKR3lJMFZNbW10bHY1WGFuMEtiTXBpTkphN3loRGhVb3AyMDdaVW42ajNjSkIKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
        service:
          name: es-quickstart
          port: 9200
          scheme: https
      parameters:
        apiVersion: appcatalog.appscode.com/v1alpha1
        kind: StashAddon
        stash:
          addon:
            backupTask:
              name: elasticsearch-backup-8.2.0
              params:
                - name: args
                  value: --match=^(?![.])(?!apm-agent-configuration)(?!kubedb-system).+
            restoreTask:
              name: elasticsearch-restore-8.2.0
              params:
                - name: args
                  value: --match=^(?![.])(?!apm-agent-configuration)(?!kubedb-system).+
      secret:
        name: es-quickstart-elastic-cred
      tlsSecret:
        name: es-quickstart-client-cert
      type: kubedb.com/elasticsearch
      version: 8.15.0
kind: List
metadata:
  resourceVersion: ""
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to insert some data into Elasticsearch.
```bash
$ kubectl get secret -n demo es-quickstart-elastic-cred -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo es-quickstart-elastic-cred -o jsonpath='{.data.password}' | base64 -d
tS$k!2IBI.ASI7FJ
```

```bash
$ kubectl port-forward -n demo svc/es-quickstart 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

```bash
$ curl -XPOST -k --user 'elastic:tS$k!2IBI.ASI7FJ' "https://localhost:9200/info/_doc?pretty" -H 'Content-Type: application/json' -d'
         {
             "Company": "AppsCode Inc",
             "Product": "KubeDB"
         }
         '

```
Now, let’s verify that the index have been created successfully.

```bash
$ curl -XGET -k --user 'elastic:tS$k!2IBI.ASI7FJ' "https://localhost:9200/_cat/indices?v&s=index&pretty"
health status index            uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   .geoip_databases FsJlvTyRSsuRWTpX8OpkOA   1   1         40            0       76mb           38mb
green  open   info             9Z2Cl5fjQWGBAfjtF9LqBw   1   1          1            0      8.9kb          4.4kb
```
Also, let’s verify the data in the indexes:

```bash
curl -XGET -k --user 'elastic:tS$k!2IBI.ASI7FJ' "https://localhost:9200/info/_search?pretty"
{
  "took" : 79,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "info",
        "_type" : "_doc",
        "_id" : "mQCvA4ABs70-lBxlFWZD",
        "_score" : 1.0,
        "_source" : {
          "Company" : "AppsCode Inc",
          "Product" : "KubeDB"
        }
      }
    ]
  }
}

```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a `S3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired S3 bucket,

```bash
$ echo -n '<your-access-key>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-secret-key>' > AWS_SECRET_ACCESS_KEY
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
      endpoint: us-east-1.linodeobjects.com
      bucket: esbackup
      region: us-east-1
      prefix: elastic
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete

```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/logical/examples/backupstorage.yaml
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

Let’s create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `es-quickstart` Elasticsearch database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD \
secret "encrypt-secret" created
```

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to take logical backup of the `es-quickstart` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: es-quickstart-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    namespace: demo
    name: es-quickstart
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
        - name: s3-elasticsearch-repo
          backend: s3-backend
          directory: /es
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: elasticsearch-addon
        tasks:
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup at `5 minutes` interval.
- `.spec.target` refers to the targeted `es-quickstart` Elasticsearch database that we created earlier.
- `.spec.sessions[*].addon.tasks[*].name[*]` specifies that the `logical-backup` tasks will be executed.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/es-quickstart-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                     PHASE   PAUSED   AGE
es-quickstart-backup     Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                      INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-elasticsearch-repo                 0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the S3 bucket, we will see the `Repository` YAML stored in the `elastic/es` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                             SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-es-quickstart-backup-frequent-backup     */5 * * * *             0        2m45s           3m25s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                              INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
es-quickstart-backup-frequent-backup-1726655113   BackupConfiguration   es-quickstart-backup   Succeeded   22s        2m7s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `es-quickstart-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo
NAME                    INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-elasticsearch-repo   true        1                1.453 KiB   Ready   2m20s                    2m30s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-elasticsearch-repo
NAME                                                              REPOSITORY              SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-elasticsearch-repo-es-quickstckup-frequent-backup-1726655113   s3-elasticsearch-repo   frequent-backup   2024-09-18T10:25:23Z   Delete            Succeeded   8m
```

> Note: KubeStash creates a `Snapshot` with the following labels:
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the Database.

```bash
$ kubectl get snapshots -n demo  s3-elasticsearch-repo-es-quickstckup-frequent-backup-1726655113 -oyaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-18T10:25:23Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 8.15.0
    kubestash.com/app-ref-kind: Elasticsearch
    kubestash.com/app-ref-name: es-quickstart
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: s3-elasticsearch-repo
  name: s3-elasticsearch-repo-es-quickstckup-frequent-backup-1726655113
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: s3-elasticsearch-repo
    uid: 0711debc-de7f-418e-9898-dc5b24affc81
  resourceVersion: "20786"
  uid: 420eb77d-db66-40eb-b78a-af4d4a3078a6
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-quickstart
    namespace: demo
  backupSession: es-quickstart-backup-frequent-backup-1726655113
  deletionPolicy: Delete
  repository: s3-elasticsearch-repo
  session: frequent-backup
  snapshotID: 01J82AMM2WCPAWCTQ08DW3VR2N
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 1.712341296s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: /kubestash-interim/data
        id: 396227e62948a4d9ca865f08b52bfcc3fbca7135b1962373c203df856bd9a260
        size: 509 B
        uploaded: 2.641 KiB
      size: 1.455 KiB
  conditions:
  - lastTransitionTime: "2024-09-18T10:25:23Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-18T10:25:40Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 1.454 KiB
  snapshotTime: "2024-09-18T10:25:23Z"
  totalComponents: 1
```

> KubeStash uses `multielasticdump` to perform backups of target `Elasticsearch` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the S3 bucket, we will see the backed up data stored in the `elastic/es/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/postgres/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

Now, we have to deploy the restored database similarly as we have deployed the original `es-quickstart` database. 

Below is the YAML for `Elasticsearch` CR we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  version: xpack-8.15.0
  enableSSL: true
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/logical/examples/restore-es.yaml
elasticsearch.kubedb.com/es-cluster created
```

If you check the database status, you will see it is stuck in **`Provisioning`** state.

```bash
$ kubectl get es -n demo restored-es
NAME               VERSION   STATUS         AGE
es-cluster         8.15.0    Provisioning   61s
```

#### Create RestoreSession:

Now, we need to create a `RestoreSession` CR pointing to targeted `Elasticsearch` database.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created `Elasticsearch` database named `restored-postgres`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: es-cluster-restore
  namespace: demo
spec:
  target:
    name: es-cluster
    namespace: demo
    apiGroup: kubedb.com
    kind: Elasticsearch
  dataSource:
    snapshot: latest
    repository: s3-elasticsearch-repo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: elasticsearch-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restore-es` Elasticsearch object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/es-cluster-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024
NAME                      REPOSITORY             FAILURE-POLICY   PHASE       DURATION   AGE
es-cluster-restore     s3-elasticsearch-repo                   Succeeded   7s         116s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get es -n demo es-cluster
NAME            VERSION        STATUS   AGE
es-cluster      xpack-8.15.0   Ready    6m14s
```

```bash
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.username}' | base64 -d
elastic
$ kubectl get secret -n demo es-cluster-elastic-cred -o jsonpath='{.data.password}' | base64 -d
tS$k!2IBI.ASI7FJ
```

```bash
$ kubectl port-forward -n demo svc/es-cluster 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```


Now, lets check either data restored in elasticsearch or not.

```bash
$ curl -XGET -k --user 'elastic:vD~b4DMXZ1iwdjnh' "https://localhost:9200/info/_search?pretty"
{
  "took" : 83,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "info",
        "_id" : "lT6pBJIBNMreROyUqVKF",
        "_score" : 1.0,
        "_source" : {
          "Company" : "AppsCode Inc",
          "Product" : "KubeDB"
        }
      }
    ]
  }
}
```

So, from the above output, we can see the `info` database we had created in the original database `es-cluster` has been restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo es-quickstart-backup
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete restoresessions.core.kubestash.com -n demo es-cluster-restore
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete postgres -n demo es-quickstart
kubectl delete postgres -n dev es-cluster
```