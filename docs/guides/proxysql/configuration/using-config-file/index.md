---
title: Run ProxySQL with Custom Configuration File
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-configuration-usingconfigfile
    name: Config File
    parent: guides-proxysql-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom bootstrap configuration for ProxySQL. This tutorial will show you how to use KubeDB to run a ProxySQL server with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/proxysql/configuration/using-config-file/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

ProxySQL bootstraps itself from a `proxysql.cnf` configuration file built from a few tables: `mysql_users`, `mysql_query_rules`, `mysql_variables` and `admin_variables`. KubeDB lets you provide this bootstrap configuration declaratively through `spec.configuration.init.secretName`. You create a Secret containing one or more of the keys `AdminVariables.cnf`, `MySQLVariables.cnf`, `MySQLUsers.cnf`, `MySQLQueryRules.cnf`, and the operator patches their contents verbatim into `proxysql.cnf` during bootstrap. This configuration is applied only once, when ProxySQL is initialized.

> Note: `spec.configuration.init.inline` lets you provide the same kind of bootstrap configuration as structured YAML instead of raw `*.cnf` text, and always takes precedence over `secretName`. See the [ProxySQL concept page](/docs/guides/proxysql/concepts/proxysql/index.md#specconfigurationsecretname) for details.

In this tutorial, we will set [`mysql-max_connections`](https://proxysql.com/documentation/global-variables/mysql-variables/#mysql-max_connections) via a custom config file.

## Prepare MySQL Backend

ProxySQL needs a backend server to proxy. We will use a 3 node MySQL Group Replication cluster set up with KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.4.3"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mysqlgrp/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Wait for the MySQL cluster to be `Ready`.

```bash
$ kubectl get my -n demo
NAME           VERSION   STATUS   AGE
mysql-server   8.4.3     Ready    7m6s
```

> You can use MariaDB or Percona XtraDB as a backend as well. Have a look at the other [ProxySQL backend examples](/docs/guides/proxysql/backends/).

## Custom Configuration

At first, let's create a `MySQLVariables.cnf` file setting `mysql-max_connections`.

```bash
$ cat MySQLVariables.cnf
mysql_variables=
{
    max_connections=2048
}
```

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo proxysql-configuration --from-file=./MySQLVariables.cnf
secret/proxysql-configuration created
```

Now, create the ProxySQL crd specifying `spec.configuration.init.secretName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/configuration/using-config-file/examples/proxysql-custom.yaml
proxysql.kubedb.com/sample-proxysql created
```

Below is the YAML for the ProxySQL crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: sample-proxysql
  namespace: demo
spec:
  version: "2.7.3-debian"
  replicas: 1
  backend:
    name: mysql-server
  configuration:
    init:
      secretName: proxysql-configuration
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-proxysql-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                READY   STATUS    RESTARTS   AGE
sample-proxysql-0   1/1     Running   0          45s

$ kubectl get proxysql -n demo
NAME              VERSION        STATUS   AGE
sample-proxysql   2.7.3-debian   Ready    71s
```

We can see the ProxySQL is in `Ready` phase. Now, we will check if ProxySQL has bootstrapped with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-proxysql-0 -- bash
root@sample-proxysql-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt "ProxySQLAdmin > "
ProxySQLAdmin > select * from global_variables where variable_name='mysql-max_connections';
+------------------------+----------------+
| variable_name          | variable_value |
+------------------------+----------------+
| mysql-max_connections  | 2048           |
+------------------------+----------------+
1 row in set (0.00 sec)

ProxySQLAdmin > exit
Bye
root@sample-proxysql-0:/# exit
exit
```

We can see that the value of `mysql-max_connections` is the same as we provided in the custom config file.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete proxysql -n demo sample-proxysql
proxysql.kubedb.com "sample-proxysql" deleted
$ kubectl delete mysql -n demo mysql-server
mysql.kubedb.com "mysql-server" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
