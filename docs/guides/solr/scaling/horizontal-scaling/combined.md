---
title: Solr Combined Horizontal Scaling
menu:
  docs_{{ .version }}:
    identifier: sl-scaling-horizontal-combined
    name: Combined Cluster
    parent: sl-scaling-horizontal
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Solr Combined Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Solr combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [Combined](/docs/guides/solr/clustering/combined_cluster.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Horizontal Scaling Overview](/docs/guides/solr/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/solr](/docs/examples/solr) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on Combined Cluster

Here, we are going to deploy a  `Solr` combined cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Solr Combined cluster

Now, we are going to deploy a `Solr` combined cluster with version `9.4.1`.

### Deploy Solr combined cluster

In this section, we are going to deploy a Solr combined cluster. Then, in the next section we will scale the cluster using `SolrOpsRequest` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```bash
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-combined
  namespace: demo
spec:
  version: 9.4.1
  replicas: 2
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/scaling/horizontal/combined/solr.yaml
solr.kubedb.com/solr-combined created
```

Now, wait until `solr-combined` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME            TYPE                  VERSION   STATUS   AGE
solr-combined   kubedb.com/v1alpha2   9.4.1     Ready    65m
```

Let's check the number of replicas has from Solr object, number of pods the petset have,

```bash
$ kubectl get solr -n demo solr-combined -o json | jq '.spec.replicas'
2
$ kubectl get petset -n demo solr-combined -o json | jq '.spec.replicas'
2

```

We can see from both command that the cluster has 2 replicas.

Also, we can verify the replicas of the combined from an internal solr command by exec into a replica.

We can see from the above output that the Solr has 2 nodes.

We are now ready to apply the `SolrOpsRequest` CR to scale this cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the combined cluster to meet the desired number of replicas after scaling.

#### Create SolrOpsRequest

In order to scale up the replicas of the combined cluster, we have to create a `SolrOpsRequest` CR with our desired replicas. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-hscale-up-combined
  namespace: demo
spec:
  databaseRef:
    name: solr-combined
  type: HorizontalScaling
  horizontalScaling:
    node: 4
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `Solr-dev` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Solr.
- `spec.horizontalScaling.node` specifies the desired replicas after scaling.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/scaling/horizontal/combined/scaling.yaml
Solropsrequest.ops.kubedb.com/kfops-hscale-up-combined created
```

#### Verify Combined cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Solr` object and related `PetSets` and `Pods`.

Let's wait for `SolrOpsRequest` to be `Successful`. Run the following command to watch `SolrOpsRequest` CR,

```bash
$ watch kubectl get Solropsrequest -n demo
NAME                        TYPE                STATUS       AGE
slops-hscale-up-combined    HorizontalScaling   Successful   106s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe slops -n demo slops-hscale-up-combined 
Name:         slops-hscale-up-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-07T09:54:54Z
  Generation:          1
  Resource Version:    1883245
  UID:                 dfc22d44-6638-43f7-97b7-9846658cc061
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-combined
  Horizontal Scaling:
    Node:  4
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-11-07T09:54:54Z
    Message:               Solr ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-11-07T09:55:52Z
    Message:               ScaleUp solr-combined nodes
    Observed Generation:   1
    Reason:                HorizontalScaleCombinedNode
    Status:                True
    Type:                  HorizontalScaleCombinedNode
    Last Transition Time:  2024-11-07T09:55:02Z
    Message:               patch pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetSet
    Last Transition Time:  2024-11-07T09:55:48Z
    Message:               is node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeInCluster
    Last Transition Time:  2024-11-07T09:55:52Z
    Message:               Successfully completed horizontally scale Solr cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age   From                         Message
  ----     ------                                     ----  ----                         -------
  Normal   Starting                                   89s   KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-hscale-up-combined
  Normal   Starting                                   89s   KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-combined
  Normal   Successful                                 89s   KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-combined for SolrOpsRequest: slops-hscale-up-combined
  Warning  patch pet set; ConditionStatus:True        81s   KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  76s   KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   60s   KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Warning  patch pet set; ConditionStatus:True        56s   KubeDB Ops-manager Operator  patch pet set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  51s   KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   35s   KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleCombinedNode                31s   KubeDB Ops-manager Operator  ScaleUp solr-combined nodes
  Normal   Starting                                   31s   KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-combined
  Normal   Successful                                 31s   KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-combined for SolrOpsRequest: slops-hscale-up-combined
```

Now, we are going to verify the number of replicas this cluster has from the Solr object, number of pods the petset have,

```bash
$ kubectl get solr -n demo solr-combined -o json | jq '.spec.replicas'
4
$ kubectl get petset -n demo solr-combined -o json | jq '.spec.replicas'
4
```

### Scale Down Replicas

Here, we are going to scale down the replicas of the Solr combined cluster to meet the desired number of replicas after scaling.

#### Create SolrOpsRequest

In order to scale down the replicas of the Solr combined cluster, we have to create a `SolrOpsRequest` CR with our desired replicas. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-hscale-down-combined
  namespace: demo
spec:
  databaseRef:
    name: solr-combined
  type: HorizontalScaling
  horizontalScaling:
    node: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `Solr-dev` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on Solr.
- `spec.horizontalScaling.node` specifies the desired replicas after scaling.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/scaling/horizontal-scaling/solr-hscale-down-combined.yaml
solropsrequest.ops.kubedb.com/slops-hscale-down-combined created
```

#### Verify Combined cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Solr` object and related `PetSets` and `Pods`.

Let's wait for `SolrOpsRequest` to be `Successful`. Run the following command to watch `SolrOpsRequest` CR,

```bash
$ watch kubectl get Solropsrequest -n demo
NAME                          TYPE                STATUS       AGE
slops-hscale-down-combined    HorizontalScaling   Successful   2m32s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe slops -n demo slops-hscale-down-combined 
Name:         slops-hscale-down-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-07T09:59:14Z
  Generation:          1
  Resource Version:    1883926
  UID:                 6f028e60-ed6f-4716-920d-eb348f9bee80
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-combined
  Horizontal Scaling:
    Node:  2
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-11-07T09:59:14Z
    Message:               Solr ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-11-07T10:01:42Z
    Message:               ScaleDown solr-combined nodes
    Observed Generation:   1
    Reason:                HorizontalScaleCombinedNode
    Status:                True
    Type:                  HorizontalScaleCombinedNode
    Last Transition Time:  2024-11-07T09:59:22Z
    Message:               reassign partitions; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReassignPartitions
    Last Transition Time:  2024-11-07T09:59:22Z
    Message:               is pet set patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetPatched
    Last Transition Time:  2024-11-07T10:01:37Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-11-07T10:00:27Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2024-11-07T10:01:37Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-11-07T10:01:42Z
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
  Normal   Starting                                   8m37s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-hscale-down-combined
  Normal   Starting                                   8m37s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-combined
  Normal   Successful                                 8m37s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-combined for SolrOpsRequest: slops-hscale-down-combined
  Warning  reassign partitions; ConditionStatus:True  8m29s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True   8m29s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:False             8m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              7m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           7m24s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False             7m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              7m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           7m24s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True              7m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  reassign partitions; ConditionStatus:True  7m19s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True   7m19s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:False             7m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              6m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           6m14s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False             6m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              6m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           6m14s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True              6m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Normal   HorizontalScaleCombinedNode                6m9s   KubeDB Ops-manager Operator  ScaleDown solr-combined nodes
  Normal   Starting                                   6m9s   KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-combined
  Normal   Successful                                 6m9s   KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-combined for SolrOpsRequest: slops-hscale-down-combined
```

Now, we are going to verify the number of replicas this cluster has from the Solr object, number of pods the petset have,

```bash
$ kubectl get solr -n demo solr-combined -o json | jq '.spec.replicas'
2
$ kubectl get petset -n demo solr-combined -o json | jq '.spec.replicas'
2
```

From all the above outputs we can see that the replicas of the combined cluster is `2`. That means we have successfully scaled down the replicas of the Solr combined cluster.

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
