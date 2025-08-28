---
title: Horizontal Scaling ClickHouse Cluster
menu:
  docs_{{ .version }}:
    identifier: ch-horizontal-scaling-cluster
    name: Cluster
    parent: ch-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale ClickHouse Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to scale the ClickHouse cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)
    - [Horizontal Scaling Overview](/docs/guides/clickhouse/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/clickhouse](/docs/examples/clickhouse) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on ClickHouse Cluster

Here, we are going to deploy a `ClickHouse` cluster using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare ClickHouse cluster

Now, we are going to deploy a `ClickHouse` cluster with version `24.4.1`.

### Deploy ClickHouse Cluster

In this section, we are going to deploy a ClickHouse cluster. Then, in the next section we will scale the cluster using `ClickHouseOpsRequest` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
        name: appscode-cluster
        shards: 2
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: clickhouse
                resources:
                  limits:
                    memory: 4Gi
                  requests:
                    cpu: 500m
                    memory: 2Gi
            initContainers:
              - name: clickhouse-init
                resources:
                  limits:
                    memory: 1Gi
                  requests:
                    cpu: 500m
                    memory: 1Gi
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/scaling/clickhouse-cluster.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
➤ kubectl get clickhouse -n demo -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   4s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   50s
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m5s
```

Let's check the number of replicas has from clickhouse object, number of pods the petset have,

```bash
➤ kubectl get petset -n demo clickhouse-prod-appscode-cluster-shard-0 -o json | jq '.spec.replicas'
2
```

We can see from commands that the cluster has 2 replicas for shard-0 as we have defined in the yaml.

We are now ready to apply the `ClickHouseOpsRequest` CR to scale this cluster.

## Scale Up Replicas

Here, we are going to scale up the replicas of the  cluster to meet the desired number of replicas after scaling.

#### Create ClickHouseOpsRequest

In order to scale up the replicas of the cluster, we have to create a `ClickHouseOpsRequest` CR with our desired replicas. Below is the YAML of the `ClickHouseOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-scale-horizontal-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: clickhouse-prod
  horizontalScaling:
    replicas: 4
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on clickhouse.
- `spec.horizontalScaling.cluster[index].replicas` specifies the desired replicas after scaling for clickhouse cluster.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/scaling/horizontal-scaling/chops-horizontal-scaling-up.yaml
clickhouseopsrequest.ops.kubedb.com/chops-scale-horizontal-up created
```

#### Verify cluster replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `ClickHouse` object and related `PetSets` and `Pods`.

Let's wait for `ClickHouseOpsRequest` to be `Successful`. Run the following command to watch `ClickHouseOpsRequest` CR,

```bash
➤ kubectl get clickhouseopsrequest -n demo
NAME                        TYPE                STATUS       AGE
chops-scale-horizontal-up   HorizontalScaling   Successful   3m16s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-scale-horizontal-up 
Name:         chops-scale-horizontal-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T12:23:20Z
  Generation:          1
  Resource Version:    820391
  UID:                 bfb2f335-01c0-4710-8784-c78f92386e76
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  clickhouse-prod
  Horizontal Scaling:
    Cluster:
      Cluster Name:  appscode-cluster
      Replicas:      4
  Type:              HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-08-25T12:23:20Z
    Message:               ClickHouse ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2025-08-25T12:25:28Z
    Message:               Successfully Scaled Up Node
    Observed Generation:   1
    Reason:                HorizontalScaleUp
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2025-08-25T12:23:28Z
    Message:               patch petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset
    Last Transition Time:  2025-08-25T12:25:18Z
    Message:               is pod running; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPodRunning
    Last Transition Time:  2025-08-25T12:25:23Z
    Message:               client failure; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure
    Last Transition Time:  2025-08-25T12:23:48Z
    Message:               is node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeInCluster
    Last Transition Time:  2025-08-25T12:25:33Z
    Message:               clickhouse-prod-appscode-cluster-shard-1 already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-08-25T12:25:33Z
    Message:               successfully reconciled the ClickHouse with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T12:25:33Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T12:25:33Z
    Message:               Successfully completed horizontally scale ClickHouse cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age    From                         Message
  ----     ------                                    ----   ----                         -------
  Normal   Starting                                  3m48s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-scale-horizontal-up
  Normal   Starting                                  3m48s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                3m48s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-scale-horizontal-up
  Warning  patch petset; ConditionStatus:True        3m40s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  is pod running; ConditionStatus:False     3m35s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      3m25s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      3m25s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:False     3m25s  KubeDB Ops-manager Operator  client failure; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      3m20s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      3m20s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:True      3m20s  KubeDB Ops-manager Operator  client failure; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:True  3m20s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Warning  patch petset; ConditionStatus:True        3m15s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  is pod running; ConditionStatus:False     3m10s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      2m55s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      2m55s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:False     2m55s  KubeDB Ops-manager Operator  client failure; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      2m50s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      2m50s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:True      2m50s  KubeDB Ops-manager Operator  client failure; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:True  2m50s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleUp                         2m45s  KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Warning  patch petset; ConditionStatus:True        2m40s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  is pod running; ConditionStatus:False     2m35s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      2m20s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      2m20s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:False     2m20s  KubeDB Ops-manager Operator  client failure; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      2m15s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      2m15s  KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:True      2m15s  KubeDB Ops-manager Operator  client failure; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:True  2m15s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Warning  patch petset; ConditionStatus:True        2m10s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  is pod running; ConditionStatus:False     2m5s   KubeDB Ops-manager Operator  is pod running; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      110s   KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      110s   KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:False     110s   KubeDB Ops-manager Operator  client failure; ConditionStatus:False
  Warning  is pod running; ConditionStatus:True      105s   KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  is pod running; ConditionStatus:True      105s   KubeDB Ops-manager Operator  is pod running; ConditionStatus:True
  Warning  client failure; ConditionStatus:True      105s   KubeDB Ops-manager Operator  client failure; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:True  105s   KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleUp                         100s   KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Warning  reconcile; ConditionStatus:True           95s    KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True           95s    KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True           95s    KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                             95s    KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with modified node
  Normal   Starting                                  95s    KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                95s    KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-scale-horizontal-up
```

Now, we are going to verify the number of replicas this cluster has from the ClickHouse object, number of pods the petset have,

```bash
➤ kubectl get petset -n demo clickhouse-prod-appscode-cluster-shard-0 -o json | jq '.spec.replicas'
4
```

### Scale Down Replicas

Here, we are going to scale down the replicas of the clickhouse cluster to meet the desired number of replicas after scaling.

#### Create ClickHouseOpsRequest

In order to scale down the replicas of the clickhouse cluster, we have to create a `ClickHouseOpsRequest` CR with our desired replicas. Below is the YAML of the `ClickHouseOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-scale-horizontal-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: clickhouse-prod
  horizontalScaling:
    cluster:
      - clusterName: appscode-cluster
        replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `HorizontalScaling` on clickhouse.
- `spec.horizontalScaling.cluster[index].replicas` specifies the desired replicas after scaling for the clickhouse nodes.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/scaling/horizontal-scaling/chops-horizontal-scaling-down.yaml
clickhouseopsrequest.ops.kubedb.com/chops-scale-horizontal-down created
```

#### Verify clickhouse cluster replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `ClickHouse` object and related `PetSets` and `Pods`.

Let's wait for `ClickHouseOpsRequest` to be `Successful`. Run the following command to watch `ClickHouseOpsRequest` CR,

```bash
➤ kubectl get clickhouseopsrequest -n demo chops-scale-horizontal-down
NAME                          TYPE                STATUS       AGE
chops-scale-horizontal-down   HorizontalScaling   Successful   60s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-scale-horizontal-down 
Name:         chops-scale-horizontal-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T12:30:09Z
  Generation:          1
  Resource Version:    821351
  UID:                 09e2b059-8ea7-4683-a5e5-c83ed14283cd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  clickhouse-prod
  Horizontal Scaling:
    Cluster:
      Cluster Name:  appscode-cluster
      Replicas:      3
  Type:              HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-08-25T12:30:09Z
    Message:               ClickHouse ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2025-08-25T12:30:52Z
    Message:               Successfully Scaled Down Node
    Observed Generation:   1
    Reason:                HorizontalScaleDown
    Status:                True
    Type:                  HorizontalScaleDown
    Last Transition Time:  2025-08-25T12:30:17Z
    Message:               reassign partitions; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReassignPartitions
    Last Transition Time:  2025-08-25T12:30:17Z
    Message:               is pet set patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetSetPatched
    Last Transition Time:  2025-08-25T12:30:58Z
    Message:               clickhouse-prod-appscode-cluster-shard-1 already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-08-25T12:30:22Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2025-08-25T12:30:22Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2025-08-25T12:30:22Z
    Message:               get pvc; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  GetPvc
    Last Transition Time:  2025-08-25T12:30:58Z
    Message:               successfully reconciled the ClickHouse with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T12:30:57Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T12:30:58Z
    Message:               Successfully completed horizontally scale ClickHouse cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age    From                         Message
  ----     ------                                     ----   ----                         -------
  Normal   Starting                                   4m33s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-scale-horizontal-down
  Normal   Starting                                   4m33s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                 4m33s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-scale-horizontal-down
  Warning  reassign partitions; ConditionStatus:True  4m25s  KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True   4m25s  KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True              4m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           4m20s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False             4m20s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              4m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           4m15s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           4m15s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Normal   HorizontalScaleDown                        4m10s  KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Warning  reassign partitions; ConditionStatus:True  4m5s   KubeDB Ops-manager Operator  reassign partitions; ConditionStatus:True
  Warning  is pet set patched; ConditionStatus:True   4m5s   KubeDB Ops-manager Operator  is pet set patched; ConditionStatus:True
  Warning  get pod; ConditionStatus:True              4m     KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           4m     KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:False             4m     KubeDB Ops-manager Operator  get pvc; ConditionStatus:False
  Warning  get pod; ConditionStatus:True              3m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           3m55s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  delete pvc; ConditionStatus:True           3m55s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Normal   HorizontalScaleDown                        3m50s  KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Warning  reconcile; ConditionStatus:True            3m45s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True            3m45s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True            3m44s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                              3m44s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with modified node
  Normal   Starting                                   3m44s  KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                 3m44s  KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-scale-horizontal-down
```

Now, we are going to verify the number of replicas this cluster has from the number of pods the petset have,

```bash
➤ kubectl get petset -n demo clickhouse-prod-appscode-cluster-shard-0 -o json | jq '.spec.replicas'
3
```

From all the above outputs we can see that the replicas of the clickhouse cluster is `3`. That means we have successfully scaled down the replicas of the ClickHouse cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete clickhouse -n demo clickhouse-prod
kubectl delete clickhouseopsrequests -n demo chops-scale-horizontal-down chops-scale-horizontal-up
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
