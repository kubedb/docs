---
title: Horizontal Scaling RabbitMQ
menu:
  docs_{{ .version }}:
    identifier: rm-horizontal-scaling-ops
    name: Scale Horizontally
    parent: rm-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale RabbitMQ

This guide will show you how to use `KubeDB` Ops-manager operator to scale the RabbitMQ Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/rabbitmq/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/rabbitmq](/docs/examples/rabbitmq) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on rabbitmq

Here, we are going to deploy a `RabbitMQ` using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Deploy RabbitMQ 

In this section, we are going to deploy a RabbitMQ. We are going to deploy a `RabbitMQ` with version `3.13.2`. Then, in the next section we will scale the rabbitmq using `RabbitMQOpsRequest` CRD. Below is the YAML of the `RabbitMQ` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq
  namespace: demo
spec:
  version: "3.13.2"
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
Let's create the `RabbitMQ` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/scaling/rabbitmq-cluster.yaml
rabbitmq.kubedb.com/rabbitmq created
```

Now, wait until `rabbitmq` has status `Ready`. i.e,

```bash
$ kubectl get rm -n demo
NAME            TYPE                  VERSION   STATUS   AGE
rabbitmq        kubedb.com/v1alpha2   3.13.2     Ready    2m
```

Let's check the number of replicas this rabbitmq has from the RabbitMQ object, number of pods the PetSet have,

```bash
$ kubectl get rabbitmq -n demo rabbitmq -o json | jq '.spec.replicas'
1

$ kubectl get petset -n demo rabbitmq -o json | jq '.spec.replicas'
1
```

We can see from both command that the rabbitmq has 3 replicas. 

We are now ready to apply the `RabbitMQOpsRequest` CR to scale this rabbitmq.

## Scale Up Replicas

Here, we are going to scale up the replicas of the rabbitmq to meet the desired number of replicas after scaling.

#### Create RabbitMQOpsRequest

In order to scale up the replicas of the rabbitmq, we have to create a `RabbitMQOpsRequest` CR with our desired replicas. Below is the YAML of the `RabbitMQOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rabbitmq-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: rabbitmq
  horizontalScaling:
    node: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `pp-horizontal` rabbitmq.
- `spec.type` specifies that we are performing `HorizontalScaling` on our rabbitmq.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/scaling/horizontal-scaling/rm-hscale-up-ops.yaml
rabbitmqopsrequest.ops.kubedb.com/rabbitmq-horizontal-scale-up created
```

#### Verify replicas scaled up successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `RabbitMQ` object and related `PetSet`.

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CR,

```bash
$ watch kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME                           TYPE                STATUS       AGE
rabbitmq-horizontal-scale-up   HorizontalScaling   Successful   2m49s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed to scale the rabbitmq.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rabbitmq-horizontal-scale-up
Name:         rabbitmq-horizontal-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
Metadata:
  Creation Timestamp:  2024-09-12T10:48:21Z
  Generation:          1
  Resource Version:    46348
  UID:                 eaa4653f-f07a-47d1-935f-5a82d64ea659
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  rabbitmq
  Horizontal Scaling:
    Node:  5
  Type:    HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-09-12T11:42:06Z
    Message:               RabbitMQ ops-request has started to horizontally scaling the nodes
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
    Message:               rabbitmq already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-09-12T11:43:04Z
    Message:               successfully reconciled the RabbitMQ with modified node
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-09-12T11:43:04Z
    Message:               Successfully completed horizontally scale RabbitMQ cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                     Age    From                         Message
  ----     ------                                     ----   ----                         -------
  Normal   Starting                                   8m40s  KubeDB Ops-manager Operator  Start processing for RabbitMQOpsRequest: demo/rabbitmq-horizontal-scale-up
  Normal   Starting                                   8m40s  KubeDB Ops-manager Operator  Pausing RabbitMQ databse: demo/rabbitmq
  Normal   Successful                                 8m40s  KubeDB Ops-manager Operator  Successfully paused RabbitMQ database: demo/rabbitmq for RabbitMQOpsRequest: rabbitmq-horizontal-scale-up
  Warning  patch petset; ConditionStatus:True         8m32s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True
  Warning  client failure; ConditionStatus:True       8m27s  KubeDB Ops-manager Operator  client failure; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:False  8m27s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:False
  Warning  is node in cluster; ConditionStatus:True   7m52s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScaleUp                          7m47s  KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdatePetSets                              7m42s  KubeDB Ops-manager Operator  successfully reconciled the RabbitMQ with modified node
  Normal   Starting                                   7m42s  KubeDB Ops-manager Operator  Resuming RabbitMQ database: demo/rabbitmq
  Normal   Successful                                 7m42s  KubeDB Ops-manager Operator  Successfully resumed RabbitMQ database: demo/rabbitmq for RabbitMQOpsRequest: rabbitmq-horizontal-scale-up
```

Now, we are going to verify the number of replicas this rabbitmq has from the Pgpool object, number of pods the PetSet have,

```bash
$ kubectl get rm -n demo rabbitmq -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo rabbitmq -o json | jq '.spec.replicas'
3
```
From all the above outputs we can see that the replicas of the rabbitmq is `3`. That means we have successfully scaled up the replicas of the RabbitMQ.


### Scale Down Replicas

Here, we are going to scale down the replicas of the rabbitmq to meet the desired number of replicas after scaling.

#### Create RabbitMQOpsRequest

In order to scale down the replicas of the rabbitmq, we have to create a `RabbitMQOpsRequest` CR with our desired replicas. Below is the YAML of the `RabbitMQOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RabbitMQOpsRequest
metadata:
  name: rabbitmq-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: rabbitmq
  horizontalScaling:
    node: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `rabbitmq` rabbitmq.
- `spec.type` specifies that we are performing `HorizontalScaling` on our rabbitmq.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `RabbitMQOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/scaling/horizontal-scaling/rmops-hscale-down-ops.yaml
rabbitmqopsrequest.ops.kubedb.com/rabbitmq-horizontal-scale-down created
```

#### Verify replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `RabbitMQ` object and related `PetSet`.

Let's wait for `RabbitMQOpsRequest` to be `Successful`.  Run the following command to watch `RabbitMQOpsRequest` CR,

```bash
$ watch kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME                           TYPE                STATUS       AGE
rabbitmq-horizontal-scale-down   HorizontalScaling   Successful   75s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed to scale the rabbitmq.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rabbitmq-horizontal-scale-down
Name:         rabbitmq-horizontal-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T08:52:28Z
  Generation:          1
  Resource Version:    63600
  UID:                 019f9d8f-c2b0-4154-b3d3-b715b8805fd7
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  rabbitmq
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
    Message:               Successfully completed horizontally scale rabbitmq cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/rabbitmq-horizontal-scale-down
  Normal   Starting                                                     96s   KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/rabbitmq
  Normal   Successful                                                   96s   KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/rabbitmq for PgpoolOpsRequest: rabbitmq-horizontal-scale-down
  Normal   patch petset; ConditionStatus:True; PodName:pp-horizontal-2  88s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   get pod; ConditionStatus:False                               83s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Normal   get pod; ConditionStatus:True; PodName:pp-horizontal-2       53s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pp-horizontal-2
  Normal   HorizontalScaleDown                                          48s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdateDatabase                                               48s   KubeDB Ops-manager Operator  Successfully updated RabbitMQ
  Normal   Starting                                                     48s   KubeDB Ops-manager Operator  Resuming Pgpool database: demo/rabbitmq
  Normal   Successful                                                   48s   KubeDB Ops-manager Operator  Successfully resumed RabbitMQ database: demo/rabbitmq for RabbitMQOpsRequest: rabbitmq-horizontal-scale-down
```

Now, we are going to verify the number of replicas this rabbitmq has from the RabbitMQ object, number of pods the petset have,

```bash
$ kubectl get rm -n demo rabbitmq -o json | jq '.spec.replicas'
2

$ kubectl get petset -n demo rabbitmq -o json | jq '.spec.replicas'
2
```
From all the above outputs we can see that the replicas of the rabbitmq is `2`. That means we have successfully scaled up the replicas of the RabbitMQ.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rm -n demo
kubectl delete rabbitmqopsrequest -n demo rabbitmq-horizontal-scale-down
```