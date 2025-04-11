---
title: Load Balance MySQL Group Replication With KubeDB ProxySQL
menu:
  docs_{{ .version }}:
    identifier: mysql-backend
    name: MySQL Group Replication
    parent: proxysql-backends
    weight: 10
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB ProxySQL with MySQL Group Replication

This guide will show you how to use `KubeDB` operator to set up a `ProxySQL` server for KubeDB managed MySQL.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/index.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare MySQL Backend 

In this tutorial we are going to set up ProxySQL using KubeDB for a MySQL Group Replication. We will use KubeDB to set up our MySQL servers. 

We need to apply the following yaml to create our MySQL Group Replication
`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.0.36"
  replicas: 3
  topology:
    mode: GroupReplication
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mysql/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's wait for the MySQL to be Ready. 

```bash
$ kubectl get my -n demo 
NAME           VERSION   STATUS   AGE
mysql-server   8.0.36    Ready    7m6s
```

Let's first create a user in the backend mysql server and a database to test the proxy traffic.

```bash
$ kubectl exec -it -n demo mysql-server-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
mysql@mysql-server-0:/$  mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 474
Server version: 8.0.36 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql>  create user `test`@'%' identified by 'pass';
Query OK, 0 rows affected (0.03 sec)

mysql> create database test;
Query OK, 1 row affected (0.04 sec)

mysql> use test;
Database changed
mysql> show tables;
Empty set (0.00 sec)

mysql> create table testtb(name varchar(103), primary key(name));
Query OK, 0 rows affected (0.13 sec)

mysql> grant all privileges on test.* to 'test'@'%';
Query OK, 0 rows affected (0.03 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.01 sec)

mysql> select * FROM performance_schema.replication_group_members;
+---------------------------+--------------------------------------+-------------------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| CHANNEL_NAME              | MEMBER_ID                            | MEMBER_HOST                               | MEMBER_PORT | MEMBER_STATE | MEMBER_ROLE | MEMBER_VERSION | MEMBER_COMMUNICATION_STACK |
+---------------------------+--------------------------------------+-------------------------------------------+-------------+--------------+-------------+----------------+----------------------------+
| group_replication_applier | b7ed7a2d-1532-11f0-9b1a-5ad095e72795 | mysql-server-0.mysql-server-pods.demo.svc |        3306 | ONLINE       | PRIMARY     | 8.0.36         | XCom                       |
| group_replication_applier | b7ed7a33-1532-11f0-9b20-6ec1a256386b | mysql-server-2.mysql-server-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.36         | XCom                       |
| group_replication_applier | bae6c112-1532-11f0-86bb-ce260ac82c19 | mysql-server-1.mysql-server-pods.demo.svc |        3306 | ONLINE       | SECONDARY   | 8.0.36         | XCom                       |
+---------------------------+--------------------------------------+-------------------------------------------+-------------+--------------+-------------+----------------+----------------------------+
3 rows in set (0.00 sec)

mysql> exit
Bye
```

This output is from the performance_schema.replication_group_members table in MySQL shows the status of nodes in a Group Replication (GR) setup.
We have 3 nodes in your MySQL Group Replication cluster. All 3 nodes are ONLINE – they are healthy and actively participating in replication. mysql-server-0 is the PRIMARY node – it's the one accepting write queries. And mysql-server-1 and mysql-server-2 are SECONDARY – they receive updates from the primary but are read-only.

Now we are ready to deploy and test our ProxySQL server. 

## Deploy ProxySQL Server 

With the following yaml we are going to create our desired ProxySQL server.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `kubedb.com/v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: mysql-proxy
  namespace: demo
spec:
  version: "2.6.3-debian"
  replicas: 3
  syncUsers: true
  backend:
    name: mysql-server
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mysql/examples/sample-proxysql.yaml
proxysql.kubedb.com/mysql-proxy created
```

Here in the `.spec.version` field we are saying that we want a ProxySQL-2.6.3 with base image of debian. In the `.spec.replicas` section we have given 3, so the operator will create 3 nodes for ProxySQL. The `spec.syncUser` field is set to  true, which means all the users in the backend MySQL server will be fetched to the ProxySQL server.

Let's wait for the ProxySQL to be Ready. 

```bash
$ kubectl get prx -n demo
NAME          VERSION        STATUS   AGE
mysql-proxy   2.6.3-debian   Ready    109s
```

Let's check the pods and associated kubernetes objects
```bash
$ kubectl get petset,pods,svc,secrets -n demo
NAME                                        AGE
petset.apps.k8s.appscode.com/mysql-proxy    3m59s

NAME                         READY   STATUS    RESTARTS        AGE
pod/mysql-proxy-0            1/1     Running   0               3m59s
pod/mysql-proxy-1            1/1     Running   0               3m58s
pod/mysql-proxy-2            1/1     Running   0               3m57s

NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/mysql-proxy            ClusterIP   10.96.246.141   <none>        6033/TCP            4m1s
service/mysql-proxy-pods       ClusterIP   None            <none>        6032/TCP,6033/TCP   4m1s

NAME                               TYPE                       DATA   AGE
secret/mysql-proxy-auth            kubernetes.io/basic-auth   2      4m1s
secret/mysql-proxy-configuration   Opaque                     1      4m1s
secret/mysql-proxy-monitor         kubernetes.io/basic-auth   2      4m1s
```

### Check Internal Configuration
Lets exec into the ProxySQL server pod and get into the admin panel. 

```bash
$ kubectl exec -it -n demo mysql-proxy-0 -- bash
proxysql@mysql-proxy-0:/$  mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > " 
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 93
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > 
```

Let's check the mysql_group_replication_hostgroups and mysql_servers table first. We didn't set it from the yaml. The KubeDB operator will do that for us. 

```bash
ProxySQLAdmin > select * from mysql_group_replication_hostgroups;
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| writer_hostgroup | backup_writer_hostgroup | reader_hostgroup | offline_hostgroup | active | max_writers | writer_is_also_reader | max_transactions_behind | comment |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
| 2                | 4                       | 3                | 1                 | 1      | 1           | 1                     | 0                       |         |
+------------------+-------------------------+------------------+-------------------+--------+-------------+-----------------------+-------------------------+---------+
1 row in set (0.001 sec)

ProxySQLAdmin > select * from mysql_servers;
+--------------+-------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                                  | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+-------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | mysql-server-0.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-0.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-1.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-2.mysql-server-pods.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+-------------------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
4 rows in set (0.001 sec)
```

Here we can see that the primary node of our MySQL Group Replication cluster has been set to the writer(hg:2) hostgroup and the secondary nodes has been set to the reader(hg:3) hostgroup.

Let's check the mysql_users table. 

```bash
ProxySQLAdmin > select username from mysql_users;
+----------+
| username |
+----------+
| root     |
| test     |
+----------+
2 rows in set (0.000 sec)
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mysql/examples/ubuntu.yaml
deployment.apps/ubuntu created
```

Lets exec into the pod and install mysql-client. 

```bash
$ kubectl exec -it -n demo ubuntu-bb47d8d6c-7wndq -- bash
root@ubuntu-bb47d8d6c-7wndq:/# apt update
... ... ..
root@ubuntu-bb47d8d6c-7wndq:/# apt install mysql-client -y
Reading package lists... Done
... .. ...
```

Now let's try to connect with the ProxySQL server through the `mysql-proxy` service as the `test` user. 

```bash
root@ubuntu-bb47d8d6c-7wndq:/# mysql -utest -ppass -hproxy-server.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 2005 (HY000): Unknown MySQL server host 'proxy-server.demo.svc' (-2)
root@ubuntu-bb47d8d6c-7wndq:/# mysql -utest -ppass -hmysql-proxy.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 348
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
Query OK, 1 row affected (0.03 sec)

mysql> insert into testtb(name) values("Tony SoFua");
Query OK, 1 row affected (0.02 sec)

mysql> select * from testtb;
+------------+
| name       |
+------------+
| Kim Torres |
| Tony SoFua |
+------------+
2 rows in set (0.00 sec)

mysql> 
```

We can see the queries are successfully executed through the ProxySQL server. 

Let's check the query splits inside the ProxySQL server by going back to the ProxySQLAdmin panel. 

```bash
ProxySQLAdmin > select hostgroup,Queries from stats_mysql_connection_pool;
+-----------+---------+
| hostgroup | Queries |
+-----------+---------+
| 2         | 6       |
| 3         | 0       |
| 3         | 1       |
| 3         | 2       |
+-----------+---------+
4 rows in set (0.002 sec)
```

We can see that the read-write split is successfully executed in the ProxySQL server. So the ProxySQL server is ready to use.

## Conclusion 

In this tutorial we have seen basic version of KubeDB ProxySQL. KubeDB provides many more for ProxySQL. In this site we have discussed on lots of other features like `TLS Secured ProxySQL` , `Declarative Configuration` , `MariaDB and Percona-XtraDB Backend` , `Reconfigure` and much more. Checkout out other docs to learn more. 