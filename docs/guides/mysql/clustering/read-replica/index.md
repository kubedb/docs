---
title: MySQL Read Replica Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-clustering-read-replica
    name: MySQL Read Replica Guide
    parent: guides-mysql-clustering
    weight: 21
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MySQL Read Replica

This tutorial will show you how to use KubeDB to provision a MySQL Read Replica from a kubedb managed mysql instance.

## Before You Begin

Before proceeding:

- Read [mysql replication concept](/docs/guides/mysql/clustering/overview/index.md) to learn about MySQL Replication.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/clustering/read-replica/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/clustering/group-replication/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).
## Read Replica

Read Replica allows us to replicate data from one mysql source to a read-only mysql server. In this section we will provision a mysql server with kubedb, and then we will create a read replica from it.


## Deploy Mysql server

The following is an example `MySQL` object which creates a MySQL standalone instance

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql
  namespace: demo
spec:
  allowedReadReplicas:
      namespaces:
        from: Same
      selector:
        matchLabels:
          kubedb.com/instance_name: ReadReplica  
  version: "8.0.29"
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/read-replica/yamls/mysql.yaml
mysql.kubedb.com/mysql created
```
Here,

 - `spec.AllowReadReplicas`  defines the types of read replicas that may be attached to a MySQL instance and the trusted namespaces where those Read Replica resources may be present.You will be able to set namespace `spec.allowReadReplicas.NameSpace` and labels `spec.allowReadReplicas.selector`.For more see [here](https://github.com/kubedb/apimachinery/blob/master/apis/kubedb/v1alpha2/mysql_types.go#L159).
 - `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete the operation of MySQL CR. *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy).

Now a MySQL instance in `demo` namespace having the label `kubedb.com/instance_name: ReadReplica` will be able to connect to this database as a read replica

KubeDB operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MySQL object name. KubeDB operator will also create a governing service for the StatefulSet with the name `<mysql-object-name>-pods`.

```bash
$ kubectl get statefulset -n demo
NAME    READY   AGE
mysql   1/1     18s


$ kubectl get pvc -n demo
NAME           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mysql-0   Bound    pvc-02b9688a-8dbb-4507-9020-8313c65f2943   1Gi        RWO            standard       41s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS   REASON   AGE
pvc-02b9688a-8dbb-4507-9020-8313c65f2943   1Gi        RWO            Delete           Bound    demo/data-mysql-0   standard                57s

$ kubectl get service -n demo
NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
mysql        ClusterIP   10.96.50.158   <none>        3306/TCP   76s
mysql-pods   ClusterIP   None           <none>        3306/TCP   76s

```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created

```bash
$ kubectl get mysql -n demo
NAME    VERSION   STATUS   AGE
mysql   8.0.31    Ready    97s
```

## Connect with MySQL database

KubeDB operator has created a new Secret called `my-group-auth` **(format: {mysql-object-name}-auth)** for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **username** for MySQL superuser and a `password` key which contains the **password** for MySQL superuser.

If you want to use an existing secret please specify that when creating the MySQL object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, you can connect to this database from your terminal using the `mysql` user and password.

```bash
$ kubectl get secrets -n demo mysql-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-auth -o jsonpath='{.data.\password}' | base64 -d
4~Dt~hvKR.(m*gZU 
```

The operator creates a standalone mysql server for the newly created `MySQL` object.

Now you can connect to the database using the above info. Ignore the warning message. It is happening for using password in the command.


##  Data Insertion

Let's insert some data to the newly created mysql server . we can use the primary service or governing service to connect with the database  
> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# create a database on primary
$ kubectl exec -it -n demo mysql-0 -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-0.mysql-pods.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo mysql-0 -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-0.mysql-pods.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$  kubectl exec -it -n demo mysql-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-0.mysql-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from primary
$ kubectl exec -it -n demo mysql-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-0.mysql-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```


#  Create  Read Replica 
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-read
  namespace: demo
  labels:
    kubedb.com/instance_name: ReadReplica
spec:
  version: "8.0.31"
  topology:
    mode: ReadReplica
    readReplica:
      sourceRef:
        name: mysql
        namespace: demo
  replicas: 2
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
Here,

- `spec.topology` contains the information about the mysql server.
- `spec.topology.mode` we are defining the server will be working a `read replica`.
- `spec.topology.readReplica.sourceref` we are referring to source to read. The  mysql instance we previously created.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete the operation of MySQL CR. *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy).
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/read-replica/read-replica.yaml
mysql.kubedb.com/mysql-read created
```

Now we will be able to see kubedb will provision a Read Replica from the source mysql instance. Lets checkout out the statefulSet , pvc , pv and services associated with it
.
```bash
$ kubectl get statefulset -n demo
NAME         READY   AGE
mysql        1/1     8m39s
mysql-read   2/2     30s

$ kubectl get pvc -n demo
NAME                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mysql-0        Bound    pvc-02b9688a-8dbb-4507-9020-8313c65f2943   1Gi        RWO            standard       9m8s
data-mysql-read-0   Bound    pvc-66b576c5-dac5-4f42-b871-ce852e3098aa   1Gi        RWO            standard       41s
data-mysql-read-1   Bound    pvc-4363d8e3-4999-485a-bd46-226db4373d27   1Gi        RWO            standard       34s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS   REASON   AGE
pvc-02b9688a-8dbb-4507-9020-8313c65f2943   1Gi        RWO            Delete           Bound    demo/data-mysql-0        standard                9m59s
pvc-4363d8e3-4999-485a-bd46-226db4373d27   1Gi        RWO            Delete           Bound    demo/data-mysql-read-1   standard                85s
pvc-66b576c5-dac5-4f42-b871-ce852e3098aa   1Gi        RWO            Delete           Bound    demo/data-mysql-read-0   standard

$ kubectl get service -n demo
mysql             ClusterIP   10.96.50.158    <none>        3306/TCP   11m
mysql-pods        ClusterIP   None            <none>        3306/TCP   11m
mysql-read        ClusterIP   10.96.151.145   <none>        3306/TCP   2m49s
mysql-read-pods   ClusterIP   None            <none>        3306/TCP   2m49s

```
KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified `MySQL` object:
```bash
$ kubectl get mysql -n demo 
NAME         VERSION   STATUS   AGE
mysql        8.0.31    Ready    15m
mysql-read   8.0.31    Ready    7m17s
```

##  Validate Read Replica
Since both source and replica database are in the ready state. we can validate Read replica is working properly by checking the replication status 

```bash
$ kubectl exec -it -n demo mysql-read-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-read-0.mysql-read-pods.demo -e "show slave status\G" 
mysql: [Warning] Using a password on the command line interface can be insecure.
*************************** 1. row ***************************
               Slave_IO_State: Waiting for source to send event
                  Master_Host: mysql.demo.svc
                  Master_User: root
                  Master_Port: 3306
                Connect_Retry: 60
              Master_Log_File: binlog.000002
          Read_Master_Log_Pos: 214789
               Relay_Log_File: mysql-read-0-relay-bin.000002
                Relay_Log_Pos: 186366
        Relay_Master_Log_File: binlog.000002
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
            ....

$ kubectl exec -it -n demo mysql-read-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-read-1.mysql-read-pods.demo -e "show slave status\G"
mysql: [Warning] Using a password on the command line interface can be insecure.
*************************** 1. row ***************************
               Slave_IO_State: Waiting for source to send event
                  Master_Host: mysql.demo.svc
                  Master_User: root
                  Master_Port: 3306
                Connect_Retry: 60
              Master_Log_File: binlog.000002
          Read_Master_Log_Pos: 230420
               Relay_Log_File: mysql-read-1-relay-bin.000003
                Relay_Log_Pos: 230630
        Relay_Master_Log_File: binlog.000002
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
           
```

# Read Data
In the previous step we have inserted into the primary pod. In the next step we will read from secondary pods to determine whether the data has been successfully copied to the secondary pods.
```bash
# read from secondary-1
$ kubectl exec -it -n demo mysql-read-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU'  --host=mysql-read-0.mysql-read-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from secondary-2
$ kubectl exec -it -n demo mysql-read-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU'  --host=mysql-read-1.mysql-read-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Write on Secondary Should Fail

Only, primary member preserves the write permission. No secondary can write data.

## Automatic Failover

To test automatic failover, we will force the primary Pod to restart. Since the primary member (`Pod`) becomes unavailable, the rest of the members will elect a new primary for these group. When the old primary comes back, it will join the group as a secondary member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# delete the primary Pod mysql-read-0
$ kubectl delete pod mysql-read-0 -n demo
pod "mysql-read-0" deleted

# check the new primary ID
$ kubectl exec -it -n demo mysql-read-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-read-0.mysql-read-pods.demo -e "show slave status\G" 
mysql: [Warning] Using a password on the command line interface can be insecure.
*************************** 1. row ***************************
               Slave_IO_State: Waiting for source to send event
                  Master_Host: mysql.demo.svc
                  Master_User: root
                  Master_Port: 3306
                Connect_Retry: 60
              Master_Log_File: binlog.000002
          Read_Master_Log_Pos: 214789
               Relay_Log_File: mysql-read-0-relay-bin.000002
                Relay_Log_Pos: 186366
        Relay_Master_Log_File: binlog.000002
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
        ...

# read data after recovery
$ kubectl exec -it -n demo mysql-read-0 -c mysql -- mysql -u root --password='4~Dt~hvKR.(m*gZU' --host=mysql-read-2.mysql-read-pods.demo -e "SELECT * FROM playground.equipment;"
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
kubectl delete -n demo my/mysql
kubectl delete -n dem my/mysql-read
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLDBVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
