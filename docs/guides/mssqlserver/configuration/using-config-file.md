---
title: Run MSSQLServer with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: ms-configuration--config-file
    name: Config File
    parent: ms-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for MSSQLServer. This tutorial will show you how to use KubeDB to run a MSSQLServer database with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mssqlserver](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mssqlserver) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

MSSQLServer allows configuring database via configuration file. The default configuration file for MSSQLServer deployed by `KubeDB` can be found in `/var/opt/mssql/mssql.conf`. When MSSQLServer starts, it will look for  configuration file in `/var/opt/mssql/mssql.conf`. If configuration file exist, this configuration will overwrite the existing defaults.

> To learn available configuration option of MSSQLServer see [Configure SQL Server on Linux](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-mssql-conf?view=sql-server-ver16).

At first, you have to create a config file named `mssql.conf` with your desired configuration. Then you have to create a [secret](https://kubernetes.io/docs/concepts/configuration/secret/) using this file. You have to specify this secret name in `spec.configSecret.name` section while creating MSSQLServer CR. 

KubeDB will create a secret named `{mssqlserver-name}-config` with configuration file contents as the value of the key `mssql.conf` and mount this secret into `/var/opt/mssql/` directory of the database pod. the secret named `{mssqlserver-name}-config` will contain your desired configurations with some default configurations.

In this tutorial, we will configure sql server via a custom config file.

## Custom Configuration

At first, create `mssql.conf` file containing required configuration settings.

```ini
$ cat mssql.conf

```

Here, `maxIncomingConnections` is set to `10000`, whereas the default value is 65536.

Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo ms-configuration --from-file=./mssql.conf
secret/ms-configuration created
```

Verify the secret has the configuration file.

```yaml
$  kubectl get secret -n demo ms-configuration -o yaml
apiVersion: v1
data:
  mssql.conf: bmV0OgogIG1heEluY29taW5nQ29ubmVjdGlvbnM6IDEwMDAwMA==
kind: Secret
metadata:
  creationTimestamp: "2021-02-09T12:59:50Z"
  name: ms-configuration
  namespace: demo
  resourceVersion: "52495"
  uid: 92ca4191-eb97-4274-980c-9430ab7cc5d1
type: Opaque

$ echo bmV0OgogIG1heEluY29taW5nQ29ubmVjdGlvbnM6IDEwMDAwMA== | base64 -d
net:
  maxIncomingConnections: 100000
```

Now, create MSSQLServer CR specifying `spec.configSecret` field.

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: mssql-custom-config
  namespace: demo
spec:
  version: "4.4.26"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  configSecret:
    name: ms-configuration
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/configuration/replicaset.yaml
mssqlserver.kubedb.com/mssql-custom-config created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `mssql-custom-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo mssql-custom-config-0
NAME                  READY     STATUS    RESTARTS   AGE
mssql-custom-config-0   1/1       Running   0          1m
```

Now, we will check if the database has started with the custom configuration we have provided.

Now, you can connect to this database through [mongo-shell](https://docs.mssqlserver.com/v4.2/mongo/). In this tutorial, we are connecting to the MSSQLServer server from inside the pod.

```bash
$ kubectl get secrets -n demo mssql-custom-config-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mssql-custom-config-auth -o jsonpath='{.data.\password}' | base64 -d
ErialNojWParBFoP

$ kubectl exec -it mssql-custom-config-0 -n demo sh

> mongo admin

> db.auth("root","ErialNojWParBFoP")
1

> db._adminCommand( {getCmdLineOpts: 1})
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--ipv6",
		"--bind_ip_all",
		"--port=27017",
		"--tlsMode=disabled",
		"--config=/data/configdb/mssql.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mssql.conf",
		"net" : {
			"bindIp" : "*",
			"ipv6" : true,
			"maxIncomingConnections" : 10000,
			"port" : 27017,
			"tls" : {
				"mode" : "disabled"
			}
		},
		"security" : {
			"authorization" : "enabled"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1
}

> exit
bye
```

As we can see from the configuration of running mssqlserver, the value of `maxIncomingConnections` has been set to 10000 successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo ms/mssql-custom-config -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo ms/mssql-custom-config

kubectl delete -n demo secret ms-configuration

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mssqlserver/backup/stash/overview/index.md) MSSQLServer databases using Stash.
- Initialize [MSSQLServer with Script](/docs/guides/mssqlserver/initialization/using-script.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mssqlserver/monitoring/using-builtin-prometheus.md).
- Use [kubedb cli](/docs/guides/mssqlserver/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Detail concepts of [MSSQLServerVersion object](/docs/guides/mssqlserver/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
