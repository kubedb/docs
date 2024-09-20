---
title: RabbitMQ Compute Resource Autoscaling
menu:
  docs_{{ .version }}:
    identifier: rm-autoscaling-compute-description
    name: Autoscaling Compute Resources
    parent: rm-autoscaling-compute
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a RabbitMQ

This guide will show you how to use `KubeDB` to autoscaling compute resources i.e. cpu and memory of a RabbitMQ.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
  - [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md)
  - [RabbitMQAutoscaler](/docs/guides/rabbitmq/concepts/autoscaler.md)
  - [RabbitMQOpsRequest](/docs/guides/rabbitmq/concepts/opsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/rabbitmq/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/rabbitmq](/docs/examples/rabbitmq) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of RabbitMQ

In this section, we are going to deploy a RabbitMQ with version `3.13.2`  Then, in the next section we will set up autoscaling for this RabbitMQ using `RabbitMQAutoscaler` CRD. Below is the YAML of the `RabbitMQ` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rabbitmq-autoscale
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
        - name: rabbitmq
          resources:
            requests:
              cpu: "0.5m"
              memory: "1Gi"
            limits:
              cpu: "1"
              memory: "2Gi"
  serviceTemplates:
    - alias: primary
      spec:
        type: LoadBalancer
```

Let's create the `RabbitMQ` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/autoscaling/compute/rabbitmq-autoscale.yaml
rabbitmq.kubedb.com/rabbitmq-autoscale created
```

Now, wait until `rabbitmq-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get rm -n demo
NAME                 TYPE                  VERSION   STATUS   AGE
rabbitmq-autoscale   kubedb.com/v1alpha2   3.13.2     Ready    22s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo rabbitmq-autoscale-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "0.5m",
    "memory": "1Gi"
  }
}
```

Let's check the RabbitMQ resources,
```bash
$ kubectl get rabbitmq -n demo rabbitmq-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "0.5m",
    "memory": "1Gi"
  }
}
```

You can see from the above outputs that the resources are same as the one we have assigned while deploying the rabbitmq.

We are now ready to apply the `RabbitMQAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute (cpu and memory) autoscaling using a RabbitMQAutoscaler Object.

#### Create RabbitMQAutoscaler Object

In order to set up compute resource autoscaling for this RabbitMQ, we have to create a `RabbitMQAutoscaler` CRO with our desired configuration. Below is the YAML of the `RabbitMQAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: RabbitMQAutoscaler
metadata:
  name: rabbitmq-autoscale-ops
  namespace: demo
spec:
  databaseRef:
    name: rabbitmq-autoscale
  compute:
    rabbitmq:
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

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on `rabbitmq-autoscale`.
- `spec.compute.rabbitmq.trigger` specifies that compute resource autoscaling is enabled for this rabbitmq.
- `spec.compute.rabbitmq.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.rabbitmq.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.rabbitmq.minAllowed` specifies the minimum allowed resources for this rabbitmq.
- `spec.compute.rabbitmq.maxAllowed` specifies the maximum allowed resources for this rabbitmq.
- `spec.compute.rabbitmq.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.rabbitmq.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields. Know more about them here :  [timeout](/docs/guides/rabbitmq/concepts/opsrequest.md#spectimeout), [apply](/docs/guides/rabbitmq/concepts/opsrequest.md#specapply).

Let's create the `RabbitMQAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/autoscaling/compute/rabbitmq-autoscaler.yaml
rabbitmqautoscaler.autoscaling.kubedb.com/rabbitmq-autoscaler-ops created
```

#### Verify Autoscaling is set up successfully

Let's check that the `rabbitmqautoscaler` resource is created successfully,

```bash
$ kubectl get rabbitmqautoscaler -n demo
NAME                   AGE
rabbitmq-autoscale-ops   6m55s

$ kubectl describe rabbitmqautoscaler rabbitmq-autoscale-ops -n demo
Name:         rabbitmq-autoscale-ops
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         RabbitMQAutoscaler
Metadata:
  Creation Timestamp:  2024-07-17T12:09:17Z
  Generation:          1
  Resource Version:    81569
  UID:                 3841c30b-3b19-4740-82f5-bf8e257ddc18
Spec:
  Compute:
    rabbitmq:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  1Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  1.2Gi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  rabbitmq-autoscale
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
      Container Name:     rabbitmq
      Vpa Object Name:    rabbitmq-autoscale
    Total Samples Count:  6
    Version:              v3
  Conditions:
    Last Transition Time:  2024-07-17T12:10:37Z
    Message:               Successfully created RabbitMQOpsRequest demo/rmops-rabbitmq-autoscale-zzell6
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
        Container Name:  rabbitmq
        Lower Bound:
          Cpu:     600m
          Memory:  1.2Gi
        Target:
          Cpu:     600m
          Memory:  1.2Gi
        Uncapped Target:
          Cpu:     500m
          Memory:  2621445k
        Upper Bound:
          Cpu:     1
          Memory:  2Gi
    Vpa Name:      rabbitmq-autoscale
Events:            <none>
```
So, the `RabbitMQautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our RabbitMQ. Our autoscaler operator continuously watches the recommendation generated and creates an `rabbitmqopsrequest` based on the recommendations, if the rabbitmq pods are needed to scaled up or down.

Let's watch the `rabbitmqopsrequest` in the demo namespace to see if any `rabbitmqopsrequest` object is created. After some time you'll see that a `rabbitmqopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME                            TYPE              STATUS        AGE
rmops-rabbitmq-autoscale-zzell6   VerticalScaling   Progressing   1m48s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get rabbitmqopsrequest -n demo
Every 2.0s: kubectl get rabbitmqopsrequest -n demo
NAME                            TYPE              STATUS       AGE
rmops-rabbitmq-autoscale-zzell6   VerticalScaling   Successful   3m40s
```

We can see from the above output that the `RabbitMQOpsRequest` has succeeded. If we describe the `RabbitMQOpsRequest` we will get an overview of the steps that were followed to scale the RabbitMQ.

```bash
$ kubectl describe rabbitmqopsrequest -n demo rmops-rabbitmq-autoscale-zzell6
Name:         rmops-rabbitmq-autoscale-zzell6
Namespace:    demo
Labels:       app.kubernetes.io/component=connection-pooler
              app.kubernetes.io/instance=rabbitmq-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=rabbitmqs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RabbitMQOpsRequest
Metadata:
  Creation Timestamp:  2024-07-17T12:10:37Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  RabbitMQAutoscaler
    Name:                  rabbitmq-autoscale-ops
    UID:                   3841c30b-3b19-4740-82f5-bf8e257ddc18
  Resource Version:        81200
  UID:                     57f99d31-af3d-4157-aa61-0f509ec89bbd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  rabbitmq-autoscale
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
    Message:               RabbitMQ ops-request has started to vertically scaling the RabbitMQ nodes
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
    Message:               get pod; ConditionStatus:True; PodName:rabbitmq-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--rabbitmq-autoscale-0
    Last Transition Time:  2024-07-17T12:10:45Z
    Message:               evict pod; ConditionStatus:True; PodName:rabbitmq-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--rabbitmq-autoscale-0
    Last Transition Time:  2024-07-17T12:11:20Z
    Message:               check pod running; ConditionStatus:True; PodName:rabbitmq-autoscale-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--rabbitmq-autoscale-0
    Last Transition Time:  2024-07-17T12:11:26Z
    Message:               Successfully completed the vertical scaling for rabbitmq
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                Age    From                         Message
  ----     ------                                                                ----   ----                         -------
  Normal   Starting                                                              8m19s  KubeDB Ops-manager Operator  Start processing for rabbitmqOpsRequest: demo/rmops-rabbitmq-autoscale-zzell6
  Normal   Starting                                                              8m19s  KubeDB Ops-manager Operator  Pausing rabbitmq databse: demo/rabbitmq-autoscale
  Normal   Successful                                                            8m19s  KubeDB Ops-manager Operator  Successfully paused rabbitmq database: demo/rabbitmq-autoscale for rabbitmqOpsRequest: rmops-rabbitmq-autoscale-zzell6
  Normal   UpdatePetSets                                                         8m16s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:rabbitmq-autoscale-0             8m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:rabbitmq-autoscale-0
  Warning  evict pod; ConditionStatus:True; PodName:rabbitmq-autoscale-0           8m11s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:rabbitmq-autoscale-0
  Warning  check pod running; ConditionStatus:False; PodName:rabbitmq-autoscale-0  8m6s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:rabbitmq-autoscale-0
  Warning  check pod running; ConditionStatus:True; PodName:rabbitmq-autoscale-0   7m36s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:rabbitmq-autoscale-0
  Normal   RestartPods                                                           7m31s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                              7m31s  KubeDB Ops-manager Operator  Resuming rabbitmq database: demo/rabbitmq-autoscale
  Normal   Successful                                                            7m30s  KubeDB Ops-manager Operator  Successfully resumed RabbitMQ database: demo/rabbitmq-autoscale for RabbitMQOpsRequest: rmops-rabbitmq-autoscale-zzell6
```

Now, we are going to verify from the Pod, and the RabbitMQ yaml whether the resources of the RabbitMQ has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo rabbitmq-autoscale-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1.2Gi"
  }
}

$ kubectl get rabbitmq -n demo rabbitmq-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1.2Gi"
  }
}
```


The above output verifies that we have successfully auto-scaled the resources of the rabbitmq.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rm -n demo rabbitmq-autoscale
kubectl delete rabbitmqautoscaler -n demo rabbitmq-autoscale-ops
```