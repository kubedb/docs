---
title: Backup & Restore MySQL | Stash
description: Backup MySQL database using Stash
menu:
  docs_{{ .version }}:
    identifier: mysql-guide-{{ .subproject_version }}
    name: Backup & Restore MySQL
    parent: stash-mysql-guides-{{ .subproject_version }}
    weight: 10
product_name: stash
menu_name: docs_{{ .version }}
section_menu_id: stash-addons
---

# Backup and Restore MySQL database using Stash

Stash 0.9.0+ supports backup and restoration of MySQL databases. This guide will show you how you can backup and restore your MySQL database with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install Stash in your cluster following the steps [here](/docs/setup/README.md).
- Install MySQL addon for Stash following the steps [here](/docs/addons/mysql/setup/install.md)
- Install [KubeDB](https://kubedb.com) in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/install/). This step is optional. You can deploy your database using any method you want. We are using KubeDB because KubeDB simplifies many of the difficult or tedious management tasks of running a production grade databases on private and public clouds.
- If you are not familiar with how Stash backup and restore MySQL databases, please check the following guide [here](/docs/addons/mysql/overview.md).

You have to be familiar with following custom resources:

- [AppBinding](/docs/concepts/crds/appbinding.md)
- [Function](/docs/concepts/crds/function.md)
- [Task](/docs/concepts/crds/task.md)
- [BackupConfiguration](/docs/concepts/crds/backupconfiguration.md)
- [RestoreSession](/docs/concepts/crds/restoresession.md)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored [here](https://github.com/stashed/mysql/tree/{{< param "info.subproject_version" >}}/docs/examples).

## Backup MySQL

This section will demonstrate how to backup MySQL database. Here, we are going to deploy a MySQL database using KubeDB. Then, we are going to backup this database into a GCS bucket. Finally, we are going to restore the backed up data into another MySQL database.

### Deploy Sample MySQL Database

Let's deploy a sample MySQL database and insert some data into it.

**Create MySQL CRD:**

Below is the YAML of a sample MySQL CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.0.21-v1"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: WipeOut
```

Create the above `MySQL` CRD,

```bash
$ kubectl apply -f https://github.com/stashed/mysql/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/sample-mysql.yaml
mysql.kubedb.com/sample-mysql created
```

KubeDB will deploy a MySQL database according to the above specification. It will also create the necessary Secrets and Services to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get my -n demo sample-mysql
NAME           VERSION   STATUS    AGE
sample-mysql   8.0.14    Running   4m22s
```

The database is `Running`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=kubedb.com/name=sample-mysql
NAME                TYPE     DATA   AGE
sample-mysql-auth   Opaque   2      4m58s

$ kubectl get service -n demo -l=kubedb.com/name=sample-mysql
NAME               TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
sample-mysql       ClusterIP   10.101.2.138   <none>        3306/TCP   5m33s
sample-mysql-gvr   ClusterIP   None           <none>        3306/TCP   5m33s
```

Here, we have to use service `sample-mysql` and secret `sample-mysql-auth` to connect with the database. KubeDB creates an [AppBinding](/docs/concepts/crds/appbinding.md) CRD that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the AppBinding has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME           AGE
sample-mysql   9m24s
```

Let's check the YAML of the above AppBinding,

```bash
$ kubectl get appbindings -n demo sample-mysql -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2019-09-27T05:07:34Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mysql
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mysqls.kubedb.com
  name: sample-mysql
  namespace: demo
spec:
  clientConfig:
    service:
      name: sample-mysql
      path: /
      port: 3306
      scheme: mysql
    url: tcp(sample-mysql:3306)/
  secret:
    name: sample-mysql-auth
  parameters:
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: StashAddon
    stash:
      addon:
        backupTask:
          name: mysql-backup-{{< param "info.subproject_version" >}}
        restoreTask:
          name: mysql-restore-{{< param "info.subproject_version" >}}
  type: kubedb.com/mysql
  version: "8.0.21"
```

Stash uses the AppBinding CRD to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `spec.parameters.stash` specifies the Stash Addon info that will be used to backup and restore this database.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Creating AppBinding Manually:**

If you deploy MySQL database without KubeDB, you have to create the AppBinding CRD manually in the same namespace as the service and secret of the database.

The following YAML shows a minimal AppBinding specification that you have to create if you deploy MySQL database without KubeDB.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: my-custom-appbinding-name
  namespace: my-database-namespace
spec:
  clientConfig:
    service:
      name: my-database-service-name
      port: 3306 # my_database_port_number
      scheme: mysql
  secret:
    name: my-database-credentials-secret-name
  # type field is optional. you can keep it empty.
  # if you keep it empty then the value of TARGET_APP_RESOURCE variable
  # will be set to "appbinding" during auto-backup.
  type: mysql
```

You have to replace the `<...>` quoted part with proper values in the above YAML.

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database Pod using the following command,

```bash
$ kubectl get pods -n demo --selector="kubedb.com/name=sample-mysql"
NAME             READY   STATUS    RESTARTS   AGE
sample-mysql-0   1/1     Running   0          33m
```

And copy the user name and password of the `root` user to access into `mysql` shell.

```bash
$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.password}'| base64 -d
5HEqoozyjgaMO97N⏎
```

Now, let's exec into the Pod to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo sample-mysql-0 -- mysql --user=root --password=5HEqoozyjgaMO97N
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 10
Server version: 8.0.14 MySQL Community Server - GPL

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

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a GCS bucket. At first, we need to create a secret with GCS credentials then we need to create a `Repository` CRD. If you want to use a different backend, please read the respective backend configuration doc from [here](/docs/guides/latest/backends/overview.md).

**Create Storage Secret:**

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

**Create Repository:**

Now, crete a `Repository` using this secret. Below is the YAML of Repository CRD we are going to create,

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcs-repo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: /demo/mysql/sample-mysql
    storageSecretName: gcs-secret
```

Let's create the `Repository` we have shown above,

```bash
$ kubectl create -f https://github.com/stashed/mysql/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/repository.yaml
repository.stash.appscode.com/gcs-repo created
```

Now, we are ready to backup our database to our desired backend.

### Backup

We have to create a `BackupConfiguration` targeting respective AppBinding CRD of our desired database. Then Stash will create a CronJob to periodically backup the database.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CRD to backup the `sample-mysql` database we have deployed earlier,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mysql-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  # task: # Uncomment if you are not using KubeDB.
  #   name: mysql-backup-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mysql
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

Here,

- `.spec.schedule` specifies that we want to backup the database at 5 minutes interval.
- `.spec.task.name` specifies the name of the Task CRD that specifies the necessary Functions and their execution order to backup a MySQL database.
- `.spec.target.ref` refers to the AppBinding CRD that was created for `sample-mysql` database.

Let's create the `BackupConfiguration` CRD we have shown above,

```bash
$ kubectl create -f https://github.com/stashed/mysql/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/backupconfiguration.yaml
backupconfiguration.stash.appscode.com/sample-mysql-backup created
```

**Verify CronJob:**

If everything goes well, Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` CRD.

Verify that the CronJob has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                  SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-mysql-backup   */5 * * * *   False     0        <none>          27s
```

**Wait for BackupSession:**

The `sample-mysql-backup` CronJob will trigger a backup on each scheduled slot by creating a `BackupSession` CRD.

Wait for a schedule to appear. Run the following command to watch `BackupSession` CRD,

```bash
$ watch -n 1 kubectl get backupsession -n demo -l=stash.appscode.com/backup-configuration=sample-mysql-backup

Every 1.0s: kubectl get backupsession -n demo -l=stash.appscode.com/backup-configuration=sample-mysql-backup   workstation: Fri Sep 27 11:14:43 2019

NAME                             INVOKER-TYPE          INVOKER-NAME          PHASE       AGE
sample-mysql-backup-1569561245   BackupConfiguration   sample-mysql-backup   Succeeded   38s
```

Here, the phase **`Succeeded`** means that the backupsession has been succeeded.

>Note: Backup CronJob creates `BackupSession` crds with the following label `stash.appscode.com/backup-configuration=<BackupConfiguration crd name>`. We can use this label to watch only the `BackupSession` of our desired `BackupConfiguration`.

**Verify Backup:**

Now, we are going to verify whether the backed up data is in the backend. Once a backup is completed, Stash will update the respective `Repository` CRD to reflect the backup completion. Check that the repository `gcs-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-repo
NAME       INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo   true        6.815 MiB   1                3m39s                    30m
```

Now, if we navigate to the GCS bucket, we will see the backed up data has been stored in `demo/mysql/sample-mysql` directory as specified by `.spec.backend.gcs.prefix` field of Repository CRD.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="../images/sample-mysql-backup.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

> Note: Stash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

## Restore MySQL

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

**Stop Taking Backup of the Old Database:**

At first, let's stop taking any further backup of the old database so that no backup is taken during restore process. We are going to pause the `BackupConfiguration` crd that we had created to backup the `sample-mysql` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-mysql-backup` BackupConfiguration,

```console
$ kubectl patch backupconfiguration -n demo sample-mysql-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mysql-backup patched
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo sample-mysql-backup
NAME                 TASK                  SCHEDULE      PAUSED   AGE
sample-mysql-backup  mysql-backup-{{< param "info.subproject_version" >}}   */5 * * * *   true     26m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

**Deploy Restored Database:**

Now, we have to deploy the restored database similarly as we have deployed the original `sample-mysql` database. However, this time there will be the following differences:

- We are going to specify `.spec.init.waitForInitialRestore` field that tells KubeDB to wait for first restore to complete before marking this database is ready to use.

Below is the YAML for `MySQL` CRD we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: restored-mysql
  namespace: demo
spec:
  version: "8.0.21-v1"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  init:
    waitForInitialRestore: true
  terminationPolicy: WipeOut
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/stashed/mysql/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/restored-mysql.yaml
mysql.kubedb.com/restored-mysql created
```

If you check the database status, you will see it is stuck in **`Provisioning`** state.

```bash
$ kubectl get my -n demo restored-mysql
NAME             VERSION   STATUS         AGE
restored-mysql   8.0.14    Provisioning   61s
```

**Create RestoreSession:**

Now, we need to create a RestoreSession CRD pointing to the AppBinding for this restored database.

Using the following command, check that another AppBinding object has been created for the `restored-mysql` object,

```bash
$ kubectl get appbindings -n demo restored-mysql
NAME             AGE
restored-mysql   6m6s
```

> If you are not using KubeDB to deploy database, create the AppBinding manually.

Below is the contents of YAML file of the RestoreSession CRD that we are going to create to restore backed up data into the newly created database provisioned by MySQL CRD named `restored-mysql`.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mysql-restore
  namespace: demo
spec:
  # task: # Uncomment if you are not using KubeDB.
  #   name: mysql-restore-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-mysql
  rules:
    - snapshots: [latest]
```

Here,

- `.spec.task.name` specifies the name of the Task CRD that specifies the necessary Functions and their execution order to restore a MySQL database.
- `.spec.repository.name` specifies the Repository CRD that holds the backend information where our backed up data has been stored.
- `.spec.target.ref` refers to the newly created AppBinding object for the `restored-mysql` MySQL object.
- `.spec.rules` specifies that we are restoring data from the latest backup snapshot of the database.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/stashed/mysql/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/restoresession.yaml
restoresession.stash.appscode.com/sample-mysql-restore created
```

Once, you have created the RestoreSession object, Stash will create a restore Job. We can watch the phase of the RestoreSession object to check whether the restore process has succeeded or not.

Run the following command to watch the phase of the RestoreSession object,

```bash
$ watch -n 1 kubectl get restoresession -n demo restore-sample-mysql

Every 1.0s: kubectl get restoresession -n demo  restore-sample-mysql    workstation: Fri Sep 27 11:18:51 2019
NAMESPACE   NAME                   REPOSITORY-NAME   PHASE       AGE
demo        restore-sample-mysql   gcs-repo          Succeeded   59s
```

Here, we can see from the output of the above command that the restore process succeeded.

**Verify Restored Data:**

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Running`** state by the following command,

```bash
$ kubectl get my -n demo restored-mysql
NAME             VERSION   STATUS    AGE
restored-mysql   8.0.14    Running   34m
```

Now, find out the database Pod by the following command,

```bash
$ kubectl get pods -n demo --selector="kubedb.com/name=restored-mysql"
NAME               READY   STATUS    RESTARTS   AGE
restored-mysql-0   1/1     Running   0          39m
```

And then copy the user name and password of the `root` user to access into `mysql` shell.

> Notice: We used the same Secret for the `restored-mysql` object. So, we will use the same commands as before.

```bash
$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.password}'| base64 -d
5HEqoozyjgaMO97N⏎
```

Now, let's exec into the Pod to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo restored-mysql-0 -- mysql --user=root --password=5HEqoozyjgaMO97N
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 9
Server version: 8.0.14 MySQL Community Server - GPL

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

So, from the above output, we can see that the `playground` database and the `equipment` table we created earlier in the original database and now, they are restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfiguration -n demo sample-mysql-backup
kubectl delete restoresession -n demo restore-sample-mysql
kubectl delete repository -n demo gcs-repo
kubectl delete my -n demo restored-mysql
kubectl delete my -n demo sample-mysql
```
