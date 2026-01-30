---
title: Run MongoDB with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: mg-using-config-file-configuration
    name: Config File
    parent: mg-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for MongoDB. This tutorial will show you how to use KubeDB to run a MongoDB database with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

MongoDB allows configuring database via configuration file. The default configuration file for MongoDB deployed by `KubeDB` can be found in `/data/configdb/mongod.conf`. When MongoDB starts, it will look for custom configuration file in `/configdb-readonly/mongod.conf`. If configuration file exist, this custom configuration will overwrite the existing default one.

> To learn available configuration option of MongoDB see [Configuration File Options](https://docs.mongodb.com/manual/reference/configuration-options/).

At first, you have to create a secret with your configuration file contents as the value of this key `mongod.conf`. Then, you have to specify the name of this secret in `spec.configuration.secretName` section while creating MongoDB crd. KubeDB will mount this secret into `/configdb-readonly/` directory of the database pod.

Here one important thing to note that, `spec.configuration.secretName` will be used for standard replicaset members & standalone mongodb only. If you want to configure a specific type of mongo nodes, you have to set the name in respective fields.
For example, to configure shard topology node, set `spec.shardTopology.<shard / configServer / mongos>.configuration.secretName` field.
Similarly, To configure arbiter node, set `spec.arbiter.configuration.secretName` field.

In this tutorial, we will configure [net.maxIncomingConnections](https://docs.mongodb.com/manual/reference/configuration-options/#net.maxIncomingConnections) (default value: 65536) via a custom config file.

## Custom Configuration

At first, create `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 10000
```

Here, `maxIncomingConnections` is set to `10000`, whereas the default value is 65536.

Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo mg-configuration --from-file=./mongod.conf
secret/mg-configuration created
```

Verify the secret has the configuration file.

```yaml
$  kubectl get secret -n demo mg-configuration -o yaml
apiVersion: v1
data:
  mongod.conf: bmV0OgogIG1heEluY29taW5nQ29ubmVjdGlvbnM6IDEwMDAwMA==
kind: Secret
metadata:
  creationTimestamp: "2021-02-09T12:59:50Z"
  name: mg-configuration
  namespace: demo
  resourceVersion: "52495"
  uid: 92ca4191-eb97-4274-980c-9430ab7cc5d1
type: Opaque

$ echo bmV0OgogIG1heEluY29taW5nQ29ubmVjdGlvbnM6IDEwMDAwMA== | base64 -d
net:
  maxIncomingConnections: 100000
```

Now, create MongoDB crd specifying `spec.configuration.secretName` field.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mgo-custom-config
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
  configuration:
    secretName: mg-configuration
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/configuration/replicaset.yaml
mongodb.kubedb.com/mgo-custom-config created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `mgo-custom-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo mgo-custom-config-0
NAME                  READY     STATUS    RESTARTS   AGE
mgo-custom-config-0   1/1       Running   0          1m
```

Now, we will check if the database has started with the custom configuration we have provided.

Now, you can connect to this database through [mongo-shell](https://docs.mongodb.com/v4.2/mongo/). In this tutorial, we are connecting to the MongoDB server from inside the pod.

```bash
$ kubectl get secrets -n demo mgo-custom-config-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mgo-custom-config-auth -o jsonpath='{.data.\password}' | base64 -d
ErialNojWParBFoP

$ kubectl exec -it mgo-custom-config-0 -n demo sh

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
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
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

As we can see from the configuration of running mongodb, the value of `maxIncomingConnections` has been set to 10000 successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mgo-custom-config -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mgo-custom-config

kubectl delete -n demo secret mg-configuration

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/stash/overview/index.md) MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
