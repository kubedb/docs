---
title: Vertical Scaling Sharded MongoDB Cluster
menu:
  docs_{{ .version }}:
    identifier: mg-vertical-scaling-shard
    name: Sharding
    parent: mg-vertical-scaling
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MongoDB Replicaset

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a MongoDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md) 
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/mongodb/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Sharded Database

Here, we are going to deploy a  `MongoDB` sharded database using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare MongoDB Sharded Database

Now, we are going to deploy a `MongoDB` sharded database with version `3.6.8`.

### Deploy MongoDB Sharded Database 

In this section, we are going to deploy a MongoDB sharded database. Then, in the next sections we will update the resources of various components (mongos, shard, configserver etc.) of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,
    
```yaml
apiVersion: kubedb.com/v1alpha2
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
mg-sharding   3.6.8-v1   Running   8m51s
```

Let's check the Pod containers resources of various components (mongos, shard, configserver etc.) of the database,

```console
$ kubectl get pod -n demo mg-sharding-mongos-0 -o json | jq '.spec.containers[].resources'
  {}

$ kubectl get pod -n demo mg-sharding-configsvr-0 -o json | jq '.spec.containers[].resources'
  {}

$ kubectl get pod -n demo mg-sharding-shard0-0 -o json | jq '.spec.containers[].resources'                                                                      
  {}
```

You can see all the Pod of mongos, configserver and shard has empty resources that means the scheduler will choose a random node to place the container of the Pod on by default.

We are now ready to apply the `MongoDBOpsRequest` CR to update the resources of mongos, configserver and shard nodes of this database.

## Vertical Scaling of Shard

Here, we are going to update the resources of the shard of the database to meet the desired resources after scaling.

#### Create MongoDBOpsRequest for shard

In order to update the resources of the shard nodes, we have to create a `MongoDBOpsRequest` CR with our desired resources. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-vscale-shard
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-sharding
  verticalScaling:
    shard:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
    configServer:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
    mongos:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mops-vscale-shard` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.shard` specifies the desired resources after scaling for the shard nodes.
- `spec.VerticalScaling.configServer` specifies the desired resources after scaling for the configServer nodes.
- `spec.VerticalScaling.mongos` specifies the desired resources after scaling for the mongos nodes.

> **Note:** If you don't want to scale all the components together, you can only specify the components (shard, configServer and mongos) that you want to scale.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/vertical-scaling/mops-vscale-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-vscale-shard created
```

#### Verify MongoDB Shard resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `MongoDB` object and related `StatefulSets` and `Pods` of shard nodes.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS       AGE
mops-vscale-shard   VerticalScaling   Successful   8m21s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-vscale-shard
Name:         mops-vscale-shard
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2020-09-21T04:51:47Z
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
        f:type:
        f:verticalScaling:
          .:
          f:configServer:
            .:
            f:limits:
              .:
              f:memory:
            f:requests:
              .:
              f:memory:
          f:mongos:
            .:
            f:limits:
              .:
              f:memory:
            f:requests:
              .:
              f:memory:
          f:shard:
            .:
            f:limits:
              .:
              f:memory:
            f:requests:
              .:
              f:memory:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2020-09-21T04:51:47Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:verticalScaling:
          f:configServer:
            f:limits:
              f:cpu:
            f:requests:
              f:cpu:
          f:mongos:
            f:limits:
              f:cpu:
            f:requests:
              f:cpu:
          f:shard:
            f:limits:
              f:cpu:
            f:requests:
              f:cpu:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2020-09-21T05:01:30Z
  Resource Version:  680319
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-vscale-shard
  UID:               bd3967d6-9671-4e0f-a1df-67622dac1579
Spec:
  Database Ref:
    Name:  mg-sharding
  Type:    VerticalScaling
  Vertical Scaling:
    Config Server:
      Limits:
        Cpu:     0.2
        Memory:  250Mi
      Requests:
        Cpu:     0.1
        Memory:  150Mi
    Mongos:
      Limits:
        Cpu:     0.2
        Memory:  250Mi
      Requests:
        Cpu:     0.1
        Memory:  150Mi
    Shard:
      Limits:
        Cpu:     0.2
        Memory:  250Mi
      Requests:
        Cpu:     0.1
        Memory:  150Mi
Status:
  Conditions:
    Last Transition Time:  2020-09-21T04:51:47Z
    Message:               MongoDB ops request is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2020-09-21T04:51:47Z
    Message:               Successfully paused mongodb: mg-sharding
    Observed Generation:   1
    Reason:                PauseDatabase
    Status:                True
    Type:                  PauseDatabase
    Last Transition Time:  2020-09-21T04:51:47Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2020-09-21T04:53:27Z
    Message:               Successfully Vertically Scaled ConfigServer Resources
    Observed Generation:   1
    Reason:                UpdateConfigServerResources
    Status:                True
    Type:                  UpdateConfigServerResources
    Last Transition Time:  2020-09-21T05:00:25Z
    Message:               Successfully Vertically Scaled Shard Resources
    Observed Generation:   1
    Reason:                UpdateShardResources
    Status:                True
    Type:                  UpdateShardResources
    Last Transition Time:  2020-09-21T05:01:30Z
    Message:               Successfully Vertically Scaled Mongos Resources
    Observed Generation:   1
    Reason:                UpdateMongosResources
    Status:                True
    Type:                  UpdateMongosResources
    Last Transition Time:  2020-09-21T05:01:30Z
    Message:               Successfully Resumed mongodb: mg-sharding
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-09-21T05:01:30Z
    Message:               Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                       Age   From                        Message
  ----    ------                       ----  ----                        -------
  Normal  UpdateStatefulSetResources   29m   KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  PauseDatabase                29m   KubeDB Enterprise Operator  Successfully Paused MongoDB mg-sharding in Namespace demo
  Normal  Starting                     29m   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-mongos
  Normal  Starting                     29m   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-configsvr
  Normal  Starting                     29m   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-shard0
  Normal  Starting                     29m   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-shard1
  Normal  Starting                     29m   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-shard2
  Normal  PauseDatabase                29m   KubeDB Enterprise Operator  Pausing MongoDB mg-sharding in Namespace demo
  Normal  UpdateConfigServerResources  28m   KubeDB Enterprise Operator  Successfully Vertically Scaled ConfigServer Resources
  Normal  UpdateShardResources         21m   KubeDB Enterprise Operator  Successfully Vertically Scaled Shard Resources
  Normal  UpdateMongosResources        20m   KubeDB Enterprise Operator  Successfully Vertically Scaled Mongos Resources
  Normal  ResumeDatabase               20m   KubeDB Enterprise Operator  Resuming MongoDB
  Normal  ResumeDatabase               20m   KubeDB Enterprise Operator  Successfully Resumed mongodb
  Normal  Successful                   20m   KubeDB Enterprise Operator  Successfully Vertically Scaled Database
  Normal  Successful                   20m   KubeDB Enterprise Operator  Successfully Vertically Scaled Database
```

Now, we are going to verify from one of the Pod yaml whether the resources of the shard nodes has updated to meet up the desired state, Let's check,

```console
$ kubectl get pod -n demo mg-sharding-shard0-0 -o json | jq '.spec.containers[].resources'                                                                           12:56:06
  {
    "limits": {
      "cpu": "200m",
      "memory": "250Mi"
    },
    "requests": {
      "cpu": "100m",
      "memory": "150Mi"
    }
  }

$ kubectl get pod -n demo mg-sharding-configsvr-0 -o json | jq '.spec.containers[].resources'
  {
    "limits": {
      "cpu": "200m",
      "memory": "250Mi"
    },
    "requests": {
      "cpu": "100m",
      "memory": "150Mi"
    }
  }

$ kubectl get pod -n demo mg-sharding-mongos-0 -o json | jq '.spec.containers[].resources'
  {
    "limits": {
      "cpu": "200m",
      "memory": "250Mi"
    },
    "requests": {
      "cpu": "100m",
      "memory": "150Mi"
    }
  }
```

The above output verifies that we have successfully scaled the resources of all components of the MongoDB sharded database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-shard
kubectl delete mongodbopsrequest -n demo mops-vscale-shard
```