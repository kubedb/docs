---
title: Reconfigure MongoDB Replicaset
menu:
  docs_{{ .version }}:
    identifier: mg-reconfigure-replicaset
    name: Replicaset
    parent: mg-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfigure MongoDB Replicaset Database

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a MongoDB Replicaset.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.


- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [ReplicaSet](/docs/guides/mongodb/clustering/replicaset.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/mongodb/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `MongoDB` Replicaset using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBOpsRequest` to reconfigure its configuration.

### Prepare MongoDB Replicaset

Now, we are going to deploy a `MongoDB` Replicaset database with version `4.2.3`.

### Deploy MongoDB 

At first, we will create `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 10000
```
Here, `maxIncomingConnections` is set to `10000`, whereas the default value is `65536`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo mg-custom-config --from-file=./mongod.conf
secret/mg-custom-config created
```

In this section, we are going to create a MongoDB object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "4.2.3"
  replicas: 3
  replicaSet:
    name: rs0
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  configSecret:
    name: mg-custom-config
```

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mg-replicaset-config.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo                                                                                                                                            
NAME            VERSION   STATUS   AGE
mg-replicaset   4.2.3     Ready    19m
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a mongodb instance,
```bash
$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\username}' | base64 -d                                                                       
root

$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         
nrKuxni0wDSMrgwy
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the configuration we have provided.

```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet                        
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--ipv6",
		"--bind_ip_all",
		"--port=27017",
		"--tlsMode=disabled",
		"--replSet=rs0",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
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
		"replication" : {
			"replSet" : "rs0"
		},
		"security" : {
			"authorization" : "enabled",
			"clusterAuthMode" : "keyFile",
			"keyFile" : "/data/configdb/key.txt"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1614668500, 1),
		"signature" : {
			"hash" : BinData(0,"7sh886HhsNYajGxYGp5Jxi52IzA="),
			"keyId" : NumberLong("6934943333319966722")
		}
	},
	"operationTime" : Timestamp(1614668500, 1)
}
```

As we can see from the configuration of ready mongodb, the value of `maxIncomingConnections` has been set to `10000`.

### Reconfigure using new config secret

Now we will reconfigure this database to set `maxIncomingConnections` to `20000`. 

Now, we will edit the `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 20000
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-custom-config --from-file=./mongod.conf
secret/new-custom-config created
```

#### Create MongoDBOpsRequest

Now, we will use this secret to replace the previous secret using a `MongoDBOpsRequest` CR. The `MongoDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfigure-replicaset
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-replicaset
  configuration:
    replicaSet:
      configSecret:
        name: new-custom-config
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-replicaset` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.customConfig.replicaSet.configSecret.name` specifies the name of the new secret.
- `spec.customConfig.arbiter.configSecret.name` could also be specified with a config-secret.
- Have a look [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-replicaset created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `MongoDB` object.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                          TYPE          STATUS       AGE
mops-reconfigure-replicaset   Reconfigure   Successful   113s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-replicaset 
Name:         mops-reconfigure-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T07:04:31Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:configuration:
          .:
          f:replicaSet:
            .:
            f:configSecret:
              .:
              f:name:
        f:databaseRef:
          .:
          f:name:
        f:readinessCriteria:
          .:
          f:objectsCountDiffPercentage:
          f:oplogMaxLagSeconds:
        f:timeout:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-02T07:04:31Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:configuration:
          f:replicaSet:
            f:podTemplate:
              .:
              f:controller:
              f:metadata:
              f:spec:
                .:
                f:resources:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-02T07:04:31Z
  Resource Version:  29869
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-reconfigure-replicaset
  UID:               064733d6-19db-4153-82f7-bc0580116ee6
Spec:
  Apply: IfReady
  Configuration:
    Replica Set:
      Config Secret:
        Name:  new-custom-config
  Database Ref:
    Name:  mg-replicaset
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2021-03-02T07:04:31Z
    Message:               MongoDB ops request is reconfiguring database
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2021-03-02T07:06:21Z
    Message:               Successfully Reconfigured MongoDB
    Observed Generation:   1
    Reason:                ReconfigureReplicaset
    Status:                True
    Type:                  ReconfigureReplicaset
    Last Transition Time:  2021-03-02T07:06:21Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                        Message
  ----    ------                 ----   ----                        -------
  Normal  PauseDatabase          2m55s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-replicaset
  Normal  PauseDatabase          2m55s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-replicaset
  Normal  ReconfigureReplicaset  65s    KubeDB Ops-manager operator  Successfully Reconfigured MongoDB
  Normal  ResumeDatabase         65s    KubeDB Ops-manager operator  Resuming MongoDB demo/mg-replicaset
  Normal  ResumeDatabase         65s    KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-replicaset
  Normal  Successful             65s    KubeDB Ops-manager operator  Successfully Reconfigured Database
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--ipv6",
		"--bind_ip_all",
		"--port=27017",
		"--tlsMode=disabled",
		"--replSet=rs0",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "*",
			"ipv6" : true,
			"maxIncomingConnections" : 20000,
			"port" : 27017,
			"tls" : {
				"mode" : "disabled"
			}
		},
		"replication" : {
			"replSet" : "rs0"
		},
		"security" : {
			"authorization" : "enabled",
			"clusterAuthMode" : "keyFile",
			"keyFile" : "/data/configdb/key.txt"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1614668887, 1),
		"signature" : {
			"hash" : BinData(0,"5q35Y51+YpbVHFKoaU7lUWi38oY="),
			"keyId" : NumberLong("6934943333319966722")
		}
	},
	"operationTime" : Timestamp(1614668887, 1)
}
```

As we can see from the configuration of ready mongodb, the value of `maxIncomingConnections` has been changed from `10000` to `20000`. So the reconfiguration of the database is successful.


### Reconfigure using inline config

Now we will reconfigure this database again to set `maxIncomingConnections` to `30000`. This time we won't use a new secret. We will use the `inlineConfig` field of the `MongoDBOpsRequest`. This will merge the new config in the existing secret.

#### Create MongoDBOpsRequest

Now, we will use the new configuration in the `inlineConfig` field in the `MongoDBOpsRequest` CR. The `MongoDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfigure-inline-replicaset
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-replicaset
  configuration:
    replicaSet:
      inlineConfig: |
          net:
            maxIncomingConnections: 30000
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-inline-replicaset` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.replicaSet.inlineConfig` specifies the new configuration that will be merged in the existing secret.
- `spec.customConfig.arbiter.configSecret.name` could also be specified with a config-secret.
- Have a look [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-inline-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-inline-replicaset created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                               TYPE          STATUS       AGE
mops-reconfigure-inline-replicaset   Reconfigure   Successful   109s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-inline-replicaset
Name:         mops-reconfigure-inline-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T07:09:39Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:configuration:
          .:
          f:replicaSet:
            .:
            f:inlineConfig:
        f:databaseRef:
          .:
          f:name:
        f:readinessCriteria:
          .:
          f:objectsCountDiffPercentage:
          f:oplogMaxLagSeconds:
        f:timeout:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-02T07:09:39Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:configuration:
          f:replicaSet:
            f:podTemplate:
              .:
              f:controller:
              f:metadata:
              f:spec:
                .:
                f:resources:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-02T07:09:39Z
  Resource Version:  31005
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-reconfigure-inline-replicaset
  UID:               0137442b-1b04-43ed-8de7-ecd913b44065
Spec:
  Apply: IfReady
  Configuration:
    Replica Set:
      Inline Config:  net:
  maxIncomingConnections: 30000

  Database Ref:
    Name:  mg-replicaset
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2021-03-02T07:09:39Z
    Message:               MongoDB ops request is reconfiguring database
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2021-03-02T07:11:14Z
    Message:               Successfully Reconfigured MongoDB
    Observed Generation:   1
    Reason:                ReconfigureReplicaset
    Status:                True
    Type:                  ReconfigureReplicaset
    Last Transition Time:  2021-03-02T07:11:14Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                        Message
  ----    ------                 ----   ----                        -------
  Normal  PauseDatabase          9m20s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-replicaset
  Normal  PauseDatabase          9m20s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-replicaset
  Normal  ReconfigureReplicaset  7m45s  KubeDB Ops-manager operator  Successfully Reconfigured MongoDB
  Normal  ResumeDatabase         7m45s  KubeDB Ops-manager operator  Resuming MongoDB demo/mg-replicaset
  Normal  ResumeDatabase         7m45s  KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-replicaset
  Normal  Successful             7m45s  KubeDB Ops-manager operator  Successfully Reconfigured Database
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--ipv6",
		"--bind_ip_all",
		"--port=27017",
		"--tlsMode=disabled",
		"--replSet=rs0",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "*",
			"ipv6" : true,
			"maxIncomingConnections" : 30000,
			"port" : 27017,
			"tls" : {
				"mode" : "disabled"
			}
		},
		"replication" : {
			"replSet" : "rs0"
		},
		"security" : {
			"authorization" : "enabled",
			"clusterAuthMode" : "keyFile",
			"keyFile" : "/data/configdb/key.txt"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1614669580, 1),
		"signature" : {
			"hash" : BinData(0,"u/xTAa4aW/8bsRvBYPffwQCeTF0="),
			"keyId" : NumberLong("6934943333319966722")
		}
	},
	"operationTime" : Timestamp(1614669580, 1)
}
```

As we can see from the configuration of ready mongodb, the value of `maxIncomingConnections` has been changed from `20000` to `30000`. So the reconfiguration of the database using the `inlineConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-reconfigure-replicaset mops-reconfigure-inline-replicaset
```