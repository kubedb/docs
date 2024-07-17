---
title: Reconfigure MongoDB Sharded Cluster
menu:
  docs_{{ .version }}:
    identifier: mg-reconfigure-shard
    name: Sharding
    parent: mg-reconfigure
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure MongoDB Shard

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a MongoDB shard.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Sharding](/docs/guides/mongodb/clustering/sharding.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/mongodb/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `MongoDB` sharded database using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBOpsRequest` to reconfigure its configuration.

### Prepare MongoDB Shard

Now, we are going to deploy a `MongoDB` sharded database with version `4.4.26`.

### Deploy MongoDB database 

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
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-sharding
  namespace: demo
spec:
  version: 4.4.26
  shardTopology:
    configServer:
      replicas: 3
      configSecret:
        name: mg-custom-config
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      replicas: 2
      configSecret:
        name: mg-custom-config
    shard:
      replicas: 3
      shards: 2
      configSecret:
        name: mg-custom-config
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mg-shard-config.yaml
mongodb.kubedb.com/mg-sharding created
```

Now, wait until `mg-sharding` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME          VERSION    STATUS    AGE
mg-sharding   4.4.26      Ready     3m23s
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a mongodb instance,
```bash
$ kubectl get secrets -n demo mg-sharding-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mg-sharding-auth -o jsonpath='{.data.\password}' | base64 -d
Dv8F55zVNiEkhHM6
```

Now let's connect to a mongodb instance from each type of nodes and run a mongodb internal command to check the configuration we have provided.

```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
{
	"bindIp" : "*",
	"ipv6" : true,
	"maxIncomingConnections" : 10000,
	"port" : 27017,
	"tls" : {
		"mode" : "disabled"
	}
}

$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
{
	"bindIp" : "*",
	"ipv6" : true,
	"maxIncomingConnections" : 10000,
	"port" : 27017,
	"tls" : {
		"mode" : "disabled"
	}
}

$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
{
	"bindIp" : "*",
	"ipv6" : true,
	"maxIncomingConnections" : 10000,
	"port" : 27017,
	"tls" : {
		"mode" : "disabled"
	}
}
```

As we can see from the configuration of ready mongodb, the value of `maxIncomingConnections` has been set to `10000` in all nodes.

### Reconfigure using new secret

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
  name: mops-reconfigure-shard
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-sharding
  configuration:
    shard:
      configSecret:
        name: new-custom-config
    configServer:
      configSecret:
        name: new-custom-config 
    mongos:
      configSecret:
        name: new-custom-config   
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-shard` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.shard.configSecret.name` specifies the name of the new secret for shard nodes.
- `spec.configuration.configServer.configSecret.name` specifies the name of the new secret for configServer nodes.
- `spec.configuration.mongos.configSecret.name` specifies the name of the new secret for mongos nodes.
- `spec.customConfig.arbiter.configSecret.name` could also be specified with a config-secret.
- Have a look [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

> **Note:** If you don't want to reconfigure all the components together, you can only specify the components (shard, configServer and mongos) that you want to reconfigure.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-shard created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `MongoDB` object.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                     TYPE          STATUS       AGE
mops-reconfigure-shard   Reconfigure   Successful   3m8s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-shard  

```

Now let's connect to a mongodb instance from each type of nodes and run a mongodb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
  {
  	"bindIp" : "0.0.0.0",
  	"maxIncomingConnections" : 20000,
  	"port" : 27017,
  	"ssl" : {
  		"mode" : "disabled"
  	}
  }

$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
  {
  	"bindIp" : "0.0.0.0",
  	"maxIncomingConnections" : 20000,
  	"port" : 27017,
  	"ssl" : {
  		"mode" : "disabled"
  	}
  }

$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
  {
  	"bindIp" : "0.0.0.0",
  	"maxIncomingConnections" : 20000,
  	"port" : 27017,
  	"ssl" : {
  		"mode" : "disabled"
  	}
  }
```

As we can see from the configuration of ready mongodb, the value of `maxIncomingConnections` has been changed from `10000` to `20000` in all type of nodes. So the reconfiguration of the database is successful.

### Reconfigure using apply config

Now we will reconfigure this database again to set `maxIncomingConnections` to `30000`. This time we won't use a new secret. We will use the `applyConfig` field of the `MongoDBOpsRequest`. This will merge the new config in the existing secret.

#### Create MongoDBOpsRequest

Now, we will use the new configuration in the `data` field in the `MongoDBOpsRequest` CR. The `MongoDBOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfigure-apply-shard
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-sharding
  configuration:
    shard:
      applyConfig:
        mongod.conf: |-
          net:
            maxIncomingConnections: 30000
    configServer:
      applyConfig:
        mongod.conf: |-
          net:
            maxIncomingConnections: 30000
    mongos:
      applyConfig:
        mongod.conf: |-
          net:
            maxIncomingConnections: 30000
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-apply-shard` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.shard.applyConfig` specifies the new configuration that will be merged in the existing secret for shard nodes.
- `spec.configuration.configServer.applyConfig` specifies the new configuration that will be merged in the existing secret for configServer nodes.
- `spec.configuration.mongos.applyConfig` specifies the new configuration that will be merged in the existing secret for mongos nodes.
- `spec.customConfig.arbiter.configSecret.name` could also be specified with a config-secret.
- Have a look [here](/docs/guides/mongodb/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

> **Note:** If you don't want to reconfigure all the components together, you can only specify the components (shard, configServer and mongos) that you want to reconfigure.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/reconfigure/mops-reconfigure-apply-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-reconfigure-apply-shard created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                          TYPE          STATUS       AGE
mops-reconfigure-apply-shard Reconfigure   Successful   3m24s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-reconfigure-apply-shard
Name:         mops-reconfigure-apply-shard
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T13:08:25Z
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
          f:configServer:
            .:
            f:configSecret:
              .:
              f:name:
          f:mongos:
            .:
            f:configSecret:
              .:
              f:name:
          f:shard:
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
    Time:         2021-03-02T13:08:25Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:configuration:
          f:configServer:
            f:podTemplate:
              .:
              f:controller:
              f:metadata:
              f:spec:
                .:
                f:resources:
          f:mongos:
            f:podTemplate:
              .:
              f:controller:
              f:metadata:
              f:spec:
                .:
                f:resources:
          f:shard:
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
    Time:            2021-03-02T13:08:25Z
  Resource Version:  103635
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-reconfigure-apply-shard
  UID:               ab454bcb-164c-4fa2-9eaa-dd47c60fe874
Spec:
  Apply: IfReady
  Configuration:
    Config Server:
      Apply Config:  net:
  maxIncomingConnections: 30000
    
    Mongos:
      Apply Config:  net:
  maxIncomingConnections: 30000
    
    Shard:
      Apply Config:  net:
  maxIncomingConnections: 30000
  
  Database Ref:
    Name:  mg-sharding
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2021-03-02T13:08:25Z
    Message:               MongoDB ops request is reconfiguring database
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2021-03-02T13:10:10Z
    Message:               Successfully Reconfigured MongoDB
    Observed Generation:   1
    Reason:                ReconfigureConfigServer
    Status:                True
    Type:                  ReconfigureConfigServer
    Last Transition Time:  2021-03-02T13:13:15Z
    Message:               Successfully Reconfigured MongoDB
    Observed Generation:   1
    Reason:                ReconfigureShard
    Status:                True
    Type:                  ReconfigureShard
    Last Transition Time:  2021-03-02T13:14:10Z
    Message:               Successfully Reconfigured MongoDB
    Observed Generation:   1
    Reason:                ReconfigureMongos
    Status:                True
    Type:                  ReconfigureMongos
    Last Transition Time:  2021-03-02T13:14:10Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                   Age    From                        Message
  ----    ------                   ----   ----                        -------
  Normal  PauseDatabase            13m    KubeDB Ops-manager operator  Pausing MongoDB demo/mg-sharding
  Normal  PauseDatabase            13m    KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-sharding
  Normal  ReconfigureConfigServer  12m    KubeDB Ops-manager operator  Successfully Reconfigured MongoDB
  Normal  ReconfigureShard         9m7s   KubeDB Ops-manager operator  Successfully Reconfigured MongoDB
  Normal  ReconfigureMongos        8m12s  KubeDB Ops-manager operator  Successfully Reconfigured MongoDB
  Normal  ResumeDatabase           8m12s  KubeDB Ops-manager operator  Resuming MongoDB demo/mg-sharding
  Normal  ResumeDatabase           8m12s  KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-sharding
  Normal  Successful               8m12s  KubeDB Ops-manager operator  Successfully Reconfigured Database
```

Now let's connect to a mongodb instance from each type of nodes and run a mongodb internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
{
	"bindIp" : "*",
	"ipv6" : true,
	"maxIncomingConnections" : 20000,
	"port" : 27017,
	"tls" : {
		"mode" : "disabled"
	}
}

$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
{
	"bindIp" : "*",
	"ipv6" : true,
	"maxIncomingConnections" : 20000,
	"port" : 27017,
	"tls" : {
		"mode" : "disabled"
	}
}

$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p Dv8F55zVNiEkhHM6 --eval "db._adminCommand( {getCmdLineOpts: 1}).parsed.net" --quiet
{
	"bindIp" : "*",
	"ipv6" : true,
	"maxIncomingConnections" : 20000,
	"port" : 27017,
	"tls" : {
		"mode" : "disabled"
	}
}
```

As we can see from the configuration of ready mongodb, the value of `maxIncomingConnections` has been changed from `20000` to `30000` in all nodes. So the reconfiguration of the database using the data field is successful.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-sharding
kubectl delete mongodbopsrequest -n demo mops-reconfigure-shard mops-reconfigure-apply-shard
```