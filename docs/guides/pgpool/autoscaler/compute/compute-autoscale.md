---
title: Pgpool Autoscaling
menu:
  docs_{{ .version }}:
    identifier: pp-auto-scaling-pgpool
    name: pgpoolCompute
    parent: pp-compute-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Pgpool

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Pgpool.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [Pgpool](/docs/guides/pgpool/concepts/pgpool.md)
  - [PgpoolAutoscaler](/docs/guides/pgpool/concepts/autoscaler.md)
  - [PgpoolOpsRequest](/docs/guides/pgpool/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/pgpool/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/pgpool](/docs/examples/pgpool) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Pgpool

### Prepare Postgres
Prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.

Here, we are going to deploy a `Pgpool` standalone using a supported version by `KubeDB` operator. Then we are going to apply `PgpoolAutoscaler` to set up autoscaling.

#### Deploy Pgpool

In this section, we are going to deploy a Pgpool with version `4.5.0`  Then, in the next section we will set up autoscaling for this pgpool using `PgpoolAutoscaler` CRD. Below is the YAML of the `Pgpool` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pgpool-autoscale
  namespace: demo
spec:
  version: "4.5.0"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      containers:
        - name: pgpool
          resources:
            requests:
              cpu: "200m"
              memory: "300Mi"
            limits:
              cpu: "200m"
              memory: "300Mi"
  deletionPolicy: WipeOut
```

Let's create the `Pgpool` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/autoscaling/compute/pgpool-autoscale.yaml
pgpool.kubedb.com/pgpool-autoscale created
```

Now, wait until `pgpool-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get pp -n demo
NAME               TYPE                  VERSION   STATUS   AGE
pgpool-autoscale   kubedb.com/v1alpha2   4.5.0     Ready    22s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo pgpool-autoscale-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

Let's check the Pgpool resources,
```bash
$ kubectl get pgpool -n demo pgpool-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "200m",
    "memory": "300Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "300Mi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the pgpool.

We are now ready to apply the `PgpoolAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a PgpoolAutoscaler Object.

#### Create PgpoolAutoscaler Object

In order to set up compute resource autoscaling for this pgpool, we have to create a `PgpoolAutoscaler` CRO with our desired configuration. Below is the YAML of the `PgpoolAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PgpoolAutoscaler
metadata:
  name: pgpool-autoscale-ops
  namespace: demo
spec:
  databaseRef:
    name: pgpool-autoscale
  compute:
    pgpool:
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

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `pgpool-autoscale`.
- `spec.compute.pgpool.trigger` specifies that compute resource autoscaling is enabled for this pgpool.
- `spec.compute.pgpool.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.replicaset.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.pgpool.minAllowed` specifies the minimum allowed resources for this pgpool.
- `spec.compute.pgpool.maxAllowed` specifies the maximum allowed resources for this pgpool.
- `spec.compute.pgpool.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.pgpool.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here :  [timeout](/docs/guides/pgpool/concepts/opsrequest.md#spectimeout), [apply](/docs/guides/pgpool/concepts/opsrequest.md#specapply).

Let's create the `PgpoolAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/autoscaling/compute/pgpool-autoscaler.yaml
pgpoolautoscaler.autoscaling.kubedb.com/pgpool-autoscaler-ops created
```

#### Verify Autoscaling is set up successfully

Let's check that the `pgpoolautoscaler` resource is created successfully,

```bash
$ kubectl get pgpoolautoscaler -n demo
NAME                   AGE
pgpool-autoscale-ops   6m55s

$ kubectl describe pgpoolautoscaler pgpool-autoscale-ops -n demo
Name:         pgpool-autoscale-ops
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         PgpoolAutoscaler
Metadata:
  Creation Timestamp:  2024-07-17T12:09:17Z
  Generation:          1
  Resource Version:    81569
  UID:                 3841c30b-3b19-4740-82f5-bf8e257ddc18
Spec:
  Compute:
    Pgpool:
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
    Name:  pgpool-autoscale
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
      Container Name:     pgpool
      Vpa Object Name:    pgpool-autoscale
    Total Samples Count:  6
    Version:              v3
  Conditions:
    Last Transition Time:  2024-07-17T12:10:37Z
    Message:               Successfully created PgpoolOpsRequest demo/ppops-pgpool-autoscale-zzell6
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
        Container Name:  pgpool
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
    Vpa Name:      pgpool-autoscale
Events:            <none>
```
So, the `pgpoolautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our pgpool. Our autoscaler operator continuously watches the recommendation generated and creates an `pgpoolopsrequest` based on the recommendations, if the pgpool pods are needed to scaled up or down.

Let's watch the `pgpoolopsrequest` in the demo namespace to see if any `pgpoolopsrequest` object is created. After some time you'll see that a `pgpoolopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                            TYPE              STATUS        AGE
ppops-pgpool-autoscale-zzell6   VerticalScaling   Progressing   1m48s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get pgpoolopsrequest -n demo
Every 2.0s: kubectl get pgpoolopsrequest -n demo
NAME                            TYPE              STATUS       AGE
ppops-pgpool-autoscale-zzell6   VerticalScaling   Successful   3m40s
```

We can see from the above output that the `PgpoolOpsRequest` has succeeded. If we describe the `PgpoolOpsRequest` we will get an overview of the steps that were followed to scale the pgpool.

```bash
$ kubectl describe pgpoolopsrequest -n demo ppops-pgpool-autoscale-zzell6
Name:         ppops-pgpool-autoscale-zzell6
Namespace:    demo
Labels:       app.kubernetes.io/component=connection-pooler
              app.kubernetes.io/instance=pgpool-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=pgpools.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PgpoolOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T12:10:37Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  PgpoolAutoscaler
    Name:                  pgpool-autoscale-ops
    UID:                   3841c30b-3b19-4740-82f5-bf8e257ddc18
  Resource Version:        81200
  UID:                     57f99d31-af3d-4157-aa61-0f509ec89bbd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  pgpool-autoscale
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
    Message:               Pgpool ops-request has started to vertically scaling the Pgpool nodes
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
    Message:               get pod; ConditionStatus:True; PodName:pgpool-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--pgpool-autoscale-0
    Last Transition Time:  2024-07-17T12:10:45Z
    Message:               evict pod; ConditionStatus:True; PodName:pgpool-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--pgpool-autoscale-0
    Last Transition Time:  2024-07-17T12:11:20Z
    Message:               check pod running; ConditionStatus:True; PodName:pgpool-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--pgpool-autoscale-0
    Last Transition Time:  2024-07-17T12:11:26Z
    Message:               Successfully completed the vertical scaling for Pgpool
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age    From                         Message
  ----     ------                                                                ----   ----                         -------
  Normal   Starting                                                              8m19s  KubeDB Ops-manager Operator  Start processing for PgpoolOpsRequest: demo/ppops-pgpool-autoscale-zzell6
  Normal   Starting                                                              8m19s  KubeDB Ops-manager Operator  Pausing Pgpool databse: demo/pgpool-autoscale
  Normal   Successful                                                            8m19s  KubeDB Ops-manager Operator  Successfully paused Pgpool database: demo/pgpool-autoscale for PgpoolOpsRequest: ppops-pgpool-autoscale-zzell6
  Normal   UpdatePetSets                                                         8m16s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:pgpool-autoscale-0             8m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:pgpool-autoscale-0
  Warning  evict pod; ConditionStatus:True; PodName:pgpool-autoscale-0           8m11s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:pgpool-autoscale-0
  Warning  check pod running; ConditionStatus:False; PodName:pgpool-autoscale-0  8m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:pgpool-autoscale-0
  Warning  check pod running; ConditionStatus:True; PodName:pgpool-autoscale-0   7m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:pgpool-autoscale-0
  Normal   RestartPods                                                           7m31s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                              7m31s  KubeDB Ops-manager Operator  Resuming Pgpool database: demo/pgpool-autoscale
  Normal   Successful                                                            7m30s  KubeDB Ops-manager Operator  Successfully resumed Pgpool database: demo/pgpool-autoscale for PgpoolOpsRequest: ppops-pgpool-autoscale-zzell6
```

Now, we are going to verify from the Pod, and the Pgpool yaml whether the resources of the pgpool has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo pgpool-autoscale-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "400Mi"
  }
}

$ kubectl get pgpool -n demo pgpool-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "400m",
    "memory": "400Mi"
  },
  "requests": {
    "cpu": "400m",
    "memory": "400Mi"
  }
}
```


The above output verifies that we have successfully auto-scaled the resources of the Pgpool.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pp -n demo pgpool-autoscale
kubectl delete pgpoolautoscaler -n demo pgpool-autoscale-ops
```