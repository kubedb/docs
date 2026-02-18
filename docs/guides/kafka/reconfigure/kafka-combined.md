---
title: Reconfigure Kafka Combined
menu:
  docs_{{ .version }}:
    identifier: kf-reconfigure-combined
    name: Combined
    parent: kf-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Kafka Combined Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a Kafka Combined cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [Combined](/docs/guides/kafka/clustering/combined-cluster/index.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Reconfigure Overview](/docs/guides/kafka/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `Kafka` Combined cluster using a supported version by `KubeDB` operator. Then we are going to apply `KafkaOpsRequest` to reconfigure its configuration.

### Prepare Kafka Combined Cluster

Now, we are going to deploy a `Kafka` combined cluster with version `3.9.0`.

### Deploy Kafka

At first, we will create a secret with the `server.properties` file containing required configuration settings.

**server.properties:**

```properties
log.retention.hours=100
```
Here, `log.retention.hours` is set to `100`, whereas the default value is `168`.

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kf-combined-custom-config
  namespace: demo
stringData:
  server.properties: |-
    log.retention.hours=100
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-combined-custom-config.yaml
secret/kf-combined-custom-config created
```

In this section, we are going to create a Kafka object specifying `spec.configuration` field to apply this custom configuration. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-dev
  namespace: demo
spec:
  replicas: 2
  version: 3.9.0
  configuration:
    secretName: kf-combined-custom-config
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-combined.yaml
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

Now, we will check if the kafka has started with the custom configuration we have provided.

Exec into the Kafka pod and execute the following commands to see the configurations:
```bash
$ kubectl exec -it -n demo kafka-dev-0 -- bash
kafka@kafka-dev-0:~$ kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep log.retention.hours
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
```
Here, we can see that our given configuration is applied to the Kafka cluster for all brokers. `log.retention.hours` is set to `100` from the default value `168`.

### Reconfigure using new config secret

Now we will reconfigure this cluster to set `log.retention.hours` to `125`.

Now, update our `server.properties` file with the new configuration.

**server.properties:**

```properties
log.retention.hours=125
```

Then, we will create a new secret with this configuration file.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-kf-combined-custom-config
  namespace: demo
stringData:
  server.properties: |-
    log.retention.hours=125
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/new-kafka-combined-custom-config.yaml
secret/new-kf-combined-custom-config created
```

#### Create KafkaOpsRequest

Now, we will use this secret to replace the previous secret using a `KafkaOpsRequest` CR. The `KafkaOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfigure-combined
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-dev
  configuration:
    configSecret:
      name: new-kf-combined-custom-config
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `kafka-dev` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-reconfigure-update-combined.yaml
kafkaopsrequest.ops.kubedb.com/kfops-reconfigure-combined created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `.spec.configuration` of `Kafka` object.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequests -n demo 
NAME                         TYPE          STATUS       AGE
kfops-reconfigure-combined   Reconfigure   Successful   4m55s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-reconfigure-combined
Name:         kfops-reconfigure-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-01T09:14:46Z
  Generation:          1
  Resource Version:    258361
  UID:                 ac2147ba-51cf-4ebf-8328-76253379108c
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-kf-combined-custom-config
  Database Ref:
    Name:   kafka-dev
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-08-01T09:14:46Z
    Message:               Kafka ops-request has started to reconfigure kafka nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-08-01T09:14:55Z
    Message:               successfully reconciled the Kafka with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-01T09:15:00Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-0
    Last Transition Time:  2024-08-01T09:15:00Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-0
    Last Transition Time:  2024-08-01T09:16:15Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-0
    Last Transition Time:  2024-08-01T09:16:20Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-1
    Last Transition Time:  2024-08-01T09:16:20Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-1
    Last Transition Time:  2024-08-01T09:17:20Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-1
    Last Transition Time:  2024-08-01T09:17:25Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-01T09:17:25Z
    Message:               Successfully completed reconfigure kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age    From                         Message
  ----     ------                                                         ----   ----                         -------
  Normal   Starting                                                       5m32s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-reconfigure-combined
  Normal   Starting                                                       5m32s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-dev
  Normal   Successful                                                     5m32s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-reconfigure-combined
  Normal   UpdatePetSets                                                  5m23s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with new configure
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-0             5m18s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-0           5m18s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-0  5m13s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-0   4m3s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-1             3m58s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-1           3m58s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-1  3m53s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-1   2m58s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-1
  Normal   RestartNodes                                                   2m53s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                       2m53s  KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-dev
  Normal   Successful                                                     2m53s  KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-reconfigure-combined
```

Now let's exec one of the instance and run a kafka-configs.sh command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo kafka-dev-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'log.retention.hours'
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
  name: kfops-reconfigure-apply-combined
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-dev
  configuration:
    applyConfig:
      server.properties: |-
        log.retention.hours=150
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `kafka-dev` cluster.
- `spec.type` specifies that we are performing `Reconfigure` on kafka.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure/kafka-reconfigure-apply-combined.yaml
kafkaopsrequest.ops.kubedb.com/kfops-reconfigure-apply-combined created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequests -n demo kfops-reconfigure-apply-combined 
NAME                               TYPE          STATUS       AGE
kfops-reconfigure-apply-combined   Reconfigure   Successful   55s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to reconfigure the cluster.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-reconfigure-apply-combined
Name:         kfops-reconfigure-apply-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-01T09:27:03Z
  Generation:          1
  Resource Version:    259123
  UID:                 fdc46ef0-e2ae-490a-aab8-6a3380ec09d1
Spec:
  Apply:  IfReady
  Configuration:
    Apply Config:
      server.properties:  log.retention.hours=150
  Database Ref:
    Name:   kafka-dev
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-08-01T09:27:03Z
    Message:               Kafka ops-request has started to reconfigure kafka nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-08-01T09:27:06Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2024-08-01T09:27:12Z
    Message:               successfully reconciled the Kafka with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-01T09:27:17Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-0
    Last Transition Time:  2024-08-01T09:27:17Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-0
    Last Transition Time:  2024-08-01T09:27:27Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-0
    Last Transition Time:  2024-08-01T09:27:32Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-dev-1
    Last Transition Time:  2024-08-01T09:27:32Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-dev-1
    Last Transition Time:  2024-08-01T09:27:52Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-dev-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-dev-1
    Last Transition Time:  2024-08-01T09:27:57Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-08-01T09:27:57Z
    Message:               Successfully completed reconfigure kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                         Age   From                         Message
  ----     ------                                                         ----  ----                         -------
  Normal   Starting                                                       2m7s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-reconfigure-apply-combined
  Normal   Starting                                                       2m7s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-dev
  Normal   Successful                                                     2m7s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-reconfigure-apply-combined
  Normal   UpdatePetSets                                                  118s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with new configure
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-0             113s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-0           113s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-0  108s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-0   103s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-dev-1             98s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-dev-1           98s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-dev-1  93s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-dev-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-dev-1   78s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-dev-1
  Normal   RestartNodes                                                   73s   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                       73s   KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-dev
  Normal   Successful                                                     73s   KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-reconfigure-apply-combined
```

Now let's exec into one of the instance and run a `kafka-configs.sh` command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo kafka-dev-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'log.retention.hours'
  log.retention.hours=150 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=150, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=150 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=150, DEFAULT_CONFIG:log.retention.hours=168}
```

As we can see from the configuration of ready kafka, the value of `log.retention.hours` has been changed from `125` to `150`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kf -n demo kafka-dev
kubectl delete kafkaopsrequest -n demo kfops-reconfigure-apply-combined kfops-reconfigure-combined
kubectl delete secret -n demo kf-combined-custom-config new-kf-combined-custom-config
kubectl delete namespace demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
