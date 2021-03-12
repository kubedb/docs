---
title: Backup & Restore Percona XtraDB Cluster | Stash
description: Backup & Restore Percona XtraDB Cluster using Stash
menu:
  docs_{{ .version }}:
    identifier: guides-px-backup-cluster
    name: Percona XtraDB Cluster
    parent: guides-px-backup
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore Percona XtraDB Cluster using Stash

Stash 0.9.0+ supports backup and restoration of Percona XtraDB cluster databases. This guide will show you how you can backup and restore your Percona XtraDB cluster with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install Stash in your cluster following the steps [here](https://stash.run/docs/latest/setup/).
- Install Percona XtraDB addon for Stash following the steps [here](https://stash.run/docs/latest/addons/percona-xtradb/setup/install/)
- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- If you are not familiar with how Stash takes backup and restores Percona XtraDB, please check the following guide [here](/docs/guides/percona-xtradb/backup/overview/index.md).

You have to be familiar with the following custom resources:

- [AppBinding](/docs/guides/percona-xtradb/concepts/appbinding.md)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [RestoreSession](https://stash.run/docs/latest/concepts/crds/restoresession/)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Backup Percona XtraDB Cluster

This section will demonstrate how to backup a Percona XtraDB cluster. Here, we are going to deploy a Percona XtraDB cluster using KubeDB. Then, we are going to back up this database into a GCS bucket. Finally, we are going to restore the backed up data into another Percona XtraDB cluster.

### Deploy Sample Percona XtraDB Cluster

Let's deploy a sample Percona XtraDB cluster and insert some data into it.

#### Create Percona XtraDB CRD

Below is the YAML of a sample `PerconaXtraDB` CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-xtradb-cluster
  namespace: demo
spec:
  version: "5.7-cluster"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Create the above `PerconaXtraDB` CRD,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/backup/examples/clustered/backup/sample-xtradb-cluster.yaml
perconaxtradb.kubedb.com/sample-xtradb-cluster created
```

KubeDB will deploy a Percona XtraDB cluster according to the above specification. It will also create the necessary Secrets and Services to access the database.

Let's check if the database is ready to use,

```bash
$  kubectl get px -n demo sample-xtradb-cluster
NAME                    VERSION       STATUS    AGE
sample-xtradb-cluster   5.7-cluster   Running   7m46s
```

The database is `Running`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=kubedb.com/name=sample-xtradb-cluster
NAME                         TYPE     DATA   AGE
sample-xtradb-cluster-auth   Opaque   2      9m2s

$ kubectl get service -n demo -l=kubedb.com/name=sample-xtradb-cluster
NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
sample-xtradb-cluster       ClusterIP   10.103.37.141   <none>        3306/TCP   11m
sample-xtradb-cluster-gvr   ClusterIP   None            <none>        3306/TCP   11m
```

Here, we have to use service `sample-xtradb-cluster` and secret `sample-xtradb-cluster-auth` to connect with the database. KubeDB creates an [AppBinding](/docs/guides/percona-xtradb/concepts/appbinding.md) CRD that holds the necessary information to connect with the database.

#### Verify AppBinding

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME                    AGE
sample-xtradb-cluster   14m
```

Let's check the YAML of the above AppBinding,

```bash
$ kubectl get appbindings -n demo sample-xtradb-cluster -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2019-10-30T11:41:20Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-xtradb-cluster
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: perconaxtradbs.kubedb.com
    kubedb.com/name: sample-xtradb-cluster
  name: sample-xtradb-cluster
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: false
    kind: PerconaXtraDB
    name: sample-xtradb-cluster
    uid: 79d90fc4-f5e8-4a8c-83d7-3eae7c12f01a
  resourceVersion: "12319"
  selfLink: /apis/appcatalog.appscode.com/v1alpha1/namespaces/demo/appbindings/sample-xtradb-cluster
  uid: 977cb8fd-b5e5-4830-a50f-58de9eb5d82c
spec:
  clientConfig:
    service:
      name: sample-xtradb-cluster
      path: /
      port: 3306
      scheme: mysql
    url: tcp(sample-xtradb-cluster:3306)/
  parameters:
    address: gcomm://sample-xtradb-cluster-0.sample-xtradb-cluster-gvr.demo,sample-xtradb-cluster-1.sample-xtradb-cluster-gvr.demo,sample-xtradb-cluster-2.sample-xtradb-cluster-gvr.demo
    apiVersion: config.kubedb.com/v1alpha2
    group: sample-xtradb-cluster
    kind: GaleraArbitratorConfiguration
    sstMethod: xtrabackup-v2
    stash:
      addon:
        backupTask:
          name: percona-xtradb-backup-5.7.0-v2
        restoreTask:
          name: percona-xtradb-restore-5.7.0-v2
  secret:
    name: sample-xtradb-cluster-auth
  type: kubedb.com/perconaxtradb
  version: "5.7-cluster"
```

Stash uses the AppBinding CRD to connect with the target database. It requires the following two fields to be set in the AppBinding's `.spec` section.

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds the necessary credentials to access the database.
- `spec.parameters.stash` contains the Stash Addon info which will be used to backup and restore this database.
- `.spec.type` specifies the type of the app that this AppBinding is pointing to. The format KubeDB generated AppBinding follows to set the value of `.spec.type` is `<app_group>/<app_resource_type>`.

#### Creating AppBinding Manually

If you deploy the Percona XtraDB cluster without KubeDB, you have to create the AppBinding CRD manually in the same namespace as the service and secret of the database.

The following YAML shows a minimal AppBinding specification that you have to create if you deploy the Percona XtraDB cluster without KubeDB.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: your-custom-appbinding-name
  namespace: your-database-namespace
spec:
  clientConfig:
    service:
      name: your-database-service-name
      port: 3306
      scheme: mysql
  secret:
    name: your-database-auth-secret-name
  # type field is optional. you can keep it empty.
  # if you keep it empty then the value of TARGET_APP_RESOURCE variable
  # will be set to "appbinding" during auto-backup.
  type: kubedb.com/perconaxtradb
```

You have to replace the `<...>` quoted part with proper values in the above YAML.

#### Insert Sample Data

Now, we are going to exec into the database pod and create some sample data. At first, find out the database pods using the following command,

```bash
$ kubectl get pods -n demo --selector="kubedb.com/name=sample-xtradb-cluster"
NAME                      READY   STATUS    RESTARTS   AGE
sample-xtradb-cluster-0   1/1     Running   0          39m
sample-xtradb-cluster-1   1/1     Running   0          38m
sample-xtradb-cluster-2   1/1     Running   0          37m
```

And copy the username and password of the `root` user to access into `mysql` shell.

```bash
$ kubectl get secret -n demo  sample-xtradb-cluster-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  sample-xtradb-cluster-auth -o jsonpath='{.data.password}'| base64 -d
CZYWy7MDXiedL2EG⏎
```

Now, let's exec into the Pod to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo sample-xtradb-cluster-0 -- mysql --user=root --password=CZYWy7MDXiedL2EG
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 275
Server version: 5.7.25-28-57 Percona XtraDB Cluster (GPL), Release rel28, Revision a2ef85f, WSREP version 31.35, wsrep_31.35

Copyright (c) 2009-2019 Percona LLC and/or its affiliates
Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE playground;
Query OK, 1 row affected (0.01 sec)

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.01 sec)

mysql> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.01 sec)

mysql> INSERT INTO playground.equipment (type, quant, color) VALUES ("slide", 2, "blue");
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

Now, we are ready to back up the database.

### Prepare Backend

We are going to store our backed up data into a GCS bucket. At first, we need to create a secret with GCS credentials then we need to create a `Repository` CRD. If you want to use a different backend, please read the respective backend configuration doc from [here](https://stash.run/docs/latest/guides/latest/backends/overview/).

#### Create Storage Secret

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat downloaded-sa-json.key > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

#### Create Repository

Now, crete a `Repository` using this secret. Below is the YAML of Repository CRD we are going to create,

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcs-repo-xtradb-cluster
  namespace: demo
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: /demo/xtradb/sample-xtradb-cluster
    storageSecretName: gcs-secret
```

Let's create the `Repository` we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clustered/backup/repository.yaml
repository.stash.appscode.com/gcs-repo-xtradb-cluster created
```

Now, we are ready to back up our database to our desired backend.

### Backup

We have to create a `BackupConfiguration` targeting respective AppBinding CRD of our desired database. Then Stash will create a CronJob to periodically backup the database.

#### Create BackupConfiguration

Below is the YAML for `BackupConfiguration` CRD to backup the `sample-xtradb-cluster` database we have deployed earlier,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-xtradb-cluster-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: gcs-repo-xtradb-cluster
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-xtradb-cluster
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

Here,

- `.spec.schedule` specifies that we want to back up the database at 5 minutes interval.
- `.spec.target.ref` refers to the AppBinding CRD that was created for the `sample-xtradb-cluster` database.

Let's create the `BackupConfiguration` CRD we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clustered/backup/backupconfiguration.yaml
backupconfiguration.stash.appscode.com/sample-xtradb-cluster-backup created
```

#### Verify CronJob

If everything goes well, Stash will create a CronJob with the schedule specified in `.spec.schedule` field of `BackupConfiguration` CRD.

Verify that the CronJob has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                           SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-xtradb-cluster-backup   */5 * * * *   False     0        49s             2m22s
```

#### Wait for BackupSession

The `sample-xtradb-cluster-backup` CronJob will trigger a backup on each scheduled slot by creating a `BackupSession` CRD.

Wait for a schedule to appear. Run the following command to watch `BackupSession` CRD,

```bash
$ kubectl get backupsession -n demo -l=stash.appscode.com/invoker-name=sample-xtradb-cluster-backup --watch
NAME                                      INVOKER-TYPE          INVOKER-NAME                   PHASE       AGE
sample-xtradb-cluster-backup-1572439801   BackupConfiguration   sample-xtradb-cluster-backup   Succeeded   4m27s
```

Here, the phase **`Succeeded`** means that the backupsession has been succeeded.

>Note: Backup CronJob creates `BackupSession` CRD the label `stash.appscode.com/invoker-name=<BackupConfiguration_crd_name>`. We can use this label to watch only the `BackupSession` of our desired `BackupConfiguration`.

#### Verify Backup

Now, we are going to verify whether the backed up data is in the backend. Once a backup is completed, Stash will update the respective `Repository` CRD to reflect the backup completion. Check that the repository `gcs-repo-xtradb-cluster` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-repo-xtradb-cluster
NAME                      INTEGRITY   SIZE          SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo-xtradb-cluster   true        304.165 MiB   3                97s                      13m
```

Now, if we navigate to the GCS bucket, we will see the backed up data has been stored in `demo/xtradb/sample-xtradb-cluster` directory as specified by `.spec.backend.gcs.prefix` field of Repository CRD.

<figure align="center">
  <img alt="Backed up data in GCS Bucket" src="/docs/guides/percona-xtradb/backup/images/sample-xtradb-cluster-backup.png">
  <figcaption align="center">Fig: Backed up data in GCS Bucket</figcaption>
</figure>

> Note: Stash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

### Restore Percona XtraDB Cluster

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

#### Stop Taking Backup of the Old Database

At first, let's stop taking any further backup of the old database so that no backup is taken during the restore process. We are going to pause the `BackupConfiguration` CRD that we had created to backup the `sample-xtradb-cluster` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-xtradb-cluster-backup` BackupConfiguration,

```console
$ kubectl patch backupconfiguration -n demo sample-xtradb-cluster-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-xtradb-cluster-backup patched
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the operator has paused the BackupConfiguration object,

```console
$ kubectl get backupconfiguration -n demo sample-xtradb-cluster-backup
NAME                           TASK                        SCHEDULE      PAUSED   AGE
sample-xtradb-cluster-backup   percona-xtradb-backup-5.7.0-v2   */5 * * * *   true     50m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

#### Deploy Restored Database

Now, we have to deploy the restored database similarly as we have deployed the original `sample-xtradb-cluster` database. However, this time there will be the following differences:

- We have to use the same secret that was used in the original database. We are going to specify it using `.spec.databaseSecret` field.
- We have to specify `.spec.init.waitForInitialRestore` field to tell KubeDB to wait for first restore to complete before marking this database as ready to use.

Below is the YAML for `PerconaXtraDB` CRD we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: restored-xtradb-cluster
  namespace: demo
spec:
  version: "5.7-cluster"
  replicas: 3
  authSecret:
    name: sample-xtradb-cluster-auth
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    waitForInitialRestore: true
  terminationPolicy: WipeOut
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clustered/restore/restored-xtradb-cluster.yaml
perconaxtradb.kubedb.com/restored-xtradb-cluster created
```

If you check the database status, you will see it is stuck in **`Provisioning`** state.

```bash
$ kubectl get px -n demo restored-xtradb-cluster
NAME                      VERSION       STATUS         AGE
restored-xtradb-cluster   5.7-cluster   Provisioning   4m10s
```

#### Create RestoreSession

Now, we need to create a `RestoreSession` CRD pointing to the newly created restored database.

In case of Percona XtraDB cluster, the RestoreSession object contains some different configurations unlike other databases supported by KubeDB. To restore Percona XtraDB cluster, Stash operator will create the required number of PVCs and mount the data in the data directory `/var/lib/mysql` with proper ownership and permission. After completing the PVC creation, KubeDB then creates AppBinding, Secret, Services, etc. objects.

Below is the contents of YAML file of the RestoreSession CRD that we are going to create to restore the backed up data into the newly created database provisioned by PerconaXtrDB CRD named `restored-xtradb-cluster`.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: restored-xtradb-cluster-restore
  namespace: demo
spec:
  repository:
    name: gcs-repo-xtradb-cluster
  target:
    replicas: 3
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-xtradb-cluster
    volumeMounts:
    - name: data-restored-xtradb-cluster
      mountPath: /var/lib/mysql
    volumeClaimTemplates:
    - metadata:
        name: data-restored-xtradb-cluster-${POD_ORDINAL}
      spec:
        accessModes: [ "ReadWriteOnce" ]
        storageClassName: "standard"
        resources:
          requests:
            storage: 1Gi
  rules:
  - targetHosts: [] # empty host match all hosts
    sourceHost: "host-0" # restore same data on all pvc
    snapshots: ["latest"]
```

Here,

- `.spec.repository.name` specifies the Repository CRD that holds the backend information where our backed up data has been stored.
- `.spec.target.replicas` specifies the number of PVCs where snapshot data will be restored.
- `.spec.target.ref` refers to the  AppBinding object for the `restored-xtradb-cluster` PerconaXtraDB object. Though the KubeDB operator will create this AppBinding object later, we need to tell Stash operator about this AppBinding object ref. Because the AppBinding object name is identical with the corresponding PerconaXtraDB object name and the names of the PVCs directly depend on this name.
- `.spec.target.volumeClaimTemplates` specifies the template used for the PVCs. The important thing here is the `.metadata.name`. In KubeDB side, the PVC name is formed by following the rule `data-<xtradb_crd_object_name>-<statefulset_pod_ordinal>`. Since we have created our restore database named `restored-xtradb-cluster` and later KubeDB will create a StatefulSet for this database, the PVC names will be `data-restored-xtradb-cluster-0`, `data-restored-xtradb-cluster-1`, `data-restored-xtradb-cluster-2`, etc. up to the number of replicas. Here Stash operator will create these PVCs by following the same convention as KubeDB. We just need to provide the `.metadata.name` as `data-<xtradb_crd_object_name>-${POD_ORDINAL}`. You must insert `${POD_ORDINAL}` at the end of the name. Stash will create the required PVCs by replacing this with the corresponding pod index. That means if the value of `.spec.target.replicas` is 3, then Stash will create 3 PVCs named `data-restored-xtradb-cluster-0`, `data-restored-xtradb-cluster-1`, and `data-restored-xtradb-cluster-2`.
- `.spec.target.volumeMounts` specifies the mount path for the volume. The `mountPath` must be  `/var/lib/mysql` as expected by Percona XtraDB server. And the volume name is form as `"data-<xtradb_crd_object_name>"`. Since for restoring purpose, we have created a PerconaXtraDB object named `restored-xtradb-cluster`, the volume name will be `"data-restored-xtradb-cluster"`.
- `.spec.rules` specifies that we are restoring data from the `latest` backup snapshot of the database. Empty (`[]`) `targetHosts` means snapshot data will be restored in all specified number of PVCs. And another obvious thing is we want to restore the same data from `host-0` to all PVCs. During the backup procedure, we took backup data as `host-0` from the Percona XtraDB cluster. So, here the source host is `host-0`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clustered/restore/restoresession.yaml
restoresession.stash.appscode.com/restored-xtradb-cluster-restore created
```

Once you have created the RestoreSession object, Stash will create a restore Job. We can watch the phase of the RestoreSession object to check whether the restore process has succeeded or not.

Run the following command to watch the phase of the RestoreSession object,

```bash
$ kubectl get restoresession -n demo restored-xtradb-cluster-restore --watch
NAME                              REPOSITORY                PHASE     AGE
restored-xtradb-cluster-restore   gcs-repo-xtradb-cluster   Running   3m33s
restored-xtradb-cluster-restore   gcs-repo-xtradb-cluster   Running   3m51s
restored-xtradb-cluster-restore   gcs-repo-xtradb-cluster   Running   3m58s
restored-xtradb-cluster-restore   gcs-repo-xtradb-cluster   Succeeded   3m58s
```

Here, we can see from the output of the above command that the restore process succeeded.

#### Verify Restored Data

In this section, we are going to verify whether the desired data has restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database have restored.

At first, check if the database has gone into **`Running`** state,

```bash
$ kubectl get px -n demo restored-xtradb-cluster --watch
NAME                      VERSION       STATUS         AGE
restored-xtradb-cluster   5.7-cluster   Provisioning   3m36s
restored-xtradb-cluster   5.7-cluster   Provisioning   4m4s
restored-xtradb-cluster   5.7-cluster   Running        4m4s
```

Now, find out the database Pod,

```bash
$ kubectl get pods -n demo --selector="kubedb.com/name=restored-xtradb-cluster" --watch
NAME                        READY   STATUS    RESTARTS   AGE
restored-xtradb-cluster-0   1/1     Running   0          115s
restored-xtradb-cluster-1   1/1     Running   0          77s
restored-xtradb-cluster-2   1/1     Running   0          41s
```

And then copy the user name and password of the `root` user to access into `mysql` shell.

> Notice: We used the same Secret for the `restored-xtradb-cluster` object. So, we will use the same commands as before.

```bash
$ kubectl get secret -n demo  sample-xtradb-cluster-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  sample-xtradb-cluster-auth -o jsonpath='{.data.password}'| base64 -d
CZYWy7MDXiedL2EG⏎
```

Now, let's exec into the Pod to enter into `mysql` shell and check the database and the table we created before,

```bash
$ kubectl exec -it -n demo restored-xtradb-cluster-0 -- mysql --user=root --password=CZYWy7MDXiedL2EG
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 275
Server version: 5.7.25-28-57 Percona XtraDB Cluster (GPL), Release rel28, Revision a2ef85f, WSREP version 31.35, wsrep_31.35

Copyright (c) 2009-2019 Percona LLC and/or its affiliates
Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.00 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

So, from the above output, we can see that the `playground` database and the `equipment` table we created before in the original database are restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete restoresession -n demo restored-xtradb-cluster-restore
kubectl delete px -n demo restored-xtradb-cluster
kubectl delete repository -n demo gcs-repo-xtradb-cluster
kubectl delete backupconfiguration -n demo sample-xtradb-cluster-backup
kubectl delete px -n demo sample-xtradb-cluster
```
