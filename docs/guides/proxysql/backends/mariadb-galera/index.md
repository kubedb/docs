---
title: Proxy MariaDB Galera Cluster With KubeDB ProxySQL
menu:
  docs_{{ .version }}:
    identifier: mariadb-galera-backend
    name: MariaDB Galera Cluster
    parent: proxysql-backends
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB ProxySQL with MariaDB Galera Cluster

This guide will show you how to use `KubeDB` operator to set up a `ProxySQL` server for KubeDB managed MariaDB.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/index.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare MariaDB Backend 

In this tutorial we are going to set up ProxySQL using KubeDB for a MariaDB Galera Cluster. We will use KubeDB to set up our MariaDB servers. 

We need to apply the following yaml to create our MariaDB Galera Cluster
`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-galera
  namespace: demo
spec:
  version: "11.6.2"
  replicas: 3
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mariadb-galera/examples/mariadb-galera.yaml
mariadb.kubedb.com/mariadb-galera created
```

Let's wait for the MariaDB to be Ready. 

```bash
$ kubectl get md -n demo 
NAME             VERSION   STATUS   AGE
mariadb-galera   11.6.2   Ready    4m20s
```

Let's first create a user in the backend mariadb server and a database to test the proxy traffic.

```bash
$ kubectl exec -it -n demo mariadb-galera-0 -- bash
Defaulted container "mariadb" out of: mariadb, md-coordinator, mariadb-init (init)
mysql@mariadb-galera-0:/$ mariadb -uroot -p$MYSQL_ROOT_PASSWORD
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 52
Server version: 11.6.2-MariaDB-1:11.6.2+maria~ubu2004 mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]>  create user `test`@'%' identified by 'pass';
Query OK, 0 rows affected (0.040 sec)

MariaDB [(none)]> create database test;
Query OK, 1 row affected (0.033 sec)

MariaDB [(none)]> use test;
Database changed
MariaDB [test]> show tables;
Empty set (0.000 sec)

MariaDB [test]> create table testtb(name varchar(103), primary key(name));
Query OK, 0 rows affected (0.077 sec)

MariaDB [test]> grant all privileges on test.* to 'test'@'%';
Query OK, 0 rows affected (0.036 sec)

MariaDB [test]> flush privileges;
Query OK, 0 rows affected (0.033 sec)

MariaDB [test]> exit
Bye
```

Now we are ready to deploy and test our ProxySQL server. 

## Deploy ProxySQL Server 

With the following yaml we are going to create our desired ProxySQL server.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: mariadb-proxy
  namespace: demo
spec:
  version: "2.7.3-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: mariadb-galera
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mariadb-galera/examples/sample-proxysql.yaml
proxysql.kubedb.com/mariadb-proxy created
```

Here in the `.spec.version` field we are saying that we want a ProxySQL-2.6.3 with base image of debian. In the `.spec.replicas` section we have given 3, so the operator will create 3 nodes for ProxySQL. The `spec.syncUser` field is set to  true, which means all the users in the backend MariaDB server will be fetched to the ProxySQL server.

Let's wait for the ProxySQL to be Ready. 

```bash
$ kubectl get prx -n demo
NAME            VERSION        STATUS   AGE
mariadb-proxy   2.7.3-debian   Ready    96s
```

Let's check the pods and associated kubernetes objects
```bash
$ kubectl get petset,pods,svc,secrets -n demo
NAME                                          AGE
petset.apps.k8s.appscode.com/mariadb-proxy    108s

NAME                         READY   STATUS    RESTARTS   AGE
pod/mariadb-proxy-0          1/1     Running   0          108s
pod/mariadb-proxy-1          1/1     Running   0          107s
pod/mariadb-proxy-2          1/1     Running   0          106s

NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/mariadb-proxy          ClusterIP   10.96.133.102   <none>        6033/TCP            110s
service/mariadb-proxy-pods     ClusterIP   None            <none>        6032/TCP,6033/TCP   110s

NAME                                 TYPE                       DATA   AGE
secret/mariadb-proxy-auth            kubernetes.io/basic-auth   2      110s
secret/mariadb-proxy-configuration   Opaque                     1      109s
secret/mariadb-proxy-monitor         kubernetes.io/basic-auth   2      110s
```

### Check Internal Configuration
Lets exec into the ProxySQL server pod and get into the admin panel. 

```bash
$ kubectl exec -it -n demo mariadb-proxy-0 -- bash
proxysql@mariadb-proxy-0:/$  mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 48
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > 
```

Let's check the mysql_galera_hostgroups and mysql_servers table first. We didn't set it from the yaml. The KubeDB operator will do that for us. 

```bash
ProxySQLAdmin > select * from mysql_galera_hostgroups;
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| writer_hostgroup | backup_writer_hostgroup | reader_hostgroup | offline_hostgroup | active | max_writers | writer_is_also_reader | max_transactions_behind | comment |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| 2                | 4                       | 3                | 1                 | 1      | 1           | 1                     | 0                       |         |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
1 row in set (0.001 sec)

ProxySQLAdmin > select * from mysql_servers;
+--------------+-----------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                      | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+-----------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | mariadb-galera-0.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-0.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mariadb-galera-1.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-1.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mariadb-galera-2.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-2.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+-----------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
6 rows in set (0.000 sec)

ProxySQLAdmin > select * from runtime_mysql_servers;
+--------------+-----------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                      | port | gtid_port | status  | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+-----------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | mariadb-galera-0.mariadb-galera-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mariadb-galera-1.mariadb-galera-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | mariadb-galera-2.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-0.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-1.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mariadb-galera-2.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | mariadb-galera-0.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | mariadb-galera-1.mariadb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+-----------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
8 rows in set (0.002 sec)
```

Here we can see that all the nodes of our MariaDB Galera cluster have been set to the writer(hg:2) hostgroup and to the reader(hg:3) hostgroup. Since `max_writers` is set to `1`, only `mariadb-galera-2` is `ONLINE` from hostgroup 2, other two Galera nodes are placed into the writer(hg:2) hostgroup but marked `SHUNNED` because only one writer is permitted.

Let's check the mysql_users table. 

```bash
ProxySQLAdmin > select username from mysql_users;
+-------------+
| username    |
+-------------+
| healthcheck |
| root        |
| test        |
+-------------+
3 rows in set (0.001 sec)
```

So test user is automatically synced in proxysql and present in mysql_users, we are now ready to test our traffic proxy.

### Check Traffic Proxy

To test the traffic routing through the ProxySQL server let's first create a pod with ubuntu base image in it. We will use the following yaml. 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ubuntu
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ubuntu
  template:
    metadata:
      labels:
        app: ubuntu
    spec:
      containers:
        - image: ubuntu
          imagePullPolicy: IfNotPresent
          name: ubuntu
          command: ["/bin/sleep", "3650d"]
```

Let's apply the yaml. 

```yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mariadb-galera/examples/ubuntu.yaml
deployment.apps/ubuntu created
```

Lets exec into the pod and install mariadb-galera-client. 

```bash
$ kubectl exec -it -n demo ubuntu-bb47d8d6c-7wndq -- bash
root@ubuntu-bb47d8d6c-7wndq:/# apt update
... ... ..
root@ubuntu-bb47d8d6c-7wndq:/# apt install mysql-client -y
Reading package lists... Done
... .. ...
```

Now let's try to connect with the ProxySQL server through the `mariadb-proxy` service as the `test` user. 

```bash
root@ubuntu-bb47d8d6c-7wndq:/# mysql -utest -ppass -hmariadb-proxy.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 126
Server version: 8.0.27 (ProxySQL)

Copyright (c) 2000, 2025, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| test               |
+--------------------+
2 rows in set (0.00 sec)

mysql> use test;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
mysql> insert into testtb(name) values("Kim Torres");
Query OK, 1 row affected (0.00 sec)

mysql> insert into testtb(name) values("Tony SoFua");
Query OK, 1 row affected (0.01 sec)

mysql> select * from testtb;
+------------+
| name       |
+------------+
| Kim Torres |
| Tony SoFua |
+------------+
2 rows in set (0.00 sec)
```

We can see the queries are successfully executed through the ProxySQL server. 

Let's check the query splits inside the ProxySQL server by going back to the ProxySQLAdmin panel. 

```bash
ProxySQLAdmin > select hostgroup,Queries from stats_mysql_connection_pool;
+-----------+---------+
| hostgroup | Queries |
+-----------+---------+
| 2         | 0       |
| 2         | 0       |
| 2         | 3       |
| 3         | 1       |
| 3         | 2       |
| 3         | 0       |
| 4         | 0       |
| 4         | 0       |
+-----------+---------+
8 rows in set (0.003 sec)
```

We can see that the read-write split is successfully executed in the ProxySQL server. So the ProxySQL server is ready to use.

## Conclusion 

In this tutorial, we have seen a basic version of KubeDB ProxySQL. KubeDB provides many more for ProxySQL. In this site we have discussed on lots of other features like `TLS Secured ProxySQL` , `Declarative Configuration` , `MySQL and Percona-XtraDB Backend` , `Reconfigure` and much more. Checkout out other docs to learn more. 