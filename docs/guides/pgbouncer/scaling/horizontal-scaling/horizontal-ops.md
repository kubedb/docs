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
NAME            VERSION   STATUS   AGE
pb-horizontal   1.18.0    Ready    2m19s
```

Let's check the number of replicas this pgbouncer has from the PgBouncer object, number of pods the petset have,

```bash
$ kubectl get pgbouncer -n demo pb-horizontal -o json | jq '.spec.replicas'
1

$ kubectl get petset -n demo pb-horizontal -o json | jq '.spec.replicas'
1
```

We can see from both command that the pgbouncer has 1 replicas. 

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/scaling/horizontal-scaling-ops.yaml
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
  Creation Timestamp:  2024-11-27T11:12:29Z
  Generation:          1
  Resource Version:    49162
  UID:                 ce390f66-e10f-490f-ad47-f28894d0569a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pb-horizontal
  Horizontal Scaling:
    Replicas:  3
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-11-27T11:12:29Z
    Message:               Controller has started to Progress with HorizontalScaling of PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-up
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2024-11-27T11:12:32Z
    Message:               Horizontal scaling started in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-up
    Observed Generation:   1
    Reason:                HorizontalScaleStarted
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-11-27T11:12:37Z
    Message:               patch p s; ConditionStatus:True; PodName:pb-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  PatchPS--pb-horizontal-1
    Last Transition Time:  2024-11-27T11:12:42Z
    Message:               is pg bouncer running; ConditionStatus:True; PodName:pb-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  IsPgBouncerRunning--pb-horizontal-1
    Last Transition Time:  2024-11-27T11:12:47Z
    Message:               patch p s; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPS--pb-horizontal-2
    Last Transition Time:  2024-11-27T11:12:52Z
    Message:               is pg bouncer running; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  IsPgBouncerRunning--pb-horizontal-2
    Last Transition Time:  2024-11-27T11:12:57Z
    Message:               Horizontal scaling Up performed successfully in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-up
    Observed Generation:   1
    Reason:                HorizontalScaleSucceeded
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2024-11-27T11:13:07Z
    Message:               Controller has successfully completed  with HorizontalScaling of PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-up
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age    From                         Message
  ----     ------                                                                ----   ----                         -------
  Normal   Starting                                                              2m13s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-up
  Normal   Starting                                                              2m13s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-horizontal
  Normal   Successful                                                            2m13s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-up
  Normal   Starting                                                              2m10s  KubeDB Ops-manager Operator  Horizontal scaling started in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-up
  Warning  patch p s; ConditionStatus:True; PodName:pb-horizontal-1              2m5s   KubeDB Ops-manager Operator  patch p s; ConditionStatus:True; PodName:pb-horizontal-1
  Warning  is pg bouncer running; ConditionStatus:True; PodName:pb-horizontal-1  2m     KubeDB Ops-manager Operator  is pg bouncer running; ConditionStatus:True; PodName:pb-horizontal-1
  Warning  patch p s; ConditionStatus:True; PodName:pb-horizontal-2              115s   KubeDB Ops-manager Operator  patch p s; ConditionStatus:True; PodName:pb-horizontal-2
  Warning  is pg bouncer running; ConditionStatus:True; PodName:pb-horizontal-2  110s   KubeDB Ops-manager Operator  is pg bouncer running; ConditionStatus:True; PodName:pb-horizontal-2
  Normal   Successful                                                            105s   KubeDB Ops-manager Operator  Horizontal scaling Up performed successfully in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-up
  Normal   Starting                                                              95s    KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-horizontal
  Normal   Successful                                                            95s    KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-horizontal
  Normal   Successful                                                            95s    KubeDB Ops-manager Operator  Controller has Successfully scaled the PgBouncer database: demo/pb-horizontal
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/scaling/horizontal-scaling-down-ops.yaml
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
  Creation Timestamp:  2024-11-27T11:16:05Z
  Generation:          1
  Resource Version:    49481
  UID:                 cf4bc042-8316-4dce-b6a2-60981af7f4db
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pb-horizontal
  Horizontal Scaling:
    Replicas:  2
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-11-27T11:16:05Z
    Message:               Controller has started to Progress with HorizontalScaling of PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-down
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2024-11-27T11:16:08Z
    Message:               Horizontal scaling started in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-down
    Observed Generation:   1
    Reason:                HorizontalScaleStarted
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-11-27T11:16:13Z
    Message:               patch p s; ConditionStatus:True; PodName:pb-horizontal-3
    Observed Generation:   1
    Status:                True
    Type:                  PatchPS--pb-horizontal-3
    Last Transition Time:  2024-11-27T11:16:18Z
    Message:               get pod; ConditionStatus:True; PodName:pb-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-horizontal-2
    Last Transition Time:  2024-11-27T11:16:23Z
    Message:               Horizontal scaling down performed successfully in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-down
    Observed Generation:   1
    Reason:                HorizontalScaleSucceeded
    Status:                True
    Type:                  HorizontalScaleDown
    Last Transition Time:  2024-11-27T11:16:33Z
    Message:               Controller has successfully completed  with HorizontalScaling of PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-down
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  2m38s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pgbouncer-horizontal-scale-down
  Normal   Starting                                                  2m38s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-horizontal
  Normal   Successful                                                2m38s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-down
  Normal   Starting                                                  2m35s  KubeDB Ops-manager Operator  Horizontal scaling started in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-down
  Warning  patch p s; ConditionStatus:True; PodName:pb-horizontal-3  2m30s  KubeDB Ops-manager Operator  patch p s; ConditionStatus:True; PodName:pb-horizontal-3
  Warning  get pod; ConditionStatus:True; PodName:pb-horizontal-2    2m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-horizontal-2
  Normal   Successful                                                2m20s  KubeDB Ops-manager Operator  Horizontal scaling down performed successfully in PgBouncer: demo/pb-horizontal for PgBouncerOpsRequest: pgbouncer-horizontal-scale-down
  Normal   Starting                                                  2m10s  KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-horizontal
  Normal   Successful                                                2m10s  KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-horizontal
  Normal   Successful                                                2m10s  KubeDB Ops-manager Operator  Controller has Successfully scaled the PgBouncer database: demo/pb-horizontal
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