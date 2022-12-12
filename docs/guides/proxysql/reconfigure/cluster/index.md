---
title: Reconfigure ProxySQL Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-reconfigure-cluster
    name: Demo
    parent: guides-proxysql-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfigure ProxySQL Cluster Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a `ProxySQL` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)
  - [ProxySQL Cluster](/docs/guides/proxysql/clustering/proxysql-cluster)
  - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest)
  - [Reconfigure Overview](/docs/guides/proxysql/reconfigure/overview)

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's wait for the MySQL to be Ready. 

```bash
$ kubectl get mysql -n demo 
NAME           VERSION   STATUS   AGE
mysql-server   5.7.36    Ready    3m51s
```

### Prepare ProxySQL Cluster

Let's create a KubeDB ProxySQL cluster with the following yaml.

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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/sample-proxysql.yaml 
proxysql.kubedb.com/proxy-server created
```

Let's wait for the ProxySQL to be Ready.

```bash
$ kubectl get proxysql -ndemo               
NAME           VERSION        STATUS   AGE
proxy-server   2.3.2-debian   Ready    98s
```

## Reconfigure MYSQL USERS

With `KubeDB` `ProxySQL` ops-request you can reconfigure `mysql_users` table. You can `add` and `delete` any users in the table. Also you can `update` any information of any user that is present in the table. To reconfigure the `mysql_users` table, you need to set the `.spec.type` to `Reconfigure`, provide the KubeDB ProxySQL instance name under the `spec.proxyRef` section and provide the desired user infos under the `spec.configuration.mysqlUsers.users` section. Set the `.spec.configuration.mysqlUsers.reqType` to either `add`, `update` or `delete` based on the operation you want to do. Below there are some samples for corresponding request type.

### Create user in mysql database

Let's first create two users in the backend mysql server. 

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

mysql> create user `testA`@'%' identified by 'passA';
Query OK, 0 rows affected (0.00 sec)

mysql> create user `testB`@'%' identified by 'passB';
Query OK, 0 rows affected (0.01 sec)

mysql> create database test;
Query OK, 1 row affected (0.01 sec)

mysql> grant all privileges on test.* to 'testA'@'%';
Query OK, 0 rows affected (0.00 sec)

mysql> grant all privileges on test.* to 'testB'@'%';
Query OK, 0 rows affected (0.00 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.00 sec)

mysql> exit
Bye
```

### Check current mysql_users table in ProxySQL

Let's check the current mysql_users table in the proxysql server. Make sure that the spec.syncUsers field was not set to true when the proxysql was deployed. Otherwise it will fetch all the users from the mysql backend and we won't be able to see the effects of reconfigure users ops requests. 

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin > "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 71
Server version: 8.0.27 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin > select * from mysql_users;
Empty set (0.001 sec)
```

### Add Users

Let's add the testA and testB user to the proxysql server with the ops-request. Make sure you have created the users in the mysql backend. As we don't provide the password in the yaml, the KubeDB operator fetches them from the backend server. So if the user is not present in the backend server, our operator will not be able to fetch the passwords and the ops-request will be failed. 

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
    name: add-user
    namespace: demo
spec:
    type: Reconfigure  
    proxyRef:
      name: proxy-server
    configuration:
      mysqlUsers:
        users: 
        - username: testA
          active: 1
          default_hostgroup: 2  
        - username: testB
          active: 1
          default_hostgroup: 2
        reqType: add
```

Let's applly the yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-add-users.yaml
proxysqlopsrequest.ops.kubedb.com/add-user created
```

Let's wait for the ops-request to be Successful. 

```bash
$ kubectl get proxysqlopsrequest -n demo     
NAME       TYPE          STATUS       AGE
add-user   Reconfigure   Successful   20s
```

Now let's check the `mysql_users` table in the proxysql server.

```bash
ProxySQLAdmin > select username,password,active,default_hostgroup from mysql_users;
+----------+-------------------------------------------+--------+-------------------+
| username | password                                  | active | default_hostgroup |
+----------+-------------------------------------------+--------+-------------------+
| testA    | *1BB8830D52D091A226FB7990D996CBC20F913475 | 1      | 2                 |
| testB    | *AE9C3C2838160D2591B6B15FA281CE712ABE94F0 | 1      | 2                 |
+----------+-------------------------------------------+--------+-------------------+
2 rows in set (0.001 sec)
```

We can see that the users has been successfuly added to the `mysql_users` table.

### Update Users

We have successfuly added new users in the `mysql_users` table with proxysqlopsrequest in the last section. Now we will see how to update any user information with proxysqlopsrequest. 

Suppose we want to update the `active` status and the `default_hostgroup` for the users "testA" and "testB". We can create an ops-request like the following. As in the `mysql_users` table the `username` is the primary key, we should always provide the `username` in the information. To update just change the `.spec.reqType` to `"update"`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: update-user
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    mysqlUsers:
      users: 
      - username: testA
        active: 0
        default_hostgroup: 3
      - username: testB
        active: 1
        default_hostgroup: 3
      reqType: update
```

Let's apply the yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-update-users.yaml 
proxysqlopsrequest.ops.kubedb.com/update-user created
```

Now wait for the ops-request to be Successful.

```bash
$ kubectl get proxysqlopsrequest -n demo     
NAME          TYPE          STATUS       AGE
add-user      Reconfigure   Successful   2m36s
update-user   Reconfigure   Successful   6s
```

Let's check the `mysql_users` table from the admin interface. 

```bash
ProxySQLAdmin > select username,password,active,default_hostgroup from mysql_users;
+----------+-------------------------------------------+--------+-------------------+
| username | password                                  | active | default_hostgroup |
+----------+-------------------------------------------+--------+-------------------+
| testA    | *1BB8830D52D091A226FB7990D996CBC20F913475 | 0      | 3                 |
| testB    | *AE9C3C2838160D2591B6B15FA281CE712ABE94F0 | 1      | 3                 |
+----------+-------------------------------------------+--------+-------------------+
2 rows in set (0.000 sec)
```

From the above output we can see that the user information has been successfuly updated.

### Delete Users 

To delete user from the `mysql_users` table, all we need to do is just provide the usernames in the `spec.configuration.mysqlUsers.users` array and set the `spec.reqType` to delete. Let's have a look at the following yaml.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: delete-user
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    mysqlUsers:
      users: 
      - username: testA
      reqType: delete
```
Let's apply the yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-remove-users.yaml 
proxysqlopsrequest.ops.kubedb.com/delete-user created
```
Let's wait for the ops-request to be successful. 

```bash
$ kubectl get proxysqlopsrequest -n demo    
NAME          TYPE          STATUS       AGE
add-user      Reconfigure   Successful   5m29s
delete-user   Reconfigure   Successful   12s
update-user   Reconfigure   Successful   2m59s
```

Now check the `mysql_users` table in the proxysql server.

```bash
ProxySQLAdmin > select username,password,active,default_hostgroup from mysql_users;
+----------+-------------------------------------------+--------+-------------------+
| username | password                                  | active | default_hostgroup |
+----------+-------------------------------------------+--------+-------------------+
| testB    | *AE9C3C2838160D2591B6B15FA281CE712ABE94F0 | 1      | 3                 |
+----------+-------------------------------------------+--------+-------------------+
1 row in set (0.001 sec)
```

We can see that the user is successfuly deleted.

## Reconfigure MYSQL QUERY RULES

With `KubeDB` `ProxySQL` ops-request you can reconfigure `mysql_query_rules` table. You can `add` and `delete` any rules in the table. Also you can `update` any information of any rule that is present in the table. To reconfigure the `mysql_query_rules` table, you need to set the `.spec.type` to `Reconfigure`, provide the KubeDB ProxySQL instance name under the `spec.proxyRef` section and provide the desired user infos under the `spec.configuration.mysqlQueryRules.rules` section. Set the `.spec.configuration.mysqlQueryRules.reqType` to either `add`, `update` or `delete` based on the operation you want to do. Below there are some samples for corresponding request type. 

### Check current mysql_query_rules table in ProxySQL

Let's check the current `mysql_query_rules` table in the proxysql server.
We might see some of the rules are already present. It happens when no rules are set in the `.spec.initConfig` section while deploying the proxysql. The operator adds some of the default query rules so that the basic operations can be run through the proxysql server. 

```bash
ProxySQLAdmin > select rule_id,active,match_digest,destination_hostgroup,apply from mysql_query_ru
les;
+---------+--------+----------------------+-----------------------+-------+
| rule_id | active | match_digest         | destination_hostgroup | apply |
+---------+--------+----------------------+-----------------------+-------+
| 1       | 1      | ^SELECT.*FOR UPDATE$ | 2                     | 1     |
| 2       | 1      | ^SELECT              | 3                     | 1     |
| 3       | 1      | .*                   | 2                     | 1     |
+---------+--------+----------------------+-----------------------+-------+
3 rows in set (0.001 sec)
```

### Add Query Rules

Let's add a query rule to the `mysql_query_rules` table with the proxysqlopsrequest. We should create a yaml like the following.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: add-rule
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    mysqlQueryRules:
      rules: 
      - rule_id: 4
        active: 1
        match_digest: "^SELECT .* FOR DELETE$"
        destination_hostgroup: 2
        apply: 1
      reqType: add
```

Let's apply the ops-request yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-add-rules.yaml  
proxysqlopsrequest.ops.kubedb.com/add-rule created
```

Wait for the ops-request to be successful. 

```bash
$ kubectl get proxysqlopsrequest -n demo | grep rule
add-rule      Reconfigure   Successful   59s
```
Now let's check the mysql_query_rules table in the proxysql server.

```bash
ProxySQLAdmin > select rule_id,active,match_digest,destination_hostgroup,apply from mysql_query_rules;
+---------+--------+------------------------+-----------------------+-------+
| rule_id | active | match_digest           | destination_hostgroup | apply |
+---------+--------+------------------------+-----------------------+-------+
| 1       | 1      | ^SELECT.*FOR UPDATE$   | 2                     | 1     |
| 2       | 1      | ^SELECT                | 3                     | 1     |
| 3       | 1      | .*                     | 2                     | 1     |
| 4       | 1      | ^SELECT .* FOR DELETE$ | 2                     | 1     |
+---------+--------+------------------------+-----------------------+-------+
4 rows in set (0.001 sec)
```
We can see that the users has been successfuly added to the `mysql_query_rules` table.

### Update Query Rules

We have successfuly added new rule in the `mysql_query_rules` table with proxysqlopsrequest in the last section. Now we will see how to update any rules information with proxysqlopsrequest. 

Suppose we want to update the `active` status rule 4. We can create an ops-request like the following. As in the `mysql_query_rules` table the `rule_id` is the primary key, we should always provide the `rule_id` in the information. To update just change the `.spec.reqType` to update.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: update-rule
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    mysqlQueryRules:
      rules: 
      - rule_id: 4
        active: 0
      reqType: update
```
Let's apply the yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-update-rules.yaml 
proxysqlopsrequest.ops.kubedb.com/update-rule created
```
Now wait for the ops-request to be successful.

```bash
$ kubectl get proxysqlopsrequest -n demo | grep rule
add-rule      Reconfigure   Successful   3m10s
update-rule   Reconfigure   Successful   71s
```
Let's check the `mysql_query_rules` table from the admin interface. 

```bash
ProxySQLAdmin > select rule_id,active,match_digest,destination_hostgroup,apply from mysql_query_rules;
+---------+--------+------------------------+-----------------------+-------+
| rule_id | active | match_digest           | destination_hostgroup | apply |
+---------+--------+------------------------+-----------------------+-------+
| 1       | 1      | ^SELECT.*FOR UPDATE$   | 2                     | 1     |
| 2       | 1      | ^SELECT                | 3                     | 1     |
| 3       | 1      | .*                     | 2                     | 1     |
| 4       | 0      | ^SELECT .* FOR DELETE$ | 2                     | 1     |
+---------+--------+------------------------+-----------------------+-------+
4 rows in set (0.001 sec)
```
From the above output we can see that the rules information has been successfuly updated.

### Delete Query Rules

To delete rules from the `mysql_query_rules` table, all we need to do is just provide the `rule_id` in the `spec.configuration.mysqlQueryRules.rules` array and set the `.spec.reqType` to `"delete"`. Let's have a look at the below yaml.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: delete-rule
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    mysqlQueryRules:
      rules: 
      - rule_id: 4
      reqType: delete
```
Let's apply the yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-remove-rules.yaml
proxysqlopsrequest.ops.kubedb.com/delete-rule created
```
Let's wait for the ops-request to be Successful. 

```bash
$ kubectl get proxysqlopsrequest -n demo | grep rule
add-rule      Reconfigure   Successful   4m13s
delete-rule   Reconfigure   Successful   12s
update-rule   Reconfigure   Successful   2m14s
```

Now check the `mysql_query_rules` table in the proxysql server.

```bash
ProxySQLAdmin > select rule_id,active,match_digest,destination_hostgroup,apply from mysql_query_rules;
+---------+--------+----------------------+-----------------------+-------+
| rule_id | active | match_digest         | destination_hostgroup | apply |
+---------+--------+----------------------+-----------------------+-------+
| 1       | 1      | ^SELECT.*FOR UPDATE$ | 2                     | 1     |
| 2       | 1      | ^SELECT              | 3                     | 1     |
| 3       | 1      | .*                   | 2                     | 1     |
+---------+--------+----------------------+-----------------------+-------+
3 rows in set (0.001 sec)
```
We can see that the user is successfuly deleted.

## Reconfigure Global Variables

With `KubeDB` `ProxySQL` ops-request you can reconfigure mysql variables and admin variables. You can reconfigure almost all the global variables except `mysql-interfaces`, `mysql-monitor_username`, `mysql-monitor_password`, `mysql-ssl_p2s_cert`, `mysql-ssl_p2s_key`, `mysql-ssl_p2s_ca`, `admin-admin_credentials` and `admin-mysql_interface`. To reconfigure any variable, you need to set the `.spec.type` to Reconfigure, provide the KubeDB ProxySQL instance name under the `spec.proxyRef` section and provide the desired configuration under the `spec.configuration.adminVariables` and the `spec.cofiguration.mysqlVariables` section. Below there are some samples for corresponding request type.

Suppose we want to update 4 global variables. Among these 2 are admin variables : cluster_check_interval_ms and refresh_interval . The other 2 are mysql variables : max_stmts_per_connection and max_transaction_time.

Let's see the current status from the proxysql server.

```bash
ProxySQLAdmin > show global variables;
+----------------------------------------------------------------------+--------------------------------------+
| Variable_name                                                        | Value                                |
+----------------------------------------------------------------------+--------------------------------------+
| ...                                                                  | ...                                  |
| admin-cluster_check_interval_ms                                      | 200                                  |
| ...                                                                  | ...                                  |
| admin-refresh_interval                                               | 2000                                 |
| ...                                                                  | ...                                  |
| mysql-max_stmts_per_connection                                       | 20                                   |
| ...                                                                  | ...                                  |
| mysql-max_transaction_time                                           | 14400000                             |
| ...                                                                  | ...                                  |
+----------------------------------------------------------------------+--------------------------------------+
193 rows in set (0.001 sec)
```

To reconfigure these variables all we need to do is create a yaml like the following. Just mention the variable name and its desired value in a key-value style under corresponding variable type i.e `mysqlVariables` and `adminVariables`. 

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: reconfigure-vars
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    adminVariables:
      refresh_interval: 2055
      cluster_check_interval_ms: 205
    mysqlVariables:
      max_transaction_time: 1540000
      max_stmts_per_connection: 19
```

Let's apply the yaml.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure/cluster/examples/proxyops-recon-vars.yaml
proxysqlopsrequest.ops.kubedb.com/recofigure-vars created
```

Wait for the ops-request to be successful.

```bash
$ kubectl get proxysqlopsrequest -n demo | grep reco
reconfigure-vars   Reconfigure   Successful   30s
```

Now let's check the variables we wanted to reconfigure. 

```bash
ProxySQLAdmin > show global variables;
+----------------------------------------------------------------------+--------------------------------------+
| Variable_name                                                        | Value                                |
+----------------------------------------------------------------------+--------------------------------------+
| ...                                                                  | ...                                  |
| admin-cluster_check_interval_ms                                      | 205                                  |
| ...                                                                  | ...                                  |
| admin-refresh_interval                                               | 2055                                 |
| ...                                                                  | ...                                  |
| mysql-max_stmts_per_connection                                       | 19                                   |
| ...                                                                  | ...                                  |
| mysql-max_transaction_time                                           | 1540000.0                            |
| ...                                                                  | ...                                  |
+----------------------------------------------------------------------+--------------------------------------+
193 rows in set (0.001 sec)
```

From the above output we can see the variables has been successfuly updated with the desired value. 

### Clean-up
```bash
$ kubectl delete proxysql -n demo proxy-server
$ kubectl delete proxysqlopsrequest -n demo --all 
$ kubectl delete mysql -n demo mysql-server
$ kubectl delete ns demo 
```