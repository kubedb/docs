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

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mariadb/configuration/using-config-file/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

MariaDB allows to configure database via configuration file. The default configuration for MariaDB can be found in `/etc/mysql/my.cnf` file. When MariaDB starts, it will look for custom configuration file in `/etc/mysql/conf.d` directory. If configuration file exist, MariaDB instance will use combined startup setting from both `/etc/mysql/my.cnf` and `*.cnf` files in `/etc/mysql/conf.d` directory. This custom configuration will overwrite the existing default one. To know more about configuring MariaDB see [here](https://mariadb.com/kb/en/configuring-mariadb-with-option-files/).

At first, you have to create a config file with `.cnf` extension with your desired configuration. Then you have to put this file into a [volume](https://kubernetes.io/docs/concepts/storage/volumes/). You have to specify this volume  in `spec.configSecret` section while creating MariaDB crd. KubeDB will mount this volume into `/etc/mysql/conf.d` directory of the database pod.

In this tutorial, we will configure [max_connections](https://mariadb.com/docs/reference/mdb/system-variables/max_connections/) and [read_buffer_size](https://mariadb.com/docs/reference/mdb/system-variables/read_buffer_size/) via a custom config file. We will use configMap as volume source.

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

Now, create a configMap with this configuration file.

```bash
$ kubectl create configmap -n demo md-configuration --from-file=./md-config.cnf
configmap/md-configuration created
```

Verify the config map has the configuration file.

```yaml
$ kubectl get configmap -n demo md-configuration -o yaml
apiVersion: v1
data:
  md-config.cnf: |
    [mysqld]
    max_connections = 200
    read_buffer_size = 1048576
kind: ConfigMap
metadata:
  name: md-configuration
  namespace: demo
  ...
```

Now, create MariaDB crd specifying `spec.configSecret` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}//docs/guides/mariadb/configuration/using-config-file/examples/md-custom.yaml
mysql.kubedb.com/custom-mysql created
```

Below is the YAML for the MariaDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: custom-mysql
  namespace: demo
spec:
  version: "8.0.23"
  configSecret:
    name: my-configuration
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi

apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.8"
  configSecret:
    name: md-configuration
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

Now, wait a few minutes. KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we will see that a pod with the name `custom-mysql-0` has been created.

Check that the statefulset's pod is running

```bash
$ kubectl get pod -n demo
NAME             READY     STATUS    RESTARTS   AGE
custom-mysql-0   1/1       Running   0          44s
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo custom-mysql-0
2020-08-26T10:45:27.892022Z 1 [System] [MY-013576] [InnoDB] InnoDB initialization has started.
2020-08-26T10:45:28.353024Z 1 [System] [MY-013577] [InnoDB] InnoDB initialization has ended.
...
2020-08-26T10:45:28.465037Z 0 [System] [MY-011323] [Server] X Plugin ready for connections. Bind-address: '::' port: 33060
2020-08-26T10:45:28.573208Z 0 [Warning] [MY-010068] [Server] CA certificate ca.pem is self signed.
2020-08-26T10:45:28.573533Z 0 [System] [MY-013602] [Server] Channel mysql_main configured to support TLS. Encrypted connections are now supported for this channel.
...
2020-08-26T10:45:28.610887Z 0 [System] [MY-010931] [Server] /usr/sbin/mysqld: ready for connections. Version: '8.0.21'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MariaDB Community Server - GPL.
....
```

Once we see `[Note] /usr/sbin/mysqld: ready for connections.` in the log, the database is ready.

Now, we will check if the database has started with the custom configuration we have provided.

First, deploy [phpMyAdmin](https://hub.docker.com/r/phpmyadmin/phpmyadmin/) to connect with the MariaDB database we have just created.

```bash
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/quickstart/demo-1.yaml
deployment.extensions/myadmin created
service/myadmin created
```

Then, open your browser and go to the following URL: _http://{node-ip}:{myadmin-svc-nodeport}_. For kind cluster, you can get this URL by running the following command:

```bash
$ kubectl get svc -n demo myadmin -o json | jq '.spec.ports[].nodePort'
30942

$ kubectl get node -o json | jq '.items[].status.addresses[].address'
"172.18.0.3"
"kind-control-plane"
"172.18.0.4"
"kind-worker"
"172.18.0.2"
"kind-worker2"

# expected url will be:
url: http://172.18.0.4:30942
```

Now, let's connect to the database from the phpMyAdmin dashboard using the database pod IP and MariaDB user password.

```bash
$ kubectl get pods custom-mysql-0 -n demo -o yaml | grep IP
  hostIP: 10.0.2.15
  podIP: 172.17.0.6

$ kubectl get secrets -n demo custom-mysql-auth -o jsonpath='{.data.\user}' | base64 -d
root

$ kubectl get secrets -n demo custom-mysql-auth -o jsonpath='{.data.\password}' | base64 -d
MLO5_fPVKcqPiEu9
```

Once, you have connected to the database with phpMyAdmin go to **Variables** tab and search for `max_connections` and `read_buffer_size`. Here are some screenshot showing those configured variables.
![max_connections](/docs/images/mysql/max_connection.png)

![read_buffer_size](/docs/images/mysql/read_buffer_size.png)

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo my/custom-mysql -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo my/custom-mysql

kubectl delete deployment -n demo myadmin
kubectl delete service -n demo myadmin

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart MariaDB](/docs/guides/mysql/quickstart/quickstart.md) with KubeDB Operator.
- Initialize [MariaDB with Script](/docs/guides/mysql/initialization/using-script.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/using-prometheus-operator.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/using-private-registry.md) to deploy MariaDB with KubeDB.
- Use [kubedb cli](/docs/guides/mysql/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MariaDB object](/docs/guides/mysql/concepts/mysql.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
