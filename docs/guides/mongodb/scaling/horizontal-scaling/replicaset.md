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

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Horizontal Scale MongoDB Replicaset

This guide will show you how to use `KubeDB` Ops-manager operator to scale the replicaset of a MongoDB database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md) 
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/mongodb/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Replicaset

Here, we are going to deploy a  `MongoDB` replicaset using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare MongoDB Replicaset Database

Now, we are going to deploy a `MongoDB` replicaset database with version `4.4.26`.

### Deploy MongoDB replicaset 

In this section, we are going to deploy a MongoDB replicaset database. Then, in the next section we will scale the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "4.4.26"
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/mg-replicaset.yaml
mongodb.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME            VERSION   STATUS    AGE
mg-replicaset   4.4.26     Ready     2m36s
```

Let's check the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```bash
$ kubectl get mongodb -n demo mg-replicaset -o json | jq '.spec.replicas'
3

$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.replicas'
3
```

We can see from both command that the database has 3 replicas in the replicaset. 

Also, we can verify the replicas of the replicaset from an internal mongodb command by execing into a replica.

First we need to get the username and password to connect to a mongodb instance,
```bash
$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\password}' | base64 -d
nrKuxni0wDSMrgwy
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,

```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 171,
		"optime" : {
			"ts" : Timestamp(1614698544, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:22:24Z"),
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614698393, 2),
		"electionDate" : ISODate("2021-03-02T15:19:53Z"),
		"configVersion" : 3,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-replicaset-1.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 128,
		"optime" : {
			"ts" : Timestamp(1614698544, 1),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614698544, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:22:24Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:22:24Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:22:32.411Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:22:31.543Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 3
	},
	{
		"_id" : 2,
		"name" : "mg-replicaset-2.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 83,
		"optime" : {
			"ts" : Timestamp(1614698544, 1),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614698544, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:22:24Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:22:24Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:22:30.615Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:22:31.543Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-up-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-up-replicaset created
```

#### Verify Replicaset replicas scaled up successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                        TYPE                STATUS       AGE
mops-hscale-up-replicaset   HorizontalScaling   Successful   106s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-hscale-up-replicaset                     
Name:         mops-hscale-up-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T15:23:14Z
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
          f:replicas:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-02T15:23:14Z
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
    Time:            2021-03-02T15:23:14Z
  Resource Version:  129882
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-up-replicaset
  UID:               e97dac5c-5e3a-4153-9b31-8ba02af54bcb
Spec:
  Database Ref:
    Name:  mg-replicaset
  Horizontal Scaling:
    Replicas:  4
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-03-02T15:23:14Z
    Message:               MongoDB ops request is horizontally scaling database
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2021-03-02T15:24:00Z
    Message:               Successfully Horizontally Scaled Up ReplicaSet
    Observed Generation:   1
    Reason:                ScaleUpReplicaSet
    Status:                True
    Type:                  ScaleUpReplicaSet
    Last Transition Time:  2021-03-02T15:24:00Z
    Message:               Successfully Horizontally Scaled MongoDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason             Age   From                        Message
  ----    ------             ----  ----                        -------
  Normal  PauseDatabase      91s   KubeDB Ops-manager operator  Pausing MongoDB demo/mg-replicaset
  Normal  PauseDatabase      91s   KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-replicaset
  Normal  ScaleUpReplicaSet  45s   KubeDB Ops-manager operator  Successfully Horizontally Scaled Up ReplicaSet
  Normal  ResumeDatabase     45s   KubeDB Ops-manager operator  Resuming MongoDB demo/mg-replicaset
  Normal  ResumeDatabase     45s   KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-replicaset
  Normal  Successful         45s   KubeDB Ops-manager operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```bash
$ kubectl get mongodb -n demo mg-replicaset -o json | jq '.spec.replicas'
4

$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.replicas'
4
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet
[
	{
		"_id" : 0,
		"name" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 344,
		"optime" : {
			"ts" : Timestamp(1614698724, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:25:24Z"),
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614698393, 2),
		"electionDate" : ISODate("2021-03-02T15:19:53Z"),
		"configVersion" : 4,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-replicaset-1.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 301,
		"optime" : {
			"ts" : Timestamp(1614698712, 2),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614698712, 2),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:25:12Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:25:12Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:25:23.889Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:25:25.179Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 4
	},
	{
		"_id" : 2,
		"name" : "mg-replicaset-2.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 256,
		"optime" : {
			"ts" : Timestamp(1614698712, 2),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614698712, 2),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:25:12Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:25:12Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:25:23.888Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:25:25.136Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 4
	},
	{
		"_id" : 3,
		"name" : "mg-replicaset-3.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 93,
		"optime" : {
			"ts" : Timestamp(1614698712, 2),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614698712, 2),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:25:12Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:25:12Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:25:23.926Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:25:24.089Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/horizontal-scaling/mops-hscale-down-replicaset.yaml
mongodbopsrequest.ops.kubedb.com/mops-hscale-down-replicaset created
```

#### Verify Replicaset replicas scaled down successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `MongoDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                          TYPE                STATUS       AGE
mops-hscale-down-replicaset   HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-hscale-down-replicaset                     
Name:         mops-hscale-down-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T15:25:57Z
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
          f:replicas:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-02T15:25:57Z
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
    Time:            2021-03-02T15:25:57Z
  Resource Version:  130393
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-hscale-down-replicaset
  UID:               fbfee7f8-1dd5-4f58-aad7-ad2e2d66b295
Spec:
  Database Ref:
    Name:  mg-replicaset
  Horizontal Scaling:
    Replicas:  3
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2021-03-02T15:25:57Z
    Message:               MongoDB ops request is horizontally scaling database
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2021-03-02T15:26:17Z
    Message:               Successfully Horizontally Scaled Down ReplicaSet
    Observed Generation:   1
    Reason:                ScaleDownReplicaSet
    Status:                True
    Type:                  ScaleDownReplicaSet
    Last Transition Time:  2021-03-02T15:26:17Z
    Message:               Successfully Horizontally Scaled MongoDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason               Age   From                        Message
  ----    ------               ----  ----                        -------
  Normal  PauseDatabase        50s   KubeDB Ops-manager operator  Pausing MongoDB demo/mg-replicaset
  Normal  PauseDatabase        50s   KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-replicaset
  Normal  ScaleDownReplicaSet  30s   KubeDB Ops-manager operator  Successfully Horizontally Scaled Down ReplicaSet
  Normal  ResumeDatabase       30s   KubeDB Ops-manager operator  Resuming MongoDB demo/mg-replicaset
  Normal  ResumeDatabase       30s   KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-replicaset
  Normal  Successful           30s   KubeDB Ops-manager operator  Successfully Horizontally Scaled Database
```

Now, we are going to verify the number of replicas this database has from the MongoDB object, number of pods the statefulset have,

```bash
$ kubectl get mongodb -n demo mg-replicaset -o json | jq '.spec.replicas' 
3

$ kubectl get sts -n demo mg-replicaset -o json | jq '.spec.replicas'
3
```

Now let's connect to a mongodb instance and run a mongodb internal command to check the number of replicas,
```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db.adminCommand( { replSetGetStatus : 1 } ).members" --quiet 
[
	{
		"_id" : 0,
		"name" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 410,
		"optime" : {
			"ts" : Timestamp(1614698784, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:26:24Z"),
		"syncingTo" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1614698393, 2),
		"electionDate" : ISODate("2021-03-02T15:19:53Z"),
		"configVersion" : 5,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 1,
		"name" : "mg-replicaset-1.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 367,
		"optime" : {
			"ts" : Timestamp(1614698784, 1),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614698784, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:26:24Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:26:24Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:26:29.423Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:26:29.330Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 5
	},
	{
		"_id" : 2,
		"name" : "mg-replicaset-2.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 322,
		"optime" : {
			"ts" : Timestamp(1614698784, 1),
			"t" : NumberLong(1)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1614698784, 1),
			"t" : NumberLong(1)
		},
		"optimeDate" : ISODate("2021-03-02T15:26:24Z"),
		"optimeDurableDate" : ISODate("2021-03-02T15:26:24Z"),
		"lastHeartbeat" : ISODate("2021-03-02T15:26:31.022Z"),
		"lastHeartbeatRecv" : ISODate("2021-03-02T15:26:31.224Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncingTo" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceHost" : "mg-replicaset-0.mg-replicaset-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 0,
		"infoMessage" : "",
		"configVersion" : 5
	}
]
```

From all the above outputs we can see that the replicas of the replicaset is `3`. That means we have successfully scaled down the replicas of the MongoDB replicaset.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete mongodbopsrequest -n demo mops-vscale-replicaset
```