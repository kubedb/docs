---
title: Run ProxySQL with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: proxysql-custom-configuration
    name: Using Custom Config File
    parent: proxysql-custom-config
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for ProxySQL. This tutorial will show you how to use KubeDB to run a ProxySQL with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [minikube](https://github.com/kubernetes/minikube).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```console
  $ kubectl create ns demo
  namespace/demo created

  $ kubectl get ns demo
  NAME    STATUS  AGE
  demo    Active  5s
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/proxysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/proxysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

ProxySQL allows to configure via configuration file. The default configuration for ProxySQL can be found in `/etc/proxysql.cnf` file. In our Docker image (for ProxySQL), we used the file `/etc/custom-config/custom-proxysql.cnf` as the custom configuration file. The procedure is such that if the configuration file exist, ProxySQL instance will use combined startup settings from both `/etc/proxysql.cnf` and `/etc/custom-config/custom-proxysql.cnf` files. This custom configuration will overwrite the existing default one. For example config file:

- [v2.0.4](https://github.com/sysown/proxysql/blob/v2.0.4/src/proxysql.cfg).

To know more about configuring ProxySQL see [configuration file](https://github.com/sysown/proxysql/wiki/Configuration-file) and [variables](https://github.com/sysown/proxysql/wiki/Global-variables).

At first, you have to create a config file with name `custom-proxysql.cnf` containing your desired configurations. Then you have to put this file into a [volume](https://kubernetes.io/docs/concepts/storage/volumes/). You have to specify this volume  in `.spec.configSource` section while creating ProxySQL object. KubeDB will mount this volume into `/etc/custom-config` directory of the ProxySQL Pod.

In this tutorial, we will configure [mysql-connect_timeout_server](https://github.com/sysown/proxysql/wiki/Global-variables#mysql-connect_timeout_server) via the `custom-proxysql.cnf` file. We will use configMap as volume source.

## Custom Configuration

At first, let's create `custom-proxysql.cnf` file setting `mysql-connect_timeout_server` parameter.

> Note: We recommend to include the line `interfaces="0.0.0.0:6033"` here in the `mysql_variables` block. Though without this line, ProxySQL will work fine but we recommend to include it.
> The important thing you should keep in mind here is that never change the credential for admin interface for current version of ProxySQL image. It must be `admin:admin` (<username>:<password>).

```console
cat <<EOF > custom-proxysql.cnf
mysql_variables=
{
  interfaces="0.0.0.0:6033"
	connect_timeout_server=20000
}
EOF

$ cat custom-proxysql.cnf
mysql_variables=
{
  interfaces="0.0.0.0:6033"
	connect_timeout_server=20000
}
```

Here, `connect_timeout_server` is set to 20 secondes in mili-second.

Now, create a configMap with this configuration file.

```console
 $ kubectl create configmap -n demo my-custom-config --from-file=./custom-proxysql.cnf
configmap/my-custom-config created
```

Verify the config map has the configuration file.

```yaml
$ kubectl get configmap -n demo my-custom-config -o yaml
apiVersion: v1
data:
  my-config.cnf: |
    mysql_variables=
    {
      interfaces="0.0.0.0:6033"
      connect_timeout_server=20000
    }
kind: ConfigMap
metadata:
  name: my-custom-config
  namespace: demo
  ...
```

> **Note:** For this tutorial there must be a MySQL object with name `my-group` (Group Replication supported) running in the `demo` namespace in the cluster. You can deploy one by following section [create MySQL object with Group Replication](/docs/guides/proxysql/quickstart/load-balance-mysql-group-replication.md#Create-MySQL-Object).

Now, create ProxySQL object specifying `.spec.configSource` field.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/custom-proxysql.yaml
proxysql.kubedb.com/custom-proxysql created
```

Below is the YAML for the ProxySQL object we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: ProxySQL
metadata:
  name: custom-proxysql
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
  configSource:
    configMap:
      name: my-custom-config
  updateStrategy:
    type: RollingUpdate
```

Now, wait a few minutes. KubeDB operator will create necessary statefulset, services, secret etc. If everything goes well, we will see that a Pod with the name `custom-proxysql-0` has been created.

Check that the StatefulSet's Pod is running

```console
$ kubectl get pod -n demo
NAME                READY     STATUS    RESTARTS   AGE
custom-proxysql-0   1/1       Running   0          44s
```

Check the Pod's log,

```console
$ kubectl logs -f -n demo custom-proxysql-0
...
2019/11/28 15:58:41 [entrypoint.sh] [INFO] Applying custom config using cmd 'proxysql -c /etc/custom-config/custom-proxysql.cnf --reload -f  &'
2019/11/28 15:58:41 [entrypoint.sh] [INFO] Configuring proxysql ...
2019/11/28 15:58:41 [configure-proxysql.sh] [] From configure-proxysql.sh
2019/11/28 15:58:41 [configure-proxysql.sh] [INFO] Provided peers are my-group-0.my-group-gvr.demo my-group-1.my-group-gvr.demo my-group-2.my-group-gvr.demo
2019/11/28 15:58:41 [configure-proxysql.sh] [INFO] Waiting for host my-group-0.my-group-gvr.demo to be online ...
2019-11-28 15:58:41 [INFO] Using config file /etc/custom-config/custom-proxysql.cnf
2019-11-28 15:58:41 [INFO] No SSL keys/certificates found in datadir (/). Generating new keys/certificates.
....
```

Now, we will check if the ProxySQL has started with the custom configuration we have provided.

```console
kubectl exec -it -n demo custom-proxysql-0 -- mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt="ProxySQL [Admin]> "
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 7
Server version: 5.5.30 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQL [Admin]> show global variables;
+-----------------------------------------------------+----------------------+
| Variable_name                                       | Value                |
+-----------------------------------------------------+----------------------+
| admin-admin_credentials                             | admin:admin          |
| admin-checksum_mysql_query_rules                    | true                 |
| admin-checksum_mysql_servers                        | true                 |
| admin-checksum_mysql_users                          | true                 |
...
| admin-mysql_ifaces                                  | 0.0.0.0:6032         |

| mysql-connect_timeout_server                        | 20000                |
...
| mysql-monitor_username                              | proxysql             |
...
+-----------------------------------------------------+----------------------+
146 rows in set (0.00 sec)

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console

$ kubectl delete proxysql -n demo custom-proxysql

$ kubectl delete configmap my-custom-config -n demo
$ rm ./custom-proxysql.cnf
$ kubectl delete my -n demo my-group

$ kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/uninstall.md).

## Next Steps

- Monitor ProxySQL with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/proxysql/monitoring/using-builtin-prometheus.md).
- Monitor ProxySQL with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/proxysql/monitoring/using-coreos-prometheus-operator.md).
- Use private Docker registry to deploy ProxySQL with KubeDB [here](/docs/guides/proxysql/private-registry/using-private-registry.md).
- Detail concepts of ProxySQL CRD [here](/docs/concepts/database-proxy/proxysql.md).
- Detail concepts of ProxySQLVersion CRD [here](/docs/concepts/catalog/proxysql.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
