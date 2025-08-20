---
title: MariaDB Failover and DR Scenarios
menu:
  docs_{{ .version }}:
    identifier: mariadb-failover
    name: Overview
    parent: guides-mariadb-FDR
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

#  Exploring Fault Tolerance in MariaDB with KubeDB

## Understanding Failover and Clustering in MariaDB on KubeDB
`Failover` refers to the process of automatically switching to a slave system or replica when the `Master`
database node fails. In high-availability database systems, failover ensures that services remain
uninterrupted even when one or more nodes go down. This capability is critical in modern, cloud-native 
infrastructure where downtime can lead to major disruptions.

When running `MariaDB` on Kubernetes using KubeDB, failover becomes more seamless. KubeDB supports two types of MariaDB clustering strategies:

- **Standard Replication (Master-Slave):**
 Standard Replication is a mechanism where one server (master) handles both read and write operations,
while one or more servers (slaves) replicate data asynchronously and serve read-only queries. 
This setup improves data redundancy, availability, and read scalability, with automatic failover and
load balancing supported through `MariaDB MaxScale Server`.

>note: Writing to a slave replica may result in a binary log (binlog) conflict issue.

- **Galera Cluster (Multi-`Master`):**
In this setup, all nodes act as `Master`, capable of handling both read and write operations. Since there’s
no single point of failure, the system provides synchronous replication and built-in high availability, 
but doesn’t use the traditional failover concept, as all pods are equal.

In the rest of this blog, we'll focus on how failover works in the `standard replication` mode, and how
KubeDB handles recovery in the event of a node failure.


## Before You Begin

Before proceeding:


- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to 
communicate with your cluster. If you do not already have a cluster, you can create one by 
using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- Read [mariadb cluster concept](/docs/guides/mariadb/clustering/overview) to learn about MariaDB
  Cluster.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. 
Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```
## Deploy MariaDB Cluster

The following is an example `MariaDB` object which creates a single-master MariaDB `standard replication` cluster with three members.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: ha-mariadb
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
mariadb.kubedb.com/ha-mariadb created
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
mariadb.kubedb.com/ha-mariadb   10.6.16   Ready    3m27s

NAME                                             AGE
petset.apps.k8s.appscode.com/ha-mariadb      3m20s
petset.apps.k8s.appscode.com/ha-mariadb-mx   3m23s

NAME                      READY   STATUS    RESTARTS   AGE
pod/ha-mariadb-0      2/2     Running   0          3m20s
pod/ha-mariadb-1      2/2     Running   0          3m20s
pod/ha-mariadb-2      2/2     Running   0          3m20s
pod/ha-mariadb-mx-0   1/1     Running   0          3m23s
pod/ha-mariadb-mx-1   1/1     Running   0          3m23s
pod/ha-mariadb-mx-2   1/1     Running   0          3m23s

```

Inspect who is `Master` and who is `slave`.

```shell
# you can inspect the role of the pods 

$ kubectl get pods -n demo --show-labels | grep role
ha-mariadb-0      2/2     Running   0          4m9s    app.kubernetes.io/component=database,app.kubernetes.io/instance=ha-mariadb,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mariadbs.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=ha-mariadb-598cd56869,kubedb.com/role=Master,statefulset.kubernetes.io/pod-name=ha-mariadb-0
ha-mariadb-1      2/2     Running   0          4m9s    app.kubernetes.io/component=database,app.kubernetes.io/instance=ha-mariadb,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mariadbs.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=ha-mariadb-598cd56869,kubedb.com/role=Slave,statefulset.kubernetes.io/pod-name=ha-mariadb-1
ha-mariadb-2      2/2     Running   0          4m9s    app.kubernetes.io/component=database,app.kubernetes.io/instance=ha-mariadb,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mariadbs.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=ha-mariadb-598cd56869,kubedb.com/role=Slave,statefulset.kubernetes.io/pod-name=ha-mariadb-2

```
The pod having `kubedb.com/role=Master` is the `Master` and `kubedb.com/role=Slave` are the slaves.
You can also check it on the cluster status:
```shell
$ kubectl exec -it -n demo svc/ha-mariadb-mx -- bash
Defaulted container "maxscale" out of: maxscale, maxscale-init (init)
bash-4.4$ maxctrl list servers
┌─────────┬─────────────────────────────────────────────────────────────┬──────┬─────────────┬─────────────────┬─────────┬────────────────────┐
│ Server  │ Address                                                     │ Port │ Connections │ State           │ GTID    │ Monitor            │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼─────────┼────────────────────┤
│ server1 │ ha-mariadb-0.ha-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Master, Running │ 0-1-217 │ ReplicationMonitor │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼─────────┼────────────────────┤
│ server2 │ ha-mariadb-1.ha-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Slave, Running  │ 0-1-217 │ ReplicationMonitor │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼─────────┼────────────────────┤
│ server3 │ ha-mariadb-2.ha-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Slave, Running  │ 0-1-217 │ ReplicationMonitor │
└─────────┴─────────────────────────────────────────────────────────────┴──────┴─────────────┴─────────────────┴─────────┴────────────────────┘

```
## Verify Pod Reachability and Status
Once the database is in running state we can connect to each of three nodes.
We will use login credentials `MYSQL_ROOT_USERNAME` and `MYSQL_ROOT_PASSWORD` saved as container's environment variable.

### Create a Test User
Writing to a slave replica can cause binlog conflicts. By default, slave-replicas are read-only, but a 
root user (with super privileges) can still make changes. For security, avoid using the root user
in production and create a dedicated user with only the needed permissions instead.
```bash
$ kubectl exec -it -n demo svc/ha-mariadb -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@ha-mariadb-0:/$ mariadb -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 443
Server version: 10.6.16-MariaDB-1:10.6.16+maria~ubu2004-log mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]>  CREATE USER 'testuser'@'%' IDENTIFIED BY 'testpassword';
Query OK, 0 rows affected (0.003 sec)

MariaDB [(none)]>  GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, INDEX, ALTER, SHOW DATABASES ON *.* TO 'testuser'@'%' WITH GRANT OPTION;
Query OK, 0 rows affected (0.003 sec)

MariaDB [(none)]> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.001 sec)

MariaDB [(none)]> quit
Bye
mysql@ha-mariadb-0:/$ exit
exit

```
### Check Connectivity using Test User

```bash
# Master Node
$ kubectl exec -it -n demo svc/ha-mariadb -- bash
mysql@ha-mariadb-0:/ mariadb -utestuser -ptestpassword
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 26
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SELECT 1; 
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

MariaDB [(none)]> quit;
Bye

# Slave Node
$ kubectl exec -it -n demo svc/ha-mariadb-slave -- bash
mysql@ha-mariadb-1:/ mariadb -utestuser -ptestpassword
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 94
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SELECT 1;
+---+
| 1 |
+---+
| 1 |
+---+
1 row in set (0.000 sec)

MariaDB [(none)]> quit;
Bye

MariaDB [(none)]> quit;
Bye
```

## Insert Data and Check Availability

In a MariaDB Replication Cluster, Only master member can write, and slave member can read. In this section, we will insert data from master node, and we will see whether we can get the data from every other slave members.
**Check which is the master pod**
```shell
$ kubectl get pods -n demo --show-labels | grep Master | awk '{ print $1 }'
ha-mariadb-0

```
let's insert data in the master node
```bash
$ kubectl exec -it -n demo ha-mariadb-0 -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@ha-mariadb-0:/$ mariadb -utestuser -ptestpassword
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 3459
Server version: 10.6.16-MariaDB-1:10.6.16+maria~ubu2004-log mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE DATABASE playground;
Query OK, 1 row affected (0.001 sec)

MariaDB [(none)]> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), `Master` KEY(id));
Query OK, 0 rows affected (0.032 sec)

MariaDB [(none)]> INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');
Query OK, 1 row affected (0.008 sec)

MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
Bye
mysql@ha-mariadb-0:/$ exit
exit
```
You can read data from the slave nodes
```shell
kubectl exec -it -n demo ha-mariadb-1 -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@ha-mariadb-1:/$ mariadb -utestuser -ptestpassword
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 2304
Server version: 10.6.16-MariaDB-1:10.6.16+maria~ubu2004-log mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.000 sec)

MariaDB [(none)]> exit
Bye
mysql@ha-mariadb-1:/$ exit
exit

```

## How Failover Works

Before simulating failover let's know how it works with the help of maxscale.

MaxScale uses a monitor (like the MariaDB-Monitor plugin) to track the health of database nodes in a replication setup (e.g. MariaDB master-slave). If the `Master` node becomes unavailable—due to a crash, network issue, or maintenance—

MaxScale:
- Detects the Failure: The monitor(ReplicationMonitor) continuously checks node status (using MySQL pings or status variables).
- Selects a New `Master`: It identifies the most suitable replica based on criteria like replication lag or server state.
- Promotes the Replica: MaxScale executes commands to promote the chosen slave-replica to `Master` (e.g. STOP SLAVE; RESET SLAVE ALL;).
- Reconfigures Replicas: Other slave-replicas are updated to replicate from the new `Master`.
- Redirects Traffic: MaxScale’s router (RW-Split-Router) seamlessly directs write queries to the new `Master` and read queries to slave-replicas.

This process happens automatically, typically within seconds, ensuring minimal disruption.
Lets open another terminal and monitor the state of all the pods:
```shell
$ watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
You'll see:
```shell
ha-mariadb-0 Master
ha-mariadb-1 Slave
ha-mariadb-2 Slave
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2
```
### Hands-on Failover Testing

#### Case 1: Delete the current `Master`

Let's delete the current `Master` pod and see how the role change happens almost immediately.

```shell
$ kubectl delete pods -n demo ha-mariadb-0 
pod "ha-mariadb-0" deleted
```
You'll see the pods' status like that:
```
ha-mariadb-0 Down
ha-mariadb-1 Master
ha-mariadb-2 Slave
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2

```
After few minutes `ha-mariadb-0` will back as `Slave` pod 
```shell
ha-mariadb-0 Slave
ha-mariadb-1 Master
ha-mariadb-2 Slave
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2
```

Now we know how failover is done, let's check if the new `Master` is working.

```shell
$ kubectl exec -it -n demo ha-mariadb-1 -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@ha-mariadb-1:/$ mariadb -utestuser -ptestpassword
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 652
Server version: 10.6.16-MariaDB-1:10.6.16+maria~ubu2004-log mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
6 rows in set (0.001 sec)

MariaDB [(none)]> CREATE DATABASE playground2;
Query OK, 1 row affected (0.000 sec)
MariaDB [(none)]> exit
Bye
mysql@ha-mariadb-1:/$ exit
exit

```

Lets check if the new Slave(`ha-mariadb-0`) got the updated data from new `Master`, `ha-mariadb-1`.

```shell
$  kubectl exec -it -n demo ha-mariadb-0 -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@ha-mariadb-0:/$ mariadb -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 340
Server version: 10.6.16-MariaDB-1:10.6.16+maria~ubu2004-log mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| playground         |
| playground2        |
| sys                |
+--------------------+
7 rows in set (0.001 sec)

MariaDB [(none)]> exit 
Bye
mysql@ha-mariadb-0:/$ exit
exit

```

#### Case 2: Delete the current `Master` and One slave

```shell
$ kubectl delete pods -n demo ha-mariadb-1 ha-mariadb-2
pod "ha-mariadb-1" deleted
pod "ha-mariadb-2" deleted
```
Again we can see the failover happened pretty quickly.
```shell
ha-mariadb-0 Master
ha-mariadb-1 Down
ha-mariadb-2 Down
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2
```
After 10-30 second, the deleted pods will be back and will have its role and both will have Slave role.
```shell
ha-mariadb-0 Master
ha-mariadb-1 Slave
ha-mariadb-2 Slave
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2
```
Lets validate the cluster state from new `Master`(`ha-mariadb-0`).

```shell
$ kubectl exec -it -n demo ha-mariadb-0 -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@ha-mariadb-0:/$ mariadb -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 377
Server version: 10.6.16-MariaDB-1:10.6.16+maria~ubu2004-log mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| playground         |
| playground2        |
| sys                |
+--------------------+
7 rows in set (0.000 sec)

MariaDB [(none)]> CREATE DATABASE playground3;
Query OK, 1 row affected (0.000 sec)

MariaDB [(none)]> exit
Bye
mysql@ha-mariadb-0:/$ exit
exit

```
Let's check whether the Slave nodes, `ha-mariadb-2` gets all the previous data 
```shell
kubectl exec -it -n demo ha-mariadb-2 -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@ha-mariadb-2:/$ mariadb -utestuser -ptestpassword
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 47
Server version: 10.6.16-MariaDB-1:10.6.16+maria~ubu2004-log mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| playground         |
| playground2        |
| playground3        |
| sys                |
+--------------------+
8 rows in set (0.001 sec)


```

#### Case3: Delete any of the slave's

Let's delete both of the slave nodes.

```shell
kubectl delete pods -n demo ha-mariadb-1 ha-mariadb-2
pod "ha-mariadb-1" deleted
pod "ha-mariadb-2" deleted

```

```shell

ha-mariadb-0 Master
ha-mariadb-1 Down
ha-mariadb-2 Down
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2
```

Shortly both of the pods will be back with its role.

```shell
ha-mariadb-0 Master
ha-mariadb-1 Slave
ha-mariadb-2 Slave
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2
```


#### Case 4: Delete both `Master` and all slave-replicas

Let's delete all the pods.

```shell
$ kubectl delete pods -n demo ha-mariadb-0 ha-mariadb-1 ha-mariadb-2
pod "ha-mariadb-0" deleted
pod "ha-mariadb-1" deleted
pod "ha-mariadb-2" deleted

```
```shell
ha-mariadb-0 Down
ha-mariadb-1 Down
ha-mariadb-2 Down
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2

```

Within 20-30 second, all of the pod should be back.

```shell
ha-mariadb-0 Master
ha-mariadb-1 Slave
ha-mariadb-2 Slave
ha-mariadb-mx-0
ha-mariadb-mx-1
ha-mariadb-mx-2

```

Lets verify the cluster state now.

```shell
$ kubectl exec -it -n demo svc/ha-mariadb-mx -- bash
Defaulted container "maxscale" out of: maxscale, maxscale-init (init)
bash-4.4$ maxctrl list servers
┌─────────┬─────────────────────────────────────────────────────────────┬──────┬─────────────┬─────────────────┬──────────┬────────────────────┐
│ Server  │ Address                                                     │ Port │ Connections │ State           │ GTID     │ Monitor            │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼──────────┼────────────────────┤
│ server1 │ ha-mariadb-0.ha-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Master, Running │ 0-1-3297 │ ReplicationMonitor │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼──────────┼────────────────────┤
│ server2 │ ha-mariadb-1.ha-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Slave, Running  │ 0-1-3297 │ ReplicationMonitor │
├─────────┼─────────────────────────────────────────────────────────────┼──────┼─────────────┼─────────────────┼──────────┼────────────────────┤
│ server3 │ ha-mariadb-2.ha-mariadb-pods.demo.svc.cluster.local │ 3306 │ 0           │ Slave, Running  │ 0-1-3297 │ ReplicationMonitor │
└─────────┴─────────────────────────────────────────────────────────────┴──────┴─────────────┴─────────────────┴──────────┴────────────────────┘

```

## CleanUp
For cleaning up what we created in this tutorial follow the following command:
```shell
$ kubectl delete mariadb -n demo ha-mariadb
mariadb.kubedb.com "ha-mariadb" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```


## Next Steps

- Learn about [backup and restore](/docs/guides/mariadb/backup/stash/overview/index.md) MariaDB database using Stash.
- Learn about initializing [MariaDB with Script](/docs/guides/mariadb/initialization/script_source.md).
- Learn about [custom MariaDBVersions](/docs/guides/mariadb/custom-versions/setup.md).
- Want to setup MariaDB cluster? Check how to [configure Highly Available MariaDB Cluster](/docs/guides/mariadb/clustering/ha_cluster.md)
- Monitor your MariaDB database with KubeDB using [built-in Prometheus](/docs/guides/mariadb/monitoring/using-builtin-prometheus.md).
- Monitor your MariaDB database with KubeDB using [Prometheus operator](/docs/guides/mariadb/monitoring/using-prometheus-operator.md).
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb.md).
- Use [private Docker registry](/docs/guides/mariadb/private-registry/using-private-registry.md) to deploy MariaDB with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).