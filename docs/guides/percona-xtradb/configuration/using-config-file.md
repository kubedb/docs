---
title: Run Percona XtraDB with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: px-using-config-file-configuration
    name: Config File
    parent: px-configuration
    weight: 10
menu_name: docs_{{ .version }}
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
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/percona-xtradb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/percona-xtradb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

PerconaXtraDB allows to configure database via configuration file. The default configuration for PerconaXtraDB can be found in `/etc/my.cnf` file. When PerconaXtraDB starts, it looks for custom configuration files.

- For standalone server, it looks into `/etc/my.cnf.d/`and `/etc/percona-server.conf.d/` directories. Here, KubeDB uses the later one `/etc/percona-server.conf.d/` for custom configurations. If configuration file exists, the `mysqld` will use combined startup setting from both `/etc/my.cnf` and `*.cnf` files in `/etc/percona-server.conf.d/` directory. This custom configuration will overwrite the existing default one.
- For cluster, the `mysqld` process looks into `/etc/my.cnf.d/`, and `/etc/percona-xtradb-cluster.conf.d/` directories. Here, KubeDB uses the later one `/etc/percona-xtradb-cluster.conf.d/` for custom configurations. If any configuration file exists, the `mysqld` will use combined startup settings from both `/etc/my.cnf` and `*.cnf` files in `/etc/percona-xtradb-cluster.conf.d/` directory. This custom configuration will overwrite the existing default one.

At first, you have to create a config file with `.cnf` extension with your desired configuration. Then you have to put this file into a [volume](https://kubernetes.io/docs/concepts/storage/volumes/). You have to specify this volume  in `.spec.configSecret` section while creating PerconaXtraDB object. KubeDB will mount this volume into the directory (specified above) of the database Pod.

In this tutorial, we will configure [max_connections](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_max_connections) and [read_buffer_size](https://dev.mysql.com/doc/refman/5.7/en/server-system-variables.html#sysvar_read_buffer_size) via a custom config file for a standalone percona server. We will use Secret as volume source.

## Custom Configuration

At first, let's create `my-config.cnf` file setting `max_connections` and `read_buffer_size` parameters.

```bash
$ cat <<EOF > my-config.cnf
[mysqld]
max_connections = 200
read_buffer_size = 1048576
EOF

$ cat my-config.cnf
[mysqld]
max_connections = 200
read_buffer_size = 1048576
```

Here, `read_buffer_size` is set to 1MB in bytes.

Now, create a Secret with this configuration file.

```bash
 $ kubectl create secret generic -n demo my-custom-config --from-file=./my-config.cnf
secret/my-custom-config created
```

Verify the Secret has the configuration file.

```bash
$ kubectl get secret -n demo my-custom-config -o yaml
```

And the output is,

```yaml
apiVersion: v1
stringData:
  my-config.cnf: |
    [mysqld]
    max_connections = 200
    read_buffer_size = 1048576
kind: Secret
metadata:
  name: my-custom-config
  namespace: demo
  ...
```

Now, create PerconaXtraDB object specifying `.spec.configSecret` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/custom-config.yaml
perconaxtradb.kubedb.com/custom-px created
```

Below is the YAML for the PerconaXtraDB object we just created above.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: custom-px
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 1
  configSecret:
    name: my-custom-config
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: DoNotTerminate
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we will see that a Pod with the name `custom-px-0` has been created.

Check that the StatefulSet's Pod is running

```bash
$ kubectl get pod -n demo
NAME          READY     STATUS    RESTARTS   AGE
custom-px-0   1/1       Running   0          44s
```

Check the Pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo custom-px-0
...
2019-12-24T13:43:51.050366Z 0 [Note] mysqld: ready for connections.
Version: '5.7.26-29'  socket: '/var/lib/mysql/mysql.sock'  port: 3306  Percona Server (GPL), Release 29, Revision 11ad961
```

Once we see `[Note] mysqld: ready for connections.` in the log, the database is ready.

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl get secret -n demo  custom-px-auth -o jsonpath='{.data.password}'| base64 -d
5ujF0R5wnUh5_gDk⏎

$ kubectl exec -it -n demo custom-px-0 -- \
  mysql --user=root --password=5ujF0R5wnUh5_gDk -e "select * from  performance_schema.global_variables where VARIABLE_NAME='max_connections';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+-----------------+----------------+
| VARIABLE_NAME   | VARIABLE_VALUE |
+-----------------+----------------+
| max_connections | 200            |
+-----------------+----------------+

$ kubectl exec -it -n demo custom-px-0 -- \
  mysql --user=root --password=5ujF0R5wnUh5_gDk -e "select * from  performance_schema.global_variables where VARIABLE_NAME='read_buffer_size';"
mysql: [Warning] Using a password on the command line interface can be insecure.
+------------------+----------------+
| VARIABLE_NAME    | VARIABLE_VALUE |
+------------------+----------------+
| read_buffer_size | 1048576        |
+------------------+----------------+
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo px/custom-px -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo px/custom-px

$ kubectl delete -n demo secret my-custom-config
$ rm ./my-config.cnf

$ kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Initialize [PerconaXtraDB with Script](/docs/guides/percona-xtradb/initialization/using-script.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- How to use [custom rbac resource](/docs/guides/percona-xtradb/custom-rbac/using-custom-rbac.md) for PerconaXtraDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/backup/overview/index.md).
- Detail concepts of [PerconaXtraDB object](/docs/guides/percona-xtradb/concepts/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/guides/percona-xtradb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
