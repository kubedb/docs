---
title: Restart Redis 
menu:
  docs_{{ .version }}:
    identifier: rd-restart-redis
    name: Restart 
    parent: rd-redis-guides
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports restarting a Redis/Valkey database via a `RedisOpsRequest`. Restarting is useful if some pods are stuck in an unexpected state or are not functioning correctly. This tutorial will guide you through the process of restarting a Redis cluster using KubeDB.

## Before You Begin

- You need a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you don’t have a cluster, you can create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a namespace called `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The YAML files used in this tutorial are stored in the [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Redis Cluster

In this section, we will deploy a Redis cluster using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version:  8.2.2
  mode: Cluster
  cluster:
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

Let’s create the `Redis` custom resource (CR) shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/restart/redis.yaml
redis.kubedb.com/redis-cluster created
```

Once the Redis cluster is created, you can check the pods created:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=redis-cluster -w
NAME                     READY   STATUS    RESTARTS   AGE
redis-cluster-shard0-0   1/1     Running   0          19h
redis-cluster-shard0-1   1/1     Running   0          19h
redis-cluster-shard1-0   1/1     Running   0          19h
redis-cluster-shard1-1   1/1     Running   0          19h
redis-cluster-shard2-0   1/1     Running   0          19h
redis-cluster-shard2-1   1/1     Running   0          19h
```

## Apply Restart OpsRequest

To restart the Redis cluster, we will create a `RedisOpsRequest` to initiate the restart operation.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: redis-cluster
  apply: Always
```

- `spec.type`: Specifies the type of operation, in this case, `Restart`.
- `spec.databaseRef`: References the name of the Redis database (`redis-cluster`) in the same namespace as the `RedisOpsRequest`.
- `spec.apply`: Determines whether the operation should always be applied (`Always`) or only when there are changes (`IfReady`).

Let’s create the `RedisOpsRequest` CR:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/restart/restart.yaml
RedisOpsRequest.ops.kubedb.com/restart created
```

### Restart Process for Redis Cluster

In a Redis cluster, pods are organized into shards, each containing multiple replicas in a master-slave 
setup. The restart process follows these steps:

- Restart replicas first: For each shard, restart all slave pods one by one.
- Restart masters last: After all slaves in all shards are restarted, restart the master pods one by one.

This ensures high availability and minimal disruption during the restart.

You can check the status of the `RedisOpsRequest` to confirm the restart operation:

```bash
$ kubectl get rdops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   6m51s

$ kubectl get rdops -n demo restart -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"RedisOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"redis-cluster"},"type":"Restart"}}
  creationTimestamp: "2025-10-21T08:41:47Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "157658"
  uid: 234fb008-ca91-420d-a2ae-4121c2c34595
spec:
  apply: Always
  databaseRef:
    name: redis-cluster
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2025-10-21T08:41:47Z"
    message: Redis ops request is restarting the database nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2025-10-21T08:41:53Z"
    message: evict pod; ConditionStatus:True; PodName:redis-cluster-shard0-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--redis-cluster-shard0-1
  - lastTransitionTime: "2025-10-21T08:42:28Z"
    message: is pod ready; ConditionStatus:True; PodName:redis-cluster-shard0-1
    observedGeneration: 1
    status: "True"
    type: IsPodReady--redis-cluster-shard0-1
  - lastTransitionTime: "2025-10-21T08:42:28Z"
    message: evict pod; ConditionStatus:True; PodName:redis-cluster-shard1-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--redis-cluster-shard1-1
  - lastTransitionTime: "2025-10-21T08:43:03Z"
    message: is pod ready; ConditionStatus:True; PodName:redis-cluster-shard1-1
    observedGeneration: 1
    status: "True"
    type: IsPodReady--redis-cluster-shard1-1
  - lastTransitionTime: "2025-10-21T08:43:03Z"
    message: evict pod; ConditionStatus:True; PodName:redis-cluster-shard2-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--redis-cluster-shard2-1
  - lastTransitionTime: "2025-10-21T08:43:38Z"
    message: is pod ready; ConditionStatus:True; PodName:redis-cluster-shard2-1
    observedGeneration: 1
    status: "True"
    type: IsPodReady--redis-cluster-shard2-1
  - lastTransitionTime: "2025-10-21T08:43:38Z"
    message: evict pod; ConditionStatus:True; PodName:redis-cluster-shard0-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--redis-cluster-shard0-0
  - lastTransitionTime: "2025-10-21T08:44:13Z"
    message: is pod ready; ConditionStatus:True; PodName:redis-cluster-shard0-0
    observedGeneration: 1
    status: "True"
    type: IsPodReady--redis-cluster-shard0-0
  - lastTransitionTime: "2025-10-21T08:44:13Z"
    message: evict pod; ConditionStatus:True; PodName:redis-cluster-shard1-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--redis-cluster-shard1-0
  - lastTransitionTime: "2025-10-21T08:44:48Z"
    message: is pod ready; ConditionStatus:True; PodName:redis-cluster-shard1-0
    observedGeneration: 1
    status: "True"
    type: IsPodReady--redis-cluster-shard1-0
  - lastTransitionTime: "2025-10-21T08:44:48Z"
    message: evict pod; ConditionStatus:True; PodName:redis-cluster-shard2-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--redis-cluster-shard2-0
  - lastTransitionTime: "2025-10-21T08:45:23Z"
    message: is pod ready; ConditionStatus:True; PodName:redis-cluster-shard2-0
    observedGeneration: 1
    status: "True"
    type: IsPodReady--redis-cluster-shard2-0
  - lastTransitionTime: "2025-10-21T08:45:23Z"
    message: Successfully restarted pods
    observedGeneration: 1
    reason: RestartPods
    status: "True"
    type: RestartPods
  - lastTransitionTime: "2025-10-21T08:45:23Z"
    message: Successfully Restarted Database
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

```

## Cleaning Up

To clean up the Kubernetes resources created in this tutorial, run:

```bash
$ kubectl delete rdops -n demo restart
$ kubectl delete redis -n demo redis-cluster
$ kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/redis/backup/kubestash/overview/index.md) Redis databases using KubeStash.
- Explore initializing a [Redis database with scripts](/docs/guides/redis/initialization/using-script.md).
- Understand the detailed concepts of the [Redis object](/docs/guides/redis/concepts/redis.md).
- Want to contribute to KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).