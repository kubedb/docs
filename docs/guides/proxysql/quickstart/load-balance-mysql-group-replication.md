---
title: Load Balance MySQL Group Replication Using ProxySQL
menu:
  docs_{{ .version }}:
    identifier: load-balance-mysql-group-replication-using-proxysql
    name: Load Balance MySQL Group Replication Using ProxySQL
    parent: proxysql-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Load Balance MySQL Group Replication Using ProxySQL

ProxySQL supports load balancing for MySQL Group Replication. This guide will show you how you can load balance MySQL Group Replication using ProxySQL.

## Before You Begin

Before proceeding:

- Read [mysql group replication concept](/docs/guides/mysql/clustering/overview.md) to learn about MySQL Group Replication.
- You need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).
- You have to be familiar with the [ProxySQL](/docs/concepts/database-proxy/proxysql.md) CRD.

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored [here](https://github.com/kubedb/docs/tree/{{< param "info.subproject_version" >}}/docs/examples).

## Load Balance Using ProxySQL

This section will demonstrate how to load balance a MySQL Group Replication (both the single-primary mode and multi-primary mode) using ProxySQL. Since KubeDB currently supports group replication only for the single-primary mode, we we are going to deploy a single-primary MySQL replication group using KubeDB. Then, we are going to load balance the read-write query requests to the MySQL database using ProxySQL.

### Deploy Sample MySQL Group Replication

Let's deploy a sample MySQL Group Replication and insert some data into it.

#### Create MySQL Object

Below is the YAML of a sample `MySQL` object with group replication that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-group
  namespace: demo
spec:
  version: "5.7.25"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
      baseServerID: 100
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Create the above `MySQL` object,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/demo-my-group.yaml
mysql.kubedb.com/my-group created
```

KubeDB will deploy a single primary MySQL Group Replication according to the above specification. It will also create the necessary Secrets and Services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get my -n demo my-group
NAME       VERSION   STATUS    AGE
my-group   5.7.25    Running   5m37s
```

The database is `Running`. Verify that KubeDB has created necessary Secret for this database using the following commands,

```console
$ kubectl get secret -n demo -l=kubedb.com/name=my-group
NAME            TYPE     DATA   AGE
my-group-auth   Opaque   2      10m
```

Here, we have to use the secret `my-group-auth` to connect with the database.

#### Insert Sample Data

Now, we are going to exec into the database pod and create some sample data. At first, find out the database Pod using the following command,

```console
$ kubectl get pods -n demo --selector="kubedb.com/name=my-group"
NAME         READY   STATUS    RESTARTS   AGE
my-group-0   1/1     Running   0          12m
my-group-1   1/1     Running   0          12m
my-group-2   1/1     Running   0          12m
```

Copy the username and password of the `root` user to access into `mysql` shell.

```console
$ kubectl get secret -n demo  my-group-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  my-group-auth -o jsonpath='{.data.password}'| base64 -d
tiIKEbjwnKLxAJP9⏎
```

Now, let's exec into the Pod to enter into `mysql` shell and create a database and a table,

```console
$ kubectl exec -it -n demo my-group-0 -- mysql --user=root --password=tiIKEbjwnKLxAJP9
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 25395
Server version: 5.7.25-log MySQL Community Server (GPL)

Copyright (c) 2000, 2019, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE playground;
Query OK, 1 row affected (0.04 sec)

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.08 sec)

mysql> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.00 sec)

mysql> INSERT INTO playground.equipment (type, quant, color) VALUES ("slide", 2, "blue");
Query OK, 1 row affected (0.02 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

Now, we are ready to see how ProxySQL can load balance the MySQL requests and route according to the defined rules.

### Prepare ProxySQL

KubeDB has a CRD named `ProxySQL` that can identify the reads, writes and route the write traffic to master and read traffic between the available slaves. We just need to create a ProxySQL object pointing to the backend database object.

#### Create ProxySQL Object

Let's create a ProxySQL object called `proxy-my-group` pointing to the previously created MySQL database. Below is the YAML of the ProxySQL object we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: ProxySQL
metadata:
  name: proxy-my-group
  namespace: demo
spec:
  version: "2.0.4"
  replicas: 1
  mode: GroupReplication
  backend:
    ref:
      apiGroup: "kubedb.com"
      kind: MySQL
      name: my-group
    replicas: 3
  updateStrategy:
    type: RollingUpdate
```

Now, create this,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/demo-proxy-my-group.yaml
proxysql.kubedb.com/proxy-my-group created
```

#### Verify ProxySQL Creation

KubeDB will deploy one replica for proxysql pointing to backend MySQL servers as we specified `.spec.backend.ref` as the previously created MySQL object named `my-group`. KubeDB will also create a Secret and a Service for it.

Let's check if the ProxySQL object is ready to use,

```console
$ kubectl get proxysql -n demo proxy-my-group
NAME             VERSION   STATUS    AGE
proxy-my-group   2.0.4     Running   129m
```

The status is `Running`. Verify that KubeDB has created necessary Secret and Service for this object using the following commands,

```console
$ kubectl get secret -n demo -l=proxysql.kubedb.com/name=proxy-my-group
NAME                  TYPE     DATA   AGE
proxy-my-group-auth   Opaque   2      132m

$ kubectl get service -n demo -l=proxysql.kubedb.com/name=proxy-my-group
NAME             TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
proxy-my-group   ClusterIP   10.32.6.128   <none>        6033/TCP   133m
```

Here, we have to use service `proxy-my-group` and secret `proxy-my-group-auth` to connect with the database through proxysql.

### Route Read/Write Requests through ProxySQL

So, KubeDB creates two different users (`proxysql` and `admin`) in the MySQL database for proxysql. The `admin` user is responsible for setting up admin configuration for proxysql and the `proxysql` user is for requesting query to the MySQL servers. KubeDB stores the credentials of `proxysql` user in the Secret `proxy-my-group-auth`, where the username and password of the `admin` user are fixed and they are `admin` and `admin` respectively.

```console
kubectl describe secret -n demo proxy-my-group-auth
Name:         proxy-my-group-auth
Namespace:    demo
Labels:       kubedb.com/kind=ProxySQL
              proxysql.kubedb.com/load-balance=GroupReplication
              proxysql.kubedb.com/name=proxy-my-group
Annotations:  <none>

Type:  Opaque

Data
====
proxysqlpass:  16 bytes
proxysqluser:  8 bytes
```

Copy the username and password of the `proxysql` user to access into `mysql` shell.

```console
$ kubectl get secret -n demo  proxy-my-group-auth -o jsonpath='{.data.proxysqluser}'| base64 -d
proxysql⏎

$ kubectl get secret -n demo  proxy-my-group-auth -o jsonpath='{.data.proxysqlpass}'| base64 -d
jxOlSObHgvvjOk1v⏎
```

#### Reads

Now, let's exec into the proxysql Pod to enter into `mysql` shell using `proxysql` user credentials and read the existing database and table,

```console
$ kubectl exec -it -n demo proxy-my-group-0 -- mysql --user=proxysql --password=jxOlSObHgvvjOk1v --host proxy-my-group.demo --port=6033 --prompt='MySQL [proxysql]> '
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 9
Server version: 5.5.30 (ProxySQL)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [proxysql]> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

MySQL [proxysql]> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.00 sec)

MySQL [proxysql]> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

MySQL [proxysql]> exit
Bye
```

Exec into the Pod to enter into `mysql` shell using `admin` user credentials and see the query counts,

```console
kubectl exec -it -n demo proxy-my-group-0 -- mysql --user=admin --password=admin --host 127.0.0.1 --port=6032 --prompt='MySQL [admin]> '
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 12
Server version: 5.5.30 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [admin]> SELECT command, Total_Time_us, Total_cnt FROM stats_mysql_commands_counters;
+-------------------+---------------+-----------+
| Command           | Total_Time_us | Total_cnt |
+-------------------+---------------+-----------+
| ALTER_TABLE       | 0             | 0         |
| ALTER_VIEW        | 0             | 0         |
| ANALYZE_TABLE     | 0             | 0         |
| BEGIN             | 0             | 0         |
| CALL              | 0             | 0         |
| CHANGE_MASTER     | 0             | 0         |
| COMMIT            | 0             | 0         |
| CREATE_DATABASE   | 0             | 0         |
| CREATE_INDEX      | 0             | 0         |
| CREATE_TABLE      | 0             | 0         |
| CREATE_TEMPORARY  | 0             | 0         |
| CREATE_TRIGGER    | 0             | 0         |
| CREATE_USER       | 0             | 0         |
| CREATE_VIEW       | 0             | 0         |
| DEALLOCATE        | 0             | 0         |
| DELETE            | 0             | 0         |
| DESCRIBE          | 0             | 0         |
| DROP_DATABASE     | 0             | 0         |
| DROP_INDEX        | 0             | 0         |
| DROP_TABLE        | 0             | 0         |
| DROP_TRIGGER      | 0             | 0         |
| DROP_USER         | 0             | 0         |
| DROP_VIEW         | 0             | 0         |
| GRANT             | 0             | 0         |
| EXECUTE           | 0             | 0         |
| EXPLAIN           | 0             | 0         |
| FLUSH             | 0             | 0         |
| INSERT            | 0             | 0         |
| KILL              | 0             | 0         |
| LOAD              | 0             | 0         |
| LOCK_TABLE        | 0             | 0         |
| OPTIMIZE          | 0             | 0         |
| PREPARE           | 0             | 0         |
| PURGE             | 0             | 0         |
| RENAME_TABLE      | 0             | 0         |
| RESET_MASTER      | 0             | 0         |
| RESET_SLAVE       | 0             | 0         |
| REPLACE           | 0             | 0         |
| REVOKE            | 0             | 0         |
| ROLLBACK          | 0             | 0         |
| SAVEPOINT         | 0             | 0         |
| SELECT            | 3789          | 7         |
| SELECT_FOR_UPDATE | 0             | 0         |
| SET               | 0             | 0         |
| SHOW_TABLE_STATUS | 0             | 0         |
| START_TRANSACTION | 0             | 0         |
| TRUNCATE_TABLE    | 0             | 0         |
| UNLOCK_TABLES     | 0             | 0         |
| UPDATE            | 0             | 0         |
| USE               | 0             | 0         |
| SHOW              | 5964          | 2         |
| UNKNOWN           | 401           | 1         |
+-------------------+---------------+-----------+
52 rows in set (0.00 sec)

MySQL [admin]> exit
Bye
```

#### Writes

This time exec into the proxysql Pod to enter into `mysql` shell using `proxysql` user credentials to create new database and table,

```console
kubectl exec -it -n demo proxy-my-group-0 -- mysql --user=proxysql --password=jxOlSObHgvvjOk1v --host proxy-my-group.demo --port=6033 --prompt='MySQL [proxysql]> '
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 13
Server version: 5.5.30 (ProxySQL)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [proxysql]> CREATE DATABASE playground_new;
Query OK, 1 row affected (0.05 sec)

MySQL [proxysql]> CREATE TABLE playground_new.equipment_new ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.16 sec)

MySQL [proxysql]> INSERT INTO playground_new.equipment_new (type, quant, color) VALUES ("slide_new", 2, "blue");
Query OK, 1 row affected (0.04 sec)

MySQL [proxysql]> SELECT * FROM playground_new.equipment_new;
+----+-----------+-------+-------+
| id | type      | quant | color |
+----+-----------+-------+-------+
|  7 | slide_new |     2 | blue  |
+----+-----------+-------+-------+
1 row in set (0.00 sec)

MySQL [proxysql]> exit;
Bye
```

Again, exec into the Pod to enter into `mysql` shell using `admin` user credentials and see the query counts,

```console
kubectl exec -it -n demo proxy-my-group-0 -- mysql --user=admin --password=admin --host 127.0.0.1 --port=6032 --prompt='MySQL [admin]> '
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 14
Server version: 5.5.30 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [admin]> SELECT command, Total_Time_us, Total_cnt FROM stats_mysql_commands_counters;
+-------------------+---------------+-----------+
| Command           | Total_Time_us | Total_cnt |
+-------------------+---------------+-----------+
| ALTER_TABLE       | 0             | 0         |
| ALTER_VIEW        | 0             | 0         |
| ANALYZE_TABLE     | 0             | 0         |
| BEGIN             | 0             | 0         |
| CALL              | 0             | 0         |
| CHANGE_MASTER     | 0             | 0         |
| COMMIT            | 0             | 0         |
| CREATE_DATABASE   | 46214         | 1         |
| CREATE_INDEX      | 0             | 0         |
| CREATE_TABLE      | 158931        | 1         |
| CREATE_TEMPORARY  | 0             | 0         |
| CREATE_TRIGGER    | 0             | 0         |
| CREATE_USER       | 0             | 0         |
| CREATE_VIEW       | 0             | 0         |
| DEALLOCATE        | 0             | 0         |
| DELETE            | 0             | 0         |
| DESCRIBE          | 0             | 0         |
| DROP_DATABASE     | 0             | 0         |
| DROP_INDEX        | 0             | 0         |
| DROP_TABLE        | 0             | 0         |
| DROP_TRIGGER      | 0             | 0         |
| DROP_USER         | 0             | 0         |
| DROP_VIEW         | 0             | 0         |
| GRANT             | 0             | 0         |
| EXECUTE           | 0             | 0         |
| EXPLAIN           | 0             | 0         |
| FLUSH             | 0             | 0         |
| INSERT            | 34228         | 1         |
| KILL              | 0             | 0         |
| LOAD              | 0             | 0         |
| LOCK_TABLE        | 0             | 0         |
| OPTIMIZE          | 0             | 0         |
| PREPARE           | 0             | 0         |
| PURGE             | 0             | 0         |
| RENAME_TABLE      | 0             | 0         |
| RESET_MASTER      | 0             | 0         |
| RESET_SLAVE       | 0             | 0         |
| REPLACE           | 0             | 0         |
| REVOKE            | 0             | 0         |
| ROLLBACK          | 0             | 0         |
| SAVEPOINT         | 0             | 0         |
| SELECT            | 8388          | 9         |
| SELECT_FOR_UPDATE | 0             | 0         |
| SET               | 0             | 0         |
| SHOW_TABLE_STATUS | 0             | 0         |
| START_TRANSACTION | 0             | 0         |
| TRUNCATE_TABLE    | 0             | 0         |
| UNLOCK_TABLES     | 0             | 0         |
| UPDATE            | 0             | 0         |
| USE               | 0             | 0         |
| SHOW              | 6713          | 3         |
| UNKNOWN           | 401           | 1         |
+-------------------+---------------+-----------+
52 rows in set (0.00 sec)

MySQL [admin]> exit
Bye
```

## Cleanup

To clean up the Kubernetes resources created by this tutorial, run:

```console
$ kubectl delete proxysql -n demo proxy-my-group
$ kubectl delete my -n demo my-group
```

## Next Steps

- Monitor ProxySQL with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/proxysql/monitoring/using-builtin-prometheus.md).
- Monitor ProxySQL with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/proxysql/monitoring/using-coreos-prometheus-operator.md).
- Use private Docker registry to deploy ProxySQL with KubeDB [here](/docs/guides/proxysql/private-registry/using-private-registry.md).
- Use custom config file to configure ProxySQL [here](/docs/guides/proxysql/configuration/using-custom-config.md).
- Detail concepts of ProxySQL CRD [here](/docs/concepts/database-proxy/proxysql.md).
- Detail concepts of ProxySQLVersion CRD [here](/docs/concepts/catalog/proxysql.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
