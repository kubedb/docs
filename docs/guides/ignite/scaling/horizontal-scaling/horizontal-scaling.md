---
title: Horizontal Scaling Ignite
menu:
  docs_{{ .version }}:
    identifier: ig-horizontal-scaling-ops
    name: Scale Horizontally
    parent: ig-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Ignite

This guide will show you how to use `KubeDB` Ops-manager operator to scale the Ignite Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Ignite](/docs/guides/ignite/concepts/ignite.md)
  - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/ignite/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/ignite](/docs/examples/ignite) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on ignite

Here, we are going to deploy a `Ignite` using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Deploy Ignite 

In this section, we are going to deploy a Ignite. We are going to deploy a `Ignite` with version `2.17.0`. Then, in the next section we will scale the ignite using `IgniteOpsRequest` CRD. Below is the YAML of the `Ignite` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite
  namespace: demo
spec:
  version: "2.17.0"
  replicas: 3
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
Let's create the `Ignite` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/scaling/ignite-cluster.yaml
ignite.kubedb.com/ignite created
```

Now, wait until `ignite` has status `Ready`. i.e,

```bash
$ kubectl get ig -n demo
NAME            TYPE                  VERSION   STATUS   AGE
ignite          kubedb.com/v1alpha2   2.17.0    Ready    2m
```

Let's check the number of replicas this ignite has from the Ignite object, number of pods the PetSet have,

```bash
$ kubectl get ignite -n demo ignite -o json | jq '.spec.replicas'
1

$ kubectl get petset -n demo ignite -o json | jq '.spec.replicas'
1
```

We can see from both command that the ignite has 3 replicas. 

We are now ready to apply the `IgniteOpsRequest` CR to scale this ignite.

## Scale Up Replicas

Here, we are going to scale up the replicas of the ignite to meet the desired number of replicas after scaling.

#### Create IgniteOpsRequest

In order to scale up the replicas of the ignite, we have to create a `IgniteOpsRequest` CR with our desired replicas. Below is the YAML of the `IgniteOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ignite-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ignite
  horizontalScaling:
    node: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `pp-horizontal` ignite.
- `spec.type` specifies that we are performing `HorizontalScaling` on our ignite.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `IgniteOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/scaling/horizontal-scaling/ig-hscale-up-ops.yaml
igniteopsrequest.ops.kubedb.com/ignite-horizontal-scale-up created
```

#### Verify replicas scaled up successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Ignite` object and related `PetSet`.

Let's wait for `IgniteOpsRequest` to be `Successful`.  Run the following command to watch `IgniteOpsRequest` CR,

```bash
$ watch kubectl get igniteopsrequest -n demo
Every 2.0s: kubectl get igniteopsrequest -n demo
NAME                           TYPE                STATUS       AGE
ignite-horizontal-scale-up   HorizontalScaling   Successful   2m49s
```

We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed to scale the ignite.

```bash
$ kubectl describe igniteopsrequest -n demo ignite-horizontal-scale-up
Name:         ignite-horizontal-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2024-09-12T10:48:21Z
  Generation:          1
  Resource Version:    46348
  UID:                 eaa4653f-f07a-47d1-935f-5a82d64ea659
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ignite
  Horizontal Scaling:
    Node:  5
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-09-12T11:42:06Z
    Message:               Ignite ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-09-12T11:42:59Z
    Message:               Successfully Scaled Up Node
    Observed Generation:   1
    Reason:                HorizontalScaleUp
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2024-09-12T11:42:14Z
    Message:               patch petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset
    Last Transition Time:  2024-09-12T11:42:19Z
    Message:               client failure; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure
    Last Transition Time:  2024-09-12T11:42:54Z
    Message:               is node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeInCluster
    Last Transition Time:  2024-09-12T11:42:39Z
    Message:               ignite already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-09-12T11:43:04Z
    Message:               successfully reconciled the Ignite with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-09-12T11:43:04Z
    Message:               Successfully completed horizontally scale Ignite cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age    From                         Message
  ----     ------                                     ----   ----                         -------
  Normal   Starting                                   8m40s  KubeDB Ops-manager Operator  Start processing for IgniteOpsRequest: demo/ignite-horizontal-scale-up
  Normal   Starting                                   8m40s  KubeDB Ops-manager Operator  Pausing Ignite databse: demo/ignite
  Normal   Successful                                 8m40s  KubeDB Ops-manager Operator  Successfully paused Ignite database: demo/ignite for IgniteOpsRequest: ignite-horizontal-scale-up
  Warning  patch petset; ConditionStatus:True         8m32s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  client failure; ConditionStatus:True       8m27s  KubeDB Ops-manager Operator  client failure; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  8m27s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   7m52s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleUp                          7m47s  KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdatePetSets                              7m42s  KubeDB Ops-manager Operator  successfully reconciled the Ignite with modified node
  Normal   Starting                                   7m42s  KubeDB Ops-manager Operator  Resuming Ignite database: demo/ignite
  Normal   Successful                                 7m42s  KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/ignite for IgniteOpsRequest: ignite-horizontal-scale-up
```

Now, we are going to verify the number of replicas this ignite has from the Pgpool object, number of pods the PetSet have,

```bash
$ kubectl get ig -n demo ignite -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo ignite -o json | jq '.spec.replicas'
3
```
From all the above outputs we can see that the replicas of the ignite is `3`. That means we have successfully scaled up the replicas of the Ignite.


### Scale Down Replicas

Here, we are going to scale down the replicas of the ignite to meet the desired number of replicas after scaling.

#### Create IgniteOpsRequest

In order to scale down the replicas of the ignite, we have to create a `IgniteOpsRequest` CR with our desired replicas. Below is the YAML of the `IgniteOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ignite-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ignite
  horizontalScaling:
    node: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `ignite` ignite.
- `spec.type` specifies that we are performing `HorizontalScaling` on our ignite.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `IgniteOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/scaling/horizontal-scaling/igops-hscale-down-ops.yaml
igniteopsrequest.ops.kubedb.com/ignite-horizontal-scale-down created
```

#### Verify replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Ignite` object and related `PetSet`.

Let's wait for `IgniteOpsRequest` to be `Successful`.  Run the following command to watch `IgniteOpsRequest` CR,

```bash
$ watch kubectl get igniteopsrequest -n demo
Every 2.0s: kubectl get igniteopsrequest -n demo
NAME                           TYPE                STATUS       AGE
ignite-horizontal-scale-down   HorizontalScaling   Successful   75s
```

We can see from the above output that the `IgniteOpsRequest` has succeeded. If we describe the `IgniteOpsRequest` we will get an overview of the steps that were followed to scale the ignite.

```bash
$ kubectl describe igniteopsrequest -n demo ignite-horizontal-scale-down
Name:         ignite-horizontal-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         IgniteOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T08:52:28Z
  Generation:          1
  Resource Version:    63600
  UID:                 019f9d8f-c2b0-4154-b3d3-b715b8805fd7
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  ignite
  Horizontal Scaling:
    Node:  2
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-07-17T08:52:28Z
    Message:               Pgpool ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-07-17T08:52:31Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-17T08:53:16Z
    Message:               Successfully Scaled Down Node
    Observed Generation:   1
    Reason:                HorizontalScaleDown
    Status:                True
    Type:                  HorizontalScaleDown
    Last Transition Time:  2024-07-17T08:52:36Z
    Message:               patch petset; ConditionStatus:True; PodName:pp-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--pp-horizontal-2
    Last Transition Time:  2024-07-17T08:52:36Z
    Message:               pp-horizontal already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-07-17T08:52:41Z
    Message:               get pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  GetPod
    Last Transition Time:  2024-07-17T08:53:11Z
    Message:               get pod; ConditionStatus:True; PodName:pp-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pp-horizontal-2
    Last Transition Time:  2024-07-17T08:53:16Z
    Message:               Successfully updated Pgpool
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-17T08:53:16Z
    Message:               Successfully completed horizontally scale ignite cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/ignite-horizontal-scale-down
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/ignite
  Normal   Successful                                                   96s   KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/ignite for PgpoolOpsRequest: ignite-horizontal-scale-down
  Normal   patch petset; ConditionStatus:True; PodName:pp-horizontal-2  88s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   get pod; ConditionStatus:False                               83s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Normal   get pod; ConditionStatus:True; PodName:pp-horizontal-2       53s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   HorizontalScaleDown                                          48s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdateDatabase                                               48s   KubeDB Ops-manager Operator  Successfully updated Ignite
  Normal   Starting                                                     48s   KubeDB Ops-manager Operator  Resuming Pgpool database: demo/ignite
  Normal   Successful                                                   48s   KubeDB Ops-manager Operator  Successfully resumed Ignite database: demo/ignite for IgniteOpsRequest: ignite-horizontal-scale-down
```

Now, we are going to verify the number of replicas this ignite has from the Ignite object, number of pods the petset have,

```bash
$ kubectl get ig -n demo ignite -o json | jq '.spec.replicas'
2

$ kubectl get petset -n demo ignite -o json | jq '.spec.replicas'
2
```
From all the above outputs we can see that the replicas of the ignite is `2`. That means we have successfully scaled up the replicas of the Ignite.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ig -n demo
kubectl delete igniteopsrequest -n demo ignite-horizontal-scale-down
```