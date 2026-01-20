---
title: Run MariaDB with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-configuration-usingconfigfile
    name: Config File
    parent: guides-mariadb-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for MariaDB. This tutorial will show you how to use KubeDB to run a MariaDB database with custom configuration.

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

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mariadb/configuration/using-config-file/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

MariaDB allows to configure database via configuration file. The default configuration for MariaDB can be found in `/etc/mysql/my.cnf` file. When MariaDB starts, it will look for custom configuration file in `/etc/mysql/conf.d` directory. If configuration file exist, MariaDB instance will use combined startup setting from both `/etc/mysql/my.cnf` and `*.cnf` files in `/etc/mysql/conf.d` directory. This custom configuration will overwrite the existing default one. To know more about configuring MariaDB see [here](https://mariadb.com/kb/en/configuring-mariadb-with-option-files/).

At first, you have to create a config file with `.cnf` extension with your desired configuration. Then you have to put this file into a [volume](https://kubernetes.io/docs/concepts/storage/volumes/). You have to specify this volume  in `spec.configuration` section while creating MariaDB crd. KubeDB will mount this volume into `/etc/mysql/conf.d` directory of the database pod.

In this tutorial, we will configure [max_connections](https://mariadb.com/docs/server/ha-and-performance/optimization-and-tuning/system-variables/server-system-variables#max_connections) and [read_buffer_size](https://mariadb.com/docs/server/ha-and-performance/optimization-and-tuning/system-variables/server-system-variables#read_buffer_size/) via a custom config file. We will use Secret as volume source.

## Custom Configuration

At first, let's create `md-config.cnf` file setting `max_connections` and `read_buffer_size` parameters.

```bash
cat <<EOF > md-config.cnf
[mysqld]
max_connections = 200
read_buffer_size = 1048576
EOF

$ cat md-config.cnf
[mysqld]
max_connections = 200
read_buffer_size = 1048576
```

Here, `read_buffer_size` is set to 1MB in bytes.

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo md-configuration --from-file=./md-config.cnf
secret/md-configuration created
```

Verify the Secret has the configuration file.

```yaml
$ kubectl get secret -n demo md-configuration -o yaml
apiVersion: v1
stringData:
  md-config.cnf: |
    [mysqld]
    max_connections = 200
    read_buffer_size = 1048576
kind: Secret
metadata:
  name: md-configuration
  namespace: demo
  ...
```

Now, create MariaDB crd specifying `spec.configuration` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/configuration/using-config-file/examples/md-custom.yaml
mysql.kubedb.com/custom-mysql created
```

Below is the YAML for the MariaDB crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  configuration:
    secretName: md-configuration
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

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-mariadb-0` has been created.

Check that the petset's pod is running

```bash
 $ kubectl get pod -n demo
NAME               READY   STATUS    RESTARTS   AGE
sample-mariadb-0   1/1     Running   0          21s

$ kubectl get mariadb -n demo 
NAME             VERSION   STATUS   AGE
sample-mariadb   10.5.23    Ready    71s
```

We can see the database is in ready phase so it can accept conncetion.

Now, we will check if the database has started with the custom configuration we have provided.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# Connceting to the database
 $ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 23
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
MariaDB [(none)]> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 200   |
+-----------------+-------+
1 row in set (0.001 sec)

# value of `read_buffer_size` is same as provided
MariaDB [(none)]> show variables like 'read_buffer_size';
+------------------+---------+
| Variable_name    | Value   |
+------------------+---------+
| read_buffer_size | 1048576 |
+------------------+---------+
1 row in set (0.001 sec)

MariaDB [(none)]> exit
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
