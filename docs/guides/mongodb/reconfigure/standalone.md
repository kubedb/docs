---
title: Reconfigure Standalone MongoDB Database
menu:
  docs_{{ .version }}:
    identifier: mg-reconfigure-standalone
    name: Standalone
    parent: mg-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfigure MongoDB Standalone Database

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a MongoDB standalone database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Reconfigure Overview](/docs/guides/mongodb/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `MongoDB` standalone using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBOpsRequest` to reconfigure its configuration.

### Prepare MongoDB Standalone Database

Now, we are going to deploy a `MongoDB` standalone database with version `3.6.8`.

### Deploy MongoDB standalone 

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
apiVersion: kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-standalone
  namespace: demo
spec:
  version: "3.6.8-v1"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  configSource:
      configMap:
        name: mg-custom-config
```

Let's create the `MongoDB` CR we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mg-standalone-config.yaml
mongodb.kubedb.com/mg-standalone created
```

Now, wait until `mg-standalone` has status `Running`. i.e,

```console
$ kubectl get mg -n demo                                                                                                                                             20:05:47
NAME            VERSION    STATUS    AGE
mg-standalone   3.6.8-v1   Running   23s
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a mongodb instance,
```console
$ kubectl get secrets -n demo mg-standalone-auth -o jsonpath='{.data.\username}' | base64 -d                                                                         11:09:51
root

$ kubectl get secrets -n demo mg-standalone-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         11:10:44
m6lXjZugrC4VEpB8
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the configuration we have provided.

```console
$ kubectl exec -n demo  mg-standalone-0  -- mongo admin -u root -p m6lXjZugrC4VEpB8 --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet                          18:35:59
  {
  	"argv" : [
  		"mongod",
  		"--dbpath=/data/db",
  		"--auth",
  		"--bind_ip=0.0.0.0",
  		"--port=27017",
  		"--sslMode=disabled",
  		"--config=/data/configdb/mongod.conf"
  	],
  	"parsed" : {
  		"config" : "/data/configdb/mongod.conf",
  		"net" : {
  			"bindIp" : "0.0.0.0",
  			"maxIncomingConnections" : 10000,
  			"port" : 27017,
  			"ssl" : {
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
  name: mops-reconfigure-standalone
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-standalone
  customConfig:
    standalone:
      configMap:
        name: new-custom-config
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-standalone` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.customConfig.standalone.configMap.name` specifies the name of the new configmap.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-standalone.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-standalone created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Enterprise operator will update the `configSource` of `MongoDB` object.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                          TYPE          STATUS       AGE
mops-reconfigure-standalone   Reconfigure   Successful   10m
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-standalone                                                                                             20:06:00
  Name:         mops-reconfigure-standalone
  Namespace:    demo
  Labels:       <none>
  Annotations:  API Version:  ops.kubedb.com/v1alpha1
  Kind:         MongoDBOpsRequest
  Metadata:
    Creation Timestamp:  2020-08-26T13:53:50Z
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
            f:standalone:
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
      Time:         2020-08-26T13:53:50Z
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
      Time:            2020-08-26T13:54:15Z
    Resource Version:  6073824
    Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-reconfigure-standalone
    UID:               a1706ad9-e3a9-468e-bc7b-1d4dcbdf82c8
  Spec:
    Custom Config:
      Standalone:
        Config Map:
          Name:  new-custom-config
    Database Ref:
      Name:  mg-standalone
    Type:    Reconfigure
  Status:
    Conditions:
      Last Transition Time:  2020-08-26T13:53:50Z
      Message:               MongoDB ops request is being processed
      Observed Generation:   1
      Reason:                Reconfigure
      Status:                True
      Type:                  Reconfigure
      Last Transition Time:  2020-08-26T13:53:50Z
      Message:               Successfully paused mongodb: mg-standalone
      Observed Generation:   1
      Reason:                PauseDatabase
      Status:                True
      Type:                  PauseDatabase
      Last Transition Time:  2020-08-26T13:54:15Z
      Message:               Successfully Reconfigured mongodb
      Observed Generation:   1
      Reason:                ReconfigureStandalone
      Status:                True
      Type:                  ReconfigureStandalone
      Last Transition Time:  2020-08-26T13:54:15Z
      Message:               Succefully Resumed mongodb: mg-standalone
      Observed Generation:   1
      Reason:                ResumeDatabase
      Status:                True
      Type:                  ResumeDatabase
      Last Transition Time:  2020-08-26T13:54:15Z
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
    Normal  PauseDatabase          12m   KubeDB Enterprise Operator  Pausing Mongodb mg-standalone in Namespace demo
    Normal  PauseDatabase          12m   KubeDB Enterprise Operator  Successfully Paused Mongodb mg-standalone in Namespace demo
    Normal  ReconfigureStandalone  11m   KubeDB Enterprise Operator  Successfully Reconfigured mongodb
    Normal  ResumeDatabase         11m   KubeDB Enterprise Operator  Resuming MongoDB
    Normal  ResumeDatabase         11m   KubeDB Enterprise Operator  Successfully Started Balancer
    Normal  Successful             11m   KubeDB Enterprise Operator  Successfully Reconfigured Database
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the new configuration we have provided.

```console
$ kubectl exec -n demo  mg-standalone-0  -- mongo admin -u root -p m6lXjZugrC4VEpB8 --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
  {
  	"argv" : [
  		"mongod",
  		"--dbpath=/data/db",
  		"--auth",
  		"--bind_ip=0.0.0.0",
  		"--port=27017",
  		"--sslMode=disabled",
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
  		"security" : {
  			"authorization" : "enabled"
  		},
  		"storage" : {
  			"dbPath" : "/data/db"
  		}
  	},
  	"ok" : 1
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
  name: mops-reconfigure-data-standalone
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-standalone
  customConfig:
    standalone:
      data:
        mongod.conf: |
          net:
            maxIncomingConnections: 30000
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-data-standalone` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.customConfig.standalone.data` specifies the new configuration that will be merged in the existing configMap.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-data-standalone.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-data-standalone created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Enterprise operator will merge this new config with the existing configuration.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                               TYPE          STATUS       AGE
mops-reconfigure-data-standalone   Reconfigure   Successful   38s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-data-standalone
Name:         mops-reconfigure-data-standalone
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2020-08-26T14:37:55Z
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
          f:standalone:
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
    Time:         2020-08-26T14:37:55Z
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
    Time:            2020-08-26T14:38:25Z
  Resource Version:  6107705
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-reconfigure-data-standalone
  UID:               a4fd1184-eabc-4038-9d33-56b7e08c7444
Spec:
  Custom Config:
    Standalone:
      Data:
        mongod.conf:  net:
  maxIncomingConnections: 30000

  Database Ref:
    Name:  mg-standalone
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2020-08-26T14:37:55Z
    Message:               MongoDB ops request is being processed
    Observed Generation:   1
    Reason:                Scaling
    Status:                True
    Type:                  Scaling
    Last Transition Time:  2020-08-26T14:38:25Z
    Message:               Successfully Reconfigured mongodb
    Observed Generation:   1
    Reason:                ReconfigureStandalone
    Status:                True
    Type:                  ReconfigureStandalone
    Last Transition Time:  2020-08-26T14:38:25Z
    Message:               Succefully Resumed mongodb: mg-standalone
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-08-26T14:38:25Z
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
  Normal  ReconfigureStandalone  27s   KubeDB Enterprise Operator  Successfully Reconfigured mongodb
  Normal  ResumeDatabase         27s   KubeDB Enterprise Operator  Resuming MongoDB
  Normal  ResumeDatabase         27s   KubeDB Enterprise Operator  Successfully Started Balancer
  Normal  Successful             27s   KubeDB Enterprise Operator  Successfully Reconfigured Database
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the new configuration we have provided.

```console
$ kubectl exec -n demo  mg-standalone-0  -- mongo admin -u root -p m6lXjZugrC4VEpB8 --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--bind_ip=0.0.0.0",
		"--port=27017",
		"--sslMode=disabled",
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
		"security" : {
			"authorization" : "enabled"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1
}
```

As we can see from the configuration of running mongodb, the value of `maxIncomingConnections` has been changed from `20000` to `30000`. So the reconfiguration of the database using the data field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-standalone
kubectl delete mongodbopsrequest -n demo mops-reconfigure-standalone mops-reconfigure-data-standalone
```