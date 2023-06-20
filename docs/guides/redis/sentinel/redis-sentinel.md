---
title: Redis Sentinel Guide
menu:
  docs_{{ .version }}:
    identifier: rd-sentinel
    name: Sentinel Guide
    parent: rd-sentinel-redis
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Redis Sentinel

This tutorial will show you how to use KubeDB to provision a Redis Sentinel .

## Before You Begin

Before proceeding:

- Read [redis sentinel concept](/docs/guides/redis/sentinel/overview.md) to learn about Redis Sentinel.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Redis Sentinel

First RedisSentinel instance needs to be deployed and then a Redis instance in Sentinel mode which will be monitored by the RedisSentinel instance.

The following is an example `RedisSentinel` object which creates a Sentinel with three replicas.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
metadata:
  name: sen-demo
  namespace: demo
spec:
  version: 6.2.8
  replicas: 3
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
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/sentinel/sentinel.yaml
redissentinel.kubedb.com/sen-demo created
```

Here,
- `spec.replicas` denotes the number of replica nodes
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `RedisSentinel` objects using Kubernetes API. When a `RedisSentinel` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching RedisSentinel object name. KubeDB operator will also create a governing service for StatefulSets named `kubedb`, if one is not already present.


Now we will deploy a Redis instance with giving the sentinelRef to the previously created RedisSentinel instance.
To deploy a Redis in Sentinel mode, specify `spec.mode` field in `Redis` CRD.

The following is an example `Redis` object which creates a Redis Sentinel with three replica node, and it is monitored by Sentinel instance `sentinel`

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: rd-demo
  namespace: demo
spec:
  version: 6.2.8
  replicas: 3
  sentinelRef:
    name: sen-demo
    namespace: demo
  mode: Sentinel
  storage:
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  terminationPolicy: Halt

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/sentinel/redis.yaml
redis.kubedb.com/rd-demo created
```

Here,

- `spec.mode` specifies the mode for Redis. Here we have used `Redis` to tell the operator that we want to deploy Redis in sentinel mode.
- `spec.replicas` denotes the number of replica nodes
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `Redis` objects using Kubernetes API. When a `Redis` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching Redis object name. KubeDB operator will also create a governing service for StatefulSets named `kubedb`, if one is not already present.

```bash
$ kubectl get redissentinel -n demo
NAME       VERSION   STATUS   AGE
sen-demo   6.2.8     Ready    2m39

$ kubectl get redis -n demo
NAME      VERSION   STATUS         AGE
rd-demo   6.2.8     Ready   2m41s

$ kubectl get statefulset -n demo
NAME       READY   AGE
rd-demo    3/3     86s
sen-demo   3/3     12m


$ kubectl get pvc -n demo
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-rd-demo-0    Bound    pvc-830fb301-512a-4de9-a110-c0ce032fabca   1Gi        RWO            standard       99s
data-rd-demo-1    Bound    pvc-0bc06618-a7ef-42ef-b2a0-4e5563d68df7   1Gi        RWO            standard       93s
data-rd-demo-2    Bound    pvc-99aebc54-c016-4376-a3a3-25f882ae86e7   1Gi        RWO            standard       87s
data-sen-demo-0   Bound    pvc-c55d804e-67e1-431c-92a6-67bdde14f59c   1Gi        RWO            standard       12m
data-sen-demo-1   Bound    pvc-171e7d75-c423-4c7f-aabd-42ce50cd0ff4   1Gi        RWO            standard       12m
data-sen-demo-2   Bound    pvc-2886e192-845b-4b44-89e0-20c2af64ec47   1Gi        RWO            standard       12m


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
pvc-0bc06618-a7ef-42ef-b2a0-4e5563d68df7   1Gi        RWO            Delete           Bound    demo/data-rd-demo-1    standard                111s
pvc-171e7d75-c423-4c7f-aabd-42ce50cd0ff4   1Gi        RWO            Delete           Bound    demo/data-sen-demo-1   standard                13m
pvc-2886e192-845b-4b44-89e0-20c2af64ec47   1Gi        RWO            Delete           Bound    demo/data-sen-demo-2   standard                12m
pvc-830fb301-512a-4de9-a110-c0ce032fabca   1Gi        RWO            Delete           Bound    demo/data-rd-demo-0    standard                117s
pvc-99aebc54-c016-4376-a3a3-25f882ae86e7   1Gi        RWO            Delete           Bound    demo/data-rd-demo-2    standard                104s
pvc-c55d804e-67e1-431c-92a6-67bdde14f59c   1Gi        RWO            Delete           Bound    demo/data-sen-demo-0   standard                13m


$ kubectl get svc -n demo
NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
rd-demo           ClusterIP   10.96.165.208   <none>        6379/TCP    2m40s
rd-demo-pods      ClusterIP   None            <none>        6379/TCP    2m40s
rd-demo-standby   ClusterIP   10.96.193.56    <none>        6379/TCP    2m40s
sen-demo          ClusterIP   10.96.249.99    <none>        26379/TCP   14m
sen-demo-pods     ClusterIP   None            <none>        26379/TCP   14m
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. `status.phase` section is similar for 
`Redis` object and `RedisSentinel` object. Run the following command to see the modified `RedisSentinel` object:

```bash
$ kubectl get redissentinel -n demo sen-demo -o yaml
```
```yaml
apiVersion: kubedb.com/v1alpha2
kind: RedisSentinel
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"RedisSentinel","metadata":{"annotations":{},"name":"sen-demo","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"Halt","version":"6.2.8"}}
  creationTimestamp: "2023-02-03T06:36:16Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: sen-demo
  namespace: demo
  resourceVersion: "531539"
  uid: 9b3785b5-4dc3-47bc-91e2-ba260dabd17e
spec:
  authSecret:
    name: sen-demo-auth
  autoOps: {}
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
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
                  app.kubernetes.io/instance: sen-demo
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: redissentinels.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: sen-demo
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: redissentinels.kubedb.com
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
      serviceAccountName: sen-demo
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: Halt
  version: 6.2.8
status:
  conditions:
  - lastTransitionTime: "2023-02-03T06:36:16Z"
    message: 'The KubeDB operator has started the provisioning of Redis: demo/sen-demo'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2023-02-03T06:36:52Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2023-02-03T06:37:13Z"
    message: 'The Sentinel: demo/sen-demo is accepting client requests.'
    observedGeneration: 2
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2023-02-03T06:37:13Z"
    message: 'The Sentinel: demo/sen-demo is ready.'
    observedGeneration: 2
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2023-02-03T06:37:20Z"
    message: 'The Redis: demo/sen-demo is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 2
  phase: Ready
```

## Connection Information

### Connect to Redis Database

- Hostname/address: you can use any of these
  - Service: `rd-demo.demo`
  - Pod IP: (`$ kubectl get pod -n demo -l app.kubernetes.io/name=redises.kubedb.com -o yaml | grep podIP`)
- Port: `6379`
- Username: Run following command to get _username_,

  ```bash
  $ kubectl get secrets -n demo rd-demo-auth -o jsonpath='{.data.\username}' | base64 -d
  default
  ```

- Password: Run the following command to get _password_,

  ```bash
  $ kubectl get secrets -n demo rd-demo-auth -o jsonpath='{.data.\password}' | base64 -d
  5VjZ7iYaoo8YRp!p
  ```
Now, you can connect to this redis database using the service using the credentials.
### Connect to Sentinel

- Hostname/address: you can use any of these
  - Service: `sen-demo.demo`
  - Pod IP: (`$ kubectl get pod -n demo -l app.kubernetes.io/name=redissentinels.kubedb.com -o yaml | grep podIP`)
- Port: `26379`
- Username: Run following command to get _username_,

  ```bash
  $ kubectl get secrets -n demo sen-demo-auth -o jsonpath='{.data.\username}' | base64 -d
  root
  ```

- Password: Run the following command to get _password_,

  ```bash
  $ kubectl get secrets -n demo sen-demo-auth -o jsonpath='{.data.\password}' | base64 -d
  Gw_sd;~Vrsj9kJSL
  ```

Now, you can connect to this sentinel using the service using the credentials. 
## Check Replication Scenario

```bash
# first list the redis pods list
$ kubectl get pods --all-namespaces -o jsonpath='{range.items[*]}{.metadata.name} ---------- {.status.podIP}:6379{"\\n"}{end}' | grep rd-demo
rd-demo-0 ---------- 10.244.0.70:6379
rd-demo-1 ---------- 10.244.0.72:6379
rd-demo-2 ---------- 10.244.0.74:6379

# enter into any pod's container named redis
$ kubectl exec -it -n demo rd-demo-0 -c redis -- bash
/data #

# now inside this container, see which role of this pod
/data #  redis-cli info replication
role:master
connected_slaves:2
slave0:ip=rd-demo-1.rd-demo-pods.demo.svc,port=6379,state=online,offset=1258038,lag=1
slave1:ip=rd-demo-2.rd-demo-pods.demo.svc,port=6379,state=online,offset=1258038,lag=0
master_failover_state:no-failover
master_replid:1ce0d55b8d8c1bd2502d4d7e63b1a2f021dbc938
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:1258038
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:1048576
repl_backlog_first_byte_offset
```
So, the node rd-demo-0 is master, and it has two connected slaves. If a replica node is being exec, it will show which master it is connected to.

## Check Sentinel Monitoring 

A sentinel can monitor multiple masters. Sentinel stores information about master and its replicas. Sentinel constantly monitor master and perform failover 
operation when master fail to respond. Sentinel pings master recurrently after a certain period a time.

```bash
$ kubectl get pods --all-namespaces -o jsonpath='{range.items[*]}{.metadata.name} ---------- {.status.podIP}:6379{"\\n"}{end}' | grep sen-demo
sen-demo-0 ---------- 10.244.0.46:6379
sen-demo-1 ---------- 10.244.0.48:6379
sen-demo-2 ---------- 10.244.0.50:6379
# enter into Sentinel pod's container named redissentinel
$ kubectl exec -it -n demo sen-demo-0 -c redissentinel -- bash

# now inside this container, see the masters information which this sentinels monitors
/data #  redis-cli -p 26379 sentinel masters
1)  1) "name"
    2) "demo/rd-demo"
    3) "ip"
    4) "rd-demo-0.rd-demo-pods.demo.svc"
    5) "port"
    6) "6379"
    7) "runid"
    8) "93cc7c6803fb14a87b6cc59d6fffdc96901153e5"
    9) "flags"
   10) "master"
   11) "link-pending-commands"
   12) "0"
   13) "link-refcount"
   14) "1"
   15) "last-ping-sent"
   16) "0"
   17) "last-ok-ping-reply"
   18) "359"
   19) "last-ping-reply"
   20) "359"
   21) "down-after-milliseconds"
   22) "5000"
   23) "info-refresh"
   24) "8563"
   25) "role-reported"
   26) "master"
   27) "role-reported-time"
   28) "4800628"
   29) "config-epoch"
   30) "0"
   31) "num-slaves"
   32) "2"
   33) "num-other-sentinels"
   34) "2"
   35) "quorum"
   36) "2"
   37) "failover-timeout"
   38) "5000"
   39) "parallel-syncs"
   40) "1"
```

It can be seen that the master `rd-demo-0.rd-demo-pods.demo.svc` has two slaves as we deployed Redis with three replicas, and it has two other sentinel instances
monitoring it as we have deployed RedisSentinel instance with three replicas as well.

## Data Availability

Now, you can connect to this database through [redis-cli](https://redis.io/topics/rediscli). In this tutorial, we will insert data, and we will see whether we can get the data from any other node (any master or replica) or not.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash

# connect to any node
$ kubectl exec -it rd-demo-0 -n demo -c redis -- bash
/data #

# now ensure that you are connected to the 1st pod
/data # redis-cli -c -h 10.244.0.140
10.244.0.140:6379>

# set 'world' as value for the key 'hello'
10.244.0.140:6379> set hello world
OK
10.244.0.140:6379> exit

# switch the connection to the replica of the current master and get the data
/data # redis-cli -c -h 10.244.0.72
10.244.0.145:6379> get hello
"world"
# trying to write data in a replica
10.244.0.145:6379> set apps code 
(error) READONLY You can't write against a read only replica.
10.244.0.145:6379> exit
```

## Automatic Failover

To test automatic failover, we will force the master node to sleep for a period. Since the master node (`pod`) becomes unavailable,
sentinel will initiate a failover, so that a replica is promoted to master. When the old master comes back, it will join the cluster
as the new replica of the new master.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# connect to any node and get the master nodes info
$ kubectl exec -it rd-demo-0 -n demo -c redis -- bash

# Check role of the first pod which has IP 10.244.0.70
/data # redis-cli -h 10.244.0.70 info replication | grep role
role:master

# let's sleep the master node 10.244.0.70 with the `DEBUG SLEEP` command
/data # redis-cli -h 10.244.0.70 debug sleep 120
OK

$ kubectl exec -it rd-demo-0 -n demo -c redis -- bash

# Check role of the first pod which has IP 10.244.0.70
/data # redis-cli -h 10.244.0.70 info replication | grep role
role:slave

# Check role of the second pod which has IP 10.244.0.72
/data # redis-cli -h 10.244.0.72 info replication | grep role
role:master

/data # exit
```

Notice that 110.244.0.72 is the new master and 10.244.0.70 has become the replica of  10.244.0.72.

## Cleaning up

First set termination policy to `WipeOut` all the things created by KubeDB operator for this Redis instance is deleted. Then delete the redis instance
to clean what you created in this tutorial.

```bash
$ kubectl patch -n demo rd/rd-demo -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/rd-demo patched

$ kubectl delete rd rd-demo -n demo
redis.kubedb.com "rd-demo" deleted
```

Now delete the RedisSentinel instance similarly.
```bash
$ kubectl patch -n demo redissentinel/sen-demo -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redissentinel.kubedb.com/sen-demo patched

$ kubectl delete redissentinel sen-demo -n demo
redis.kubedb.com "sen-demo" deleted
```

## Next Steps

- Replace Sentinel of your database with a new [Sentinel](/docs/guides/redis/sentinel/replacesentinel/overview.md)
- Monitor your Redis database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [RedisVersion object](/docs/guides/redis/concepts/catalog.md).
