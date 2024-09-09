---
title: Kafka Topology Autoscaling
menu:
  docs_{{ .version }}:
    identifier: kf-auto-scaling-topology
    name: Topology Cluster
    parent: kf-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Kafka Topology Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Kafka topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [KafkaAutoscaler](/docs/guides/kafka/concepts/kafkaautoscaler.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Compute Resource Autoscaling Overview](/docs/guides/kafka/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Autoscaling of Topology Cluster

Here, we are going to deploy a `Kafka` Topology Cluster using a supported version by `KubeDB` operator. Then we are going to apply `KafkaAutoscaler` to set up autoscaling.

#### Deploy Kafka Topology Cluster

In this section, we are going to deploy a Kafka Topology cluster with version `3.6.1`.  Then, in the next section we will set up autoscaling for this database using `KafkaAutoscaler` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.6.1
  topology:
    broker:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 1Gi
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: kafka
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 1Gi
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

Let's create the `Kafka` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaler/kafka-topology.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait until `kafka-prod` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
kafka-prod    kubedb.com/v1   3.6.1     Provisioning   0s
kafka-prod    kubedb.com/v1   3.6.1     Provisioning   24s
.
.
kafka-prod    kubedb.com/v1   3.6.1     Ready          118s
```

## Kafka Topology Autoscaler(Broker) 

Let's check the Broker Pod containers resources,

```bash
$ kubectl get pod -n demo kafka-prod-broker-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

Let's check the Kafka resources for broker,
```bash
$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.broker.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see from the above outputs that the resources for broker are same as the one we have assigned while deploying the kafka.

We are now ready to apply the `KafkaAutoscaler` CRO to set up autoscaling for this broker nodes.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a KafkaAutoscaler Object.

#### Create KafkaAutoscaler Object

In order to set up compute resource autoscaling for this topology cluster, we have to create a `KafkaAutoscaler` CRO with our desired configuration. Below is the YAML of the `KafkaAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kf-broker-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: kafka-prod
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    broker:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 1.5Gi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `kafka-prod` cluster.
- `spec.compute.broker.trigger` specifies that compute autoscaling is enabled for this node.
- `spec.compute.broker.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.broker.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.broker.minAllowed` specifies the minimum allowed resources for the cluster.
- `spec.compute.broker.maxAllowed` specifies the maximum allowed resources for the cluster.
- `spec.compute.broker.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.broker.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields.
    - `timeout` specifies the timeout for the OpsRequest.
    - `apply` specifies when the OpsRequest should be applied. The default is "IfReady".

Let's create the `KafkaAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaler/compute/kafka-broker-autoscaler.yaml
kafkaautoscaler.autoscaling.kubedb.com/kf-broker-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `kafkaautoscaler` resource is created successfully,

```bash
$ kubectl describe kafkaautoscaler kf-broker-autoscaler -n demo
$ kubectl describe kafkaautoscaler kf-broker-autoscaler -n demo
Name:         kf-broker-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         KafkaAutoscaler
Metadata:
  Creation Timestamp:  2024-08-27T06:17:07Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Kafka
    Name:                  kafka-prod
    UID:                   7cee41e0-259c-4a5e-856a-e8ca90056120
  Resource Version:        1113275
  UID:                     7e3be99f-cd4d-440a-a477-8e8994840ebb
Spec:
  Compute:
    Broker:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  2Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  1536Mi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  kafka-prod
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              1
        Weight:             10000
        Index:              2
        Weight:             2485
        Index:              30
        Weight:             1923
      Reference Timestamp:  2024-08-27T06:20:00Z
      Total Weight:         0.8587070656303101
    First Sample Start:     2024-08-27T06:20:45Z
    Last Sample Start:      2024-08-27T06:23:53Z
    Last Update Time:       2024-08-27T06:24:10Z
    Memory Histogram:
      Bucket Weights:
        Index:              20
        Weight:             9682
        Index:              21
        Weight:             10000
      Reference Timestamp:  2024-08-27T06:25:00Z
      Total Weight:         1.9636285054518687
    Ref:
      Container Name:     kafka
      Vpa Object Name:    kafka-prod-broker
    Total Samples Count:  6
    Version:              v3
  Conditions:
    Last Transition Time:  2024-08-27T06:21:32Z
    Message:               Successfully created kafkaOpsRequest demo/kfops-kafka-prod-broker-f6qbth
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-08-27T06:21:10Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  kafka
        Lower Bound:
          Cpu:     600m
          Memory:  1536Mi
        Target:
          Cpu:     813m
          Memory:  1536Mi
        Uncapped Target:
          Cpu:     813m
          Memory:  442809964
        Upper Bound:
          Cpu:     1
          Memory:  2Gi
    Vpa Name:      kafka-prod-broker
Events:            <none>
```
So, the `kafkaautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `kafkaopsrequest` based on the recommendations, if the database pods resources are needed to scaled up or down.

Let's watch the `kafkaopsrequest` in the demo namespace to see if any `kafkaopsrequest` object is created. After some time you'll see that a `kafkaopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get kafkaopsrequest -n demo
Every 2.0s: kubectl get kafkaopsrequest -n demo
NAME                                TYPE              STATUS       AGE
kfops-kafka-prod-broker-f6qbth      VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                                 TYPE              STATUS       AGE
kfops-kafka-prod-broker-f6qbth       VerticalScaling   Successful   3m2s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-kafka-prod-broker-f6qbth 
Name:         kfops-kafka-prod-broker-f6qbth
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=kafka-prod
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=kafkas.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-27T06:21:32Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  KafkaAutoscaler
    Name:                  kf-broker-autoscaler
    UID:                   7e3be99f-cd4d-440a-a477-8e8994840ebb
  Resource Version:        1113011
  UID:                     a040a45b-135c-454a-8ddd-d4bd5000ffba
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Broker:
      Resources:
        Limits:
          Memory:  1536Mi
        Requests:
          Cpu:     813m
          Memory:  1536Mi
Status:
  Conditions:
    Last Transition Time:  2024-08-27T06:21:32Z
    Message:               Kafka ops-request has started to vertically scaling the kafka nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-08-27T06:21:35Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-27T06:21:40Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-27T06:21:41Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-27T06:21:55Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-08-27T06:22:00Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-27T06:22:01Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-27T06:22:21Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-08-27T06:22:25Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-08-27T06:22:26Z
    Message:               Successfully completed the vertical scaling for kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                 Age    From                         Message
  ----     ------                                                                 ----   ----                         -------
  Normal   Starting                                                               4m55s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-kafka-prod-broker-f6qbth
  Normal   Starting                                                               4m55s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                             4m55s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-broker-f6qbth
  Normal   UpdatePetSets                                                          4m52s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0             4m47s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0           4m46s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0  4m42s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0   4m32s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1             4m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1           4m26s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1  4m22s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1   4m7s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartPods                                                            4m2s   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                               4m1s   KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                             4m1s   KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-broker-f6qbth
```

Now, we are going to verify from the Pod, and the Kafka yaml whether the resources of the broker node has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo kafka-prod-broker-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1536Mi"
  }
}


$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.broker.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1536Mi"
  }
}
```

## Kafka Topology Autoscaler(Controller)

Let's check the Controller Pod containers resources,

```bash
$ kubectl get pod -n demo kafka-prod-controller-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

Let's check the Kafka resources for broker,
```bash
$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.controller.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see from the above outputs that the resources for controller are same as the one we have assigned while deploying the kafka.

We are now ready to apply the `KafkaAutoscaler` CRO to set up autoscaling for this broker nodes.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a KafkaAutoscaler Object.

#### Create KafkaAutoscaler Object

In order to set up compute resource autoscaling for this topology cluster, we have to create a `KafkaAutoscaler` CRO with our desired configuration. Below is the YAML of the `KafkaAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kf-controller-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: kafka-prod
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    controller:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 1.5Gi
      maxAllowed:
        cpu: 1
        memory: 2Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `kafka-prod` cluster.
- `spec.compute.controller.trigger` specifies that compute autoscaling is enabled for this node.
- `spec.compute.controller.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.controller.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.controller.minAllowed` specifies the minimum allowed resources for the cluster.
- `spec.compute.controller.maxAllowed` specifies the maximum allowed resources for the cluster.
- `spec.compute.controller.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.controller.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields.
  - `timeout` specifies the timeout for the OpsRequest.
  - `apply` specifies when the OpsRequest should be applied. The default is "IfReady".

Let's create the `KafkaAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaler/compute/kafka-controller-autoscaler.yaml
kafkaautoscaler.autoscaling.kubedb.com/kf-controller-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `kafkaautoscaler` resource is created successfully,

```bash
$ kubectl describe kafkaautoscaler kf-controller-autoscaler -n demo
Name:         kf-controller-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         KafkaAutoscaler
Metadata:
  Creation Timestamp:  2024-08-27T06:29:45Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Kafka
    Name:                  kafka-prod
    UID:                   7cee41e0-259c-4a5e-856a-e8ca90056120
  Resource Version:        1116548
  UID:                     49461872-3628-4bc2-8692-f147bc55aa49
Spec:
  Compute:
    Controller:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     1
        Memory:  2Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  1536Mi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  kafka-prod
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              1
        Weight:             10000
        Index:              3
        Weight:             4666
      Reference Timestamp:  2024-08-27T06:30:00Z
      Total Weight:         0.3085514112801626
    First Sample Start:     2024-08-27T06:29:52Z
    Last Sample Start:      2024-08-27T06:30:49Z
    Last Update Time:       2024-08-27T06:31:11Z
    Memory Histogram:
      Reference Timestamp:  2024-08-27T06:35:00Z
    Ref:
      Container Name:     kafka
      Vpa Object Name:    kafka-prod-controller
    Total Samples Count:  3
    Version:              v3
  Conditions:
    Last Transition Time:  2024-08-27T06:30:32Z
    Message:               Successfully created kafkaOpsRequest demo/kfops-kafka-prod-controller-3vlvzr
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-08-27T06:30:11Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  kafka
        Lower Bound:
          Cpu:     600m
          Memory:  1536Mi
        Target:
          Cpu:     600m
          Memory:  1536Mi
        Uncapped Target:
          Cpu:     100m
          Memory:  297164212
        Upper Bound:
          Cpu:     1
          Memory:  2Gi
    Vpa Name:      kafka-prod-controller
Events:            <none>
```
So, the `kafkaautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our controller cluster. Our autoscaler operator continuously watches the recommendation generated and creates an `kafkaopsrequest` based on the recommendations, if the controller node pods resources are needed to scaled up or down.

Let's watch the `kafkaopsrequest` in the demo namespace to see if any `kafkaopsrequest` object is created. After some time you'll see that a `kafkaopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get kafkaopsrequest -n demo
Every 2.0s: kubectl get kafkaopsrequest -n demo
NAME                                    TYPE              STATUS       AGE
kfops-kafka-prod-controller-3vlvzr      VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                                  TYPE              STATUS       AGE
kfops-kafka-prod-controller-3vlvzr    VerticalScaling   Successful   3m2s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-kafka-prod-controller-3vlvzr 
Name:         kfops-kafka-prod-controller-3vlvzr
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=kafka-prod
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=kafkas.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-27T06:30:32Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  KafkaAutoscaler
    Name:                  kf-controller-autoscaler
    UID:                   49461872-3628-4bc2-8692-f147bc55aa49
  Resource Version:        1117285
  UID:                     22228813-bf11-4d8a-9bea-53a1995fe4d0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Controller:
      Resources:
        Limits:
          Memory:  1536Mi
        Requests:
          Cpu:     600m
          Memory:  1536Mi
Status:
  Conditions:
    Last Transition Time:  2024-08-27T06:30:32Z
    Message:               Kafka ops-request has started to vertically scaling the kafka nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-08-27T06:30:35Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-27T06:30:40Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-27T06:30:40Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-27T06:31:11Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-08-27T06:31:15Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-27T06:31:16Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-27T06:31:30Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-08-27T06:31:35Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-08-27T06:31:36Z
    Message:               Successfully completed the vertical scaling for kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   2m33s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-kafka-prod-controller-3vlvzr
  Normal   Starting                                                                   2m33s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 2m33s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-controller-3vlvzr
  Normal   UpdatePetSets                                                              2m30s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             2m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           2m25s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  2m20s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   115s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             110s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           109s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  105s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   95s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Normal   RestartPods                                                                90s    KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                   90s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 90s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-controller-3vlvzr
```

Now, we are going to verify from the Pod, and the Kafka yaml whether the resources of the controller node has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo kafka-prod-controller-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1536Mi"
  }
}


$ kubectl get kafka -n demo kafka-prod -o json | jq '.spec.topology.controller.podTemplate.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1536Mi"
  }
}
```

The above output verifies that we have successfully auto scaled the resources of the Kafka topology cluster for broker and controller. You can create a similar `KafkaAutoscaler` object with both broker and controller resources to auto scale the resources of the Kafka topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequest -n demo kfops-kafka-prod-broker-f6qbth kfops-kafka-prod-controller-3vlvzr
kubectl delete kafkaautoscaler -n demo kf-broker-autoscaler kf-controller-autoscaler
kubectl delete kf -n demo kafka-prod
kubectl delete ns demo
```
## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
