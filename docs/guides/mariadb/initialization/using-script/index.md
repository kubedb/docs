---
title: Initialize MariaDB using Script
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-initialization-usingscript
    name: Using Script
    parent: guides-mariadb-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize MariaDB using Script

This tutorial will show you how to use KubeDB to initialize a MariaDB database with \*.sql, \*.sh and/or \*.sql.gz script.
In this tutorial we will use .sql script stored in GitHub repository [kubedb/mysql-init-scripts](https://github.com/kubedb/mysql-init-scripts).

> Note: The yaml files that are used in this tutorial are stored [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mariadb/initialization/using-script/example) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
namespace/demo created
```

## Prepare Initialization Scripts

MariaDB supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use `init.sql` script from [mysql-init-scripts](https://github.com/kubedb/mysql-init-scripts) git repository to create a TABLE `kubedb_table` in `mysql` database.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from `init.sql` file. Then, we will provide this ConfigMap as script source in `init.script` of MariaDB crd spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo md-init-script \
--from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/mysql-init-scripts/raw/master/init.sql)"
configmap/md-init-script created
```

## Create a MariaDB database with Init-Script

Below is the `MariaDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: md-init-script
  terminationPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/Initialization/demo-1.yaml
mysql.kubedb.com/mysql-init-script created
```

Here,

- `spec.init.script` specifies a script source used to initialize the database before database server starts. The scripts will be executed alphabatically. In this tutorial, a sample .sql script from the git repository `https://github.com/kubedb/mysql-init-scripts.git` is used to create a test database. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) instead of `ConfigMap`.  The \*.sql, \*sql.gz and/or \*.sh sripts that are stored inside the root folder will be executed alphabatically. The scripts inside child folders will be skipped.

KubeDB operator watches for `MariaDB` objects using Kubernetes api. When a `MariaDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MariaDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present. No MariaDB specific RBAC roles are required for [RBAC enabled clusters](/docs/setup/README.md#using-yaml).

```yaml
$ kubectl get mariadb -n demo sample-mariadb -oyaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  ...
  name: sample-mariadb
  namespace: demo
  ...
spec:
  authSecret:
    name: sample-mariadb-auth
  init:
    initialized: true
    script:
      configMap:
        name: md-init-script
  ...
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: WipeOut
  version: 10.5.23
status:
  ...
  phase: Ready
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created.

Now, we will connect to this database and check the data inserted by the initlization script.

```bash
# Connecting to the database
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 40
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> use mysql;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

# Showing the inserted `kubedb_table`
Database changed
MariaDB [mysql]> select * from  kubedb_table;
+----+-------+
| id | name  |
+----+-------+
|  1 | name1 |
|  2 | name2 |
|  3 | name3 |
+----+-------+
3 rows in set (0.001 sec)

MariaDB [mysql]> quit;
Bye

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
mariadb.kubedb.com "sample-mariadb" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
