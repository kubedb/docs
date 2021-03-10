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

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MongoDB Replicaset

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a MongoDB replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [Replicaset](/docs/guides/mongodb/clustering/replicaset.md) 
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/mongodb/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Sharded Database

Here, we are going to deploy a  `MongoDB` sharded database using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare MongoDB Sharded Database

Now, we are going to deploy a `MongoDB` sharded database with version `4.2.3`.

### Deploy MongoDB Sharded Database 

In this section, we are going to deploy a MongoDB sharded database. Then, in the next sections we will update the resources of various components (mongos, shard, configserver etc.) of the database using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,
    
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-sharding
  namespace: demo
spec:
  version: 4.2.3
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
mg-sharding   4.2.3      Ready     8m51s
```

Let's check the Pod containers resources of various components (mongos, shard, configserver etc.) of the database,

```bash
$ kubectl get pod -n demo mg-sharding-mongos-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}

$ kubectl get pod -n demo mg-sharding-configsvr-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}

$ kubectl get pod -n demo mg-sharding-shard0-0 -o json | jq '.spec.containers[].resources'                                                                      
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see all the Pod of mongos, configserver and shard has default resources which is assigned by Kubedb operator.

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
        memory: "1100Mi"
        cpu: "0.55"
      limits:
        memory: "1100Mi"
        cpu: "0.55"
    configServer:
      requests:
        memory: "1100Mi"
        cpu: "0.55"
      limits:
        memory: "1100Mi"
        cpu: "0.55"
    mongos:
      requests:
        memory: "1100Mi"
        cpu: "0.55"
      limits:
        memory: "1100Mi"
        cpu: "0.55"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mops-vscale-shard` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.shard` specifies the desired resources after scaling for the shard nodes.
- `spec.VerticalScaling.configServer` specifies the desired resources after scaling for the configServer nodes.
- `spec.VerticalScaling.mongos` specifies the desired resources after scaling for the mongos nodes.

> **Note:** If you don't want to scale all the components together, you can only specify the components (shard, configServer and mongos) that you want to scale.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/scaling/vertical-scaling/mops-vscale-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-vscale-shard created
```

#### Verify MongoDB Shard resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `MongoDB` object and related `StatefulSets` and `Pods` of shard nodes.

Let's wait for `MongoDBOpsRequest` to be `Successful`.  Run the following command to watch `MongoDBOpsRequest` CR,

```bash
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS       AGE
mops-vscale-shard   VerticalScaling   Successful   8m21s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-vscale-shard
Name:         mops-vscale-shard
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T17:10:43Z
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
    Time:         2021-03-02T17:10:43Z
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
    Time:            2021-03-02T17:10:43Z
  Resource Version:  157614
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-vscale-shard
  UID:               6354eabe-0008-4ef7-8ec7-30ce58985012
Spec:
  Database Ref:
    Name:  mg-sharding
  Type:    VerticalScaling
  Vertical Scaling:
    Config Server:
      Limits:
        Cpu:     0.55
        Memory:  1100Mi
      Requests:
        Cpu:     0.55
        Memory:  1100Mi
    Mongos:
      Limits:
        Cpu:     0.55
        Memory:  1100Mi
      Requests:
        Cpu:     0.55
        Memory:  1100Mi
    Shard:
      Limits:
        Cpu:     0.55
        Memory:  1100Mi
      Requests:
        Cpu:     0.55
        Memory:  1100Mi
Status:
  Conditions:
    Last Transition Time:  2021-03-02T17:10:43Z
    Message:               MongoDB ops request is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2021-03-02T17:10:43Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2021-03-02T17:12:39Z
    Message:               Successfully Vertically Scaled ConfigServer Resources
    Observed Generation:   1
    Reason:                UpdateConfigServerResources
    Status:                True
    Type:                  UpdateConfigServerResources
    Last Transition Time:  2021-03-02T17:16:33Z
    Message:               Successfully Vertically Scaled Shard Resources
    Observed Generation:   1
    Reason:                UpdateShardResources
    Status:                True
    Type:                  UpdateShardResources
    Last Transition Time:  2021-03-02T17:17:44Z
    Message:               Successfully Vertically Scaled Mongos Resources
    Observed Generation:   1
    Reason:                UpdateMongosResources
    Status:                True
    Type:                  UpdateMongosResources
    Last Transition Time:  2021-03-02T17:17:44Z
    Message:               Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                       Age    From                        Message
  ----    ------                       ----   ----                        -------
  Normal  UpdateStatefulSetResources   7m9s   KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  PauseDatabase                7m9s   KubeDB Enterprise Operator  Successfully paused MongoDB demo/mg-sharding
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-mongos
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-configsvr
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-shard0
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-shard1
  Normal  UpdateStatefulSetResources   7m9s   KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-mongos
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-configsvr
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-shard0
  Normal  Starting                     7m9s   KubeDB Enterprise Operator  Updating Resources of StatefulSet: mg-sharding-shard1
  Normal  PauseDatabase                7m9s   KubeDB Enterprise Operator  Pausing MongoDB demo/mg-sharding
  Normal  UpdateConfigServerResources  5m13s  KubeDB Enterprise Operator  Successfully Vertically Scaled ConfigServer Resources
  Normal  UpdateShardResources         79s    KubeDB Enterprise Operator  Successfully Vertically Scaled Shard Resources
  Normal  UpdateMongosResources        8s     KubeDB Enterprise Operator  Successfully Vertically Scaled Mongos Resources
  Normal  ResumeDatabase               8s     KubeDB Enterprise Operator  Resuming MongoDB demo/mg-sharding
  Normal  ResumeDatabase               8s     KubeDB Enterprise Operator  Successfully resumed MongoDB demo/mg-sharding
  Normal  Successful                   8s     KubeDB Enterprise Operator  Successfully Vertically Scaled Database
```

Now, we are going to verify from one of the Pod yaml whether the resources of the shard nodes has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo mg-sharding-shard0-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "550m",
    "memory": "1100Mi"
  },
  "requests": {
    "cpu": "550m",
    "memory": "1100Mi"
  }
}

$ kubectl get pod -n demo mg-sharding-configsvr-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "550m",
    "memory": "1100Mi"
  },
  "requests": {
    "cpu": "550m",
    "memory": "1100Mi"
  }
}

$ kubectl get pod -n demo mg-sharding-mongos-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "550m",
    "memory": "1100Mi"
  },
  "requests": {
    "cpu": "550m",
    "memory": "1100Mi"
  }
}
```

The above output verifies that we have successfully scaled the resources of all components of the MongoDB sharded database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-shard
kubectl delete mongodbopsrequest -n demo mops-vscale-shard
```