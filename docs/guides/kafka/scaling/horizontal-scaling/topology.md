---
title: Horizontal Scaling Topology Kafka
menu:
  docs_{{ .version }}:
    identifier: kf-horizontal-scaling-topology
    name: Topology Cluster
    parent: kf-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Kafka Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Kafka topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [Topology](/docs/guides/kafka/clustering/topology-cluster/index.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Horizontal Scaling Overview](/docs/guides/kafka/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Topology Cluster

Here, we are going to deploy a `Kafka` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Kafka Topology cluster

Now, we are going to deploy a `Kafka` topology cluster with version `3.6.1`.

### Deploy Kafka topology cluster

In this section, we are going to deploy a Kafka topology cluster. Then, in the next section we will scale the cluster using `KafkaOpsRequest` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.6.1
  topology:
    broker:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Kafka` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/scaling/kafka-topology.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait until `kafka-prod` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
kafka-prod    kubedb.com/v1   3.6.1     Provisioning   0s
kafka-prod    kubedb.com/v1   3.6.1     Provisioning   24s
.
.
kafka-prod    kubedb.com/v1   3.6.1     Ready          92s
```

Let's check the number of replicas has from kafka object, number of pods the petset have,

**Broker Replicas**

```bash
$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.broker.replicas'
2

$ kubectl get petset -n demo kafka-prod-broker -o json | jq '.spec.replicas'
2
```

**Controller Replicas**

```bash
$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.controller.replicas'
2

$ kubectl get petset -n demo kafka-prod-controller -o json | jq '.spec.replicas'
2
```

We can see from commands that the cluster has 2 replicas for both broker and controller.

Also, we can verify the replicas of the topology from an internal kafka command by exec into a replica.

Now let's exec to a broker instance and run a kafka internal command to check the number of replicas for broker and controller.,

**Broker**

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-broker-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
kafka-prod-broker-0.kafka-prod-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
	Produce(0): 0 to 9 [usable: 9],
	Fetch(1): 0 to 15 [usable: 15],
	ListOffsets(2): 0 to 8 [usable: 8],
	Metadata(3): 0 to 12 [usable: 12],
	LeaderAndIsr(4): UNSUPPORTED,
	StopReplica(5): UNSUPPORTED,
	UpdateMetadata(6): UNSUPPORTED,
	ControlledShutdown(7): UNSUPPORTED,
	OffsetCommit(8): 0 to 8 [usable: 8],
	OffsetFetch(9): 0 to 8 [usable: 8],
	FindCoordinator(10): 0 to 4 [usable: 4],
	JoinGroup(11): 0 to 9 [usable: 9],
	Heartbeat(12): 0 to 4 [usable: 4],
	LeaveGroup(13): 0 to 5 [usable: 5],
	SyncGroup(14): 0 to 5 [usable: 5],
	DescribeGroups(15): 0 to 5 [usable: 5],
	ListGroups(16): 0 to 4 [usable: 4],
	SaslHandshake(17): 0 to 1 [usable: 1],
	ApiVersions(18): 0 to 3 [usable: 3],
	CreateTopics(19): 0 to 7 [usable: 7],
	DeleteTopics(20): 0 to 6 [usable: 6],
	DeleteRecords(21): 0 to 2 [usable: 2],
	InitProducerId(22): 0 to 4 [usable: 4],
	OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
	AddPartitionsToTxn(24): 0 to 4 [usable: 4],
	AddOffsetsToTxn(25): 0 to 3 [usable: 3],
	EndTxn(26): 0 to 3 [usable: 3],
	WriteTxnMarkers(27): 0 to 1 [usable: 1],
	TxnOffsetCommit(28): 0 to 3 [usable: 3],
	DescribeAcls(29): 0 to 3 [usable: 3],
	CreateAcls(30): 0 to 3 [usable: 3],
	DeleteAcls(31): 0 to 3 [usable: 3],
	DescribeConfigs(32): 0 to 4 [usable: 4],
	AlterConfigs(33): 0 to 2 [usable: 2],
	AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
	DescribeLogDirs(35): 0 to 4 [usable: 4],
	SaslAuthenticate(36): 0 to 2 [usable: 2],
	CreatePartitions(37): 0 to 3 [usable: 3],
	CreateDelegationToken(38): 0 to 3 [usable: 3],
	RenewDelegationToken(39): 0 to 2 [usable: 2],
	ExpireDelegationToken(40): 0 to 2 [usable: 2],
	DescribeDelegationToken(41): 0 to 3 [usable: 3],
	DeleteGroups(42): 0 to 2 [usable: 2],
	ElectLeaders(43): 0 to 2 [usable: 2],
	IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
	AlterPartitionReassignments(45): 0 [usable: 0],
	ListPartitionReassignments(46): 0 [usable: 0],
	OffsetDelete(47): 0 [usable: 0],
	DescribeClientQuotas(48): 0 to 1 [usable: 1],
	AlterClientQuotas(49): 0 to 1 [usable: 1],
	DescribeUserScramCredentials(50): 0 [usable: 0],
	AlterUserScramCredentials(51): 0 [usable: 0],
	DescribeQuorum(55): 0 to 1 [usable: 1],
	AlterPartition(56): UNSUPPORTED,
	UpdateFeatures(57): 0 to 1 [usable: 1],
	Envelope(58): UNSUPPORTED,
	DescribeCluster(60): 0 [usable: 0],
	DescribeProducers(61): 0 [usable: 0],
	UnregisterBroker(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
kafka-prod-broker-1.kafka-prod-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
	Produce(0): 0 to 9 [usable: 9],
	Fetch(1): 0 to 15 [usable: 15],
	ListOffsets(2): 0 to 8 [usable: 8],
	Metadata(3): 0 to 12 [usable: 12],
	LeaderAndIsr(4): UNSUPPORTED,
	StopReplica(5): UNSUPPORTED,
	UpdateMetadata(6): UNSUPPORTED,
	ControlledShutdown(7): UNSUPPORTED,
	OffsetCommit(8): 0 to 8 [usable: 8],
	OffsetFetch(9): 0 to 8 [usable: 8],
	FindCoordinator(10): 0 to 4 [usable: 4],
	JoinGroup(11): 0 to 9 [usable: 9],
	Heartbeat(12): 0 to 4 [usable: 4],
	LeaveGroup(13): 0 to 5 [usable: 5],
	SyncGroup(14): 0 to 5 [usable: 5],
	DescribeGroups(15): 0 to 5 [usable: 5],
	ListGroups(16): 0 to 4 [usable: 4],
	SaslHandshake(17): 0 to 1 [usable: 1],
	ApiVersions(18): 0 to 3 [usable: 3],
	CreateTopics(19): 0 to 7 [usable: 7],
	DeleteTopics(20): 0 to 6 [usable: 6],
	DeleteRecords(21): 0 to 2 [usable: 2],
	InitProducerId(22): 0 to 4 [usable: 4],
	OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
	AddPartitionsToTxn(24): 0 to 4 [usable: 4],
	AddOffsetsToTxn(25): 0 to 3 [usable: 3],
	EndTxn(26): 0 to 3 [usable: 3],
	WriteTxnMarkers(27): 0 to 1 [usable: 1],
	TxnOffsetCommit(28): 0 to 3 [usable: 3],
	DescribeAcls(29): 0 to 3 [usable: 3],
	CreateAcls(30): 0 to 3 [usable: 3],
	DeleteAcls(31): 0 to 3 [usable: 3],
	DescribeConfigs(32): 0 to 4 [usable: 4],
	AlterConfigs(33): 0 to 2 [usable: 2],
	AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
	DescribeLogDirs(35): 0 to 4 [usable: 4],
	SaslAuthenticate(36): 0 to 2 [usable: 2],
	CreatePartitions(37): 0 to 3 [usable: 3],
	CreateDelegationToken(38): 0 to 3 [usable: 3],
	RenewDelegationToken(39): 0 to 2 [usable: 2],
	ExpireDelegationToken(40): 0 to 2 [usable: 2],
	DescribeDelegationToken(41): 0 to 3 [usable: 3],
	DeleteGroups(42): 0 to 2 [usable: 2],
	ElectLeaders(43): 0 to 2 [usable: 2],
	IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
	AlterPartitionReassignments(45): 0 [usable: 0],
	ListPartitionReassignments(46): 0 [usable: 0],
	OffsetDelete(47): 0 [usable: 0],
	DescribeClientQuotas(48): 0 to 1 [usable: 1],
	AlterClientQuotas(49): 0 to 1 [usable: 1],
	DescribeUserScramCredentials(50): 0 [usable: 0],
	AlterUserScramCredentials(51): 0 [usable: 0],
	DescribeQuorum(55): 0 to 1 [usable: 1],
	AlterPartition(56): UNSUPPORTED,
	UpdateFeatures(57): 0 to 1 [usable: 1],
	Envelope(58): UNSUPPORTED,
	DescribeCluster(60): 0 [usable: 0],
	DescribeProducers(61): 0 [usable: 0],
	UnregisterBroker(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
```

**Controller**

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-metadata-quorum.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties describe --status | grep CurrentObservers
CurrentObservers:       [0,1]
```

We can see from the above output that the kafka has 2 nodes for broker and 2 nodes for controller.

We are now ready to apply the `KafkaOpsRequest` CR to scale this cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the topology cluster to meet the desired number of replicas after scaling.

#### Create KafkaOpsRequest

In order to scale up the replicas of the topology cluster, we have to create a `KafkaOpsRequest` CR with our desired replicas. Below is the YAML of the `KafkaOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-hscale-up-topology
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: kafka-prod
  horizontalScaling:
    topology: 
      broker: 3
      controller: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on kafka.
- `spec.horizontalScaling.topology.broker` specifies the desired replicas after scaling for broker.
- `spec.horizontalScaling.topology.controller` specifies the desired replicas after scaling for controller.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/scaling/horizontal-scaling/kafka-hscale-up-topology.yaml
kafkaopsrequest.ops.kubedb.com/kfops-hscale-up-topology created
```

> **Note:** If you want to scale down only broker or controller, you can specify the desired replicas for only broker or controller in the `KafkaOpsRequest` CR. You can specify one at a time. If you want to scale broker only, no node will need restart to apply the changes. But if you want to scale controller, all nodes will need restart to apply the changes.

#### Verify Topology cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Kafka` object and related `PetSets` and `Pods`.

Let's wait for `KafkaOpsRequest` to be `Successful`. Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ watch kubectl get kafkaopsrequest -n demo
NAME                        TYPE                STATUS       AGE
kfops-hscale-up-topology    HorizontalScaling   Successful   106s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-hscale-up-topology 
Name:         kfops-hscale-up-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T11:02:51Z
  Generation:          1
  Resource Version:    356503
  UID:                 44e0db0c-2094-4c13-a3be-9ca680888545
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Horizontal Scaling:
    Topology:
      Broker:      3
      Controller:  3
  Type:            HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-08-02T11:02:51Z
    Message:               Kafka ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-08-02T11:02:59Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T11:03:00Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T11:03:09Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T11:03:14Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T11:03:14Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T11:03:24Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T11:03:29Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T11:03:30Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T11:03:59Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T11:04:04Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T11:04:05Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T11:04:19Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T11:04:24Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T11:04:59Z
    Message:               Successfully Scaled Up Broker
    Observed Generation:   1
    Reason:                ScaleUpBroker
    Status:                True
    Type:                  ScaleUpBroker
    Last Transition Time:  2024-08-02T11:04:30Z
    Message:               patch pet setkafka-prod-broker; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSetkafka-prod-broker
    Last Transition Time:  2024-08-02T11:04:55Z
    Message:               node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  NodeInCluster
    Last Transition Time:  2024-08-02T11:05:15Z
    Message:               Successfully Scaled Up Controller
    Observed Generation:   1
    Reason:                ScaleUpController
    Status:                True
    Type:                  ScaleUpController
    Last Transition Time:  2024-08-02T11:05:05Z
    Message:               patch pet setkafka-prod-controller; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSetkafka-prod-controller
    Last Transition Time:  2024-08-02T11:05:15Z
    Message:               Successfully completed horizontally scale kafka cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   4m19s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-hscale-up-topology
  Normal   Starting                                                                   4m19s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 4m19s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-hscale-up-topology
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             4m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           4m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  4m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   4m1s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             3m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           3m56s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  3m50s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   3m46s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 3m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               3m41s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      3m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       3m11s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 3m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               3m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      3m1s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       2m51s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               2m46s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Warning  patch pet setkafka-prod-broker; ConditionStatus:True                       2m40s  KubeDB Ops-manager Operator  patch pet setkafka-prod-broker; ConditionStatus:True
  Warning  node in cluster; ConditionStatus:False                                     2m36s  KubeDB Ops-manager Operator  node in cluster; ConditionStatus:False
  Warning  node in cluster; ConditionStatus:True                                      2m15s  KubeDB Ops-manager Operator  node in cluster; ConditionStatus:True
  Normal   ScaleUpBroker                                                              2m10s  KubeDB Ops-manager Operator  Successfully Scaled Up Broker
  Warning  patch pet setkafka-prod-controller; ConditionStatus:True                   2m5s   KubeDB Ops-manager Operator  patch pet setkafka-prod-controller; ConditionStatus:True
  Warning  node in cluster; ConditionStatus:True                                      2m     KubeDB Ops-manager Operator  node in cluster; ConditionStatus:True
  Normal   ScaleUpController                                                          115s   KubeDB Ops-manager Operator  Successfully Scaled Up Controller
  Normal   Starting                                                                   115s   KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 115s   KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-hscale-up-topology
```

Now, we are going to verify the number of replicas this cluster has from the Kafka object, number of pods the petset have,

**Broker Replicas**

```bash
$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.broker.replicas'
3

$ kubectl get petset -n demo kafka-prod-broker -o json | jq '.spec.replicas'
3
```

Now let's connect to a kafka instance and run a kafka internal command to check the number of replicas of topology cluster for both broker and controller.,

**Broker**

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-broker-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
kafka-prod-broker-0.kafka-prod-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
	Produce(0): 0 to 9 [usable: 9],
	Fetch(1): 0 to 15 [usable: 15],
	ListOffsets(2): 0 to 8 [usable: 8],
	Metadata(3): 0 to 12 [usable: 12],
	LeaderAndIsr(4): UNSUPPORTED,
	StopReplica(5): UNSUPPORTED,
	UpdateMetadata(6): UNSUPPORTED,
	ControlledShutdown(7): UNSUPPORTED,
	OffsetCommit(8): 0 to 8 [usable: 8],
	OffsetFetch(9): 0 to 8 [usable: 8],
	FindCoordinator(10): 0 to 4 [usable: 4],
	JoinGroup(11): 0 to 9 [usable: 9],
	Heartbeat(12): 0 to 4 [usable: 4],
	LeaveGroup(13): 0 to 5 [usable: 5],
	SyncGroup(14): 0 to 5 [usable: 5],
	DescribeGroups(15): 0 to 5 [usable: 5],
	ListGroups(16): 0 to 4 [usable: 4],
	SaslHandshake(17): 0 to 1 [usable: 1],
	ApiVersions(18): 0 to 3 [usable: 3],
	CreateTopics(19): 0 to 7 [usable: 7],
	DeleteTopics(20): 0 to 6 [usable: 6],
	DeleteRecords(21): 0 to 2 [usable: 2],
	InitProducerId(22): 0 to 4 [usable: 4],
	OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
	AddPartitionsToTxn(24): 0 to 4 [usable: 4],
	AddOffsetsToTxn(25): 0 to 3 [usable: 3],
	EndTxn(26): 0 to 3 [usable: 3],
	WriteTxnMarkers(27): 0 to 1 [usable: 1],
	TxnOffsetCommit(28): 0 to 3 [usable: 3],
	DescribeAcls(29): 0 to 3 [usable: 3],
	CreateAcls(30): 0 to 3 [usable: 3],
	DeleteAcls(31): 0 to 3 [usable: 3],
	DescribeConfigs(32): 0 to 4 [usable: 4],
	AlterConfigs(33): 0 to 2 [usable: 2],
	AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
	DescribeLogDirs(35): 0 to 4 [usable: 4],
	SaslAuthenticate(36): 0 to 2 [usable: 2],
	CreatePartitions(37): 0 to 3 [usable: 3],
	CreateDelegationToken(38): 0 to 3 [usable: 3],
	RenewDelegationToken(39): 0 to 2 [usable: 2],
	ExpireDelegationToken(40): 0 to 2 [usable: 2],
	DescribeDelegationToken(41): 0 to 3 [usable: 3],
	DeleteGroups(42): 0 to 2 [usable: 2],
	ElectLeaders(43): 0 to 2 [usable: 2],
	IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
	AlterPartitionReassignments(45): 0 [usable: 0],
	ListPartitionReassignments(46): 0 [usable: 0],
	OffsetDelete(47): 0 [usable: 0],
	DescribeClientQuotas(48): 0 to 1 [usable: 1],
	AlterClientQuotas(49): 0 to 1 [usable: 1],
	DescribeUserScramCredentials(50): 0 [usable: 0],
	AlterUserScramCredentials(51): 0 [usable: 0],
	DescribeQuorum(55): 0 to 1 [usable: 1],
	AlterPartition(56): UNSUPPORTED,
	UpdateFeatures(57): 0 to 1 [usable: 1],
	Envelope(58): UNSUPPORTED,
	DescribeCluster(60): 0 [usable: 0],
	DescribeProducers(61): 0 [usable: 0],
	UnregisterBroker(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
kafka-prod-broker-1.kafka-prod-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
	Produce(0): 0 to 9 [usable: 9],
	Fetch(1): 0 to 15 [usable: 15],
	ListOffsets(2): 0 to 8 [usable: 8],
	Metadata(3): 0 to 12 [usable: 12],
	LeaderAndIsr(4): UNSUPPORTED,
	StopReplica(5): UNSUPPORTED,
	UpdateMetadata(6): UNSUPPORTED,
	ControlledShutdown(7): UNSUPPORTED,
	OffsetCommit(8): 0 to 8 [usable: 8],
	OffsetFetch(9): 0 to 8 [usable: 8],
	FindCoordinator(10): 0 to 4 [usable: 4],
	JoinGroup(11): 0 to 9 [usable: 9],
	Heartbeat(12): 0 to 4 [usable: 4],
	LeaveGroup(13): 0 to 5 [usable: 5],
	SyncGroup(14): 0 to 5 [usable: 5],
	DescribeGroups(15): 0 to 5 [usable: 5],
	ListGroups(16): 0 to 4 [usable: 4],
	SaslHandshake(17): 0 to 1 [usable: 1],
	ApiVersions(18): 0 to 3 [usable: 3],
	CreateTopics(19): 0 to 7 [usable: 7],
	DeleteTopics(20): 0 to 6 [usable: 6],
	DeleteRecords(21): 0 to 2 [usable: 2],
	InitProducerId(22): 0 to 4 [usable: 4],
	OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
	AddPartitionsToTxn(24): 0 to 4 [usable: 4],
	AddOffsetsToTxn(25): 0 to 3 [usable: 3],
	EndTxn(26): 0 to 3 [usable: 3],
	WriteTxnMarkers(27): 0 to 1 [usable: 1],
	TxnOffsetCommit(28): 0 to 3 [usable: 3],
	DescribeAcls(29): 0 to 3 [usable: 3],
	CreateAcls(30): 0 to 3 [usable: 3],
	DeleteAcls(31): 0 to 3 [usable: 3],
	DescribeConfigs(32): 0 to 4 [usable: 4],
	AlterConfigs(33): 0 to 2 [usable: 2],
	AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
	DescribeLogDirs(35): 0 to 4 [usable: 4],
	SaslAuthenticate(36): 0 to 2 [usable: 2],
	CreatePartitions(37): 0 to 3 [usable: 3],
	CreateDelegationToken(38): 0 to 3 [usable: 3],
	RenewDelegationToken(39): 0 to 2 [usable: 2],
	ExpireDelegationToken(40): 0 to 2 [usable: 2],
	DescribeDelegationToken(41): 0 to 3 [usable: 3],
	DeleteGroups(42): 0 to 2 [usable: 2],
	ElectLeaders(43): 0 to 2 [usable: 2],
	IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
	AlterPartitionReassignments(45): 0 [usable: 0],
	ListPartitionReassignments(46): 0 [usable: 0],
	OffsetDelete(47): 0 [usable: 0],
	DescribeClientQuotas(48): 0 to 1 [usable: 1],
	AlterClientQuotas(49): 0 to 1 [usable: 1],
	DescribeUserScramCredentials(50): 0 [usable: 0],
	AlterUserScramCredentials(51): 0 [usable: 0],
	DescribeQuorum(55): 0 to 1 [usable: 1],
	AlterPartition(56): UNSUPPORTED,
	UpdateFeatures(57): 0 to 1 [usable: 1],
	Envelope(58): UNSUPPORTED,
	DescribeCluster(60): 0 [usable: 0],
	DescribeProducers(61): 0 [usable: 0],
	UnregisterBroker(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
kafka-prod-broker-2.kafka-prod-pods.demo.svc.cluster.local:9092 (id: 2 rack: null) -> (
	Produce(0): 0 to 9 [usable: 9],
	Fetch(1): 0 to 15 [usable: 15],
	ListOffsets(2): 0 to 8 [usable: 8],
	Metadata(3): 0 to 12 [usable: 12],
	LeaderAndIsr(4): UNSUPPORTED,
	StopReplica(5): UNSUPPORTED,
	UpdateMetadata(6): UNSUPPORTED,
	ControlledShutdown(7): UNSUPPORTED,
	OffsetCommit(8): 0 to 8 [usable: 8],
	OffsetFetch(9): 0 to 8 [usable: 8],
	FindCoordinator(10): 0 to 4 [usable: 4],
	JoinGroup(11): 0 to 9 [usable: 9],
	Heartbeat(12): 0 to 4 [usable: 4],
	LeaveGroup(13): 0 to 5 [usable: 5],
	SyncGroup(14): 0 to 5 [usable: 5],
	DescribeGroups(15): 0 to 5 [usable: 5],
	ListGroups(16): 0 to 4 [usable: 4],
	SaslHandshake(17): 0 to 1 [usable: 1],
	ApiVersions(18): 0 to 3 [usable: 3],
	CreateTopics(19): 0 to 7 [usable: 7],
	DeleteTopics(20): 0 to 6 [usable: 6],
	DeleteRecords(21): 0 to 2 [usable: 2],
	InitProducerId(22): 0 to 4 [usable: 4],
	OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
	AddPartitionsToTxn(24): 0 to 4 [usable: 4],
	AddOffsetsToTxn(25): 0 to 3 [usable: 3],
	EndTxn(26): 0 to 3 [usable: 3],
	WriteTxnMarkers(27): 0 to 1 [usable: 1],
	TxnOffsetCommit(28): 0 to 3 [usable: 3],
	DescribeAcls(29): 0 to 3 [usable: 3],
	CreateAcls(30): 0 to 3 [usable: 3],
	DeleteAcls(31): 0 to 3 [usable: 3],
	DescribeConfigs(32): 0 to 4 [usable: 4],
	AlterConfigs(33): 0 to 2 [usable: 2],
	AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
	DescribeLogDirs(35): 0 to 4 [usable: 4],
	SaslAuthenticate(36): 0 to 2 [usable: 2],
	CreatePartitions(37): 0 to 3 [usable: 3],
	CreateDelegationToken(38): 0 to 3 [usable: 3],
	RenewDelegationToken(39): 0 to 2 [usable: 2],
	ExpireDelegationToken(40): 0 to 2 [usable: 2],
	DescribeDelegationToken(41): 0 to 3 [usable: 3],
	DeleteGroups(42): 0 to 2 [usable: 2],
	ElectLeaders(43): 0 to 2 [usable: 2],
	IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
	AlterPartitionReassignments(45): 0 [usable: 0],
	ListPartitionReassignments(46): 0 [usable: 0],
	OffsetDelete(47): 0 [usable: 0],
	DescribeClientQuotas(48): 0 to 1 [usable: 1],
	AlterClientQuotas(49): 0 to 1 [usable: 1],
	DescribeUserScramCredentials(50): 0 [usable: 0],
	AlterUserScramCredentials(51): 0 [usable: 0],
	DescribeQuorum(55): 0 to 1 [usable: 1],
	AlterPartition(56): UNSUPPORTED,
	UpdateFeatures(57): 0 to 1 [usable: 1],
	Envelope(58): UNSUPPORTED,
	DescribeCluster(60): 0 [usable: 0],
	DescribeProducers(61): 0 [usable: 0],
	UnregisterBroker(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
```

**Controller**

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-metadata-quorum.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties describe --status | grep CurrentObservers
CurrentObservers:       [0,1,2]
```

From all the above outputs we can see that the both brokers and controller of the topology kafka is `3`. That means we have successfully scaled up the replicas of the Kafka topology cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the kafka topology cluster to meet the desired number of replicas after scaling.

#### Create KafkaOpsRequest

In order to scale down the replicas of the kafka topology cluster, we have to create a `KafkaOpsRequest` CR with our desired replicas. Below is the YAML of the `KafkaOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-hscale-down-topology
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: kafka-prod
  horizontalScaling:
    topology:
      broker: 2
      controller: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on kafka.
- `spec.horizontalScaling.topology.broker` specifies the desired replicas after scaling for the broker nodes.
- `spec.horizontalScaling.topology.controller` specifies the desired replicas after scaling for the controller nodes.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/scaling/horizontal-scaling/kafka-hscale-down-topology.yaml
kafkaopsrequest.ops.kubedb.com/kfops-hscale-down-topology created
```

#### Verify Topology cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Kafka` object and related `PetSets` and `Pods`.

Let's wait for `KafkaOpsRequest` to be `Successful`. Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ watch kubectl get kafkaopsrequest -n demo
NAME                          TYPE                STATUS       AGE
kfops-hscale-down-topology    HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-hscale-down-topology
Name:         kfops-hscale-down-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T11:14:18Z
  Generation:          1
  Resource Version:    357545
  UID:                 b786d791-6ba8-4f1c-ade8-9443e049cede
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Horizontal Scaling:
    Topology:
      Broker:      2
      Controller:  2
  Type:            HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-08-02T11:14:18Z
    Message:               Kafka ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-08-02T11:14:46Z
    Message:               Successfully Scaled Down Broker
    Observed Generation:   1
    Reason:                ScaleDownBroker
    Status:                True
    Type:                  ScaleDownBroker
    Last Transition Time:  2024-08-02T11:14:36Z
    Message:               reassign partitions; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReassignPartitions
    Last Transition Time:  2024-08-02T11:14:36Z
    Message:               is pet set patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetPatched
    Last Transition Time:  2024-08-02T11:14:37Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-08-02T11:14:37Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2024-08-02T11:15:26Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-08-02T11:15:31Z
    Message:               Successfully Scaled Down Controller
    Observed Generation:   1
    Reason:                ScaleDownController
    Status:                True
    Type:                  ScaleDownController
    Last Transition Time:  2024-08-02T11:15:38Z
    Message:               successfully reconciled the Kafka with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-02T11:15:43Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T11:15:44Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T11:15:53Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T11:15:58Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T11:15:58Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T11:16:08Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T11:16:13Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T11:16:13Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T11:16:58Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T11:17:03Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T11:17:03Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T11:17:13Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T11:17:18Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T11:17:19Z
    Message:               Successfully completed horizontally scale kafka cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   8m35s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-hscale-down-topology
  Normal   Starting                                                                   8m35s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 8m35s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-hscale-down-topology
  Warning  reassign partitions; ConditionStatus:True                                  8m17s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True                                   8m17s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                                              8m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           8m16s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                                             8m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                                              8m12s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           8m12s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                                              8m12s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   ScaleDownBroker                                                            8m7s   KubeDB Ops-manager Operator  Successfully Scaled Down Broker
  Warning  reassign partitions; ConditionStatus:True                                  7m31s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True                                   7m31s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                                              7m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           7m31s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                                             7m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                                              7m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           7m27s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                                              7m27s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   ScaleDownController                                                        7m22s  KubeDB Ops-manager Operator  Successfully Scaled Down Controller
  Normal   UpdatePetSets                                                              7m15s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with modified node
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             7m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           7m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  7m5s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   7m     KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             6m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           6m55s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  6m50s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   6m45s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 6m40s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               6m40s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      6m35s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       5m55s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 5m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               5m50s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      5m45s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       5m40s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               5m35s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   5m35s  KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 5m34s  KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-hscale-down-topology
```

Now, we are going to verify the number of replicas this cluster has from the Kafka object, number of pods the petset have,

**Broker Replicas**

```bash
$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.broker.replicas' 
2

$ kubectl get petset -n demo kafka-prod-broker -o json | jq '.spec.replicas'
2
```

**Controller Replicas**

```bash
$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.controller.replicas'
2

$ kubectl get petset -n demo kafka-prod-controller -o json | jq '.spec.replicas'
2
```

Now let's connect to a kafka instance and run a kafka internal command to check the number of replicas for both broker and controller nodes,

**Broker**

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-broker-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
kafka-prod-broker-0.kafka-prod-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
	Produce(0): 0 to 9 [usable: 9],
	Fetch(1): 0 to 15 [usable: 15],
	ListOffsets(2): 0 to 8 [usable: 8],
	Metadata(3): 0 to 12 [usable: 12],
	LeaderAndIsr(4): UNSUPPORTED,
	StopReplica(5): UNSUPPORTED,
	UpdateMetadata(6): UNSUPPORTED,
	ControlledShutdown(7): UNSUPPORTED,
	OffsetCommit(8): 0 to 8 [usable: 8],
	OffsetFetch(9): 0 to 8 [usable: 8],
	FindCoordinator(10): 0 to 4 [usable: 4],
	JoinGroup(11): 0 to 9 [usable: 9],
	Heartbeat(12): 0 to 4 [usable: 4],
	LeaveGroup(13): 0 to 5 [usable: 5],
	SyncGroup(14): 0 to 5 [usable: 5],
	DescribeGroups(15): 0 to 5 [usable: 5],
	ListGroups(16): 0 to 4 [usable: 4],
	SaslHandshake(17): 0 to 1 [usable: 1],
	ApiVersions(18): 0 to 3 [usable: 3],
	CreateTopics(19): 0 to 7 [usable: 7],
	DeleteTopics(20): 0 to 6 [usable: 6],
	DeleteRecords(21): 0 to 2 [usable: 2],
	InitProducerId(22): 0 to 4 [usable: 4],
	OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
	AddPartitionsToTxn(24): 0 to 4 [usable: 4],
	AddOffsetsToTxn(25): 0 to 3 [usable: 3],
	EndTxn(26): 0 to 3 [usable: 3],
	WriteTxnMarkers(27): 0 to 1 [usable: 1],
	TxnOffsetCommit(28): 0 to 3 [usable: 3],
	DescribeAcls(29): 0 to 3 [usable: 3],
	CreateAcls(30): 0 to 3 [usable: 3],
	DeleteAcls(31): 0 to 3 [usable: 3],
	DescribeConfigs(32): 0 to 4 [usable: 4],
	AlterConfigs(33): 0 to 2 [usable: 2],
	AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
	DescribeLogDirs(35): 0 to 4 [usable: 4],
	SaslAuthenticate(36): 0 to 2 [usable: 2],
	CreatePartitions(37): 0 to 3 [usable: 3],
	CreateDelegationToken(38): 0 to 3 [usable: 3],
	RenewDelegationToken(39): 0 to 2 [usable: 2],
	ExpireDelegationToken(40): 0 to 2 [usable: 2],
	DescribeDelegationToken(41): 0 to 3 [usable: 3],
	DeleteGroups(42): 0 to 2 [usable: 2],
	ElectLeaders(43): 0 to 2 [usable: 2],
	IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
	AlterPartitionReassignments(45): 0 [usable: 0],
	ListPartitionReassignments(46): 0 [usable: 0],
	OffsetDelete(47): 0 [usable: 0],
	DescribeClientQuotas(48): 0 to 1 [usable: 1],
	AlterClientQuotas(49): 0 to 1 [usable: 1],
	DescribeUserScramCredentials(50): 0 [usable: 0],
	AlterUserScramCredentials(51): 0 [usable: 0],
	DescribeQuorum(55): 0 to 1 [usable: 1],
	AlterPartition(56): UNSUPPORTED,
	UpdateFeatures(57): 0 to 1 [usable: 1],
	Envelope(58): UNSUPPORTED,
	DescribeCluster(60): 0 [usable: 0],
	DescribeProducers(61): 0 [usable: 0],
	UnregisterBroker(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
kafka-prod-broker-1.kafka-prod-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
	Produce(0): 0 to 9 [usable: 9],
	Fetch(1): 0 to 15 [usable: 15],
	ListOffsets(2): 0 to 8 [usable: 8],
	Metadata(3): 0 to 12 [usable: 12],
	LeaderAndIsr(4): UNSUPPORTED,
	StopReplica(5): UNSUPPORTED,
	UpdateMetadata(6): UNSUPPORTED,
	ControlledShutdown(7): UNSUPPORTED,
	OffsetCommit(8): 0 to 8 [usable: 8],
	OffsetFetch(9): 0 to 8 [usable: 8],
	FindCoordinator(10): 0 to 4 [usable: 4],
	JoinGroup(11): 0 to 9 [usable: 9],
	Heartbeat(12): 0 to 4 [usable: 4],
	LeaveGroup(13): 0 to 5 [usable: 5],
	SyncGroup(14): 0 to 5 [usable: 5],
	DescribeGroups(15): 0 to 5 [usable: 5],
	ListGroups(16): 0 to 4 [usable: 4],
	SaslHandshake(17): 0 to 1 [usable: 1],
	ApiVersions(18): 0 to 3 [usable: 3],
	CreateTopics(19): 0 to 7 [usable: 7],
	DeleteTopics(20): 0 to 6 [usable: 6],
	DeleteRecords(21): 0 to 2 [usable: 2],
	InitProducerId(22): 0 to 4 [usable: 4],
	OffsetForLeaderEpoch(23): 0 to 4 [usable: 4],
	AddPartitionsToTxn(24): 0 to 4 [usable: 4],
	AddOffsetsToTxn(25): 0 to 3 [usable: 3],
	EndTxn(26): 0 to 3 [usable: 3],
	WriteTxnMarkers(27): 0 to 1 [usable: 1],
	TxnOffsetCommit(28): 0 to 3 [usable: 3],
	DescribeAcls(29): 0 to 3 [usable: 3],
	CreateAcls(30): 0 to 3 [usable: 3],
	DeleteAcls(31): 0 to 3 [usable: 3],
	DescribeConfigs(32): 0 to 4 [usable: 4],
	AlterConfigs(33): 0 to 2 [usable: 2],
	AlterReplicaLogDirs(34): 0 to 2 [usable: 2],
	DescribeLogDirs(35): 0 to 4 [usable: 4],
	SaslAuthenticate(36): 0 to 2 [usable: 2],
	CreatePartitions(37): 0 to 3 [usable: 3],
	CreateDelegationToken(38): 0 to 3 [usable: 3],
	RenewDelegationToken(39): 0 to 2 [usable: 2],
	ExpireDelegationToken(40): 0 to 2 [usable: 2],
	DescribeDelegationToken(41): 0 to 3 [usable: 3],
	DeleteGroups(42): 0 to 2 [usable: 2],
	ElectLeaders(43): 0 to 2 [usable: 2],
	IncrementalAlterConfigs(44): 0 to 1 [usable: 1],
	AlterPartitionReassignments(45): 0 [usable: 0],
	ListPartitionReassignments(46): 0 [usable: 0],
	OffsetDelete(47): 0 [usable: 0],
	DescribeClientQuotas(48): 0 to 1 [usable: 1],
	AlterClientQuotas(49): 0 to 1 [usable: 1],
	DescribeUserScramCredentials(50): 0 [usable: 0],
	AlterUserScramCredentials(51): 0 [usable: 0],
	DescribeQuorum(55): 0 to 1 [usable: 1],
	AlterPartition(56): UNSUPPORTED,
	UpdateFeatures(57): 0 to 1 [usable: 1],
	Envelope(58): UNSUPPORTED,
	DescribeCluster(60): 0 [usable: 0],
	DescribeProducers(61): 0 [usable: 0],
	UnregisterBroker(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
```

**Controller**

```bash
$ kubectl exec -it -n demo kafka-prod-controller-0 -- kafka-metadata-quorum.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties describe --status | grep CurrentObservers
CurrentObservers:       [0,1]
```

From all the above outputs we can see that the replicas of both broker and controller of the topology cluster is `2`. That means we have successfully scaled down the replicas of the Kafka topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kf -n demo kafka-prod
kubectl delete kafkaopsrequest -n demo kfops-hscale-up-topology kfops-hscale-down-topology
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
