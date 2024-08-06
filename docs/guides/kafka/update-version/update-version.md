---
title: Update Version of Kafka
menu:
  docs_{{ .version }}:
    identifier: kf-update-version-kafka
    name: Kafka
    parent: kf-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of Kafka

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `Kafka` Combined or Topology.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Updating Overview](/docs/guides/kafka/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare Kafka

Now, we are going to deploy a `Kafka` replicaset database with version `3.6.8`.

### Deploy Kafka

In this section, we are going to deploy a Kafka topology cluster. Then, in the next section we will update the version using `KafkaOpsRequest` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.6.1
  configSecret:
    name: configsecret-topology
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
  deletionPolicy: WipeOut
```

Let's create the `Kafka` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/update-version/kafka.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait until `kafka-prod` created has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w                                                                                                                                           
NAME         TYPE            VERSION   STATUS         AGE
kafka-prod   kubedb.com/v1   3.5.2     Provisioning   0s
kafka-prod   kubedb.com/v1   3.5.2     Provisioning   55s
.
.
kafka-prod   kubedb.com/v1   3.5.2     Ready          119s
```

We are now ready to apply the `KafkaOpsRequest` CR to update.

### update Kafka Version

Here, we are going to update `Kafka` from `3.5.2` to `3.6.1`.

#### Create KafkaOpsRequest:

In order to update the version, we have to create a `KafkaOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `KafkaOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kafka-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: kafka-prod
  updateVersion:
    targetVersion: 3.6.1
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `kafka-prod` Kafka.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `3.6.1`.

> **Note:** If you want to update combined Kafka, you just refer to the `Kafka` combined object name in `spec.databaseRef.name`. To create a combined Kafka, you can refer to the [Kafka Combined](/docs/guides/kafka/clustering/combined-cluster/index.md) guide.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/update-version/update-version.yaml
kafkaopsrequest.ops.kubedb.com/kafka-update-version created
```

#### Verify Kafka version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `Kafka` object and related `PetSets` and `Pods`.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                   TYPE            STATUS        AGE
kafka-update-version   UpdateVersion   Successful    2m6s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe kafkaopsrequest -n demo kafka-update-version
Name:         kafka-update-version
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-07-30T10:18:44Z
  Generation:          1
  Resource Version:    90131
  UID:                 a274197b-c379-485b-9a36-9eb1e673eee4
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m
  Type:     UpdateVersion
  Update Version:
    Target Version:  3.6.1
Status:
  Conditions:
    Last Transition Time:  2024-07-30T10:18:44Z
    Message:               Kafka ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2024-07-30T10:18:54Z
    Message:               successfully reconciled the Kafka with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-30T10:18:59Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-30T10:18:59Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-30T10:19:19Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-07-30T10:19:24Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-30T10:19:24Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-30T10:19:49Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-07-30T10:19:54Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-30T10:19:54Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-30T10:20:14Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-07-30T10:20:19Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-30T10:20:19Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-30T10:20:44Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-07-30T10:20:49Z
    Message:               Successfully Restarted Kafka nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-30T10:20:50Z
    Message:               Successfully completed update kafka version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   3m7s   KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kafka-update-version
  Normal   Starting                                                                   3m7s   KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 3m7s   KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kafka-update-version
  Normal   UpdatePetSets                                                              2m57s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with updated version
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             2m52s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           2m52s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  2m47s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   2m32s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             2m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           2m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  2m22s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   2m2s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 117s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               117s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      112s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       97s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 92s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               92s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      87s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       67s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartPods                                                                62s    KubeDB Ops-manager Operator  Successfully Restarted Kafka nodes
  Normal   Starting                                                                   62s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 61s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kafka-update-version
```

Now, we are going to verify whether the `Kafka` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get kf -n demo kafka-prod -o=jsonpath='{.spec.version}{"\n"}'
3.6.1

$ kubectl get petset -n demo kafka-prod-broker -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/kafka-kraft:3.6.1@sha256:e251d3c0ceee0db8400b689e42587985034852a8a6c81b5973c2844e902e6d11

$ kubectl get pods -n demo kafka-prod-broker-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/kafka-kraft:3.6.1@sha256:e251d3c0ceee0db8400b689e42587985034852a8a6c81b5973c2844e902e6d11
```

You can see from above, our `Kafka` has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequest -n demo kafka-update-version
kubectl delete kf -n demo kafka-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
