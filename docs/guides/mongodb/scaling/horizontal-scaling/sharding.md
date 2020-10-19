---
title: Horizontal Scaling MongoDB Shard
menu:
  docs_{{ .version }}:
    identifier: mg-horizontal-scaling-shard
    name: Sharding
    parent: mg-horizontal-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---


{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Horizontal Scale MongoDB Shard

This guide will show you how to use `KubeDB` Enterprise operator to scale the shard of a MongoDB database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [Sharding](/docs/guides/mongodb/clustering/sharding.md) 
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/mongodb/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Sharded Database

Here, we are going to deploy a  `MongoDB` sharded database using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare MongoDB Sharded Database

Now, we are going to deploy a `MongoDB` sharded database with version `3.6.8`.

### Deploy MongoDB Sharded Database 

In this section, we are going to deploy a MongoDB sharded database. Then, in the next sections we will scale shards of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,
    
```yaml
apiVersion: kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-sharding
  namespace: demo
spec:
  version: 3.6.8-v1
  shardTopology:
    configServer:
      replicas: 2
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 2
      shards: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

Let's create the `MongoDB` CR we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/mg-shard.yaml
mongodb.kubedb.com/mg-sharding created
```

Now, wait until `mg-sharding` has status `Running`. i.e,

```console
$ kubectl get mg -n demo                                                            
NAME          VERSION    STATUS    AGE
mg-sharding   3.6.8-v1   Running   10m
```

##### Verify Number of Shard and Shard Replicas

Let's check the number of shards this database from the MongoDB object and the number of statefulsets it has,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.shards'
3

$ kubectl get sts -n demo                                                                 
NAME                    READY   AGE
mg-sharding-configsvr   2/2     23m
mg-sharding-mongos      2/2     22m
mg-sharding-shard0      2/2     23m
mg-sharding-shard1      2/2     23m
mg-sharding-shard2      2/2     23m
```

So, We can see from the both output that the database has 3 shards.

Now, Let's check the number of replicas each shard has from the MongoDB object and the number of pod the statefulsets have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.replicas'
2

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.replicas'  
2
```

We can see from both output that the database has 2 replicas in each shards. 

Also, we can verify the number of shard from an internal mongodb command by execing into a mongos node.

First we need to get the username and password to connect to a mongos instance,
```console
$ kubectl get secrets -n demo mg-sharding-auth -o jsonpath='{.data.\username}' | base64 -d 
root

$ kubectl get secrets -n demo mg-sharding-auth -o jsonpath='{.data.\password}' | base64 -d  
xBC-EwMFivFCgUlK
```

Now let's connect to a mongos instance and run a mongodb internal command to check the number of shards,

```console
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet  
--- Sharding Status --- 
 sharding version: {
    "_id" : 1,
    "minCompatibleVersion" : 5,
    "currentVersion" : 6,
    "clusterId" : ObjectId("5f45fadd48c42afd901e6265")
 }
 shards:
       {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
       {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
       {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
 active mongoses:
       "3.6.8" : 2
 autosplit:
       Currently enabled: yes
 balancer:
       Currently enabled:  yes
       Currently running:  no
       Failed balancer rounds in last 5 attempts:  0
       Migration Results for the last 24 hours: 
               No recent migrations
 databases:
       {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
               config.system.sessions
                       shard key: { "_id" : 1 }
                       unique: false
                       balancing: true
                       chunks:
                               shard0	1
                       { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
```

We can see from the above output that the number of shard is 3.

Also, we can verify the number of replicas each shard has from an internal mongodb command by execing into a shard node.

Now let's connect to a shard instance and run a mongodb internal command to check the number of replicas,

```console
$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
  [
  	{
  		"_id" : 0,
  		"name" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 1,
  		"stateStr" : "PRIMARY",
  		"uptime" : 2403,
  		"optime" : {
  			"ts" : Timestamp(1598424103, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T06:41:43Z"),
  		"syncingTo" : "",
  		"syncSourceHost" : "",
  		"syncSourceId" : -1,
  		"infoMessage" : "",
  		"electionTime" : Timestamp(1598421711, 1),
  		"electionDate" : ISODate("2020-08-26T06:01:51Z"),
  		"configVersion" : 2,
  		"self" : true,
  		"lastHeartbeatMessage" : ""
  	},
  	{
  		"_id" : 1,
  		"name" : "mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 2380,
  		"optime" : {
  			"ts" : Timestamp(1598424103, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598424103, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T06:41:43Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T06:41:43Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T06:41:52.541Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T06:41:52.259Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 2
  	}
  ]
```

We can see from the above output that the number of replica is 2.

##### Verify Number of ConfigServer

Let's check the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.configServer.replicas'                                                                                           11:02:09
2

$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.replicas'                                                                                                11:03:27
2
```

We can see from both command that the database has `2` replicas in the configServer. 

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,

```console
$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet     15:37:23
  
  [
  	{
  		"_id" : 0,
  		"name" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 1,
  		"stateStr" : "PRIMARY",
  		"uptime" : 388,
  		"optime" : {
  			"ts" : Timestamp(1598434641, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T09:37:21Z"),
  		"syncingTo" : "",
  		"syncSourceHost" : "",
  		"syncSourceId" : -1,
  		"infoMessage" : "",
  		"electionTime" : Timestamp(1598434260, 1),
  		"electionDate" : ISODate("2020-08-26T09:31:00Z"),
  		"configVersion" : 2,
  		"self" : true,
  		"lastHeartbeatMessage" : ""
  	},
  	{
  		"_id" : 1,
  		"name" : "mg-sharding-configsvr-1.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 360,
  		"optime" : {
  			"ts" : Timestamp(1598434641, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598434641, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T09:37:21Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T09:37:21Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T09:37:24.731Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T09:37:23.295Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 2
  	}
  ]
```

We can see from the above output that the configServer has 2 nodes.

##### Verify Number of Mongos
Let's check the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.mongos.replicas'                                                                                           11:02:09
2

$ kubectl get sts -n demo mg-sharding-mongos -o json | jq '.spec.replicas'                                                                                                11:03:27
2
```

We can see from both command that the database has `2` replicas in the mongos. 

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,

```console
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet
 --- Sharding Status --- 
    sharding version: {
    	"_id" : 1,
    	"minCompatibleVersion" : 5,
    	"currentVersion" : 6,
    	"clusterId" : ObjectId("5f463327bd21df369bb338bc")
    }
    shards:
          {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
          {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
          {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
    active mongoses:
          "3.6.8" : 2
    autosplit:
          Currently enabled: yes
    balancer:
          Currently enabled:  yes
          Currently running:  no
          Failed balancer rounds in last 5 attempts:  0
          Migration Results for the last 24 hours: 
                  No recent migrations
    databases:
          {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
```

We can see from the above output that the mongos has 2 active nodes.

We are now ready to apply the `MongoDBOpsRequest` CR to update scale up and down all the components of the database.

### Scale Up

Here, we are going to scale up all the components of the database to meet the desired number of replicas after scaling.

#### Create MongoDBOpsRequest

In order to scale up, we have to create a `MongoDBOpsRequest` CR with our configuration. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-hscale-up-shard
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mg-sharding
  horizontalScaling:
    shard: 
      shards: 4
      replicas: 3
    mongos:
      replicas: 3
    configServer:
      replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mops-hscale-up-shard` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.shard.shards` specifies the desired number of shards after scaling.
- `spec.horizontalScaling.shard.replicas` specifies the desired number of replicas of each shard after scaling.
- `spec.horizontalScaling.mongos.replicas` specifies the desired replicas after scaling.
- `spec.horizontalScaling.configServer.replicas` specifies the desired replicas after scaling.

> **Note:** If you don't want to scale all the components together, you can only specify the components (shard, configServer and mongos) that you want to scale.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-up-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-up-shard created
```

#### Verify scaling up is successful 

If everything goes well, `KubeDB` Enterprise operator will update the shard and replicas of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                   TYPE                STATUS       AGE
mops-hscale-up-shard   HorizontalScaling   Successful   9m57s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-hscale-up-shard                     
Name:         mops-hscale-up-shard
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2020-09-30T05:36:41Z
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
        f:databaseRef:
          .:
          f:name:
        f:horizontalScaling:
          .:
          f:configServer:
            .:
            f:replicas:
          f:mongos:
            .:
            f:replicas:
          f:shard:
            .:
            f:replicas:
            f:shards:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2020-09-30T05:36:41Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2020-09-30T05:45:58Z
  Resource Version:  1889687
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-up-shard
  UID:               5dd83c5d-81e1-494d-92c3-5dcbd8ebcaf2
Spec:
  Database Ref:
    Name:  mg-sharding
  Horizontal Scaling:
    Config Server:
      Replicas:  3
    Mongos:
      Replicas:  3
    Shard:
      Replicas:  3
      Shards:    4
  Type:          HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2020-09-30T05:36:41Z
    Message:               MongoDB ops request is horizontally scaling database
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2020-09-30T05:45:58Z
    Message:               Successfully Resumed mongodb: mg-sharding
    Observed Generation:   1
    Reason:                PauseDatabase
    Status:                False
    Type:                  PauseDatabase
    Last Transition Time:  2020-09-30T05:38:31Z
    Message:               Successfully Horizontally Scaled Up Shard Replicas
    Observed Generation:   1
    Reason:                ScaleUpShardReplicas
    Status:                True
    Type:                  ScaleUpShardReplicas
    Last Transition Time:  2020-09-30T05:40:31Z
    Message:               Successfully Horizontally Scaled Up Shard
    Observed Generation:   1
    Reason:                ScaleUpShard
    Status:                True
    Type:                  ScaleUpShard
    Last Transition Time:  2020-09-30T05:45:43Z
    Message:               Successfully Horizontally Scaled Up ConfigServer
    Observed Generation:   1
    Reason:                ScaleUpConfigServer 
    Status:                True
    Type:                  ScaleUpConfigServer 
    Last Transition Time:  2020-09-30T05:45:58Z
    Message:               Successfully Horizontally Scaled Mongos
    Observed Generation:   1
    Reason:                ScaleMongos
    Status:                True
    Type:                  ScaleMongos
    Last Transition Time:  2020-09-30T05:45:58Z
    Message:               Successfully Horizontally Scaled MongoDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                Age   From                        Message
  ----    ------                ----  ----                        -------
  Normal  PauseDatabase         54m   KubeDB Enterprise Operator  Pausing MongoDB mg-sharding in Namespace demo
  Normal  PauseDatabase         54m   KubeDB Enterprise Operator  Successfully Paused MongoDB mg-sharding in Namespace demo
  Normal  ScaleUpShard          50m   KubeDB Enterprise Operator  Successfully Horizontally Scaled Up Shard
  Normal  PauseDatabase         50m   KubeDB Enterprise Operator  Pausing MongoDB mg-sharding in Namespace demo
  Normal  PauseDatabase         50m   KubeDB Enterprise Operator  Successfully Paused MongoDB mg-sharding in Namespace demo
  Normal  Progressing           50m   KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  Progressing           49m   KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  ScaleUpShard          49m   KubeDB Enterprise Operator  Successfully Horizontally Scaled Up Shard
  Normal  ScaleUpConfigServer   45m   KubeDB Enterprise Operator  Successfully Horizontally Scaled Up ConfigServer
  Normal  ScaleMongos           44m   KubeDB Enterprise Operator  Successfully Horizontally Scaled Mongos
  Normal  ResumeDatabase        44m   KubeDB Enterprise Operator  Resuming MongoDB
  Normal  ResumeDatabase        44m   KubeDB Enterprise Operator  Successfully Resumed mongodb
  Normal  Successful            44m   KubeDB Enterprise Operator  Successfully Horizontally Scaled Database
```

#### Verify Number of Shard and Shard Replicas

Now, we are going to verify the number of shards this database has from the MongoDB object, number of statefulsets it has,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.shards'         
4

$ kubectl get sts -n demo                                                                      
NAME                    READY   AGE
mg-sharding-configsvr   2/2     58m
mg-sharding-mongos      2/2     57m
mg-sharding-shard0      3/3     58m
mg-sharding-shard1      3/3     58m
mg-sharding-shard2      3/3     58m
mg-sharding-shard3      3/3     15m
```

Now let's connect to a mongos instance and run a mongodb internal command to check the number of shards,
```console
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet  
--- Sharding Status --- 
 sharding version: {
    "_id" : 1,
    "minCompatibleVersion" : 5,
    "currentVersion" : 6,
    "clusterId" : ObjectId("5f45fadd48c42afd901e6265")
 }
 shards:
       {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017,mg-sharding-shard0-2.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
       {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017,mg-sharding-shard1-2.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
       {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017,mg-sharding-shard2-2.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
       {  "_id" : "shard3",  "host" : "shard3/mg-sharding-shard3-0.mg-sharding-shard3-gvr.demo.svc.cluster.local:27017,mg-sharding-shard3-1.mg-sharding-shard3-gvr.demo.svc.cluster.local:27017,mg-sharding-shard3-2.mg-sharding-shard3-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
 active mongoses:
       "3.6.8" : 2
 autosplit:
       Currently enabled: yes
 balancer:
       Currently enabled:  yes
       Currently running:  no
       Failed balancer rounds in last 5 attempts:  0
       Migration Results for the last 24 hours: 
               No recent migrations
 databases:
       {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
               config.system.sessions
                       shard key: { "_id" : 1 }
                       unique: false
                       balancing: true
                       chunks:
                               shard0	1
                       { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0)
```

From all the above outputs we can see that the number of shards are `4`.

Now, we are going to verify the number of replicas each shard has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.replicas'              
3

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.replicas'          
3
```

Now let's connect to a shard instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
  [
  	{
  		"_id" : 0,
  		"name" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 1,
  		"stateStr" : "PRIMARY",
  		"uptime" : 3907,
  		"optime" : {
  			"ts" : Timestamp(1598425614, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T07:06:54Z"),
  		"syncingTo" : "",
  		"syncSourceHost" : "",
  		"syncSourceId" : -1,
  		"infoMessage" : "",
  		"electionTime" : Timestamp(1598421711, 1),
  		"electionDate" : ISODate("2020-08-26T06:01:51Z"),
  		"configVersion" : 3,
  		"self" : true,
  		"lastHeartbeatMessage" : ""
  	},
  	{
  		"_id" : 1,
  		"name" : "mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 3884,
  		"optime" : {
  			"ts" : Timestamp(1598425614, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598425614, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T07:06:54Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T07:06:54Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T07:06:57.308Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T07:06:57.383Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 3
  	},
  	{
  		"_id" : 2,
  		"name" : "mg-sharding-shard0-2.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 1173,
  		"optime" : {
  			"ts" : Timestamp(1598425614, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598425614, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T07:06:54Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T07:06:54Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T07:06:57.491Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T07:06:56.659Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 3
  	}
  ]
```

From all the above outputs we can see that the replicas of each shard has is `3`. 

#### Verify Number of ConfigServer Replicas
Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.configServer.replicas'                                                                                           11:02:09
3

$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.replicas'
3
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
  [
  	{
  		"_id" : 0,
  		"name" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 1,
  		"stateStr" : "PRIMARY",
  		"uptime" : 1058,
  		"optime" : {
  			"ts" : Timestamp(1598435313, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T09:48:33Z"),
  		"syncingTo" : "",
  		"syncSourceHost" : "",
  		"syncSourceId" : -1,
  		"infoMessage" : "",
  		"electionTime" : Timestamp(1598434260, 1),
  		"electionDate" : ISODate("2020-08-26T09:31:00Z"),
  		"configVersion" : 3,
  		"self" : true,
  		"lastHeartbeatMessage" : ""
  	},
  	{
  		"_id" : 1,
  		"name" : "mg-sharding-configsvr-1.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 1031,
  		"optime" : {
  			"ts" : Timestamp(1598435313, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598435313, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T09:48:33Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T09:48:33Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T09:48:35.250Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T09:48:34.860Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 3
  	},
  	{
  		"_id" : 2,
  		"name" : "mg-sharding-configsvr-2.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 460,
  		"optime" : {
  			"ts" : Timestamp(1598435313, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598435313, 2),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T09:48:33Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T09:48:33Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T09:48:35.304Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T09:48:34.729Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-sharding-configsvr-1.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-sharding-configsvr-1.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 1,
  		"infoMessage" : "",
  		"configVersion" : 3
  	}
  ]
```

From all the above outputs we can see that the replicas of the configServer is `3`. That means we have successfully scaled up the replicas of the MongoDB configServer replicas.

#### Verify Number of Mongos Replicas
Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.mongos.replicas'                                                                                           11:02:09
3

$ kubectl get sts -n demo mg-sharding-mongos -o json | jq '.spec.replicas'
3
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet
  --- Sharding Status --- 
    sharding version: {
    	"_id" : 1,
    	"minCompatibleVersion" : 5,
    	"currentVersion" : 6,
    	"clusterId" : ObjectId("5f463327bd21df369bb338bc")
    }
    shards:
          {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
          {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
          {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
    active mongoses:
          "3.6.8" : 3
    autosplit:
          Currently enabled: yes
    balancer:
          Currently enabled:  yes
          Currently running:  no
          Failed balancer rounds in last 5 attempts:  0
          Migration Results for the last 24 hours: 
                  No recent migrations
    databases:
          {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                  config.system.sessions
                          shard key: { "_id" : 1 }
                          unique: false
                          balancing: true
                          chunks:
                                  shard0	1
                          { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0)
```

From all the above outputs we can see that the replicas of the mongos is `3`. That means we have successfully scaled up the replicas of the MongoDB mongos replicas.


So, we have successfully scaled up all the components of the MongoDB database.

### Scale Down

Here, we are going to scale down both the shard and their replicas to meet the desired number of replicas after scaling.

#### Create MongoDBOpsRequest

In order to scale down, we have to create a `MongoDBOpsRequest` CR with our configuration. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-hscale-down-shard
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mg-sharding
  horizontalScaling:
    shard: 
      shards: 3
      replicas: 2
    mongos:
      replicas: 2
    configServer:
      replicas: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mops-hscale-down-shard` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.shard.shards` specifies the desired number of shards after scaling.
- `spec.horizontalScaling.shard.replicas` specifies the desired number of replicas of each shard after scaling.
- `spec.horizontalScaling.configServer.replicas` specifies the desired replicas after scaling.
- `spec.horizontalScaling.mongos.replicas` specifies the desired replicas after scaling.

> **Note:** If you don't want to scale all the components together, you can only specify the components (shard, configServer and mongos) that you want to scale.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-down-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-down-shard created
```

#### Verify scaling down is successful 

If everything goes well, `KubeDB` Enterprise operator will update the shards and replicas `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                     TYPE                STATUS       AGE
mops-hscale-down-shard   HorizontalScaling   Successful   81s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale down the the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-hscale-down-shard                     
 Name:         mops-hscale-down-shard
 Namespace:    demo
 Labels:       <none>
 Annotations:  <none>
 API Version:  ops.kubedb.com/v1alpha1
 Kind:         MongoDBOpsRequest
 Metadata:
   Creation Timestamp:  2020-09-30T07:03:52Z
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
         f:databaseRef:
           .:
           f:name:
         f:horizontalScaling:
           .:
           f:configServer:
             .:
             f:replicas:
           f:mongos:
             .:
             f:replicas:
           f:shard:
             .:
             f:replicas:
             f:shards:
         f:type:
     Manager:      kubectl-client-side-apply
     Operation:    Update
     Time:         2020-09-30T07:03:52Z
     API Version:  ops.kubedb.com/v1alpha1
     Fields Type:  FieldsV1
     fieldsV1:
       f:status:
         .:
         f:conditions:
         f:observedGeneration:
         f:phase:
     Manager:         kubedb-enterprise
     Operation:       Update
     Time:            2020-09-30T07:05:34Z
   Resource Version:  1908351
   Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-down-shard
   UID:               8edc1db9-a537-4bae-ac93-d1f5e9fcb758
 Spec:
   Database Ref:
     Name:  mg-sharding
   Horizontal Scaling:
     Config Server:
       Replicas:  2
     Mongos:
       Replicas:  2
     Shard:
       Replicas:  2
       Shards:    3
   Type:          HorizontalScaling
 Status:
   Conditions:
     Last Transition Time:  2020-09-30T07:03:52Z
     Message:               MongoDB ops request is horizontally scaling database
     Observed Generation:   1
     Reason:                HorizontalScaling
     Status:                True
     Type:                  HorizontalScaling
     Last Transition Time:  2020-09-30T07:05:34Z
     Message:               Successfully paused mongodb: mg-sharding
     Observed Generation:   1
     Reason:                PauseDatabase
     Status:                True
     Type:                  PauseDatabase
     Last Transition Time:  2020-09-30T07:04:42Z
     Message:               Successfully Horizontally Scaled Down Shard Replicas
     Observed Generation:   1
     Reason:                ScaleDownShardReplicas
     Status:                True
     Type:                  ScaleDownShardReplicas
     Last Transition Time:  2020-09-30T07:04:42Z
     Message:               Successfully started mongodb load balancer
     Observed Generation:   1
     Reason:                StartingBalancer
     Status:                True
     Type:                  StartingBalancer
     Last Transition Time:  2020-09-30T07:05:04Z
     Message:               Successfully Horizontally Scaled Down Shard
     Observed Generation:   1
     Reason:                ScaleDownShard
     Status:                True
     Type:                  ScaleDownShard
     Last Transition Time:  2020-09-30T07:05:14Z
     Message:               Successfully Horizontally Scaled Down ConfigServer
     Observed Generation:   1
     Reason:                ScaleDownConfigServer 
     Status:                True
     Type:                  ScaleDownConfigServer 
     Last Transition Time:  2020-09-30T07:05:34Z
     Message:               Successfully Horizontally Scaled Mongos
     Observed Generation:   1
     Reason:                ScaleMongos
     Status:                True
     Type:                  ScaleMongos
     Last Transition Time:  2020-09-30T07:05:34Z
     Message:               Successfully Horizontally Scaled MongoDB
     Observed Generation:   1
     Reason:                Successful
     Status:                True
     Type:                  Successful
   Observed Generation:     1
   Phase:                   Successful
 Events:
   Type    Reason                  Age    From                        Message
   ----    ------                  ----   ----                        -------
   Normal  PauseDatabase           2m17s  KubeDB Enterprise Operator  Pausing MongoDB mg-sharding in Namespace demo
   Normal  PauseDatabase           2m17s  KubeDB Enterprise Operator  Successfully Paused MongoDB mg-sharding in Namespace demo
   Normal  ScaleDownShardReplicas  87s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Down Shard Replicas
   Normal  StartingBalancer        87s    KubeDB Enterprise Operator  Starting Balancer
   Normal  StartingBalancer        87s    KubeDB Enterprise Operator  Successfully Started Balancer
   Normal  ScaleDownShard          65s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Down Shard
   Normal  ScaleDownConfigServer   55s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Down ConfigServer
   Normal  ScaleMongos             35s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Mongos
   Normal  ResumeDatabase          35s    KubeDB Enterprise Operator  Resuming MongoDB
   Normal  ResumeDatabase          35s    KubeDB Enterprise Operator  Successfully Resumed mongodb
   Normal  Successful              35s    KubeDB Enterprise Operator  Successfully Horizontally Scaled Database
   Normal  PauseDatabase           35s    KubeDB Enterprise Operator  Pausing MongoDB mg-sharding in Namespace demo
   Normal  PauseDatabase           35s    KubeDB Enterprise Operator  Successfully Paused MongoDB mg-sharding in Namespace demo
```

##### Verify Number of Shard and Shard Replicas

Now, we are going to verify the number of shards this database has from the MongoDB object, number of statefulsets it has,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.shards'     
3

$ kubectl get sts -n demo                                                                      
NAME                    READY   AGE
mg-sharding-configsvr   2/2     78m
mg-sharding-mongos      2/2     77m
mg-sharding-shard0      2/2     78m
mg-sharding-shard1      2/2     78m
mg-sharding-shard2      2/2     78m
```

Now let's connect to a mongos instance and run a mongodb internal command to check the number of shards,
```console
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet  
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("5f45fadd48c42afd901e6265")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "3.6.8" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                No recent migrations
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0)
```

From all the above outputs we can see that the number of shards are `3`.

Now, we are going to verify the number of replicas each shard has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.replicas'                                                                            13:05:25
2

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.replicas'                                                                                           13:05:30
2
```

Now let's connect to a shard instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet        13:06:31
  [
  	{
  		"_id" : 0,
  		"name" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 1,
  		"stateStr" : "PRIMARY",
  		"uptime" : 4784,
  		"optime" : {
  			"ts" : Timestamp(1598426494, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T07:21:34Z"),
  		"syncingTo" : "",
  		"syncSourceHost" : "",
  		"syncSourceId" : -1,
  		"infoMessage" : "",
  		"electionTime" : Timestamp(1598421711, 1),
  		"electionDate" : ISODate("2020-08-26T06:01:51Z"),
  		"configVersion" : 4,
  		"self" : true,
  		"lastHeartbeatMessage" : ""
  	},
  	{
  		"_id" : 1,
  		"name" : "mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 4761,
  		"optime" : {
  			"ts" : Timestamp(1598426494, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598426494, 1),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T07:21:34Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T07:21:34Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T07:21:34.403Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T07:21:34.383Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 4
  	}
  ]
```

From all the above outputs we can see that the replicas of each shard has is `2`. 

##### Verify Number of ConfigServer Replicas

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.configServer.replicas'                                                                                           11:02:09
3

$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.replicas'
3
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 1316,
		"optime" : {
			"ts" : Timestamp(1598435569, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2020-08-26T09:52:49Z"),
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1598434260, 1),
		"electionDate" : ISODate("2020-08-26T09:31:00Z"),
		"configVersion" : 4,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-sharding-configsvr-1.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 1288,
		"optime" : {
			"ts" : Timestamp(1598435569, 1),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1598435569, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2020-08-26T09:52:49Z"),
		"optimeDurableDate" : ISODate("2020-08-26T09:52:49Z"),
		"lastHeartbeat" : ISODate("2020-08-26T09:52:52.348Z"),
		"lastHeartbeatRecv" : ISODate("2020-08-26T09:52:52.347Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-0.mg-sharding-configsvr-gvr.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 4
	}
]
```

From all the above outputs we can see that the replicas of the configServer is `2`. That means we have successfully scaled down the replicas of the MongoDB configServer replicas.

##### Verify Number of Mongos Replicas

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.mongos.replicas'                                                                                           11:02:09
2

$ kubectl get sts -n demo mg-sharding-mongos -o json | jq '.spec.replicas'
2
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("5f463327bd21df369bb338bc")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-gvr.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "3.6.8" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                No recent migrations
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "_id" : { "$minKey" : 1 } } -->> { "_id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0)
```

From all the above outputs we can see that the replicas of the mongos is `2`. That means we have successfully scaled down the replicas of the MongoDB mongos replicas.

So, we have successfully scaled down all the components of the MongoDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-sharding
kubectl delete mongodbopsrequest -n demo mops-vscale-up-shard mops-vscale-down-shard 
```