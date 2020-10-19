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

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfigure MongoDB Replicaset Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a MongoDB Replicaset.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [ReplicaSet](/docs/guides/mongodb/clustering/replicaset.md)
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Reconfigure Overview](/docs/guides/mongodb/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `MongoDB` Replicaset using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBOpsRequest` to reconfigure its configuration.

### Prepare MongoDB Replicaset

Now, we are going to deploy a `MongoDB` Replicaset database with version `3.6.8`.

### Deploy MongoDB 

At first, we will create `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 10000
```
Here, `maxIncomingConnections` is set to `10000`, whereas the default value is `65536`.

Now, we will create a configMap with this configuration file.

```console
$ kubectl create configmap -n demo mg-custom-config --from-file=./mongod.conf
configmap/mg-custom-config created
```

In this section, we are going to create a MongoDB object specifying `spec.configSource` field to apply this custom configuration. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml

```

Let's create the `MongoDB` CR we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mg-replicaset-config.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` has status `Running`. i.e,

```console
$ kubectl get mg -n demo                                                                                                                                             20:05:47

```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a mongodb instance,
```console
$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\username}' | base64 -d                                                                         11:09:51
root

$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         11:10:44
nrKuxni0wDSMrgwy
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the configuration we have provided.

```console
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet                          18:35:59

```

As we can see from the configuration of running mongodb, the value of `maxIncomingConnections` has been set to `10000`.

### Reconfigure using new ConfigMap

Now we will reconfigure this database to set `maxIncomingConnections` to `20000`. 

Now, we will edit the `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 20000
```

Then, we will create a new configMap with this configuration file.

```console
$ kubectl create configmap -n demo new-custom-config --from-file=./mongod.conf
configmap/mg-custom-config created
```

#### Create MongoDBOpsRequest

Now, we will use this configMap to replace the previous configMap using a `MongoDBOpsRequest` CR. The `MongoDBOpsRequest` yaml is given below,

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
  customConfig:
    replicaSet:
      configMap:
        name: new-custom-config
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-replicaset` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.customConfig.replicaSet.configMap.name` specifies the name of the new configmap.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-replicaset created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Enterprise operator will update the `configSource` of `MongoDB` object.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                          TYPE          STATUS       AGE
mops-reconfigure-replicaset   Reconfigure   Successful   113s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-replicaset 
 Name:         mops-reconfigure-replicaset
 Namespace:    demo
 Labels:       <none>
 Annotations:  API Version:  ops.kubedb.com/v1alpha1
 Kind:         MongoDBOpsRequest
 Metadata:
   Creation Timestamp:  2020-08-26T18:40:06Z
   Finalizers:
     kubedb.com
   Generation:  1
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
         f:customConfig:
           .:
           f:replicaSet:
             .:
             f:configMap:
               .:
               f:name:
         f:databaseRef:
           .:
           f:name:
         f:type:
     Manager:      kubectl
     Operation:    Update
     Time:         2020-08-26T18:40:06Z
     API Version:  ops.kubedb.com/v1alpha1
     Fields Type:  FieldsV1
     fieldsV1:
       f:metadata:
         f:finalizers:
       f:status:
         .:
         f:conditions:
         f:observedGeneration:
         f:phase:
     Manager:         kubedb-enterprise
     Operation:       Update
     Time:            2020-08-26T18:41:26Z
   Resource Version:  6294401
   Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-reconfigure-replicaset
   UID:               bd17e9ed-35ed-4ab9-a844-659800ef4f39
 Spec:
   Custom Config:
     ReplicaSet:
       Config Map:
         Name:  new-custom-config
   Database Ref:
     Name:  mg-replicaset
   Type:    Reconfigure
 Status:
   Conditions:
     Last Transition Time:  2020-08-26T18:40:06Z
     Message:               MongoDB ops request is being processed
     Observed Generation:   1
     Reason:                Scaling
     Status:                True
     Type:                  Scaling
     Last Transition Time:  2020-08-26T18:40:06Z
     Message:               Successfully paused mongodb: mg-replicaset
     Observed Generation:   1
     Reason:                PauseDatabase
     Status:                True
     Type:                  PauseDatabase
     Last Transition Time:  2020-08-26T18:41:26Z
     Message:               Successfully Reconfigured mongodb
     Observed Generation:   1
     Reason:                ReconfigureReplicaset
     Status:                True
     Type:                  ReconfigureReplicaset
     Last Transition Time:  2020-08-26T18:41:26Z
     Message:               Succefully Resumed mongodb: mg-replicaset
     Observed Generation:   1
     Reason:                ResumeDatabase
     Status:                True
     Type:                  ResumeDatabase
     Last Transition Time:  2020-08-26T18:41:26Z
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
   Normal  PauseDatabase          2m23s  KubeDB Enterprise Operator  Pausing Mongodb mg-replicaset in Namespace demo
   Normal  PauseDatabase          2m23s  KubeDB Enterprise Operator  Successfully Paused Mongodb mg-replicaset in Namespace demo
   Normal  ReconfigureReplicaset  63s    KubeDB Enterprise Operator  Successfully Reconfigured mongodb
   Normal  ResumeDatabase         63s    KubeDB Enterprise Operator  Resuming MongoDB
   Normal  ResumeDatabase         63s    KubeDB Enterprise Operator  Successfully Started Balancer
   Normal  Successful             63s    KubeDB Enterprise Operator  Successfully Reconfigured Database
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the new configuration we have provided.

```console
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--bind_ip=0.0.0.0",
		"--port=27017",
		"--sslMode=disabled",
		"--replSet=replicaset",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "0.0.0.0",
			"maxIncomingConnections" : 20000,
			"port" : 27017,
			"ssl" : {
				"mode" : "disabled"
			}
		},
		"replication" : {
			"replSet" : "replicaset"
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
	"operationTime" : Timestamp(1598467358, 1),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1598467358, 1),
		"signature" : {
			"hash" : BinData(0,"ZumeKUzV3YmFMlmHy9twU1tJxlw="),
			"keyId" : NumberLong("6865318254139670530")
		}
	}
}  
```

As we can see from the configuration of running mongodb, the value of `maxIncomingConnections` has been changed from `10000` to `20000`. So the reconfiguration of the database is successful.


### Reconfigure using new Data

Now we will reconfigure this database again to set `maxIncomingConnections` to `30000`. This time we won't use a new configMap. We will use the data field of the `MongoDBOpsRequest`. This will merge the new config in the existing configMap.

#### Create MongoDBOpsRequest

Now, we will use the new configuration in the `data` field in the `MongoDBOpsRequest` CR. The `MongoDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfigure-data-replicaset
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-replicaset
  customConfig:
    replicaSet:
      data:
        mongod.conf: |
          net:
            maxIncomingConnections: 30000
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-data-replicaset` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.customConfig.replicaSet.data` specifies the new configuration that will be merged in the existing configMap.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-data-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-data-replicaset created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Enterprise operator will merge this new config with the existing configuration.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                               TYPE          STATUS       AGE
mops-reconfigure-data-replicaset   Reconfigure   Successful   109s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-data-replicaset
Name:         mops-reconfigure-data-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2020-08-26T18:36:14Z
  Finalizers:
    kubedb.com
  Generation:  1
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
        f:customConfig:
          .:
          f:replicaSet:
            .:
            f:data:
              .:
              f:mongod.conf:
        f:databaseRef:
          .:
          f:name:
        f:type:
    Manager:      kubectl
    Operation:    Update
    Time:         2020-08-26T18:36:14Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:finalizers:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2020-08-26T18:37:55Z
  Resource Version:  6291551
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-reconfigure-data-replicaset
  UID:               0e3a72d8-c906-48cb-bc0c-e87671367454
Spec:
  Custom Config:
    ReplicaSet:
      Data:
        mongod.conf:  net:
  maxIncomingConnections: 30000

  Database Ref:
    Name:  mg-replicaset
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2020-08-26T18:36:14Z
    Message:               MongoDB ops request is being processed
    Observed Generation:   1
    Reason:                Scaling
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2020-08-26T18:37:55Z
    Message:               Successfully Reconfigured mongodb
    Observed Generation:   1
    Reason:                ReconfigureReplicaset
    Status:                True
    Type:                  ReconfigureReplicaset
    Last Transition Time:  2020-08-26T18:37:55Z
    Message:               Succefully Resumed mongodb: mg-replicaset
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-26T18:37:55Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age   From                        Message
  ----    ------                 ----  ----                        -------
  Normal  ReconfigureReplicaset  25s   KubeDB Enterprise Operator  Successfully Reconfigured mongodb
  Normal  ResumeDatabase         25s   KubeDB Enterprise Operator  Resuming MongoDB
  Normal  ResumeDatabase         25s   KubeDB Enterprise Operator  Successfully Started Balancer
  Normal  Successful             25s   KubeDB Enterprise Operator  Successfully Reconfigured Database
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the new configuration we have provided.

```console
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--bind_ip=0.0.0.0",
		"--port=27017",
		"--sslMode=disabled",
		"--replSet=replicaset",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "0.0.0.0",
			"maxIncomingConnections" : 30000,
			"port" : 27017,
			"ssl" : {
				"mode" : "disabled"
			}
		},
		"replication" : {
			"replSet" : "replicaset"
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
	"operationTime" : Timestamp(1598467113, 1),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1598467113, 1),
		"signature" : {
			"hash" : BinData(0,"jDOBwlqD1dG9mIKgTwX7K5NnJfs="),
			"keyId" : NumberLong("6865318254139670530")
		}
	}
}
```

As we can see from the configuration of running mongodb, the value of `maxIncomingConnections` has been changed from `20000` to `30000`. So the reconfiguration of the database using the data field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-reconfigure-replicaset mops-reconfigure-data-replicaset
```