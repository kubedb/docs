---
title: MySQL Semi-synchronous cluster guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-clustering-semi-sync
    name: MySQL Semi-sync cluster Guide
    parent: guides-mysql-clustering
    weight: 23
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MySQL Semi-sync cluster 

This tutorial will show you how to use KubeDB to provision a MySQL semi-synchronous cluster.

## Before You Begin

Before proceeding:

- Read [mysql semi synchronous concept](/docs/guides/mysql/clustering/overview/index.md) to learn about MySQL Semi sync cluster.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/clustering/semi-sync/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/clustering/semi-sync/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MySQL Semi-sync Cluster

To deploy a single primary MySQL semi-synchronous cluster , specify `spec.topology` field in `MySQL` CRD.

The following is an example `MySQL` object which creates a semi-synchronous cluster with three members (one is primary member and the two others are secondary members).

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: semi-sync-mysql
  namespace: demo
spec:
  version: "9.1.0"
  replicas: 3
  topology:
    mode: SemiSync
    semiSync:
      sourceWaitForReplicaCount: 1
      sourceTimeout: 23h
      errantTransactionRecoveryPolicy: PseudoTransaction
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
   deletionPolicy: WipeOut

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/semi-sync/yamls/semi-sync.yaml
mysql.kubedb.com/semi-sync-mysql created
```

Here,

- `spec.topology` tells about the clustering configuration for MySQL.
- `spec.topology.mode` specifies the mode for MySQL cluster. Here we have used `SemiSync` to tell the operator that we want to deploy a MySQL Semi-synchronous cluster.
- `spec.topology.semiSync` contains semi-synchronous cluster  info.
- `spec.topology.semiSync.sourceWaitForReplicaCount:` explains the number of replica semi-sync  primary wait before commit a transaction
- `spec.topology.semiSync.sourceTimeout:` explains the timeout  for primary to wait for a replica and fall back to asynchronous replication
- `spec.topology.semiSync.errantTransactionRecoveryPolicy:` it's possible to have errant transaction during a Primary failover . kubedb supports two types of recovery using `PseudoTransaction` and `Clone`
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, KubeDB operator will create a new PetSet and a Service with the matching MySQL object name. KubeDB operator will also create a governing service for the PetSet with the name `<mysql-object-name>-pods`.

```bash
$ kubectl dba describe my -n demo semi-sync-mysql
Name:               semi-sync-mysql
Namespace:          demo
CreationTimestamp:  Wed, 16 Nov 2022 11:45:53 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"semi-sync-mysql","namespace":"demo"},"spec":{"replicas":3,"stor...
Replicas:           3  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  WipeOut

PetSet:          
  Name:               semi-sync-mysql
  CreationTimestamp:  Wed, 16 Nov 2022 11:45:53 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=semi-sync-mysql
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:        <none>
  Replicas:           824640444456 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         semi-sync-mysql
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=semi-sync-mysql
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.121.252
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.18:3306

Service:        
  Name:         semi-sync-mysql-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=semi-sync-mysql
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.18:3306,10.244.0.20:3306,10.244.0.22:3306
  Port:         coordinator  2380/TCP
  TargetPort:   coordinator/TCP
  Endpoints:    10.244.0.18:2380,10.244.0.20:2380,10.244.0.22:2380
  Port:         coordinatclient  2379/TCP
  TargetPort:   coordinatclient/TCP
  Endpoints:    10.244.0.18:2379,10.244.0.20:2379,10.244.0.22:2379

Service:        
  Name:         semi-sync-mysql-standby
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=semi-sync-mysql
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.133.61
  Port:         standby  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.20:3306,10.244.0.22:3306

Auth Secret:
  Name:         semi-sync-mysql-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=semi-sync-mysql
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"MySQL","metadata":{"annotations":{},"name":"semi-sync-mysql","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","topology":{"mode":"SemiSync","semiSync":{"errantTransactionRecoveryPolicy":"PseudoTransaction","sourceTimeout":"23h","sourceWaitForReplicaCount":1}},"version":"9.1.0"}}

    Creation Timestamp:  2022-11-16T05:45:53Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    semi-sync-mysql
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysqls.kubedb.com
    Name:                            semi-sync-mysql
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    semi-sync-mysql
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(semi-sync-mysql.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  mysql-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  mysql-restore-8.0.21
    Secret:
      Name:   semi-sync-mysql-auth
    Type:     kubedb.com/mysql
    Version:  9.1.0

Events:
  Type    Reason         Age   From            Message
  ----    ------         ----  ----            -------
  Normal  Phase Changed  7m    MySQL operator  phase changed from  to Provisioning reason:
  Normal  Successful     7m    MySQL operator  Successfully created governing service
  Normal  Successful     7m    MySQL operator  Successfully created service for primary/standalone
  Normal  Successful     7m    MySQL operator  Successfully created service for secondary replicas
  Normal  Successful     7m    MySQL operator  Successfully created database auth secret
  Normal  Successful     7m    MySQL operator  Successfully created PetSet
  Normal  Successful     7m    MySQL operator  Successfully created MySQL
  Normal  Successful     7m    MySQL operator  Successfully created appbinding
  Normal  Successful     7m    MySQL operator  Successfully patched governing service
  Normal  Successful     6m    MySQL operator  Successfully patched governing service
  Normal  Successful     6m    MySQL operator  Successfully patched governing service
  Normal  Successful     6m    MySQL operator  Successfully patched governing service
  Normal  Successful     6m    MySQL operator  Successfully patched governing service
  Normal  Successful     5m    MySQL operator  Successfully patched governing service
  Normal  Successful     5m    MySQL operator  Successfully patched governing service
  Normal  Successful     5m    MySQL operator  Successfully patched governing service
  Normal  Successful     5m    MySQL operator  Successfully patched governing service
  Normal  Phase Changed  5m    MySQL operator  phase changed from Provisioning to Ready reason:
  Normal  Successful     5m    MySQL operator  Successfully patched governing service
g


$ kubectl get petset -n demo
NAME       READY   AGE
semi-sync-mysql   3/3     3m47s

$ kubectl get pvc -n demo
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-semi-sync-mysql-0   Bound    pvc-4f8538f6-a6ce-4233-b533-8566852f5b98   1Gi        RWO            standard       4m16s
data-semi-sync-mysql-1   Bound    pvc-8823d3ad-d614-4172-89ac-c2284a17f502   1Gi        RWO            standard       4m11s
data-semi-sync-mysql-2   Bound    pvc-94f1c312-50e3-41e1-94a8-a820be0abc08   1Gi        RWO            standard       4m7s
s

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
pvc-4f8538f6-a6ce-4233-b533-8566852f5b98   1Gi        RWO            Delete           Bound    demo/data-semi-sync-mysql-0   standard                4m39s
pvc-8823d3ad-d614-4172-89ac-c2284a17f502   1Gi        RWO            Delete           Bound    demo/data-semi-sync-mysql-1   standard                4m35s
pvc-94f1c312-50e3-41e1-94a8-a820be0abc08   1Gi        RWO            Delete           Bound    demo/data-semi-sync-mysql-2   standard                4m31s

$ kubectl get service -n demo
NAME               TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
NAME                      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
semi-sync-mysql           ClusterIP   10.96.121.252   <none>        3306/TCP                     10m
semi-sync-mysql-pods      ClusterIP   None            <none>        3306/TCP,2380/TCP,2379/TCP   10m
semi-sync-mysql-standby   ClusterIP   10.96.133.61    <none>        3306/TCP                     10m

```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully provisioned. Run the following command to see the modified `MySQL` object:

```bash
$ kubectl get mysql -n demo 
NAME              VERSION   STATUS   AGE
semi-sync-mysql   9.1.0    Ready    16m
```

```yaml
$ kubectl get  my -n demo semi-sync-mysql -o yaml | kubectl neat
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: semi-sync-mysql
  namespace: demo
spec:
  allowedReadReplicas:
    namespaces:
      from: Same
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: semi-sync-mysql-auth
  autoOps: {}
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  podTemplate:
    ...
  
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
  topology:
    mode: SemiSync
    semiSync:
      errantTransactionRecoveryPolicy: PseudoTransaction
      sourceTimeout: 23h0m0s
      sourceWaitForReplicaCount: 1
  useAddressType: DNS
  version: 9.1.0
status:
  conditions:
    ...
  observedGeneration: 2
  phase: Ready
```

## Connect with MySQL database

KubeDB operator has created a new Secret called `semi-sync-mysql-auth` **(format: {mysql-object-name}-auth)** for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **username** for MySQL superuser and a `password` key which contains the **password** for MySQL superuser.

If you want to use an existing secret please specify that when creating the MySQL object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, you can connect to this database from your terminal using the `mysql` user and password.

```bash
$ kubectl get secrets -n demo semi-sync-mysql-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo semi-sync-mysql-auth -o jsonpath='{.data.\password}' | base64 -d
y~EC~984Et1Yfs~i
```

The operator creates a cluster according to the newly created `MySQL` object. This cluster has 3 members (one primary and two secondary).

You can connect to any of these cluster members. In that case you just need to specify the host name of that member Pod (either PodIP or the fully-qualified-domain-name for that Pod using the governing service named `<mysql-object-name>-pods`) by `--host` flag.

```bash
# first list the mysql pods list
$ kubectl get pods -n demo -l app.kubernetes.io/instance=semi-sync-mysql
NAME                READY   STATUS    RESTARTS   AGE
semi-sync-mysql-0   2/2     Running   0          21m
semi-sync-mysql-1   2/2     Running   0          20m
semi-sync-mysql-2   2/2     Running   0          20m

# get the governing service
$ kubectl get service semi-sync-mysql-pods -n demo
NAME                   TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                      AGE
semi-sync-mysql-pods   ClusterIP   None         <none>        3306/TCP,2380/TCP,2379/TCP   21m

# list the pods with PodIP
$ kubectl get pods -n demo -l app.kubernetes.io/instance=semi-sync-mysql -o jsonpath='{range.items[*]}{.metadata.name} ........... {.status.podIP} ............ {.metadata.name}.semi-sync-mysql-pods.{.metadata.namespace}{"\\n"}{end}'
semi-sync-mysql-0 ........... 10.244.0.18 ............ semi-sync-mysql-0.semi-sync-mysql-pods.demo
semi-sync-mysql-1 ........... 10.244.0.20 ............ semi-sync-mysql-1.semi-sync-mysql-pods.demo
semi-sync-mysql-2 ........... 10.244.0.22 ............ semi-sync-mysql-2.semi-sync-mysql-pods.demo
```

Now you can connect to these database using the above info. Ignore the warning message. It is happening for using password in the command.

```bash
# connect to the 1st server
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 2nd server
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i'  --host=semi-sync-mysql-1.semi-sync-mysql-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 3rd server
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i'  --host=semi-sync-mysql-2.semi-sync-mysql-pods.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+
```

## Check the Semi-sync cluster Status

Now, you are ready to check newly created semi-sync. Connect and run the following commands from any of the hosts and you will get the same results.

```bash

$ kubectl get pods -n demo --show-labels
NAME                READY   STATUS    RESTARTS   AGE    LABELS
semi-sync-mysql-0   2/2     Running   0          171m   app.kubernetes.io/component=database,app.kubernetes.io/instance=semi-sync-mysql,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,controller-revision-hash=semi-sync-mysql-77775485f8,kubedb.com/role=primary,petset.kubernetes.io/pod-name=semi-sync-mysql-0
semi-sync-mysql-1   2/2     Running   0          170m   app.kubernetes.io/component=database,app.kubernetes.io/instance=semi-sync-mysql,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,controller-revision-hash=semi-sync-mysql-77775485f8,kubedb.com/role=standby,petset.kubernetes.io/pod-name=semi-sync-mysql-1
semi-sync-mysql-2   2/2     Running   0          169m   app.kubernetes.io/component=database,app.kubernetes.io/instance=semi-sync-mysql,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,controller-revision-hash=semi-sync-mysql-77775485f8,kubedb.com/role=standby,petset.kubernetes.io/pod-name=semi-sync-mysql-2

From the labels we can see that the `semi-sync-mysql-0` is running as primary and the rest are running as standby.Lets validate with the mysql semisync status

$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "show  status like 'Rpl%_status';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+-----------------------------+-------+
| Variable_name               | Value |
+-----------------------------+-------+
| Rpl_semi_sync_master_status | ON    |
| Rpl_semi_sync_slave_status  | OFF   |
+-----------------------------+-------+


$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-1.semi-sync-mysql-pods.demo -e "show  status like 'Rpl%_status';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+-----------------------------+-------+
| Variable_name               | Value |
+-----------------------------+-------+
| Rpl_semi_sync_master_status | OFF   |
| Rpl_semi_sync_slave_status  | ON    |
+-----------------------------+-------+


$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-2.semi-sync-mysql-pods.demo -e "show  status like 'Rpl%_status';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+-----------------------------+-------+
| Variable_name               | Value |
+-----------------------------+-------+
| Rpl_semi_sync_master_status | OFF   |
| Rpl_semi_sync_slave_status  | ON    |
+-----------------------------+-------+

```

## Data Availability

In a MySQL semi-sync cluster, only the primary member can write not the secondary. But you can read data from any member. In this tutorial, we will insert data from primary, and we will see whether we can get the data from any other member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# create a database on primary
$ kubectl exec -it -n demo semi-sync-mysql-0 -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo semi-sync-mysql-0 -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$  kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from primary
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```
In the previous step we have inserted into the primary pod. In the next step we will read from secondary pods to determine whether the data has been successfully copied to the secondary pods.
```bash
# read from secondary-1
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i'  --host=semi-sync-mysql-1.semi-sync-mysql-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from secondary-2
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i'  --host=semi-sync-mysql-2.semi-sync-mysql-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Write on Secondary Should Fail

Only, primary member preserves the write permission. No secondary can write data.

```bash
# try to write on secondary-1
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i'  --host=semi-sync-mysql-1.semi-sync-mysql-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MySQL server is running with the --super-read-only option so it cannot execute this statement
command terminated with exit code 1

# try to write on secondary-2
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i'  --host=semi-sync-mysql-2.semi-sync-mysql-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('mango', 5, 'yellow');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MySQL server is running with the --super-read-only option so it cannot execute this statement
command terminated with exit code 1
```

## Automatic Failover

To test automatic failover, we will force the primary Pod to restart. Since the primary member (`Pod`) becomes unavailable, the rest of the members will elect a new primary for the cluster. When the old primary comes back, it will join the cluster as a secondary member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# delete the primary Pod semi-sync-mysql-0
$ kubectl delete pod semi-sync-mysql-0 -n demo
pod "semi-sync-mysql-0" deleted

# check the new primary ID
$ kubectl get pod -n demo --show-labels | grep primary
semi-sync-mysql-1   2/2     Running   0          3h9m   app.kubernetes.io/component=database,app.kubernetes.io/instance=semi-sync-mysql,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,controller-revision-hash=semi-sync-mysql-77775485f8,kubedb.com/role=primary,petset.kubernetes.io/pod-name=semi-sync-mysql-1

# now check the cluster status
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "show  status like 'Rpl%_status';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+-----------------------------+-------+
| Variable_name               | Value |
+-----------------------------+-------+
| Rpl_semi_sync_master_status | OFF   |
| Rpl_semi_sync_slave_status  | ON    |
+-----------------------------+-------+
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-1.semi-sync-mysql-pods.demo -e "show  status like 'Rpl%_status';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+-----------------------------+-------+
| Variable_name               | Value |
+-----------------------------+-------+
| Rpl_semi_sync_master_status | ON    |
| Rpl_semi_sync_slave_status  | OFF   |
+-----------------------------+-------+

$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-2.semi-sync-mysql-pods.demo -e "show  status like 'Rpl%_status';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+-----------------------------+-------+
| Variable_name               | Value |
+-----------------------------+-------+
| Rpl_semi_sync_master_status | OFF   |
| Rpl_semi_sync_slave_status  | ON    |


# read data from new primary semi-sync-mysql-1.semi-sync-mysql-pods.demo
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-1.semi-sync-mysql-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```
Now Let's read the data from secondary pods to see if the data is consistant.
```bash
# read data from secondary-1 semi-sync-mysql-0.semi-sync-mysql-pods.demo
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-0.semi-sync-mysql-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

# read data from secondary-2 semi-sync-mysql-2.semi-sync-mysql-pods.demo
$ kubectl exec -it -n demo semi-sync-mysql-0 -c mysql -- mysql -u root --password='y~EC~984Et1Yfs~i' --host=semi-sync-mysql-2.semi-sync-mysql-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Cleaning up

Clean what you created in this tutorial.

```bash
kubectl delete -n demo my/semi-sync-mysql
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLDBVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
