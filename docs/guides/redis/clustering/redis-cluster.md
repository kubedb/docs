---
title: Redis Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: rd-cluster
    name: Clustering Guide
    parent: rd-clustering-redis
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Redis Cluster

This tutorial will show you how to use KubeDB to provision a Redis cluster.

## Before You Begin

Before proceeding:

- Read [redis clustering concept](/docs/guides/redis/clustering/overview.md) to learn about Redis clustering.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Redis Cluster

To deploy a Redis Cluster, specify `spec.mode` and `spec.cluster` fields in `Redis` CRD.

The following is an example `Redis` object which creates a Redis cluster with three master nodes each of which has one replica node.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: 6.2.14
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
  deletionPolicy: Halt
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/clustering/demo-1.yaml
redis.kubedb.com/redis-cluster created
```

Here,

- `spec.mode` specifies the mode for Redis. Here we have used `Cluster` to tell the operator that we want to deploy Redis in cluster mode.
- `spec.cluster` represents the cluster configuration.
  - `master` denotes the number of master nodes.
  - `replicas` denotes the number of replica nodes per master.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `Redis` objects using Kubernetes API. When a `Redis` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching Redis object name. KubeDB operator will also create a governing service for StatefulSets named `kubedb`, if one is not already present.

```bash
$ kubectl get rd -n demo
NAME            VERSION   STATUS   AGE
redis-cluster   6.2.14     Ready    82s


$ kubectl get petset -n demo
NAME                   READY   AGE
redis-cluster-shard0   2/2     92s
redis-cluster-shard1   2/2     88s
redis-cluster-shard2   2/2     84s

$ kubectl get pvc -n demo
NAME                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-redis-cluster-shard0-0   Bound    pvc-4dd44ddd-06d8-4f2d-bb57-4324c3385d06   1Gi        RWO            standard       112s
data-redis-cluster-shard0-1   Bound    pvc-fb431bb5-036d-4bd8-a89d-4b2477136c1c   1Gi        RWO            standard       105s
data-redis-cluster-shard1-0   Bound    pvc-1be09fa7-6c26-4d5c-8aae-c0cc99e41c73   1Gi        RWO            standard       108s
data-redis-cluster-shard1-1   Bound    pvc-3206ff9e-1ca3-4cef-846d-f91f60c5d572   1Gi        RWO            standard       98s
data-redis-cluster-shard2-0   Bound    pvc-40ccbe7c-e414-4e7b-b40b-2816f42efa63   1Gi        RWO            standard       104s
data-redis-cluster-shard2-1   Bound    pvc-be02792b-b033-407b-a376-9b34001c561f   1Gi        RWO            standard       92s


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                              STORAGECLASS   REASON   AGE
pvc-1be09fa7-6c26-4d5c-8aae-c0cc99e41c73   1Gi        RWO            Delete           Bound    demo/data-redis-cluster-shard1-0   standard                2m33s
pvc-3206ff9e-1ca3-4cef-846d-f91f60c5d572   1Gi        RWO            Delete           Bound    demo/data-redis-cluster-shard1-1   standard                2m21s
pvc-40ccbe7c-e414-4e7b-b40b-2816f42efa63   1Gi        RWO            Delete           Bound    demo/data-redis-cluster-shard2-0   standard                2m29s
pvc-4dd44ddd-06d8-4f2d-bb57-4324c3385d06   1Gi        RWO            Delete           Bound    demo/data-redis-cluster-shard0-0   standard                2m39s
pvc-be02792b-b033-407b-a376-9b34001c561f   1Gi        RWO            Delete           Bound    demo/data-redis-cluster-shard2-1   standard                2m17s
pvc-fb431bb5-036d-4bd8-a89d-4b2477136c1c   1Gi        RWO            Delete           Bound    demo/data-redis-cluster-shard0-1   standard                2m30s

$ kubectl get svc -n demo
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
redis-cluster        ClusterIP   10.96.115.92   <none>        6379/TCP   3m4s
redis-cluster-pods   ClusterIP   None           <none>        6379/TCP   3m4s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified `Redis` object:

```bash
$ kubectl get rd -n demo redis-cluster -o yaml
```
``` yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Redis","metadata":{"annotations":{},"name":"redis-cluster","namespace":"demo"},"spec":{"cluster":{"master":3,"replicas":1},"mode":"Cluster","storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"Halt","version":"6.2.14"}}
  creationTimestamp: "2023-02-02T11:16:57Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: redis-cluster
  namespace: demo
  resourceVersion: "493812"
  uid: d3809d4b-b244-40a5-9570-77141cb1864b
spec:
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: redis-cluster-auth
  autoOps: {}
  cluster:
    master: 3
    replicas: 1
  coordinator:
    resources: {}
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  mode: Cluster
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: redis-cluster
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: redises.kubedb.com
                  redis.kubedb.com/shard: ${SHARD_INDEX}
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: redis-cluster
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: redises.kubedb.com
                  redis.kubedb.com/shard: ${SHARD_INDEX}
              namespaces:
              - demo
              topologyKey: failure-domain.beta.kubernetes.io/zone
            weight: 50
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: redis-cluster
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: Halt
  version: 6.2.14
status:
  conditions:
  - lastTransitionTime: "2023-02-02T11:16:57Z"
    message: 'The KubeDB operator has started the provisioning of Redis: demo/redis-cluster'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2023-02-02T11:17:31Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2023-02-02T11:17:44Z"
    message: 'The Redis: demo/redis-cluster is accepting rdClient requests.'
    observedGeneration: 2
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2023-02-02T11:17:54Z"
    message: 'The Redis: demo/redis-cluster is ready.'
    observedGeneration: 2
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2023-02-02T11:18:14Z"
    message: 'The Redis: demo/redis-cluster is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready
```

## Connection Information

- Hostname/address: you can use any of these
  - Service: `redis-cluster.demo`
  - Pod IP: (`$ kubectl get pod -n demo -l app.kubernetes.io/name=redises.kubedb.com -o yaml | grep podIP`)
- Port: `6379`
- Username: Run following command to get _username_,

  ```bash
  $ kubectl get secrets -n demo redis-cluster-auth -o jsonpath='{.data.\username}' | base64 -d
  default
  ```

- Password: Run the following command to get _password_,

  ```bash
  $ kubectl get secrets -n demo redis-cluster-auth -o jsonpath='{.data.\password}' | base64 -d
  AO8iK)s);o5kQVFs
  ```

Now, you can connect to this database using the service using the credentials. 
## Check Cluster Scenario

The operator creates a cluster according to the newly created `Redis` object. This cluster has 3 masters and one replica per master. And every node in the cluster is responsible for a subset of the total **16384** hash slots.

```bash
# first list the redis pods list
$ kubectl get pods --all-namespaces -o jsonpath='{range.items[*]}{.metadata.name} ---------- {.status.podIP}:6379{"\\n"}{end}' | grep redis
redis-cluster-shard0-0 ---------- 10.244.0.140:6379
redis-cluster-shard0-1 ---------- 10.244.0.145:6379
redis-cluster-shard1-0 ---------- 10.244.0.144:6379
redis-cluster-shard1-1 ---------- 10.244.0.149:6379
redis-cluster-shard2-0 ---------- 10.244.0.146:6379
redis-cluster-shard2-1 ---------- 10.244.0.150:637

# enter into any pod's container named redis
$ kubectl exec -it redis-cluster-shard0-0 -n demo -c redis -- bash
/data #

# now inside this container, see which ones are the masters
# which ones are the replicas
/data # redis-cli -c cluster nodes
d3d7d5924fa4aa7347acb2d4c86f7cd5d18a2950 10.244.0.145:6379@16379 slave f9af25d8db7bb742346b0130fb1cc749ffcd4d1e 0 1675337399550 1 connected
3b4048d43fa982dd246703c899602f5c2472a995 10.244.0.149:6379@16379 slave b49398da2eefac62a3b668a60f36bf4ccc3ccf4f 0 1675337400854 2 connected
b49398da2eefac62a3b668a60f36bf4ccc3ccf4f 10.244.0.144:6379@16379 master - 0 1675337400352 2 connected 5461-10922
31d3f90e1bde3835ca7b08ae8b145b230d9b1ba8 10.244.0.146:6379@16379 master - 0 1675337399000 3 connected 10923-16383
6acca34b192445b888649a839bb7537d2cbb1cf4 10.244.0.150:6379@16379 slave 31d3f90e1bde3835ca7b08ae8b145b230d9b1ba8 0 1675337400553 3 connected
f9af25d8db7bb742346b0130fb1cc749ffcd4d1e 10.244.0.140:6379@16379 myself,master - 0 1675337398000 1 connected 0-5460
```
Each master has assigned some slots from slot 0 to slot 16383, and each master has one replica following it. 

## Data Availability

Now, you can connect to this database through [redis-cli](https://redis.io/topics/rediscli). In this tutorial, we will insert data, and we will see whether we can get the data from any other node (any master or replica) or not.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# here the hash slot for key 'hello' is 866 which is in 1st node
# named 'redis-cluster-shard0-0' (0-5460)
$ kubectl exec -it redis-cluster-shard0-0 -n demo -c redis -- redis-cli -c cluster keyslot hello
(integer) 866

# connect to any node
$ kubectl exec -it redis-cluster-shard0-0 -n demo -c redis -- bash
/data #

# now ensure that you are connected to the 1st pod
/data # redis-cli -c -h 10.244.0.140
10.244.0.140:6379>

# set 'world' as value for the key 'hello'
10.244.0.140:6379> set hello world
OK
10.244.0.140:6379> exit

# switch the connection to the replica of the current master and get the data
/data # redis-cli -c -h 10.244.0.145
10.244.0.145:6379> get hello
-> Redirected to slot [866] located at 10.244.0.140:6379
"world"
10.244.0.145:6379> exit

# switch the connection to any other node
# get the data
/data # redis-cli -c -h 10.244.0.146
10.244.0.146:6379> get hello
-> Redirected to slot [866] located at 10.244.0.140:6379
"world"
10.244.0.146:6379> exit
```

## Automatic Failover

To test automatic failover, we will force a master node to sleep for a period. Since the master node (`pod`) becomes unavailable, the rest of the members will elect a replica (one of its replica in case of more than one replica under this master) of this master node as the new master. When the old master comes back, it will join the cluster as the new replica of the new master.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# connect to any node and get the master nodes info
$ kubectl exec -it redis-cluster-shard0-0 -n demo -c redis -- bash
/data # redis-cli -c cluster nodes | grep master
b49398da2eefac62a3b668a60f36bf4ccc3ccf4f 10.244.0.144:6379@16379 master - 0 1675338070000 2 connected 5461-10922
31d3f90e1bde3835ca7b08ae8b145b230d9b1ba8 10.244.0.146:6379@16379 master - 0 1675338070000 3 connected 10923-16383
f9af25d8db7bb742346b0130fb1cc749ffcd4d1e 10.244.0.140:6379@16379 myself,master - 0 1675338070000 1 connected 0-5460

# let's sleep node 10.244.0.144 with the `DEBUG SLEEP` command
/data # redis-cli -h 10.244.0.144 debug sleep 120
OK

# now again connect to a node and get the master nodes info
$ kubectl exec -it redis-cluster-shard0-0 -n demo -c redis -- bash
/data # redis-cli -c cluster nodes | grep master
3b4048d43fa982dd246703c899602f5c2472a995 10.244.0.149:6379@16379 master - 0 1675338334000 4 connected 5461-10922
31d3f90e1bde3835ca7b08ae8b145b230d9b1ba8 10.244.0.146:6379@16379 master - 0 1675338335355 3 connected 10923-16383
f9af25d8db7bb742346b0130fb1cc749ffcd4d1e 10.244.0.140:6379@16379 myself,master - 0 1675338334000 1 connected 0-5460


/data # redis-cli -c cluster nodes
d3d7d5924fa4aa7347acb2d4c86f7cd5d18a2950 10.244.0.145:6379@16379 slave f9af25d8db7bb742346b0130fb1cc749ffcd4d1e 0 1675338355429 1 connected
3b4048d43fa982dd246703c899602f5c2472a995 10.244.0.149:6379@16379 master - 0 1675338355530 4 connected 5461-10922
b49398da2eefac62a3b668a60f36bf4ccc3ccf4f 10.244.0.144:6379@16379 slave 3b4048d43fa982dd246703c899602f5c2472a995 0 1675338353521 4 connected
31d3f90e1bde3835ca7b08ae8b145b230d9b1ba8 10.244.0.146:6379@16379 master - 0 1675338355000 3 connected 10923-16383
6acca34b192445b888649a839bb7537d2cbb1cf4 10.244.0.150:6379@16379 slave 31d3f90e1bde3835ca7b08ae8b145b230d9b1ba8 0 1675338355000 3 connected
f9af25d8db7bb742346b0130fb1cc749ffcd4d1e 10.244.0.140:6379@16379 myself,master - 0 1675338355000 1 connected 0-5460

/data # exit
```

Notice that 110.244.0.149 is the new master and 10.244.0.144 has become the replica of  10.244.0.149.

## Cleaning up

First set termination policy to `WipeOut` all the things created by KubeDB operator for this Redis instance is deleted. Then delete the redis instance
to clean what you created in this tutorial.

```bash
$ kubectl patch -n demo rd/redis-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/redis-cluster patched

$ kubectl delete rd redis-cluster -n demo
redis.kubedb.com "redis-cluster" deleted
```

## Next Steps

- Deploy [Redis Sentinel](/docs/guides/redis/sentinel/overview.md)
- Monitor your Redis database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [RedisVersion object](/docs/guides/redis/concepts/catalog.md).
