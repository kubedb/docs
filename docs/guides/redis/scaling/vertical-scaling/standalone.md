---
title: Vertical Scaling Standalone Redis
menu:
  docs_{{ .version }}:
    identifier: rd-vertical-scaling-standalone
    name: Standalone
    parent: rd-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Standalone Redis

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a standalone Redis database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/redis/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Standalone

Here, we are going to deploy a  `Redis` standalone using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Redis Standalone Database

Now, we are going to deploy a `Redis` standalone database with version `6.2.14`.

### Deploy Redis standalone 

In this section, we are going to deploy a Redis standalone database. Then, in the next section we will update the resources of the database using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-quickstart
  namespace: demo
spec:
  version: 6.2.14
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "100m"
          memory: "100Mi"
```

Let's create the `Redis` CR we have shown above, 

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-scaling/rd-standalone.yaml
redis.kubedb.com/redis-quickstart created
```

Now, wait until `rd-quickstart` has status `Ready`. i.e. ,

```bash
$ kubectl get redis -n demo
NAME               VERSION    STATUS   AGE
redis-quickstart   6.2.14   Ready    2m30s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo redis-quickstart-0 -o json | jq '.spec.containers[].resources'
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

We can see from the above output that there are some default resources set by the operator. And the scheduler will choose the best suitable node to place the container of the Pod.

We are now ready to apply the `RedisOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the standalone database to meet the desired resources after scaling.

#### Create RedisOpsRequest

In order to update the resources of the database, we have to create a `RedisOpsRequest` CR with our desired resources. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: redisopsstandalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: redis-quickstart
  verticalScaling:
    redis:
      resources:
        requests:
          memory: "300Mi"
          cpu: "200m"
        limits:
          memory: "800Mi"
          cpu: "500m"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `redis-quickstart` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.redis` specifies the desired resources after scaling.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-scaling/vertical-standalone.yaml
redisopsrequest.ops.kubedb.com/redisopsstandalone created
```

#### Verify Redis Standalone resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `Redis` object and related `PetSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo redisopsstandalone
NAME                 TYPE              STATUS       AGE
redisopsstandalone   VerticalScaling   Successful   26s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 
Now, we are going to verify from the Pod yaml whether the resources of the standalone database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo redis-quickstart-0 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the Redis standalone database.

## Cleaning up

To clean up the Kubernetes resources created by this turorial, run:

```bash

$ kubectl patch -n demo rd/redis-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/redis-quickstart patched

$ kubectl delete -n demo redis redis-quickstart
redis.kubedb.com "redis-quickstart" deleted

$ kubectl delete redisopsrequest -n demo redisopsstandalone
redisopsrequest.ops.kubedb.com "redisopsstandalone" deleted
```