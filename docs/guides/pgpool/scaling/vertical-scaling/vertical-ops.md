---
title: Vertical Scaling Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-vertical-scaling-ops
    name: VerticalScaling OpsRequest
    parent: pp-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Pgpool

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Pgpool.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Pgpool](/docs/guides/pgpool/concepts/pgpool.md)
  - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/pgpool/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgpool](/docs/examples/pgpool) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Pgpool

Here, we are going to deploy a  `Pgpool` using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare Pgpool

Now, we are going to deploy a `Pgpool` with version `4.5.0`.

### Deploy Pgpool 

In this section, we are going to deploy a Pgpool. Then, in the next section we will update the resources using `PgpoolOpsRequest` CRD. Below is the YAML of the `Pgpool` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-vertical
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  deletionPolicy: WipeOut
```

Let's create the `Pgpool` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/scaling/pp-vertical.yaml
pgpool.kubedb.com/pp-vertical created
```

Now, wait until `pp-vertical` has status `Ready`. i.e,

```bash
$ kubectl get pp -n demo
NAME          TYPE                  VERSION   STATUS   AGE
pp-vertical   kubedb.com/v1alpha2   4.5.0     Ready    17s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo pp-vertical-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see the Pod has default resources which is assigned by the KubeDB operator.

We are now ready to apply the `PgpoolOpsRequest` CR to update the resources of this pgpool.

### Vertical Scaling

Here, we are going to update the resources of the pgpool to meet the desired resources after scaling.

#### Create PgpoolOpsRequest

In order to update the resources of the pgpool, we have to create a `PgpoolOpsRequest` CR with our desired resources. Below is the YAML of the `PgpoolOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgpoolOpsRequest
metadata:
  name: pgpool-scale-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: pp-vertical
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `pp-vertical` pgpool.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.standalone` specifies the desired resources after scaling.
- Have a look [here](/docs/guides/pgpool/concepts/opsrequest.md) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `PgpoolOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/scaling/vertical-scaling/pp-vertical-ops.yaml
pgpoolopsrequest.ops.kubedb.com/pgpool-scale-vertical created
```

#### Verify Pgpool resources updated successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Pgpool` object and related `PetSet` and `Pods`.

Let's wait for `PgpoolOpsRequest` to be `Successful`.  Run the following command to watch `PgpoolOpsRequest` CR,

```bash
$ kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                    TYPE              STATUS       AGE
pgpool-scale-vertical   VerticalScaling   Successful   3m42s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed to scale the pgpool.

```bash
$ kubectl describe pgpoolopsrequest -n demo pgpool-scale-vertical
Name:         pgpool-scale-vertical
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T09:44:22Z
  Generation:          1
  Resource Version:    68270
  UID:                 62a105f7-e7b9-444e-9303-79818fccfdef
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   pp-vertical
  Timeout:  5m
  Type:     VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2Gi
        Requests:
          Cpu:     1
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-07-17T09:44:22Z
    Message:               Pgpool ops-request has started to vertically scaling the Pgpool nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-07-17T09:44:25Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-17T09:44:25Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-17T09:45:10Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-17T09:44:30Z
    Message:               get pod; ConditionStatus:True; PodName:pp-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pp-vertical-0
    Last Transition Time:  2024-07-17T09:44:30Z
    Message:               evict pod; ConditionStatus:True; PodName:pp-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pp-vertical-0
    Last Transition Time:  2024-07-17T09:45:05Z
    Message:               check pod running; ConditionStatus:True; PodName:pp-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pp-vertical-0
    Last Transition Time:  2024-07-17T09:45:10Z
    Message:               Successfully completed the vertical scaling for Pgpool
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                           Age    From                         Message
  ----     ------                                                           ----   ----                         -------
  Normal   Starting                                                         4m16s  KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/pgpool-scale-vertical
  Normal   Starting                                                         4m16s  KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pp-vertical
  Normal   Successful                                                       4m16s  KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pp-vertical for PgpoolOpsRequest: pgpool-scale-vertical
  Normal   UpdatePetSets                                                    4m13s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:pp-vertical-0             4m8s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pp-vertical-0
  Warning  evict pod; ConditionStatus:True; PodName:pp-vertical-0           4m8s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pp-vertical-0
  Warning  check pod running; ConditionStatus:False; PodName:pp-vertical-0  4m3s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pp-vertical-0
  Warning  check pod running; ConditionStatus:True; PodName:pp-vertical-0   3m33s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pp-vertical-0
  Normal   RestartPods                                                      3m28s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                         3m28s  KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pp-vertical
  Normal   Successful                                                       3m28s  KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pp-vertical for PgpoolOpsRequest: pgpool-scale-vertical
```

Now, we are going to verify from the Pod yaml whether the resources of the pgpool has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo pp-vertical-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the Pgpool.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pp -n demo pp-vertical
kubectl delete pgpoolopsrequest -n demo pgpool-scale-vertical
```