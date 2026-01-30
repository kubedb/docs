---
title: Restart Kafka
menu:
  docs_{{ .version }}:
    identifier: kf-restart-details
    name: Restart Kafka
    parent: kf-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Kafka

KubeDB supports restarting the Kafka database via a KafkaOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/kafka](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Kafka

In this section, we are going to deploy a Kafka database using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 4.0.0
  topology:
    broker:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: kafka
            resources:
              requests:
                cpu: "500m"
                memory: "1Gi"
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
              requests:
                  cpu: "500m"
                  memory: "1Gi"
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  deletionPolicy: DoNotTerminate
```

- `spec.topology` represents the specification for kafka topology.
    - `broker` denotes the broker node of kafka topology.
    - `controller` denotes the controller node of kafka topology.

Let's create the `Kafka` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/restart/kafka.yaml
kafka.kubedb.com/kafka-prod created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: kafka-prod
  timeout: 5m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Kafka CR. It should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/kafka/concepts/kafkaopsrequest.md#spectimeout)

> Note: The method of restarting the combined node is exactly same as above. All you need, is to specify the corresponding Kafka name in `spec.databaseRef.name` section.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/restart/ops.yaml
kafkaopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the controller pods, then broker of the referenced kafka.

```shell
$ kubectl get kfops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   119s

$ kubectl get kfops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"KafkaOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"kafka-prod"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-07-26T10:12:10Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "24434"
  uid: 956a374e-1d6f-4f68-828f-cfed4410b175
spec:
  apply: Always
  databaseRef:
    name: kafka-prod
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-07-26T10:12:10Z"
    message: Kafka ops-request has started to restart kafka nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-07-26T10:12:18Z"
    message: get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    observedGeneration: 1
    status: "True"
    type: GetPod--kafka-prod-controller-0
  - lastTransitionTime: "2024-07-26T10:12:18Z"
    message: evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--kafka-prod-controller-0
  - lastTransitionTime: "2024-07-26T10:12:23Z"
    message: check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--kafka-prod-controller-0
  - lastTransitionTime: "2024-07-26T10:12:28Z"
    message: get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    observedGeneration: 1
    status: "True"
    type: GetPod--kafka-prod-controller-1
  - lastTransitionTime: "2024-07-26T10:12:28Z"
    message: evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--kafka-prod-controller-1
  - lastTransitionTime: "2024-07-26T10:12:38Z"
    message: check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--kafka-prod-controller-1
  - lastTransitionTime: "2024-07-26T10:12:43Z"
    message: get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    observedGeneration: 1
    status: "True"
    type: GetPod--kafka-prod-broker-0
  - lastTransitionTime: "2024-07-26T10:12:43Z"
    message: evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--kafka-prod-broker-0
  - lastTransitionTime: "2024-07-26T10:13:18Z"
    message: check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--kafka-prod-broker-0
  - lastTransitionTime: "2024-07-26T10:13:23Z"
    message: get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    observedGeneration: 1
    status: "True"
    type: GetPod--kafka-prod-broker-1
  - lastTransitionTime: "2024-07-26T10:13:23Z"
    message: evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--kafka-prod-broker-1
  - lastTransitionTime: "2024-07-26T10:13:28Z"
    message: check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--kafka-prod-broker-1
  - lastTransitionTime: "2024-07-26T10:13:33Z"
    message: Successfully Restarted Kafka nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-07-26T10:13:33Z"
    message: Controller has successfully restart the Kafka replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequest -n demo restart
kubectl delete kafka -n demo kafka-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
