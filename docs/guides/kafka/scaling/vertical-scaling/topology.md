---
title: Vertical Scaling Kafka Topology Cluster
menu:
  docs_{{ .version }}:
    identifier: kf-vertical-scaling-topology
    name: Topology Cluster
    parent: kf-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Kafka Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Kafka topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [Topology](/docs/guides/kafka/clustering/topology-cluster/index.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/kafka/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Topology Cluster

Here, we are going to deploy a `Kafka` topology cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Kafka Topology Cluster

Now, we are going to deploy a `Kafka` topology cluster database with version `3.9.0`.

### Deploy Kafka Topology Cluster

In this section, we are going to deploy a Kafka topology cluster. Then, in the next section we will update the resources of the database using `KafkaOpsRequest` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 2
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

Let's create the `Kafka` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/scaling/kafka-topology.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait until `kafka-prod` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   0s
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   24s
.
.
kafka-prod    kubedb.com/v1   3.9.0     Ready          92s
```

Let's check the Pod containers resources for both `broker` and `controller` of the Kafka topology cluster. Run the following command to get the resources of the `broker` and `controller` containers of the Kafka topology cluster

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
This is the default resources of the Kafka topology cluster set by the `KubeDB` operator.

We are now ready to apply the `KafkaOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the topology cluster to meet the desired resources after scaling.

#### Create KafkaOpsRequest

In order to update the resources of the database, we have to create a `KafkaOpsRequest` CR with our desired resources. Below is the YAML of the `KafkaOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-vscale-topology
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: kafka-prod
  verticalScaling:
    broker:
      resources:
        requests:
          memory: "1.2Gi"
          cpu: "0.6"
        limits:
          memory: "1.2Gi"
          cpu: "0.6"
    controller:
      resources:
        requests:
          memory: "1.1Gi"
          cpu: "0.6"
        limits:
          memory: "1.1Gi"
          cpu: "0.6"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on kafka.
- `spec.VerticalScaling.node` specifies the desired resources after scaling.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/scaling/vertical-scaling/kafka-vertical-scaling-topology.yaml
kafkaopsrequest.ops.kubedb.com/kfops-vscale-topology created
```

#### Verify Kafka Topology cluster resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Kafka` object and related `PetSets` and `Pods`.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                     TYPE              STATUS       AGE
kfops-vscale-topology    VerticalScaling   Successful   3m56s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-vscale-topology
Name:         kfops-vscale-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T06:09:46Z
  Generation:          1
  Resource Version:    337300
  UID:                 ca298c0a-e08d-4c78-acbc-40eb5e96532d
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m
  Type:     VerticalScaling
  Vertical Scaling:
    Broker:
      Resources:
        Limits:
          Cpu:     0.6
          Memory:  1.2Gi
        Requests:
          Cpu:     0.6
          Memory:  1.2Gi
    Controller:
      Resources:
        Limits:
          Cpu:     0.6
          Memory:  1.1Gi
        Requests:
          Cpu:     0.6
          Memory:  1.1Gi
Status:
  Conditions:
    Last Transition Time:  2024-08-02T06:09:46Z
    Message:               Kafka ops-request has started to vertically scaling the kafka nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-08-02T06:09:50Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-02T06:09:55Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T06:09:55Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T06:10:00Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T06:10:05Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T06:10:05Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T06:10:15Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T06:10:20Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T06:10:20Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T06:10:35Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T06:10:40Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T06:10:40Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T06:10:55Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T06:11:00Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-08-02T06:11:00Z
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
  Normal   Starting                                                                   3m32s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-vscale-topology
  Normal   Starting                                                                   3m32s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 3m32s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-vscale-topology
  Normal   UpdatePetSets                                                              3m28s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 3m23s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               3m23s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       3m18s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 3m13s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               3m13s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      3m8s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       3m3s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             2m58s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           2m58s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  2m53s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   2m43s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             2m38s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           2m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  2m33s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   2m23s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Normal   RestartPods                                                                2m18s  KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                                   2m18s  KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 2m18s  KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-vscale-topology
```
Now, we are going to verify from one of the Pod yaml whether the resources of the topology cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo kafka-prod-broker-1 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1288490188800m"
  }
}
$ kubectl get pod -n demo kafka-prod-controller-1 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1181116006400m"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1181116006400m"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the Kafka topology cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kf -n demo kafka-prod
kubectl delete kafkaopsrequest -n demo kfops-vscale-topology
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
