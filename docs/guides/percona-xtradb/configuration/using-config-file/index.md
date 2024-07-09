---
title: Run PerconaXtraDB with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-configuration-usingconfigfile
    name: Config File
    parent: guides-perconaxtradb-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for PerconaXtraDB. This tutorial will show you how to use KubeDB to run a PerconaXtraDB database with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created

  $ kubectl get ns demo
  NAME    STATUS  AGE
  demo    Active  5s
  ```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/percona-xtradb/configuration/using-config-file/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

PerconaXtraDB allows to configure database via configuration file. The default configuration for PerconaXtraDB can be found in `/etc/my.cnf` file. KubeDB adds a new custom configuration directory `/etc/mysql/custom.conf.d` if it's enabled. When PerconaXtraDB starts, it will look for custom configuration file in `/etc/mysql/custom.conf.d` directory. If configuration file exist, PerconaXtraDB instance will use combined startup setting from both `/etc/my.cnf` and `*.cnf` files in `/etc/mysql/conf.d` and `/etc/mysql/custom.conf.d` directory. This custom configuration will overwrite the existing default one.

At first, you have to create a config file with `.cnf` extension with your desired configuration. Then you have to put this file into a [volume](https://kubernetes.io/docs/concepts/storage/volumes/). You have to specify this volume  in `spec.configSecret` section while creating PerconaXtraDB crd. KubeDB will mount this volume into `/etc/mysql/custom.conf.d` directory of the database pod.

In this tutorial, we will configure [max_connections](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_max_connections/) and [read_buffer_size](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_read_buffer_size) via a custom config file. We will use Secret as volume source.

## Custom Configuration

At first, let's create `px-config.cnf` file setting `max_connections` and `read_buffer_size` parameters.

```bash
cat <<EOF > px-config.cnf
[mysqld]
max_connections = 200
read_buffer_size = 1048576
EOF

$ cat px-config.cnf
[mysqld]
max_connections = 200
read_buffer_size = 1048576
```

Here, `read_buffer_size` is set to 1MB in bytes.

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo px-configuration --from-file=./px-config.cnf
secret/md-configuration created
```

Verify the Secret has the configuration file.

```bash
$ kubectl get secret -n demo px-configuration -o yaml
apiVersion: v1
stringData:
  px-config.cnf: |
    [mysqld]
    max_connections = 200
    read_buffer_size = 1048576
kind: Secret
metadata:
  name: px-configuration
  namespace: demo
  ...
```

Now, create PerconaXtraDB crd specifying `spec.configSecret` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/configuration/using-config-file/examples/px-custom.yaml
mysql.kubedb.com/custom-mysql created
```

Below is the YAML for the PerconaXtraDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  configSecret:
    name: px-configuration
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

Now, wait a few minutes. KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-pxc-0` has been created.

Check that the statefulset's pod is running

```bash
$ kubectl get pod -n demo
NAME           READY   STATUS    RESTARTS   AGE
sample-pxc-0   2/2     Running   0          75m
sample-pxc-1   2/2     Running   0          95m
sample-pxc-2   2/2     Running   0          95m

$ kubectl get perconaxtradb -n demo 
NAME             VERSION   STATUS   AGE
NAME         VERSION   STATUS   AGE
sample-pxc   8.0.26    Ready    96m
```

We can see the database is in ready phase so it can accept connection.

Now, we will check if the database has started with the custom configuration we have provided.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# Connecting to the database
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$  mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1390
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 200   |
+-----------------+-------+
1 row in set (0.01 sec)


# value of `read_buffer_size` is same as provided
mysql> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 1048576 |
+------------------+---------+
1 row in set (0.001 sec)

mysql> exit
Bye
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
perconaxtradb.kubedb.com "sample-pxc" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
