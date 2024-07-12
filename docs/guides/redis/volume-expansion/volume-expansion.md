---
title: Redis Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: redis-volume-expansion-volume-expansion
    name: Redis Volume Expansion
    parent: rd-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis Volume Expansion

This guide will show you how to use `KubeDB` Enterprise operator to expand the volume of a Redis.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Volume Expansion Overview](/docs/guides/redis/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Expand Volume of Redis

Here, we are going to deploy a  `Redis` cluster using a supported version by `KubeDB` operator. Then we are going to apply `RedisOpsRequest` to expand its volume. The process of expanding Redis `standalone` is same as Redis cluster.

### Prepare Redis Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  69s
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   37s

```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We will use this storage class. You can install topolvm from [here](https://github.com/topolvm/topolvm).

Now, we are going to deploy a `Redis` database with in `Cluster` Mode version `6.2.14`.

### Deploy Redis

In this section, we are going to deploy a Redis Cluster with 1GB volume. Then, in the next section we will expand its volume to 2GB using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: 6.2.14
  mode: Cluster
  cluster:
    shards: 3
    replicas: 1
  storageType: Durable
  storage:
    storageClassName: "topolvm-provisioner"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Halt
```

Let's create the `Redis` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/example/redis/volume-expansion/sample-redis.yaml
redis.kubedb.com/sample-redis created
```

Now, wait until `sample-redis` has status `Ready`. i.e,

```bash
$ kubectl get redis -n demo
NAME             VERSION   STATUS   AGE
sample-redis     6.2.14    Ready    5m4s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo sample-redis-shard0 -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                             STORAGECLASS              REASON   AGE
pvc-032f1355-1720-4d85-b1e5-b86427bc4662   1Gi        RWO            Delete           Bound    demo/data-sample-redis-shard0-1   topolvm-provisioner                2m49s
pvc-207ac9aa-2ba2-432b-ac00-8cc1cd46e20a   1Gi        RWO            Delete           Bound    demo/data-sample-redis-shard2-0   topolvm-provisioner                2m49s
pvc-20c946e4-4812-4dfc-a76e-4629bcd385dc   1Gi        RWO            Delete           Bound    demo/data-sample-redis-shard2-1   topolvm-provisioner                2m38s
pvc-69158d05-c715-4dd5-afee-2f5d196ba1f9   1Gi        RWO            Delete           Bound    demo/data-sample-redis-shard1-0   topolvm-provisioner                2m53s
pvc-aee29446-eff0-430e-95ff-ae853e73a244   1Gi        RWO            Delete           Bound    demo/data-sample-redis-shard1-1   topolvm-provisioner                2m41s
pvc-d37fbdf9-90bd-4b5e-b3b2-7e40156c13a8   1Gi        RWO            Delete           Bound    demo/data-sample-redis-shard0-0   topolvm-provisioner                2m56s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `RedisOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the Redis cluster.

#### Create RedisOpsRequest

In order to expand the volume of the database, we have to create a `RedisOpsRequest` CR with our desired volume size. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-online-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: sample-redis
  volumeExpansion:   
    mode: "Online"
    redis: 2Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `sample-redis` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.redis` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`). Storageclass `topolvm-provisioner` supports `Online` volume expansion.

> **Note:** If the Storageclass you are using doesn't support `Online` Volume Expansion, Try offline volume expansion by using `spec.volumeExpansion.mode:"Offline"`.

During `Online` VolumeExpansion KubeDB expands volume without pausing database object, it directly updates the underlying PVC. And for Offline volume expansion, the database is paused. The Pods 
are deleted and PVC is updated. Then the database Pods are recreated with updated PVC.


Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/example/redis/volume-expansion/online-vol-expansion.yaml
redisopsrequest.ops.kubedb.com/rd-online-volume-expansion created
```

#### Verify Redis volume expanded successfully

If everything goes well, `KubeDB` Enterprise operator will update the volume size of `Redis` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ kubectl get redisopsrequest -n demo
NAME                         TYPE              STATUS       AGE
rd-online-volume-expansion   VolumeExpansion   Successful   96s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo sample-redis-shard0 -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get sts -n demo sample-redis-shard1 -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                             STORAGECLASS              REASON   AGE
pvc-032f1355-1720-4d85-b1e5-b86427bc4662   2Gi        RWO            Delete           Bound    demo/data-sample-redis-shard0-1   topolvm-provisioner                7m9s
pvc-207ac9aa-2ba2-432b-ac00-8cc1cd46e20a   2Gi        RWO            Delete           Bound    demo/data-sample-redis-shard2-0   topolvm-provisioner                7m9s
pvc-20c946e4-4812-4dfc-a76e-4629bcd385dc   2Gi        RWO            Delete           Bound    demo/data-sample-redis-shard2-1   topolvm-provisioner                7m8s
pvc-69158d05-c715-4dd5-afee-2f5d196ba1f9   2Gi        RWO            Delete           Bound    demo/data-sample-redis-shard1-0   topolvm-provisioner                7m3s
pvc-aee29446-eff0-430e-95ff-ae853e73a244   2Gi        RWO            Delete           Bound    demo/data-sample-redis-shard1-1   topolvm-provisioner                7m1s
pvc-d37fbdf9-90bd-4b5e-b3b2-7e40156c13a8   2Gi        RWO            Delete           Bound    demo/data-sample-redis-shard0-0   topolvm-provisioner                7m6s
```

The above output verifies that we have successfully expanded the volume of the Redis database.

## Standalone Mode and Sentinel Mode

The volume expansion process is same for all the Redis modes. The `RedisOpsRequest` CR has the sample fields. The database needs to refer to a redis database 
in standalone or sentinel mode.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete redis -n demo sample-redis
$ kubectl delete redisopsrequest -n demo rd-online-volume-expansion
```

```bash
$ kubectl patch -n demo rd/sample-redis -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/sample-redis patched

$ kubectl delete -n demo redis sample-redis
redis.kubedb.com "sample-redis" deleted

$ kubectl delete -n demo redisopsrequest rd-online-volume-expansion
redisopsrequest.ops.kubedb.com "rd-online-volume-expansion" deleted
```
