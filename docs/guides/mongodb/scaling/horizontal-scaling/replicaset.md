---
title: Horizontal Scaling MongoDB Replicaset
menu:
  docs_{{ .version }}:
    identifier: mg-horizontal-scaling-replicaset
    name: Replicaset
    parent: mg-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Horizontal Scale MongoDB Replicaset

This guide will show you how to use `KubeDB` Enterprise operator to scale the replicaset of a MongoDB database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md) 
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/mongodb/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Replicaset

Here, we are going to deploy a  `MongoDB` replicaset using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare MongoDB Replicaset Database

Now, we are going to deploy a `MongoDB` replicaset database with version `3.6.8`.

### Deploy MongoDB replicaset 

In this section, we are going to deploy a MongoDB replicaset database. Then, in the next section we will scale the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MongoDB
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "3.6.8-v1"
  replicaSet: 
    name: "replicaset"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `MongoDB` CR we have shown above,

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/mg-replicaset.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` has status `Running`. i.e,

```console
$ kubectl get mg -n demo                                                                                                                                             20:05:47
  NAME            VERSION    STATUS    AGE
  mg-replicaset   3.6.8-v1   Running   2m36s
```

Let's check the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-replicaset -o json | jq '.spec.replicas'                                                                                            11:02:09
3

$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.replicas'                                                                                                11:03:27
3
```

We can see from both command that the database has 3 replicas in the replicaset. 

Also, we can verify the replicas of the replicaset from an internal mongodb command by execing into a replica.

First we need to get the username and password to connect to a mongodb instance,
```console
$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\username}' | base64 -d                                                                         11:09:51
root

$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         11:10:44
nrKuxni0wDSMrgwy
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,

```console
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
    {
        "_id" : 0,
        "name" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
        "health" : 1,
        "state" : 1,
        "stateStr" : "PRIMARY",
        "uptime" : 631,
        "optime" : {
            "ts" : Timestamp(1598418585, 1),
            "t" : NumberLong(2)
        },
        "optimeDate" : ISODate("2020-08-26T05:09:45Z"),
        "syncingTo" : "",
        "syncSourceHost" : "",
        "syncSourceId" : -1,
        "infoMessage" : "",
        "electionTime" : Timestamp(1598417963, 1),
        "electionDate" : ISODate("2020-08-26T04:59:23Z"),
        "configVersion" : 3,
        "self" : true,
        "lastHeartbeatMessage" : ""
    },
    {
        "_id" : 1,
        "name" : "mg-replicaset-1.mg-replicaset-gvr.demo.svc.cluster.local:27017",
        "health" : 1,
        "state" : 2,
        "stateStr" : "SECONDARY",
        "uptime" : 606,
        "optime" : {
            "ts" : Timestamp(1598418585, 1),
            "t" : NumberLong(2)
        },
        "optimeDurable" : {
            "ts" : Timestamp(1598418585, 1),
            "t" : NumberLong(2)
        },
        "optimeDate" : ISODate("2020-08-26T05:09:45Z"),
        "optimeDurableDate" : ISODate("2020-08-26T05:09:45Z"),
        "lastHeartbeat" : ISODate("2020-08-26T05:09:49.489Z"),
        "lastHeartbeatRecv" : ISODate("2020-08-26T05:09:50.484Z"),
        "pingMs" : NumberLong(0),
        "lastHeartbeatMessage" : "",
        "syncingTo" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
        "syncSourceHost" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
        "syncSourceId" : 0,
        "infoMessage" : "",
        "configVersion" : 3
    },
    {
        "_id" : 2,
        "name" : "mg-replicaset-2.mg-replicaset-gvr.demo.svc.cluster.local:27017",
        "health" : 1,
        "state" : 2,
        "stateStr" : "SECONDARY",
        "uptime" : 590,
        "optime" : {
            "ts" : Timestamp(1598418585, 1),
            "t" : NumberLong(2)
        },
        "optimeDurable" : {
            "ts" : Timestamp(1598418585, 1),
            "t" : NumberLong(2)
        },
        "optimeDate" : ISODate("2020-08-26T05:09:45Z"),
        "optimeDurableDate" : ISODate("2020-08-26T05:09:45Z"),
        "lastHeartbeat" : ISODate("2020-08-26T05:09:49.539Z"),
        "lastHeartbeatRecv" : ISODate("2020-08-26T05:09:50.330Z"),
        "pingMs" : NumberLong(0),
        "lastHeartbeatMessage" : "",
        "syncingTo" : "mg-replicaset-1.mg-replicaset-gvr.demo.svc.cluster.local:27017",
        "syncSourceHost" : "mg-replicaset-1.mg-replicaset-gvr.demo.svc.cluster.local:27017",
        "syncSourceId" : 1,
        "infoMessage" : "",
        "configVersion" : 3
    }
]
```

We can see from the above output that the replicaset has 3 nodes.

We are now ready to apply the `MongoDBOpsRequest` CR to scale this database.

## Scale Up Replicas

Here, we are going to scale up the replicas of the replicaset to meet the desired number of replicas after scaling.

#### Create MongoDBOpsRequest

In order to scale up the replicas of the replicaset of the database, we have to create a `MongoDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-hscale-up-replicaset
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mg-replicaset
  horizontalScaling:
    replicas: 4
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `mops-hscale-up-replicaset` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-up-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-up-replicaset created
```

#### Verify Replicaset replicas scaled up successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                        TYPE                STATUS       AGE
mops-hscale-up-replicaset   HorizontalScaling   Successful   106s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-hscale-up-replicaset                     
  Name:         mops-hscale-up-replicaset
  Namespace:    demo
  Labels:       <none>
  Annotations:  API Version:  ops.kubedb.com/v1alpha1
  Kind:         MongoDBOpsRequest
  Metadata:
    Creation Timestamp:  2020-08-26T05:22:33Z
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
          f:databaseRef:
            .:
            f:name:
          f:horizontalScaling:
            .:
            f:replicas:
          f:type:
      Manager:      kubectl
      Operation:    Update
      Time:         2020-08-26T05:22:33Z
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
      Time:            2020-08-26T05:23:18Z
    Resource Version:  5681626
    Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-up-replicaset
    UID:               8b9b03c4-d95d-41af-b418-312ad81c49de
  Spec:
    Database Ref:
      Name:  mg-replicaset
    Horizontal Scaling:
      Replicas:  4
    Type:        HorizontalScaling
  Status:
    Conditions:
      Last Transition Time:  2020-08-26T05:22:33Z
      Message:               MongoDB ops request is being processed
      Observed Generation:   1
      Reason:                Scaling
      Status:                True
      Type:                  Scaling
      Last Transition Time:  2020-08-26T05:22:33Z
      Message:               Successfully paused mongodb: mg-replicaset
      Observed Generation:   1
      Reason:                PauseDatabase
      Status:                True
      Type:                  PauseDatabase
      Last Transition Time:  2020-08-26T05:23:18Z
      Message:               Successfully Scaled Up Replicas of StatefulSet
      Observed Generation:   1
      Reason:                ScaleUpReplicaSet
      Status:                True
      Type:                  ScaleUpReplicaSet
      Last Transition Time:  2020-08-26T05:23:18Z
      Message:               Successfully Resumed mongodb: mg-replicaset
      Observed Generation:   1
      Reason:                ResumeDatabase
      Status:                True
      Type:                  ResumeDatabase
      Last Transition Time:  2020-08-26T05:23:18Z
      Message:               Successfully completed the modification process
      Observed Generation:   1
      Reason:                Successful
      Status:                True
      Type:                  Successful
    Observed Generation:     1
    Phase:                   Successful
  Events:
    Type    Reason             Age    From                        Message
    ----    ------             ----   ----                        -------
    Normal  PauseDatabase      2m10s  KubeDB Enterprise Operator  Pausing Mongodb mg-replicaset in Namespace demo
    Normal  PauseDatabase      2m10s  KubeDB Enterprise Operator  Successfully Paused Mongodb mg-replicaset in Namespace demo
    Normal  PauseDatabase      2m10s  KubeDB Enterprise Operator  Pausing Mongodb mg-replicaset in Namespace demo
    Normal  PauseDatabase      2m10s  KubeDB Enterprise Operator  Successfully Paused Mongodb mg-replicaset in Namespace demo
    Normal  ScaleUpReplicaSet  85s    KubeDB Enterprise Operator  Successfully Scaled Up Replicas of StatefulSet
    Normal  ResumeDatabase     85s    KubeDB Enterprise Operator  Resuming MongoDB
    Normal  ResumeDatabase     85s    KubeDB Enterprise Operator  Successfully Resumed mongodb
    Normal  Successful         85s    KubeDB Enterprise Operator  Successfully Scaled Database
```

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-replicaset -o json | jq '.spec.replicas'                                                                                            11:26:38
4

$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.replicas'                                                                                                11:27:13
4
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet             11:28:20
  
  [
  	{
  		"_id" : 0,
  		"name" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 1,
  		"stateStr" : "PRIMARY",
  		"uptime" : 1749,
  		"optime" : {
  			"ts" : Timestamp(1598419697, 4),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T05:28:17Z"),
  		"syncingTo" : "",
  		"syncSourceHost" : "",
  		"syncSourceId" : -1,
  		"infoMessage" : "",
  		"electionTime" : Timestamp(1598417963, 1),
  		"electionDate" : ISODate("2020-08-26T04:59:23Z"),
  		"configVersion" : 4,
  		"self" : true,
  		"lastHeartbeatMessage" : ""
  	},
  	{
  		"_id" : 1,
  		"name" : "mg-replicaset-1.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 1724,
  		"optime" : {
  			"ts" : Timestamp(1598419697, 4),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598419697, 4),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T05:28:17Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T05:28:17Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T05:28:28.990Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T05:28:27.959Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 4
  	},
  	{
  		"_id" : 2,
  		"name" : "mg-replicaset-2.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 1708,
  		"optime" : {
  			"ts" : Timestamp(1598419697, 4),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598419697, 4),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T05:28:17Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T05:28:17Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T05:28:28.990Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T05:28:27.959Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 0,
  		"infoMessage" : "",
  		"configVersion" : 4
  	},
  	{
  		"_id" : 3,
  		"name" : "mg-replicaset-3.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"health" : 1,
  		"state" : 2,
  		"stateStr" : "SECONDARY",
  		"uptime" : 310,
  		"optime" : {
  			"ts" : Timestamp(1598419697, 4),
  			"t" : NumberLong(2)
  		},
  		"optimeDurable" : {
  			"ts" : Timestamp(1598419697, 4),
  			"t" : NumberLong(2)
  		},
  		"optimeDate" : ISODate("2020-08-26T05:28:17Z"),
  		"optimeDurableDate" : ISODate("2020-08-26T05:28:17Z"),
  		"lastHeartbeat" : ISODate("2020-08-26T05:28:29.153Z"),
  		"lastHeartbeatRecv" : ISODate("2020-08-26T05:28:28.379Z"),
  		"pingMs" : NumberLong(0),
  		"lastHeartbeatMessage" : "",
  		"syncingTo" : "mg-replicaset-2.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"syncSourceHost" : "mg-replicaset-2.mg-replicaset-gvr.demo.svc.cluster.local:27017",
  		"syncSourceId" : 2,
  		"infoMessage" : "",
  		"configVersion" : 4
  	}
  ]
```

From all the above outputs we can see that the replicas of the replicaset is `4`. That means we have successfully scaled up the replicas of the MongoDB replicaset.


### Scale Down Replicas

Here, we are going to scale down the replicas of the replicaset to meet the desired number of replicas after scaling.

#### Create MongoDBOpsRequest

In order to scale down the replicas of the replicaset of the database, we have to create a `MongoDBOpsRequest` CR with our desired replicas. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-hscale-down-replicaset
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mg-replicaset
  horizontalScaling:
    replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `mops-hscale-down-replicaset` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-down-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-down-replicaset created
```

#### Verify Replicaset replicas scaled down successfully 

If everything goes well, `KubeDB` Enterprise operator will update the replicas of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                          TYPE                STATUS       AGE
mops-hscale-down-replicaset   HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-hscale-down-replicaset                     
 Name:         mops-hscale-down-replicaset
 Namespace:    demo
 Labels:       <none>
 Annotations:  API Version:  ops.kubedb.com/v1alpha1
 Kind:         MongoDBOpsRequest
 Metadata:
   Creation Timestamp:  2020-08-26T05:36:49Z
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
         f:databaseRef:
           .:
           f:name:
         f:horizontalScaling:
           .:
           f:replicas:
         f:type:
     Manager:      kubectl
     Operation:    Update
     Time:         2020-08-26T05:36:49Z
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
     Time:            2020-08-26T05:36:54Z
   Resource Version:  5691961
   Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-down-replicaset
   UID:               9563b401-2d20-4624-8374-c008a83a58ad
 Spec:
   Database Ref:
     Name:  mg-replicaset
   Horizontal Scaling:
     Replicas:  3
   Type:        HorizontalScaling
 Status:
   Conditions:
     Last Transition Time:  2020-08-26T05:36:49Z
     Message:               MongoDB ops request is being processed
     Observed Generation:   1
     Reason:                Scaling
     Status:                True
     Type:                  Scaling
     Last Transition Time:  2020-08-26T05:36:49Z
     Message:               Successfully paused mongodb: mg-replicaset
     Observed Generation:   1
     Reason:                PauseDatabase
     Status:                True
     Type:                  PauseDatabase
     Last Transition Time:  2020-08-26T05:36:54Z
     Message:               Successfully Scale Down Replicas of Replicaset
     Observed Generation:   1
     Reason:                ScaleDownReplicaSet
     Status:                True
     Type:                  ScaleDownReplicaSet
     Last Transition Time:  2020-08-26T05:36:54Z
     Message:               Successfully Resumed mongodb: mg-replicaset
     Observed Generation:   1
     Reason:                ResumeDatabase
     Status:                True
     Type:                  ResumeDatabase
     Last Transition Time:  2020-08-26T05:36:54Z
     Message:               Successfully completed the modification process
     Observed Generation:   1
     Reason:                Successful
     Status:                True
     Type:                  Successful
   Observed Generation:     1
   Phase:                   Successful
 Events:
   Type    Reason               Age    From                        Message
   ----    ------               ----   ----                        -------
   Normal  PauseDatabase        3m1s   KubeDB Enterprise Operator  Pausing Mongodb mg-replicaset in Namespace demo
   Normal  PauseDatabase        3m1s   KubeDB Enterprise Operator  Successfully Paused Mongodb mg-replicaset in Namespace demo
   Normal  ScalingDown          3m1s   KubeDB Enterprise Operator  Scaling Down Replicas of replicaSet
   Normal  ScalingDown          3m1s   KubeDB Enterprise Operator  Scaling Down Replicas of replicaSet
   Normal  ScaleDownReplicaSet  2m56s  KubeDB Enterprise Operator  Successfully Scale Down Replicas of Replicaset
   Normal  ResumeDatabase       2m56s  KubeDB Enterprise Operator  Resuming MongoDB
   Normal  ResumeDatabase       2m56s  KubeDB Enterprise Operator  Successfully Resumed mongodb
   Normal  Successful           2m56s  KubeDB Enterprise Operator  Successfully Scaled Database
```

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```console
$ kubectl get mongodb -n demo mg-replicaset -o json | jq '.spec.replicas' 
3

$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.replicas'
3
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```console
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet 

[
	{
		"_id" : 0,
		"name" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 2475,
		"optime" : {
			"ts" : Timestamp(1598420435, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2020-08-26T05:40:35Z"),
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1598417963, 1),
		"electionDate" : ISODate("2020-08-26T04:59:23Z"),
		"configVersion" : 5,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-replicaset-1.mg-replicaset-gvr.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 2450,
		"optime" : {
			"ts" : Timestamp(1598420425, 1),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1598420425, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2020-08-26T05:40:25Z"),
		"optimeDurableDate" : ISODate("2020-08-26T05:40:25Z"),
		"lastHeartbeat" : ISODate("2020-08-26T05:40:34.917Z"),
		"lastHeartbeatRecv" : ISODate("2020-08-26T05:40:33.976Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 5
	},
	{
		"_id" : 2,
		"name" : "mg-replicaset-2.mg-replicaset-gvr.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 2434,
		"optime" : {
			"ts" : Timestamp(1598420425, 1),
			"t" : NumberLong(2)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1598420425, 1),
			"t" : NumberLong(2)
		},
		"optimeDate" : ISODate("2020-08-26T05:40:25Z"),
		"optimeDurableDate" : ISODate("2020-08-26T05:40:25Z"),
		"lastHeartbeat" : ISODate("2020-08-26T05:40:34.917Z"),
		"lastHeartbeatRecv" : ISODate("2020-08-26T05:40:33.976Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-gvr.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 5
	}
]
```

From all the above outputs we can see that the replicas of the replicaset is `3`. That means we have successfully scaled down the replicas of the MongoDB replicaset.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-vscale-replicaset
```