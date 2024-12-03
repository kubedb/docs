---
title: PgBouncer Autoscaling
menu:
  docs_{{ .version }}:
    identifier: pb-auto-scaling-pgbouncer
    name: pgbouncerCompute
    parent: pb-compute-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a PgBouncer

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a PgBouncer.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerAutoscaler](/docs/guides/pgbouncer/concepts/autoscaler.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/pgbouncer/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgbouncer](/docs/examples/pgbouncer) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of PgBouncer

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgbouncer/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

Here, we are going to deploy a `PgBouncer` standalone using a supported version by `KubeDB` operator. Then we are going to apply `PgBouncerAutoscaler` to set up autoscaling.

#### Deploy PgBouncer

In this section, we are going to deploy a PgBouncer with version `1.18.0`  Then, in the next section we will set up autoscaling for this pgbouncer using `PgBouncerAutoscaler` CRD. Below is the YAML of the `PgBouncer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: pgbouncer-autoscale
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

Let's create the `PgBouncer` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/autoscaling/compute/pgbouncer-autoscale.yaml
pgbouncer.kubedb.com/pgbouncer-autoscale created
```

Now, wait until `pgbouncer-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get pb -n demo
NAME                  TYPE                  VERSION   STATUS   AGE
pgbouncer-autoscale   kubedb.com/v1         1.18.0    Ready    22s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo pgbouncer-autoscale-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

Let's check the PgBouncer resources,
```bash
$ kubectl get pgbouncer -n demo pgbouncer-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the pgbouncer.

We are now ready to apply the `PgBouncerAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a PgBouncerAutoscaler Object.

#### Create PgBouncerAutoscaler Object

In order to set up compute resource autoscaling for this pgbouncer, we have to create a `PgBouncerAutoscaler` CRO with our desired configuration. Below is the YAML of the `PgBouncerAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PgBouncerAutoscaler
metadata:
  name: pgbouncer-autoscale-ops
  namespace: demo
spec:
  databaseRef:
    name: pgbouncer-autoscale
  compute:
    pgbouncer:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 400m
        memory: 400Mi
      maxAllowed:
        cpu: 1
        memory: 1Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `pgbouncer-autoscale`.
- `spec.compute.pgbouncer.trigger` specifies that compute resource autoscaling is enabled for this pgbouncer.
- `spec.compute.pgbouncer.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.replicaset.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.pgbouncer.minAllowed` specifies the minimum allowed resources for this pgbouncer.
- `spec.compute.pgbouncer.maxAllowed` specifies the maximum allowed resources for this pgbouncer.
- `spec.compute.pgbouncer.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.pgbouncer.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here :  [timeout](/docs/guides/pgbouncer/concepts/opsrequest.md#spectimeout), [apply](/docs/guides/pgbouncer/concepts/opsrequest.md#specapply).

Let's create the `PgBouncerAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/autoscaling/compute/pgbouncer-autoscaler.yaml
pgbouncerautoscaler.autoscaling.kubedb.com/pgbouncer-autoscaler-ops created
```

#### Verify Autoscaling is set up successfully

Let's check that the `pgbouncerautoscaler` resource is created successfully,

```bash
$ kubectl get pgbouncerautoscaler -n demo
NAME                      AGE
pgbouncer-autoscale-ops   6m55s

$ kubectl describe pgbouncerautoscaler pgbouncer-autoscale-ops -n demo
Name:         pgbouncer-autoscale-ops
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         PgBouncerAutoscaler
Metadata:
  Creation Timestamp:  2024-07-17T12:09:17Z
  Generation:          1
  Resource Version:    81569
  UID:                 3841c30b-3b19-4740-82f5-bf8e257ddc18
Spec:
  Compute:
    PgBouncer:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:                     400m
        Memory:                  400Mi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  pgbouncer-autoscale
  Ops Request Options:
    Apply:  IfReady
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
      Reference Timestamp:  2024-07-17T12:10:00Z
      Total Weight:         0.8733542386168607
    First Sample Start:     2024-07-17T12:09:14Z
    Last Sample Start:      2024-07-17T12:15:06Z
    Last Update Time:       2024-07-17T12:15:38Z
    Memory Histogram:
      Bucket Weights:
        Index:              11
        Weight:             10000
      Reference Timestamp:  2024-07-17T12:15:00Z
      Total Weight:         0.7827734162991002
    Ref:
      Container Name:     pgbouncer
      Vpa Object Name:    pgbouncer-autoscale
    Total Samples Count:  6
    Version:              v3
  Conditions:
    Last Transition Time:  2024-07-17T12:10:37Z
    Message:               Successfully created PgBouncerOpsRequest demo/pbops-pgbouncer-autoscale-zzell6
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-07-17T12:09:37Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  pgbouncer
        Lower Bound:
          Cpu:     400m
          Memory:  400Mi
        Target:
          Cpu:     400m
          Memory:  400Mi
        Uncapped Target:
          Cpu:     100m
          Memory:  262144k
        Upper Bound:
          Cpu:     1
          Memory:  1Gi
    Vpa Name:      pgbouncer-autoscale
Events:            <none>
```
So, the `pgbouncerautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our pgbouncer. Our autoscaler operator continuously watches the recommendation generated and creates an `pgbounceropsrequest` based on the recommendations, if the pgbouncer pods are needed to scaled up or down.

Let's watch the `pgbounceropsrequest` in the demo namespace to see if any `pgbounceropsrequest` object is created. After some time you'll see that a `pgbounceropsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                               TYPE              STATUS        AGE
pbops-pgbouncer-autoscale-zzell6   VerticalScaling   Progressing   1m48s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get pgbounceropsrequest -n demo
Every 2.0s: kubectl get pgbounceropsrequest -n demo
NAME                               TYPE              STATUS       AGE
pbops-pgbouncer-autoscale-zzell6   VerticalScaling   Successful   3m40s
```

We can see from the above output that the `PgBouncerOpsRequest` has succeeded. If we describe the `PgBouncerOpsRequest` we will get an overview of the steps that were followed to scale the pgbouncer.

```bash
$ kubectl describe pgbounceropsrequest -n demo pbops-pgbouncer-autoscale-zzell6
Name:         pbops-pgbouncer-autoscale-zzell6
Namespace:    demo
Labels:       app.kubernetes.io/component=connection-pooler
              app.kubernetes.io/instance=pgbouncer-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=pgbouncers.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgBouncerOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T12:10:37Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  PgBouncerAutoscaler
    Name:                  pgbouncer-autoscale-ops
    UID:                   3841c30b-3b19-4740-82f5-bf8e257ddc18
  Resource Version:        81200
  UID:                     57f99d31-af3d-4157-aa61-0f509ec89bbd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgbouncer-autoscale
  Type:    VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     400m
          Memory:  400Mi
        Requests:
          Cpu:     400m
          Memory:  400Mi
Status:
  Conditions:
    Last Transition Time:  2024-07-17T12:10:37Z
    Message:               PgBouncer ops-request has started to vertically scaling the PgBouncer nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-07-17T12:10:40Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-07-17T12:10:40Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-17T12:11:25Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-17T12:10:45Z
    Message:               get pod; ConditionStatus:True; PodName:pgbouncer-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgbouncer-autoscale-0
    Last Transition Time:  2024-07-17T12:10:45Z
    Message:               evict pod; ConditionStatus:True; PodName:pgbouncer-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgbouncer-autoscale-0
    Last Transition Time:  2024-07-17T12:11:20Z
    Message:               check pod running; ConditionStatus:True; PodName:pgbouncer-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgbouncer-autoscale-0
    Last Transition Time:  2024-07-17T12:11:26Z
    Message:               Successfully completed the vertical scaling for PgBouncer
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                   Age    From                         Message
  ----     ------                                                                   ----   ----                         -------
  Normal   Starting                                                                 8m19s  KubeDB Ops-manager Operator  Start processing for PgBouncerOpsRequest: demo/pbops-pgbouncer-autoscale-zzell6
  Normal   Starting                                                                 8m19s  KubeDB Ops-manager Operator  Pausing PgBouncer databse: demo/pgbouncer-autoscale
  Normal   Successful                                                               8m19s  KubeDB Ops-manager Operator  Successfully paused PgBouncer database: demo/pgbouncer-autoscale for PgBouncerOpsRequest: pbops-pgbouncer-autoscale-zzell6
  Normal   UpdatePetSets                                                            8m16s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:pgbouncer-autoscale-0             8m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgbouncer-autoscale-0
  Warning  evict pod; ConditionStatus:True; PodName:pgbouncer-autoscale-0           8m11s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgbouncer-autoscale-0
  Warning  check pod running; ConditionStatus:False; PodName:pgbouncer-autoscale-0  8m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgbouncer-autoscale-0
  Warning  check pod running; ConditionStatus:True; PodName:pgbouncer-autoscale-0   7m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgbouncer-autoscale-0
  Normal   RestartPods                                                              7m31s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                 7m31s  KubeDB Ops-manager Operator  Resuming PgBouncer database: demo/pgbouncer-autoscale
  Normal   Successful                                                               7m30s  KubeDB Ops-manager Operator  Successfully resumed PgBouncer database: demo/pgbouncer-autoscale for PgBouncerOpsRequest: pbops-pgbouncer-autoscale-zzell6
```

Now, we are going to verify from the Pod, and the PgBouncer yaml whether the resources of the pgbouncer has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo pgbouncer-autoscale-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "400Mi"
  }
}

$ kubectl get pgbouncer -n demo pgbouncer-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "400Mi"
  }
}
```


The above output verifies that we have successfully auto-scaled the resources of the PgBouncer.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pb -n demo pgbouncer-autoscale
kubectl delete pgbouncerautoscaler -n demo pgbouncer-autoscale-ops
```