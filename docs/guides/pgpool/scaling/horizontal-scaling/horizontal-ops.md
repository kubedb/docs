---
title: Horizontal Scaling Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-horizontal-scaling-ops
    name: HorizontalScaling OpsRequest
    parent: pp-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Pgpool

This guide will show you how to use `KubeDB` Ops-manager operator to scale the replicaset of a Pgpool.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Pgpool](/docs/guides/pgpool/concepts/pgpool.md)
  - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/pgpool/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgpool](/docs/examples/pgpool) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on pgpool

Here, we are going to deploy a  `Pgpool` using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare Pgpool

Now, we are going to deploy a `Pgpool` with version `4.5.0`.

### Deploy Pgpool 

In this section, we are going to deploy a Pgpool. Then, in the next section we will scale the pgpool using `PgpoolOpsRequest` CRD. Below is the YAML of the `Pgpool` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-horizontal
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  initConfig:
    pgpoolConfig:
      max_pool : 60
  deletionPolicy: WipeOut
```
Here we are creating the pgpool with `max_pool=60`, it is necessary because we will up scale the pgpool replicas so for that we need larger `max_pool`. Let's create the `Pgpool` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/scaling/pp-horizontal.yaml
pgpool.kubedb.com/pp-horizontal created
```

Now, wait until `pp-horizontal ` has status `Ready`. i.e,

```bash
$ kubectl get pp -n demo
NAME            TYPE                  VERSION   STATUS   AGE
pp-horizontal   kubedb.com/v1alpha2   4.5.0     Ready    2m
```

Let's check the number of replicas this pgpool has from the Pgpool object, number of pods the petset have,

```bash
$ kubectl get pgpool -n demo pp-horizontal -o json | jq '.spec.replicas'
1

$ kubectl get petset -n demo pp-horizontal -o json | jq '.spec.replicas'
1
```

We can see from both command that the pgpool has 3 replicas. 

We are now ready to apply the `PgpoolOpsRequest` CR to scale this pgpool.

## Scale Up Replicas

Here, we are going to scale up the replicas of the pgpool to meet the desired number of replicas after scaling.

#### Create PgpoolOpsRequest

In order to scale up the replicas of the pgpool, we have to create a `PgpoolOpsRequest` CR with our desired replicas. Below is the YAML of the `PgpoolOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: pgpool-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pp-horizontal
  horizontalScaling:
    node: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `pp-horizontal` pgpool.
- `spec.type` specifies that we are performing `HorizontalScaling` on our pgpool.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/scaling/horizontal-scaling/ppops-hscale-up-ops.yaml
pgpoolopsrequest.ops.kubedb.com/pgpool-horizontal-scale-up created
```

#### Verify replicas scaled up successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Pgpool` object and related `PetSet`.

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CR,

```bash
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                         TYPE                STATUS       AGE
pgpool-horizontal-scale-up   HorizontalScaling   Successful   2m49s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed to scale the pgpool.

```bash
$ kubectl describe pgpoolopsrequest -n demo pgpool-horizontal-scale-up
Name:         pgpool-horizontal-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T08:35:13Z
  Generation:          1
  Resource Version:    62002
  UID:                 ce44f7a1-e78d-4248-a691-62fe1efd11f3
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pp-horizontal
  Horizontal Scaling:
    Node:  3
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-07-17T08:35:13Z
    Message:               Pgpool ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-07-17T08:35:16Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-17T08:35:41Z
    Message:               Successfully Scaled Up Node
    Observed Generation:   1
    Reason:                HorizontalScaleUp
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2024-07-17T08:35:21Z
    Message:               patch petset; ConditionStatus:True; PodName:pp-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--pp-horizontal-1
    Last Transition Time:  2024-07-17T08:35:26Z
    Message:               is pod ready; ConditionStatus:True; PodName:pp-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--pp-horizontal-1
    Last Transition Time:  2024-07-17T08:35:26Z
    Message:               client failure; ConditionStatus:True; PodName:pp-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure--pp-horizontal-1
    Last Transition Time:  2024-07-17T08:35:26Z
    Message:               is node healthy; ConditionStatus:True; PodName:pp-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeHealthy--pp-horizontal-1
    Last Transition Time:  2024-07-17T08:35:31Z
    Message:               patch petset; ConditionStatus:True; PodName:pp-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--pp-horizontal-2
    Last Transition Time:  2024-07-17T08:35:31Z
    Message:               pp-horizontal already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-07-17T08:35:36Z
    Message:               is pod ready; ConditionStatus:True; PodName:pp-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--pp-horizontal-2
    Last Transition Time:  2024-07-17T08:35:36Z
    Message:               client failure; ConditionStatus:True; PodName:pp-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure--pp-horizontal-2
    Last Transition Time:  2024-07-17T08:35:36Z
    Message:               is node healthy; ConditionStatus:True; PodName:pp-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeHealthy--pp-horizontal-2
    Last Transition Time:  2024-07-17T08:35:41Z
    Message:               Successfully updated Pgpool
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-17T08:35:41Z
    Message:               Successfully completed horizontally scale pgpool cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age    From                         Message
  ----     ------                                                          ----   ----                         -------
  Normal   Starting                                                        4m5s   KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/pgpool-horizontal-scale-up
  Normal   Starting                                                        4m5s   KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pp-horizontal
  Normal   Successful                                                      4m5s   KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pp-horizontal for PgpoolOpsRequest: pgpool-horizontal-scale-up
  Normal   patch petset; ConditionStatus:True; PodName:pp-horizontal-1     3m57s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pp-horizontal-1
  Normal   is pod ready; ConditionStatus:True; PodName:pp-horizontal-1     3m52s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:pp-horizontal-1
  Normal   is node healthy; ConditionStatus:True; PodName:pp-horizontal-1  3m52s  KubeDB Ops-manager Operator  is node healthy; ConditionStatus:True; PodName:pp-horizontal-1
  Normal   patch petset; ConditionStatus:True; PodName:pp-horizontal-2     3m47s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   is pod ready; ConditionStatus:True; PodName:pp-horizontal-2     3m42s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   is node healthy; ConditionStatus:True; PodName:pp-horizontal-2  3m42s  KubeDB Ops-manager Operator  is node healthy; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   HorizontalScaleUp                                               3m37s  KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdateDatabase                                                  3m37s  KubeDB Ops-manager Operator  Successfully updated Pgpool
  Normal   Starting                                                        3m37s  KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pp-horizontal
  Normal   Successful                                                      3m37s  KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pp-horizontal for PgpoolOpsRequest: pgpool-horizontal-scale-up
```

Now, we are going to verify the number of replicas this pgpool has from the Pgpool object, number of pods the petset have,

```bash
$ kubectl get pp -n demo pp-horizontal -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo pp-horizontal -o json | jq '.spec.replicas'
3
```
From all the above outputs we can see that the replicas of the pgpool is `3`. That means we have successfully scaled up the replicas of the Pgpool.


### Scale Down Replicas

Here, we are going to scale down the replicas of the pgpool to meet the desired number of replicas after scaling.

#### Create PgpoolOpsRequest

In order to scale down the replicas of the pgpool, we have to create a `PgpoolOpsRequest` CR with our desired replicas. Below is the YAML of the `PgpoolOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: pgpool-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pp-horizontal
  horizontalScaling:
    node: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `pp-horizontal` pgpool.
- `spec.type` specifies that we are performing `HorizontalScaling` on our pgpool.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/scaling/horizontal-scaling/ppops-hscale-down-ops.yaml
pgpoolopsrequest.ops.kubedb.com/pgpool-horizontal-scale-down created
```

#### Verify replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `Pgpool` object and related `PetSet`.

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CR,

```bash
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                           TYPE                STATUS       AGE
pgpool-horizontal-scale-down   HorizontalScaling   Successful   75s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed to scale the pgpool.

```bash
$ kubectl describe pgpoolopsrequest -n demo pgpool-horizontal-scale-down
Name:         pgpool-horizontal-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T08:52:28Z
  Generation:          1
  Resource Version:    63600
  UID:                 019f9d8f-c2b0-4154-b3d3-b715b8805fd7
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pp-horizontal
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
    Message:               Successfully completed horizontally scale pgpool cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/pgpool-horizontal-scale-down
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pp-horizontal
  Normal   Successful                                                   96s   KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pp-horizontal for PgpoolOpsRequest: pgpool-horizontal-scale-down
  Normal   patch petset; ConditionStatus:True; PodName:pp-horizontal-2  88s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   get pod; ConditionStatus:False                               83s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Normal   get pod; ConditionStatus:True; PodName:pp-horizontal-2       53s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   HorizontalScaleDown                                          48s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdateDatabase                                               48s   KubeDB Ops-manager Operator  Successfully updated Pgpool
  Normal   Starting                                                     48s   KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pp-horizontal
  Normal   Successful                                                   48s   KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pp-horizontal for PgpoolOpsRequest: pgpool-horizontal-scale-down
```

Now, we are going to verify the number of replicas this pgpool has from the Pgpool object, number of pods the petset have,

```bash
$ kubectl get pp -n demo pp-horizontal -o json | jq '.spec.replicas'
2

$ kubectl get petset -n demo pp-horizontal -o json | jq '.spec.replicas'
2
```
From all the above outputs we can see that the replicas of the pgpool is `2`. That means we have successfully scaled up the replicas of the Pgpool.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n pp-horizontal
kubectl delete pgpoolopsrequest -n demo pgpool-horizontal-scale-down
```