---
title: Upgrading Redis Standalone
menu:
  docs_{{ .version }}:
    identifier: rd-upgrading-standalone
    name: Standalone
    parent: rd-upgrading
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade version of Redis Standalone

This guide will show you how to use `KubeDB` Enterprise operator to upgrade the version of `Redis` standalone.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/opsrequest.md)
  - [Upgrading Overview](/docs/guides/redis/upgrading/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Redis Standalone Database

Now, we are going to deploy a `Redis` standalone database with version `5.0.3-v1`.

### Deploy Redis standalone :

In this section, we are going to deploy a Redis standalone database. Then, in the next section we will upgrade the version of the database using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/rd-standalone.yaml
redis.kubedb.com/redis-quickstart created
```

Now, wait until `redis-quickstart` created has status `Ready`. i.e,

```bash
$ kubectl get rd -n demo
NAME               VERSION    STATUS   AGE
redis-quickstart   5.0.3-v1   Ready    5m14s
```

We are now ready to apply the `RedisOpsRequest` CR to upgrade this database.

### Upgrade Redis Version

Here, we are going to upgrade `Redis` standalone from `5.0.3-v1` to `6.0.6`.

#### Create RedispsRequest:

In order to upgrade the standalone database, we have to create a `RedisOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: upgrade-standalone
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: redis-quickstart
  upgrade:
    targetVersion: 6.0.6
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `redis-quickstart` Redis database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `6.0.6`.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/upgrading/upgrade-standalone.yaml
redisopsrequest.ops.kubedb.com/upgrade-standalone created
```

#### Verify Redis version upgraded successfully :

If everything goes well, `KubeDB` Enterprise operator will update the image of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                    TYPE      STATUS       AGE
upgrade-standalone      Upgrade   Successful   3m45s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to upgrade the database.

```bash
$ kubectl describe redisopsrequest -n demo upgrade-standalone
Name:         upgrade-standalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2020-11-26T05:22:35Z
  Generation:          1
  Resource Version:    15075
  Self Link:           /apis/ops.kubedb.com/v1alpha1/namespaces/demo/redisopsrequests/upgrade-standalone
  UID:                 cfcefd4b-4cf8-49af-8322-121c7f666982
Spec:
  Database Ref:
    Name:  redis-quickstart
  Type:    Upgrade
  Upgrade:
    Target Version:  6.0.6
Status:
  Conditions:
    Last Transition Time:  2020-11-26T05:22:35Z
    Message:               RedisOpsRequest: demo/upgrade-standalone is upgrading database
    Observed Generation:   1
    Reason:                Upgrade
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-11-26T05:22:35Z
    Message:               Successfully paused Redis: redis-quickstart
    Observed Generation:   1
    Reason:                PauseDatabase
    Status:                True
    Type:                  PauseDatabase
    Last Transition Time:  2020-11-26T05:22:35Z
    Message:               Successfully Updated StatefulSets Image
    Observed Generation:   1
    Reason:                UpdateStatefulSetImage
    Status:                True
    Type:                  UpdateStatefulSetImage
    Last Transition Time:  2020-11-26T05:22:50Z
    Message:               Successfully Restarted Pods With Updated Version Image
    Observed Generation:   1
    Reason:                RestartedPodsWithImage
    Status:                True
    Type:                  RestartedPodsWithImage
    Last Transition Time:  2020-11-26T05:22:50Z
    Message:               Upgrading have been done successfully
    Observed Generation:   1
    Reason:                upgradingDone
    Status:                True
    Type:                  upgradingDone
    Last Transition Time:  2020-11-26T05:22:50Z
    Message:               Successfully resumed Redis: redis-quickstart
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-11-26T05:22:50Z
    Message:               RedisOpsRequest: demo/upgrade-standalone Successfully Upgraded Database
    Observed Generation:   1
    Reason:                Upgrade
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                  Age    From                        Message
  ----    ------                  ----   ----                        -------
  Normal  PauseDatabase           7m10s  KubeDB Enterprise Operator  Pausing Redis demo/redis-quickstart
  Normal  PauseDatabase           7m10s  KubeDB Enterprise Operator  Successfully paused Redis demo/redis-quickstart
  Normal  Starting                7m10s  KubeDB Enterprise Operator  Updating Image of StatefulSet: redis-quickstart
  Normal  UpdateStatefulSetImage  7m10s  KubeDB Enterprise Operator  Successfully updated StatefulSets Image
  Normal  RestartedPodsWithImage  6m55s  KubeDB Enterprise Operator  Successfully Restarted Pods With Updated Version Image
  Normal  ResumeDatabase          6m55s  KubeDB Enterprise Operator  Pausing Redis demo/redis-quickstart
  Normal  ResumeDatabase          6m55s  KubeDB Enterprise Operator  Successfully resumed Redis demo/redis-quickstart
  Normal  Successful              6m55s  KubeDB Enterprise Operator  Successfully Completed the OpsRequest
```

Now, we are going to verify whether the `Redis` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get redis -n demo redis-quickstart -o=jsonpath='{.spec.version}{"\n"}'
6.0.6

$ kubectl get statefulset -n demo redis-quickstart -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/redis:6.0.6

$ kubectl get pods -n demo redis-quickstart-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/redis:6.0.6
```

You can see from above, our `Redis` standalone database has been updated with the new version. So, the upgrade process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete redis -n demo redis-quickstart
kubectl delete redisopsrequest -n demo upgrade-standalone
```