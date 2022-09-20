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

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale Standalone Redis

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a standalone Redis database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/opsrequest.md)
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

Now, we are going to deploy a `Redis` standalone database with version `5.0.3-v1`.

### Deploy Redis standalone 

In this section, we are going to deploy a Redis standalone database. Then, in the next section we will update the resources of the database using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/rd-standalone.yaml
redis.kubedb.com/redis-quickstart created
```

Now, wait until `rd-quickstart` has status `Ready`. i.e. ,

```bash
$ kubectl get redis -n demo
NAME               VERSION    STATUS   AGE
redis-quickstart   5.0.3-v1   Ready    2m30s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo redis-quickstart-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "250m",
    "memory": "512Mi"
  },
  "requests": {
    "cpu": "100m",
    "memory": "256Mi"
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
- `spec.VerticalScaling.redis` specifies the desired resources after scaling.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-standalone.yaml
redisopsrequest.ops.kubedb.com/redisopsstandalone created
```

#### Verify Redis Standalone resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ kubectl get redisopsrequest -n demo -w
NAME                 TYPE              STATUS       AGE
redisopsstandalone   VerticalScaling   Successful   5m43s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe redisopsrequest -n demo redisopsstandalone
Name:         redisopsstandalone
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2020-11-26T11:14:13Z
  Generation:          1
  Resource Version:    72954
  Self Link:           /apis/ops.kubedb.com/v1alpha1/namespaces/demo/redisopsrequests/redisopsstandalone
  UID:                 92d06d0b-774e-4d57-be62-a4015865034c
Spec:
  Database Ref:
    Name:  redis-quickstart
  Type:    VerticalScaling
  Vertical Scaling:
    Redis:
      Limits:
        Cpu:     500m
        Memory:  800Mi
      Requests:
        Cpu:     200m
        Memory:  300Mi
Status:
  Conditions:
    Last Transition Time:  2020-11-26T11:14:13Z
    Message:               RedisOpsRequest: demo/redisopsstandalone is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-11-26T11:14:13Z
    Message:               Successfully paused Redis: redis-quickstart
    Observed Generation:   1
    Reason:                PauseDatabase
    Status:                True
    Type:                  PauseDatabase
    Last Transition Time:  2020-11-26T11:14:13Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2020-11-26T11:14:33Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartedPodsWithResources
    Status:                True
    Type:                  RestartedPodsWithResources
    Last Transition Time:  2020-11-26T11:14:33Z
    Message:               Vertical scaling have been done successfully
    Observed Generation:   1
    Reason:                ScalingDone
    Status:                True
    Type:                  ScalingDone
    Last Transition Time:  2020-11-26T11:14:33Z
    Message:               Successfully resumed Redis: redis-quickstart
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-11-26T11:14:33Z
    Message:               RedisOpsRequest: demo/redisopsstandalone Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                      Age   From                        Message
  ----    ------                      ----  ----                        -------
  Normal  PauseDatabase               19m   KubeDB Enterprise Operator  Pausing Redis demo/redis-quickstart
  Normal  PauseDatabase               19m   KubeDB Enterprise Operator  Successfully paused Redis demo/redis-quickstart
  Normal  Starting                    19m   KubeDB Enterprise Operator  Updating Resources of StatefulSet: redis-quickstart
  Normal  UpdateStatefulSetResources  19m   KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  RestartedPodsWithResources  18m   KubeDB Enterprise Operator  Successfully Restarted Pods With Resources
  Normal  ResumeDatabase              18m   KubeDB Enterprise Operator  Pausing Redis demo/redis-quickstart
  Normal  ResumeDatabase              18m   KubeDB Enterprise Operator  Successfully resumed Redis demo/redis-quickstart
  Normal  Successful                  18m   KubeDB Enterprise Operator  Successfully Completed the OpsRequest
  Normal  ResumeDatabase              18m   KubeDB Enterprise Operator  Pausing Redis demo/redis-quickstart
  Normal  ResumeDatabase              18m   KubeDB Enterprise Operator  Successfully resumed Redis demo/redis-quickstart
  Normal  Successful                  18m   KubeDB Enterprise Operator  Successfully Completed the OpsRequest
```

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
kubectl delete redis -n demo redis-quickstart
kubectl delete redisopsrequest -n demo redisopsstandalone
```