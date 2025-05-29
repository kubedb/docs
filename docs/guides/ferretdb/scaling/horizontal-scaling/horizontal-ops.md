---
title: Horizontal Scaling FerretDB
menu:
  docs_{{ .version }}:
    identifier: fr-horizontal-scaling-ops
    name: HorizontalScaling OpsRequest
    parent: fr-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale FerretDB

This guide will show you how to use `KubeDB` Ops-manager operator to scale the replicaset of a FerretDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md)
    - [FerretDBOpsRequest](/docs/guides/ferretdb/concepts/opsrequest.md)
    - [Horizontal Scaling Overview](/docs/guides/ferretdb/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/ferretdb](/docs/examples/ferretdb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Horizontal Scaling on ferretdb

Here, we are going to deploy a  `FerretDB` using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare FerretDB

Now, we are going to deploy a `FerretDB` with version `1.23.0`.

### Deploy FerretDB

In this section, we are going to deploy a FerretDB. Then, in the next section we will scale the ferretdb using `FerretDBOpsRequest` CRD. Below is the YAML of the `FerretDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: fr-horizontal
  namespace: demo
spec:
  version: "2.0.0"
  backend:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 500Mi
  deletionPolicy: WipeOut
  server:
    primary:
      replicas: 1
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/scaling/fr-horizontal.yaml
ferretdb.kubedb.com/fr-horizontal created
```

Now, wait until `fr-horizontal ` has status `Ready`. i.e,

```bash
$ kubectl get fr -n demo
NAME            NAMESPACE   VERSION   STATUS   AGE
fr-horizontal   demo        2.0.0     Ready    86s
```

Let's check the number of replicas this ferretdb has from the FerretDB object, number of pods the petset have,

```bash
$ kubectl get fr -n demo fr-horizontal -o json | jq '.spec.server.primary.replicas'
1

$ kubectl get petset -n demo fr-horizontal -o json | jq '.spec.replicas'
1
```

We can see from both command that the ferretdb primary server has 1 replicas.

We are now ready to apply the `FerretDBOpsRequest` CR to scale this ferretdb.

## Scale Up Replicas

Here, we are going to scale up the replicas of the ferretdb to meet the desired number of replicas after scaling.

#### Create FerretDBOpsRequest

In order to scale up the replicas of the ferretdb primary server, we have to create a `FerretDBOpsRequest` CR with our desired replicas. Below is the YAML of the `FerretDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: ferretdb-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: fr-horizontal
  horizontalScaling:
    primary:
      replicas: 3
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `fr-horizontal` ferretdb.
- `spec.type` specifies that we are performing `HorizontalScaling` on our ferretdb.
- `spec.horizontalScaling.primary.replicas` specifies the desired primary server replicas after scaling.

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/scaling/horizontal-scaling/frops-hscale-up-ops.yaml
ferretdbopsrequest.ops.kubedb.com/ferretdb-horizontal-scale-up created
```

#### Verify replicas scaled up successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `FerretDB` object and related `PetSet`.

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CR,

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME                           TYPE                STATUS       AGE
ferretdb-horizontal-scale-up   HorizontalScaling   Successful   102s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed to scale the ferretdb.

```bash
$ kubectl describe ferretdbopsrequest -n demo ferretdb-horizontal-scale-up
Name:         ferretdb-horizontal-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2025-04-08T06:23:50Z
  Generation:          1
  Resource Version:    64728
  UID:                 ba4b687c-7127-46a6-bf01-9c5d5d916555
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  fr-horizontal
  Horizontal Scaling:
    Primary:
      Replicas:  3
  Type:          HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-04-08T06:23:51Z
    Message:               FerretDB ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2025-04-08T06:23:54Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-04-08T06:24:19Z
    Message:               Successfully Scaled Up Node
    Observed Generation:   1
    Reason:                HorizontalScaleUp
    Status:                True
    Type:                  HorizontalScaleUp
    Last Transition Time:  2025-04-08T06:23:59Z
    Message:               patch petset; ConditionStatus:True; PodName:fr-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--fr-horizontal-1
    Last Transition Time:  2025-04-08T06:24:04Z
    Message:               is pod ready; ConditionStatus:True; PodName:fr-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--fr-horizontal-1
    Last Transition Time:  2025-04-08T06:24:04Z
    Message:               client failure; ConditionStatus:True; PodName:fr-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure--fr-horizontal-1
    Last Transition Time:  2025-04-08T06:24:04Z
    Message:               is node healthy; ConditionStatus:True; PodName:fr-horizontal-1
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeHealthy--fr-horizontal-1
    Last Transition Time:  2025-04-08T06:24:09Z
    Message:               patch petset; ConditionStatus:True; PodName:fr-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--fr-horizontal-2
    Last Transition Time:  2025-04-08T06:24:09Z
    Message:               fr-horizontal already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-04-08T06:24:14Z
    Message:               is pod ready; ConditionStatus:True; PodName:fr-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--fr-horizontal-2
    Last Transition Time:  2025-04-08T06:24:14Z
    Message:               client failure; ConditionStatus:True; PodName:fr-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  ClientFailure--fr-horizontal-2
    Last Transition Time:  2025-04-08T06:24:14Z
    Message:               is node healthy; ConditionStatus:True; PodName:fr-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeHealthy--fr-horizontal-2
    Last Transition Time:  2025-04-08T06:24:19Z
    Message:               Successfully updated FerretDB
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-04-08T06:24:19Z
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
  Normal   Starting                                                        43s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/ferretdb-horizontal-scale-up
  Normal   Starting                                                        43s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/fr-horizontal
  Normal   Successful                                                      43s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/fr-horizontal for FerretDBOpsRequest: ferretdb-horizontal-scale-up
  Warning  patch petset; ConditionStatus:True; PodName:fr-horizontal-1     35s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:fr-horizontal-1
  Warning  is pod ready; ConditionStatus:True; PodName:fr-horizontal-1     30s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:fr-horizontal-1
  Warning  client failure; ConditionStatus:True; PodName:fr-horizontal-1   30s   KubeDB Ops-manager Operator  client failure; ConditionStatus:True; PodName:fr-horizontal-1
  Warning  is node healthy; ConditionStatus:True; PodName:fr-horizontal-1  30s   KubeDB Ops-manager Operator  is node healthy; ConditionStatus:True; PodName:fr-horizontal-1
  Warning  patch petset; ConditionStatus:True; PodName:fr-horizontal-2     25s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:fr-horizontal-2
  Warning  is pod ready; ConditionStatus:True; PodName:fr-horizontal-2     20s   KubeDB Ops-manager Operator  is pod ready; ConditionStatus:True; PodName:fr-horizontal-2
  Warning  client failure; ConditionStatus:True; PodName:fr-horizontal-2   20s   KubeDB Ops-manager Operator  client failure; ConditionStatus:True; PodName:fr-horizontal-2
  Warning  is node healthy; ConditionStatus:True; PodName:fr-horizontal-2  20s   KubeDB Ops-manager Operator  is node healthy; ConditionStatus:True; PodName:fr-horizontal-2
  Normal   HorizontalScaleUp                                               15s   KubeDB Ops-manager Operator  Successfully Scaled Up Node
  Normal   UpdateDatabase                                                  15s   KubeDB Ops-manager Operator  Successfully updated FerretDB
  Normal   Starting                                                        15s   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/fr-horizontal
  Normal   Successful                                                      15s   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/fr-horizontal for FerretDBOpsRequest: ferretdb-horizontal-scale-up
```

Now, we are going to verify the number of replicas this ferretdb has from the FerretDB object, number of pods the petset have,

```bash
$ kubectl get fr -n demo fr-horizontal -o json | jq '.spec.server.primary.replicas'
3

$ kubectl get petset -n demo fr-horizontal -o json | jq '.spec.replicas'
3
```
From all the above outputs we can see that the replicas of the ferretdb primary server is `3`. That means we have successfully scaled up the replicas of the FerretDB.


### Scale Down Replicas

Here, we are going to scale down the replicas of the ferretdb primary server to meet the desired number of replicas after scaling.

#### Create FerretDBOpsRequest

In order to scale down the replicas of the ferretdb, we have to create a `FerretDBOpsRequest` CR with our desired replicas. Below is the YAML of the `FerretDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: ferretdb-horizontal-scale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: fr-horizontal
  horizontalScaling:
    primary:
      replicas: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling down operation on `fr-horizontal` ferretdb.
- `spec.type` specifies that we are performing `HorizontalScaling` on our ferretdb.
- `spec.horizontalScaling.primary.replicas` specifies the desired primary server replicas after scaling.

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/scaling/horizontal-scaling/frops-hscale-down-ops.yaml
ferretdbopsrequest.ops.kubedb.com/ferretdb-horizontal-scale-down created
```

#### Verify replicas scaled down successfully

If everything goes well, `KubeDB` Ops-manager operator will update the replicas of `FerretDB` object and related `PetSet`.

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CR,

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME                             TYPE                STATUS       AGE
ferretdb-horizontal-scale-down   HorizontalScaling   Successful   40s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed to scale the ferretdb.

```bash
$ kubectl describe ferretdbopsrequest -n demo ferretdb-horizontal-scale-down
Name:         ferretdb-horizontal-scale-down
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2025-04-08T06:26:09Z
  Generation:          1
  Resource Version:    64829
  UID:                 369e5413-9fa3-47de-b491-f7f7568a45c1
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  fr-horizontal
  Horizontal Scaling:
    Primary:
      Replicas:  2
  Type:          HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-04-08T06:26:09Z
    Message:               FerretDB ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2025-04-08T06:26:12Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2025-04-08T06:26:27Z
    Message:               Successfully Scaled Down Node
    Observed Generation:   1
    Reason:                HorizontalScaleDown
    Status:                True
    Type:                  HorizontalScaleDown
    Last Transition Time:  2025-04-08T06:26:17Z
    Message:               patch petset; ConditionStatus:True; PodName:fr-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  PatchPetset--fr-horizontal-2
    Last Transition Time:  2025-04-08T06:26:17Z
    Message:               fr-horizontal already has desired replicas
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-04-08T06:26:22Z
    Message:               get pod; ConditionStatus:True; PodName:fr-horizontal-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--fr-horizontal-2
    Last Transition Time:  2025-04-08T06:26:27Z
    Message:               Successfully updated FerretDB
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2025-04-08T06:26:27Z
    Message:               Successfully completed the HorizontalScaling for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                       Age   From                         Message
  ----     ------                                                       ----  ----                         -------
  Normal   Starting                                                     43s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/ferretdb-horizontal-scale-down
  Normal   Starting                                                     43s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/fr-horizontal
  Normal   Successful                                                   43s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/fr-horizontal for FerretDBOpsRequest: ferretdb-horizontal-scale-down
  Warning  patch petset; ConditionStatus:True; PodName:fr-horizontal-2  35s   KubeDB Ops-manager Operator  patch petset; ConditionStatus:True; PodName:fr-horizontal-2
  Warning  get pod; ConditionStatus:True; PodName:fr-horizontal-2       30s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:fr-horizontal-2
  Normal   HorizontalScaleDown                                          25s   KubeDB Ops-manager Operator  Successfully Scaled Down Node
  Normal   UpdateDatabase                                               25s   KubeDB Ops-manager Operator  Successfully updated FerretDB
  Normal   Starting                                                     25s   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/fr-horizontal
  Normal   Successful                                                   25s   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/fr-horizontal for FerretDBOpsRequest: ferretdb-horizontal-scale-down
```

Now, we are going to verify the number of replicas this ferretdb has from the FerretDB object, number of pods the petset have,

```bash
$ kubectl get fr -n demo fr-horizontal -o json | jq '.spec.server.primary.replicas'
2

$ kubectl get petset -n demo fr-horizontal -o json | jq '.spec.replicas'
2
```
From all the above outputs we can see that the replicas of the ferretdb is `2`. That means we have successfully scaled up the replicas of the FerretDB.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n fr-horizontal
kubectl delete ferretdbopsrequest -n demo ferretdb-horizontal-scale-down
```