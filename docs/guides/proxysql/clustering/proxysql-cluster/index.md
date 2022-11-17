---
title: ProxySQL Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-clustering-cluster
    name: ProxySQL Cluster Guide
    parent: guides-proxysql-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# ProxySQL Cluster

This guide will show you how to use `KubeDB` Enterprise operator to set up a `ProxySQL` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)
  - [ProxySQL Cluster](/docs/guides/proxysql/clustering/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

### Prepare MySQL backend

We need a mysql backend for the proxysql server. So we are creating one with the below yaml.

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/clustering/proxysql-cluster/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's wait for the MySQL to be Ready. 

```bash
$ kubectl get mysql -n demo 
NAME           VERSION   STATUS   AGE
mysql-server   5.7.36    Ready    3m51s
```

Let's first create an user in the backend mysql server and a database to test test the proxy traffic . 

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

## Deploy ProxySQL Cluster

The following is an example `ProxySQL` object which creates a proxysql cluster with three members. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"  
  replicas: 3
  mode: GroupReplication
  backend:
    name: mysql-server
  terminationPolicy: WipeOut
```

To deploy a simple proxysql cluster all you need to do is just set the `.spec.replicas` field to a higher value than 2. 

Let's apply the yaml. 

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/clustering/proxysql-cluster/examples/sample-proxysql.yaml
proxysql.kubedb.com/proxysql-server created
```

Let's wait for the ProxySQL to be Ready. 

```bash
$ kubectl get proxysql -n demo
NAME           VERSION        STATUS   AGE
proxy-server   2.3.2-debian   Ready    4m
```

Let's see the pods 

```bash
$ kubectl get pods -n demo | grep proxy
proxy-server-0   1/1     Running   3          4m
proxy-server-1   1/1     Running   3          4m
proxy-server-2   1/1     Running   3          4m
```

We can see that three nodes are up now. 

## Check proxysql_servers table

Let's check the proxysql_servers table inside the ProxySQL pods.

```bash 
#first node
$ kubectl exec -it -n demo proxy-server-0 -- bash 
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin >"
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 316
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select * from runtime_proxysql_servers;
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
3 rows in set (0.000 sec)

ProxySQLAdmin >exit
Bye
```
```bash 
#second node
$ kubectl exec -it -n demo proxy-server-1 -- bash 
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin >"
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 316
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select * from runtime_proxysql_servers;
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
3 rows in set (0.000 sec)

ProxySQLAdmin >exit
Bye
```

```bash 
#third node
$ kubectl exec -it -n demo proxy-server-2 -- bash 
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin >"
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 316
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select * from runtime_proxysql_servers;
+---------------------------------------+------+--------+---------+
| hostname                              | port | weight | comment |
+---------------------------------------+------+--------+---------+
| proxy-server-2.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-1.proxy-server-pods.demo | 6032 | 1      |         |
| proxy-server-0.proxy-server-pods.demo | 6032 | 1      |         |
+---------------------------------------+------+--------+---------+
3 rows in set (0.000 sec)

ProxySQLAdmin >exit
Bye
```

From the above output we can see that the proxysql_servers tables has been successfuly set up. 

## Create test user in proxysql 

Let's insert the test user inside the proxysql server 

```bash
$ kubectl exec -it -n demo proxy-server-1 -- bash 
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin >"
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 316
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > insert into mysql_users(username,password,default_hostgroup) values('test','pass',2);
Query OK, 1 row affected (0.001 sec)

ProxySQLAdmin > LOAD MYSQL USERS TO RUNTIME;
Query OK, 0 rows affected (0.000 sec)

ProxySQLAdmin > SAVE MYSQL USERS TO DISK;
Query OK, 0 rows affected (0.009 sec)

```

## Check load balance

Now lets check the load balancing through the cluster. 

First we need to create a script to sent load over the ProxySQL. We will use the test user and the test table to check send the load.

```bash
$ kubectl exec -it -n demo proxy-server-1 -- bash
root@proxy-server-1:/# apt update
... ... ... 
root@proxy-server-1:/# apt install nano
... ... ... 
root@proxy-server-1:/# nano load.sh
# copy paste the load.sh file here
GNU nano 5.4                    load.sh                                                                      
#!/bin/bash

COUNTER=0

USER='test'
PROXYSQL_NAME='proxy-server'
NAMESPACE='demo'
PASS='pass'

VAR="x"

while [  $COUNTER -lt 100 ]; do
    let COUNTER=COUNTER+1
    VAR=a$VAR
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e 'select 1;' > /dev/null 2>&1
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e "INSERT INTO test.testtb(name) VALUES ('$VAR');" > /dev/null 2>&1
    mysql -u$USER -h$PROXYSQL_NAME.$NAMESPACE.svc -P6033 -p$PASS -e "select * from test.testtb;" > /dev/null 2>&1
    sleep 0.0001
done

root@proxy-server-1:/# chmod +x load.sh

root@proxy-server-1:/# ./load.sh
```

```bash
$ kubectl exec -it -n demo proxy-server-1 -- bash 
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin >"
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 316
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select hostname, Queries from stats_proxysql_servers_metrics;
+---------------------------------------+---------+
| hostname                              | Queries |
+---------------------------------------+---------+
| proxy-server-2.proxy-server-pods.demo | 122     |
| proxy-server-1.proxy-server-pods.demo | 94      |
| proxy-server-0.proxy-server-pods.demo | 101     |
+---------------------------------------+---------+
3 rows in set (0.000 sec)

ProxySQLAdmin > select hostgroup,srv_host,Queries from stats_mysql_connection_pool;
+-----------+-------------------------------+---------+
| hostgroup | srv_host                      | Queries |
+-----------+-------------------------------+---------+
| 2         | mysql-server.demo.svc         | 30      |
| 3         | mysql-server-standby.demo.svc | 100     |
| 3         | mysql-server.demo.svc         | 34      |
+-----------+-------------------------------+---------+
```

From the above output we can see that the loads are properly distributed over the proxysql servers and the backend mysqls. 

## Chekc cluster sync

Let's check if any configuration change is automatically propagated to other in out proxysql cluster. 

We will change the `admin-restapi_enabled` in one cluster and observe the change in others.

First check the current status. 

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash 
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "show variables like 'admin-restapi_enabled';"
+-----------------------+-------+
| Variable_name         | Value |
+-----------------------+-------+
| admin-restapi_enabled | false |
+-----------------------+-------+
root@proxy-server-0:/# exit
exit 

$ kubectl exec -it -n demo proxy-server-1 -- bash 
root@proxy-server-1:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "show variables like 'admin-restapi_enabled';"
+-----------------------+-------+
| Variable_name         | Value |
+-----------------------+-------+
| admin-restapi_enabled | false |
+-----------------------+-------+
root@proxy-server-1:/# exit
exit 

$ kubectl exec -it -n demo proxy-server-2 -- bash 
root@proxy-server-2:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "show variables like 'admin-restapi_enabled';"
+-----------------------+-------+
| Variable_name         | Value |
+-----------------------+-------+
| admin-restapi_enabled | false |
+-----------------------+-------+
root@proxy-server-2:/# exit
exit 

```

Now set the value to `true` in server 0 . 

```bash

$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "set admin-restapi_enabled='true';"
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "show variables like 'admin-restapi_enabled';"
+-----------------------+-------+
| Variable_name         | Value |
+-----------------------+-------+
| admin-restapi_enabled | true  |
+-----------------------+-------+
root@proxy-server-0:/# exit
exit 

$ kubectl exec -it -n demo proxy-server-1 -- bash
root@proxy-server-1:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "show variables like 'admin-restapi_enabled';"
+-----------------------+-------+
| Variable_name         | Value |
+-----------------------+-------+
| admin-restapi_enabled | true  |
+-----------------------+-------+
root@proxy-server-1:/# exit
exit 

$ kubectl exec -it -n demo proxy-server-2 -- bash
root@proxy-server-2:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 -e "show variables like 'admin-restapi_enabled';"
+-----------------------+-------+
| Variable_name         | Value |
+-----------------------+-------+
| admin-restapi_enabled | true  |
+-----------------------+-------+
root@proxy-server-2:/# exit
exit 

```
From the above output we can see that the cluster is always in sync and the configuration change is always propagated to other cluster nodes. 

## Cluster failover recovery 
In case of any pod crash for proxysql cluster, the statefulset which was created by KubeDb operator creates another pod and the is auto joins the cluster. We can delete a pod and wait for that to create again and join the cluster and test this feature. 

Let's see the current status first.

```bash
ProxySQLAdmin > SELECT hostname, checksum, FROM_UNIXTIME(changed_at) changed_at, FROM_UNIXTIME(updated_at) updated_at FROM stats_proxysql_servers_checksums WHERE name='mysql_users' ORDER BY hostname;
+---------------------------------------+--------------------+---------------------+---------------------+
| hostname                              | checksum           | changed_at          | updated_at          |
+---------------------------------------+--------------------+---------------------+---------------------+
| proxy-server-0.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:09:49 | 2022-11-15 06:34:28 |
| proxy-server-1.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:22 | 2022-11-15 06:34:28 |
| proxy-server-2.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:17 | 2022-11-15 06:34:28 |
+---------------------------------------+--------------------+---------------------+---------------------+
3 rows in set (0.000 sec)
```

Now let's delete the pod-2. 

```bash
$ kubectl delete pod -n demo proxy-server-2
pod "proxy-server-2" deleted
```

Let's watch the cluster status now. 
```bash
ProxySQLAdmin > SELECT hostname, checksum, FROM_UNIXTIME(changed_at) changed_at, FROM_UNIXTIME(updated_at) updated_at FROM stats_proxysql_servers_checksums WHERE name='mysql_users' ORDER BY hostname;
+---------------------------------------+--------------------+---------------------+---------------------+
| hostname                              | checksum           | changed_at          | updated_at          |
+---------------------------------------+--------------------+---------------------+---------------------+
| proxy-server-0.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:09:49 | 2022-11-15 06:34:28 |
| proxy-server-1.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:22 | 2022-11-15 06:34:28 |
| proxy-server-2.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:17 | 2022-11-15 06:34:28 |
+---------------------------------------+--------------------+---------------------+---------------------+
3 rows in set (0.000 sec)
ProxySQLAdmin > SELECT hostname, checksum, FROM_UNIXTIME(changed_at) changed_at, FROM_UNIXTIME(updated_at) updated_at FROM stats_proxysql_servers_checksums WHERE name='mysql_users' ORDER BY hostname;
+---------------------------------------+--------------------+---------------------+---------------------+
| hostname                              | checksum           | changed_at          | updated_at          |
+---------------------------------------+--------------------+---------------------+---------------------+
| proxy-server-0.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:09:49 | 2022-11-15 06:34:28 |
| proxy-server-1.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:22 | 2022-11-15 06:34:28 |
| proxy-server-2.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:17 | 2022-11-15 06:34:28 |
+---------------------------------------+--------------------+---------------------+---------------------+
3 rows in set (0.000 sec)

... ... ...

ProxySQLAdmin > SELECT hostname, checksum, FROM_UNIXTIME(changed_at) changed_at, FROM_UNIXTIME(updated_at) updated_at FROM stats_proxysql_servers_checksums WHERE name='mysql_users' ORDER BY hostname;
+---------------------------------------+--------------------+---------------------+---------------------+
| hostname                              | checksum           | changed_at          | updated_at          |
+---------------------------------------+--------------------+---------------------+---------------------+
| proxy-server-0.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:09:49 | 2022-11-15 06:34:40 |
| proxy-server-1.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:22 | 2022-11-15 06:34:40 |
| proxy-server-2.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:17 | 2022-11-15 06:34:28 |
+---------------------------------------+--------------------+---------------------+---------------------+
3 rows in set (0.000 sec)

... ... ...

ProxySQLAdmin > SELECT hostname, checksum, FROM_UNIXTIME(changed_at) changed_at, FROM_UNIXTIME(updated_at) updated_at FROM stats_proxysql_servers_checksums WHERE name='mysql_users' ORDER BY hostname;
+---------------------------------------+--------------------+---------------------+---------------------+
| hostname                              | checksum           | changed_at          | updated_at          |
+---------------------------------------+--------------------+---------------------+---------------------+
| proxy-server-0.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:09:49 | 2022-11-15 06:34:40 |
| proxy-server-1.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:22 | 2022-11-15 06:34:40 |
| proxy-server-2.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:17 | 2022-11-15 06:34:28 |
+---------------------------------------+--------------------+---------------------+---------------------+
3 rows in set (0.000 sec)
```

From the above output we can see that the third server is out of sync as it is not available right now. But the other two are in sync. 

Wait for the new pod come up. 

```bash
$ kubectl get pod -n demo proxy-server-2
NAME             READY   STATUS    RESTARTS   AGE
proxy-server-2   1/1     Running   0          94s
```

Now check the status again. 

```
ProxySQLAdmin > SELECT hostname, checksum, FROM_UNIXTIME(changed_at) changed_at, FROM_UNIXTIME(updated_at) updated_at FROM stats_proxysql_servers_checksums WHERE name='mysql_users' ORDER BY hostname;
+---------------------------------------+--------------------+---------------------+---------------------+
| hostname                              | checksum           | changed_at          | updated_at          |
+---------------------------------------+--------------------+---------------------+---------------------+
| proxy-server-0.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:09:49 | 2022-11-15 06:34:50 |
| proxy-server-1.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:22 | 2022-11-15 06:34:50 |
| proxy-server-2.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:17 | 2022-11-15 06:34:28 |
+---------------------------------------+--------------------+---------------------+---------------------+
3 rows in set (0.000 sec)

... ... ...

ProxySQLAdmin > SELECT hostname, checksum, FROM_UNIXTIME(changed_at) changed_at, FROM_UNIXTIME(updated_at) updated_at FROM stats_proxysql_servers_checksums WHERE name='mysql_users' ORDER BY hostname;
+---------------------------------------+--------------------+---------------------+---------------------+
| hostname                              | checksum           | changed_at          | updated_at          |
+---------------------------------------+--------------------+---------------------+---------------------+
| proxy-server-0.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:09:49 | 2022-11-15 06:35:15 |
| proxy-server-1.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:22 | 2022-11-15 06:35:15 |
| proxy-server-2.proxy-server-pods.demo | 0x49728B20D3BC91AC | 2022-11-15 06:10:17 | 2022-11-15 06:35:15 |
+---------------------------------------+--------------------+---------------------+---------------------+
3 rows in set (0.000 sec)
```

From the above output we can see that the new pod is now in sync with the two others. So the failover recovery is successful. 

## Cleaning up 

```bash
$ kubectl delete proxysql -n demo proxy-server
proxysql.kubedb.com "proxy-server" deleted
$ kubectl delete mysql -n demo mysql-server
mysql.kubedb.com "mysql-server" deleted
```