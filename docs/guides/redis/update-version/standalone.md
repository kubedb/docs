---
title: Updating Redis Standalone
menu:
  docs_{{ .version }}:
    identifier: rd-update-version-standalone
    name: Standalone
    parent: rd-update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Update version of Redis Standalone

This guide will show you how to use `KubeDB` Enterprise operator to update the version of `Redis` standalone.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [updating Overview](/docs/guides/redis/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Redis Standalone Database

Now, we are going to deploy a `Redis` standalone database with version `5.0.3-v1`.

### Deploy Redis standalone :

In this section, we are going to deploy a Redis standalone database. Then, in the next section we will update the version of the database using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-quickstart
  namespace: demo
spec:
  version: 5.0.3-v1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `Redis` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/update-version/rd-standalone.yaml
redis.kubedb.com/redis-quickstart created
```

Now, wait until `redis-quickstart` created has status `Ready`. i.e,

```bash
$ kubectl get rd -n demo
NAME               VERSION    STATUS   AGE
redis-quickstart   5.0.3-v1   Ready    5m14s
```

We are now ready to apply the `RedisOpsRequest` CR to update this database.

### Update Redis Version

Here, we are going to update `Redis` standalone from `5.0.3-v1` to `7.0.5`.

#### Create RedisOpsRequest:

In order to update the standalone database, we have to create a `RedisOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: update-standalone
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: redis-quickstart
  updateVersion:
    targetVersion: 7.0.5
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `redis-quickstart` Redis database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `7.0.5`.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/update-version/update-standalone.yaml
redisopsrequest.ops.kubedb.com/update-standalone created
```

#### Verify Redis version updated successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                    TYPE            STATUS       AGE
update-standalone       UpdateVersion   Successful   3m45s
```

We can see from the above output that the `RedisOpsRequest` has succeeded.

Now, we are going to verify whether the `Redis` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get redis -n demo redis-quickstart -o=jsonpath='{.spec.version}{"\n"}'
7.0.5

$ kubectl get statefulset -n demo redis-quickstart -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
redis:7.0.5@sha256:dfeb5451fce377ab47c5bb6b6826592eea534279354bbfc3890c0b5e9b57c763

$ kubectl get pods -n demo redis-quickstart-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
redis:7.0.5@sha256:dfeb5451fce377ab47c5bb6b6826592eea534279354bbfc3890c0b5e9b57c763
```

You can see from above, our `Redis` standalone database has been updated with the new version. So, the UpdateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo rd/redis-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/redis-quickstart patched

$ kubectl delete -n demo redis redis-quickstart
redis.kubedb.com "redis-quickstart" deleted

$ kubectl delete -n demo redisopsrequest update-standalone
redisopsrequest.ops.kubedb.com "update-standalone" deleted
```
