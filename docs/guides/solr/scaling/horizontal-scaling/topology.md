---
title: Solr Topology Horizontal Scaling
menu:
  docs_{{ .version }}:
    identifier: sl-scaling-horizontal-topology
    name: Topology Cluster
    parent: sl-scaling-horizontal
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Solr Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Solr topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [Topology](/docs/guides/solr/clustering/topology_cluster.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Horizontal Scaling Overview](/docs/guides/solr/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/solr](/docs/examples/solr) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Topology Cluster

Here, we are going to deploy a `Solr` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Solr Topology cluster

Now, we are going to deploy a `Solr` topology cluster with version `9.4.1`.

### Deploy Solr topology cluster

In this section, we are going to deploy a Solr topology cluster. Then, in the next section we will scale the cluster using `SolrOpsRequest` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.4.1
  zookeeperRef:
    name: zoo
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi

```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/scaling/horizontal/topology/solr.yaml
solr.kubedb.com/solr-cluster created
```

Now, wait until `solr-cluster` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.4.1     Ready    90m
```

Let's check the number of replicas has from Solr object, number of pods the petset have,

**Data Replicas**

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.data.replicas'
1
$ kubectl get petset -n demo solr-cluster-data -o json | jq '.spec.replicas'
1
```

**Overseer Replicas**

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.overseer.replicas'
1
$ kubectl get petset -n demo solr-cluster-overseer -o json | jq '.spec.replicas'
1

```

**Coordinator Replicas**

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.coordinator.replicas'
1
$ kubectl get petset -n demo solr-cluster-coordinator -o json | jq '.spec.replicas'
1

```

We can see from commands that the cluster has 3 replicas for data, overseer, coordinator.



## Scale Up Replicas

Here, we are going to scale up the replicas of the topology cluster to meet the desired number of replicas after scaling.

#### Create SolrOpsRequest

In order to scale up the replicas of the topology cluster, we have to create a `SolrOpsRequest` CR with our desired replicas. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-hscale-up-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: HorizontalScaling
  horizontalScaling:
    data: 2
    overseer: 2
    coordinator: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `Solr-prod` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Solr.
- `spec.horizontalScaling.topology.data` specifies the desired replicas after scaling for data.
- `spec.horizontalScaling.topology.overseer` specifies the desired replicas after scaling for overseer.
- `spec.horizontalScaling.topology.coordinator` specifies the desired replicas after scaling for coordinator.


Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/scaling/horizontal/topology/slops-hscale-up-topology.yaml
solropsrequest.ops.kubedb.com/slops-hscale-up-topology created
```

> **Note:** If you want to scale down only broker or controller, you can specify the desired replicas for only broker or controller in the `SolrOpsRequest` CR. You can specify one at a time. If you want to scale broker only, no node will need restart to apply the changes. But if you want to scale controller, all nodes will need restart to apply the changes.

#### Verify Topology cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Solr` object and related `PetSets` and `Pods`.

Let's wait for `SolrOpsRequest` to be `Successful`. Run the following command to watch `SolrOpsRequest` CR,

```bash
$ watch kubectl get Solropsrequest -n demo
NAME                        TYPE                STATUS       AGE
slops-hscale-up-topology    HorizontalScaling   Successful   106s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe slops -n demo slops-hscale-up-topology 
Name:         slops-hscale-up-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-07T07:40:35Z
  Generation:          1
  Resource Version:    1870552
  UID:                 142fb5b9-26ec-4dab-ad39-ebfd4470b1db
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Horizontal Scaling:
    Data:      2
    Overseer:  2
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-11-07T07:40:35Z
    Message:               Solr ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-11-07T07:41:08Z
    Message:               ScaleUp solr-cluster-data nodes
    Observed Generation:   1
    Reason:                HorizontalScaleDataNode
    Status:                True
    Type:                  HorizontalScaleDataNode
    Last Transition Time:  2024-11-07T07:40:43Z
    Message:               patch pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSet
    Last Transition Time:  2024-11-07T07:41:38Z
    Message:               is node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeInCluster
    Last Transition Time:  2024-11-07T07:41:43Z
    Message:               ScaleUp solr-cluster-overseer nodes
    Observed Generation:   1
    Reason:                HorizontalScaleOverseerNode
    Status:                True
    Type:                  HorizontalScaleOverseerNode
    Last Transition Time:  2024-11-07T07:41:43Z
    Message:               Successfully completed horizontally scale Solr cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age    From                         Message
  ----     ------                                     ----   ----                         -------
  Normal   Starting                                   7m24s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-hscale-up-topology
  Normal   Starting                                   7m24s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                                 7m24s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: slops-hscale-up-topology
  Warning  patch pet set; ConditionStatus:True        7m16s  KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  7m11s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   6m54s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleDataNode                    6m51s  KubeDB Ops-manager Operator  ScaleUp solr-cluster-data nodes
  Warning  patch pet set; ConditionStatus:True        6m46s  KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  6m41s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   6m21s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleOverseerNode                6m16s  KubeDB Ops-manager Operator  ScaleUp solr-cluster-overseer nodes
  Normal   Starting                                   6m16s  KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                                 6m16s  KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: slops-hscale-up-topolog
```

Now, we are going to verify the number of replicas this cluster has from the Solr object, number of pods the petset have,

**Broker Replicas**

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.data.replicas'
2
$ kubectl get petset -n demo solr-cluster-data -o json | jq '.spec.replicas'
2
```

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.overseer.replicas'
2
$ kubectl get petset -n demo solr-cluster-overseer -o json | jq '.spec.replicas'
2
```

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.coordinator.replicas'
2
$ kubectl get petset -n demo solr-cluster-coordinator -o json | jq '.spec.replicas'
2
```

From all the above outputs we can see that all data, overseer, coordinator of the topology Solr is `2`. That means we have successfully scaled up the replicas of the Solr topology cluster.

### Scale Down Replicas

Here, we are going to scale down the replicas of the Solr topology cluster to meet the desired number of replicas after scaling.

#### Create SolrOpsRequest

In order to scale down the replicas of the Solr topology cluster, we have to create a `SolrOpsRequest` CR with our desired replicas. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-hscale-down-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: HorizontalScaling
  horizontalScaling:
    data: 1
    overseer: 1
    coordinator: 1
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `solr-cluster` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Solr.
- `spec.horizontalScaling.topology.data` specifies the desired replicas after scaling for data.
- `spec.horizontalScaling.topology.overseer` specifies the desired replicas after scaling for overseer.
- `spec.horizontalScaling.topology.coordinator` specifies the desired replicas after scaling for coordinator.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Solr/scaling/horizontal-scaling/Solr-hscale-down-topology.yaml
solropsrequest.ops.kubedb.com/slops-hscale-down-topology created
```

#### Verify Topology cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Solr` object and related `PetSets` and `Pods`.

Let's wait for `SolrOpsRequest` to be `Successful`. Run the following command to watch `SolrOpsRequest` CR,

```bash
$ watch kubectl get solropsrequest -n demo
NAME                          TYPE                STATUS       AGE
slops-hscale-down-topology    HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe slops -n demo slops-hscale-down-topology 
Name:         slops-hscale-down-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-07T07:54:53Z
  Generation:          1
  Resource Version:    1872016
  UID:                 67c6912b-0658-43ed-af65-8cf6b249c567
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Horizontal Scaling:
    Data:      1
    Overseer:  1
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-11-07T07:54:53Z
    Message:               Solr ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-11-07T07:56:11Z
    Message:               ScaleDown solr-cluster-data nodes
    Observed Generation:   1
    Reason:                HorizontalScaleDataNode
    Status:                True
    Type:                  HorizontalScaleDataNode
    Last Transition Time:  2024-11-07T07:55:01Z
    Message:               reassign partitions; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReassignPartitions
    Last Transition Time:  2024-11-07T07:55:01Z
    Message:               is pet set patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetPatched
    Last Transition Time:  2024-11-07T07:57:21Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-11-07T07:56:06Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2024-11-07T07:57:21Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-11-07T07:57:26Z
    Message:               ScaleDown solr-cluster-overseer nodes
    Observed Generation:   1
    Reason:                HorizontalScaleOverseerNode
    Status:                True
    Type:                  HorizontalScaleOverseerNode
    Last Transition Time:  2024-11-07T07:57:26Z
    Message:               Successfully completed horizontally scale Solr cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age    From                         Message
  ----     ------                                     ----   ----                         -------
  Normal   Starting                                   2m46s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-hscale-down-topology
  Normal   Starting                                   2m46s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                                 2m46s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: slops-hscale-down-topology
  Warning  reassign partitions; ConditionStatus:True  2m38s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True   2m38s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:False             2m33s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              93s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           93s    KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False             93s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              93s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           93s    KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True              93s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   HorizontalScaleDataNode                    88s    KubeDB Ops-manager Operator  ScaleDown solr-cluster-data nodes
  Warning  is pet set patched; ConditionStatus:True   83s    KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:False             78s    KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              18s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           18s    KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False             18s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              18s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           18s    KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True              18s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   HorizontalScaleOverseerNode                13s    KubeDB Ops-manager Operator  ScaleDown solr-cluster-overseer nodes
  Normal   Starting                                   13s    KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                                 13s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: slops-hscale-down-topology
```

Now, we are going to verify the number of replicas this cluster has from the Solr object, number of pods the petset have,

Let's check the number of replicas has from Solr object, number of pods the petset have,

**Data Replicas**

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.data.replicas'
1
$ kubectl get petset -n demo solr-cluster-data -o json | jq '.spec.replicas'
1
```

**Overseer Replicas**

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.overseer.replicas'
1
$ kubectl get petset -n demo solr-cluster-overseer -o json | jq '.spec.replicas'
1

```

**Coordinator Replicas**

```bash
$ kubectl get solr -n demo solr-cluster -o json | jq '.spec.topology.coordinator.replicas'
1
$ kubectl get petset -n demo solr-cluster-coordinator -o json | jq '.spec.replicas'
1

```


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete sl -n demo solr-cluster
kubectl delete solropsrequest -n demo slops-hscale-up-topology slops-hscale-down-topology
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
