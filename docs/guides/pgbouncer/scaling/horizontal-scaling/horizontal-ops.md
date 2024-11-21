---
title: Horizontal Scaling PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-horizontal-scaling-ops
    name: HorizontalScaling OpsRequest
    parent: pb-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale PgBouncer

This guide will show you how to use `KubeDB` Ops-manager operator to scale the replicaset of a PgBouncer.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/pgbouncer/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgbouncer](/docs/examples/pgbouncer) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on pgbouncer

Here, we are going to deploy a  `PgBouncer` using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare PgBouncer

Now, we are going to deploy a `PgBouncer` with version `1.23.1`.

### Deploy PgBouncer 

In this section, we are going to deploy a PgBouncer. Then, in the next section we will scale the pgbouncer using `PgBouncerOpsRequest` CRD. Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-horizontal
  namespace: demo
spec:
  replicas: 1
  version: "1.18.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "ha-postgres"
      namespace: demo
  connectionPool:
    poolMode: session
    port: 5432
    reservePoolSize: 5
    maxClientConnections: 87
    defaultPoolSize: 2
    minPoolSize: 1
    authType: md5
  deletionPolicy: WipeOut
```
Let's create the `PgBouncer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/scaling/pb-horizontal.yaml
pgbouncer.kubedb.com/pb-horizontal created
```

Now, wait until `pb-horizontal ` has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
NAME            TYPE                  VERSION   STATUS   AGE
pb-horizontal   kubedb.com/v1   1.18.0    Ready    2m
```

Let's check the number of replicas this pgbouncer has from the PgBouncer object, number of pods the petset have,

```bash
$ kubectl get pgbouncer -n demo pb-horizontal -o json | jq '.spec.replicas'
1

$ kubectl get petset -n demo pb-horizontal -o json | jq '.spec.replicas'
1
```

We can see from both command that the pgbouncer has 3 replicas. 

We are now ready to apply the `PgBouncerOpsRequest` CR to scale this pgbouncer.

## Scale Up Replicas

Here, we are going to scale up the replicas of the pgbouncer to meet the desired number of replicas after scaling.

#### Create PgBouncerOpsRequest

In order to scale up the replicas of the pgbouncer, we have to create a `PgBouncerOpsRequest` CR with our desired replicas. Below is the YAML of the `PgBouncerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pb-horizontal
  horizontalScaling:
    replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `pb-horizontal` pgbouncer.
- `spec.type` specifies that we are performing `HorizontalScaling` on our pgbouncer.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/scaling/horizontal-scaling/pbops-hscale-up-ops.yaml
pgbounceropsrequest.ops.kubedb.com/pgbouncer-horizontal-scale-up created
```

#### Verify replicas scaled up successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `PgBouncer` object and related `PetSet`.

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CR,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                           TYPE                STATUS       AGE
pgbouncer-horizontal-scale-up  HorizontalScaling   Successful   2m49s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to scale the pgbouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pgbouncer-horizontal-scale-up
Name:         pgbouncer-horizontal-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T08:35:13Z
  Generation:          1
  Resource Version:    62002
  UID:                 ce44f7a1-e78d-4248-a691-62fe1efd11f3
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pb-horizontal
  Horizontal Scaling:
    Node:  3
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-07-17T08:35:13Z
    Message:               PgBouncer ops-request has started to horizontally scaling the nodes
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
    Message:               patch petset; ConditionStatus:True; PodName:pb-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--pb-horizontal-1
    Last Transition Time:  2024-07-17T08:35:26Z
    Message:               is pod ready; ConditionStatus:True; PodName:pb-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--pb-horizontal-1
    Last Transition Time:  2024-07-17T08:35:26Z
    Message:               client failure; ConditionStatus:True; PodName:pb-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure--pb-horizontal-1
    Last Transition Time:  2024-07-17T08:35:26Z
    Message:               is node healthy; ConditionStatus:True; PodName:pb-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeHealthy--pb-horizontal-1
    Last Transition Time:  2024-07-17T08:35:31Z
    Message:               patch petset; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--pb-horizontal-2
    Last Transition Time:  2024-07-17T08:35:31Z
    Message:               pb-horizontal already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-07-17T08:35:36Z
    Message:               is pod ready; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--pb-horizontal-2
    Last Transition Time:  2024-07-17T08:35:36Z
    Message:               client failure; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure--pb-horizontal-2
    Last Transition Time:  2024-07-17T08:35:36Z
    Message:               is node healthy; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeHealthy--pb-horizontal-2
    Last Transition Time:  2024-07-17T08:35:41Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-17T08:35:41Z
    Message:               Successfully completed horizontally scale pgbouncer cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age    From                         Message
  ----     ------                                                          ----   ----                         -------
  Normal   Starting                                                        4m5s   KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-up
  Normal   Starting                                                        4m5s   KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-horizontal
  Normal   Successful                                                      4m5s   KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-up
  Normal   patch petset; ConditionStatus:True; PodName:pb-horizontal-1     3m57s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pb-horizontal-1
  Normal   is pod ready; ConditionStatus:True; PodName:pb-horizontal-1     3m52s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:pb-horizontal-1
  Normal   is node healthy; ConditionStatus:True; PodName:pb-horizontal-1  3m52s  KubeDB Ops-manager Operator  is node healthy; ConditionStatus:True; PodName:pb-horizontal-1
  Normal   patch petset; ConditionStatus:True; PodName:pb-horizontal-2     3m47s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pb-horizontal-2
  Normal   is pod ready; ConditionStatus:True; PodName:pb-horizontal-2     3m42s  KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:pb-horizontal-2
  Normal   is node healthy; ConditionStatus:True; PodName:pb-horizontal-2  3m42s  KubeDB Ops-manager Operator  is node healthy; ConditionStatus:True; PodName:pb-horizontal-2
  Normal   HorizontalScaleUp                                               3m37s  KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdateDatabase                                                  3m37s  KubeDB Ops-manager Operator  Successfully updated PgBouncer
  Normal   Starting                                                        3m37s  KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-horizontal
  Normal   Successful                                                      3m37s  KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-up
```

Now, we are going to verify the number of replicas this pgbouncer has from the PgBouncer object, number of pods the petset have,

```bash
$ kubectl get pb -n demo pb-horizontal -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo pb-horizontal -o json | jq '.spec.replicas'
3
```
From all the above outputs we can see that the replicas of the pgbouncer is `3`. That means we have successfully scaled up the replicas of the PgBouncer.


### Scale Down Replicas

Here, we are going to scale down the replicas of the pgbouncer to meet the desired number of replicas after scaling.

#### Create PgBouncerOpsRequest

In order to scale down the replicas of the pgbouncer, we have to create a `PgBouncerOpsRequest` CR with our desired replicas. Below is the YAML of the `PgBouncerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pb-horizontal
  horizontalScaling:
    replicas: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `pb-horizontal` pgbouncer.
- `spec.type` specifies that we are performing `HorizontalScaling` on our pgbouncer.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/scaling/horizontal-scaling/pbops-hscale-down-ops.yaml
pgbounceropsrequest.ops.kubedb.com/pgbouncer-horizontal-scale-down created
```

#### Verify replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `PgBouncer` object and related `PetSet`.

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CR,

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                              TYPE                STATUS       AGE
pgbouncer-horizontal-scale-down   HorizontalScaling   Successful   75s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to scale the pgbouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pgbouncer-horizontal-scale-down
Name:         pgbouncer-horizontal-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T08:52:28Z
  Generation:          1
  Resource Version:    63600
  UID:                 019f9d8f-c2b0-4154-b3d3-b715b8805fd7
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pb-horizontal
  Horizontal Scaling:
    Node:  2
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-07-17T08:52:28Z
    Message:               PgBouncer ops-request has started to horizontally scaling the nodes
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
    Message:               patch petset; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--pb-horizontal-2
    Last Transition Time:  2024-07-17T08:52:36Z
    Message:               pb-horizontal already has desired replicas
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
    Message:               get pod; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-horizontal-2
    Last Transition Time:  2024-07-17T08:53:16Z
    Message:               Successfully updated PgBouncer
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-07-17T08:53:16Z
    Message:               Successfully completed horizontally scale pgbouncer cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-down
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-horizontal
  Normal   Successful                                                   96s   KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-down
  Normal   patch petset; ConditionStatus:True; PodName:pb-horizontal-2  88s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pb-horizontal-2
  Normal   get pod; ConditionStatus:False                               83s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Normal   get pod; ConditionStatus:True; PodName:pb-horizontal-2       53s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-horizontal-2
  Normal   HorizontalScaleDown                                          48s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdateDatabase                                               48s   KubeDB Ops-manager Operator  Successfully updated PgBouncer
  Normal   Starting                                                     48s   KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-horizontal
  Normal   Successful                                                   48s   KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-down
```

Now, we are going to verify the number of replicas this pgbouncer has from the PgBouncer object, number of pods the petset have,

```bash
$ kubectl get pb -n demo pb-horizontal -o json | jq '.spec.replicas'
2

$ kubectl get petset -n demo pb-horizontal -o json | jq '.spec.replicas'
2
```
From all the above outputs we can see that the replicas of the pgbouncer is `2`. That means we have successfully scaled down the replicas of the PgBouncer.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pb -n demo pb-horizontal
kubectl delete pgbounceropsrequest -n demo pgbouncer-horizontal-scale-down
```