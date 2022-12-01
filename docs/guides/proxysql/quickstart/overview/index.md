---
title: Proxy Load To MySQL Group Replication With KubeDB Provisioned ProxySQL
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-quickstart-overview
    name: Concepts
    parent: guides-proxysql-quickstart
    weight: 20
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# KubeDB ProxySQL Quickstart

This guide will show you how to use `KubeDB` Enterprise operator to set up a `ProxySQL` server for KubeDB managed MySQL.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare MySQL Backend 

In this tutorial we are going to test set up a ProxySQL server with KubeDB operator for a MySQL Group Replication. We will use KubeDb to set up our MySQL servers here. 
By applying the following yaml we are going to create our MySQL Group Replication 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "5.7.36"
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
  terminationPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/quickstart/overview/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's wait for the MySQL to be Ready. 

```bash
$ kubectl get mysql -n demo 
NAME           VERSION   STATUS   AGE
mysql-server   5.7.36    Ready    3m51s
```

Let's first create a user in the backend mysql server and a database to test the proxy traffic .

```bash
$ kubectl exec -it -n demo mysql-server-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
root@mysql-server-0:/# mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 195
Server version: 5.7.36-log MySQL Community Server (GPL)

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create user `test`@'%' identified by 'pass';
Query OK, 0 rows affected (0.00 sec)

mysql> create database test;
Query OK, 1 row affected (0.01 sec)

mysql> use test;
Database changed

mysql> show tables;
Empty set (0.00 sec)

mysql> create table testtb(name varchar(103), primary key(name));
Query OK, 0 rows affected (0.01 sec)

mysql> grant all privileges on test.* to 'test'@'%';
Query OK, 0 rows affected (0.00 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.00 sec)

mysql> exit
Bye
```

Now we are ready to deploy and test our ProxySQL server. 

## Deploy ProxySQL Server 

With the following yaml we are going to create our desired ProxySQL server.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 1
  syncUsers: true
  backend:
    name: mysql-server
  terminationPolicy: WipeOut
```

This is the simplest version of a KubeDB ProxySQL server. Here in the `.spec.version` field we are saying that we want a ProxySQL-2.3.2 with base image of debian. In the `.spec.replicas` section we have written 1, so the operator will create a single node ProxySQL. The `spec.syncUser` field is set to  true, which means all the users in the backend MySQL server will be fetched to the ProxySQL server. 

Now let's apply the yaml. 

```yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/quickstart/overview/examples/sample-proxysql.yaml
  proxysql.kubedb.com/proxysql-server created
```

Let's wait for the ProxySQL to be Ready. 

```bash
$ kubectl get proxysql -n demo
NAME           VERSION        STATUS   AGE
proxy-server   2.3.2-debian   Ready    4m
```

Let's check the pod.

```bash
$ kubectl get pods -n demo | grep proxy
proxy-server-0   1/1     Running   0          4m
```

### Check Associated Kubernetes Objects

KubeDB operator will create some services and secrets for the ProxySQL object. Let's check. 

```bash
$ kubectl get svc,secret -n demo | grep proxy
service/proxy-server          ClusterIP   10.96.181.182   <none>        6033/TCP            4m
service/proxy-server-pods     ClusterIP   None            <none>        6032/TCP,6033/TCP   4m
secret/proxy-server-auth             kubernetes.io/basic-auth              2      4m
secret/proxy-server-configuration    Opaque                                1      4m
secret/proxy-server-monitor          kubernetes.io/basic-auth              2      4m
```

You can find the description of the associated objects here. 

### Check Internal Configuration 

Let's exec into the ProxySQL server pod and get into the admin panel. 

```bash
$ kubectl exec -it -n demo proxy-mysql-0 -- bash                                                  11:20
root@proxy-mysql-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQLAdmin > " 
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 1204
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > 
```

Let's check the mysql_servers table first. We didn't set it from the yaml. The KubeDB operator will do that for us. 

```bash
ProxySQLAdmin > select * from mysql_servers;
+--------------+-------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| hostgroup_id | hostname                      | port | gtid_port | status | weight | compression | max_connections | max_replication_lag | use_ssl | max_latency_ms | comment |
+--------------+-------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
| 2            | mysql-server.demo.svc         | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
| 3            | mysql-server-standby.demo.svc | 3306 | 0         | ONLINE | 1      | 0           | 1000            | 0                   | 0       | 0              |         |
+--------------+-------------------------------+------+-----------+--------+--------+-------------+-----------------+---------------------+---------+----------------+---------+
2 rows in set (0.000 sec)
```

Here we can see that the primary service of our MySQL instance has been set to the writer(hg:2) hostgroup and the secondary service has been set to the reader(hg:3) hostgroup. KubeDB MySQL group replication usually creates two services. The primary one forwards query to the writer node and the secondary one to the readers. 

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

So we are now ready to test our traffic proxy. In the next section we are going to have some demo's. 

### Check Traffic Proxy

To test the traffic routing through the ProxySQL server let's first create a pod with ubuntu base image in it. We will use the following yaml. 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: ubuntu
  name: ubuntu
  namespace: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ubuntu
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: ubuntu
    spec:
      containers:
        - image: ubuntu
          imagePullPolicy: IfNotPresent
          name: ubuntu
          command: ["/bin/sleep", "3650d"]
          resources: {}
```

Let's apply the yaml. 

```yaml
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/quickstart/overview/examples/ubuntu.yaml
deployment.apps/ubuntu created
```

Let's exec into the pod and install mysql-client. 

```bash
$ kubectl exec -it -n demo ubuntu-867d4588d8-tl7hh -- bash                12:00
root@ubuntu-867d4588d8-tl7hh:/# apt update
... ... ..
root@ubuntu-867d4588d8-tl7hh:/# apt install mysql-client -y
Reading package lists... Done
... .. ...
root@ubuntu-867d4588d8-tl7hh:/#
```

Now let's try to connect with the ProxySQL server through the `proxy-server` service as the `test` user. 

```bash
root@ubuntu-867d4588d8-tl7hh:/# mysql -utest -ppass -hproxy-server.demo.svc -P6033
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1881
Server version: 8.0.27 (ProxySQL)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> 
```

We are successfully connected as the `test` user. Let's run some read/write query on this connection.

```bash
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
mysql> show tables;
+----------------+
| Tables_in_test |
+----------------+
| testtb         |
+----------------+
1 row in set (0.00 sec)

mysql> insert into testtb(name) values("Kim Torres");
Query OK, 1 row affected (0.01 sec)

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
| 3         | 3       |
| 4         | 0       |
+-----------+---------+
4 rows in set (0.003 sec)
```

We can see that the read-write split is successfully executed in the ProxySQL server. So the ProxySQL server is ready to use.

## Conclusion 

In this tutorial we have seen some very basic version of KubeDB ProxySQL. KubeDB provides many more for ProxySQL. In this site we have discussed on lot's of other features like `TLS Secured ProxySQL` , `Declarative Configuration` , `MariaDB and Percona-XtraDB Backend` , `Reconfigure` and much more. Checkout out other docs to learn more. 