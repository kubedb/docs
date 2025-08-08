---
title: Horizontal Scaling Redis Cluster With Horizon DNS
menu:
  docs_{{ .version }}:
    identifier: rd-horizontal-scaling-cluster-horizon
    name: Horizon DNS
    parent: rd-horizontal-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Redis Cluster With Horizon DNS

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Redis` database shards and replicas Redis in Cluster mode which are running with Announces.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Redis](/docs/guides/redis/concepts/redis.md)
    - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
    - [External Client Connection](/docs/guides/redis/external-connections/exposure.md)
    - [Horizontal Scaling Overview](/docs/guides/redis/scaling/horizontal-scaling/overview.md)

## Apply Horizontal Scaling on Cluster

Here, we are going to deploy a `Redis/Valkey` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Redis Cluster Database

Deploy `Redis/Valkey` cluster as shown in [External Connection Exposer](/docs/guides/redis/external-connections/exposure.md).

After it gets `Ready` check the number of shards and replicas this database has from the Redis object

```bash
$ kubectl get redis -n demo redis-announce -o json | jq '.spec.cluster.shards'
3
$ kubectl get redis -n demo redis-announce -o json | jq '.spec.cluster.replicas'
2
```

Now let's connect to redis-cluster using `redis-cli` and verify master and replica count of the cluster
```bash
$ kubectl exec -it -n demo redis-announce-shard0-0 -c redis -- redis-cli -c cluster nodes | grep master
fc7c635c745b8c74c4422300e945eadb4251add6 10.2.0.87:10050@10056,rd0-0.kubedb.appscode myself,master - 0 1754481552000 1 connected 0-5460
e45749edaf324b980bbf5148644d500d6842ff5c 10.2.0.87:10054@10060,rd0-0.kubedb.appscode master - 0 1754481555065 3 connected 10923-16383
673060b3b589f06fe6a12e6f47ea8910042b6be6 10.2.0.87:10052@10058,rd0-0.kubedb.appscode master - 0 1754481555000 2 connected 5461-10922

$ kubectl exec -it -n demo redis-cluster-shard0-0 -c redis -- redis-cli -c cluster nodes | grep slave | wc -l
3
```

We can see from above output that there are 3 masters and each master has 2 replicas. So, total 6 replicas in the cluster. Each master and its two replicas belongs to a shard.

We are now ready to apply the `RedisOpsRequest` CR to update the resources of this database.

### Horizontal Scaling

Here, we are going to scale up the shards and replicas of the redis cluster to meet the desired resources after scaling.

#### Create RedisOpsRequest

In order to scale up the shards and replicas of the redis cluster, we have to create a `RedisOpsRequest` CR with our desired number of shards and replicas and with desired FQDN. Below is the YAML of the `RedisOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: redisops-horizontal-external
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: redis-announce
  horizontalScaling:
    replicas: 3
    shards: 4
    announce:
      shards:
        - endpoints:
            - rd0-0.kubedb.appscode
        - endpoints:
            - rd0-0.kubedb.appscode
        - endpoints:
            - rd0-0.kubedb.appscode
        - endpoints:
            - rd0-0.kubedb.appscode
            - rd0-0.kubedb.appscode
            - rd0-0.kubedb.appscode
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `redis-announce` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.shards` specifies the desired number of shards after scaling.
- `spec.horizontalScaling.replicas` specifies the desired number of replicas after scaling.
- `spec.horizontalScaling.announce.shards` specifies endpoints for newly created replicas and shards. As we have two replicas in the first shard, and we will increase it by one and make it 3, we have added one FQDN for this.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/scaling/horizontal-scaling/horizontal-cluster.yaml
redisopsrequest.ops.kubedb.com/redisops-horizontal-external created
```

#### Verify Redis Cluster resources updated successfully

If everything goes well, `KubeDB` Enterprise operator will update the replicas and shards of `Redis` object and related `PetSets`.

> `Note`: Newly created pod will be started with the Governing Service DNS and will wait for another `announce` OpsRequest

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo redisops-horizontal
NAME                           TYPE                STATUS       AGE
redisops-horizontal-external   HorizontalScaling   Successful   3m8s
```

Now, we are going to verify if the number of shards and replicas the redis cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get redis -n demo redis-anounce -o json | jq '.spec.cluster.shards'
4
$ kubectl get redis -n demo redis-anounce -o json | jq '.spec.cluster.replicas'
3
```

Let's wait for the new `Announce` opsRequest to be created and Successful

```bash
$ watch kubectl get rdops -n demo
NAME                           TYPE                STATUS       AGE
rd-at86h7                      Announce            Successful   2m
redisops-horizontal-external   HorizontalScaling   Successful   4m
```

Now let's connect to redis-anounce using `redis-cli` and verify master and replica count of the cluster
```bash
$ kubectl exec -it -n demo redis-anounce-shard0-0 -c redis -- redis-cli -c cluster nodes | grep master
fc7c635c745b8c74c4422300e945eadb4251add6 10.2.0.87:10050@10056,rd0-0.kubedb.appscode myself,master - 0 1754484135000 1 connected 1365-5460
039d9b38874ee6dca807836646bbdc8b25f544d5 10.2.0.87:10065@10071,rd0-0.kubedb.appscode master - 0 1754484137945 4 connected 0-1364 5461-6826 10923-12287
e45749edaf324b980bbf5148644d500d6842ff5c 10.2.0.87:10054@10060,rd0-0.kubedb.appscode master - 0 1754484137000 3 connected 12288-16383
673060b3b589f06fe6a12e6f47ea8910042b6be6 10.2.0.87:10052@10058,rd0-0.kubedb.appscode master - 0 1754484136539 2 connected 6827-10922

$ kubectl exec -it -n demo redis-cluster-shard0-0 -c redis -- redis-cli -c cluster nodes | grep slave | wc -l
8
```

The above output verifies that we have successfully scaled up the shards and scaled down the replicas of the Redis cluster database. The slots in redis shard
is also distributed among 4 master.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash

$ kubectl patch -n demo rd/redis-announce -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/redis-announce patched

$ kubectl delete -n demo redis redis-announce
redis.kubedb.com "redis-announce" deleted

$ kubectl delete -n demo redisopsrequest redisops-horizontal-external rd-at86h7
redisopsrequest.ops.kubedb.com "redisops-horizontal-external" deleted
redisopsrequest.ops.kubedb.com "rd-at86h7" deleted
```