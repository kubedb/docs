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

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale MongoDB Shard

This guide will show you how to use `KubeDB` Ops-manager operator to scale the shard of a MongoDB database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Sharding](/docs/guides/mongodb/clustering/sharding.md) 
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/mongodb/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Sharded Database

Here, we are going to deploy a  `MongoDB` sharded database using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare MongoDB Sharded Database

Now, we are going to deploy a `MongoDB` sharded database with version `4.4.26`.

### Deploy MongoDB Sharded Database 

In this section, we are going to deploy a MongoDB sharded database. Then, in the next sections we will scale shards of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,
    
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
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 3
      shards: 2
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

Let's create the `MongoDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/mg-shard.yaml
mongodb.kubedb.com/mg-sharding created
```

Now, wait until `mg-sharding` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo                                                            
NAME          VERSION    STATUS    AGE
mg-sharding   4.4.26      Ready     10m
```

##### Verify Number of Shard and Shard Replicas

Let's check the number of shards this database from the MongoDB object and the number of petsets it has,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.shards'
2

$ kubectl get sts -n demo                                                                 
NAME                    READY   AGE
mg-sharding-configsvr   3/3     23m
mg-sharding-mongos      2/2     22m
mg-sharding-shard0      3/3     23m
mg-sharding-shard1      3/3     23m
```

So, We can see from the both output that the database has 2 shards.

Now, Let's check the number of replicas each shard has from the MongoDB object and the number of pod the petsets have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.replicas'
3

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.replicas'  
3
```

We can see from both output that the database has 3 replicas in each shards. 

Also, we can verify the number of shard from an internal mongodb command by execing into a mongos node.

First we need to get the username and password to connect to a mongos instance,
```bash
$ kubectl get secrets -n demo mg-sharding-auth -o jsonpath='{.data.\username}' | base64 -d 
root

$ kubectl get secrets -n demo mg-sharding-auth -o jsonpath='{.data.\password}' | base64 -d  
xBC-EwMFivFCgUlK
```

Now let's connect to a mongos instance and run a mongodb internal command to check the number of shards,

```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet  
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("603e5a4bec470e6b4197e10b")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-2.mg-sharding-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
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

We can see from the above output that the number of shard is 2.

Also, we can verify the number of replicas each shard has from an internal mongodb command by execing into a shard node.

Now let's connect to a shard instance and run a mongodb internal command to check the number of replicas,

```bash
$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 338,
		"optime" : {
			"ts" : Timestamp(1614699416, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:36:56Z"),
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614699092, 1),
		"electionDate" : ISODate("2021-03-02T15:31:32Z"),
		"configVersion" : 3,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 291,
		"optime" : {
			"ts" : Timestamp(1614699413, 1),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614699413, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:36:53Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:36:53Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:36:56.692Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:36:56.015Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 3
	},
	{
		"_id" : 2,
		"name" : "mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 259,
		"optime" : {
			"ts" : Timestamp(1614699413, 1),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614699413, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:36:53Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:36:53Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:36:56.732Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:36:57.773Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 3
	}
]
```

We can see from the above output that the number of replica is 3.

##### Verify Number of ConfigServer

Let's check the number of replicas this database has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.configServer.replicas'
3

$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.replicas'
3
```

We can see from both command that the database has `3` replicas in the configServer. 

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,

```bash
$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-sharding-configsvr-0.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 423,
		"optime" : {
			"ts" : Timestamp(1614699492, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:38:12Z"),
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614699081, 2),
		"electionDate" : ISODate("2021-03-02T15:31:21Z"),
		"configVersion" : 3,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 385,
		"optime" : {
			"ts" : Timestamp(1614699492, 1),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614699492, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:38:12Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:38:12Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:38:13.573Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:38:12.725Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-configsvr-0.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-0.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 3
	},
	{
		"_id" : 2,
		"name" : "mg-sharding-configsvr-2.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 340,
		"optime" : {
			"ts" : Timestamp(1614699490, 8),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614699490, 8),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:38:10Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:38:10Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:38:11.665Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:38:11.827Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-configsvr-0.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-0.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 3
	}
]
```

We can see from the above output that the configServer has 3 nodes.

##### Verify Number of Mongos
Let's check the number of replicas this database has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.mongos.replicas'
2

$ kubectl get sts -n demo mg-sharding-mongos -o json | jq '.spec.replicas'
2
```

We can see from both command that the database has `2` replicas in the mongos. 

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,

```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("603e5a4bec470e6b4197e10b")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-2.mg-sharding-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
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
      shards: 3
      replicas: 4
    mongos:
      replicas: 3
    configServer:
      replicas: 4
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-up-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-up-shard created
```

#### Verify scaling up is successful 

If everything goes well, `KubeDB` Ops-manager operator will update the shard and replicas of `MongoDB` object and related `PetSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                   TYPE                STATUS       AGE
mops-hscale-up-shard   HorizontalScaling   Successful   9m57s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-hscale-up-shard                     
Name:         mops-hscale-up-shard
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T16:23:16Z
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
    Time:         2021-03-02T16:23:16Z
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
    Time:            2021-03-02T16:23:16Z
  Resource Version:  147313
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-up-shard
  UID:               982014fc-1655-44e7-946c-859626ae0247
Spec:
  Database Ref:
    Name:  mg-sharding
  Horizontal Scaling:
    Config Server:
      Replicas:  4
    Mongos:
      Replicas:  3
    Shard:
      Replicas:  4
      Shards:    3
  Type:          HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-03-02T16:23:16Z
    Message:               MongoDB ops request is horizontally scaling database
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2021-03-02T16:25:31Z
    Message:               Successfully Horizontally Scaled Up Shard Replicas
    Observed Generation:   1
    Reason:                ScaleUpShardReplicas
    Status:                True
    Type:                  ScaleUpShardReplicas
    Last Transition Time:  2021-03-02T16:33:07Z
    Message:               Successfully Horizontally Scaled Up Shard
    Observed Generation:   1
    Reason:                ScaleUpShard
    Status:                True
    Type:                  ScaleUpShard
    Last Transition Time:  2021-03-02T16:34:35Z
    Message:               Successfully Horizontally Scaled Up ConfigServer
    Observed Generation:   1
    Reason:                ScaleUpConfigServer 
    Status:                True
    Type:                  ScaleUpConfigServer 
    Last Transition Time:  2021-03-02T16:36:30Z
    Message:               Successfully Horizontally Scaled Mongos
    Observed Generation:   1
    Reason:                ScaleMongos
    Status:                True
    Type:                  ScaleMongos
    Last Transition Time:  2021-03-02T16:36:30Z
    Message:               Successfully Horizontally Scaled MongoDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                Age    From                        Message
  ----    ------                ----   ----                        -------
  Normal  PauseDatabase         13m    KubeDB Ops-manager operator  Pausing MongoDB demo/mg-sharding
  Normal  PauseDatabase         13m    KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-sharding
  Normal  ScaleUpShardReplicas  11m    KubeDB Ops-manager operator  Successfully Horizontally Scaled Up Shard Replicas
  Normal  ResumeDatabase        11m    KubeDB Ops-manager operator  Resuming MongoDB demo/mg-sharding
  Normal  ResumeDatabase        11m    KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-sharding
  Normal  ScaleUpShardReplicas  11m    KubeDB Ops-manager operator  Successfully Horizontally Scaled Up Shard Replicas
  Normal  ScaleUpShardReplicas  11m    KubeDB Ops-manager operator  Successfully Horizontally Scaled Up Shard Replicas
  Normal  Progressing           8m20s  KubeDB Ops-manager operator  Successfully updated PetSets Resources
  Normal  Progressing           4m5s   KubeDB Ops-manager operator  Successfully updated PetSets Resources
  Normal  ScaleUpShard          3m59s  KubeDB Ops-manager operator  Successfully Horizontally Scaled Up Shard
  Normal  PauseDatabase         3m59s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-sharding
  Normal  PauseDatabase         3m59s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-sharding
  Normal  ScaleUpConfigServer   2m31s  KubeDB Ops-manager operator  Successfully Horizontally Scaled Up ConfigServer
  Normal  ScaleMongos           36s    KubeDB Ops-manager operator  Successfully Horizontally Scaled Mongos
  Normal  ResumeDatabase        36s    KubeDB Ops-manager operator  Resuming MongoDB demo/mg-sharding
  Normal  ResumeDatabase        36s    KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-sharding
  Normal  Successful            36s    KubeDB Ops-manager operator  Successfully Horizontally Scaled Database
```

#### Verify Number of Shard and Shard Replicas

Now, we are going to verify the number of shards this database has from the MongoDB object, number of petsets it has,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.shards'         
3

$ kubectl get sts -n demo                                                                      
NAME                    READY   AGE
mg-sharding-configsvr   4/4     66m
mg-sharding-mongos      3/3     64m
mg-sharding-shard0      4/4     66m
mg-sharding-shard1      4/4     66m
mg-sharding-shard2      4/4     12m
```

Now let's connect to a mongos instance and run a mongodb internal command to check the number of shards,
```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet  
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("603e5a4bec470e6b4197e10b")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-3.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-2.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-3.mg-sharding-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-pods.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-pods.demo.svc.cluster.local:27017,mg-sharding-shard2-2.mg-sharding-shard2-pods.demo.svc.cluster.local:27017,mg-sharding-shard2-3.mg-sharding-shard2-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 3
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  2
        Last reported error:  Couldn't get a connection within the time limit
        Time of Reported error:  Tue Mar 02 2021 16:17:53 GMT+0000 (UTC)
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

Now, we are going to verify the number of replicas each shard has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.replicas'              
4

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.replicas'          
4
```

Now let's connect to a shard instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 1464,
		"optime" : {
			"ts" : Timestamp(1614703143, 10),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:39:03Z"),
		"syncingTo" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 4,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 1433,
		"optime" : {
			"ts" : Timestamp(1614703143, 10),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703143, 10),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:39:03Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:39:03Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:39:07.800Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:39:08.087Z"),
		"pingMs" : NumberLong(6),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614701678, 2),
		"electionDate" : ISODate("2021-03-02T16:14:38Z"),
		"configVersion" : 4
	},
	{
		"_id" : 2,
		"name" : "mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 1433,
		"optime" : {
			"ts" : Timestamp(1614703143, 10),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703143, 10),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:39:03Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:39:03Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:39:08.575Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:39:08.580Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 4
	},
	{
		"_id" : 3,
		"name" : "mg-sharding-shard0-3.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 905,
		"optime" : {
			"ts" : Timestamp(1614703143, 10),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703143, 10),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:39:03Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:39:03Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:39:06.683Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:39:07.980Z"),
		"pingMs" : NumberLong(10),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 4
	}
]
```

From all the above outputs we can see that the replicas of each shard has is `4`. 

#### Verify Number of ConfigServer Replicas
Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.configServer.replicas'
4

$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.replicas'
4
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-sharding-configsvr-0.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 1639,
		"optime" : {
			"ts" : Timestamp(1614703138, 2),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:38:58Z"),
		"syncingTo" : "mg-sharding-configsvr-2.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-2.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 2,
		"infoMessage" : "",
		"configVersion" : 4,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 1623,
		"optime" : {
			"ts" : Timestamp(1614703138, 2),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703138, 2),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:38:58Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:38:58Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:38:58.979Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:38:59.291Z"),
		"pingMs" : NumberLong(3),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614701497, 2),
		"electionDate" : ISODate("2021-03-02T16:11:37Z"),
		"configVersion" : 4
	},
	{
		"_id" : 2,
		"name" : "mg-sharding-configsvr-2.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 1623,
		"optime" : {
			"ts" : Timestamp(1614703138, 2),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703138, 2),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:38:58Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:38:58Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:38:58.885Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:39:00.188Z"),
		"pingMs" : NumberLong(3),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 4
	},
	{
		"_id" : 3,
		"name" : "mg-sharding-configsvr-3.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 296,
		"optime" : {
			"ts" : Timestamp(1614703138, 2),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703138, 2),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:38:58Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:38:58Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:38:58.977Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:39:00.276Z"),
		"pingMs" : NumberLong(1),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 4
	}
]
```

From all the above outputs we can see that the replicas of the configServer is `3`. That means we have successfully scaled up the replicas of the MongoDB configServer replicas.

#### Verify Number of Mongos Replicas
Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.mongos.replicas'
3

$ kubectl get sts -n demo mg-sharding-mongos -o json | jq '.spec.replicas'
3
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("603e5a4bec470e6b4197e10b")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-3.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-2.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-3.mg-sharding-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard2",  "host" : "shard2/mg-sharding-shard2-0.mg-sharding-shard2-pods.demo.svc.cluster.local:27017,mg-sharding-shard2-1.mg-sharding-shard2-pods.demo.svc.cluster.local:27017,mg-sharding-shard2-2.mg-sharding-shard2-pods.demo.svc.cluster.local:27017,mg-sharding-shard2-3.mg-sharding-shard2-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 3
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  2
        Last reported error:  Couldn't get a connection within the time limit
        Time of Reported error:  Tue Mar 02 2021 16:17:53 GMT+0000 (UTC)
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
      shards: 2
      replicas: 3
    mongos:
      replicas: 2
    configServer:
      replicas: 3
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-down-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-down-shard created
```

#### Verify scaling down is successful 

If everything goes well, `KubeDB` Ops-manager operator will update the shards and replicas `MongoDB` object and related `PetSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                     TYPE                STATUS       AGE
mops-hscale-down-shard   HorizontalScaling   Successful   81s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale down the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-hscale-down-shard                     
Name:         mops-hscale-down-shard
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T16:41:11Z
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
    Time:         2021-03-02T16:41:11Z
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
    Time:            2021-03-02T16:41:11Z
  Resource Version:  149077
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-down-shard
  UID:               0f83c457-9498-4144-a397-226141851751
Spec:
  Database Ref:
    Name:  mg-sharding
  Horizontal Scaling:
    Config Server:
      Replicas:  3
    Mongos:
      Replicas:  2
    Shard:
      Replicas:  3
      Shards:    2
  Type:          HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-03-02T16:41:11Z
    Message:               MongoDB ops request is horizontally scaling database
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2021-03-02T16:42:11Z
    Message:               Successfully Horizontally Scaled Down Shard Replicas
    Observed Generation:   1
    Reason:                ScaleDownShardReplicas
    Status:                True
    Type:                  ScaleDownShardReplicas
    Last Transition Time:  2021-03-02T16:42:12Z
    Message:               Successfully started mongodb load balancer
    Observed Generation:   1
    Reason:                StartingBalancer
    Status:                True
    Type:                  StartingBalancer
    Last Transition Time:  2021-03-02T16:43:03Z
    Message:               Successfully Horizontally Scaled Down Shard
    Observed Generation:   1
    Reason:                ScaleDownShard
    Status:                True
    Type:                  ScaleDownShard
    Last Transition Time:  2021-03-02T16:43:24Z
    Message:               Successfully Horizontally Scaled Down ConfigServer
    Observed Generation:   1
    Reason:                ScaleDownConfigServer 
    Status:                True
    Type:                  ScaleDownConfigServer 
    Last Transition Time:  2021-03-02T16:43:34Z
    Message:               Successfully Horizontally Scaled Mongos
    Observed Generation:   1
    Reason:                ScaleMongos
    Status:                True
    Type:                  ScaleMongos
    Last Transition Time:  2021-03-02T16:43:34Z
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
  Normal  PauseDatabase           6m29s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-sharding
  Normal  PauseDatabase           6m29s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-sharding
  Normal  ScaleDownShardReplicas  5m29s  KubeDB Ops-manager operator  Successfully Horizontally Scaled Down Shard Replicas
  Normal  StartingBalancer        5m29s  KubeDB Ops-manager operator  Starting Balancer
  Normal  StartingBalancer        5m28s  KubeDB Ops-manager operator  Successfully Started Balancer
  Normal  ScaleDownShard          4m37s  KubeDB Ops-manager operator  Successfully Horizontally Scaled Down Shard
  Normal  ScaleDownConfigServer   4m16s  KubeDB Ops-manager operator  Successfully Horizontally Scaled Down ConfigServer
  Normal  ScaleMongos             4m6s   KubeDB Ops-manager operator  Successfully Horizontally Scaled Mongos
  Normal  ResumeDatabase          4m6s   KubeDB Ops-manager operator  Resuming MongoDB demo/mg-sharding
  Normal  ResumeDatabase          4m6s   KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-sharding
  Normal  Successful              4m6s   KubeDB Ops-manager operator  Successfully Horizontally Scaled Database
```

##### Verify Number of Shard and Shard Replicas

Now, we are going to verify the number of shards this database has from the MongoDB object, number of petsets it has,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.shards'     
2

$ kubectl get sts -n demo                                                                      
NAME                    READY   AGE
mg-sharding-configsvr   3/3     77m
mg-sharding-mongos      2/2     75m
mg-sharding-shard0      3/3     77m
mg-sharding-shard1      3/3     77m
```

Now let's connect to a mongos instance and run a mongodb internal command to check the number of shards,
```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet  
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("603e5a4bec470e6b4197e10b")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-2.mg-sharding-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  2
        Last reported error:  Couldn't get a connection within the time limit
        Time of Reported error:  Tue Mar 02 2021 16:17:53 GMT+0000 (UTC)
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

From all the above outputs we can see that the number of shards are `2`.

Now, we are going to verify the number of replicas each shard has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.shard.replicas'
3

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.replicas'
3
```

Now let's connect to a shard instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-sharding-shard0-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 2096,
		"optime" : {
			"ts" : Timestamp(1614703771, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:49:31Z"),
		"syncingTo" : "mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 2,
		"infoMessage" : "",
		"configVersion" : 5,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 2065,
		"optime" : {
			"ts" : Timestamp(1614703771, 1),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703771, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:49:31Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:49:31Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:49:39.092Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:49:40.074Z"),
		"pingMs" : NumberLong(18),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614701678, 2),
		"electionDate" : ISODate("2021-03-02T16:14:38Z"),
		"configVersion" : 5
	},
	{
		"_id" : 2,
		"name" : "mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 2065,
		"optime" : {
			"ts" : Timestamp(1614703771, 1),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703771, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:49:31Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:49:31Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:49:38.712Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:49:39.885Z"),
		"pingMs" : NumberLong(4),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 5
	}
]
```

From all the above outputs we can see that the replicas of each shard has is `3`. 

##### Verify Number of ConfigServer Replicas

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.configServer.replicas'
3

$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.replicas'
3
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-sharding-configsvr-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-sharding-configsvr-0.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 2345,
		"optime" : {
			"ts" : Timestamp(1614703841, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:50:41Z"),
		"syncingTo" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 5,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 2329,
		"optime" : {
			"ts" : Timestamp(1614703841, 1),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703841, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:50:41Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:50:41Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:50:45.874Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:50:44.194Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614701497, 2),
		"electionDate" : ISODate("2021-03-02T16:11:37Z"),
		"configVersion" : 5
	},
	{
		"_id" : 2,
		"name" : "mg-sharding-configsvr-2.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 2329,
		"optime" : {
			"ts" : Timestamp(1614703841, 1),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614703841, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2021-03-02T16:50:41Z"),
		"optimeDurableDate" : ISODate("2021-03-02T16:50:41Z"),
		"lastHeartbeat" : ISODate("2021-03-02T16:50:45.778Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T16:50:46.091Z"),
		"pingMs" : NumberLong(1),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-sharding-configsvr-1.mg-sharding-configsvr-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 5
	}
]
```

From all the above outputs we can see that the replicas of the configServer is `3`. That means we have successfully scaled down the replicas of the MongoDB configServer replicas.

##### Verify Number of Mongos Replicas

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the petset have,

```bash
$ kubectl get mongodb -n demo mg-sharding -o json | jq '.spec.shardTopology.mongos.replicas'
2

$ kubectl get sts -n demo mg-sharding-mongos -o json | jq '.spec.replicas'
2
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-sharding-mongos-0  -- mongo admin -u root -p xBC-EwMFivFCgUlK --eval "sh.status()" --quiet
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("603e5a4bec470e6b4197e10b")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mg-sharding-shard0-0.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-1.mg-sharding-shard0-pods.demo.svc.cluster.local:27017,mg-sharding-shard0-2.mg-sharding-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mg-sharding-shard1-0.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-1.mg-sharding-shard1-pods.demo.svc.cluster.local:27017,mg-sharding-shard1-2.mg-sharding-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  2
        Last reported error:  Couldn't get a connection within the time limit
        Time of Reported error:  Tue Mar 02 2021 16:17:53 GMT+0000 (UTC)
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

```bash
kubectl delete mg -n demo mg-sharding
kubectl delete mongodbopsrequest -n demo mops-vscale-up-shard mops-vscale-down-shard 
```