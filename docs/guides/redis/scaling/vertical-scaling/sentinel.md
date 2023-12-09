---
title: Vertical Scaling Sentinel Redis
menu:
  docs_{{ .version }}:
    identifier: rd-vertical-scaling-sentinel
    name: Sentinel
    parent: rd-vertical-scaling
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale of Redis Sentinel

This guide will show you how to use `KubeDB` Enterprise operator to perform vertical scaling of `Redis` in Sentinel mode and `RedisSentinel`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisSentinel](/docs/guides/redis/concepts/redissentinel.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/redis/scaling/vertical-scaling/overview.md).

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Redis Sentinel Database

Now, we are going to deploy a `RedisSentinel` instance with version `6.2.7` and a `Redis` database with version `6.2.5`. Then, in the next section we are going to apply vertical scaling on the sentinel and the database using `RedisOpsRequest` CRD

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
  replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "100m"
          memory: "100Mi"
  terminationPolicy: DoNotTerminate
```

Let's create the `RedisSentinel` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-scaling/sentinel.yaml
redissentinel.kubedb.com/sen-sample created
```

Now, wait until `sen-sample` created has status `Ready`. i.e,

```bash
$ kubectl get redissentinel -n demo
NAME         VERSION   STATUS   AGE
sen-sample   6.2.7     Ready    5m20s
```

Let's check the Pod containers resources,
```bash
$ kubectl get pod -n demo sen-sample-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "100Mi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "100Mi"
  }
}
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
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "100m"
          memory: "100Mi"
  terminationPolicy: DoNotTerminate
```

Let's create the `Redis` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-scaling/rd-sentinel.yaml
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
$ kubectl get pod -n demo rd-sample-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "100Mi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "100Mi"
  }
}
```

We are now ready to apply the `RedisSentinelOpsRequest` CR to vertical scale on sentinel and `RedisOpsRequest` CR to vertical scale database.

### Vertical Scale RedisSentinel

Here, we are going to update the resources of the sentinel to meet the desired resources after scaling.

#### Create RedisSentinelOpsRequest:

In order to update the resources of the sentinel, we have to create a `RedisSentinelOpsRequest` CR with our desired resources. Below is the YAML of the `RedisSentinelOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: sen-ops-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: sen-sample
  verticalScaling:
    redissentinel:
      requests:
        memory: "300Mi"
        cpu: "200m"
      limits:
        memory: "800Mi"
        cpu: "500m"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `sen-sample` RedisSentinel instance.
- `spec.type` specifies that we are going to perform `VerticalScaling` on our database.
- `spec.verticalScaling.redissentinel` specifies the desired resources after scaling.

Let's create the `RedisSentinelOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-scaling/vertical-sentinel.yaml
redissentinelopsrequest.ops.kubedb.com/sen-ops-vertical created
```

#### Verify RedisSentinel resources updated successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `RedisSentinel` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisSentinelOpsRequest` to be `Successful`.  Run the following command to watch `RedisSentinelOpsRequest` CR,

```bash
$ watch kubectl get redissentinelopsrequest -n demo
Every 2.0s: kubectl get redissentinelopsrequest -n demo
NAME               TYPE              STATUS       AGE
sen-ops-vertical   VerticalScaling   Successful   5m27s
```

We can see from the above output that the `RedisSentinelOpsRequest` has succeeded.

Now, we are going to verify from the Pod yaml whether the resources of the sentinel has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sen-sample-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "800Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the sentinel instance.
### Vertical Scale Redis

Here, we are going to update the resources of the redis database to meet the desired resources after scaling.

#### Create RedisOpsRequest:

In order to update the resources of the database, we have to create a `RedisOpsRequest` CR with our desired resources. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rd-ops-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: rd-sample
  verticalScaling:
    redis:
      requests:
        memory: "300Mi"
        cpu: "200m"
      limits:
        memory: "800Mi"
        cpu: "500m"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `rd-sample` Redis database.
- `spec.type` specifies that we are going to perform `VerticalScaling` on our database.
- `spec.VerticalScaling.redis` specifies the desired resources after scaling.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-scaling/vertical-redis-sentinel.yaml
redisopsrequest.ops.kubedb.com/rd-ops-vertical created
```

#### Verify Redis resources updated successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
NAME              TYPE              STATUS       AGE
rd-ops-vertical   VerticalScaling   Successful   4m4s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 
Now, we are going to verify from the Pod yaml whether the resources of the database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo rd-sample-0 -o json | jq '.spec.containers[].resources'
{}
{
  "limits": {
    "cpu": "500m",
    "memory": "800Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the redis database.
## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
# Delete Redis and RedisOpsRequest
$ kubectl patch -n demo rd/rd-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-sample patched

$ kubectl delete -n demo redis rd-sample
redis.kubedb.com "rd-sample" deleted

$ kubectl delete -n demo redisopsrequest rd-ops-vertical 
redisopsrequest.ops.kubedb.com "rd-ops-vertical" deleted

# Delete RedisSentinel and RedisSentinelOpsRequest
$ kubectl patch -n demo redissentinel/sen-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-sample patched

$ kubectl delete -n demo redissentinel sen-sample
redissentinel.kubedb.com "sen-sample" deleted

$ kubectl delete -n demo redissentinelopsrequests sen-ops-vertical 
redissentinelopsrequest.ops.kubedb.com "sen-ops-vertical" deleted
```
