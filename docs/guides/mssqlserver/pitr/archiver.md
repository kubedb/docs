---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: pitr-mssqlserver-archiver
    name: Overview
    parent: pitr-mssqlserver
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MSSQLServer - Continuous Archiving and Point-in-time Recovery

Here, will show you how to use KubeDB to provision a Microsoft SQL Server to Archive continuously and Restore point-in-time.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, 
- install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).
- install `KubeStash` operator in your cluster following the steps [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/mssqlserver/pitr/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mssqlserver/pitr/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Continuous Archiving

Continuous archiving involves making regular copies (or "archives") of the Microsoft SQL Server transaction log files.To ensure continuous archiving to a remote location we need prepare `BackupStorage`, `RetentionPolicy`, `MSSQLServerArchiver` for the KubeDB Managed Microsoft SQL Server Databases.

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

Now, create a `BackupStorage` that references this secret. Below is the YAML of `BackupStorage` CR we are going to create,

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
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/pitr/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/gcs-storage created
```

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/pitr/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

**Create Encryption Secret**

Let’s create a secret called encrypt-secret with the `Restic` password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret "encrypt-secret" created
```

**Create MSSQLServerArchiver CR:**

`MSSQLServerArchiver` is a CR provided by KubeDB for managing the archiving of `MSSQLServer` transaction log files and performing volume-level backups.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MSSQLServerArchiver
metadata:
  name: mssqlserverarchiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: demo-retention
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: WalG
    task:
      params:
        databases: demo
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "/30 * * * *"
    sessionHistoryLimit: 2
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 0
  walBackup:
    runtimeSettings:
      pod:
        securityContext:
          runAsUser: 0
      container:
        securityContext:
          runAsUser: 0
  backupStorage:
    ref:
      name: gcs-storage
      namespace: demo
```

Let’s create the above `MSSQLServerArchiver`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/pitr/examples/sample-mssqlserverarchiver.yaml
mssqlserverarchiver.archiver.kubedb.com/sample-mssqlserverarchiver created
```

Here,
- The `databases` field within `spec.fullBackup.task.params` specifies the target databases for the archive. If no database list is provided, the archiver will target all non-system databases by default.


> The `KubeDB` provisioner uses `KubeStash` for full-backup and `mssqlserver-archiver` for transaction log-backup. Both utilizes `Wal-G` to perform backups of Microsoft SQL Server databases. Since `Wal-G` operates with `root` user privileges, it’s necessary to configure our full-backup job and `archiver pod` to run as a `root` user.

### Deploy Sample Microsoft SQL Server Database

First, an `Issuer/ClusterIssuer` needs to be created, even if TLS is not enabled for `Microsoft SQL Server`. The issuer will be used to configure the TLS-enabled `Wal-G proxy server`, which is required for the SQL Server backup and restore operations.

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

Now, we are going to create an `Issuer` CR using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` CR that we are going to create,

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/pitr/examples/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer.yaml created
```

**Create MSSQLServer CR:**

Below is the YAML of a sample `MSSQLServer` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: sample-mssqlserver-ag
  namespace: demo
  labels:
    archiver: "true"
spec:
  healthChecker:
    timeoutSeconds: 100
  archiver:
    ref:
      name: sample-mssqlserverarchiver
      namespace: demo
  version: "2022-cu12"
  replicas: 2
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - demo
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssqlserver-ca-issuer
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation # Change it 
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Here,
- We don't have to specify `spec.archiver.ref` field. The `archiver` will be auto-selected using the `archiver.spec.databases` field.

Create the above `MSSQLServer` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/pitr/examples/sample-mssqlserver-ag.yaml
mssqlserver.kubedb.com/sample-mssqlserver-ag created
```

Let’s check the pods which are related to the backup,

```bash
$ kubectl get pods -n demo
NAME                                                          READY   STATUS    RESTARTS   AGE
sample-mssqlserver-ag-0                                       2/2     Running   0          6m18s
sample-mssqlserver-ag-1                                       2/2     Running   0          6m12s
sample-mssqlserver-ag-archiver-full-backup-1728973299-7gmh5   1/1     Running   0          41s
sample-mssqlserver-ag-sidekick                                1/1     Running   0          17s
```

Here,
- Pod `sample-mssqlserver-ag-archiver-full-backup-1728973299-7gmh5` is responsible for application backup. i.e (target databases and manifest)
- Pod `sample-mssqlserver-ag-sidekick` is used for transaction log-backup. This pod always be in running phase due to continuous archiving.  


### Verify Backup Setup Successful

If everything goes well, kubedb provisioner will create a `BackupConfiguration` and  the phase should be Ready. The `Ready` phase indicates that the backup setup is successful. 

Let’s verify the Phase of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                             PHASE   PAUSED   AGE
sample-mssqlserver-ag-archiver   Ready            7m49s
```

***Verify BackupSession:***

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo 
NAME                                                    INVOKER-TYPE          INVOKER-NAME                     PHASE       DURATION   AGE
sample-mssqlserver-ag-archiver-full-backup-1728973299   BackupConfiguration   sample-mssqlserver-ag-archiver   Succeeded   51s        8m31s
```

***Verify Snapshot:***

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=sample-mssqlserver-ag-archiver
NAME                                                              REPOSITORY                       SESSION       SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
sample-mssqlserver-ag-archiver-sarchiver-full-backup-1728973299   sample-mssqlserver-ag-archiver   full-backup   2024-10-15T06:21:39Z   Delete            Succeeded   10m
```

### Insert Some Data

Every successful transaction log will be recorded during the log backup by Sidekick. By default, log backups occur at `25-second` intervals.

```bash
$ kubectl get secret -n demo  sample-mssqlserver-ag-auth -o jsonpath='{.data.username}'| base64 -d
sa⏎

$ kubectl get secret -n demo  sample-mssqlserver-ag-auth -o jsonpath='{.data.password}'| base64 -d
XhGrsDvJ7ATrPp7n⏎

$ kubectl exec -it -n demo sample-mssqlserver-ag-0 -c mssql -- /opt/mssql-tools/bin/sqlcmd -S sample-mssqlserver-ag -U sa -P "XhGrsDvJ7ATrPp7n"
1> SELECT name from sys.databases;
2> GO
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
demo                                                                                                                            
kubedb_system                                                                                                                   

(6 rows affected)


# Now create a 'equipment' table and insert multiple rows of data
1> USE demo;
2> CREATE TABLE equipment (id INT NOT NULL IDENTITY(1,1) PRIMARY KEY, type NVARCHAR(50), quant INT, color NVARCHAR(25));
3> INSERT INTO equipment (type, quant, color) VALUES ('Swing', 10, 'Red'), ('Slide', 5, 'Blue'), ('Monkey Bars', 3, 'Yellow');
4> INSERT INTO equipment (type, quant, color) VALUES ('Seesaw', 4, 'Green'), ('Trampoline', 2, 'Orange'), ('Climbing Wall', 6, 'Purple');
5> INSERT INTO equipment (type, quant, color) VALUES ('Sandbox', 8, 'Brown'), ('Balance Beam', 1, 'Pink'), ('Tire Swing', 7, 'Black'), ('Ladder', 9, 'White');
6> GO
Changed database context to 'demo'.

(3 rows affected)

(3 rows affected)

(4 rows affected

# Verify that data hase been inserted successfully
1> SELECT * FROM equipment;
2> GO
id          type                                               quant       color                    
----------- -------------------------------------------------- ----------- -------------------------
          1 Swing                                                       10 Red                      
          2 Slide                                                        5 Blue                     
          3 Monkey Bars                                                  3 Yellow                   
          4 Seesaw                                                       4 Green                    
          5 Trampoline                                                   2 Orange                   
          6 Climbing Wall                                                6 Purple                   
          7 Sandbox                                                      8 Brown                    
          8 Balance Beam                                                 1 Pink                     
          9 Tire Swing                                                   7 Black                    
         10 Ladder                                                       9 White                    

(10 rows affected)


# Number of rows in equipment table 
1> SELECT COUNT(*) FROM equipment
2> GO
           
-----------
         10

(1 rows affected)

# At this point we have a table named `equipment` with 10 rows database `demo`. we will restore here.
1> SELECT GETDATE();
2> GO
                       
-----------------------
2024-10-15 09:57:10.530

(1 rows affected)

# exit from the pod
1> exit
```

### Point-in-time Recovery

Point-In-Time Recovery allows you to restore a `Microsoft SQL Server` database to a specific point in time using the archived transaction logs. This is particularly useful in scenarios where you need to recover to a state just before a specific error or data corruption occurred. 

Let’s say accidentally drops the table `equipment`.

```bash
$ kubectl exec -it -n demo sample-mssqlserver-ag-0 -c mssql -- /opt/mssql-tools/bin/sqlcmd -S sample-mssqlserver-ag -U sa -P "XhGrsDvJ7ATrPp7n"
1> use demo
2> DROP table equipment;
3> GO

Changed database context to 'demo'.

1> SELECT name from sys.tables;
2> GO
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------

(0 rows affected) # It confirms that no tables are exist in `demo` database.
```

We can’t restore from a full backup since at this point no full backup was perform. So we can choose a specific time in which time we want to restore.

For the demo, I will use the time previous drop table,

```bash
-----------------------
2024-10-15 09:57:10.530
```

**Create Restored MSSQLServer CR:**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: restored-mssqlserver-ag
  namespace: demo
spec:
  healthChecker:
    timeoutSeconds: 200
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      recoveryTimestamp: "2024-10-15T09:57:10.530Z"
      fullDBRepository:
        name: sample-mssqlserver-ag-archiver
        namespace: demo
  version: "2022-cu12"
  replicas: 2
  topology:
    mode: AvailabilityGroup
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: cert-manager.io
    clientTLS: false
  podTemplate:
    spec:
      containers:
        - name: mssql
          env:
            - name: ACCEPT_EULA
              value: "Y"
            - name: MSSQL_PID
              value: Evaluation # Change it 
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f restored-mssqlserver-ag.yaml
mssqlserver.kubedb.com/restored-mssqlserver-ag created
```

Let’s check the pods which are related to the restore,

```bash
$ kubectl get pods -n demo
NAME                                                 READY   STATUS      RESTARTS   AGE
restored-mssqlserver-ag-0                            2/2     Running     0          2m10s
restored-mssqlserver-ag-1                            2/2     Running     0          2m3s
restored-mssqlserver-ag-full-backup-restorer-7kpn8   0/1     Completed   0          65s
restored-mssqlserver-ag-log-restorer-d9sjd           1/1     Running     0          11s
restored-mssqlserver-ag-manifest-restorer-7kpn8      0/1     Completed   0          2m31s
```

Here,
- Pod `restored-mssqlserver-ag-manifest-restorer-7kpn8` is responsible for manifest restore. 
- Pod `restored-mssqlserver-ag-full-backup-restorer-6kpny` is use for full-backup restore.
- Pod `restored-mssqlserver-ag-log-restorer-d9sjd` is responsible for transaction logs restore.

> Note: Restore process works sequentially. Manifest Restore --> Full-backup Restore -->  Transaction Logs Restore.

#### Verify Restored Data:

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get mssqlserver -n demo restored-mssqlserver-ag 
NAME                      VERSION     STATUS   AGE
restored-mssqlserver-ag   2022-cu12   Ready    10m
```

Now, Lets exec into the Pod to enter into mssqlserver shell and verify restored data,

```bash
$ kubectl get secret -n demo  restored-mssqlserver-ag-auth -o jsonpath='{.data.username}'| base64 -d
sa⏎

$ kubectl get secret -n demo  restored-mssqlserver-ag-auth -o jsonpath='{.data.password}'| base64 -d
Q2YKiGgqr5ju62NL⏎

$ kubectl exec -it -n demo restored-mssqlserver-ag-0 -c mssql -- /opt/mssql-tools/bin/sqlcmd -S restored-mssqlserver-ag -U sa -P "Q2YKiGgqr5ju62NL"
1> SELECT name from sys.databases;
2> GO
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
demo                                                                                                                            
kubedb_system                                                                                                                   

(6 rows affected)
1> use demo;
2> SELECT name from sys.tables;
3> GO
Changed database context to 'demo'.
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
          4 Seesaw                                                       4 Green                    
          5 Trampoline                                                   2 Orange                   
          6 Climbing Wall                                                6 Purple                   
          7 Sandbox                                                      8 Brown                    
          8 Balance Beam                                                 1 Pink                     
          9 Tire Swing                                                   7 Black                    
         10 Ladder                                                       9 White                    

(10 rows affected)
1> SELECT COUNT(*) FROM equipment;
2> GO
           
-----------
         10

(1 rows affected)
1> exit
```

So, we are able to successfully recover from a disaster.

### Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mssqlserverarchivers.archiver.kubedb.com  -n demo sample-mssqlserverarchiver
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete mssqlserver -n demo restored-mssqlserver-ag
kubectl delete mssqlserver -n demo sample-mssqlserver-ag
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secrets -n demo mssqlserver-ca
kubectl delete issuer -n demo mssqlserver-ca-issuer
```

