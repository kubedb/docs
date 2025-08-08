---
title: Cassandra Compute Autoscaling
menu:
  docs_{{ .version }}:
    identifier: cas-auto-scaling-compute-autoscale
    name: Topology Cluster
    parent: cas-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Cassandra

This guide will show you how to use `KubeDB` to autoscaling compute resources i.e. cpu and memory of a Cassandra.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
  - [CassandraAutoscaler](/docs/guides/cassandra/concepts/cassandraautoscaler.md)
  - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/cassandra/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/cassandra](/docs/examples/cassandra) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Cassandra

In this section, we are going to deploy a Cassandra with version `5.0.3`  Then, in the next section we will set up autoscaling for this Cassandra using `CassandraAutoscaler` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-autoscale
  namespace: demo
spec:
  version: "5.0.3"
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 600Mi
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 2Gi
                    cpu: 1000m
                  requests:
                    memory: 600Mi
                    cpu: 500m
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/autoscaling/compute/cassandra-autoscale.yaml
cassandra.kubedb.com/cassandra-autoscale created
```

Now, wait until `cassandra-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get cas -n demo
NAME                 TYPE                  VERSION   STATUS   AGE
cassandra-autoscale   kubedb.com/v1alpha2   5.0.3     Ready    22s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo cassandra-autoscale-rack-r0-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "600Mi"
  }
}
```

Let's check the Cassandra resources,
```bash
$ kubectl get cassandra -n demo cassandra-autoscale -o json | jq '.spec.topology.rack[0].podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "600Mi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the cassandra.

We are now ready to apply the `CassandraAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a CassandraAutoscaler Object.

#### Create CassandraAutoscaler Object

In order to set up compute resource autoscaling for this Cassandra, we have to create a `CassandraAutoscaler` CRO with our desired configuration. Below is the YAML of the `CassandraAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: CassandraAutoscaler
metadata:
  name: cassandra-autoscale-ops
  namespace: demo
spec:
  databaseRef:
    name: cassandra-autoscale
  compute:
    cassandra:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 1.2Gi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `cassandra-autoscale`.
- `spec.compute.cassandra.trigger` specifies that compute resource autoscaling is enabled for this cassandra.
- `spec.compute.cassandra.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.cassandra.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.cassandra.minAllowed` specifies the minimum allowed resources for this cassandra.
- `spec.compute.cassandra.maxAllowed` specifies the maximum allowed resources for this cassandra.
- `spec.compute.cassandra.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.cassandra.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here :  [timeout](/docs/guides/cassandra/concepts/cassandraopsrequest.md#spectimeout), [apply](/docs/guides/cassandra/concepts/cassandraopsrequest.md#specapply).

Let's create the `CassandraAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/autoscaling/compute/cassandra-autoscaler-ops.yaml
cassandraautoscaler.autoscaling.kubedb.com/cassandra-autoscaler-ops created
```

#### Verify Autoscaling is set up successfully

Let's check that the `cassandraautoscaler` resource is created successfully,

```bash
$ kubectl get cassandraautoscaler -n demo
NAME                   AGE
cassandra-autoscale-ops   6m55s

$ kubectl describe cassandraautoscaler cassandra-autoscale-ops -n demo
Name:         cassandra-autoscale-ops
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         CassandraAutoscaler
Metadata:
  Creation Timestamp:  2025-07-14T09:51:36Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Cassandra
    Name:                  cassandra-autoscale
    UID:                   4abde3d7-a0f4-4f4e-bcdc-6518444f1e23
  Resource Version:        10064
  UID:                     e29a989b-becd-4812-9ab3-65adead1d578
Spec:
  Compute:
    Cassandra:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:                     800m
        Memory:                  2Gi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  cassandra-autoscale
  Ops Request Options:
    Apply:  IfReady
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              1
        Weight:             4631
        Index:              3
        Weight:             10000
        Index:              44
        Weight:             6125
      Reference Timestamp:  2025-07-14T10:10:00Z
      Total Weight:         0.4133802430760871
    First Sample Start:     2025-07-14T10:09:25Z
    Last Sample Start:      2025-07-14T10:11:26Z
    Last Update Time:       2025-07-14T10:11:46Z
    Memory Histogram:
      Reference Timestamp:  2025-07-14T10:15:00Z
    Ref:
      Container Name:     cassandra
      Vpa Object Name:    cassandra-autoscale-rack-r0
    Total Samples Count:  4
    Version:              v3
  Conditions:
    Last Transition Time:  2025-07-14T10:09:46Z
    Message:               Successfully created CassandraOpsRequest demo/casops-cassandra-autoscale-rack-r0-kefyuq
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2025-07-14T10:09:46Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  cassandra
        Lower Bound:
          Cpu:     800m
          Memory:  2Gi
        Target:
          Cpu:     1836m
          Memory:  2Gi
        Uncapped Target:
          Cpu:     1836m
          Memory:  1168723596
        Upper Bound:
          Cpu:     2
          Memory:  3Gi
    Vpa Name:      cassandra-autoscale-rack-r0
Events:            <none>
```
So, the `Cassandraautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our Cassandra. Our autoscaler operator continuously watches the recommendation generated and creates an `cassandraopsrequest` based on the recommendations, if the cassandra pods are needed to scaled up or down.

Let's watch the `cassandraopsrequest` in the demo namespace to see if any `cassandraopsrequest` object is created. After some time you'll see that a `cassandraopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get cassandraopsrequest -n demo
Every 2.0s: kubectl get cassandraopsrequest -n demo
NAME                                        TYPE              STATUS        AGE
casops-cassandra-autoscale-rack-r0-kefyuq   VerticalScaling   Progressing   1m28s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get cassandraopsrequest -n demo
Every 2.0s: kubectl get cassandraopsrequest -n demo
NAME                                        TYPE              STATUS       AGE
casops-cassandra-autoscale-rack-r0-kefyuq   VerticalScaling   Successful   3m34s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to scale the Cassandra.

```bash
$ kubectl describe cassandraopsrequest -n demo casops-cassandra-autoscale-rack-r0-kefyuq
Name:         casops-cassandra-autoscale-rack-r0-kefyuq
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=cassandra-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=cassandras.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-14T10:09:46Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  CassandraAutoscaler
    Name:                  cassandra-autoscale-ops
    UID:                   e29a989b-becd-4812-9ab3-65adead1d578
  Resource Version:        10149
  UID:                     e55a25aa-c629-48a8-a6d6-ff7df56edf2d
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-autoscale
  Type:    VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     1600m
          Memory:  7330077518
        Requests:
          Cpu:     800m
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2025-07-14T10:09:46Z
    Message:               Cassandra ops-request has started to vertically scaling the Cassandra nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2025-07-14T10:09:49Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-14T10:12:34Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-07-14T10:09:54Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-autoscale-rack-r0-0
    Last Transition Time:  2025-07-14T10:09:54Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-autoscale-rack-r0-0
    Last Transition Time:  2025-07-14T10:09:59Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-14T10:10:34Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-autoscale-rack-r0-1
    Last Transition Time:  2025-07-14T10:10:34Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-autoscale-rack-r0-1
    Last Transition Time:  2025-07-14T10:12:34Z
    Message:               Successfully completed the vertical scaling for Cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                  Age    From                         Message
  ----     ------                                                                  ----   ----                         -------
  Normal   Starting                                                                4m17s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-cassandra-autoscale-rack-r0-kefyuq
  Normal   Starting                                                                4m17s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-autoscale
  Normal   Successful                                                              4m17s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-autoscale for CassandraOpsRequest: casops-cassandra-autoscale-rack-r0-kefyuq
  Normal   UpdatePetSets                                                           4m14s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0    4m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0  4m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0
  Warning  running pod; ConditionStatus:False                                      4m4s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1    3m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1  3m29s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0    2m49s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0  2m49s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1    2m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1  2m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-autoscale-rack-r0-1
  Normal   RestartPods                                                             89s    KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                89s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-autoscale
  Normal   Successful                                                              89s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-autoscale for CassandraOpsRequest: casops-cassandra-autoscale-rack-r0-kefyuq
```

Now, we are going to verify from the Pod, and the Cassandra yaml whether the resources of the Cassandra has updated to meet up the desired state, Let's check,

```bash
$  kubectl get pod -n demo cassandra-autoscale-rack-r0-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1600m",
    "memory": "7330077518"
  },
  "requests": {
    "cpu": "800m",
    "memory": "2Gi"
  }
}

$  kubectl get cassandra -n demo cassandra-autoscale -o json | jq '.spec.topology.rack[0].podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "600Mi"
  }
}
```

The above output verifies that we have successfully auto-scaled the resources of the cassandra.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cas -n demo cassandra-autoscale
```