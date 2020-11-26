---
title: Vertical Scaling Redis Cluster
menu:
  docs_{{ .version }}:
    identifier: rd-vertical-scaling-cluster
    name: Cluster
    parent: rd-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale Redis Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a Redis cluster database.

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

## Apply Vertical Scaling on Cluster

Here, we are going to deploy a `Redis` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Redis Cluster Database

Now, we are going to deploy a `Redis` cluster database with version `5.0.3-v1`.

### Deploy Redis Cluster 

In this section, we are going to deploy a Redis cluster database. Then, in the next section we will update the resources of the database using `RedisOpsRequest` CRD. Below is the YAML of the `Redis` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/rd-cluster.yaml
redis.kubedb.com/redis-cluster created
```

Now, wait until `rd-cluster` has status `Ready`. i.e. ,

```bash
$ kubectl get redis -n demo
NAME               VERSION    STATUS   AGE
redis-cluster      5.0.3-v1   Ready    2m30s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo redis-cluster-shard0-0 -o json | jq '.spec.containers[].resources'
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
$ kubectl get pod -n demo redis-cluster-shard1-1 -o json | jq '.spec.containers[].resources'
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

We can see from the above output that there are some default resources set by the operator for pods across all shards. And the scheduler will choose the best suitable node to place the container of the Pod.

We are now ready to apply the `RedisOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the cluster database to meet the desired resources after scaling.

#### Create RedisOpsRequest

In order to update the resources of the database, we have to create a `RedisOpsRequest` CR with our desired resources. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: redisops-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: redis-cluster
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

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `redis-cluster` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.redis` specifies the desired resources after scaling.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/vertical-cluster.yaml
redisopsrequest.ops.kubedb.com/redisops-vertical created
```

#### Verify Redis Cluster resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `Redis` object and related `StatefulSets` and `Pods`.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ kubectl get redisopsrequest -n demo redisops-vertical -w
NAME                TYPE              STATUS       AGE
redisops-vertical   VerticalScaling   Successful   2m11s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
kubectl describe redisopsrequest -n demo redisops-vertical
Name:         redisops-vertical
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2020-11-26T12:02:56Z
  Generation:          1
  Resource Version:    81466
  Self Link:           /apis/ops.kubedb.com/v1alpha1/namespaces/demo/redisopsrequests/redisops-vertical
  UID:                 53b77e5d-31f4-4a24-b282-a85e44420518
Spec:
  Database Ref:
    Name:  redis-cluster
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
    Last Transition Time:  2020-11-26T12:02:56Z
    Message:               RedisOpsRequest: demo/redisops-vertical is vertically scaling database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2020-11-26T12:02:56Z
    Message:               Successfully paused Redis: redis-cluster
    Observed Generation:   1
    Reason:                PauseDatabase
    Status:                True
    Type:                  PauseDatabase
    Last Transition Time:  2020-11-26T12:02:56Z
    Message:               Successfully updated StatefulSets Resources
    Observed Generation:   1
    Reason:                UpdateStatefulSetResources
    Status:                True
    Type:                  UpdateStatefulSetResources
    Last Transition Time:  2020-11-26T12:04:01Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartedPodsWithResources
    Status:                True
    Type:                  RestartedPodsWithResources
    Last Transition Time:  2020-11-26T12:04:02Z
    Message:               Vertical scaling have been done successfully
    Observed Generation:   1
    Reason:                ScalingDone
    Status:                True
    Type:                  ScalingDone
    Last Transition Time:  2020-11-26T12:04:02Z
    Message:               Successfully resumed Redis: redis-cluster
    Observed Generation:   1
    Reason:                ResumeDatabase
    Status:                True
    Type:                  ResumeDatabase
    Last Transition Time:  2020-11-26T12:04:02Z
    Message:               RedisOpsRequest: demo/redisops-vertical Successfully Vertically Scaled Database
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                      Age    From                        Message
  ----    ------                      ----   ----                        -------
  Normal  PauseDatabase               3m44s  KubeDB Enterprise Operator  Pausing Redis demo/redis-cluster
  Normal  PauseDatabase               3m44s  KubeDB Enterprise Operator  Successfully paused Redis demo/redis-cluster
  Normal  Starting                    3m44s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: redis-cluster-shard0
  Normal  Starting                    3m44s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: redis-cluster-shard1
  Normal  Starting                    3m44s  KubeDB Enterprise Operator  Updating Resources of StatefulSet: redis-cluster-shard2
  Normal  UpdateStatefulSetResources  3m44s  KubeDB Enterprise Operator  Successfully updated StatefulSets Resources
  Normal  RestartedPodsWithResources  2m38s  KubeDB Enterprise Operator  Successfully Restarted Pods With Resources
  Normal  ResumeDatabase              2m38s  KubeDB Enterprise Operator  Resuming Redis demo/redis-cluster
  Normal  ResumeDatabase              2m38s  KubeDB Enterprise Operator  Successfully resumed Redis demo/redis-cluster
  Normal  Successful                  2m38s  KubeDB Enterprise Operator  Successfully Completed the OpsRequest

```

Now, we are going to verify from the Pod yaml whether the resources of the cluster database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo redis-cluster-shard0-0 -o json | jq '.spec.containers[].resources'
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
$ kubectl get pod -n demo redis-cluster-shard1-1 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the Redis cluster database.

## Cleaning up

To clean up the kubernetes resources created by this turorial, run:

```bash
kubectl delete redis -n demo redis-cluster
kubectl delete redisopsrequest -n demo redisops-vertical
```