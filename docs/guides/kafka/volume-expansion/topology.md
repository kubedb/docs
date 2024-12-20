---
title: Kafka Topology Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: kf-volume-expansion-topology
    name: Topology
    parent: kf-volume-expansion
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka Topology Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Kafka Topology Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [Topology](/docs/guides/kafka/clustering/topology-cluster/index.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Volume Expansion Overview](/docs/guides/kafka/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/kafka](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of Topology Kafka Cluster

Here, we are going to deploy a `Kafka` topology using a supported version by `KubeDB` operator. Then we are going to apply `KafkaOpsRequest` to expand its volume.

### Prepare Kafka Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Kafka` combined cluster with version `3.9.0`.

### Deploy Kafka

In this section, we are going to deploy a Kafka topology cluster for broker and controller with 1GB volume. Then, in the next section we will expand its volume to 2GB using `KafkaOpsRequest` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/volume-expansion/kafka-topology.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait until `kafka-prod` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   0s
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   9s
.
.
kafka-prod    kubedb.com/v1   3.9.0     Ready          2m10s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo kafka-prod-broker -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get petset -n demo kafka-prod-controller -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo                                       
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                                           STORAGECLASS   REASON   AGE
pvc-3f177a92721440bb   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-0    standard                106s
pvc-86ff354122324b1c   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-1        standard                78s
pvc-9fa35d773aa74bd0   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-1    standard                75s
pvc-ccf50adf179e4162   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-0        standard                106s
```

You can see the petsets have 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `KafkaOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the kafka topology cluster.

#### Create KafkaOpsRequest

In order to expand the volume of the database, we have to create a `KafkaOpsRequest` CR with our desired volume size. Below is the YAML of the `KafkaOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kf-volume-exp-topology
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: kafka-prod
  volumeExpansion:
    broker: 3Gi
    controller: 2Gi
    mode: Online
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `kafka-prod`.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.broker` specifies the desired volume size for broker node.
- `spec.volumeExpansion.controller` specifies the desired volume size for controller node.

> If you want to expand the volume of only one node, you can specify the desired volume size for that node only.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/volume-expansion/kafka-volume-expansion-topology.yaml
kafkaopsrequest.ops.kubedb.com/kf-volume-exp-topology created
```

#### Verify Kafka Topology volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `Kafka` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CR,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                     TYPE              STATUS       AGE
kf-volume-exp-topology   VolumeExpansion   Successful   3m1s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to expand the volume of kafka.

```bash
$ kubectl describe kafkaopsrequest -n demo kf-volume-exp-topology   
Name:         kf-volume-exp-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-07-31T04:44:17Z
  Generation:          1
  Resource Version:    149682
  UID:                 e0e19d97-7150-463c-9a7d-53eff05ea6c4
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Type:    VolumeExpansion
  Volume Expansion:
    Broker:      3Gi
    Controller:  2Gi
    Mode:        Online
Status:
  Conditions:
    Last Transition Time:  2024-07-31T04:44:17Z
    Message:               Kafka ops-request has started to expand volume of kafka nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-07-31T04:44:25Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-07-31T04:44:25Z
    Message:               is petset deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetsetDeleted
    Last Transition Time:  2024-07-31T04:44:45Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-07-31T04:44:50Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-07-31T04:44:50Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-07-31T04:44:55Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-07-31T04:45:10Z
    Message:               successfully updated controller node PVC sizes
    Observed Generation:   1
    Reason:                UpdateControllerNodePVCs
    Status:                True
    Type:                  UpdateControllerNodePVCs
    Last Transition Time:  2024-07-31T04:45:35Z
    Message:               successfully updated broker node PVC sizes
    Observed Generation:   1
    Reason:                UpdateBrokerNodePVCs
    Status:                True
    Type:                  UpdateBrokerNodePVCs
    Last Transition Time:  2024-07-31T04:45:42Z
    Message:               successfully reconciled the Kafka resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-31T04:45:47Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-07-31T04:45:47Z
    Message:               Successfully completed volumeExpansion for kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                   Age   From                         Message
  ----     ------                                   ----  ----                         -------
  Normal   Starting                                 116s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kf-volume-exp-topology
  Normal   Starting                                 116s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                               116s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kf-volume-exp-topology
  Warning  get pet set; ConditionStatus:True        108s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is petset deleted; ConditionStatus:True  108s  KubeDB Ops-manager Operator  is petset deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        103s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        98s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is petset deleted; ConditionStatus:True  98s   KubeDB Ops-manager Operator  is petset deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        93s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                         88s   KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True            83s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     83s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            78s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    78s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            73s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     73s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            68s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    68s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateControllerNodePVCs                 63s   KubeDB Ops-manager Operator  successfully updated controller node PVC sizes
  Warning  get pvc; ConditionStatus:True            58s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     58s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            53s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    53s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            48s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     48s   KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            43s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    43s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateBrokerNodePVCs                     38s   KubeDB Ops-manager Operator  successfully updated broker node PVC sizes
  Normal   UpdatePetSets                            31s   KubeDB Ops-manager Operator  successfully reconciled the Kafka resources
  Warning  get pet set; ConditionStatus:True        26s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        26s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             26s   KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 26s   KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                               26s   KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kf-volume-exp-topology
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo kafka-prod-broker -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"3Gi"

$ kubectl get petset -n demo kafka-prod-controller -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo                                       
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                                           STORAGECLASS   REASON   AGE
pvc-3f177a92721440bb   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-0    standard                5m25s
pvc-86ff354122324b1c   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-1        standard                4m51s
pvc-9fa35d773aa74bd0   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-1    standard                5m1s
pvc-ccf50adf179e4162   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-0        standard                5m30s
```

The above output verifies that we have successfully expanded the volume of the Kafka.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequest -n demo kf-volume-exp-topology
kubectl delete kf -n demo kafka-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- 
[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
