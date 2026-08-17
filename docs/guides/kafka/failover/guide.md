---
title: Kafka Failover and DR Scenarios
menu:
  docs_{{ .version }}:
    identifier: kafka-failover
    name: Overview
    parent: guides-kafka-FDR
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Exploring Fault Tolerance in Kafka with KubeDB

## Understanding Failover and Clustering in Kafka on KubeDB

`Failover` refers to the process of automatically re-electing leadership when the node currently
responsible for a piece of work fails. In high-availability database systems, failover ensures that
services remain uninterrupted even when one or more nodes go down. This capability is critical in modern,
cloud-native infrastructure where downtime can lead to major disruptions.

When running `Kafka` on Kubernetes using KubeDB, the operator deploys Kafka in **KRaft mode** (no external
ZooKeeper). Every node plays one or both of these roles:

- **Controller role:**
  A quorum of nodes runs the Raft consensus protocol to maintain cluster metadata — topic configuration,
  partition assignment, and broker membership. Exactly one controller is the **active controller** at any
  time; the rest are voters that replicate the metadata log and are ready to take over.

- **Broker role:**
  Nodes that host topic partitions and serve produce/consume traffic. For each partition, one broker is the
  **partition leader** (handles all reads/writes for that partition), and the rest holding a replica are
  **followers**, kept in the partition's **in-sync replica (ISR)** set as long as they keep up.

KubeDB supports two topologies:

- **Combined cluster:** every pod is simultaneously a broker *and* a controller — one PetSet, `spec.replicas`
  pods. Simpler and lighter, good for dev/test. **This is what this guide deploys and tests.**
- **Dedicated topology cluster:** separate `spec.topology.broker` and `spec.topology.controller` node
  pools, each with their own PetSet — recommended for production, since broker load and controller/metadata
  load no longer compete for the same pod's resources. See the
  [topology cluster guide](/docs/guides/kafka/clustering/topology-cluster/index.md) to deploy this instead.

Because every pod in a combined cluster is both broker and controller, deleting **any single pod** can
simultaneously trigger a partition-leader election (if it held one) and a controller election (if it was
the active controller) — they aren't independent failures the way they are in a dedicated topology cluster.

## Before You Begin

Before proceeding:

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to
  communicate with your cluster. If you do not already have a cluster, you can create one by
  using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- Read the [Kafka topology cluster concept](/docs/guides/kafka/clustering/topology-cluster/index.md) to learn about broker/controller roles.

- To keep things isolated, this tutorial uses a separate namespace called `kf-demo` throughout this tutorial.
  Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns kf-demo
  namespace/kf-demo created
  ```

## Deploy Kafka Combined Cluster

The following is an example `Kafka` object which creates a 3-node combined cluster — every pod acts as both
broker and controller, enough to tolerate a single pod failure while keeping controller quorum.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-multinode
  namespace: kf-demo
spec:
  replicas: 3
  version: 4.2.0
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
  storageType: Durable
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f kafka-multinode.yaml
kafka.kubedb.com/kafka-multinode created
```

Here,

- `spec.replicas` — number of combined broker+controller pods. Kept **odd and ≥ 3** so controller quorum
  survives a single node failure. (A dedicated topology cluster would instead set
  `spec.topology.broker.replicas` and `spec.topology.controller.replicas` independently.)

KubeDB operator watches for `Kafka` objects using the Kubernetes API. Since no `spec.topology` is set, the
operator creates a single PetSet where every pod runs both roles.

Watch the bootstrap progress:

```bash
$ kubectl get kf -n kf-demo -w
NAME              VERSION   STATUS   AGE
kafka-multinode   4.2.0     Ready    18m
```

```bash
$ kubectl get all -n kf-demo -l 'app.kubernetes.io/instance=kafka-multinode'
NAME                    READY   STATUS    RESTARTS   AGE
pod/kafka-multinode-0   1/1     Running   0          16m
pod/kafka-multinode-1   1/1     Running   0          15m
pod/kafka-multinode-2   1/1     Running   0          15m

NAME                           TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-multinode-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   16m

NAME                                                 TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-multinode   kubedb.com/kafka   4.2.0     16m
```

Since this is a combined cluster, there's only one pod set (`kafka-multinode-0/1/2`) — no separate
`-broker-`/`-controller-` suffixes. Each pod's ordinal is used as both its broker ID and its controller
voter ID.

## Verify Controller Quorum and Partition Leadership

Exec into any pod and describe the KRaft metadata quorum:

```bash
$ kubectl exec -it -n kf-demo kafka-multinode-0 -- bash
Defaulted container "kafka" out of: kafka, kafka-init (init)
kafka@kafka-multinode-0:~$ kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
ClusterId:              11f1-88a2-9ae67b83fafw
LeaderId:               0
LeaderEpoch:            1
HighWatermark:          2168
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   0
CurrentVoters:          [{"id": 0, "endpoints": ["CONTROLLER://kafka-multinode-0.kafka-multinode-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1, "endpoints": ["CONTROLLER://kafka-multinode-1.kafka-multinode-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 2, "endpoints": ["CONTROLLER://kafka-multinode-2.kafka-multinode-pods.kf-demo.svc.cluster.local:9093"]}]
CurrentObservers:       []
```

`LeaderId: 0` means `kafka-multinode-0` is currently the active controller among the 3 voters.

Now create a topic with a replication factor of 3, so every pod holds a copy of the partition:

```bash
kafka@kafka-multinode-0:~$ kafka-topics.sh --command-config config/clientauth.properties --create --topic sample --partitions 1 --replication-factor 3 --bootstrap-server localhost:9092
Created topic sample.

kafka@kafka-multinode-0:~$ kafka-topics.sh --command-config config/clientauth.properties --describe --topic sample --bootstrap-server localhost:9092
Topic: sample	TopicId: clHkbrXASlSJd5FgWZu5kA	PartitionCount: 1	ReplicationFactor: 3	Configs: min.insync.replicas=1,segment.bytes=1073741824,min.compaction.lag.ms=60000
	Topic: sample	Partition: 0	Leader: 1	Replicas: 1,2,0	Isr: 1,2,0	Elr: 	LastKnownElr:
```

Broker `1` (`kafka-multinode-1`) is the partition leader for `sample-0`, and all three pods are in the ISR
(fully caught up). Note `min.insync.replicas=1` on this topic — a produce with `acks=all` only needs 1
in-sync replica to acknowledge, so writes can keep going even with just one pod up (more on this in Case 2).

## Produce and Consume Across the Cluster

```bash
kafka@kafka-multinode-1:~$ kafka-console-producer.sh --producer.config config/clientauth.properties --topic sample --request-required-acks all --bootstrap-server localhost:9092
>hello
>hi
```

```bash
kafka@kafka-multinode-0:~$ kafka-console-consumer.sh --consumer.config config/clientauth.properties --topic sample --from-beginning --bootstrap-server localhost:9092
hello
hi
```

## How Failover Works

Kafka in KRaft mode runs two independent failover flows — but on a **combined** cluster, both can be
triggered by deleting the very same pod, since every pod plays both roles:

**Controller quorum failover.** Controllers replicate a metadata log using Raft. If the active controller
stops responding, the remaining voters detect the missing heartbeats and hold an election; a candidate must
get votes from a **majority** of the controller voters to become the new active controller. With `3`
combined pods, any `2` form a majority — the quorum survives losing `1`. If a majority is unreachable, no
new active controller can be elected, and metadata operations (creating topics, altering partition
assignments, broker registration) stall until quorum returns.

**Partition leader failover.** For each partition, if the pod currently acting as leader goes down, the
active controller picks a replacement leader from the partition's in-sync replica (ISR) set — followers
that were fully caught up with the old leader — and updates the partition metadata. Because the new leader
was already in-sync, no data already acknowledged (`acks=all`, bounded by `min.insync.replicas`) is lost.

### Hands-on Failover Testing

>note: The Deploy/Verify/Produce steps above were run against a live cluster. The cases below describe the
> expected KRaft/ISR behavior for this topology and are a good next step to verify live yourself — treat
> the exact IDs/output as illustrative.

#### Case 1: Delete a single pod (partition leader and/or active controller)

`kafka-multinode-1` is currently both the partition leader for `sample-0` and (per Case 2 below) potentially
the active controller too — on a combined cluster you don't get to isolate the two roles.

```bash
$ kubectl delete pod -n kf-demo kafka-multinode-1
pod "kafka-multinode-1" deleted
```

With `2` of `3` pods still up, controller quorum is intact (still a majority), and a new partition leader is
elected from the remaining ISR members:

```bash
kafka@kafka-multinode-0:~$ kafka-topics.sh --command-config config/clientauth.properties --describe --topic sample --bootstrap-server localhost:9092
Topic: sample	TopicId: clHkbrXASlSJd5FgWZu5kA	PartitionCount: 1	ReplicationFactor: 3	Configs: min.insync.replicas=1,segment.bytes=1073741824,min.compaction.lag.ms=60000
	Topic: sample	Partition: 0	Leader: 2	Replicas: 1,2,0	Isr: 2,0	Elr: 	LastKnownElr:
```

Producing and consuming continue uninterrupted against the new leader (`kafka-multinode-2`):

```bash
kafka@kafka-multinode-0:~$ kafka-console-producer.sh --producer.config config/clientauth.properties --topic sample --request-required-acks all --bootstrap-server localhost:9092
>still working
```

Once `kafka-multinode-1` comes back, it catches up on the partition log and rejoins the ISR as a follower,
and rejoins the controller quorum as a voter:

```bash
Topic: sample	Partition: 0	Leader: 2	Replicas: 1,2,0	Isr: 2,0,1
```

#### Case 2: Delete two pods at once (controller quorum lost)

```bash
$ kubectl delete pod -n kf-demo kafka-multinode-0 kafka-multinode-2
pod "kafka-multinode-0" deleted
pod "kafka-multinode-2" deleted
```

With only `1` of `3` voters left, controller quorum is lost — no new active controller can be elected, and
metadata operations (new topics, partition reassignment, broker (re)registration) stall:

```bash
kubectl exec -it -n kf-demo kafka-multinode-0 -- bash
Defaulted container "kafka" out of: kafka, kafka-init (init)
kafka@kafka-multinode-0:~$ kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
[2026-08-04 05:34:51,762] WARN [AdminClient clientId=adminclient-1] Connection to node -1 (localhost/127.0.0.1:9092) could not be established. Node may not be available. (org.apache.kafka.clients.NetworkClient)
[2026-08-04 05:34:51,858] WARN [AdminClient clientId=adminclient-1] Connection to node -1 (localhost/127.0.0.1:9092) could not be established. Node may not be available. (org.apache.kafka.clients.NetworkClient)
[2026-08-04 05:34:51,964] WARN [AdminClient clientId=adminclient-1] Connection to node -1 (localhost/127.0.0.1:9092) could not be established. Node may not be available. (org.apache.kafka.clients.NetworkClient)
[2026-08-04 05:34:52,261] WARN [AdminClient clientId=adminclient-1] Connection to node -1 (localhost/127.0.0.1:9092) could not be established. Node may not be available. (org.apache.kafka.clients.NetworkClient)
[2026-08-04 05:34:52,672] WARN [AdminClient clientId=adminclient-1] Connection to node -1 (localhost/127.0.0.1:9092) could not be established. Node may not be available. (org.apache.kafka.clients.NetworkClient)
[2026-08-04 05:34:53,594] WARN [AdminClient clientId=adminclient-1] Connection to node -1 (localhost/127.0.0.1:9092) could not be established. Node may not be available. (org.apache.kafka.clients.NetworkClient)
ClusterId:              11f1-88a2-9ae67b83fafw
LeaderId:               0
LeaderEpoch:            12
HighWatermark:          7412
MaxFollowerLag:         7413
MaxFollowerLagTimeMs:   -1
CurrentVoters:          [{"id": 0, "endpoints": ["CONTROLLER://kafka-multinode-0.kafka-multinode-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1, "endpoints": ["CONTROLLER://kafka-multinode-1.kafka-multinode-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 2, "endpoints": ["CONTROLLER://kafka-multinode-2.kafka-multinode-pods.kf-demo.svc.cluster.local:9093"]}]
CurrentObservers:       []
kafka@kafka-multinode-0:~$ 

```

Because `min.insync.replicas=1` on the `sample` topic, the surviving pod (`kafka-multinode-1`, if it's an
ISR member for the partition) can often still accept produce/consume traffic for that specific partition
even without controller quorum — existing data-plane traffic and cluster metadata are separate concerns.
What it can't do is anything that requires the controller: elect a new partition leader if this pod also
fails, create topics, or let the two deleted pods rejoin as recognized voters. Once at least one of
`kafka-multinode-0`/`kafka-multinode-2` returns (restoring `2` of `3`), quorum is restored.

#### Case 3: Delete every pod in the cluster

```bash
$ kubectl delete pod -n kf-demo -l app.kubernetes.io/instance=kafka-multinode
pod "kafka-multinode-0" deleted
pod "kafka-multinode-1" deleted
pod "kafka-multinode-2" deleted
```

As the PetSet brings pods back up one at a time (each reattaching its own PVC), they rejoin the metadata
quorum once a majority is reachable, elect an active controller, and resync partition data from disk.
Produce/consume resumes once a partition leader is elected again.

```bash
$ watch kubectl exec -it -n kf-demo kafka-multinode-0 -- kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
```

## CleanUp

For cleaning up what we created in this tutorial follow the following commands:

```bash
$ kubectl patch -n kf-demo kf kafka-multinode -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kafka.kubedb.com/kafka-multinode patched

$ kubectl delete kf -n kf-demo kafka-multinode
kafka.kubedb.com "kafka-multinode" deleted

$ kubectl delete ns kf-demo
namespace "kf-demo" deleted
```

## Next Steps

- Deploy a [dedicated topology cluster](/docs/guides/kafka/clustering/topology-cluster/index.md) for Apache Kafka, for independently-scaled broker/controller pools.
- Monitor your Kafka cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Learn to use KubeDB managed Kafka objects using [CLIs](/docs/guides/kafka/cli/cli.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
