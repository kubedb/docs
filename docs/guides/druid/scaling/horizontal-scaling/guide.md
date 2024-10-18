---
title: Horizontal Scaling Druid Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-druid-scaling-horizontal-scaling-guide
    name: Guide
    parent: guides-druid-scaling-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Druid Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Druid topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [Topology](/docs/guides/druid/clustering/topology-cluster/index.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)
    - [Horizontal Scaling Overview](/docs/guides/druid/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/druid](/docs/examples/druid) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Druid Cluster

Here, we are going to deploy a `Druid` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Druid Topology cluster

Now, we are going to deploy a `Druid` topology cluster with version `28.0.1`.

### Create External Dependency (Deep Storage)

Before proceeding further, we need to prepare deep storage, which is one of the external dependency of Druid and used for storing the segments. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/scaling/horizontal-scaling/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

### Deploy Druid topology cluster

In this section, we are going to deploy a Druid topology cluster. Then, in the next section we will scale the cluster using `DruidOpsRequest` CRD. Below is the YAML of the `Druid` CR that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-quickstart
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

Let's create the `Druid` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/scaling/druid-topology.yaml
druid.kubedb.com/druid-cluster created
```

Now, wait until `druid-cluster` has status `Ready`. i.e,

```bash
$ kubectl get dr -n demo -w
NAME             TYPE                  VERSION    STATUS         AGE
druid-cluster    kubedb.com/v1aplha2   28.0.1     Provisioning   0s
druid-cluster    kubedb.com/v1aplha2   28.0.1     Provisioning   24s
.
.
druid-cluster    kubedb.com/v1aplha2   28.0.1     Ready          92s
```

Let's check the number of replicas has from druid object, number of pods the petset have,

**Coordinators Replicas**

```bash
$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.coordinators.replicas'
1

$ kubectl get petset -n demo druid-cluster-coordinators -o json | jq '.spec.replicas'
1
```

**Historicals Replicas**

```bash
$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.historicals.replicas'
1

$ kubectl get petset -n demo druid-cluster-historicals -o json | jq '.spec.replicas'
1
```

We can see from commands that the cluster has 1 replicas for both coordinators and historicals.

Also, we can verify the replicas of the topology from an internal druid command by exec into a replica.

Now lets exec to a coordinators instance and run a druid internal command to check the number of replicas for coordinators and historicals.,

**Coordinators**

```bash
$ kubectl exec -it -n demo druid-cluster-coordinators-0 -- druid-coordinators-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
druid-cluster-coordinators-0.druid-cluster-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
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
	UnregisterCoordinators(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
druid-cluster-coordinators-1.druid-cluster-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
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
	UnregisterCoordinators(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
```

**Historicals**

```bash
$ kubectl exec -it -n demo druid-cluster-historicals-0 -- druid-metadata-quorum.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties describe --status | grep CurrentObservers
CurrentObservers:       [0,1]
```

We can see from the above output that the druid has 1 node for coordinators and 1 node for historicals.

We are now ready to apply the `DruidOpsRequest` CR to scale this cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the topology cluster to meet the desired number of replicas after scaling.

#### Create DruidOpsRequest

In order to scale up the replicas of the topology cluster, we have to create a `DruidOpsRequest` CR with our desired replicas. Below is the YAML of the `DruidOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druid-hscale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: druid-cluster
  horizontalScaling:
    topology: 
      coordinators: 2
      historicals: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `druid-cluster` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on druid.
- `spec.horizontalScaling.topology.coordinators` specifies the desired replicas after scaling for coordinators.
- `spec.horizontalScaling.topology.historicals` specifies the desired replicas after scaling for historicals.

> **Note:** Similarly you can scale other druid nodes by specifying the following fields:
  > - For `overlords` use `spec.horizontalScaling.topology.overlords`.
  > - For `brokers` use `spec.horizontalScaling.topology.brokers`.
  > - For `middleManagers` use `spec.horizontalScaling.topology.middleManagers`.
  > - For `routers` use `spec.horizontalScaling.topology.routers`.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/scaling/horizontal-scaling/yamls/druid-hscale-up.yaml
druidopsrequest.ops.kubedb.com/druid-hscale-up created
```

> **Note:** If you want to scale down only coordinators or historicals, you can specify the desired replicas for only coordinators or historicals in the `DruidOpsRequest` CR. You can specify one at a time. If you want to scale coordinators only, no node will need restart to apply the changes. But if you want to scale historicals, all nodes will need restart to apply the changes.

#### Verify Topology cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Druid` object and related `PetSets` and `Pods`.

Let's wait for `DruidOpsRequest` to be `Successful`. Run the following command to watch `DruidOpsRequest` CR,

```bash
$ watch kubectl get druidopsrequest -n demo
NAME                        TYPE                STATUS       AGE
druid-hscale-up    HorizontalScaling   Successful   106s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe druidopsrequests -n demo druid-hscale-up 
Name:         druid-hscale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T11:02:51Z
  Generation:          1
  Resource Version:    356503
  UID:                 44e0db0c-2094-4c13-a3be-9ca680888545
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Horizontal Scaling:
    Topology:
      Coordinators:      3
      Historicals:  3
  Type:            HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-08-02T11:02:51Z
    Message:               Druid ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-08-02T11:02:59Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-08-02T11:03:00Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-08-02T11:03:09Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-08-02T11:03:14Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-1
    Last Transition Time:  2024-08-02T11:03:14Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-1
    Last Transition Time:  2024-08-02T11:03:24Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-1
    Last Transition Time:  2024-08-02T11:03:29Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-08-02T11:03:30Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-08-02T11:03:59Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-08-02T11:04:04Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-1
    Last Transition Time:  2024-08-02T11:04:05Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-1
    Last Transition Time:  2024-08-02T11:04:19Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-1
    Last Transition Time:  2024-08-02T11:04:24Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T11:04:59Z
    Message:               Successfully Scaled Up Coordinators
    Observed Generation:   1
    Reason:                ScaleUpCoordinators
    Status:                True
    Type:                  ScaleUpCoordinators
    Last Transition Time:  2024-08-02T11:04:30Z
    Message:               patch pet setdruid-cluster-coordinators; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSetdruid-cluster-coordinators
    Last Transition Time:  2024-08-02T11:04:55Z
    Message:               node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  NodeInCluster
    Last Transition Time:  2024-08-02T11:05:15Z
    Message:               Successfully Scaled Up Historicals
    Observed Generation:   1
    Reason:                ScaleUpHistoricals
    Status:                True
    Type:                  ScaleUpHistoricals
    Last Transition Time:  2024-08-02T11:05:05Z
    Message:               patch pet setdruid-cluster-historicals; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSetdruid-cluster-historicals
    Last Transition Time:  2024-08-02T11:05:15Z
    Message:               Successfully completed horizontally scale druid cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   4m19s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/druid-hscale-up
  Normal   Starting                                                                   4m19s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                 4m19s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: druid-hscale-up
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             4m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0           4m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-0  4m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0   4m1s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-1             3m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-1           3m56s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-1  3m50s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-1
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-1   3m46s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-1
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0                 3m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0               3m41s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-0      3m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0       3m11s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1                 3m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1               3m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-1      3m1s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-1
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-1       2m51s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-1
  Normal   RestartNodes                                                               2m46s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Warning  patch pet setdruid-cluster-coordinators; ConditionStatus:True                       2m40s  KubeDB Ops-manager Operator  patch pet setdruid-cluster-coordinators; ConditionStatus:True
  Warning  node in cluster; ConditionStatus:False                                     2m36s  KubeDB Ops-manager Operator  node in cluster; ConditionStatus:False
  Warning  node in cluster; ConditionStatus:True                                      2m15s  KubeDB Ops-manager Operator  node in cluster; ConditionStatus:True
  Normal   ScaleUpCoordinators                                                              2m10s  KubeDB Ops-manager Operator  Successfully Scaled Up Coordinators
  Warning  patch pet setdruid-cluster-historicals; ConditionStatus:True                   2m5s   KubeDB Ops-manager Operator  patch pet setdruid-cluster-historicals; ConditionStatus:True
  Warning  node in cluster; ConditionStatus:True                                      2m     KubeDB Ops-manager Operator  node in cluster; ConditionStatus:True
  Normal   ScaleUpHistoricals                                                          115s   KubeDB Ops-manager Operator  Successfully Scaled Up Historicals
  Normal   Starting                                                                   115s   KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                 115s   KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: druid-hscale-up
```

Now, we are going to verify the number of replicas this cluster has from the Druid object, number of pods the petset have,

**Coordinators Replicas**

```bash
$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.coordinators.replicas'
3

$ kubectl get petset -n demo druid-cluster-coordinators -o json | jq '.spec.replicas'
3
```

Now let's connect to a druid instance and run a druid internal command to check the number of replicas of topology cluster for both coordinators and historicals.,

**Coordinators**

```bash
$ kubectl exec -it -n demo druid-cluster-coordinators-0 -- druid-coordinators-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
druid-cluster-coordinators-0.druid-cluster-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
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
	UnregisterCoordinators(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
druid-cluster-coordinators-1.druid-cluster-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
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
	UnregisterCoordinators(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
druid-cluster-coordinators-2.druid-cluster-pods.demo.svc.cluster.local:9092 (id: 2 rack: null) -> (
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
	UnregisterCoordinators(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
```

**Historicals**

```bash
$ kubectl exec -it -n demo druid-cluster-coordinators-0 -- druid-metadata-quorum.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties describe --status | grep CurrentObservers
CurrentObservers:       [0,1,2]
```

From all the above outputs we can see that the both coordinatorss and historicals of the topology druid is `3`. That means we have successfully scaled up the replicas of the Druid topology cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the druid topology cluster to meet the desired number of replicas after scaling.

#### Create DruidOpsRequest

In order to scale down the replicas of the druid topology cluster, we have to create a `DruidOpsRequest` CR with our desired replicas. Below is the YAML of the `DruidOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druid-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: druid-cluster
  horizontalScaling:
    topology:
      coordinators: 1
      historicals: 1
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `druid-cluster` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on druid.
- `spec.horizontalScaling.topology.coordinators` specifies the desired replicas after scaling for the coordinators nodes.
- `spec.horizontalScaling.topology.historicals` specifies the desired replicas after scaling for the historicals nodes.

> **Note:** Similarly you can scale other druid nodes by specifying the following fields:
> - For `overlords` use `spec.horizontalScaling.topology.overlords`.
> - For `brokers` use `spec.horizontalScaling.topology.brokers`.
> - For `middleManagers` use `spec.horizontalScaling.topology.middleManagers`.
> - For `routers` use `spec.horizontalScaling.topology.routers`.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/scaling/horizontal-scaling/druid-hscale-down-topology.yaml
druidopsrequest.ops.kubedb.com/druid-hscale-down created
```

#### Verify Topology cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Druid` object and related `PetSets` and `Pods`.

Let's wait for `DruidOpsRequest` to be `Successful`. Run the following command to watch `DruidOpsRequest` CR,

```bash
$ watch kubectl get druidopsrequest -n demo
NAME                          TYPE                STATUS       AGE
druid-hscale-down    HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe druidopsrequests -n demo druid-hscale-down
Name:         druid-hscale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T11:14:18Z
  Generation:          1
  Resource Version:    357545
  UID:                 b786d791-6ba8-4f1c-ade8-9443e049cede
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  druid-cluster
  Horizontal Scaling:
    Topology:
      Coordinators:      2
      Historicals:  2
  Type:            HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-08-02T11:14:18Z
    Message:               Druid ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-08-02T11:14:46Z
    Message:               Successfully Scaled Down Coordinators
    Observed Generation:   1
    Reason:                ScaleDownCoordinators
    Status:                True
    Type:                  ScaleDownCoordinators
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
    Message:               Successfully Scaled Down Historicals
    Observed Generation:   1
    Reason:                ScaleDownHistoricals
    Status:                True
    Type:                  ScaleDownHistoricals
    Last Transition Time:  2024-08-02T11:15:38Z
    Message:               successfully reconciled the Druid with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-02T11:15:43Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-0
    Last Transition Time:  2024-08-02T11:15:44Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-0
    Last Transition Time:  2024-08-02T11:15:53Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-0
    Last Transition Time:  2024-08-02T11:15:58Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-historicals-1
    Last Transition Time:  2024-08-02T11:15:58Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-historicals-1
    Last Transition Time:  2024-08-02T11:16:08Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-historicals-1
    Last Transition Time:  2024-08-02T11:16:13Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-08-02T11:16:13Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-0
    Last Transition Time:  2024-08-02T11:16:58Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-0
    Last Transition Time:  2024-08-02T11:17:03Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-coordinators-1
    Last Transition Time:  2024-08-02T11:17:03Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-coordinators-1
    Last Transition Time:  2024-08-02T11:17:13Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-coordinators-1
    Last Transition Time:  2024-08-02T11:17:18Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T11:17:19Z
    Message:               Successfully completed horizontally scale druid cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   8m35s  KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/druid-hscale-down
  Normal   Starting                                                                   8m35s  KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                 8m35s  KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: druid-hscale-down
  Warning  reassign partitions; ConditionStatus:True                                  8m17s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True                                   8m17s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                                              8m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           8m16s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                                             8m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                                              8m12s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           8m12s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                                              8m12s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   ScaleDownCoordinators                                                            8m7s   KubeDB Ops-manager Operator  Successfully Scaled Down Coordinators
  Warning  reassign partitions; ConditionStatus:True                                  7m31s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True                                   7m31s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                                              7m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           7m31s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False                                             7m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                                              7m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True                                           7m27s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                                              7m27s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   ScaleDownHistoricals                                                        7m22s  KubeDB Ops-manager Operator  Successfully Scaled Down Historicals
  Normal   UpdatePetSets                                                              7m15s  KubeDB Ops-manager Operator  successfully reconciled the Druid with modified node
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0             7m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0           7m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-0  7m5s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0   7m     KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-1             6m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-1           6m55s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-historicals-1
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-1  6m50s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-historicals-1
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-1   6m45s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-historicals-1
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0                 6m40s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0               6m40s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-0      6m35s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0       5m55s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1                 5m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1               5m50s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-coordinators-1
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-1      5m45s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-coordinators-1
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-1       5m40s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-coordinators-1
  Normal   RestartNodes                                                               5m35s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   5m35s  KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                 5m34s  KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: druid-hscale-down
```

Now, we are going to verify the number of replicas this cluster has from the Druid object, number of pods the petset have,

**Coordinators Replicas**

```bash
$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.coordinators.replicas' 
2

$ kubectl get petset -n demo druid-cluster-coordinators -o json | jq '.spec.replicas'
2
```

**Historicals Replicas**

```bash
$ kubectl get druid -n demo druid-cluster -o json | jq '.spec.topology.historicals.replicas'
2

$ kubectl get petset -n demo druid-cluster-historicals -o json | jq '.spec.replicas'
2
```

Now let's connect to a druid instance and run a druid internal command to check the number of replicas for both coordinators and historicals nodes,

**Coordinators**

```bash
$ kubectl exec -it -n demo druid-cluster-coordinators-0 -- druid-coordinators-api-versions.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties
druid-cluster-coordinators-0.druid-cluster-pods.demo.svc.cluster.local:9092 (id: 0 rack: null) -> (
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
	UnregisterCoordinators(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
druid-cluster-coordinators-1.druid-cluster-pods.demo.svc.cluster.local:9092 (id: 1 rack: null) -> (
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
	UnregisterCoordinators(64): 0 [usable: 0],
	DescribeTransactions(65): 0 [usable: 0],
	ListTransactions(66): 0 [usable: 0],
	AllocateProducerIds(67): UNSUPPORTED,
	ConsumerGroupHeartbeat(68): UNSUPPORTED
)
```

**Historicals**

```bash
$ kubectl exec -it -n demo druid-cluster-historicals-0 -- druid-metadata-quorum.sh --bootstrap-server localhost:9092 --command-config config/clientauth.properties describe --status | grep CurrentObservers
CurrentObservers:       [0,1]
```

From all the above outputs we can see that the replicas of both coordinators and historicals of the topology cluster is `2`. That means we have successfully scaled down the replicas of the Druid topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete dr -n demo druid-cluster
kubectl delete druidopsrequest -n demo druid-hscale-up druid-hscale-down
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Druid with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
