---
title: Vertical Scaling PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-vertical-scaling-ops
    name: VerticalScaling OpsRequest
    parent: pb-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale PgBouncer

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a PgBouncer.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/pgbouncer/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgbouncer](/docs/examples/pgbouncer) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on PgBouncer

Here, we are going to deploy a `PgBouncer` using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

### Prepare PgBouncer

Now, we are going to deploy a `PgBouncer` with version `1.18.0`.

### Deploy PgBouncer 

In this section, we are going to deploy a PgBouncer. Then, in the next section we will update the resources using `PgBouncerOpsRequest` CRD. Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pb-vertical
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/scaling/pb-vertical.yaml
pgbouncer.kubedb.com/pb-vertical created
```

Now, wait until `pb-vertical` has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
NAME          TYPE                  VERSION   STATUS   AGE
pb-vertical   kubedb.com/v1         1.18.0    Ready    17s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo pb-vertical-0 -o json | jq '.spec.containers[].resources'
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

We are now ready to apply the `PgBouncerOpsRequest` CR to update the resources of this pgbouncer.

### Vertical Scaling

Here, we are going to update the resources of the pgbouncer to meet the desired resources after scaling.

#### Create PgBouncerOpsRequest

In order to update the resources of the pgbouncer, we have to create a `PgBouncerOpsRequest` CR with our desired resources. Below is the YAML of the `PgBouncerOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-scale-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: pb-vertical
  verticalScaling:
    pgbouncer:
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

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `pb-vertical` pgbouncer.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.pgbouncer` specifies the desired resources after scaling.
- Have a look [here](/docs/guides/pgbouncer/concepts/opsrequest.md) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `PgBouncerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/scaling/vertical-scaling/pb-vertical-ops.yaml
pgbounceropsrequest.ops.kubedb.com/pgbouncer-scale-vertical created
```

#### Verify PgBouncer resources updated successfully 

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `PgBouncer` object and related `PetSet` and `Pods`.

Let's wait for `PgBouncerOpsRequest` to be `Successful`.  Run the following command to watch `PgBouncerOpsRequest` CR,

```bash
$ kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                       TYPE              STATUS       AGE
pgbouncer-scale-vertical   VerticalScaling   Successful   3m42s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to scale the pgbouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pgbouncer-scale-vertical
Name:         pgbouncer-scale-vertical
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T09:44:22Z
  Generation:          1
  Resource Version:    68270
  UID:                 62a105f7-e7b9-444e-9303-79818fccfdef
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   pb-vertical
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
    Message:               PgBouncer ops-request has started to vertically scaling the PgBouncer nodes
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
    Message:               get pod; ConditionStatus:True; PodName:pb-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pb-vertical-0
    Last Transition Time:  2024-07-17T09:44:30Z
    Message:               evict pod; ConditionStatus:True; PodName:pb-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pb-vertical-0
    Last Transition Time:  2024-07-17T09:45:05Z
    Message:               check pod running; ConditionStatus:True; PodName:pb-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pb-vertical-0
    Last Transition Time:  2024-07-17T09:45:10Z
    Message:               Successfully completed the vertical scaling for PgBouncer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                           Age    From                         Message
  ----     ------                                                           ----   ----                         -------
  Normal   Starting                                                         4m16s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pgbouncer-scale-vertical
  Normal   Starting                                                         4m16s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pb-vertical
  Normal   Successful                                                       4m16s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pb-vertical for PgBouncerOpsRequest: pgbouncer-scale-vertical
  Normal   UpdatePetSets                                                    4m13s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:pb-vertical-0             4m8s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pb-vertical-0
  Warning  evict pod; ConditionStatus:True; PodName:pb-vertical-0           4m8s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pb-vertical-0
  Warning  check pod running; ConditionStatus:False; PodName:pb-vertical-0  4m3s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pb-vertical-0
  Warning  check pod running; ConditionStatus:True; PodName:pb-vertical-0   3m33s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pb-vertical-0
  Normal   RestartPods                                                      3m28s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                         3m28s  KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pb-vertical
  Normal   Successful                                                       3m28s  KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pb-vertical for PgBouncerOpsRequest: pgbouncer-scale-vertical
```

Now, we are going to verify from the Pod yaml whether the resources of the pgbouncer has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo pb-vertical-0 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the PgBouncer.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pb -n demo pb-vertical
kubectl delete pgbounceropsrequest -n demo pgbouncer-scale-vertical
```