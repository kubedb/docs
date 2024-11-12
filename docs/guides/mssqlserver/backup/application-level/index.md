---
title: Application Level Backup & Restore Microsoft SQL Server | KubeStash
description: Application Level Backup and Restore using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-application-level-backup
    name: Application Level Backup
    parent: guides-mssqlserver-backup
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---


# Application Level Backup and Restore Microsoft SQL Server database using KubeStash

KubeStash offers application-level backup and restore functionality for `Microsoft SQL Server` databases. It captures both manifest and data backups of any `Microsoft SQL Server` database in a single snapshot. During the restore process, KubeStash first applies the `Microsoft SQL Server` manifest to the cluster and then restores the data into it.

This guide will give you an overview how you can take application-level backup and restore your `Microsoft SQL Server` databases using `Kubestash`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore `Microsoft SQL Server` databases, please check the following guide [here](/docs/guides/mssqlserver/backup/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mssqlserver/backup/application-level/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/application-level/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


## Backup Microsoft SQL Server

KubeStash supports backups for `Microsoft SQL Server` instances across different configurations, including Standalone and  Availability Group setups. In this demonstration, we'll focus on a `Microsoft SQL Server` database using Standalone configuration. The backup and restore process is similar for Availability Group configuration.

This section will demonstrate how to backup a `Microsoft SQL Server` database. Here, we are going to deploy a `Microsoft SQL Server` database using KubeDB. Then, we are going to backup this database into a `GCS` bucket. Finally, we are going to restore the backup up data into another `Microsoft SQL Server` database.

### Deploy Sample Microsoft SQL Server Database

By default, a KubeDB-managed `Microsoft SQL Server` instance does not run with TLS enabled. However, the `.spec.tls` field is mandatory and will be used during backup and restore operations.

**Create Issuer/ClusterIssuer:**

Now, we are going to create an example `Issuer` CR that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager](https://cert-manager.io/docs/configuration/ca/) tutorial to create your own `Issuer`.

By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mssqlserver/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```

Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/application-level/examples/mssqlserver-ca-issuer-demo.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer-demo.yaml created
```

**Create MSSQLServer CR:**

Below is the YAML of a sample `MSSQLServer` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: sample-mssqlserver
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Create the above `MSSQLServer` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/application-level/examples/sample-mssqlserver.yaml
mssqlserver.kubedb.com/sample-mssqlserver created
```

KubeDB will deploy a `Microsoft SQL Server` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get mssqlserver -n demo sample-mssqlserver
NAME                 VERSION     STATUS   AGE
sample-mssqlserver   2022-cu12   Ready    3m27
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo
NAME                                    TYPE                       DATA   AGE
mssqlserver-ca                          kubernetes.io/tls          2      2d20h
sample-mssqlserver-auth                 kubernetes.io/basic-auth   2      3m44s
sample-mssqlserver-client-cert          kubernetes.io/tls          3      3m14s
sample-mssqlserver-server-cert          kubernetes.io/tls          3      3m14s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mssqlserver
NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
sample-mssqlserver        ClusterIP   10.96.165.94   <none>        1433/TCP   4m32s
sample-mssqlserver-pods   ClusterIP   None           <none>        1433/TCP   4m32s
```

Here, we have to use service `sample-mssqlserver` and secret `sample-mssqlserver-auth` to connect with the database. `KubeDB` creates an AppBinding CR that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME                 TYPE                     VERSION   AGE
sample-mssqlserver   kubedb.com/mssqlserver   2022      4m18s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-mssqlserver -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MSSQLServer","metadata":{"annotations":{},"name":"sample-mssqlserver","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"storageType":"Durable","tls":{"clientTLS":false,"issuerRef":{"apiGroup":"cert-manager.io","kind":"Issuer","name":"mssqlserver-ca-issuer"}},"version":"2022-cu12"}}
  creationTimestamp: "2024-09-20T09:09:38Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mssqlserver
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mssqlservers.kubedb.com
  name: sample-mssqlserver
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: MSSQLServer
    name: sample-mssqlserver
    uid: 212fef79-23fb-4f3a-aea9-564ce1362174
  resourceVersion: "277078"
  uid: 01955aa0-f68e-410c-b952-c8516ea24922
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MSSQLServer
    name: sample-mssqlserver
    namespace: demo
  clientConfig:
    service:
      name: sample-mssqlserver
      path: /
      port: 1433
      scheme: tcp
    url: tcp(sample-mssqlserver.demo.svc:1433)/
  secret:
    name: sample-mssqlserver-auth
  type: kubedb.com/mssqlserver
  version: "2022"
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.


**Insert Sample Data:**

Now, we are going to exec into one of the database pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-mssqlserver"
NAME                   READY   STATUS    RESTARTS   AGE
sample-mssqlserver-0   1/1     Running   0          4m44s
```

And copy the username and password of the `sa` user to access into `mssqlserver` shell.

```bash
$ kubectl get secret -n demo  sample-mssqlserver-auth -o jsonpath='{.data.username}'| base64 -d
sa⏎

$ kubectl get secret -n demo  sample-mssqlserver-auth -o jsonpath='{.data.password}'| base64 -d
kkvAFfl8sIxRO2i3⏎
```

Now, Lets exec into the `Pod` to enter into `mssqlserver` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo sample-mssqlserver-0 -c mssql -- /opt/mssql-tools/bin/sqlcmd -S sample-mssqlserver -U sa -P "kkvAFfl8sIxRO2i3"
# list available databases
1> SELECT name from sys.databases;
2> GO
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
kubedb_system                                                                                                                   

(5 rows affected)

# create a database named "playground"
1> CREATE DATABASE playground;
2> GO

# verify that the "playground" database has been created
1> SELECT name from sys.databases;
2> GO
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
kubedb_system                                                                                                                   
playground                                                                                                                      

(6 rows affected)
                                                                                                               
# Now create a 'equipment' table and insert multiple rows of data
1> USE playground;
2> CREATE TABLE equipment (id INT NOT NULL IDENTITY(1,1) PRIMARY KEY, type NVARCHAR(50), quant INT, color NVARCHAR(25));
3> INSERT INTO equipment (type, quant, color) VALUES ('Swing', 10, 'Red'), ('Slide', 5, 'Blue'), ('Monkey Bars', 3, 'Yellow');
4> GO

(3 rows affected)

# Verify that data hase been inserted successfully
1> SELECT * FROM equipment;
2> GO
id          type                                               quant       color                    
----------- -------------------------------------------------- ----------- -------------------------
          1 Swing                                                       10 Red                      
          2 Slide                                                        5 Blue                     
          3 Monkey Bars                                                  3 Yellow                   

(3 rows affected)

# exit from the pod
1> exit
```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a `GCS` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/application-level/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/application-level/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `sample-mssqlserver` Microsoft SQL Server database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

Below is the YAML for `BackupConfiguration` CR to backup the `sample-mssqlserver` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mssqlserver-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: sample-mssqlserver
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
        - name: gcs-mssqlserver-repo
          backend: gcs-backend
          directory: /mssqlserver
      addon:
        name: mssqlserver-addon
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 0
        tasks:
          - name: manifest-backup
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `sample-mssqlserver` Microsoft SQL Server database that we created earlier.
- `.spec.sessions[*].addon.tasks[*].name[*]` specifies that both the `manifest-backup` and `logical-backup` tasks will be executed.

> KubeStash utilizes [Wal-G](https://wal-g.readthedocs.io/SQLServer/) to perform logical backups of `Microsoft SQL Server` databases. Since Wal-G operates with `root` user privileges, it’s necessary to configure our backup job to run as a `root` user by specifying `runAsUser: 0` in the `spec.sessions[*].addon.jobTemplate.spec.securityContext` section.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/application-level/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-mssqlserver-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                        PHASE   PAUSED   AGE
sample-mssqlserver-backup   Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                     INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-mssqlserver-repo                 0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the GCS bucket, we will see the `Repository` YAML stored in the `demo/mssqlserver` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                                SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-mssqlserver-backup-frequent-backup   */5 * * * *   False     0        4m52s           15m
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                   INVOKER-TYPE          INVOKER-NAME                 PHASE       DURATION   AGE
sample-mssqlserver-backup-frequent-backup-1725449400   BackupConfiguration   sample-mssqlserver-backup    Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `gcs-mssqlserver-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-mssqlserver-repo
NAME                          INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-mssqlserver-repo          true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=gcs-mssqlserver-repo
NAME                                                              REPOSITORY             SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-mssqlserver-repo-sample-mssqckup-frequent-backup-1725449400   gcs-mssqlserver-repo   frequent-backup   2024-01-23T13:10:54Z   Delete            Succeeded   16h
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
$ kubectl get snapshots -n demo gcs-mssqlserver-repo-sample-mssqckup-frequent-backup-1725449400 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  annotations:
    kubedb.com/db-version: "2022"
  creationTimestamp: "2024-09-20T11:25:00Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: MSSQLServer
    kubestash.com/app-ref-name: sample-mssqlserver
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: gcs-mssqlserver-repo
  name: gcs-mssqlserver-repo-sample-mssqckup-frequent-backup-1725449400
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: gcs-mssqlserver-repo
    uid: 5774142d-a81d-44d6-9459-20c16a0d7ade
  resourceVersion: "293781"
  uid: b84c608c-8da8-444a-8db4-1632d04736e3
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MSSQLServer
    name: sample-mssqlserver
    namespace: demo
  backupSession: sample-mssqlserver-backup-frequent-backup-1725449400
  deletionPolicy: Delete
  repository: gcs-mssqlserver-repo
  session: frequent-backup
  snapshotID: 01J87JV7HTH2FW71RBCTM56QWQ
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: WalG
      duration: 19.996377s
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      walGStats:
        databases:
        - playground
        id: base_20240920T112503Z
        startTime: "2024-09-20T11:25:03Z"
        stopTime: "2024-09-20T11:25:23Z"
  conditions:
  - lastTransitionTime: "2024-09-20T11:25:00Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-20T11:25:26Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  phase: Succeeded
  snapshotTime: "2024-09-20T11:25:00Z"
  totalComponents: 1
```

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `demo/mssqlserver/repository/v1/frequent-backup/dump/basebackups_005` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/mssqlserver/snapshots` directory.

## Restore

In this section, we are going to restore the entire database from the backup that we have taken in the previous section.

For this tutorial, we will restore the database in a separate namespace called `dev`.

First, create the namespace by running the following command:

```bash
$ kubectl create ns dev
namespace/dev created
```

**Create Issuer/ClusterIssuer:**

Now, we are going to create another example `Issuer` CR  that will be used throughout the restore of this tutorial. Alternatively, you can follow this [cert-manager](https://cert-manager.io/docs/configuration/ca/) tutorial to create your own `Issuer`.

By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mssqlserver/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=dev 
secret/mssqlserver-ca created
```

Now, we are going to create an `Issuer` CR using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: dev
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/application-level/examples/mssqlserver-ca-issuer-dev.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer-dev.yaml created
```

#### Create RestoreSession:

We need to create a RestoreSession CR.

Below, is the contents of YAML file of the `RestoreSession` CR that we are going to create to restore the entire database.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-mssqlserver
  namespace: dev
spec:
  manifestOptions:
    msSQLServer:
      db: true
      restoreNamespace: dev
      tlsIssuerRef:
        name: mssqlserver-ca-issuer
        kind: Issuer
        apiGroup: cert-manager.io
  dataSource:
    namespace: demo
    repository: gcs-mssqlserver-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: mssqlserver-addon
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 0
    tasks:
      - name: manifest-restore
      - name: logical-backup-restore
```

Here,

- `.spec.manifestOptions.msSQLServer.db` specifies whether to restore the DB manifest or not.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.namespace` specifies the namespace name of Repository object.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.
- `.spec.addon.tasks[*]` specifies that both the `manifest-restore` and `logical-backup-restore` tasks.

> KubeStash utilizes [Wal-G](https://wal-g.readthedocs.io/SQLServer/) to perform logical restores of `Restore Microsoft SQL Server` databases. Since Wal-G operates with `root` user privileges, it’s necessary to configure our restore job to run as a `root` user by specifying `runAsUser: 0` in the `.spe.addon.jobTemplate.spec.securityContext` section.

> Note: Set the RestoreSession namespace and `.spe.manifestOptions.msSQLServer.restoreNamespace` to the same value, as kubeStash internally creates a proxy server. Currently, only the same namespace is supported.

Let's create the RestoreSession CR object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/application-level/examples/restoresession.yaml
restoresession.core.kubestash.com/restore-sample-mssqlserver created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n dev
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024
NAME                          REPOSITORY            FAILURE-POLICY      PHASE       DURATION   AGE
restore-sample-mssqlserver    gcs-mssqlserver-repo                      Succeeded   3s         53s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get mssqlserver -n dev sample-mssqlserver
NAME                   VERSION     STATUS   AGE
sample-mssqlserver     2022-cu12   Ready    13m
```

Now, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n dev --selector="app.kubernetes.io/instance=sample-mssqlserver"
NAME                         READY   STATUS      RESTARTS   AGE
restored-mssqlserver-0       1/1     Running     0          16m
```

And copy the username and password of the `sa` user to access into `mssqlserver` shell.

```bash
$ kubectl get secret -n dev  sample-mssqlserver-auth -o jsonpath='{.data.username}'| base64 -d
sa⏎

$ kubectl get secret -n dev  sample-mssqlserver-auth -o jsonpath='{.data.password}'| base64 -d
Ag9qi8zQiFew0xHo⏎
```

Now, Lets exec into the `Pod` to enter into `mssqlserver` shell and verify restored data,

```bash
$ kubectl exec -it -n dev sample-mssqlserver-0 -c mssql -- /opt/mssql-tools/bin/sqlcmd -S sample-mssqlserver -U sa -P "Ag9qi8zQiFew0xHo"
1> SELECT name from sys.databases;
2> GO
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
kubedb_system                                                                                                                   
playground                                                                                                                      

(6 rows affected)

1> USE playground;
2> SELECT name from sys.tables;
3> GO
Changed database context to 'playground'.
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
equipment 

(1 rows affected)

1> SELECT * FROM equipment;
2> GO
id          type                                               quant       color                    
----------- -------------------------------------------------- ----------- -------------------------
          1 Swing                                                       10 Red                      
          2 Slide                                                        5 Blue                     
          3 Monkey Bars                                                  3 Yellow                   

(3 rows affected)
> exit
```

Based on the output above, we can confirm that the `playground` database and the `equipment` table, which were previously created in the original database, have now been successfully restored.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-mssqlserver-backup
kubectl delete restoresessions.core.kubestash.com -n dev restore-sample-mssqlserver
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secrets -n demo mssqlserver-ca
kubectl delete secrets -n dev mssqlserver-ca
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete issuer -n dev mssqlserver-ca-issuer
kubectl delete mssqlserver -n demo sample-mssqlserver
kubectl delete mssqlserver -n dev sample-mssqlserver
```