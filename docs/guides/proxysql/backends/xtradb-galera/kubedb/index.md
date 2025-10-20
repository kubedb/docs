---
title: Proxy KubeDB Percona XtraDB Galera Cluster With KubeDB ProxySQL
menu:
  docs_{{ .version }}:
    identifier: kubedb-percona-xtradb
    name: KubeDB Managed
    parent: percona-xtradb-backend
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB ProxySQL with PerconaXtraDB Galera Cluster

This guide will show you how to use `KubeDB` operator to set up a `ProxySQL` server for KubeDB managed PerconaXtraDB.

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

## Prepare PerconaXtraDB Backend 

In this tutorial, we are going to set up ProxySQL using KubeDB for a PerconaXtraDB Galera Cluster. We will use KubeDB to set up our PerconaXtraDB servers. 

We need to apply the following YAML to create our PerconaXtraDB Galera Cluster
`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: xtradb-galera
  namespace: demo
spec:
  version: "8.0.40"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/xtradb-galera/kubedb/examples/xtradb-galera.yaml
perconaxtradb.kubedb.com/xtradb-galera created
```

Let's wait for the PerconaXtraDB to be Ready. 

```bash
$ kubectl get px -n demo 
NAME            VERSION   STATUS   AGE
xtradb-galera   8.0.40    Ready    8m
```

Let's first create a user in the backend percona-xtradb server and a database to test the proxy traffic.

```bash
$ kubectl exec -it -n demo xtradb-galera-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-5.1$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 471
Server version: 8.0.40-31.1 Percona XtraDB Cluster (GPL), Release rel3, Revision cf742b4, WSREP version 26.1.4.3

Copyright (c) 2009-2024 Percona LLC and/or its affiliates
Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create user `test`@'%' identified by 'pass';
Query OK, 0 rows affected (0.03 sec)

mysql> create database test;
Query OK, 1 row affected (0.03 sec)

mysql> use test;
Database changed
mysql> show tables;
Empty set (0.00 sec)

mysql> create table testtb(name varchar(103), primary key(name));
Query OK, 0 rows affected (0.12 sec)

mysql> grant all privileges on test.* to 'test'@'%';
Query OK, 0 rows affected (0.05 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.04 sec)

mysql> exit
Bye
```

Now we are ready to deploy and test our ProxySQL server. 

## Deploy ProxySQL Server 

With the following YAML, we are going to create our desired ProxySQL server.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: xtradb-proxy
  namespace: demo
spec:
  version: "2.7.3-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: xtradb-galera
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/xtradb-galera/kubedb/examples/xtradb-proxy.yaml
proxysql.kubedb.com/xtradb-proxy created
```

Here in the `.spec.version` field we are saying that we want a ProxySQL-2.6.3 with base image of debian. In the `.spec.replicas` section we have given 3, so the operator will create 3 nodes for ProxySQL. The `spec.syncUser` field is set to  true, which means all the users in the backend PerconaXtraDB server will be fetched to the ProxySQL server.

Let's wait for the ProxySQL to be Ready. 

```bash
$ kubectl get prx -n demo
NAME            VERSION        STATUS   AGE
xtradb-proxy    2.7.3-debian   Ready    17m
```

Let's check the pods and associated kubernetes objects
```bash
$ kubectl get petset,pods,svc,secrets -n demo
petset.apps.k8s.appscode.com/xtradb-proxy     18m

NAME                         READY   STATUS    RESTARTS        AGE
pod/xtradb-proxy-0           1/1     Running   0               18m
pod/xtradb-proxy-1           1/1     Running   0               18m
pod/xtradb-proxy-2           1/1     Running   0               18m

NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/xtradb-proxy           ClusterIP   10.96.87.237    <none>        6033/TCP            18m
service/xtradb-proxy-pods      ClusterIP   None            <none>        6032/TCP,6033/TCP   18m

NAME                                 TYPE                       DATA   AGE
secret/xtradb-proxy-auth             kubernetes.io/basic-auth   2      18m
secret/xtradb-proxy-configuration    Opaque                     1      18m
secret/xtradb-proxy-monitor          kubernetes.io/basic-auth   2      18m
```

### Check Internal Configuration
Lets exec into the ProxySQL server pod and get into the admin panel. 

```bash
$ kubectl exec -it -n demo xtradb-proxy-0 -- bash
proxysql@xtradb-proxy-0:/$ mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 275
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > 
```

Let's check the mysql_galera_hostgroups and mysql_servers table first. We didn't set it from the YAML. The KubeDB operator will do that for us.

```bash
ProxySQLAdmin > select * from mysql_galera_hostgroups;
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| writer_hostgroup | backup_writer_hostgroup | reader_hostgroup | offline_hostgroup | active | max_writers | writer_is_also_reader | max_transactions_behind | comment |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| 2                | 4                       | 3                | 1                 | 1      | 1           | 1                     | 0                       |         |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
1 row in set (0.000 sec)


ProxySQLAdmin > select * from mysql_servers;
+--------------+---------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                    | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+---------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | xtradb-galera-2.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-2.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+---------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
6 rows in set (0.001 sec)

ProxySQLAdmin > select * from runtime_mysql_servers;
+--------------+---------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                    | port | gtid_port | status  | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+---------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | SHUNNED | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 2            | xtradb-galera-2.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | xtradb-galera-2.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | xtradb-galera-0.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 4            | xtradb-galera-1.xtradb-galera-pods.demo.svc | 3306 | 0         | ONLINE  | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+---------------------------------------------+------+-----------+---------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
8 rows in set (0.003 sec)
```

Here we can see that all the nodes of our PerconaXtraDB Galera cluster has been set to the writer(hg:2) hostgroup and to the reader(hg:3) hostgroup. Since `max_writers` is set to `1`, only `xtradb-galera-2` is `ONLINE` from  hostgroup 2, other nodes are `SHUNNED`.



Let's check the mysql_users table. 

```bash
ProxySQLAdmin > select username from mysql_users;
+----------+
| username |
+----------+
| root     |
| test     |
+----------+
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

Now let's try to connect with the ProxySQL server through the `xtradb-proxy` service as the `test` user.

```bash
root@ubuntu-bb47d8d6c-7wndq:/# mysql -utest -ppass -hxtradb-proxy.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 40
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
| performance_schema |
| test               |
+--------------------+
3 rows in set (0.01 sec)

mysql> use test;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
mysql> insert into testtb(name) values("Kim Torres");
Query OK, 1 row affected (0.02 sec)

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

In this tutorial, we have seen a basic version of KubeDB ProxySQL. KubeDB provides many more for ProxySQL. In this site we have discussed on lots of other features like `TLS Secured ProxySQL` , `Declarative Configuration` , `MySQL and MariaDB Backend` , `Reconfigure` and much more. Checkout out other docs to learn more. 