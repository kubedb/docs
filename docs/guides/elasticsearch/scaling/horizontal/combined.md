---
title: Horizontal Scaling Combined Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-horizontal-scaling-combined
    name: Combined Cluster
    parent: es-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Elasticsearch Combined Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Elasticsearch combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [Combined](/docs/guides/elasticsearch/clustering/combined-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
    - [Horizontal Scaling Overview](/docs/guides/elasticsearch/scaling/horizontal/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Combined Cluster

Here, we are going to deploy a  `Elasticsearch` combined cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Elasticsearch Combined cluster

Now, we are going to deploy a `Elasticsearch` combined cluster with version `3.9.0`.

### Deploy Elasticsearch combined cluster

In this section, we are going to deploy a Elasticsearch combined cluster. Then, in the next section we will scale the cluster using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: Elasticsearch-dev
  namespace: demo
spec:
  replicas: 2
  version: 3.9.0
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

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/scaling/Elasticsearch-combined.yaml
Elasticsearch.kubedb.com/Elasticsearch-dev created
```

Now, wait until `Elasticsearch-dev` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo -w
NAME         TYPE            VERSION   STATUS         AGE
Elasticsearch-dev    kubedb.com/v1   3.9.0     Provisioning   0s
Elasticsearch-dev    kubedb.com/v1   3.9.0     Provisioning   24s
.
.
Elasticsearch-dev    kubedb.com/v1   3.9.0     Ready          92s
```

Let's check the number of replicas has from Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get Elasticsearch -n demo Elasticsearch-dev -o json | jq '.spec.replicas'
2

$ kubectl get petset -n demo Elasticsearch-dev -o json | jq '.spec.replicas'
2
```

We can see from both command that the cluster has 2 replicas.

Also, we can verify the replicas of the combined from an internal Elasticsearch command by exec into a replica.

Now let's exec to a instance and run a Elasticsearch internal command to check the number of replicas,

```bash
$ kubectl exec -it -n demo Elasticsearch-dev-0 -- Elasticsearch-broker-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
Elasticsearch-dev-0.Elasticsearch-dev-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
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
Elasticsearch-dev-1.Elasticsearch-dev-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
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

We can see from the above output that the Elasticsearch has 2 nodes.

We are now ready to apply the `ElasticsearchOpsRequest` CR to scale this cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the combined cluster to meet the desired number of replicas after scaling.

#### Create ElasticsearchOpsRequest

In order to scale up the replicas of the combined cluster, we have to create a `ElasticsearchOpsRequest` CR with our desired replicas. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-hscale-up-combined
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: Elasticsearch-dev
  horizontalScaling:
    node: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `Elasticsearch-dev` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Elasticsearch.
- `spec.horizontalScaling.node` specifies the desired replicas after scaling.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/scaling/horizontal-scaling/Elasticsearch-hscale-up-combined.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-hscale-up-combined created
```

#### Verify Combined cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ watch kubectl get Elasticsearchopsrequest -n demo
NAME                        TYPE                STATUS       AGE
esops-hscale-up-combined    HorizontalScaling   Successful   106s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe Elasticsearchopsrequests -n demo esops-hscale-up-combined
Name:         esops-hscale-up-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T10:19:56Z
  Generation:          1
  Resource Version:    353093
  UID:                 f91de2da-82c4-4175-aab4-de0f3e1ce498
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  Elasticsearch-dev
  Horizontal Scaling:
    Node:  3
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-08-02T10:19:57Z
    Message:               Elasticsearch ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-08-02T10:20:05Z
    Message:               get pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--Elasticsearch-dev-0
    Last Transition Time:  2024-08-02T10:20:05Z
    Message:               evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--Elasticsearch-dev-0
    Last Transition Time:  2024-08-02T10:20:15Z
    Message:               check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--Elasticsearch-dev-0
    Last Transition Time:  2024-08-02T10:20:20Z
    Message:               get pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--Elasticsearch-dev-1
    Last Transition Time:  2024-08-02T10:20:20Z
    Message:               evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--Elasticsearch-dev-1
    Last Transition Time:  2024-08-02T10:21:00Z
    Message:               check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--Elasticsearch-dev-1
    Last Transition Time:  2024-08-02T10:21:05Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T10:22:15Z
    Message:               Successfully Scaled Up Server Node
    Observed Generation:   1
    Reason:                ScaleUpCombined
    Status:                True
    Type:                  ScaleUpCombined
    Last Transition Time:  2024-08-02T10:21:10Z
    Message:               patch pet setElasticsearch-dev; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSetElasticsearch-dev
    Last Transition Time:  2024-08-02T10:22:10Z
    Message:               node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  NodeInCluster
    Last Transition Time:  2024-08-02T10:22:15Z
    Message:               Successfully completed horizontally scale Elasticsearch cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age    From                         Message
  ----     ------                                                         ----   ----                         -------
  Normal   Starting                                                       4m34s  KubeDB Ops-manager Operator  Start processing for ElasticsearchOpsRequest: demo/esops-hscale-up-combined
  Normal   Starting                                                       4m34s  KubeDB Ops-manager Operator  Pausing Elasticsearch databse: demo/Elasticsearch-dev
  Normal   Successful                                                     4m34s  KubeDB Ops-manager Operator  Successfully paused Elasticsearch database: demo/Elasticsearch-dev for ElasticsearchOpsRequest: esops-hscale-up-combined
  Warning  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-0             4m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
  Warning  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-0           4m26s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
  Warning  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-0  4m21s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-0
  Warning  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-0   4m16s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-0
  Warning  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-1             4m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
  Warning  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-1           4m11s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
  Warning  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-1  4m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-1
  Warning  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-1   3m31s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-1
  Normal   RestartNodes                                                   3m26s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Warning  patch pet setElasticsearch-dev; ConditionStatus:True                   3m21s  KubeDB Ops-manager Operator  patch pet setElasticsearch-dev; ConditionStatus:True
  Warning  node in cluster; ConditionStatus:False                         2m46s  KubeDB Ops-manager Operator  node in cluster; ConditionStatus:False
  Warning  node in cluster; ConditionStatus:True                          2m21s  KubeDB Ops-manager Operator  node in cluster; ConditionStatus:True
  Normal   ScaleUpCombined                                                2m16s  KubeDB Ops-manager Operator  Successfully Scaled Up Server Node
  Normal   Starting                                                       2m16s  KubeDB Ops-manager Operator  Resuming Elasticsearch database: demo/Elasticsearch-dev
  Normal   Successful                                                     2m16s  KubeDB Ops-manager Operator  Successfully resumed Elasticsearch database: demo/Elasticsearch-dev for ElasticsearchOpsRequest: esops-hscale-up-combined
```

Now, we are going to verify the number of replicas this cluster has from the Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get Elasticsearch -n demo Elasticsearch-dev -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo Elasticsearch-dev -o json | jq '.spec.replicas'
3
```

Now let's connect to a Elasticsearch instance and run a Elasticsearch internal command to check the number of replicas,
```bash
$ kubectl exec -it -n demo Elasticsearch-dev-0 -- Elasticsearch-broker-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
Elasticsearch-dev-0.Elasticsearch-dev-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
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
Elasticsearch-dev-1.Elasticsearch-dev-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
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
Elasticsearch-dev-2.Elasticsearch-dev-pods.demo.svc.cluster.local:9092 (id: 2 rack: null) -> (
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

From all the above outputs we can see that the brokers of the combined Elasticsearch is `3`. That means we have successfully scaled up the replicas of the Elasticsearch combined cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the Elasticsearch combined cluster to meet the desired number of replicas after scaling.

#### Create ElasticsearchOpsRequest

In order to scale down the replicas of the Elasticsearch combined cluster, we have to create a `ElasticsearchOpsRequest` CR with our desired replicas. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-hscale-down-combined
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: Elasticsearch-dev
  horizontalScaling:
    node: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `Elasticsearch-dev` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Elasticsearch.
- `spec.horizontalScaling.node` specifies the desired replicas after scaling.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/scaling/horizontal-scaling/Elasticsearch-hscale-down-combined.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-hscale-down-combined created
```

#### Verify Combined cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`. Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ watch kubectl get Elasticsearchopsrequest -n demo
NAME                          TYPE                STATUS       AGE
esops-hscale-down-combined    HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe Elasticsearchopsrequests -n demo esops-hscale-down-combined
Name:         esops-hscale-down-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T10:46:39Z
  Generation:          1
  Resource Version:    354924
  UID:                 f1a0b85d-1a86-463c-a3e4-72947badd108
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  Elasticsearch-dev
  Horizontal Scaling:
    Node:  2
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-08-02T10:46:39Z
    Message:               Elasticsearch ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-08-02T10:47:07Z
    Message:               Successfully Scaled Down Server Node
    Observed Generation:   1
    Reason:                ScaleDownCombined
    Status:                True
    Type:                  ScaleDownCombined
    Last Transition Time:  2024-08-02T10:46:57Z
    Message:               reassign partitions; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReassignPartitions
    Last Transition Time:  2024-08-02T10:46:57Z
    Message:               is pet set patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetPatched
    Last Transition Time:  2024-08-02T10:46:57Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-08-02T10:46:58Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2024-08-02T10:47:02Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-08-02T10:47:13Z
    Message:               successfully reconciled the Elasticsearch with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-02T10:47:18Z
    Message:               get pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--Elasticsearch-dev-0
    Last Transition Time:  2024-08-02T10:47:18Z
    Message:               evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--Elasticsearch-dev-0
    Last Transition Time:  2024-08-02T10:47:28Z
    Message:               check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--Elasticsearch-dev-0
    Last Transition Time:  2024-08-02T10:47:33Z
    Message:               get pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--Elasticsearch-dev-1
    Last Transition Time:  2024-08-02T10:47:33Z
    Message:               evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--Elasticsearch-dev-1
    Last Transition Time:  2024-08-02T10:48:53Z
    Message:               check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--Elasticsearch-dev-1
    Last Transition Time:  2024-08-02T10:48:58Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T10:48:58Z
    Message:               Successfully completed horizontally scale Elasticsearch cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age    From                         Message
  ----     ------                                                         ----   ----                         -------
  Normal   Starting                                                       2m39s  KubeDB Ops-manager Operator  Start processing for ElasticsearchOpsRequest: demo/esops-hscale-down-combined
  Normal   Starting                                                       2m39s  KubeDB Ops-manager Operator  Pausing Elasticsearch databse: demo/Elasticsearch-dev
  Normal   Successful                                                     2m39s  KubeDB Ops-manager Operator  Successfully paused Elasticsearch database: demo/Elasticsearch-dev for ElasticsearchOpsRequest: esops-hscale-down-combined
  Warning  reassign partitions; ConditionStatus:True                      2m21s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True                       2m21s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                                  2m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                               2m20s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                                 2m20s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                                  2m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                               2m16s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                                  2m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   ScaleDownCombined                                              2m11s  KubeDB Ops-manager Operator  Successfully Scaled Down Server Node
  Normal   UpdatePetSets                                                  2m5s   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch with modified node
  Warning  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-0             2m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
  Warning  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-0           2m     KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-0
  Warning  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-0  115s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-0
  Warning  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-0   110s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-0
  Warning  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-1             105s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
  Warning  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-1           105s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:Elasticsearch-dev-1
  Warning  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-1  100s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:Elasticsearch-dev-1
  Warning  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-1   25s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:Elasticsearch-dev-1
  Normal   RestartNodes                                                   20s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                       20s    KubeDB Ops-manager Operator  Resuming Elasticsearch database: demo/Elasticsearch-dev
  Normal   Successful                                                     20s    KubeDB Ops-manager Operator  Successfully resumed Elasticsearch database: demo/Elasticsearch-dev for ElasticsearchOpsRequest: esops-hscale-down-combined
```

Now, we are going to verify the number of replicas this cluster has from the Elasticsearch object, number of pods the petset have,

```bash
$ kubectl get Elasticsearch -n demo Elasticsearch-dev -o json | jq '.spec.replicas' 
2

$ kubectl get petset -n demo Elasticsearch-dev -o json | jq '.spec.replicas'
2
```

Now let's connect to a Elasticsearch instance and run a Elasticsearch internal command to check the number of replicas,

```bash
$ kubectl exec -it -n demo Elasticsearch-dev-0 -- Elasticsearch-broker-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
Elasticsearch-dev-0.Elasticsearch-dev-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
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
Elasticsearch-dev-1.Elasticsearch-dev-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
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

From all the above outputs we can see that the replicas of the combined cluster is `2`. That means we have successfully scaled down the replicas of the Elasticsearch combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete es -n demo Elasticsearch-dev
kubectl delete Elasticsearchopsrequest -n demo esops-hscale-up-combined esops-hscale-down-combined
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/Elasticsearch.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/_index.md).
- Monitor your Elasticsearch with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Elasticsearch with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
