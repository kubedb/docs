---
title: ClickHouse Compute Autoscaling
menu:
  docs_{{ .version }}:
    identifier: ch-auto-scaling-compute-autoscale
    name: Cluster
    parent: ch-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a ClickHouse

This guide will show you how to use `KubeDB` to autoscaling compute resources i.e. cpu and memory of a ClickHouse.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
    - [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md)
    - [ClickHouseAutoscaler](/docs/guides/clickhouse/concepts/clickhouseautoscaler.md)
    - [ClickHouseOpsRequest](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md)
    - [Compute Resource Autoscaling Overview](/docs/guides/clickhouse/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/clickhouse](/docs/examples/clickhouse) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of ClickHouse

In this section, we are going to deploy a ClickHouse with version `24.4.1`  Then, in the next section we will set up autoscaling for this ClickHouse using `ClickHouseAutoscaler` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      name: appscode-cluster
      shards: 2
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: clickhouse
              resources:
                limits:
                  memory: 2Gi
                requests:
                  memory: 1Gi
                  cpu: 900m
          initContainers:
            - name: clickhouse-init
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 1Gi
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/autoscaling/compute/clickhouse-autoscale.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
➤ kubectl get ch -n demo clickhouse-prod -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   114s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   117s
.
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m44s
```

Let's check the Pod containers resources,

```bash
➤ kubectl get pod -n demo clickhouse-prod-appscode-cluster-shard-0-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "1Gi"
  }
}
```

Let's check the ClickHouse resources,
```bash
➤ kubectl get clickhouse -n demo clickhouse-prod -o json | jq '.spec.clusterTopology.cluster.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "900m",
    "memory": "1Gi"
  }
}

```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the clickhouse.

We are now ready to apply the `ClickHouseAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a ClickHouseAutoscaler Object.

#### Create ClickHouseAutoscaler Object

In order to set up compute resource autoscaling for this ClickHouse, we have to create a `ClickHouseAutoscaler` CRO with our desired configuration. Below is the YAML of the `ClickHouseAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: ClickHouseAutoscaler
metadata:
  name: ch-compute-autoscale
  namespace: demo
spec:
  databaseRef:
    name: clickhouse-prod
  compute:
    clickhouse:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 1
        memory: 2Gi
      maxAllowed:
        cpu: 2
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `clickhouse-autoscale`.
- `spec.compute.clickhouse.trigger` specifies that compute resource autoscaling is enabled for this clickhouse.
- `spec.compute.clickhouse.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.clickhouse.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.clickhouse.minAllowed` specifies the minimum allowed resources for this clickhouse.
- `spec.compute.clickhouse.maxAllowed` specifies the maximum allowed resources for this clickhouse.
- `spec.compute.clickhouse.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.clickhouse.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here :  [timeout](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md#spectimeout), [apply](/docs/guides/clickhouse/concepts/clickhouseopsrequest.md#specapply).

Let's create the `ClickHouseAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/autoscaling/compute/clickhouse-autoscaler-ops.yaml
clickhouseautoscaler.autoscaling.kubedb.com/ch-compute-autoscale created
```

#### Verify Autoscaling is set up successfully

Let's check that the `clickhouseautoscaler` resource is created successfully,

```bash
➤ kubectl get clickhouseautoscaler -n demo
NAME                   AGE
ch-compute-autoscale   4m3s

➤ kubectl describe clickhouseautoscaler -n demo ch-compute-autoscale
Name:         ch-compute-autoscale
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         ClickHouseAutoscaler
Metadata:
  Creation Timestamp:  2025-10-07T05:36:00Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ClickHouse
    Name:                  clickhouse-prod
    UID:                   35393926-cb6b-46c2-814c-1ca563783128
  Resource Version:        698948
  UID:                     e8a75e40-c03c-401b-9b1d-647479c4fe67
Spec:
  Compute:
    Clickhouse:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:                     1
        Memory:                  2Gi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  clickhouse-prod
  Ops Request Options:
    Apply:  IfReady
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              2
        Weight:             2070
        Index:              5
        Weight:             10000
        Index:              6
        Weight:             1344
      Reference Timestamp:  2025-10-07T05:40:00Z
      Total Weight:         0.5562658287128397
    First Sample Start:     2025-10-07T05:35:47Z
    Last Sample Start:      2025-10-07T05:38:54Z
    Last Update Time:       2025-10-07T05:39:06Z
    Memory Histogram:
      Reference Timestamp:  2025-10-07T05:40:00Z
    Ref:
      Container Name:     clickhouse
      Vpa Object Name:    clickhouse-prod-appscode-cluster-shard-1
    Total Samples Count:  8
    Version:              v3
    Cpu Histogram:
      Bucket Weights:
        Index:              5
        Weight:             10000
        Index:              6
        Weight:             5380
      Reference Timestamp:  2025-10-07T05:40:00Z
      Total Weight:         0.6734523183104338
    First Sample Start:     2025-10-07T05:35:42Z
    Last Sample Start:      2025-10-07T05:39:45Z
    Last Update Time:       2025-10-07T05:40:06Z
    Memory Histogram:
      Bucket Weights:
        Index:              15
        Weight:             9840
        Index:              16
        Weight:             10000
        Index:              17
        Weight:             9840
        Index:              19
        Weight:             10000
      Reference Timestamp:  2025-10-07T05:40:00Z
      Total Weight:         3.868337950095009
    Ref:
      Container Name:     clickhouse
      Vpa Object Name:    clickhouse-prod-appscode-cluster-shard-0
    Total Samples Count:  9
    Version:              v3
  Conditions:
    Last Transition Time:  2025-10-07T05:37:06Z
    Message:               Successfully created ClickHouseOpsRequest demo/chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2025-10-07T05:38:06Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  clickhouse
        Lower Bound:
          Cpu:     1
          Memory:  2Gi
        Target:
          Cpu:     1
          Memory:  2Gi
        Uncapped Target:
          Cpu:     100m
          Memory:  351198544
        Upper Bound:
          Cpu:     2
          Memory:  3Gi
    Vpa Name:      clickhouse-prod-appscode-cluster-shard-1
    Conditions:
      Last Transition Time:  2025-10-07T05:36:06Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  clickhouse
        Lower Bound:
          Cpu:     1
          Memory:  2Gi
        Target:
          Cpu:     1
          Memory:  2Gi
        Uncapped Target:
          Cpu:     100m
          Memory:  380258472
        Upper Bound:
          Cpu:     2
          Memory:  3Gi
    Vpa Name:      clickhouse-prod-appscode-cluster-shard-0
Events:            <none>
```
So, the `ClickHouseautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our ClickHouse. Our autoscaler operator continuously watches the recommendation generated and creates an `clickhouseopsrequest` based on the recommendations, if the clickhouse pods are needed to scaled up or down.

Let's watch the `clickhouseopsrequest` in the demo namespace to see if any `clickhouseopsrequest` object is created. After some time you'll see that a `clickhouseopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get clickhouseopsrequest -n demo
Every 2.0s: kubectl get clickhouseopsrequest -n demo
NAME                                        TYPE              STATUS        AGE
chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v   VerticalScaling   Progressing   1m28s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get clickhouseopsrequest -n demo
Every 2.0s: kubectl get clickhouseopsrequest -n demo
NAME                                        TYPE              STATUS       AGE
chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v   VerticalScaling   Successful   3m34s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed to scale the ClickHouse.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v 
Name:         chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=clickhouse-prod
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=clickhouses.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-10-07T05:37:06Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ClickHouseAutoscaler
    Name:                  ch-compute-autoscale
    UID:                   e8a75e40-c03c-401b-9b1d-647479c4fe67
  Resource Version:        698821
  UID:                     171e1151-821b-4c51-b028-a18ff75a2133
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  clickhouse-prod
  Type:    VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Memory:  4Gi
        Requests:
          Cpu:     1
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2025-10-07T05:37:06Z
    Message:               ClickHouse ops-request has started to vertically scaling the ClickHouse nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2025-10-07T05:37:09Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-10-07T05:39:04Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-10-07T05:37:15Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-10-07T05:37:15Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-10-07T05:37:19Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-10-07T05:37:44Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-10-07T05:37:44Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-10-07T05:38:04Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-10-07T05:38:04Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-10-07T05:38:24Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-10-07T05:38:24Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-10-07T05:39:04Z
    Message:               Successfully completed the vertical scaling for ClickHouse
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age    From                         Message
  ----     ------                                                                               ----   ----                         -------
  Normal   Starting                                                                             7m42s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v
  Normal   Starting                                                                             7m42s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           7m42s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v
  Normal   UpdatePetSets                                                                        7m39s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    7m33s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  7m33s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   7m29s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    7m4s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  7m4s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    6m44s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  6m44s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    6m24s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  6m24s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartPods                                                                          5m44s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                             5m44s  KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           5m44s  KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-clickhouse-prod-appscode-cluster-shard-0-ckc28v
```

Now, we are going to verify from the Pod, and the ClickHouse yaml whether the resources of the ClickHouse has updated to meet up the desired state, Let's check,

```bash
➤ kubectl get pod -n demo clickhouse-prod-appscode-cluster-shard-0-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}


➤ kubectl get clickhouse -n demo clickhouse-prod -o json | jq '.spec.clusterTopology.cluster.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}

```

The above output verifies that we have successfully auto-scaled the resources of the clickhouse.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ch -n demo clickhouse-prod
kubectl delete clickhouseautoscaler -n demo ch-compute-autoscale
```