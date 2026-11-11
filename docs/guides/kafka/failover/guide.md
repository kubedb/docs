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
  pods. Simpler and lighter, good for dev/test. See the
  [combined cluster guide](/docs/guides/kafka/clustering/combined-cluster/index.md) to deploy this instead.
- **Dedicated topology cluster:** separate `spec.topology.broker` and `spec.topology.controller` node
  pools, each with their own PetSet — recommended for production, since broker load and controller/metadata
  load no longer compete for the same pod's resources. **This is what this guide deploys and tests.**

Because brokers and controllers are separate pods in separate PetSets on a topology cluster, their failures
are genuinely independent — deleting a broker pod never touches controller quorum, and deleting a controller
pod never touches partition leadership. That's the opposite of a combined cluster, where every pod plays
both roles and a single deletion can trigger both elections at once.

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

## Create Topology Kafka Cluster

Here, we are going to create a TLS secured Kafka topology cluster in Kraft mode.

### Create Issuer/ ClusterIssuer

At first, make sure you have cert-manager installed on your k8s for enabling TLS. KubeDB operator uses cert manager to inject certificates into kubernetes secret & uses them for secure `SASL` encrypted communication among kafka brokers and controllers. We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in Kafka. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you CA certificates using openssl.

```bash
openssl req -newkey rsa:2048 -keyout ca.key -nodes -x509 -days 3650 -out ca.crt
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls kafka-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=kf-demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: kafka-ca-issuer
  namespace: kf-demo
spec:
  ca:
    secretName: kafka-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tls/kf-issuer.yaml
issuer.cert-manager.io/kafka-ca-issuer created
```

## Deploy Kafka Topology Cluster

For this demo, we are going to provision Kafka version `4.2.0` with 3 controllers and 3 brokers — enough for
each pool to independently tolerate a single pod failure. To learn more about the Kafka CR, visit
[here](/docs/guides/kafka/concepts/kafka.md). Visit [here](/docs/guides/kafka/concepts/kafkaversion.md) to
learn more about the KafkaVersion CR.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Kafka
metadata:
  name: kafka-topology
  namespace: kf-demo
spec:
  version: 4.2.0
  enableSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: kafka-ca-issuer
      kind: Issuer
  topology:
    broker:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: local-path
    controller:
      replicas: 3
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

Here,

- `spec.topology.broker.replicas` — number of dedicated broker pods, each hosting topic partitions.
- `spec.topology.controller.replicas` — number of dedicated controller pods, forming the Raft metadata
  quorum. Kept **odd and ≥ 3** so quorum survives a single node failure.

Let's deploy the above example by the following command:

```bash
$ kubectl create -f kafka-topology.yaml
kafka.kubedb.com/kafka-topology created
```

KubeDB operator watches for `Kafka` objects using the Kubernetes API. Since `spec.topology` is set, the
operator creates **two** PetSets — one for the broker pool, one for the controller pool.

Watch the bootstrap progress:

```bash
$ kubectl get kf -n kf-demo -w
NAME             TYPE                  VERSION   STATUS         AGE
kafka-topology   kubedb.com/v1alpha2   4.2.0     Provisioning   6s
kafka-topology   kubedb.com/v1alpha2   4.2.0     Provisioning   14s
kafka-topology   kubedb.com/v1alpha2   4.2.0     Provisioning   50s
kafka-topology   kubedb.com/v1alpha2   4.2.0     Ready          68s
```

Hence, the cluster is ready to use. Let's check the Kubernetes resources created by the operator on the
deployment of the Kafka CRO:

```bash
$ kubectl get all,petset,secret,pvc -n kf-demo -l 'app.kubernetes.io/instance=kafka-topology'
NAME                              READY   STATUS    RESTARTS   AGE
pod/kafka-topology-broker-0       1/1     Running   0          37m
pod/kafka-topology-broker-1       1/1     Running   0          36m
pod/kafka-topology-broker-2       1/1     Running   0          36m
pod/kafka-topology-controller-0   1/1     Running   0          37m
pod/kafka-topology-controller-1   1/1     Running   0          36m
pod/kafka-topology-controller-2   1/1     Running   0          36m

NAME                          TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                       AGE
service/kafka-topology-pods   ClusterIP   None         <none>        9092/TCP,9093/TCP,29092/TCP   37m

NAME                                                TYPE               VERSION   AGE
appbinding.appcatalog.appscode.com/kafka-topology   kubedb.com/kafka   4.2.0     37m

NAME                                                     AGE
petset.apps.k8s.appscode.com/kafka-topology-broker       37m
petset.apps.k8s.appscode.com/kafka-topology-controller   37m

NAME                                  TYPE                       DATA   AGE
secret/kafka-topology-auth            kubernetes.io/basic-auth   2      37m
secret/kafka-topology-client-cert     kubernetes.io/tls          6      37m
secret/kafka-topology-keystore-cred   Opaque                     3      37m
secret/kafka-topology-server-cert     kubernetes.io/tls          5      37m

NAME                                                                    STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/kafka-topology-data-kafka-topology-broker-0       Bound    pvc-20783f0f-1a29-493b-8884-9e8158564164   1Gi        RWO            local-path     <unset>                 37m
persistentvolumeclaim/kafka-topology-data-kafka-topology-broker-1       Bound    pvc-604fb4f4-e93f-4c46-a168-6c788b946d7b   1Gi        RWO            local-path     <unset>                 36m
persistentvolumeclaim/kafka-topology-data-kafka-topology-broker-2       Bound    pvc-99e44fbf-a7e9-41ca-a8a5-f9f3d6b8315a   1Gi        RWO            local-path     <unset>                 36m
persistentvolumeclaim/kafka-topology-data-kafka-topology-controller-0   Bound    pvc-af4d1cea-be6f-4a09-bba3-df41308e4935   1Gi        RWO            local-path     <unset>                 37m
persistentvolumeclaim/kafka-topology-data-kafka-topology-controller-1   Bound    pvc-1c72f3c3-5f54-43e8-89c5-967351dbf1f8   1Gi        RWO            local-path     <unset>                 36m
persistentvolumeclaim/kafka-topology-data-kafka-topology-controller-2   Bound    pvc-1a5a3b6c-4d27-49d1-9a12-96bf11418547   1Gi        RWO            local-path     <unset>                 36m

```

Unlike a combined cluster, the pod names now tell you the role directly: `kafka-topology-broker-0/1/2` and
`kafka-topology-controller-0/1/2` are separate pools with their own PetSets (and their own PVCs), sharing a
single headless `kafka-topology-pods` service for inter-node and client traffic.

## Verify Controller Quorum and Partition Leadership

Exec into a broker pod and describe the KRaft metadata quorum:

```bash
$ kubectl exec -it -n kf-demo kafka-topology-broker-0 -- bash
Defaulted container "kafka" out of: kafka, kafka-init (init)
kafka@kafka-topology-broker-0:~$ kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
ClusterId:              11f1-afb2-0e1c891a390w
LeaderId:               1002
LeaderEpoch:            2
HighWatermark:          4468
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   331
CurrentVoters:          [{"id": 1000, "endpoints": ["CONTROLLER://kafka-topology-controller-0.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1001, "endpoints": ["CONTROLLER://kafka-topology-controller-1.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1002, "endpoints": ["CONTROLLER://kafka-topology-controller-2.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}]
CurrentObservers:       [{"id": 1, "directoryId": "g8Ur58TMEmTnlhDRE_-3aA"}, {"id": 0, "directoryId": "w2a0-O46nY2UQTVg-tsh7A"}, {"id": 2, "directoryId": "a3tVkcAY4iZI8VlD1SK4iQ"}]
```

Voter `id`s are offset by `1000` from the controller pod's ordinal (`1000` → `controller-0`, `1001` →
`controller-1`, `1002` → `controller-2`). `LeaderId: 1002` means `kafka-topology-controller-2` is currently
the active controller among the 3 voters — notice the voters are exclusively `controller` pods; broker pods
show up only as `CurrentObservers`, tracking the metadata log without voting rights.

Now create a topic with a replication factor of 3, so every broker holds a copy of the partition:

```bash
kafka@kafka-topology-broker-0:~$ kafka-topics.sh --command-config config/clientauth.properties --create --topic sample --partitions 1 --replication-factor 3 --bootstrap-server localhost:9092
Created topic sample.
kafka@kafka-topology-broker-0:~$ kafka-topics.sh --command-config config/clientauth.properties --describe --topic sample --bootstrap-server localhost:9092
Topic: sample	TopicId: tLoAbMaTQEatzHjGxS1owg	PartitionCount: 1	ReplicationFactor: 3	Configs: min.insync.replicas=1,segment.bytes=1073741824,min.compaction.lag.ms=60000
	Topic: sample	Partition: 0	Leader: 0	Replicas: 0,1,2	Isr: 0,1,2	Elr: 	LastKnownElr: 
kafka@kafka-topology-broker-0:~$ 
```

Broker `0` (`kafka-topology-broker-0`) is the partition leader for `sample-0`, and all three brokers are in
the ISR (fully caught up). Note `min.insync.replicas=1` on this topic — a produce with `acks=all` only needs
1 in-sync replica to acknowledge, so writes can keep going even with just one broker up (more on this in
Case 3).

## Produce and Consume Across the Cluster

```bash
kafka@kafka-topology-broker-1:~$ kafka-console-producer.sh --producer.config config/clientauth.properties --topic sample --request-required-acks all --bootstrap-server localhost:9092
>hello
>hi
```

```bash
kafka@kafka-topology-broker-0:~$ kafka-console-consumer.sh --consumer.config config/clientauth.properties --topic sample --from-beginning --bootstrap-server localhost:9092
hello
hi
```

## How Failover Works

Kafka in KRaft mode runs two independent failover flows. On this **dedicated topology cluster**, they stay
genuinely independent — brokers and controllers are separate pods in separate PetSets, so losing a broker
never touches controller quorum, and losing a controller never touches partition leadership.

**Controller quorum failover.** Controllers replicate a metadata log using Raft. If the active controller
stops responding, the remaining controller voters detect the missing heartbeats and hold an election; a
candidate must get votes from a **majority** of the controller voters to become the new active controller.
With `3` controllers, any `2` form a majority — quorum survives losing `1`. If a majority is unreachable, no
new active controller can be elected, and metadata operations (creating topics, altering partition
assignments, broker registration) stall until quorum returns. Broker pods and existing partition leadership
are unaffected either way.

**Partition leader failover.** For each partition, if the broker currently acting as leader goes down, the
active controller picks a replacement leader from the partition's in-sync replica (ISR) set — follower
brokers that were fully caught up with the old leader — and updates the partition metadata. Because the new
leader was already in-sync, no data already acknowledged (`acks=all`, bounded by `min.insync.replicas`) is
lost. Controller quorum is untouched either way, since controllers are separate pods.

### Hands-on Failover Testing

>note: The Deploy/Verify/Produce steps above were run against a live cluster. The cases below describe the
> expected KRaft/ISR behavior for this topology and are a good next step to verify live yourself — treat
> the exact IDs/output as illustrative.

#### Case 1: Delete the broker holding the partition leader

`kafka-topology-broker-0` is the current partition leader (see the `describe --topic` output above), so
deleting it forces a leader election.

```bash
$ kubectl delete pod -n kf-demo kafka-topology-broker-0
pod "kafka-topology-broker-0" deleted
```

A new partition leader is elected from the remaining ISR members:

```bash
kafka@kafka-topology-broker-1:~$ kafka-topics.sh --command-config config/clientauth.properties --describe --topic sample --bootstrap-server localhost:9092
Topic: sample	TopicId: tLoAbMaTQEatzHjGxS1owg	PartitionCount: 1	ReplicationFactor: 3	Configs: min.insync.replicas=1,segment.bytes=1073741824,min.compaction.lag.ms=60000
	Topic: sample	Partition: 0	Leader: 1	Replicas: 0,1,2	Isr: 1,2	Elr: 	LastKnownElr:
```

Producing and consuming continue uninterrupted against the new leader (`kafka-topology-broker-1`). Controller
quorum is untouched — all 3 controller voters are unaffected by this deletion:

```bash
kafka@kafka-topology-broker-1:~$ kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
LeaderId:               1002
CurrentVoters:          [{"id": 1000, ...}, {"id": 1001, ...}, {"id": 1002, ...}]
```

Once `kafka-topology-broker-0` comes back, it catches up on the partition log and rejoins the ISR as a
follower:

```bash
Topic: sample	Partition: 0	Leader: 1	Replicas: 0,1,2	Isr: 1,2,0
```

#### Case 2: Delete one controller (tolerated)

```bash
$ kubectl delete pod -n kf-demo kafka-topology-controller-0
pod "kafka-topology-controller-0" deleted
```

`kafka-topology-controller-0` wasn't the active controller (`controller-2` still is), so there's nothing to
re-elect here — this just confirms that losing *any one* voter, active or not, is fine as long as a majority
(`2` of `3`) remains. Brokers and partition leadership are entirely unaffected:

```bash
kafka@kafka-topology-broker-0:~$ kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
ClusterId:              11f1-afb2-0e1c891a390w
LeaderId:               1002
LeaderEpoch:            2
HighWatermark:          6380
MaxFollowerLag:         5
MaxFollowerLagTimeMs:   2870
CurrentVoters:          [{"id": 1000, "endpoints": ["CONTROLLER://kafka-topology-controller-0.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1001, "endpoints": ["CONTROLLER://kafka-topology-controller-1.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1002, "endpoints": ["CONTROLLER://kafka-topology-controller-2.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}]
CurrentObservers:       [{"id": 1, "directoryId": "g8Ur58TMEmTnlhDRE_-3aA"}, {"id": 0, "directoryId": "w2a0-O46nY2UQTVg-tsh7A"}, {"id": 2, "directoryId": "a3tVkcAY4iZI8VlD1SK4iQ"}]
kafka@kafka-topology-broker-0:~$ kafka-topics.sh --command-config config/clientauth.properties --describe --topic sample --bootstrap-server localhost:9092
Topic: sample	TopicId: tLoAbMaTQEatzHjGxS1owg	PartitionCount: 1	ReplicationFactor: 3	Configs: min.insync.replicas=1,segment.bytes=1073741824,min.compaction.lag.ms=60000
	Topic: sample	Partition: 0	Leader: 0	Replicas: 0,1,2	Isr: 0,1,2	Elr: 	LastKnownElr: 
```

Once `kafka-topology-controller-0` comes back, it rejoins the quorum as a voter.

#### Case 3: Delete a second controller (quorum lost)

```bash
$ kubectl delete pod -n kf-demo kafka-topology-controller-1
pod "kafka-topology-controller-1" deleted
```

With only `1` of `3` controller voters left, quorum is lost — no new active controller can be elected, and
metadata operations (new topics, partition reassignment, broker (re)registration) stall:

```bash
kafka@kafka-topology-broker-0:~$ kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
ClusterId:              11f1-afb2-0e1c891a390w
LeaderId:               1002
LeaderEpoch:            2
HighWatermark:          6535
MaxFollowerLag:         0
MaxFollowerLagTimeMs:   24
CurrentVoters:          [{"id": 1000, "endpoints": ["CONTROLLER://kafka-topology-controller-0.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1001, "endpoints": ["CONTROLLER://kafka-topology-controller-1.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}, {"id": 1002, "endpoints": ["CONTROLLER://kafka-topology-controller-2.kafka-topology-pods.kf-demo.svc.cluster.local:9093"]}]
CurrentObservers:       [{"id": 1, "directoryId": "g8Ur58TMEmTnlhDRE_-3aA"}, {"id": 0, "directoryId": "w2a0-O46nY2UQTVg-tsh7A"}, {"id": 2, "directoryId": "a3tVkcAY4iZI8VlD1SK4iQ"}]

```

Because `min.insync.replicas=1` on the `sample` topic, and the broker pool is completely unaffected by this
controller-only failure, produce/consume traffic keeps flowing normally throughout — the data plane and the
controller quorum are separate concerns on a topology cluster. What stalls is anything that requires the
controller: electing a new partition leader if the current one also fails, creating topics, or letting the
two deleted controllers rejoin as recognized voters. Once at least one of
`kafka-topology-controller-0`/`kafka-topology-controller-1` returns (restoring `2` of `3`), quorum is
restored.

#### Case 4: Delete every pod in the cluster

```bash
$ kubectl delete pod -n kf-demo -l app.kubernetes.io/instance=kafka-topology
pod "kafka-topology-broker-0" deleted
pod "kafka-topology-broker-1" deleted
pod "kafka-topology-broker-2" deleted
pod "kafka-topology-controller-0" deleted
pod "kafka-topology-controller-1" deleted
pod "kafka-topology-controller-2" deleted
```

As both PetSets bring their pods back up (each reattaching its own PVC), the controller pool rejoins the
metadata quorum once a majority is reachable and elects an active controller, while the broker pool resyncs
partition data from disk. Produce/consume resumes once a partition leader is elected again.

```bash
$ watch kubectl exec -it -n kf-demo kafka-topology-broker-0 -- kafka-metadata-quorum.sh --command-config config/clientauth.properties --bootstrap-server localhost:9092 describe --status
```

## CleanUp

For cleaning up what we created in this tutorial follow the following commands:

```bash
$ kubectl patch -n kf-demo kf kafka-topology -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kafka.kubedb.com/kafka-topology patched

$ kubectl delete kf -n kf-demo kafka-topology
kafka.kubedb.com "kafka-topology" deleted

$ kubectl delete ns kf-demo
namespace "kf-demo" deleted
```

## Next Steps

- Deploy a [combined cluster](/docs/guides/kafka/clustering/combined-cluster/index.md) for Apache Kafka, for a lighter-weight dev/test setup.
- Monitor your Kafka cluster with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Learn to use KubeDB managed Kafka objects using [CLIs](/docs/guides/kafka/cli/cli.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
