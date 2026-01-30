---
title: Kafka Combined Autoscaling
menu:
  docs_{{ .version }}:
    identifier: kf-auto-scaling-combined
    name: Combined Cluster
    parent: kf-compute-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a Kafka Combined Cluster

This guide will show you how to use `KubeDB` to autoscale compute resources i.e. cpu and memory of a Kafka combined cluster.

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

## Autoscaling of Combined Cluster

Here, we are going to deploy a `Kafka` Combined Cluster using a supported version by `KubeDB` operator. Then we are going to apply `KafkaAutoscaler` to set up autoscaling.

#### Deploy Kafka Combined Cluster

In this section, we are going to deploy a Kafka Topology database with version `3.9.0`.  Then, in the next section we will set up autoscaling for this database using `KafkaAutoscaler` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-dev
  namespace: demo
spec:
  replicas: 2
  version: 4.0.0
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaler/kafka-combined.yaml
kafka.kubedb.com/kafka-dev created
```

Now, wait until `kafka-dev` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME         TYPE            VERSION   STATUS         AGE
kafka-dev    kubedb.com/v1   3.9.0     Provisioning   0s
kafka-dev    kubedb.com/v1   3.9.0     Provisioning   24s
.
.
kafka-dev    kubedb.com/v1   3.9.0     Ready          92s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo kafka-dev-0 -o json | jq '.spec.containers[].resources'
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

Let's check the Kafka resources,
```bash
$ kubectl get kafka -n demo kafka-dev -o json | jq '.spec.podTemplate.spec.containers[].resources'
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

You can see from the above outputs that the resources are same as the one we have assigned while deploying the kafka.

We are now ready to apply the `KafkaAutoscaler` CRO to set up autoscaling for this database.

### Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a KafkaAutoscaler Object.

#### Create KafkaAutoscaler Object

In order to set up compute resource autoscaling for this combined cluster, we have to create a `KafkaAutoscaler` CRO with our desired configuration. Below is the YAML of the `KafkaAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kf-combined-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: kafka-dev
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    node:
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

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on `kafka-dev` cluster.
- `spec.compute.node.trigger` specifies that compute autoscaling is enabled for this cluster.
- `spec.compute.node.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pod to initiate a vertical scaling.
- `spec.compute.node.resourceDiffPercentage` specifies the minimum resource difference in percentage. The default is 10%.
  If the difference between current & recommended resource is less than ResourceDiffPercentage, Autoscaler Operator will ignore the updating.
- `spec.compute.node.minAllowed` specifies the minimum allowed resources for the cluster.
- `spec.compute.node.maxAllowed` specifies the maximum allowed resources for the cluster.
- `spec.compute.node.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.node.containerControlledValues` specifies which resource values should be controlled. The default is "RequestsAndLimits".
- `spec.opsRequestOptions` contains the options to pass to the created OpsRequest. It has 2 fields.
  - `timeout` specifies the timeout for the OpsRequest.
  - `apply` specifies when the OpsRequest should be applied. The default is "IfReady".

Let's create the `KafkaAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaler/compute/kafka-combined-autoscaler.yaml
kafkaautoscaler.autoscaling.kubedb.com/kf-combined-autoscaler created
```

#### Verify Autoscaling is set up successfully

Let's check that the `kafkaautoscaler` resource is created successfully,

```bash
$ kubectl describe kafkaautoscaler kf-combined-autoscaler -n demo
Name:         kf-combined-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         KafkaAutoscaler
Metadata:
  Creation Timestamp:  2024-08-27T05:55:51Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Kafka
    Name:                  kafka-dev
    UID:                   a0153c7f-1e1e-4070-a318-c7c1153b810a
  Resource Version:        1104655
  UID:                     817602cc-f851-4fc5-b2c1-1d191462ac56
Spec:
  Compute:
    Node:
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
    Name:  kafka-dev
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             4610
        Index:              1
        Weight:             10000
      Reference Timestamp:  2024-08-27T05:55:00Z
      Total Weight:         0.35081120875606336
    First Sample Start:     2024-08-27T05:55:44Z
    Last Sample Start:      2024-08-27T05:56:49Z
    Last Update Time:       2024-08-27T05:57:10Z
    Memory Histogram:
      Reference Timestamp:  2024-08-27T06:00:00Z
    Ref:
      Container Name:     kafka
      Vpa Object Name:    kafka-dev
    Total Samples Count:  3
    Version:              v3
  Conditions:
    Last Transition Time:  2024-08-27T05:56:32Z
    Message:               Successfully created kafkaOpsRequest demo/kfops-kafka-dev-z8d3l5
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2024-08-27T05:56:10Z
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
          Memory:  511772986
        Upper Bound:
          Cpu:     1
          Memory:  2Gi
    Vpa Name:      kafka-dev
Events:            <none>
```
So, the `kafkaautoscaler` resource is created successfully.

you can see in the `Status.VPAs.Recommendation` section, that recommendation has been generated for our database. Our autoscaler operator continuously watches the recommendation generated and creates an `kafkaopsrequest` based on the recommendations, if the database pods resources are needed to scaled up or down.

Let's watch the `kafkaopsrequest` in the demo namespace to see if any `kafkaopsrequest` object is created. After some time you'll see that a `kafkaopsrequest` will be created based on the recommendation.

```bash
$ watch kubectl get kafkaopsrequest -n demo
Every 2.0s: kubectl get kafkaopsrequest -n demo
NAME                         TYPE              STATUS       AGE
kfops-kafka-dev-z8d3l5       VerticalScaling   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                         TYPE              STATUS       AGE
kfops-kafka-dev-z8d3l5       VerticalScaling   Successful   3m2s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-kafka-dev-z8d3l5 
Name:         kfops-kafka-dev-z8d3l5
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=kafka-dev
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=kafkas.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-27T05:56:32Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  KafkaAutoscaler
    Name:                  kf-combined-autoscaler
    UID:                   817602cc-f851-4fc5-b2c1-1d191462ac56
  Resource Version:        1104871
  UID:                     8b7615c6-d38b-4d5a-b733-6aa93cd41a29
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-dev
  Timeout:  5m0s
  Type:     VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Memory:  1536Mi
        Requests:
          Cpu:     600m
          Memory:  1536Mi
Status:
  Conditions:
    Last Transition Time:  2024-08-27T05:56:32Z
    Message:               Kafka ops-request has started to vertically scaling the kafka nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-08-27T05:56:35Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-27T05:56:40Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-0
    Last Transition Time:  2024-08-27T05:56:40Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-0
    Last Transition Time:  2024-08-27T05:57:10Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-0
    Last Transition Time:  2024-08-27T05:57:15Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-1
    Last Transition Time:  2024-08-27T05:57:16Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-1
    Last Transition Time:  2024-08-27T05:57:25Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-1
    Last Transition Time:  2024-08-27T05:57:30Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-08-27T05:57:30Z
    Message:               Successfully completed the vertical scaling for kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age    From                         Message
  ----     ------                                                         ----   ----                         -------
  Normal   Starting                                                       4m33s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-kafka-dev-z8d3l5
  Normal   Starting                                                       4m33s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-dev
  Normal   Successful                                                     4m33s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-kafka-dev-z8d3l5
  Normal   UpdatePetSets                                                  4m30s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-0             4m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-0           4m25s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-0  4m19s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-0   3m55s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-1             3m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-1           3m49s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-1  3m45s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-1   3m40s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-1
  Normal   RestartPods                                                    3m35s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                       3m35s  KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-dev
  Normal   Successful                                                     3m35s  KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-kafka-dev-z8d3l5
```

Now, we are going to verify from the Pod, and the Kafka yaml whether the resources of the topology database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo kafka-dev-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1536Mi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1536Mi"
  }
}


$ kubectl get kafka -n demo kafka-dev -o json | jq '.spec.podTemplate.spec.containers[].resources'
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


The above output verifies that we have successfully auto scaled the resources of the Kafka combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequest -n demo kfops-kafka-dev-z8d3l5
kubectl delete kafkaautoscaler -n demo kf-combined-autoscaler
kubectl delete kf -n demo kafka-dev
kubectl delete ns demo
```
## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
