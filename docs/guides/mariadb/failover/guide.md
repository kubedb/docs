---
title: Postgres Failover and DR Scenarios
menu:
  docs_{{ .version }}:
    identifier: mariadb-failover
    name: Overview
    parent: mariadb-failure-disaster-recovery
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

#  Exploring Fault Tolerance in MariaDB with KubeDB

## Understanding Failover and Clustering in MariaDB on KubeDB
`Failover` refers to the process of automatically switching to a standby system or replica when the primary
database node fails. In high-availability database systems, failover ensures that services remain
uninterrupted even when one or more nodes go down. This capability is critical in modern, cloud-native 
infrastructure where downtime can lead to major disruptions.

When running `MariaDB` on Kubernetes using KubeDB, failover becomes more seamless. KubeDB supports two types of MariaDB clustering strategies:

- **Standard Replication (Primary-Replica):**
This is a traditional setup where one pod acts as the primary (read/write) node, and others are replicas
(read-only). If the primary fails, KubeDB detects it and automatically promotes a healthy replica to become 
the new primary. This mode supports automatic failover.

- **Galera Cluster (Multi-Primary):**
In this setup, all nodes act as primary, capable of handling both read and write operations. Since there’s
no single point of failure, the system provides synchronous replication and built-in high availability, 
but doesn’t use the traditional failover concept, as all pods are equal.

In the rest of this blog, we'll focus on how failover works in the standard replication mode, and how
KubeDB handles recovery in the event of a node failure.


## Before You Begin

Before proceeding:

- Read [mariadb galera cluster concept](/docs/guides/mariadb/clustering/overview) to learn about MariaDB
Galera Cluster.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to 
communicate with your cluster. If you do not already have a cluster, you can create one by 
using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```
## Deploy MariaDB Cluster

The following is an example `MariaDB` object which creates a single-master MariaDB standard replication cluster with three members.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.6.16"
  replicas: 3
  topology:
    mode: MariaDBReplication
    maxscale:
      replicas: 3
      enableUI: true
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 50Mi
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/clustering/galera-cluster/examples/demo-1.yaml
mariadb.kubedb.com/sample-mariadb created
```

Here,

- `spec.replicas` Defines the number of MariaDB pods (instances) in the cluster.
- `spec.storage` Specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.topology` Configures the database topology and associated components.
- `spec.topology.maxscale` Specifies the Maxscale proxy server configuration.
- `spec.topology.maxscale.replicas` Defines the number of MaxScale replicas in the petset managed by the KubeDB Operator.
- `spec.topology.maxscale.enableUI` A boolean parameter (e.g. true or false) that controls whether the MaxScale GUI (accessible via the REST API) is enabled for the MaxScale instance.

KubeDB operator watches for `MariaDB` objects using Kubernetes API. When a `MariaDB` object is created, KubeDB operator will create a new PetSet and a Service with the matching MariaDB object name. KubeDB operator will also create a governing service for the PetSet with the name `<mariadb-object-name>-pods`. 

You can monitor the status until all pods are ready:
```shell
watch kubectl get mariadb,petset,pods -n demo
```
See the database is ready.

```shell
$ kubectl get mariadb,petset,pods -n demo
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/sample-mariadb   10.6.16   Ready    3m27s

NAME                                             AGE
petset.apps.k8s.appscode.com/sample-mariadb      3m20s
petset.apps.k8s.appscode.com/sample-mariadb-mx   3m23s

NAME                      READY   STATUS    RESTARTS   AGE
pod/sample-mariadb-0      2/2     Running   0          3m20s
pod/sample-mariadb-1      2/2     Running   0          3m20s
pod/sample-mariadb-2      2/2     Running   0          3m20s
pod/sample-mariadb-mx-0   1/1     Running   0          3m23s
pod/sample-mariadb-mx-1   1/1     Running   0          3m23s
pod/sample-mariadb-mx-2   1/1     Running   0          3m23s

```

Inspect who is primary and who is standby.

```shell
# you can inspect the role of the pods 

$ kubectl get pods -n demo --show-labels | grep role
sample-mariadb-0      2/2     Running   0          4m9s    app.kubernetes.io/component=database,app.kubernetes.io/instance=sample-mariadb,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mariadbs.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=sample-mariadb-598cd56869,kubedb.com/role=Master,statefulset.kubernetes.io/pod-name=sample-mariadb-0
sample-mariadb-1      2/2     Running   0          4m9s    app.kubernetes.io/component=database,app.kubernetes.io/instance=sample-mariadb,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mariadbs.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=sample-mariadb-598cd56869,kubedb.com/role=Slave,statefulset.kubernetes.io/pod-name=sample-mariadb-1
sample-mariadb-2      2/2     Running   0          4m9s    app.kubernetes.io/component=database,app.kubernetes.io/instance=sample-mariadb,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mariadbs.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=sample-mariadb-598cd56869,kubedb.com/role=Slave,statefulset.kubernetes.io/pod-name=sample-mariadb-2

```
The pod having `kubedb.com/role=Master` is the primary and `kubedb.com/role=Slave` are the standby’s.
You can also check it on the cluster status:
```shell
 kubectl exec -it -n demo svc/sample-mariadb-mx -- bash
Defaulted container "maxscale" out of: maxscale, maxscale-init (init)
bash-4.4$ maxctrl list servers
┌─────────┬─────────────────────────────────────────────────────────────┬──────┬─────────────┬─────────────────┬─────────┬────────────────────┐
│ Server  │ Address                                                     │ Port │ Connections │ State           │ GTID    │ Monitor            │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼─────────┼────────────────────┤
│ server1 │ sample-mariadb-0.sample-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Master, Running │ 0-1-217 │ ReplicationMonitor │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼─────────┼────────────────────┤
│ server2 │ sample-mariadb-1.sample-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Slave, Running  │ 0-1-217 │ ReplicationMonitor │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼─────────┼────────────────────┤
│ server3 │ sample-mariadb-2.sample-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Slave, Running  │ 0-1-217 │ ReplicationMonitor │
└─────────┴─────────────────────────────────────────────────────────────┴──────┴─────────────┴─────────────────┴─────────┴────────────────────┘

```
Let’s make sure all pods are reachable and running as expected.
```shell

$ kubectl exec -it -n demo sample-mariadb-0  -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@sample-mariadb-0:/$  mysql -u$ {MYSQL_ROOT_USERNAME} -p$ {MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 413
Server version: 10.5.23-MariaDB-1:10.5.23+maria~ubu2004 mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE DB_1;
Query OK, 1 row affected (0.018 sec)

MariaDB [(none)]> Show Databases;
+--------------------+
| Database           |
+--------------------+
| DB_1               |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
+--------------------+
5 rows in set (0.000 sec)

MariaDB [(none)]> exit
Bye

```
Here, a database is created which can be used in any other nodes. Let's see:

```shell
$ kubectl exec -it -n demo sample-mariadb-1  -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@sample-mariadb-1:/$  mysql -u$ {MYSQL_ROOT_USERNAME} -p$ {MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 744
Server version: 10.5.23-MariaDB-1:10.5.23+maria~ubu2004 mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE TABlE DB_1.tb1 ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.046 sec)

MariaDB [(none)]> Show Databases;
+--------------------+
| Database           |
+--------------------+
| DB_1               |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
+--------------------+
5 rows in set (0.000 sec)

MariaDB [(none)]> exit
Bye
mysql@sample-mariadb-1:/$  exit
exit

$ kubectl exec -it -n demo sample-mariadb-2  -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@sample-mariadb-2:/$   mysql -u$ {MYSQL_ROOT_USERNAME} -p$ {MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 883
Server version: 10.5.23-MariaDB-1:10.5.23+maria~ubu2004 mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> INSERT INTO DB_1.tb1 (type, quant, color) VALUES ('slide', 2, 'blue');
Query OK, 1 row affected (0.004 sec)

MariaDB [(none)]> Show * From DB_1.tb1;
ERROR 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MariaDB server version for the right syntax to use near '* From DB_1.tb1' at line 1
MariaDB [(none)]> Select * From DB_1.tb1;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.000 sec)

```
So, users can connect to any pod to access databases and perform read or write operations on any of them.
### Step 2: Simulating a Failover

Before simulating failover, let's discuss how we handle these failover scenarios in KubeDB-managed MariaDB.
We use sidecar container with all db pods, and inside that sidecar container, we use [raft](https://raft.github.io/)
protocol to detect the viable primary of the MariaDB cluster. Raft will choose a db pod as a leader of the MariaDB cluster, we will
check if that pod can really run as a leader. If everything is good with that chosen pod, we will run it as primary. This whole process of failover
generally takes less than 10 seconds to complete. So you can expect very rapid failover to ensure high availability of your MariaDB cluster.



Now current running primary is `mariadb-ha-demo-0`. Let's open another terminal and run the command below.


```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"

```
It will show current mariadb cluster roles.

![img.png](/docs/guides/postgres/failure-and-disaster-recovery/img.png)

#### Case 1: Delete the current primary

Lets delete the current primary and see how the role change happens almost immediately.

```shell
$ kubectl delete pods -n demo mariadb-ha-demo-0 
pod "mariadb-ha-demo-0" deleted

```

![img_1.png](/docs/guides/postgres/failure-and-disaster-recovery/img_1.png)

You see almost immediately the failover happened. Here's what happened internally:

- Distributed raft algorithm implementation is running 24 * 7 in your each db sidecar. You can configure this behavior as shown below.
- As soon as `mariadb-ha-demo-0` was being deleted and raft inside `mariadb-ha-demo-0` senses the termination, it immediately switches the leadership to any other viable leader before termination.
- In our case, raft inside `mariadb-ha-demo-1` got the leadership.
- Now this leader switch only means raft leader switch, not the **database leader switch(aka failover)** yet. So `mariadb-ha-demo-1` still running as replica. It will be primary after the next step.
- Once raft sidecar inside `mariadb-ha-demo-1` see it has become leader of the cluster, it initiates the database failover process and start running as primary.
- So, now `mariadb-ha-demo-1` is running as primary.

```yaml
# You can find this part in your db yaml by running
# kubectl get mariadb -n demo mariadb-ha-demo -oyaml
# under db.spec section
# vist below link for more information
# https://github.com/kubedb/apimachinery/blob/97c18a62d4e33a112e5f887dc3ad910edf3f3c82/apis/kubedb/v1/postgres_types.go#L204

leaderElection:
  electionTick: 10
  heartbeatTick: 1
  maximumLagBeforeFailover: 67108864
  period: 300ms
  transferLeadershipInterval: 1s
  transferLeadershipTimeout: 1m0s
  
```

Now we know how failover is done, let's check if the new primary is working.

```shell
$ kubectl exec -it -n demo mariadb-ha-demo-1  -- bash
Defaulted container "postgres" out of: postgres, mariadb-coordinator, postgres-init-container (init)
mariadb-ha-demo-1:/$  psql
psql (17.2)
Type "help" for help.

postgres=# create table hi(id int);
CREATE TABLE # See we were able to create the database. so failover was successful.
postgres=# 

```


You will see the deleted pod (mariadb-ha-demo-0) is brought back by the kubedb operator and it is now assigned to standby role.


![img_2.png](/docs/guides/postgres/failure-and-disaster-recovery/img_2.png)

Lets check if the standby(`mariadb-ha-demo-0`) got the updated data from new primary `mariadb-ha-demo-1`.

```shell
$ kubectl exec -it -n demo mariadb-ha-demo-0  -- bash
Defaulted container "postgres" out of: postgres, mariadb-coordinator, postgres-init-container (init)
mariadb-ha-demo-0:/$  psql
psql (17.2)
Type "help" for help.

postgres=# \dt
               List of relations
 Schema |        Name        | Type  |  Owner   
--------+--------------------+-------+----------
 public | hello              | table | postgres
 public | hi                 | table | postgres # this was created in the new primary
 public | kubedb_write_check | table | postgres
(3 rows)

```

#### Case 2: Delete the current primary and One replica

```shell
$ kubectl delete pods -n demo mariadb-ha-demo-1 mariadb-ha-demo-2
pod "mariadb-ha-demo-1" deleted
pod "mariadb-ha-demo-2" deleted
```
Again we can see the failover happened pretty quickly.

![img_3.png](/docs/guides/postgres/failure-and-disaster-recovery/img_3.png)

After 10-30 second, the deleted pods will be back and will have its role.

![img_4.png](/docs/guides/postgres/failure-and-disaster-recovery/img_4.png)

Lets validate the cluster state from new primary(`mariadb-ha-demo-0`).

```shell
$ kubectl exec -it -n demo mariadb-ha-demo-0  -- bash
Defaulted container "postgres" out of: postgres, mariadb-coordinator, postgres-init-container (init)
mariadb-ha-demo-0:/$  psql
psql (17.2)
Type "help" for help.

postgres=# select * from mariadb_stat_replication;
 pid  | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |    flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
------+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+-----------------+-----------------+---------------+------------+-------------------------------
 1098 |       10 | postgres | mariadb-ha-demo-1     | 10.42.0.191 |                 |       49410 | 2025-06-20 09:56:36.989448+00 |              | streaming | 0/70016A8 | 0/70016A8 | 0/70016A8 | 0/70016A8  | 00:00:00.000142 | 00:00:00.00066  | 00:00:00.000703 |             0 | async      | 2025-06-20 09:59:40.217223+00
 1129 |       10 | postgres | mariadb-ha-demo-2     | 10.42.0.192 |                 |       35216 | 2025-06-20 09:56:39.042789+00 |              | streaming | 0/70016A8 | 0/70016A8 | 0/70016A8 | 0/70016A8  | 00:00:00.000219 | 00:00:00.000745 | 00:00:00.00079  |             0 | async      | 2025-06-20 09:59:40.217308+00
(2 rows)

```

#### Case3: Delete any of the replica's

Let's delete both of the standby's.

```shell
kubectl delete pods -n demo mariadb-ha-demo-1 mariadb-ha-demo-2
pod "mariadb-ha-demo-1" deleted
pod "mariadb-ha-demo-2" deleted

```

![img_5.png](/docs/guides/postgres/failure-and-disaster-recovery/img_5.png)

Shortly both of the pods will be back with its role.

![img_6.png](/docs/guides/postgres/failure-and-disaster-recovery/img_6.png)

Lets verify cluster state.
```shell
$ kubectl exec -it -n demo mariadb-ha-demo-0  -- bash
Defaulted container "postgres" out of: postgres, mariadb-coordinator, postgres-init-container (init)
mariadb-ha-demo-0:/$  psql
psql (17.2)
Type "help" for help.

postgres=# select * from mariadb_stat_replication;
 pid  | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |    flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
------+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+-----------------+-----------------+---------------+------------+-------------------------------
 5564 |       10 | postgres | mariadb-ha-demo-2     | 10.42.0.194 |                 |       51560 | 2025-06-20 10:06:26.988807+00 |              | streaming | 0/7014A58 | 0/7014A58 | 0/7014A58 | 0/7014A58  | 00:00:00.000178 | 00:00:00.000811 | 00:00:00.000848 |             0 | async      | 2025-06-20 10:07:50.218299+00
 5572 |       10 | postgres | mariadb-ha-demo-1     | 10.42.0.193 |                 |       36158 | 2025-06-20 10:06:27.980841+00 |              | streaming | 0/7014A58 | 0/7014A58 | 0/7014A58 | 0/7014A58  | 00:00:00.000194 | 00:00:00.000818 | 00:00:00.000895 |             0 | async      | 2025-06-20 10:07:50.218337+00
(2 rows)


```

#### Case 4: Delete both primary and all replicas

Let's delete all the pods.

```shell
$ kubectl delete pods -n demo mariadb-ha-demo-0 mariadb-ha-demo-1 mariadb-ha-demo-2
pod "mariadb-ha-demo-0" deleted
pod "mariadb-ha-demo-1" deleted
pod "mariadb-ha-demo-2" deleted

```
![img_7.png](/docs/guides/postgres/failure-and-disaster-recovery/img_7.png)

Within 20-30 second, all of the pod should be back.

![img_8.png](/docs/guides/postgres/failure-and-disaster-recovery/img_8.png)

Lets verify the cluster state now.

```shell
$ kubectl exec -it -n demo mariadb-ha-demo-0  -- bash
Defaulted container "postgres" out of: postgres, mariadb-coordinator, postgres-init-container (init)
mariadb-ha-demo-0:/$  psql
psql (17.2)
Type "help" for help.

postgres=# select * from mariadb_stat_replication;
 pid | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |    flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
-----+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+-----------------+-----------------+---------------+------------+-------------------------------
 132 |       10 | postgres | mariadb-ha-demo-2     | 10.42.0.197 |                 |       34244 | 2025-06-20 10:09:20.27726+00  |              | streaming | 0/9001848 | 0/9001848 | 0/9001848 | 0/9001848  | 00:00:00.00021  | 00:00:00.000841 | 00:00:00.000894 |             0 | async      | 2025-06-20 10:11:02.527633+00
 133 |       10 | postgres | mariadb-ha-demo-1     | 10.42.0.196 |                 |       40102 | 2025-06-20 10:09:20.279987+00 |              | streaming | 0/9001848 | 0/9001848 | 0/9001848 | 0/9001848  | 00:00:00.000225 | 00:00:00.000848 | 00:00:00.000905 |             0 | async      | 2025-06-20 10:11:02.527653+00
(2 rows)


```

> **We make sure the pod with highest lsn (you can think lsn as the highest data point available in your cluster) always run as primary, so if a case occur where the pod with highest lsn is being terminated, we will not perform the failover until the highest lsn pod is back online. So in a case, where that highest lsn primary is not recoverable, read [this](https://appscode.com/blog/post/kubedb-v2025.2.19/#forcefailover) to do a force failover.**


## A Guide to Postgres Backup And Restore

You can configure Backup and Restore following the below documentation.

[Backup and Restore](/docs/guides/postgres/backup)

Youtube video Links: [link](https://www.youtube.com/watch?v=j9y5MsB-guQ)


## A Guide to Postgres PITR
Documentaion Link: [PITR](/docs/guides/postgres/pitr)

Concepts and Demo: [link](https://www.youtube.com/watch?v=gR5UdN6Y99c)

Basic Demo: [link](https://www.youtube.com/watch?v=BdMSVjNQtCA)

Full Demo: [link](https://www.youtube.com/watch?v=KAl3rdd8i6k)

## A Guide to Handling Postgres Storage

It is often possible that your database storage become full and your database has stopped working. We have got you covered. You just apply a VolumeExpansion `PostgresOpsRequest` and your database storage will be increased, and the database will be ready to use again.

### Disaster Scenario and Recovery

#### Scenario

You deploy a `MariaDB` database. The database was running fine. Someday, your database storage becomes full. As your postgres process can't write to the filesystem,
clients won't be able to connect to the database. Your database status will be `Not Ready`.

#### Recovery

In order to recover from this, you can create a `VolumeExpansion` `PostgresOpsRequest` with expanded resource requests.
As soon as you create this, KubeDB will trigger the necessary steps to expand your volume based on your specifications on the `PostgresOpsRequest` manifest. A sample `PostgresOpsRequest` manifest for `VolumeExpansion` is given below:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: mariadbops-vol-exp-ha-demo
  namespace: demo
spec:
  apply: Always
  databaseRef:
    name: mariadb-ha-demo
  type: VolumeExpansion
  volumeExpansion:
    mode: Online # see the notes, your storageclass must support this mode
    postgres: 20Gi # expanded resource
```


For more details, please check the full section [here](/docs/guides/postgres/volume-expansion/Overview/overview.md).

> **Note**: There are two ways to update your volume: 1.Online 2.Offline. Which Mode to choose?
> It depends on your `StorageClass`. If your storageclass supports online volume expansion, you can go with it. Otherwise, you can go with `Ofline` Volume Expansion.

## CleanUp

```shell
kubectl delete mariadb -n demo mariadb-ha-demo
```


## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/stash/overview/index.md) MariaDB database using Stash.
- Learn about initializing [MariaDB with Script](/docs/guides/postgres/initialization/script_source.md).
- Learn about [custom PostgresVersions](/docs/guides/postgres/custom-versions/setup.md).
- Want to setup MariaDB cluster? Check how to [configure Highly Available MariaDB Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your MariaDB database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your MariaDB database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy MariaDB with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).