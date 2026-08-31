---
title: Redis Failover and DR Scenarios
menu:
  docs_{{ .version }}:
    identifier: redis-failover
    name: Overview
    parent: guides-redis-FDR
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Exploring Fault Tolerance in Redis with KubeDB

## Understanding Failover and Clustering in Redis on KubeDB

`Failover` refers to the process of automatically promoting a replacement node when the currently active
node fails. In high-availability database systems, failover ensures that services remain uninterrupted
even when one or more nodes go down. This capability is critical in modern, cloud-native infrastructure
where downtime can lead to major disruptions.

When running `Redis` on Kubernetes using KubeDB, failover becomes more seamless. KubeDB supports two types
of Redis high-availability strategies:

- **Cluster (Sharded, Multi-Primary):**
  Data is sharded across multiple primary (`master`) nodes using hash slots (`0`-`16383`), and each shard
  can have its own replica(s) (`slave`). There's no single primary for the whole dataset — every shard
  independently elects a new primary from among its own replica(s) if its current master fails. Because
  failover happens per-shard, a single node failure only affects the slots owned by that shard, not the
  whole cluster.

- **Sentinel (Primary-Replica):**
  A dedicated `RedisSentinel` instance monitors a single Redis primary and its replicas, and drives failover
  by promoting a replica when the primary goes down. This is the classic Redis high-availability setup and
  is analogous to a master-slave replication topology. See the
  [Redis Sentinel overview](/docs/guides/redis/sentinel/overview.md) if you want to deploy this mode instead.

In the rest of this guide, we'll focus on how failover works in `Cluster` mode, and how KubeDB handles
recovery in the event of a node failure.

>note: A replica (`slave`) never accepts writes — Redis rejects them with a `READONLY` error. Reads are only served directly by a replica once the client has issued `READONLY` on that connection (which `redis-cli -c` does for you); otherwise even reads get redirected.

## Before You Begin

Before proceeding:

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to
  communicate with your cluster. If you do not already have a cluster, you can create one by
  using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- Read [redis clustering concept](/docs/guides/redis/clustering/overview.md) to learn about Redis Cluster.

- To keep things isolated, this tutorial uses a separate namespace called `redis` throughout this tutorial.
  Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns redis
  namespace/redis created
  ```

## Deploy Redis Cluster

The following is an example `Redis` object which creates a Redis Cluster with three shards, each having
one master and one replica (2 pods per shard, 6 pods total).

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis
  namespace: redis
spec:
  version: 8.2.2
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: "standard"
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f redis-cluster.yaml
redis.kubedb.com/redis created
```

Here,

- `spec.mode: Cluster` tells the operator to deploy Redis in Cluster mode instead of standalone or Sentinel mode.
- `spec.cluster.shards` — number of shards. Each shard independently owns a subset of the cluster's hash slots.
- `spec.cluster.replicas` — number of replicas per shard. With `2`, each shard has exactly one master and one replica (no spare to fail over to if you lose both pods of a shard at once — see Case 4 below).
- `spec.storage` specifies the StorageClass of the PVC dynamically allocated for each pod.

KubeDB operator watches for `Redis` objects using the Kubernetes API. When one is created, the operator
creates one PetSet per shard, plus the Services needed for cluster-bus and client traffic.

You can monitor the status until all pods are ready:

```bash
$ watch kubectl get redis,petset,pods -n redis
```

See the setup is ready.

```bash
$ kubectl get redis,petset,pods -n redis
NAME                     VERSION   STATUS   AGE
redis.kubedb.com/redis   8.2.2     Ready    5m25s

NAME                                        AGE
petset.apps.k8s.appscode.com/redis-shard0   5m21s
petset.apps.k8s.appscode.com/redis-shard1   5m19s
petset.apps.k8s.appscode.com/redis-shard2   5m17s

NAME                 READY   STATUS    RESTARTS   AGE
pod/redis-shard0-0   1/1     Running   0          5m20s
pod/redis-shard0-1   1/1     Running   0          4m52s
pod/redis-shard1-0   1/1     Running   0          5m19s
pod/redis-shard1-1   1/1     Running   0          4m52s
pod/redis-shard2-0   1/1     Running   0          5m16s
pod/redis-shard2-1   1/1     Running   0          4m52s
```

```bash
$ kubectl get svc -n redis
NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)              AGE
redis        ClusterIP   10.43.114.61   <none>        6379/TCP             6m25s
redis-pods   ClusterIP   None           <none>        6379/TCP,16379/TCP   6m26s
```

## Verify Pod Reachability and Role

KubeDB labels every Redis pod with its current replication role (`kubedb.com/role=master` or `slave`) and
the shard it belongs to (`redis.kubedb.com/shard`). This is the easiest, and safest, way to check who's
currently the master of each shard — no exec or credentials required.

```bash
$ kubectl get pods -n redis --show-labels | grep role
redis-shard0-0   1/1     Running   0   12m   ...,kubedb.com/role=master,redis.kubedb.com/shard=0,...
redis-shard0-1   1/1     Running   0   12m   ...,kubedb.com/role=slave,redis.kubedb.com/shard=0,...
redis-shard1-0   1/1     Running   0   12m   ...,kubedb.com/role=master,redis.kubedb.com/shard=1,...
redis-shard1-1   1/1     Running   0   12m   ...,kubedb.com/role=slave,redis.kubedb.com/shard=1,...
redis-shard2-0   1/1     Running   0   12m   ...,kubedb.com/role=master,redis.kubedb.com/shard=2,...
redis-shard2-1   1/1     Running   0   12m   ...,kubedb.com/role=slave,redis.kubedb.com/shard=2,...
```

For a more compact view while watching failover happen live, use `jsonpath`:

```bash
$ kubectl get pods -n redis -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.kubedb\.com/role}{"\tshard="}{.metadata.labels.redis\.kubedb\.com/shard}{"\n"}{end}'
redis-shard0-0	master	shard=0
redis-shard0-1	slave	shard=0
redis-shard1-0	master	shard=1
redis-shard1-1	slave	shard=1
redis-shard2-0	master	shard=2
redis-shard2-1	slave	shard=2
```

You can cross-check this against Redis's own view of the cluster. Fetch the auto-generated auth credentials
and exec into any pod to run `cluster nodes`:

```bash
$ export REDIS_PASSWORD=$(kubectl get secrets -n redis redis-auth -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -it -n redis redis-shard0-0 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c cluster nodes
af70e731bc1581bd952e5e124bf8e5e119d89cba 10.42.0.179:6379@16379 master - 0 1785744427172 2 connected 5461-10922
9afbd2dd395f37e334ab5a296f11e8a3982a524a 10.42.0.185:6379@16379 slave af70e731bc1581bd952e5e124bf8e5e119d89cba 0 1785744426569 2 connected
3e08d7245bcc15402b3a78084406c8dca179e170 10.42.0.184:6379@16379 slave 498c469543eaa09f4cc5f1200ac5413216de54b8 0 1785744426167 1 connected
498c469543eaa09f4cc5f1200ac5413216de54b8 10.42.0.178:6379@16379 myself,master - 0 0 1 connected 0-5460
65d34f5894b6b4bfb4c671a5c18b3482e04a576a 10.42.0.180:6379@16379 master - 0 1785744427676 3 connected 10923-16383
11b53b95e464aa866b01198cecb338616ba87e37 10.42.0.186:6379@16379 slave 65d34f5894b6b4bfb4c671a5c18b3482e04a576a 0 1785744426000 3 connected
```

Each master owns a distinct slot range, and its replica follows it — matching the `kubedb.com/role` labels.

## Insert Data and Check Availability

In Cluster mode, writes for a given key must land on the master that owns the hash slot for that key. Just
as importantly, by default a Redis Cluster **replica redirects even read commands** with `MOVED` unless the
client has sent `READONLY` on that connection first. `redis-cli -c` (cluster mode) handles both cases for
you: it follows `MOVED`/`ASK` redirects automatically, and sends `READONLY` whenever it lands on a replica —
so plain `redis-cli` without `-c` will bounce you around with `MOVED`/`ASK` even against a perfectly healthy
node.

```bash
$ kubectl exec -it -n redis redis-shard0-0 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c set hello world
OK

$ kubectl exec -it -n redis redis-shard0-0 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c get hello
"world"
```

Reading with `-c` works from any pod, master or replica, regardless of which shard it belongs to — the
client transparently follows the redirect to whichever node actually owns the slot:

```bash
$ kubectl exec -it -n redis redis-shard0-1 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c get hello
"world"

$ kubectl exec -it -n redis redis-shard2-1 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c get hello
"world"
```

Drop the `-c` and try the same read directly against a replica — it redirects instead of answering, because
`READONLY` was never sent on this connection:

```bash
$ kubectl exec -it -n redis redis-shard0-1 -c redis -- redis-cli -a "$REDIS_PASSWORD" get hello
(error) MOVED 866 10.42.0.178:6379
```

Writes are rejected by a replica outright, with `READONLY` rather than a redirect — a replica never accepts
a write no matter which slot it's for, `-c` or not:

```bash
$ kubectl exec -it -n redis redis-shard0-1 -c redis -- redis-cli -a "$REDIS_PASSWORD" set hello world2
(error) READONLY You can't write against a read only replica.
```

## How Failover Works

Redis Cluster has no external coordinator like Sentinel or MaxScale — every node gossips with every other
node over the cluster bus (port `16379`) and participates in failover decisions for its own shard.

When a shard's master stops responding:

- **Detection**: Other nodes mark it `PFAIL` (possibly failing) once they can't reach it within
  `cluster-node-timeout`. Once enough nodes (a majority of masters in the whole cluster) agree, it's
  upgraded to `FAIL`.
- **Election**: The replica (or replicas) of the failed master start an election, requesting votes from
  every master node in the cluster. Only a replica whose data is reasonably fresh (bounded by
  `cluster-replica-validity-factor`) is eligible to run.
- **Promotion**: The replica that gets votes from a majority of masters wins, increments the shard's
  `configEpoch`, and takes over the failed master's hash slots by issuing a cluster-wide slot claim.
- **Rejoin**: When the old master comes back, it detects it's been superseded (lower epoch) and
  automatically demotes itself to a replica of the new master.

Because voting requires a majority of the cluster's masters, a single shard's failover doesn't depend on the
health of the other shards — but the cluster as a whole needs a majority of master nodes reachable for any
new election to succeed at all.

### Hands-on Failover Testing

#### Case 1: Delete a shard's master pod

Let's delete `redis-shard0-0` (the master of shard 0) and watch the role change.

```bash
$ kubectl delete pod -n redis redis-shard0-0
pod "redis-shard0-0" deleted
```

```bash
$ watch -n 2 "kubectl get pods -n redis -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role} shard={.metadata.labels.redis\\.kubedb\\.com/shard}{\"\\n\"}{end}'"
```

While the pod is down, its old replica takes over shard 0:

```
redis-shard0-1  master  shard=0
redis-shard1-0  master  shard=1
redis-shard1-1  slave   shard=1
redis-shard2-0  master  shard=2
redis-shard2-1  slave   shard=2
```

Once `redis-shard0-0` comes back, it rejoins as the replica of `redis-shard0-1`:

```
redis-shard0-0  slave   shard=0
redis-shard0-1  master  shard=0
redis-shard1-0  master  shard=1
redis-shard1-1  slave   shard=1
redis-shard2-0  master  shard=2
redis-shard2-1  slave   shard=2
```

Confirm the data survived the promotion — both the new master and the recovered old master (now a replica)
serve it correctly with `-c`:

```bash
$ kubectl exec -it -n redis redis-shard0-1 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c get hello
"world"

$ kubectl exec -it -n redis redis-shard0-0 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c get hello
"world"
```

#### Case 2: Delete two different shards' masters at once

```bash
$ kubectl delete pod -n redis redis-shard1-0 redis-shard2-0
pod "redis-shard1-0" deleted
pod "redis-shard2-0" deleted
```

Each affected shard fails over independently, in parallel. While the two deleted pods are still restarting
(their `kubedb.com/role` label is briefly absent):

```
redis-shard0-0  master  shard=0
redis-shard0-1  slave   shard=0
redis-shard1-0
redis-shard1-1  master  shard=1
redis-shard2-0
redis-shard2-1  master  shard=2
```

The deleted pods rejoin as replicas of their shard's new master once they're back up:

```
redis-shard0-0  master  shard=0
redis-shard0-1  slave   shard=0
redis-shard1-0  slave   shard=1
redis-shard1-1  master  shard=1
redis-shard2-0  slave   shard=2
redis-shard2-1  master  shard=2
```

#### Case 3: Delete a shard's replica only

```bash
$ kubectl delete pod -n redis redis-shard0-1
pod "redis-shard0-1" deleted
```

Since the master of shard 0 never went down, there's nothing to fail over — the master keeps serving reads
and writes for its slot range uninterrupted while the replica pod restarts, reattaches its PVC, and
resynchronizes:

```
redis-shard0-0  master  shard=0
redis-shard0-1
redis-shard1-0  master  shard=1
redis-shard1-1  slave   shard=1
redis-shard2-0  master  shard=2
redis-shard2-1  slave   shard=2
```

```
redis-shard0-0  master  shard=0
redis-shard0-1  slave   shard=0
redis-shard1-0  master  shard=1
redis-shard1-1  slave   shard=1
redis-shard2-0  master  shard=2
redis-shard2-1  slave   shard=2
```

#### Case 4: Delete both pods of a shard at once

Since this deployment has only 1 replica per shard, deleting both wipes out the shard entirely — there's no
surviving copy of that shard's data anywhere in the cluster to fail over to.

```bash
$ kubectl delete pod -n redis redis-shard1-0 redis-shard1-1
pod "redis-shard1-0" deleted
pod "redis-shard1-1" deleted
```

Shard 1's hash slots become unowned until at least one of the two pods comes back and recovers its data
from its PVC (check `cluster nodes` to see exactly which range that shard was assigned):

```bash
$ kubectl exec -it -n redis redis-shard0-0 -c redis -- redis-cli -a "$REDIS_PASSWORD" -c cluster info | grep cluster_state
cluster_state:fail
```

Once the PetSet brings a pod back with its PVC reattached, it resumes ownership of its previously assigned
slots and `cluster_state` returns to `ok`. This case is why production clusters should run with
`spec.cluster.replicas` of at least `1` spread across failure domains (e.g. different nodes/zones via
pod anti-affinity), so a single node loss can never take out every copy of a shard.

## CleanUp

For cleaning up what we created in this tutorial follow the following commands:

```bash
$ kubectl patch -n redis rd/redis -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/redis patched

$ kubectl delete rd redis -n redis
redis.kubedb.com "redis" deleted

$ kubectl delete ns redis
namespace "redis" deleted
```

## Next Steps

- Learn about [backup and restore](/docs/guides/redis/backup/stash/overview/index.md) Redis database using Stash.
- Want the classic primary-replica model instead? Deploy [Redis Sentinel](/docs/guides/redis/sentinel/overview.md).
- Detailed concepts of the [Redis Cluster guide](/docs/guides/redis/clustering/redis-cluster.md), including hash slot assignment.
- Monitor your Redis database with KubeDB using [built-in Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
