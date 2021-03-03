---
title: Upgrading Redis Cluster
menu:
  docs_{{ .version }}:
    identifier: rd-upgrading-cluster
    name: Cluster
    parent: rd-upgrading
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade version of Redis Cluster

This guide will show you how to use `KubeDB` Enterprise operator to upgrade the version of `Redis` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [Redis Clustering](/docs/guides/redis/clustering/redis-cluster.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/opsrequest.md)
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

In this section, we are going to deploy a Redis cluster database. Then, in the next section we will upgrade the version of the database using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: 5.0.3-v1
  mode: Cluster
  cluster:
    master: 3
    replicas: 1
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/rd-cluster.yaml
redis.kubedb.com/redis-cluster created
```

Now, wait until `redis-cluster` created has status `Ready`. i.e,

```bash
$ kubectl get rd -n demo
NAME              VERSION    STATUS   AGE
redis-cluster     5.0.3-v1   Ready    3m14s
```

We are now ready to apply the `RedisOpsRequest` CR to upgrade this database.

### Upgrade Redis Version

Here, we are going to upgrade `Redis` cluster from `5.0.3-v1` to `6.0.6`.

#### Create RedispsRequest:

In order to upgrade the cluster database, we have to create a `RedisOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: upgrade-cluster
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: redis-cluster
  upgrade:
    targetVersion: 6.0.6
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `redis-cluster` Redis database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `6.0.6`.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/upgrade-cluster.yaml
redisopsrequest.ops.kubedb.com/upgrade-cluster created
```

#### Verify Redis version upgraded successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                    TYPE      STATUS       AGE
upgrade-cluster         Upgrade   Successful   90s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to upgrade the database.

```bash
$ kubectl describe redisopsrequest -n demo upgrade-cluster
Name:         upgrade-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2020-11-26T06:18:15Z
  Generation:          1
  Resource Version:    24726
  Self Link:           /apis/ops.kubedb.com/v1alpha1/namespaces/demo/redisopsrequests/upgrade-cluster
  UID:                 02224da5-4bc9-437b-9bea-34325c867b20
Spec:
  Database Ref:
    Name:  redis-cluster
  Type:    Upgrade
  Upgrade:
    Target Version:  6.0.6
Status:
  Conditions:
    Last Transition Time:  2020-11-26T06:18:15Z
    Message:               RedisOpsRequest: demo/upgrade-cluster is upgrading database
    Observed Generation:   1
    Reason:                Upgrade
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-11-26T06:18:15Z
    Message:               Successfully paused Redis: redis-cluster
    Observed Generation:   1
    Reason:                PauseDatabase
    Status:                True
    Type:                  PauseDatabase
    Last Transition Time:  2020-11-26T06:18:16Z
    Message:               Successfully Updated StatefulSets Image
    Observed Generation:   1
    Reason:                UpdateStatefulSetImage
    Status:                True
    Type:                  UpdateStatefulSetImage
    Last Transition Time:  2020-11-26T06:19:21Z
    Message:               Successfully Restarted Pods With Updated Version Image
    Observed Generation:   1
    Reason:                RestartedPodsWithImage
    Status:                True
    Type:                  RestartedPodsWithImage
    Last Transition Time:  2020-11-26T06:19:21Z
    Message:               Upgrading have been done successfully
    Observed Generation:   1
    Reason:                upgradingDone
    Status:                True
    Type:                  upgradingDone
    Last Transition Time:  2020-11-26T06:19:21Z
    Message:               Successfully resumed Redis: redis-cluster
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-11-26T06:19:21Z
    Message:               RedisOpsRequest: demo/upgrade-cluster Successfully Upgraded Database
    Observed Generation:   1
    Reason:                Upgrade
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                  Age   From                        Message
  ----    ------                  ----  ----                        -------
  Normal  PauseDatabase           2m8s  KubeDB Enterprise Operator  Pausing Redis demo/redis-cluster
  Normal  PauseDatabase           2m8s  KubeDB Enterprise Operator  Successfully paused Redis demo/redis-cluster
  Normal  Starting                2m8s  KubeDB Enterprise Operator  Updating Image of StatefulSet: redis-cluster-shard0
  Normal  Starting                2m8s  KubeDB Enterprise Operator  Updating Image of StatefulSet: redis-cluster-shard1
  Normal  Starting                2m7s  KubeDB Enterprise Operator  Updating Image of StatefulSet: redis-cluster-shard2
  Normal  UpdateStatefulSetImage  2m7s  KubeDB Enterprise Operator  Successfully updated StatefulSets Image
  Normal  RestartedPodsWithImage  62s   KubeDB Enterprise Operator  Successfully Restarted Pods With Updated Version Image
  Normal  ResumeDatabase          62s   KubeDB Enterprise Operator  Pausing Redis demo/redis-cluster
  Normal  ResumeDatabase          62s   KubeDB Enterprise Operator  Successfully resumed Redis demo/redis-cluster
  Normal  Successful              62s   KubeDB Enterprise Operator  Successfully Completed the OpsRequest
```

Now, we are going to verify whether the `Redis` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get redis -n demo redis-cluster -o=jsonpath='{.spec.version}{"\n"}'
6.0.6

$ kubectl get statefulset -n demo redis-cluster-shard0 -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/redis:6.0.6

$ kubectl get pods -n demo redis-cluster-shard1-1 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/redis:6.0.6

```

You can see from above, our `Redis` cluster database has been updated with the new version. So, the upgrade process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete redis -n demo redis-cluster
kubectl delete redisopsrequest -n demo upgrade-cluster
```