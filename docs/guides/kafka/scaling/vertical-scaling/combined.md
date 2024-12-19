---
title: Vertical Scaling Kafka Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: kf-vertical-scaling-combined
    name: Combined Cluster
    parent: kf-vertical-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Kafka Combined Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a Kafka combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [Combined](/docs/guides/kafka/clustering/combined-cluster/index.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/kafka/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on Combined Cluster

Here, we are going to deploy a `Kafka` combined cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare Kafka Combined Cluster

Now, we are going to deploy a `Kafka` combined cluster database with version `3.9.0`.

### Deploy Kafka Combined Cluster

In this section, we are going to deploy a Kafka combined cluster. Then, in the next section we will update the resources of the database using `KafkaOpsRequest` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-dev
  namespace: demo
spec:
  replicas: 2
  version: 3.9.0
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/scaling/kafka-combined.yaml
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
This is the default resources of the Kafka combined cluster set by the `KubeDB` operator.

We are now ready to apply the `KafkaOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the combined cluster to meet the desired resources after scaling.

#### Create KafkaOpsRequest

In order to update the resources of the database, we have to create a `KafkaOpsRequest` CR with our desired resources. Below is the YAML of the `KafkaOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-vscale-combined
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: kafka-dev
  verticalScaling:
    node:
      resources:
        requests:
          memory: "1.2Gi"
          cpu: "0.6"
        limits:
          memory: "1.2Gi"
          cpu: "0.6"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `kafka-dev` cluster.
- `spec.type` specifies that we are performing `VerticalScaling` on kafka.
- `spec.VerticalScaling.node` specifies the desired resources after scaling.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/scaling/vertical-scaling/kafka-vertical-scaling-combined.yaml
kafkaopsrequest.ops.kubedb.com/kfops-vscale-combined created
```

#### Verify Kafka Combined cluster resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `Kafka` object and related `PetSets` and `Pods`.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                     TYPE              STATUS       AGE
kfops-vscale-combined    VerticalScaling   Successful   3m56s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to scale the cluster.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-vscale-combined 
Name:         kfops-vscale-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T05:59:06Z
  Generation:          1
  Resource Version:    336197
  UID:                 5fd90feb-eed2-4130-8762-442f2f4d2698
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-dev
  Timeout:  5m
  Type:     VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     0.6
          Memory:  1.2Gi
        Requests:
          Cpu:     0.6
          Memory:  1.2Gi
Status:
  Conditions:
    Last Transition Time:  2024-08-02T05:59:06Z
    Message:               Kafka ops-request has started to vertically scaling the kafka nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-08-02T05:59:09Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-02T05:59:14Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-0
    Last Transition Time:  2024-08-02T05:59:14Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-0
    Last Transition Time:  2024-08-02T05:59:29Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-0
    Last Transition Time:  2024-08-02T05:59:34Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-1
    Last Transition Time:  2024-08-02T05:59:34Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-1
    Last Transition Time:  2024-08-02T06:00:59Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-1
    Last Transition Time:  2024-08-02T06:01:04Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-08-02T06:01:04Z
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
  Normal   Starting                                                       2m38s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-vscale-combined
  Normal   Starting                                                       2m38s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-dev
  Normal   Successful                                                     2m38s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-vscale-combined
  Normal   UpdatePetSets                                                  2m35s  KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-0             2m30s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-0           2m30s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-0  2m25s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-0   2m15s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-1             2m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-1           2m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-1  2m5s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-1   45s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-1
  Normal   RestartPods                                                    40s    KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                       40s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-dev
  Normal   Successful                                                     40s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-vscale-combined
```

Now, we are going to verify from one of the Pod yaml whether the resources of the combined cluster has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo kafka-dev-1 -o json | jq '.spec.containers[].resources'
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
```

The above output verifies that we have successfully scaled up the resources of the Kafka combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo kafka-dev
kubectl delete kafkaopsrequest -n demo kfops-vscale-combined
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
