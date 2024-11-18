---
title: Horizontal Scaling ZooKeeper
menu:
  docs_{{ .version }}:
    identifier: zk-horizontal-scaling-ops
    name: Scale Horizontally
    parent: zk-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale ZooKeeper

This guide will show you how to use `KubeDB` Ops-manager operator to scale the ZooKeeper Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
    - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)
    - [Horizontal Scaling Overview](/docs/guides/zookeeper/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/zookeeper](/docs/examples/zookeeper) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on zookeeper

Here, we are going to deploy a `ZooKeeper` using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Deploy ZooKeeper

In this section, we are going to deploy a ZooKeeper. We are going to deploy a `ZooKeeper` with version `3.8.3`. Then, in the next section we will scale the zookeeper using `ZooKeeperOpsRequest` CRD. Below is the YAML of the `ZooKeeper` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-quickstart
  namespace: demo
spec:
  version: "3.8.3"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```
Let's create the `ZooKeeper` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/scaling/zookeeper.yaml
zookeeper.kubedb.com/zk-quickstart created
```

Now, wait until `zk-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get zk -n demo
NAME            VERSION    STATUS    AGE
zk-quickstart   3.8.3      Ready     5m56s
```

Let's check the number of replicas this zookeeper has from the ZooKeeper object, number of pods the PetSet have,

```bash
$ kubectl get zookeeper -n demo zk-quickstart -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo zk-quickstart -o json | jq '.spec.replicas'
3
```

We can see from both command that the zookeeper has 3 replicas.

We are now ready to apply the `ZooKeeperOpsRequest` CR to scale this zookeeper.

## Scale Up Replicas

Here, we are going to scale up the replicas of the zookeeper to meet the desired number of replicas after scaling.

#### Create ZooKeeperOpsRequest

In order to scale up the replicas of the zookeeper, we have to create a `ZooKeeperOpsRequest` CR with our desired replicas. Below is the YAML of the `ZooKeeperOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zookeeper-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: zk-quickstart
  horizontalScaling:
    replicas: 5
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `zk-quickstart` zookeeper.
- `spec.type` specifies that we are performing `HorizontalScaling` on our zookeeper.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/scaling/horizontal-scaling/zk-hscale-up-ops.yaml
zookeeperopsrequest.ops.kubedb.com/zookeeper-horizontal-scale-up created
```

#### Verify replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `ZooKeeper` object and related `PetSet`.

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CR,

```bash
$ watch kubectl get zookeeperopsrequest -n demo
NAME                            TYPE                STATUS       AGE
zookeeper-horizontal-scale-up   HorizontalScaling   Successful   2m49s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed to scale the zookeeper.

```bash
$ kubectl describe zookeeperopsrequest -n demo zookeeper-horizontal-scale-up
Name:         zookeeper-horizontal-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-10-25T13:37:43Z
  Generation:          1
  Resource Version:    1198117
  UID:                 bfa6fb3f-5eb2-456c-8a3e-7a59097add0a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  zk-quickstart
  Horizontal Scaling:
    Replicas:  5
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-10-25T13:37:43Z
    Message:               ZooKeeper ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-10-25T13:38:03Z
    Message:               Successfully Scaled Up Node
    Observed Generation:   1
    Reason:                HorizontalScaleUp
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2024-10-25T13:37:48Z
    Message:               patch petset; ConditionStatus:True; PodName:zk-quickstart-4
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--zk-quickstart-4
    Last Transition Time:  2024-10-25T13:37:48Z
    Message:               zk-quickstart already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-10-25T13:37:58Z
    Message:               is pod ready; ConditionStatus:True; PodName:zk-quickstart-4
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--zk-quickstart-4
    Last Transition Time:  2024-10-25T13:37:58Z
    Message:               is node healthy; ConditionStatus:True; PodName:zk-quickstart-4
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeHealthy--zk-quickstart-4
    Last Transition Time:  2024-10-25T13:38:03Z
    Message:               Successfully updated ZooKeeper
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-10-25T13:38:03Z
    Message:               Successfully completed the HorizontalScaling for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age   From                         Message
  ----     ------                                                          ----  ----                         -------
  Normal   Starting                                                        47s   KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/horizontal-scale-up
  Warning  patch petset; ConditionStatus:True; PodName:zk-quickstart-4     42s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:zk-quickstart-4
  Warning  is pod ready; ConditionStatus:False; PodName:zk-quickstart-4    37s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:False; PodName:zk-quickstart-4
  Warning  is pod ready; ConditionStatus:True; PodName:zk-quickstart-4     32s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:zk-quickstart-4
  Warning  is node healthy; ConditionStatus:True; PodName:zk-quickstart-4  32s   KubeDB Ops-manager Operator  is node healthy; ConditionStatus:True; PodName:zk-quickstart-4
  Normal   HorizontalScaleUp                                               27s   KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdateDatabase                                                  27s   KubeDB Ops-manager Operator  Successfully updated ZooKeeper
  Normal   Starting                                                        27s   KubeDB Ops-manager Operator  Resuming ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                      27s   KubeDB Ops-manager Operator  Successfully resumed ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: horizontal-scale-up
```

Now, we are going to verify the number of replicas this zookeeper has from the Pgpool object, number of pods the PetSet have,

```bash
$ kubectl get zookeeper -n demo zk-quickstart -o json | jq '.spec.replicas'
5

$ kubectl get petset -n demo zk-quickstart -o json | jq '.spec.replicas'
5
```
From all the above outputs we can see that the replicas of the zookeeper is `5`. That means we have successfully scaled up the replicas of the ZooKeeper.


### Scale Down Replicas

Here, we are going to scale down the replicas of the zookeeper to meet the desired number of replicas after scaling.

#### Create ZooKeeperOpsRequest

In order to scale down the replicas of the zookeeper, we have to create a `ZooKeeperOpsRequest` CR with our desired replicas. Below is the YAML of the `ZooKeeperOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zookeeper-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: zk-quickstart
  horizontalScaling:
    replicas: 3

```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `zookeeper` zookeeper.
- `spec.type` specifies that we are performing `HorizontalScaling` on our zookeeper.
- `spec.horizontalScaling.replicas` specifies the desired replicas after scaling.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/scaling/horizontal-scaling/zk-hscale-down-ops.yaml
zookeeperopsrequest.ops.kubedb.com/zookeeper-horizontal-scale-down created
```

#### Verify replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `ZooKeeper` object and related `PetSet`.

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CR,

```bash
$ watch kubectl get zookeeperopsrequest -n demo
NAME                              TYPE                STATUS       AGE
zookeeper-horizontal-scale-down   HorizontalScaling   Successful   75s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed to scale the zookeeper.

```bash
$ kubectl describe zookeeperopsrequest -n demo zookeeper-horizontal-scale-down
Name:         zookeeper-horizontal-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-10-25T13:58:45Z
  Generation:          1
  Resource Version:    1199568
  UID:                 18b2adbb-9fbd-44fe-a265-e7eb4a292798
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  zk-quickstart
  Horizontal Scaling:
    Replicas:  3
  Type:        HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2024-10-25T13:58:45Z
    Message:               ZooKeeper ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2024-10-25T14:00:23Z
    Message:               Successfully Scaled Down Node
    Observed Generation:   1
    Reason:                HorizontalScaleDown
    Status:                True
    Type:                  HorizontalScaleDown
    Last Transition Time:  2024-10-25T13:58:53Z
    Message:               patch petset; ConditionStatus:True; PodName:zk-quickstart-4
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--zk-quickstart-4
    Last Transition Time:  2024-10-25T13:58:58Z
    Message:               get pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  GetPod
    Last Transition Time:  2024-10-25T13:59:28Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-4
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-4
    Last Transition Time:  2024-10-25T13:59:28Z
    Message:               delete pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePvc
    Last Transition Time:  2024-10-25T13:59:38Z
    Message:               patch petset; ConditionStatus:True; PodName:zk-quickstart-3
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--zk-quickstart-3
    Last Transition Time:  2024-10-25T13:59:38Z
    Message:               zk-quickstart already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2024-10-25T14:00:13Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-3
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-3
    Last Transition Time:  2024-10-25T14:00:23Z
    Message:               Successfully updated ZooKeeper
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-10-25T14:00:23Z
    Message:               Successfully completed the HorizontalScaling for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age    From                         Message
  ----     ------                                                       ----   ----                         -------
  Normal   Starting                                                     3m27s  KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/horizontal-scale-down
  Normal   Starting                                                     3m27s  KubeDB Ops-manager Operator  Pausing ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                   3m27s  KubeDB Ops-manager Operator  Successfully paused ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: horizontal-scale-down
  Warning  patch petset; ConditionStatus:True; PodName:zk-quickstart-4  3m19s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:zk-quickstart-4
  Warning  get pod; ConditionStatus:False                               3m14s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-4       2m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-4
  Warning  delete pvc; ConditionStatus:True                             2m44s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:False                               2m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-4       2m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-4
  Warning  delete pvc; ConditionStatus:True                             2m39s  KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-4       2m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-4
  Warning  patch petset; ConditionStatus:True; PodName:zk-quickstart-3  2m34s  KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:zk-quickstart-3
  Warning  get pod; ConditionStatus:False                               2m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-3       119s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-3
  Warning  delete pvc; ConditionStatus:True                             119s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:False                               119s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-3       114s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-3
  Warning  delete pvc; ConditionStatus:True                             114s   KubeDB Ops-manager Operator  delete pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-3       114s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-3
  Normal   HorizontalScaleDown                                          109s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdateDatabase                                               109s   KubeDB Ops-manager Operator  Successfully updated ZooKeeper
  Normal   Starting                                                     109s   KubeDB Ops-manager Operator  Resuming ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                   109s   KubeDB Ops-manager Operator  Successfully resumed ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: horizontal-scale-down
```

Now, we are going to verify the number of replicas this zookeeper has from the ZooKeeper object, number of pods the petset have,

```bash
$ kubectl get zookeeper -n demo zk-quickstart -o json | jq '.spec.replicas'
3

$ kubectl get petset -n demo zk-quickstart -o json | jq '.spec.replicas'
3
```
From all the above outputs we can see that the replicas of the zookeeper is `3`. That means we have successfully scaled up the replicas of the ZooKeeper.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete zk -n demo
kubectl delete zookeeperopsrequest -n demo zookeeper-horizontal-scale-down
```