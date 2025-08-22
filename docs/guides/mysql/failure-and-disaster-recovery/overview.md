---
title: MySQL Failover and DR Scenarios
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-failure-and-disaster-recovery-overview
    name: Guide
    parent: guides-mysql-failure-and-disaster-recovery
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Maximizing MySQL Uptime and Reliability

## A Guide to KubeDB's High Availability and Auto-Failover

In today’s always-on digital landscape, database downtime can lead to significant business disruption,
data loss, or degraded user experience. Ensuring continuous database availability is especially important
for mission-critical applications running in dynamic environments like Kubernetes. KubeDB addresses this 
need by offering built-in support for high availability and automated failover for MySQL. It continuously
monitors the health of database nodes and automatically detects failures. When a primary node becomes 
unavailable, the system quickly promotes a healthy replica to maintain service continuity — all without 
manual intervention. This seamless failover mechanism ensures that your MySQL workloads remain highly 
available, resilient, and ready to scale in production environments.
This article will guide you through KubeDB's automated failover capabilities for MySQL. We will set up
an HA cluster and then simulate a leader failure to see KubeDB's auto-recovery mechanism in action.

> You will see how fast the failover happens when it's truly necessary. Failover in KubeDB-managed MySQL 
will generally happen within 10 seconds depending on your cluster networking. There is an exception 
scenario that we discussed later in this doc where failover might take a bit longer up to 45 seconds. 
But that is a bit rare though.

### Before You Start

To follow along with this tutorial, you will need:

1. A running Kubernetes cluster.
2. KubeDB [installed](https://kubedb.com/docs/{{< param "info.version" >}}/setup/install/kubedb/) in your cluster.
3. kubectl command-line tool configured to communicate with your cluster.

### Step 1: Create a High-Availability MySQL Cluster

To begin, we’ll deploy a MySQL cluster configured for High Availability (HA).
Unlike a standalone MySQL instance, an HA cluster includes:
- A primary pod that handles all write operations, and
- One or more standby pods that are ready to take over automatically if the primary node fails.

The following YAML manifest defines a 3-node MySQL cluster with streaming replication enabled.
Save it as ha-mysql.yaml:
```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: ha-mysql
  namespace: demo
spec:
  version: "8.2.0"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut

```

Now, create the namespace and apply the manifest:

```shell
# Create the namespace if it doesn't exist
$ kubectl create ns demo
namespace/demo created

# Apply the manifest to deploy the cluster
$ kubectl apply -f ha-mysql.yaml
mysql.kubedb.com/ha-mysql created
```

You can monitor on another terminal the status until all pods are ready:
```shell
$ watch kubectl get my,petset,pods -n demo
```
See the database is ready.

```shell
$ kubectl get my,petset,pods -n demo
NAME                             VERSION   STATUS   AGE
mysql.kubedb.com/ha-mysql   8.2.0     Ready    19h

NAME                                         AGE
petset.apps.k8s.appscode.com/ha-mysql   19h

NAME                  READY   STATUS    RESTARTS      AGE
pod/ha-mysql-0   2/2     Running   3 (24m ago)   16h
pod/ha-mysql-1   2/2     Running   2 (24m ago)   16h
pod/ha-mysql-2   2/2     Running   3 (24m ago)   16h

```

Inspect who is primary and who is standby.

```shell
# you can inspect who is primary
# and who is secondary like below

$ kubectl get pods -n demo --show-labels | grep role
ha-mysql-0   2/2     Running   0          34m   app.kubernetes.io/component=database,app.kubernetes.io/instance=ha-mysql,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=ha-mysql-7f595bb48b,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=ha-mysql-0
ha-mysql-1   2/2     Running   0          34m   app.kubernetes.io/component=database,app.kubernetes.io/instance=ha-mysql,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=ha-mysql-7f595bb48b,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=ha-mysql-1
ha-mysql-2   2/2     Running   0          34m   app.kubernetes.io/component=database,app.kubernetes.io/instance=ha-mysql,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=ha-mysql-7f595bb48b,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=ha-mysql-2

```
The pod having `kubedb.com/role=primary` is the primary and `kubedb.com/role=standby` are the standby's.


Lets create a table in the primary.

```shell
# find the primary pod
$ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1 }'
ha-mysql-0

# exec into the primary pod
$ kubectl exec -it -n demo ha-mysql-0  -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-4.4$  mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 742
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create database hello;
Query OK, 1 row affected (0.06 sec)

mysql> show Databases;
+--------------------+
| Database           |
+--------------------+
| hello              |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.09 sec)

```

Verify that the table has been created on the standby nodes. Note that standby pods have read-only access, 
so you won't be able to perform any write operations.

```shell
$ kubectl exec -it -n demo ha-mysql-1  -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 564
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| hello              |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.47 sec)

mysql> create database Hi;
ERROR 1290 (HY000): The MySQL server is running with the --super-read-only option so it cannot execute this statement

```
### Step 2: Simulating a Failover

Before simulating failover, let’s discuss how KubeDB-managed MySQL handles such scenarios.
KubeDB deploys a high availability MySQL cluster that continuously monitors the health and 
status of each database node. When a failure is detected — such as a primary node becoming 
unreachable — the system automatically reconfigures the cluster to promote a healthy member 
to take over as the new primary. This ensures continued availability of write operations with 
minimal disruption. The failover process is fast, coordinated, and typically completes within 
a few seconds, allowing your MySQL database to remain highly available and resilient in dynamic 
Kubernetes environments.

Now current running primary is `ha-mysql-0`. Let's open another terminal and run the command below.


```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
It will show current mysql cluster roles like that:

```shell
ha-mysql-0 primary
ha-mysql-1 standby
ha-mysql-2 standby
```

#### Case 1: Delete the current primary

Lets delete the current primary and see how the role change happens almost immediately.

```shell
$ kubectl delete pods -n demo ha-mysql-0 
pod "ha-mysql-0" deleted
```
You see almost immediately the failover happened. 
```shell
ha-mysql-0 
ha-mysql-1 primary
ha-mysql-2 standby
```

Here's what happened internally is mainly managed by
MySQL group Replication like this steps:

**1.Cluster Formation**

A group of MySQL servers is initialized, each maintaining a full copy of the data. They form a replication group that communicates via a reliable, ordered messaging system powered by `Paxos-based` consensus.

**2.Group Membership & Coordination**

Each server joins a consistent, updated view of the group using a group membership service. This ensures that every member knows which nodes are active and can participate in decision-making.

**3.Transaction Agreement**

When a read/write transaction is submitted, the originating server proposes it to the group. All members must certify and agree on the order and validity of the transaction before it can commit. This decision is coordinated using atomic message delivery and total ordering guarantees.

**4.Conflict Detection & Certification**

Before committing, each server checks for possible conflicts with other concurrent transactions by comparing write sets. If no conflict is found, the transaction is approved and applied in the agreed global order.

**5.Commit and Replication**

Once approved, the transaction is committed on the originating server and sent to all others. Each server applies the transaction, ensuring all members eventually reach the same consistent state.

**6.Failure Detection**

If the primary server fails, the group detects the failure automatically using built-in heartbeat and monitoring mechanisms within the Paxos-based group communication engine.

**7.Automatic Reconfiguration**

The group reconfigures itself to exclude the failed node and triggers a new primary election from among the remaining healthy servers.

**8.New Primary Assignment**

A healthy replica is promoted as the new primary, and it resumes accepting writes. The group continues processing transactions without manual intervention.

Now we know how failover is done, let's check if the new primary is working.

```shell
$ kubectl exec -it -n demo ha-mysql-1  -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1870
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE hi;
Query OK, 1 row affected (0.16 sec)
```

You will see the deleted pod (ha-mysql-0) is brought back by the kubedb operator and it is now assigned to standby role.

```shell
ha-mysql-0 standby
ha-mysql-1 primary
ha-mysql-2 standby
```

Lets check if the standby(`ha-mysql-0`) got the updated data from new primary `ha-mysql-1`.

```shell
$ kubectl exec -it -n demo ha-mysql-1  -- bash

Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1909
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> Show Databases;
+--------------------+
| Database           |
+--------------------+
| hello              |
| hi                 |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
7 rows in set (0.12 sec)

```

#### Case 2: Delete the current primary and One replica

```shell
$ kubectl delete pods -n demo ha-mysql-1 ha-mysql-2
pod "ha-mysql-1" deleted
pod "ha-mysql-2" deleted
```
Again we can see the failover happened pretty quickly.

```shell
ha-mysql-0 primary
ha-mysql-1 
ha-mysql-2
```

After 10-40 second, the deleted pods will be back and will have its role.

```shell
ha-mysql-0 primary
ha-mysql-1 standby
ha-mysql-2 standby
```
Lets validate the cluster state from new primary(`ha-mysql-0`).

```shell
$ kubectl exec -it -n demo ha-mysql-0  -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 974
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
+---------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                 | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+---------------------------------------------+-------------+--------------+-------------+
| ha-mysql-1.ha-mysql-pods.demo.svc |        3306 | ONLINE       | SECONDARY   |
| ha-mysql-0.ha-mysql-pods.demo.svc |        3306 | ONLINE       | PRIMARY     |
| ha-mysql-2.ha-mysql-pods.demo.svc |        3306 | ONLINE       | SECONDARY   |
+---------------------------------------------+-------------+--------------+-------------+
3 rows in set (0.00 sec)

```

#### Case3: Delete any of the replica's

Let's delete both of the standby's.

```shell
$ kubectl delete pods -n demo ha-mysql-1 ha-mysql-2
pod "ha-mysql-1" deleted
pod "ha-mysql-2" deleted

```

```shell
ha-mysql-0 primary
ha-mysql-1 
ha-mysql-2
```

Shortly both of the pods will be back with its role.

```shell
ha-mysql-0 primary
ha-mysql-1 standby
ha-mysql-2 standby
```

Lets verify cluster state.
```shell
$ kubectl exec -it -n demo ha-mysql-0  -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 974
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
+---------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                 | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+---------------------------------------------+-------------+--------------+-------------+
| ha-mysql-1.ha-mysql-pods.demo.svc |        3306 | ONLINE       | SECONDARY   |
| ha-mysql-0.ha-mysql-pods.demo.svc |        3306 | ONLINE       | PRIMARY     |
| ha-mysql-2.ha-mysql-pods.demo.svc |        3306 | ONLINE       | SECONDARY   |
+---------------------------------------------+-------------+--------------+-------------+
3 rows in set (0.01 sec)
```

#### Case 4: Delete both primary and all replicas

Let's delete all the pods.

```shell
$ kubectl delete pods -n demo ha-mysql-0 ha-mysql-1 ha-mysql-2
pod "ha-mysql-0" deleted
pod "ha-mysql-1" deleted
pod "ha-mysql-2" deleted
```
```bash
ha-mysql-0 
ha-mysql-1
ha-mysql-2
```

Within 20-30 second, all of the pod should be back.

```shell
ha-mysql-0 primary
ha-mysql-1 standby
ha-mysql-2 standby
```

Lets verify the cluster state now.

```shell
$ kubectl exec -it -n demo ha-mysql-0  -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 136
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT MEMBER_HOST, MEMBER_PORT, MEMBER_STATE, MEMBER_ROLE FROM performance_schema.replication_group_members;
+---------------------------------------------+-------------+--------------+-------------+
| MEMBER_HOST                                 | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE |
+---------------------------------------------+-------------+--------------+-------------+
| ha-mysql-1.ha-mysql-pods.demo.svc |        3306 | ONLINE       | SECONDARY   |
| ha-mysql-0.ha-mysql-pods.demo.svc |        3306 | ONLINE       | PRIMARY     |
| ha-mysql-2.ha-mysql-pods.demo.svc |        3306 | ONLINE       | SECONDARY   |
+---------------------------------------------+-------------+--------------+-------------+
3 rows in set (0.00 sec)
```

## A Guide to Handling MySQL Storage

It is often possible that your database storage become full and your database has stopped working. We have got you covered. You just apply a VolumeExpansion `mysqlOpsRequest` and your database storage will be increased, and the database will be ready to use again.

### Disaster Scenario and Recovery


#### Scenario
You deploy a `MySQL` database. The database was running fine. Someday, your database storage becomes full. As your mysql 
process can't write to the filesystem,
clients won't be able to connect to the database. Your database status will be `Not Ready`.

#### Recovery

In order to recover from this, you can create a `VolumeExpansion` `mysqlOpsRequest` with expanded resource requests.

As soon as you create this, KubeDB will trigger the necessary steps to expand your volume based on your specifications 
on the `mysqlOpsRequest` manifest. A sample `mysqlOpsRequest` manifest for `VolumeExpansion` is given below:


```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: my-online-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: ha-mysql
  volumeExpansion:
    mode: "Offline"
    mysql: 2Gi
```

For more details, please check the full section [here](/docs/guides/mysql/volume-expansion/overview/index.md).


> **Note**: There are two ways to update your volume: 1.Online 2.Offline. Which Mode to choose? <br>
It depends on your `StorageClass`. If your storageclass supports online volume expansion, you can go with it. Otherwise, you can go with `Offline` Volume Expansion.


## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:
```shell
$ kubectl delete my -n demo ha-mysql
$ kubectl delete ns demo
```

### Next Steps

- You can configure [Backup and Restore](/docs/guides/mysql/backup/kubestash/overview/index.md)
- Learn about [PITR](/docs/guides/mysql/pitr/volumesnapshot/archiver.md)
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLDBVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
