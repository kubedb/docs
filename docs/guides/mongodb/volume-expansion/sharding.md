---
title: MongoDB Sharded Database Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: mg-volume-expansion-shard
    name: Sharding
    parent: mg-volume-expansion
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# MongoDB Sharded Database Volume Expansion

This guide will show you how to use `KubeDB` Enterprise operator to expand the volume of a MongoDB Sharded Database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here]().

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/concepts/databases/mongodb.md)
  - [Sharding](/docs/guides/mongodb/clustering/sharding.md)
  - [MongoDBOpsRequest](/docs/concepts/day-2-operations/mongodbopsrequest.md)
  - [Volume Expansion Overview](/docs/guides/mongodb/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Expand Volume of Sharded Database

Here, we are going to deploy a `MongoDB` Sharded Database using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBOpsRequest` to expand the volume of shard nodes and config servers.

### Prepare MongoDB Sharded Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```console
$ kubectl get storageclass                                                                                                                                           20:22:33
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `MongoDB` standalone database with version `3.6.8`.

### Deploy MongoDB

In this section, we are going to deploy a MongoDB Sharded database with 1GB volume for each of the shard nodes and config servers. Then, in the next sections we will expand the volume of shard nodes and config servers to 2GB using `MongoDBOpsRequest` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/volume-expansion/mg-shard.yaml
mongodb.kubedb.com/mg-sharding created
```

Now, wait until `mg-sharding` has status `Running`. i.e,

```console
$ kubectl get mg -n demo
NAME          VERSION    STATUS    AGE
mg-sharding   3.6.8-v1   Running   2m45s
```

Let's check volume size from statefulset, and from the persistent volume of shards and config servers,

```console
$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                  STORAGECLASS   REASON   AGE
pvc-194f6e9c-b9a7-4d00-a125-a6c01273468c   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard0-0      standard                68s
pvc-390b6343-f97e-4761-a516-e3c9607c55d6   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard1-1      standard                2m26s
pvc-51ab98e8-d468-4a74-b176-3853dada41c2   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-configsvr-1   standard                2m33s
pvc-5209095e-561f-4601-a0bf-0c705234da5b   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard1-0      standard                3m6s
pvc-5be2ab13-e12c-4053-8680-7c5588dff8eb   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard2-1      standard                2m32s
pvc-7e11502d-13e0-4a84-9ebe-29bc2b15f026   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard0-1      standard                44s
pvc-7e20906c-462d-47b7-b4cf-ba0ef69ba26e   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard2-0      standard                3m7s
pvc-87634059-0f95-4595-ae8a-121944961103   1Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-configsvr-0   standard                3m7s
```

You can see the statefulsets have 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `MongoDBOpsRequest` CR to expand the volume of this database.

### Volume Expansion of Shard and ConfigServer Nodes

Here, we are going to expand the volume of the shard and configServer nodes of the database.

#### Create MongoDBOpsRequest

In order to expand the volume of the shard nodes of the database, we have to create a `MongoDBOpsRequest` CR with our desired volume size. Below is the YAML of the `MongoDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-volume-exp-shard
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: mg-sharding
  volumeExpansion:
    shard: 2Gi
    configServer: 2Gi
```

Here,
- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `mops-volume-exp-shard` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.shard` specifies the desired volume size of shard nodes.
- `spec.volumeExpansion.configServer` specifies the desired volume size of configServer nodes.

> **Note:** If you don't want to expand the volume of all the components together, you can only specify the components (shard and configServer) that you want to expand.

Let's create the `MongoDBOpsRequest` CR we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/volume-expansion/mops-volume-exp-shard.yaml
mongodbopsrequest.ops.kubedb.com/mops-volume-exp-shard created
```

#### Verify MongoDB shard volumes expanded successfully

If everything goes well, `KubeDB` Enterprise operator will update the volume size of `MongoDB` object and related `StatefulSets` and `Persistent Volumes`.

Let's wait for `MongoDBOpsRequest` to be `Successful`. Run the following command to watch `MongoDBOpsRequest` CR,

```console
$ kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                    TYPE              STATUS       AGE
mops-volume-exp-shard   VolumeExpansion   Successful   3m49s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```console
$ kubectl describe mongodbopsrequest -n demo mops-volume-exp-shard
Name:         mops-volume-exp-shard
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2020-09-30T04:24:37Z
  Generation:          1
  Resource Version:    140791
  Self Link:           /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-volume-exp-shard
  UID:                 fc23a0a2-3a48-4b76-95c5-121f3d56df78
Spec:
  Database Ref:
    Name:  mg-sharding
  Type:    VolumeExpansion
  Volume Expansion:
    Config Server:  2Gi
    Shard:          2Gi
Status:
  Conditions:
    Last Transition Time:  2020-09-30T04:25:48Z
    Message:               MongoDB ops request is expanding volume of database
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2020-09-30T04:26:58Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                ConfigServerVolumeExpansion
    Status:                True
    Type:                  ConfigServerVolumeExpansion
    Last Transition Time:  2020-09-30T04:29:28Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                ShardVolumeExpansion
    Status:                True
    Type:                  ShardVolumeExpansion
    Last Transition Time:  2020-09-30T04:29:33Z
    Message:               Successfully Resumed mongodb: mg-sharding
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-09-30T04:29:33Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                       Age    From                        Message
  ----    ------                       ----   ----                        -------
  Normal  ConfigServerVolumeExpansion  3m25s  KubeDB Enterprise Operator  Successfully Expanded Volume
  Normal  ShardVolumeExpansion         55s    KubeDB Enterprise Operator  Successfully Expanded Volume
  Normal  ResumeDatabase               50s    KubeDB Enterprise Operator  Resuming MongoDB
  Normal  ResumeDatabase               50s    KubeDB Enterprise Operator  Successfully Resumed mongodb
  Normal  Successful                   50s    KubeDB Enterprise Operator  Successfully Expanded Volume
```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```console
$ kubectl get sts -n demo mg-sharding-configsvr -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get sts -n demo mg-sharding-shard0 -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'                                             10:15:51
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                  STORAGECLASS   REASON   AGE
pvc-194f6e9c-b9a7-4d00-a125-a6c01273468c   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard0-0      standard                3m38s
pvc-390b6343-f97e-4761-a516-e3c9607c55d6   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard1-1      standard                4m56s
pvc-51ab98e8-d468-4a74-b176-3853dada41c2   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-configsvr-1   standard                5m3s
pvc-5209095e-561f-4601-a0bf-0c705234da5b   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard1-0      standard                5m36s
pvc-5be2ab13-e12c-4053-8680-7c5588dff8eb   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard2-1      standard                5m2s
pvc-7e11502d-13e0-4a84-9ebe-29bc2b15f026   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard0-1      standard                3m14s
pvc-7e20906c-462d-47b7-b4cf-ba0ef69ba26e   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-shard2-0      standard                5m37s
pvc-87634059-0f95-4595-ae8a-121944961103   2Gi        RWO            Delete           Bound    demo/datadir-mg-sharding-configsvr-0   standard                5m37s
```

The above output verifies that we have successfully expanded the volume of the shard nodes and configServer nodes of the MongoDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete mg -n demo mg-sharding
kubectl delete mongodbopsrequest -n demo mops-volume-exp-shard mops-volume-exp-configserver
```
