---
title: Updating Redis Sentinel
menu:
  docs_{{ .version }}:
    identifier: rd-update-version-sentinel
    name: Sentinel
    parent: rd-update-version
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Update version of Redis Sentinel

This guide will show you how to use `KubeDB` Enterprise operator to update the version of `Redis` in Sentinel mode and `RedisSentinel`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisSentinel](/docs/guides/redis/concepts/redissentinel.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Upgrading Overview](/docs/guides/redis/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Redis Sentinel Database

Now, we are going to deploy a `RedisSentinel` instance with version `6.2.7` and a `Redis` database with version `6.2.5`. Then, in the next section we will update the version of the sentinel and the database using `RedisOpsRequest` CRD

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
  terminationPolicy: DoNotTerminate
```

Let's create the `RedisSentinel` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/sentinel.yaml
redissentinel.kubedb.com/sen-sample created
```

Now, wait until `sen-sample` created has status `Ready`. i.e,

```bash
$ kubectl get redissentinel -n demo
NAME         VERSION   STATUS   AGE
sen-sample   6.2.7     Ready    5m20s
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/rd-sentinel.yaml
redis.kubedb.com/rd-sample created
```

Now, wait until `rd-sample` created has status `Ready`. i.e,

```bash
$ kubectl get redis -n demo
NAME        VERSION   STATUS   AGE
rd-sample   6.2.5     Ready    2m11s
```

We are now ready to apply the `RedisSentinelOpsRequest` CR to update the sentinel version and `RedisOpsRequest` CR to update the database version.

### Update RedisSentinel Version

Here, we are going to update `RedisSentinel` standalone from `6.2.7` to `7.0.5`.

#### Create RedisSentinelOpsRequest:

In order to update the sentinel, we have to create a `RedisSentinelOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `RedisSentinelOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisSentinelOpsRequest
metadata:
  name: update-sen-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: sen-sample
  upgrade:
    targetVersion: 7.0.5
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `sen-sample` RedisSentinel instance.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `7.0.5`.

Let's create the `RedisSentinelOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/upgrade-sentinel.yaml
redissentinelopsrequest.ops.kubedb.com/update-sen-version created
```

#### Verify RedisSentinel version updated successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `RedisSentinel` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisSentinelOpsRequest` to be `Successful`.  Run the following command to watch `RedisSentinelOpsRequest` CR,

```bash
$ watch kubectl get redissentinelopsrequest -n demo
Every 2.0s: kubectl get redissentinelopsrequest -n demo
NAME                  TYPE            STATUS       AGE
update-sen-version    UpdateVersion   Successful   3m30s
```

We can see from the above output that the `RedisOpsRequest` has succeeded.

Now, we are going to verify whether the `RedisSentinel` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get redissentinel -n demo sen-sample -o=jsonpath='{.spec.version}{"\n"}'
7.0.5

$ kubectl get statefulset -n demo sen-sample -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
redis:7.0.5@sha256:dfeb5451fce377ab47c5bb6b6826592eea534279354bbfc3890c0b5e9b57c763

$ kubectl get pods -n demo sen-sample-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
redis:7.0.5@sha256:dfeb5451fce377ab47c5bb6b6826592eea534279354bbfc3890c0b5e9b57c763
```

You can see from above, our `RedisSentinel` sen-demo has been updated with the new version. So, the UpdateVersion process is successfully completed.
### Update Redis Version

Here, we are going to update `Redis` standalone from `6.2.5` to `7.0.4`.

#### Create RedispsRequest:

In order to update the redis database, we have to create a `RedisOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: update-rd-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: rd-sample
  upgrade:
    targetVersion: 7.0.4
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `rd-sample` Redis database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `7.0.5`.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/upgrade-redis-sentinel.yaml
redisopsrequest.ops.kubedb.com/update-rd-version created
```

#### Verify Redis version updated successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                 TYPE            STATUS       AGE
update-rd-version    UpdateVersion   Successful   5m40s
```

We can see from the above output that the `RedisOpsRequest` has succeeded.

Now, we are going to verify whether the `Redis` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get redis -n demo rd-sample -o=jsonpath='{.spec.version}{"\n"}'
7.0.4

$ kubectl get statefulset -n demo rd-sample -o=jsonpath='{.spec.template.spec.containers[1].image}{"\n"}'
redis:7.0.4@sha256:091a7b5de688f283b30a4942280b64cf822bbdab0abfb2d2ce6db989f2d3c3f4

$ kubectl get pods -n demo rd-sample-0 -o=jsonpath='{.spec.containers[1].image}{"\n"}'
redis:7.0.4@sha256:091a7b5de688f283b30a4942280b64cf822bbdab0abfb2d2ce6db989f2d3c3f4
```

You can see from above, our `Redis` standalone database has been updated with the new version. So, the UpdateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
# Delete Redis and RedisOpsRequest
$ kubectl patch -n demo rd/rd-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-sample patched

$ kubectl delete -n demo redis rd-sample
redis.kubedb.com "rd-sample" deleted

$ kubectl delete -n demo redisopsrequest update-rd-version
redisopsrequest.ops.kubedb.com "update-rd-version" deleted

# Delete RedisSentinel and RedisSentinelOpsRequest
$ kubectl patch -n demo redissentinel/sen-sample -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-sample patched

$ kubectl delete -n demo redissentinel sen-sample
redissentinel.kubedb.com "sen-sample" deleted

$ kubectl delete -n demo redissentinelopsrequests update-sen-version
redissentinelopsrequest.ops.kubedb.com "update-sen-version" deleted
```
