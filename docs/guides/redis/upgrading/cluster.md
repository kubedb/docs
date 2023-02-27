---
title: Updating Redis Cluster
menu:
  docs_{{ .version }}:
    identifier: rd-upgrading-cluster
    name: Cluster
    parent: rd-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Update version of Redis Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the version of `Redis` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [Redis Clustering](/docs/guides/redis/clustering/redis-cluster.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Upgrading Overview](/docs/guides/redis/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Redis Cluster Database

Now, we are going to deploy a `Redis` cluster database with version `5.0.3-v1`.

### Deploy Redis cluster :

In this section, we are going to deploy a Redis cluster database. Then, in the next section we will update the version of the database using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: 6.0.6
  mode: Cluster
  cluster:
    master: 3
    replicas: 1
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "100Mi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  terminationPolicy: Halt
```

Let's create the `Redis` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/rd-cluster.yaml
redis.kubedb.com/redis-cluster created
```

Now, wait until `redis-cluster` created has status `Ready`. i.e,

```bash
$ kubectl get rd -n demo
NAME            VERSION   STATUS   AGE
redis-cluster   6.0.6     Ready    88s
```

We are now ready to apply the `RedisOpsRequest` CR to upgrade this database.

### Update Redis Version

Here, we are going to update `Redis` cluster from `6.0.6` to `7.0.5`.

#### Create RedisOpsRequest:

In order to update the cluster database, we have to create a `RedisOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: redis-cluster
  upgrade:
    targetVersion: 7.0.5
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `redis-cluster` Redis database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `7.0.5`.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/update-version.yaml
redisopsrequest.ops.kubedb.com/update-version created
```

#### Verify Redis version upgraded successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME              TYPE            STATUS       AGE
update-version    UpdateVersion   Successful   4m6s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. 

Now, we are going to verify whether the `Redis` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get redis -n demo redis-cluster -o=jsonpath='{.spec.version}{"\n"}'
7.0.5

$ kubectl get statefulset -n demo redis-cluster-shard0 -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
redis:7.0.5@sha256:dfeb5451fce377ab47c5bb6b6826592eea534279354bbfc3890c0b5e9b57c763

$ kubectl get pods -n demo redis-cluster-shard1-1 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
redis:7.0.5@sha256:dfeb5451fce377ab47c5bb6b6826592eea534279354bbfc3890c0b5e9b57c763
```

You can see from above, our `Redis` cluster database has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo rd/redis-cluster -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/redis-quickstart patched

$ kubectl delete -n demo redis redis-cluster
redis.kubedb.com "redis-cluster" deleted

$ kubectl delete -n demo redisopsrequest update-version
redisopsrequest.ops.kubedb.com "update-version" deleted
```
