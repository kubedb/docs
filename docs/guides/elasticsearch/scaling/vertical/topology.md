---
title: Vertical Scaling Elasticsearch Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: es-vertical-scaling-topology
    name: Topology Cluster
    parent: es-vertical-scalling-elasticsearch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Elasticsearch Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Elasticsearch topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/Elasticsearch/concepts/Elasticsearch.md)
    - [Topology](/docs/guides/Elasticsearch/clustering/topology-cluster/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/Elasticsearch/concepts/elasticsearch-ops-request.md)
    - [Vertical Scaling Overview](/docs/guides/Elasticsearch/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Elasticsearch](/docs/examples/elasticsearch) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Topology Cluster

Here, we are going to deploy a `Elasticsearch` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Elasticsearch Topology Cluster

Now, we are going to deploy a `Elasticsearch` topology cluster database with version `xpack-8.11.1`.

### Deploy Elasticsearch Topology Cluster

In this section, we are going to deploy a Elasticsearch topology cluster. Then, in the next section we will update the resources of the database using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-cluster
  namespace: demo
spec:
  enableSSL: true
  version: xpack-8.11.1
  storageType: Durable
  topology:
    master:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/scaling/Elasticsearch-topology.yaml
Elasticsearch.kubedb.com/es-cluster created
```

Now, wait until `es-cluster` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo -w
NAME         VERSION        STATUS   AGE
es-cluster   xpack-8.11.1   Ready    53m

```

Let's check the Pod containers resources for both `data`,`ingest` and `master` of the Elasticsearch topology cluster. Run the following command to get the resources of the `broker` and `controller` containers of the Elasticsearch topology cluster

```bash
$ kubectl get pod -n demo es-cluster-data-0  -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}
$ kubectl get pod -n demo es-cluster-ingest-0  -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}

$ kubectl get pod -n demo es-cluster-master-0  -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1536Mi"
  }
}

```
This is the default resources of the Elasticsearch topology cluster set by the `KubeDB` operator.

We are now ready to apply the `ElasticsearchOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the topology cluster to meet the desired resources after scaling.

#### Create ElasticsearchOpsRequest

In order to update the resources of the database, we have to create a `ElasticsearchOpsRequest` CR with our desired resources. Below is the YAML of the `ElasticsearchOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: vscale-topology
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: es-cluster
  verticalScaling:
    master:
      resources:
        limits:
          cpu: 750m
          memory: 800Mi
    data:
      resources:
        requests:
          cpu: 760m
          memory: 900Mi
    ingest:
      resources:
        limits:
          cpu: 900m
          memory: 1.2Gi
        requests:
          cpu: 800m
          memory: 1Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `es-cluster` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on Elasticsearch.
- `spec.VerticalScaling.node` specifies the desired resources after scaling.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/scaling/vertical/Elasticsearch-vertical-scaling-topology.yaml
Elasticsearchopsrequest.ops.kubedb.com/vscale-topology created
```

#### Verify Elasticsearch Topology cluster resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Elasticsearch` object and related `PetSets` and `Pods`.

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CR,

```bash
$ kubectl get elasticsearchopsrequest -n demo 
NAME              TYPE              STATUS       AGE
vscale-topology   VerticalScaling   Successful   18m

```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo vscale-topology
Name:         vscale-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-19T11:55:28Z
  Generation:          1
  Resource Version:    71748
  UID:                 be8b4117-90d3-4122-8705-993ce8621635
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-cluster
  Type:    VerticalScaling
  Vertical Scaling:
    Data:
      Resources:
        Requests:
          Cpu:     760m
          Memory:  900Mi
    Ingest:
      Resources:
        Limits:
          Cpu:     900m
          Memory:  1.2Gi
        Requests:
          Cpu:     800m
          Memory:  1Gi
    Master:
      Resources:
        Limits:
          Cpu:     750m
          Memory:  800Mi
Status:
  Conditions:
    Last Transition Time:  2025-11-19T11:55:29Z
    Message:               Elasticsearch ops request is vertically scaling the nodes
    Observed Generation:   1
    Reason:                VerticalScale
    Status:                True
    Type:                  VerticalScale
    Last Transition Time:  2025-11-19T11:55:50Z
    Message:               successfully reconciled the Elasticsearch resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-11-19T11:55:55Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-ingest-0
    Last Transition Time:  2025-11-19T11:55:55Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-ingest-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-ingest-0
    Last Transition Time:  2025-11-19T11:55:55Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-ingest-0
    Last Transition Time:  2025-11-19T11:56:50Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-ingest-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-ingest-0
    Last Transition Time:  2025-11-19T12:03:25Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-19T11:56:35Z
    Message:               re enable shard allocation; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReEnableShardAllocation
    Last Transition Time:  2025-11-19T11:56:40Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-ingest-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-ingest-1
    Last Transition Time:  2025-11-19T11:56:40Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-ingest-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-ingest-1
    Last Transition Time:  2025-11-19T11:56:40Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-1
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-ingest-1
    Last Transition Time:  2025-11-19T11:57:35Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-ingest-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-ingest-1
    Last Transition Time:  2025-11-19T11:57:25Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-ingest-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-ingest-2
    Last Transition Time:  2025-11-19T11:57:25Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-ingest-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-ingest-2
    Last Transition Time:  2025-11-19T11:57:25Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-2
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-ingest-2
    Last Transition Time:  2025-11-19T11:57:25Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-ingest-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-ingest-2
    Last Transition Time:  2025-11-19T11:58:10Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-data-0
    Last Transition Time:  2025-11-19T11:58:10Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-data-0
    Last Transition Time:  2025-11-19T11:58:10Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-data-0
    Last Transition Time:  2025-11-19T11:59:10Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-data-0
    Last Transition Time:  2025-11-19T11:58:35Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-data-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-data-1
    Last Transition Time:  2025-11-19T11:58:35Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-data-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-data-1
    Last Transition Time:  2025-11-19T11:58:35Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-1
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-data-1
    Last Transition Time:  2025-11-19T11:58:35Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-data-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-data-1
    Last Transition Time:  2025-11-19T11:59:00Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-data-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-data-2
    Last Transition Time:  2025-11-19T11:59:00Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-data-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-data-2
    Last Transition Time:  2025-11-19T11:59:00Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-2
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-data-2
    Last Transition Time:  2025-11-19T11:59:00Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-data-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-data-2
    Last Transition Time:  2025-11-19T11:59:25Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-master-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-master-0
    Last Transition Time:  2025-11-19T11:59:25Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-master-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-master-0
    Last Transition Time:  2025-11-19T11:59:25Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-0
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-master-0
    Last Transition Time:  2025-11-19T12:00:25Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-master-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-master-0
    Last Transition Time:  2025-11-19T12:00:15Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-master-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-master-1
    Last Transition Time:  2025-11-19T12:00:15Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-master-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-master-1
    Last Transition Time:  2025-11-19T12:00:15Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-1
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-master-1
    Last Transition Time:  2025-11-19T12:00:15Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-master-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-master-1
    Last Transition Time:  2025-11-19T12:01:05Z
    Message:               pod exists; ConditionStatus:True; PodName:es-cluster-master-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-cluster-master-2
    Last Transition Time:  2025-11-19T12:01:05Z
    Message:               create es client; ConditionStatus:True; PodName:es-cluster-master-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-cluster-master-2
    Last Transition Time:  2025-11-19T12:01:05Z
    Message:               disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-2
    Observed Generation:   1
    Status:                True
    Type:                  DisableShardAllocation--es-cluster-master-2
    Last Transition Time:  2025-11-19T12:01:05Z
    Message:               evict pod; ConditionStatus:True; PodName:es-cluster-master-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-cluster-master-2
    Last Transition Time:  2025-11-19T12:02:10Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-11-19T12:02:15Z
    Message:               successfully updated Elasticsearch CR
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-11-19T12:02:15Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                       Age   From                         Message
  ----     ------                                                                       ----  ----                         -------
  Normal   PauseDatabase                                                                19m   KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-cluster
  Normal   UpdatePetSets                                                                19m   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0                19m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0          19m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0  19m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-0                 19m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  create es client; ConditionStatus:False                                      19m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Normal   UpdatePetSets                                                                19m   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Normal   UpdatePetSets                                                                19m   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Normal   UpdatePetSets                                                                18m   KubeDB Ops-manager Operator  successfully reconciled the Elasticsearch resources
  Warning  create es client; ConditionStatus:True                                       18m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             18m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-1                18m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-1          18m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-1  18m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-1                 18m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0                18m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0          18m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0  18m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  evict pod; ConditionStatus:False; PodName:es-cluster-ingest-0                18m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:False; PodName:es-cluster-ingest-0
  Warning  create es client; ConditionStatus:False                                      18m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0                18m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0          18m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0  18m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0                18m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0          18m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0  18m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-0                 18m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-0
  Warning  create es client; ConditionStatus:False                                      18m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             17m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-2                17m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-2          17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-2  17m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-2                 17m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  create es client; ConditionStatus:True                                       17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             17m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  create es client; ConditionStatus:False                                      17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-1                17m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-1          17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-1  17m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  evict pod; ConditionStatus:False; PodName:es-cluster-ingest-1                17m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:False; PodName:es-cluster-ingest-1
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-1                17m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-1          17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-1  17m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-1                 17m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-1
  Warning  create es client; ConditionStatus:False                                      17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             17m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-data-0                  17m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-data-0            17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-0    17m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-data-0                   17m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  create es client; ConditionStatus:False                                      17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       17m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             17m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-2                16m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-ingest-2          16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-2  16m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-2                 16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-ingest-2
  Warning  create es client; ConditionStatus:False                                      16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             16m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-data-1                  16m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-data-1            16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-1    16m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-data-1                   16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  create es client; ConditionStatus:False                                      16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             16m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-data-2                  16m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-data-2            16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  create es client; ConditionStatus:True                                       16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-2    16m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  re enable shard allocation; ConditionStatus:True                             16m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-data-2                   16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  create es client; ConditionStatus:False                                      16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-data-0                  16m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-data-0            16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-0    16m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  evict pod; ConditionStatus:False; PodName:es-cluster-data-0                  16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:False; PodName:es-cluster-data-0
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-data-0                  16m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-data-0            16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-0    16m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-data-0                   16m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-data-0
  Warning  create es client; ConditionStatus:False                                      16m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             15m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-master-0                15m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-master-0          15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-0  15m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-master-0                 15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  create es client; ConditionStatus:False                                      15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             15m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-data-1                  15m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-data-1            15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-1    15m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-data-1                   15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-data-1
  Warning  create es client; ConditionStatus:False                                      15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             15m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-data-2                  15m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-data-2            15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-2    15m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-data-2                   15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-data-2
  Warning  create es client; ConditionStatus:False                                      15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             15m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-master-1                15m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-master-1          15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-1  15m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-master-1                 15m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  create es client; ConditionStatus:True                                       15m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             15m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  create es client; ConditionStatus:False                                      14m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-master-0                14m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-master-0          14m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-0  14m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  evict pod; ConditionStatus:False; PodName:es-cluster-master-0                14m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:False; PodName:es-cluster-master-0
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-master-0                14m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-master-0          14m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-0  14m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-master-0                 14m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-master-0
  Warning  create es client; ConditionStatus:False                                      14m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       14m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             14m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-master-2                14m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-master-2          14m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-2  14m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-master-2                 14m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  create es client; ConditionStatus:False                                      14m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       13m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             13m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-master-1                13m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-master-1          13m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-1  13m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-master-1                 13m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-master-1
  Warning  create es client; ConditionStatus:False                                      13m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       13m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             13m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Normal   RestartNodes                                                                 13m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   UpdateDatabase                                                               13m   KubeDB Ops-manager Operator  successfully updated Elasticsearch CR
  Normal   ResumeDatabase                                                               13m   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-cluster
  Normal   ResumeDatabase                                                               13m   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-cluster
  Normal   Successful                                                                   13m   KubeDB Ops-manager Operator  Successfully Updated Database
  Warning  create es client; ConditionStatus:True                                       12m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             12m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-cluster-master-2                12m   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  create es client; ConditionStatus:True; PodName:es-cluster-master-2          12m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-2  12m   KubeDB Ops-manager Operator  disable shard allocation; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  evict pod; ConditionStatus:True; PodName:es-cluster-master-2                 12m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-cluster-master-2
  Warning  create es client; ConditionStatus:False                                      12m   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                                       11m   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  re enable shard allocation; ConditionStatus:True                             11m   KubeDB Ops-manager Operator  re enable shard allocation; ConditionStatus:True
  Normal   RestartNodes                                                                 11m   KubeDB Ops-manager Operator  Successfully restarted all nodes

```
Now, we are going to verify from one of the Pod yaml whether the resources of the topology cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo es-cluster-ingest-0  -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "900m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "800m",
    "memory": "1Gi"
  }
}
$ kubectl get pod -n demo es-cluster-data-0  -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "900Mi"
  },
  "requests": {
    "cpu": "760m",
    "memory": "900Mi"
  }
}
$ kubectl get pod -n demo es-cluster-master-0  -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "750m",
    "memory": "800Mi"
  },
  "requests": {
    "cpu": "750m",
    "memory": "800Mi"
  }
}

```

The above output verifies that we have successfully scaled up the resources of the Elasticsearch topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete es -n demo es-cluster
kubectl delete Elasticsearchopsrequest -n demo vscale-topology
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/Elasticsearch/concepts/Elasticsearch.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/Elasticsearch/clustering/_index.md).
- Monitor your Elasticsearch database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/Elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
