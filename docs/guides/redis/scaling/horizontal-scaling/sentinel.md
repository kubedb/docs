---
title: Horizontal Scaling Redis Sentinel
menu:
  docs_{{ .version }}:
    identifier: rd-horizontal-scaling-sentinel
    name: Sentinel
    parent: rd-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Horizontal Scale of Redis Sentinel

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Redis` database and `RedisSentinel` instance.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisSentinel](/docs/guides/redis/concepts/redissentinel.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/redis/scaling/horizontal-scaling/overview.md).

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Redis Sentinel Database

Now, we are going to deploy a `RedisSentinel` instance with version `6.2.7` and a `Redis` database with version `6.2.5`. Then, in the next section we are going to apply horizontal scaling on the sentinel and the database using `RedisSentinelOpsRequest` and `RedisOpsRequest` CRD.

### Deploy RedisSentinel :

In this section, we are going to deploy a `RedisSentinel` instance. Below is the YAML of the `RedisSentinel` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
metadata:
  name: sen-sample
  namespace: demo
spec:
  version: 6.2.7
  replicas: 5
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  terminationPolicy: DoNotTerminate
```

Let's create the `RedisSentinel` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/horizontal-scaling/sentinel.yaml
redissentinel.kubedb.com/sen-sample created
```

Now, wait until `sen-sample` created has status `Ready`. i.e,

```bash
$ kubectl get redissentinel -n demo
NAME         VERSION   STATUS   AGE
sen-sample   6.2.7     Ready    5m20s
```

Let's check the number of replicas this sentinel has from the RedisSentinel object

```bash
$ kubectl get redissentinel -n demo sen-sample -o json | jq '.spec.replicas'
5
```

### Deploy Redis :

In this section, we are going to deploy a `Redis` instance which will be monitored by previously created `sen-sample`. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-sample
  namespace: demo
spec:
  version: 6.2.5
  replicas: 3
  sentinelRef:
    name: sen-sample
    namespace: demo
  mode: Sentinel
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  terminationPolicy: DoNotTerminate
```

Let's create the `Redis` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/horizontal-scaling/rd-sentinel.yaml
redis.kubedb.com/rd-sample created
```

Now, wait until `rd-sample` created has status `Ready`. i.e,

```bash
$ kubectl get redis -n demo
NAME        VERSION   STATUS   AGE
rd-sample   6.2.5     Ready    2m11s
```
Let's check the Pod containers resources,
```bash
$ kubectl get redis -n demo rd-sample -o json | jq '.spec.replicas'
3
```

Now let's connect to redis with redis-cli to check the replication configuration
```bash
$ kubectl exec -it -n demo rd-sample-0 -c redis -- redis-cli info replication
# Replication
role:master
connected_slaves:2
slave0:ip=rd-sample-1.rd-sample-pods.demo.svc,port=6379,state=online,offset=35478,lag=0
slave1:ip=rd-sample-2.rd-sample-pods.demo.svc,port=6379,state=online,offset=35478,lag=0
master_failover_state:no-failover
master_replid:4ac5cc7292e84c6d1b69d3732869557f2854db2d
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:35492
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset:1
repl_backlog_histlen:35492
```

Additionally, the sentinel monitoring can be checked with following command : 
```bash
kubectl exec -it -n demo sen-sample-0 -c redissentinel -- redis-cli -p 26379 sentinel masters
```

We are now ready to apply the `RedisSentinelOpsRequest` CR to horizontal scale on sentinel and `RedisOpsRequest` CR to horizontal scale database.

### Horizontal Scale RedisSentinel

Here, we are going to scale down the replicas count of the sentinel to meet the desired resources after scaling.

#### Create RedisSentinelOpsRequest:

In order to scale the replicas of the sentinel, we have to create a `RedisSentinelOpsRequest` CR with our desired number of replicas. Below is the YAML of the `RedisSentinelOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: sen-ops-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sen-sample
  horizontalScaling:
    replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `sen-sample` RedisSentinel instance.
- `spec.type` specifies that we are going to perform `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.

Let's create the `RedisSentinelOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/horizontal-scaling/horizontal-sentinel.yaml
redissentinelopsrequest.ops.kubedb.com/sen-ops-horizontal created
```

#### Verify RedisSentinel replicas updated successfully :

If everything goes well, `KubeDB` Enterprise operator will scale down the replicas of `RedisSentinel` object.

Let's wait for `RedisSentinelOpsRequest` to be `Successful`.  Run the following command to watch `RedisSentinelOpsRequest` CR,

```bash
$ watch kubectl get redissentinelopsrequest -n demo
Every 2.0s: kubectl get redissentinelopsrequest -n demo
NAME                 TYPE              STATUS       AGE
sen-ops-horizontal   HorizontalScaling   Successful   5m27s
```

We can see from the above output that the `RedisSentinelOpsRequest` has succeeded.

Let's check the number of master and replicas this database has from the RedisSentinel object

```bash
$ kubectl get redissentinel -n demo sen-sample -o json | jq '.spec.replicas'
3
```

The above output verifies that we have successfully scaled up the resources of the sentinel instance.
### Horizontal Scale Redis

Here, we are going to update the resources of the redis database to meet the desired resources after scaling.

#### Create RedisOpsRequest:

In order to scale the replicas of the redis database, we have to create a `RedisOpsRequest` CR with our desired number of replicas. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-ops-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: rd-sample
  horizontalScaling:
    replicas: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `rd-sample` Redis database.
- `spec.type` specifies that we are going to perform `HorizontalScaling` on our database.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/horizontal-scaling//horizontal-redis-sentinel.yaml
redisopsrequest.ops.kubedb.com/rd-ops-horizontal created
```

#### Verify Redis resources updated successfully :

If everything goes well, `KubeDB` Enterprise operator will scale up the replicas of `Redis` object.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
NAME                TYPE                STATUS       AGE
rd-ops-horizontal   HorizontalScaling   Successful   4m4s
```

We can see from the above output that the `RedisOpsRequest` has succeeded.
Now, we are going to verify if the number of replicas the redis sentinel has updated to meet up the desired state, Let's check,

```bash
$ kubectl get redis -n demo rd-sample -o json | jq '.spec.replicas'
5
```

Now let's connect to redis with redis-cli to check the replication configuration
```bash
$ kubectl exec -it -n demo rd-sample-0 -c redis -- redis-cli info replication
# Replication
role:master
connected_slaves:4
slave0:ip=rd-sample-1.rd-sample-pods.demo.svc,port=6379,state=online,offset=325651,lag=1
slave1:ip=rd-sample-2.rd-sample-pods.demo.svc,port=6379,state=online,offset=325651,lag=1
slave2:ip=rd-sample-3.rd-sample-pods.demo.svc,port=6379,state=online,offset=325651,lag=1
slave3:ip=rd-sample-4.rd-sample-pods.demo.svc,port=6379,state=online,offset=325651,lag=1
master_failover_state:no-failover
master_replid:4871c4756eebbadc7f2c56a4dd1dff11e20a04ba
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:325651
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset:1
repl_backlog_histlen:325651
```

The above output verifies that we have successfully scaled up the resources of the redis database. There are 1 master and 4 connected slaves. So, the Ops Request
scaled up the replicas to 5.

Additionally, the sentinel monitoring can be checked with following command :
```bash
kubectl exec -it -n demo sen-sample-0 -c redissentinel -- redis-cli -p 26379 sentinel masters
```

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
# Delete Redis and RedisOpsRequest
$ kubectl patch -n demo rd/rd-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-sample patched

$ kubectl delete -n demo redis rd-sample
redis.kubedb.com "rd-sample" deleted

$ kubectl delete -n demo redisopsrequest rd-ops-horizontal 
redisopsrequest.ops.kubedb.com "rd-ops-horizontal" deleted

# Delete RedisSentinel and RedisSentinelOpsRequest
$ kubectl patch -n demo redissentinel/sen-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-sample patched

$ kubectl delete -n demo redissentinel sen-sample
redissentinel.kubedb.com "sen-sample" deleted

$ kubectl delete -n demo redissentinelopsrequests sen-ops-horizontal 
redissentinelopsrequest.ops.kubedb.com "sen-ops-horizontal" deleted
```
