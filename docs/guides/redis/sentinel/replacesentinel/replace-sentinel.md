---
title: Replace Sentinel
menu:
  docs_{{ .version }}:
    identifier: rd-replacing-sentinel
    name: Replace Sentinel
    parent: rd-replace-sentinel
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Replace Sentinel

This guide will show you how to use `KubeDB` Enterprise operator to replace Sentinel instance of Redis Database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisSentinel](/docs/guides/redis/concepts/redissentinel.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply ReplaceSentinel

Here, we are going to deploy a  `Redis` and `RedisSentinel` instance using a supported version by `KubeDB` operator. Then we are going to create another `RedisSentinel`, and it will replace the old sentinel.

### Prepare RedisSentinel

Now, we are going to deploy a `RedisSentinel` version `6.2.14`.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
metadata:
  name: sen-demo
  namespace: demo
spec:
  version: 6.2.14
  replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  terminationPolicy: WipeOut
```
Let's create the `RedisSentinel` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/sentinel/sentinel.yaml
redissentinel.kubedb.com/sen-demo created
```

Now, wait until `sen-dmo` has status `Ready`. i.e. ,

```bash
$ kubectl get redissentinel -n demo
NAME       VERSION   STATUS   AGE
sen-demo   6.2.14     Ready    96s
```
### Deploy Redis in Sentinel Mode

In this section, we are going to deploy a Redis database in Sentinel Mode. 
```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-demo
  namespace: demo
spec:
  version: 6.2.14
  replicas: 3
  sentinelRef:
    name: sen-demo
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
  terminationPolicy: WipeOut
```

Let's create the `Redis` CR we have shown above, 

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/sentinel/redis.yaml
redis.kubedb.com/rd-demo created
```

Now, wait until `rd-demo` has status `Ready`. i.e. ,

```bash
NAME      VERSION   STATUS   AGE
rd-demo   6.2.14     Ready    67s
```

Lets exec into a sentinel pod, and make sure sentinel monitors redis master
```bash
$ kubectl exec -it -n demo sen-demo-0 -c redissentinel -- bash
root@sen-demo-0:/data# redis-cli -p 26379 sentinel masters
1)  1) "name"
    2) "demo/rd-demo"
    3) "ip"
    4) "rd-demo-0.rd-demo-pods.demo.svc"
    5) "port"
    6) "6379"
    7) "runid"
    8) "ae368ff430018c9ef2e4c418aa1d5af1869e01a6"
    9) "flags"
   10) "master"
   11) "link-pending-commands"
   12) "0"
   13) "link-refcount"
   14) "1"
   15) "last-ping-sent"
   16) "0"
   17) "last-ok-ping-reply"
   18) "144"
   19) "last-ping-reply"
   20) "145"
   21) "down-after-milliseconds"
   22) "5000"
   23) "info-refresh"
   24) "755"
   25) "role-reported"
   26) "master"
   27) "role-reported-time"
   28) "103241"
   29) "config-epoch"
   30) "0"
   31) "num-slaves"
   32) "2"
   33) "num-other-sentinels"
   34) "2"
   35) "quorum"
   36) "2"
   37) "failover-timeout"
   38) "5000"
   39) "parallel-syncs"
   40) "1"
root@sen-demo-0:/data# exit
exit
```

### Replace Sentinel

We are going to create a new `RedisSentinel` object for replacing.
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
metadata:
  name: new-sentinel
  namespace: demo
spec:
  version: 6.2.14
  replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  terminationPolicy: WipeOut
```
Let's create the `RedisSentinel` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/sentinel/new-sentinel.yaml
redissentinel.kubedb.com/new-sentinel created
```

Now, wait until `new-sentinel` has status `Ready`. i.e. ,

```bash
$ kubectl get redissentinel -n demo
NAME           VERSION   STATUS   AGE
new-sentinel   6.2.14     Ready    60s
sen-demo       6.2.14     Ready    11m
```

Here, we are going to replace `sen-demo` with `new-sentinel`

#### Create RedisOpsRequest

In order to replace sentinel, we have to create a `RedisOpsRequest` CR with our desired resources. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: replace-sentinel
  namespace: demo
spec:
  type: ReplaceSentinel
  databaseRef:
    name: rd-demo
  sentinel:
    ref:
      name: new-sentinel
      namespace: demo
    removeUnusedSentinel: true
```


Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `redis` database.
- `spec.type` specifies that we are performing `ReplaceSentinel` on our database.
- `spec.sentinel.ref` specifies reference of new sentinel.
- `spec.sentienl.removeUnusedSentinel` specifies whether KubeDB operator should remove orphan Sentinel instance after replacing

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/sentinel/replace-sentinel.yaml
redisopsrequest.ops.kubedb.com/replace-sentinel created
```

#### Verify Replacement

If everything goes well, `KubeDB` Enterprise operator will update the sentinel of `Redis` object.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ kubectl get redisopsrequest -n demo 
NAME               TYPE              STATUS       AGE
replace-sentinel   ReplaceSentinel   Successful   2m34s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 

Now, we are going to verify from the Pod whether the sentinel of the database has updated, Let's check.
Lets exec into one of the new-sentinel pod and verify if it is following the master. And we can additionally check if old sentinel still following the 
database if it exists.

```bash
$ kubectl exec -it -n demo new-sentinel-0 -c redissentinel -- bash
root@new-sentinel-0:/data# redis-cli -p 26379 sentinel masters
1)  1) "name"
    2) "demo/rd-demo"
    3) "ip"
    4) "rd-demo-0.rd-demo-pods.demo.svc"
    5) "port"
    6) "6379"
    7) "runid"
    8) "8af7fc2d42da77f92745b30c9e6bf7d2c21e3d33"
    9) "flags"
   10) "master"
   11) "link-pending-commands"
   12) "0"
   13) "link-refcount"
   14) "1"
   15) "last-ping-sent"
   16) "0"
   17) "last-ok-ping-reply"
   18) "798"
   19) "last-ping-reply"
   20) "798"
   21) "down-after-milliseconds"
   22) "5000"
   23) "info-refresh"
   24) "1350"
   25) "role-reported"
   26) "master"
   27) "role-reported-time"
   28) "240103"
   29) "config-epoch"
   30) "0"
   31) "num-slaves"
   32) "2"
   33) "num-other-sentinels"
   34) "2"
   35) "quorum"
   36) "2"
   37) "failover-timeout"
   38) "5000"
   39) "parallel-syncs"
   40) "1"
root@new-sentinel-0:/data# exit
exit
```

The above output verifies that we have successfully replaced sentinel of Redis database.

## Cleaning up

First set termination policy to `WipeOut` all the things created by KubeDB operator for this Redis instance is deleted. Then delete the redis instance
to clean what you created in this tutorial.

```bash
$ kubectl patch -n demo rd/rd-demo -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-demo patched

$ kubectl delete rd rd-demo -n demo
redis.kubedb.com "rd-demo" deleted

$ kubectl delete -n demo redisopsrequest replace-sentinel
redisopsrequest.ops.kubedb.com "replace-sentinel" deleted
```

Now delete the RedisSentinel instance similarly.
```bash
$ kubectl patch -n demo redissentinel/sen-demo -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-demo patched

$ kubectl delete redissentinel sen-demo -n demo
redis.kubedb.com "sen-demo" deleted

$ kubectl patch -n demo redissentinel/new-sentinel -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/new-sentinel patched

$ kubectl delete redissentinel new-sentinel -n demo
redis.kubedb.com "new-sentinel" deleted
```


## Next Steps

- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [RedisSentinel object](/docs/guides/redis/concepts/redissentinel.md).
- Detail concepts of [RedisVersion object](/docs/guides/redis/concepts/catalog.md).