---
title: Reconfigure Kafka Topology
menu:
  docs_{{ .version }}:
    identifier: kf-reconfigure-topology
    name: Topology
    parent: kf-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Kafka Topology Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a Kafka Topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [Topology](/docs/guides/kafka/clustering/topology-cluster/index.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Reconfigure Overview](/docs/guides/kafka/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `Kafka` Topology cluster using a supported version by `KubeDB` operator. Then we are going to apply `KafkaOpsRequest` to reconfigure its configuration.

### Prepare Kafka Topology Cluster

Now, we are going to deploy a `Kafka` topology cluster with version `3.9.0`.

### Deploy Kafka

At first, we will create a secret with the `broker.properties` and `controller.properties` file containing required configuration settings.

**broker.properties:**

```properties
log.retention.hours=100
```

**controller.properties:**

```properties
controller.quorum.election.timeout.ms=2000
```

Here, `log.retention.hours` is set to `100`, whereas the default value is `168` for broker and `controller.quorum.election.timeout.ms` is set to `2000` for controller.

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kf-topology-custom-config
  namespace: demo
stringData:
  broker.properties: |-
    log.retention.hours=100
  controller.properties: |-
    controller.quorum.election.timeout.ms=2000
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-topology-custom-config.yaml
secret/kf-topology-custom-config created
```

> **Note:** 

In this section, we are going to create a Kafka object specifying `spec.configuration.secretName` field to apply this custom configuration. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.9.0
  configuration:
    secretName: kf-topology-custom-config
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-topology.yaml
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

Now, we will check if the kafka has started with the custom configuration we have provided.

Exec into the Kafka pod and execute the following commands to see the configurations:
```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- bash
kafka@kafka-prod-broker-0:~$ kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep log.retention.hours
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
```
Here, we can see that our given configuration is applied to the Kafka cluster for all brokers. `log.retention.hours` is set to `100` from the default value `168`.

### Reconfigure using new config secret

Now we will reconfigure this cluster to set `log.retention.hours` to `125`.

Now, update our `broker.properties` and `controller.properties` file with the new configuration.

**broker.properties:**

```properties
log.retention.hours=125
```

**controller.properties:**

```properties
controller.quorum.election.timeout.ms=3000
controller.quorum.fetch.timeout.ms=4000
```

Then, we will create a new secret with this configuration file.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-kf-topology-custom-config
  namespace: demo
stringData:
  broker.properties: |-
    log.retention.hours=125
  controller.properties: |-
    controller.quorum.election.timeout.ms=3000
    controller.quorum.fetch.timeout.ms=4000
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/new-kafka-topology-custom-config.yaml
secret/new-kf-topology-custom-config created
```

#### Create KafkaOpsRequest

Now, we will use this secret to replace the previous secret using a `KafkaOpsRequest` CR. The `KafkaOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfigure-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-prod
  configuration:
    configSecret:
      name: new-kf-topology-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `kafka-prod` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-reconfigure-update-topology.yaml
kafkaopsrequest.ops.kubedb.com/kfops-reconfigure-topology created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configuration` of `Kafka` object.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequests -n demo 
NAME                         TYPE          STATUS       AGE
kfops-reconfigure-topology   Reconfigure   Successful   4m55s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-reconfigure-topology
Name:         kfops-reconfigure-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T05:08:37Z
  Generation:          1
  Resource Version:    332491
  UID:                 b6e8cb1b-d29f-445e-bb01-60d29012c7eb
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-kf-topology-custom-config
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-08-02T05:08:37Z
    Message:               Kafka ops-request has started to reconfigure kafka nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-08-02T05:08:45Z
    Message:               check reconcile; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  CheckReconcile
    Last Transition Time:  2024-08-02T05:09:42Z
    Message:               successfully reconciled the Kafka with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-02T05:09:47Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T05:09:47Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T05:10:02Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T05:10:07Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T05:10:07Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T05:10:22Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T05:10:27Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T05:10:27Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T05:11:12Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T05:11:17Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T05:11:17Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T05:11:32Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T05:11:37Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T05:11:39Z
    Message:               Successfully completed reconfigure kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   3m7s   KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-reconfigure-topology
  Normal   Starting                                                                   3m7s   KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 3m7s   KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-reconfigure-topology
  Warning  check reconcile; ConditionStatus:False                                     2m59s  KubeDB Ops-manager Operator  check reconcile; ConditionStatus:False
  Normal   UpdatePetSets                                                              2m2s   KubeDB Ops-manager Operator  successfully reconciled the Kafka with new configure
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             117s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           117s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  112s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   102s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             97s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           97s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  92s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   82s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 77s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               77s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      72s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       32s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 27s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               27s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      22s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       12s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               7s     KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   5s     KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 5s     KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-reconfigure-topology
```

Now let's exec one of the instance and run a kafka-configs.sh command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'log.retention.hours'
  log.retention.hours=125 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=125, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=125 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=125, DEFAULT_CONFIG:log.retention.hours=168}
```

As we can see from the configuration of ready kafka, the value of `log.retention.hours` has been changed from `100` to `125`. So the reconfiguration of the cluster is successful.


### Reconfigure using apply config

Now we will reconfigure this cluster again to set `log.retention.hours` to `150`. This time we won't use a new secret. We will use the `applyConfig` field of the `KafkaOpsRequest`. This will merge the new config in the existing secret.

#### Create KafkaOpsRequest

Now, we will use the new configuration in the `applyConfig` field in the `KafkaOpsRequest` CR. The `KafkaOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfigure-apply-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-prod
  configuration:
    applyConfig:
      broker.properties: |-
        log.retention.hours=150
      controller.properties: |-
        controller.quorum.election.timeout.ms=4000
        controller.quorum.fetch.timeout.ms=5000
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `kafka-prod` cluster.
- `spec.type` specifies that we are performing `Reconfigure` on kafka.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-reconfigure-apply-topology.yaml
kafkaopsrequest.ops.kubedb.com/kfops-reconfigure-apply-topology created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequests -n demo kfops-reconfigure-apply-topology 
NAME                               TYPE          STATUS       AGE
kfops-reconfigure-apply-topology   Reconfigure   Successful   55s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to reconfigure the cluster.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-reconfigure-apply-topology 
Name:         kfops-reconfigure-apply-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-02T05:14:42Z
  Generation:          1
  Resource Version:    332996
  UID:                 551d2c92-9431-47a7-a699-8f8115131b49
Spec:
  Apply:  IfReady
  Configuration:
    Apply Config:
      broker.properties:      log.retention.hours=150
      controller.properties:  controller.quorum.election.timeout.ms=4000
controller.quorum.fetch.timeout.ms=5000
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-08-02T05:14:42Z
    Message:               Kafka ops-request has started to reconfigure kafka nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-08-02T05:14:45Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2024-08-02T05:14:52Z
    Message:               successfully reconciled the Kafka with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-02T05:14:57Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T05:14:57Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T05:15:07Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-08-02T05:15:12Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T05:15:12Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T05:15:27Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-08-02T05:15:32Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T05:15:32Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T05:16:07Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-08-02T05:16:12Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T05:16:12Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T05:16:27Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-08-02T05:16:32Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-02T05:16:35Z
    Message:               Successfully completed reconfigure kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age   From                         Message
  ----     ------                                                                     ----  ----                         -------
  Normal   Starting                                                                   2m6s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-reconfigure-apply-topology
  Normal   Starting                                                                   2m6s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 2m6s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-reconfigure-apply-topology
  Normal   UpdatePetSets                                                              116s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with new configure
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             111s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           111s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  106s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   101s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             96s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           96s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  91s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   81s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 76s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               76s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      71s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       41s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 36s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               36s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      31s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       21s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               15s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   14s   KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 14s   KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-reconfigure-apply-topology
```

Now let's exec into one of the instance and run a `kafka-configs.sh` command to check the new configuration we have provided.

```bash
$ $ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'log.retention.hours'
  log.retention.hours=150 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=150, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=150 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=150, DEFAULT_CONFIG:log.retention.hours=168}
```

As we can see from the configuration of ready kafka, the value of `log.retention.hours` has been changed from `125` to `150`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kf -n demo kafka-dev
kubectl delete kafkaopsrequest -n demo kfops-reconfigure-apply-topology kfops-reconfigure-topology
kubectl delete secret -n demo kf-topology-custom-config new-kf-topology-custom-config
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
