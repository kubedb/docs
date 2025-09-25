---
title: Backup & Restore SingleStore | KubeStash
description: Backup SingleStore database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-backup-logical-backup-stashv2
    name: Logical Backup
    parent: guides-sdb-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore SingleStore database using KubeStash

KubeStash allows you to backup and restore `SingleStore` databases. It supports backups for `SingleStore` instances running in Standalone, and cluster configurations. KubeStash makes managing your `SingleStore` backups and restorations more straightforward and efficient.

This guide will give you how you can take backup and restore your `SingleStore` databases using `Kubestash`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore SingleStore databases, please check the following guide [here](/docs/guides/singlestore/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/singlestore/backup/kubestash/logical/examples](/docs/guides/singlestore/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Backup SingleStore

KubeStash supports backups for `SingleStore` instances across different configurations, including Standalone, and Cluster setups. In this demonstration, we'll focus on a `SingleStore` database using Clustering mode. The backup and restore process is similar for Standalone and Cluster configurations as well.

This section will demonstrate how to backup a `SingleStore` database. Here, we are going to deploy a `SingleStore` database using KubeDB. Then, we are going to backup this database into a `GCS` bucket. Finally, we are going to restore the backup up data into another `SingleStore` database.

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```


### Deploy Sample SingleStore Database

Let's deploy a sample `SingleStore` database and insert some data into it.

**Create SingleStore CR:**

Below is the YAML of a sample `SingleStore` CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 3
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    kind: Secret
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Here,

- `spec.version` is the name of the SinglestoreVersion CRD where the docker images are specified. In this tutorial, a SingleStore `8.7.10` database is going to be created.
- `spec.topology` specifies that it will be used as cluster mode. If this field is nil it will be work as standalone mode.
- `spec.topology.aggregator.replicas` or `spec.topology.leaf.replicas` specifies that the number replicas that will be used for aggregator or leaf.
- `spec.storageType` specifies the type of storage that will be used for SingleStore database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create SingleStore database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.topology.aggregator.storage` or `spec.topology.leaf.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Singlestore` crd or which resources KubeDB should keep or delete when you delete `Singlestore` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy)

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in `storage.resources.requests` field. Don't specify limits here. PVC does not get resized automatically.

Create the above `SingleStore` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/backup/kubestash/logical/examples/sdb-sample.yaml
singlestore.kubedb.com/sdb-sample created
```

KubeDB will deploy a SingleStore database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get singlestores.kubedb.com -n demo
NAME           VERSION   STATUS    AGE
sdb-sample      8.7.10   Ready     4m22s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sdb-sample
NAME                   TYPE                       DATA   AGE
sdb-sample-root-cred   kubernetes.io/basic-auth   2      4m58s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sdb-sample
NAME              TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE
sdb-sample        ClusterIP   10.128.230.168   <none>        3306/TCP,8081/TCP   5m10s
sdb-sample-pods   ClusterIP   None             <none>        3306/TCP            5m10s

```

Here, we have to use service `sdb-sample` and secret `sdb-sample-root-cred` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/mysql/concepts/appbinding/index.md) CR that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME           AGE
sdb-sample    9m24s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sdb-sample -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Singlestore","metadata":{"annotations":{},"name":"sdb-sample","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","licenseSecret":{"name":"license-secret"},"storageType":"Durable","topology":{"aggregator":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"0.6","memory":"2Gi"},"requests":{"cpu":"0.6","memory":"2Gi"}}}]}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}}},"leaf":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"0.6","memory":"2Gi"},"requests":{"cpu":"0.6","memory":"2Gi"}}}]}},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"10Gi"}}}}},"version":"8.7.10"}}
  creationTimestamp: "2024-09-11T07:03:44Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sdb-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: singlestores.kubedb.com
  name: sdb-sample
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Singlestore
    name: sdb-sample
    uid: e08e1f37-d869-437d-9b15-14c6aef3f406
  resourceVersion: "4904220"
  uid: 92b2b318-1874-4471-97ec-d789c6e16809
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Singlestore
    name: sdb-sample
    namespace: demo
  clientConfig:
    service:
      name: sdb-sample
      path: /
      port: 3306
      scheme: tcp
    url: tcp(sdb-sample.demo.svc:3306)/
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: SinglestoreConfiguration
    masterAggregator: sdb-sample-aggregator-0.sdb-sample-pods.demo.svc
    stash:
      addon:
        backupTask:
          name: ""
        restoreTask:
          name: ""
  secret:
    name: sdb-sample-root-cred
  type: kubedb.com/singlestore
  version: 8.7.10

```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

- `.spec.parameters.masterAggregator` specifies the dns of master aggregator node that we have to mention in mysqldump command when taken backup or restore.
- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into the any aggregator pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sdb-sample"
NAME                      READY   STATUS    RESTARTS   AGE
sdb-sample-aggregator-0   2/2     Running   0          15m
sdb-sample-aggregator-1   2/2     Running   0          15m
sdb-sample-leaf-0         2/2     Running   0          15m
sdb-sample-leaf-1         2/2     Running   0          15m
sdb-sample-leaf-2         2/2     Running   0          15m
```

And copy the username and password of the `root` user to access into `memsql` shell.

```bash
$ kubectl get secret -n demo  sdb-sample-root-cred -o jsonpath='{.data.username}'| base64 -d
root⏎           

kubectl get secret -n demo  sdb-sample-root-cred -o jsonpath='{.data.password}'| base64 -d
xEJv73q3w_m1~H.G⏎ 
```

Now, Lets exec into the any aggregator `Pod` to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo sdb-sample-aggregator-0 -- singlestore --user=root --password=xEJv73q3w_m1~H.G
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 300070
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> CREATE DATABASE playground partitions 2;
Query OK, 1 row affected (0.01 sec)

singlestore> show databases;
+--------------------+
| Database           |
+--------------------+
| cluster            |
| demo               |
| det                |
| information_schema |
| memsql             |
| playground         |
| singlestore_health |
| test               |
+--------------------+
8 rows in set (0.00 sec)

singlestore> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected, 1 warning (0.27 sec)

singlestore> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.00 sec)

singlestore> INSERT INTO playground.equipment (type, quant, color) VALUES ("slide", 2, "blue");
Query OK, 1 row affected (1.15 sec)

singlestore> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.14 sec)

singlestore> exit
Bye

```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a GCS bucket. We have to create a Secret with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```bash
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat /path/to/downloaded-sa-key.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of `BackupStorage` CR we are going to create,

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: gcs-storage
  namespace: demo
spec:
  storage:
    provider: gcs
    gcs:
      bucket: kubestash-qa
      prefix: demo
      secretName: gcs-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/backup/kubestash/logical/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/gcs-storage created
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `sdb-sample` SingleStore database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

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

Below is the YAML for `BackupConfiguration` CR to backup the `sdb-sample` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-sdb-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Singlestore
    namespace: demo
    name: sdb-sample
  backends:
    - name: gcs-backend
      storageRef:
        namespace: demo
        name: gcs-storage
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
        - name: gcs-sdb-repo
          backend: gcs-backend
          directory: /singlestore
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: singlestore-addon
        tasks:
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `sdb-sample` SigleStore database that we created earlier.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/backup/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-sdb-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                  PHASE   PAUSED   AGE
sample-sdb-backup   Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME               INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-sdb-repo          0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the GCS bucket, we will see the `Repository` YAML stored in the `demo/singlestore` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                          SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-sdb-backup-frequent-backup     */5 * * * *             0        2m45s           3m25s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w

NAME                                             INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
sample-sdb-backup-frequent-backup-1724065200   BackupConfiguration   sdb-sample-backup    Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `sample-sdb-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-sdb-repo
NAME                    INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-sdb-repo          true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=gcs-demo-repo
NAME                                                            REPOSITORY            SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-sdb-repo-sample-sdb-backup-frequent-backup-1724065200   sample-sdb-backup   frequent-backup   2024-01-23T13:10:54Z   Delete            Succeeded   16h
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
$ kubectl get snapshots -n demo gcs-sdb-repo-sample-sdb-backup-frequent-backup-1724065200 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-08-19T12:10:00Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Singlestore
    kubestash.com/app-ref-name: sdb-sample
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: gcs-sdb-repo
  annotations:
    kubedb.com/db-version: 8.7.10
  name: gcs-sdb-repo-sample-sdb-backup-frequent-backup-1724065200
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: gcs-sdb-repo
    uid: 036a2605-0dcc-43bd-a921-363d5c3e8cf0
  resourceVersion: "33742"
  uid: d7f14919-4c31-4b46-84e9-0a75f303ad92
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Singlestore
    name: sdb-sample
    namespace: demo
  backupSession: sample-sdb-backup-frequent-backup-1724065200
  deletionPolicy: Delete
  repository: gcs-sdb-repo
  session: frequent-backup
  snapshotID: 01J6V4P4J0R33C8EG0JWK82118
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 7.472499073s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: dumpfile.sql
        id: fbab3af5c38f51b7aa9096799d8ce4b71ea0092dd8297526fed0adbd9f62f3f1
        size: 3.657 MiB
        uploaded: 1.036 MiB
      size: 1.456 MiB
  ...
```

> KubeStash uses the `mysqldump` command to take backups of target SingleStore databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `demo/singlestore/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/dep/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

#### Deploy Restored Database:

Now, we have to deploy the restored database similarly as we have deployed the original `sdb-sample` database. However, this time there will be the following differences:

- We are going to specify `.spec.init.waitForInitialRestore` field that tells KubeDB to wait for first restore to complete before marking this database is ready to use.

Below is the YAML for `SingleStore` CRD we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: restored-singlestore
  namespace: demo
spec:
  version: "8.7.10"
  init:
    waitForInitialRestore: true
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 3
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    kind: Secret
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/backup/kubestash/logical/examples/restored-singlestore.yaml
singlestore.kubedb.com/restored-singlestore created
```

If you check the database status, you will see it is stuck in `Provisioning` state.

```bash
$ kubectl get singlestore -n demo restored-singlestore
NAME                   VERSION     STATUS         AGE
restored-singlestore   8.7.10     Provisioning    61s
```

#### Create RestoreSession:

Now, we need to create a RestoreSession CRD pointing to targeted `SingleStore` database.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created database provisioned by SingleStore object named `restored-singlestore`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-singlestore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Singlestore
    namespace: demo
    name: restored-singlestore
  dataSource:
    repository: gcs-singlestore-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: singlestore-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restored-singlestore` SingleStore object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-singlestore-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Sep 20 10:44:05 2024

NAME             REPOSITORY        FAILURE-POLICY   PHASE       DURATION   AGE
sample-restore   gcs-demo-repo                      Succeeded   3s         53s
```

The `Succeeded` phase means that the restore process has been completed successfully.


#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get singlestore -n demo restored-singlestore
NAME                   VERSION   STATUS  AGE
restored-singlestore   8.7.10     Ready   34m
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-singlestore"
NAME                                READY   STATUS    RESTARTS   AGE
restored-singlestore-aggregator-0   2/2     Running   0          34m
restored-singlestore-aggregator-1   2/2     Running   0          34m
restored-singlestore-leaf-0         2/2     Running   0          34m
restored-singlestore-leaf-1         2/2     Running   0          34m
restored-singlestore-leaf-2         2/2     Running   0          34m
```

And then copy the user name and password of the `root` user to access into `memsql` shell.

```bash
$ kubectl get secret -n demo  restored-singlestore-root-cred -o jsonpath='{.data.username}'| base64 -d
root⏎           

kubectl get secret -n demo  restored-singlestore-root-cred -o jsonpath='{.data.password}'| base64 -d
QMm1hi0T*7QFz_yh⏎ 
```

Now, Lets exec into the any aggregator `Pod` to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo restored-singlestore-aggregator-0 -- singlestore --user=root --password=QMm1hi0T*7QFz_yh
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 300070
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> show databases;
+--------------------+
| Database           |
+--------------------+
| cluster            |
| demo               |
| det                |
| information_schema |
| memsql             |
| playground         |
| singlestore_health |
| test               |
+--------------------+
8 rows in set (0.00 sec)

singlestore> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.00 sec)

singlestore> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.14 sec)

singlestore> exit
Bye

```

So, from the above output, we can see that the `playground` database and the `equipment` table we have created earlier in the original database and now, they are restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-sdb-backup
kubectl delete restoresessions.core.kubestash.com -n demo sample-restore
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete my -n demo restored-singlestore
kubectl delete my -n demo sdb-sample
```