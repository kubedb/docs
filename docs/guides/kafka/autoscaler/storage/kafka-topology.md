---
title: Kafka Topology Autoscaling
menu:
  docs_{{ .version }}:
    identifier: kf-storage-auto-scaling-topology
    name: Topology Cluster
    parent: kf-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Kafka Topology Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Kafka Topology cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [KafkaAutoscaler](/docs/guides/kafka/concepts/kafkaautoscaler.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)
    - [Storage Autoscaling Overview](/docs/guides/kafka/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/kafka](/docs/examples/kafka) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of Topology Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Kafka` topology using a supported version by `KubeDB` operator. Then we are going to apply `KafkaAutoscaler` to set up autoscaling.

#### Deploy Kafka topology

In this section, we are going to deploy a Kafka topology cluster with version `4.4.26`.  Then, in the next section we will set up autoscaling for this cluster using `KafkaAutoscaler` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

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

Let's create the `Kafka` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaler/kafka-topology.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait until `kafka-dev` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   0s
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   24s
.
.
kafka-prod    kubedb.com/v1   3.9.0     Ready          119s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo kafka-prod-broker -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get petset -n demo kafka-prod-controller -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                                                      STORAGECLASS          REASON     AGE
pvc-128d9138-64da-4021-8a7c-7ca80823e842   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-1               longhorn              <unset>    33s
pvc-27fe9102-2e7d-41e0-b77d-729a82c64e21   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-0                   longhorn              <unset>    51s
pvc-3bb98ba1-9cea-46ad-857f-fc843c265d57   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-0               longhorn              <unset>    50s
pvc-68f86aac-33d1-423a-bc56-8a905b546db2   1Gi        RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-1                   longhorn              <unset>    32s
```

You can see the petset for both broker and controller has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `KafkaAutoscaler` CRO to set up storage autoscaling for this cluster(broker and controller).

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a KafkaAutoscaler Object.

#### Create KafkaAutoscaler Object

In order to set up vertical autoscaling for this topology cluster, we have to create a `KafkaAutoscaler` CRO with our desired configuration. Below is the YAML of the `KafkaAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kf-storage-autoscaler-topology
  namespace: demo
spec:
  databaseRef:
    name: kafka-prod
  storage:
    broker:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 100
    controller:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 100
```

Here,

- `spec.clusterRef.name` specifies that we are performing vertical scaling operation on `kafka-prod` cluster.
- `spec.storage.broker.trigger/spec.storage.controller.trigger` specifies that storage autoscaling is enabled for broker and controller of topology cluster.
- `spec.storage.broker.usageThreshold/spec.storage.controller.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.broker.scalingThreshold/spec.storage.broker.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `100%` of the current amount.
- It has another field `spec.storage.broker.expansionMode/spec.storage.controller.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `KafkaAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaling/storage/kafka-storage-autoscaler-topology.yaml
kafkaautoscaler.autoscaling.kubedb.com/kf-storage-autoscaler-topology created
```

#### Storage Autoscaling is set up successfully

Let's check that the `kafkaautoscaler` resource is created successfully,

```bash
NAME                             AGE
kf-storage-autoscaler-topology   8s

$ kubectl describe kafkaautoscaler -n demo kf-storage-autoscaler-topology 
Name:         kf-storage-autoscaler-topology
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         KafkaAutoscaler
Metadata:
  Creation Timestamp:  2024-08-27T08:54:35Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Kafka
    Name:                  kafka-prod
    UID:                   1ae37155-dd92-4547-8aba-589140d1d2cf
  Resource Version:        1142604
  UID:                     bca444d0-d860-4588-9b51-412c614c4771
Spec:
  Database Ref:
    Name:  kafka-prod
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Broker:
      Expansion Mode:  Online
      Scaling Rules:
        Applies Upto:     
        Threshold:        100pc
      Scaling Threshold:  100
      Trigger:            On
      Usage Threshold:    60
    Controller:
      Expansion Mode:  Online
      Scaling Rules:
        Applies Upto:     
        Threshold:        100pc
      Scaling Threshold:  100
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```
So, the `kafkaautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

We are autoscaling volume for both broker and controller. So we need to fill up the persistent volume for both broker and controller.

1. Let's exec into the broker pod and fill the cluster volume using the following commands:

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- bash
kafka@kafka-prod-broker-0:~$ df -h /var/log/kafka
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-27fe9102-2e7d-41e0-b77d-729a82c64e21  974M  256K  958M   1% /var/log/kafka
kafka@kafka-prod-broker-0:~$ dd if=/dev/zero of=/var/log/kafka/file.img bs=600M count=1
1+0 records in
1+0 records out
629145600 bytes (629 MB, 600 MiB) copied, 5.58851 s, 113 MB/s
kafka@kafka-prod-broker-0:~$ df -h /var/log/kafka
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-27fe9102-2e7d-41e0-b77d-729a82c64e21  974M  601M  358M  63% /var/log/kafka
```

2. Let's exec into the controller pod and fill the cluster volume using the following commands:

```bash
$ kubectl exec -it -n demo kafka-prod-controller-0 -- bash
kafka@kafka-prod-controller-0:~$ df -h /var/log/kafka
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-3bb98ba1-9cea-46ad-857f-fc843c265d57  974M  192K  958M   1% /var/log/kafka
kafka@kafka-prod-controller-0:~$ dd if=/dev/zero of=/var/log/kafka/file.img bs=600M count=1
1+0 records in
1+0 records out
629145600 bytes (629 MB, 600 MiB) copied, 3.39618 s, 185 MB/s
kafka@kafka-prod-controller-0:~$ df -h /var/log/kafka
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-3bb98ba1-9cea-46ad-857f-fc843c265d57  974M  601M  358M  63% /var/log/kafka
```

So, from the above output we can see that the storage usage is 63% for both nodes, which exceeded the `usageThreshold` 60%.

There will be two `KafkaOpsRequest` created for both broker and controller to expand the volume of the cluster for both nodes.
Let's watch the `kafkaopsrequest` in the demo namespace to see if any `kafkaopsrequest` object is created. After some time you'll see that a `kafkaopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get kafkaopsrequest -n demo
Every 2.0s: kubectl get kafkaopsrequest -n demo
NAME                     TYPE              STATUS        AGE
kfops-kafka-prod-7qwpbn  VolumeExpansion   Progressing   10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get kafkaopsrequest -n demo 
NAME                    TYPE              STATUS        AGE
kfops-kafka-prod-7qwpbn  VolumeExpansion   Successful   2m37s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to expand the volume of the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-kafka-prod-7qwpbn 
Name:         kfops-kafka-prod-7qwpbn
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=kafka-prod
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=kafkas.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-27T08:59:43Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  KafkaAutoscaler
    Name:                  kf-storage-autoscaler-topology
    UID:                   bca444d0-d860-4588-9b51-412c614c4771
  Resource Version:        1144249
  UID:                     2a9bd422-c6ce-47c9-bfd6-ba7f79774c17
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Type:    VolumeExpansion
  Volume Expansion:
    Broker:  2041405440
    Mode:    Online
Status:
  Conditions:
    Last Transition Time:  2024-08-27T08:59:43Z
    Message:               Kafka ops-request has started to expand volume of kafka nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-08-27T08:59:51Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-08-27T08:59:51Z
    Message:               is petset deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetsetDeleted
    Last Transition Time:  2024-08-27T09:00:01Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-08-27T09:00:06Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-08-27T09:00:06Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-08-27T09:03:51Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-08-27T09:03:56Z
    Message:               successfully updated broker node PVC sizes
    Observed Generation:   1
    Reason:                UpdateBrokerNodePVCs
    Status:                True
    Type:                  UpdateBrokerNodePVCs
    Last Transition Time:  2024-08-27T09:04:03Z
    Message:               successfully reconciled the Kafka resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-27T09:04:08Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-08-27T09:04:08Z
    Message:               Successfully completed volumeExpansion for kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                   Age    From                         Message
  ----     ------                                   ----   ----                         -------
  Normal   Starting                                 6m6s   KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-kafka-prod-7qwpbn
  Normal   Starting                                 6m6s   KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                               6m6s   KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-7qwpbn
  Warning  get pet set; ConditionStatus:True        5m58s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is petset deleted; ConditionStatus:True  5m58s  KubeDB Ops-manager Operator  is petset deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        5m53s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                         5m48s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True            5m43s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     5m43s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m38s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   5m38s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True            5m33s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m28s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m23s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m18s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m13s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m8s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m3s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m58s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m53s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m48s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m43s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m38s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m33s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m28s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m23s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m18s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m13s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m8s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m3s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m58s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    3m58s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m53s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     3m53s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m48s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   3m48s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True            3m43s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m38s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m33s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m28s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m23s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m18s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m13s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m8s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m3s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m58s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m53s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m48s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m43s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m38s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m33s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m28s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m23s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m18s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m13s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m8s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m3s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            118s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    118s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateBrokerNodePVCs                     113s   KubeDB Ops-manager Operator  successfully updated broker node PVC sizes
  Normal   UpdatePetSets                            106s   KubeDB Ops-manager Operator  successfully reconciled the Kafka resources
  Warning  get pet set; ConditionStatus:True        101s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             101s   KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 101s   KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                               101s   KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-7qwpbn
```

After a few minutes, another `KafkaOpsRequest` of type `VolumeExpansion` will be created for the controller node.

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                     TYPE              STATUS        AGE
kfops-kafka-prod-7qwpbn  VolumeExpansion   Successful   2m47s
kfops-kafka-prod-sa4thn  VolumeExpansion   Progressing  10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                     TYPE              STATUS        AGE
kfops-kafka-prod-7qwpbn  VolumeExpansion   Successful   4m47s
kfops-kafka-prod-sa4thn  VolumeExpansion   Successful   2m10s
```

We can see from the above output that the `KafkaOpsRequest` `kfops-kafka-prod-sa4thn` has also succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to expand the volume of the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-kafka-prod-2ta9m6 
Name:         kfops-kafka-prod-2ta9m6
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=kafka-prod
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=kafkas.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-27T09:04:43Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  KafkaAutoscaler
    Name:                  kf-storage-autoscaler-topology
    UID:                   bca444d0-d860-4588-9b51-412c614c4771
  Resource Version:        1145309
  UID:                     c965e481-8dbd-4b1d-8a9a-40239753cbf0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Type:    VolumeExpansion
  Volume Expansion:
    Controller:  2041405440
    Mode:        Online
Status:
  Conditions:
    Last Transition Time:  2024-08-27T09:04:43Z
    Message:               Kafka ops-request has started to expand volume of kafka nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-08-27T09:04:51Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-08-27T09:04:51Z
    Message:               is petset deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetsetDeleted
    Last Transition Time:  2024-08-27T09:05:01Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-08-27T09:05:06Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-08-27T09:05:06Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-08-27T09:09:36Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-08-27T09:09:41Z
    Message:               successfully updated controller node PVC sizes
    Observed Generation:   1
    Reason:                UpdateControllerNodePVCs
    Status:                True
    Type:                  UpdateControllerNodePVCs
    Last Transition Time:  2024-08-27T09:09:47Z
    Message:               successfully reconciled the Kafka resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-27T09:09:53Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-08-27T09:09:53Z
    Message:               Successfully completed volumeExpansion for kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                   Age    From                         Message
  ----     ------                                   ----   ----                         -------
  Normal   Starting                                 8m17s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-kafka-prod-2ta9m6
  Normal   Starting                                 8m17s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                               8m17s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-2ta9m6
  Warning  get pet set; ConditionStatus:True        8m9s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is petset deleted; ConditionStatus:True  8m9s   KubeDB Ops-manager Operator  is petset deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        8m4s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                         7m59s  KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True            7m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     7m54s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   7m49s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True            7m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m29s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m19s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m9s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            7m4s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m29s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m19s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m9s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            6m4s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    6m4s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     5m59s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   5m54s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True            5m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m29s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m19s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m9s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m4s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m29s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m19s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m14s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m9s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m4s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m59s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m54s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m49s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m44s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m39s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m34s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m29s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m24s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    3m24s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateControllerNodePVCs                 3m19s  KubeDB Ops-manager Operator  successfully updated controller node PVC sizes
  Normal   UpdatePetSets                            3m12s  KubeDB Ops-manager Operator  successfully reconciled the Kafka resources
  Warning  get pet set; ConditionStatus:True        3m7s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             3m7s   KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 3m7s   KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                               3m7s   KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-kafka-prod-2ta9m6
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the topology cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo kafka-prod-broker -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2041405440"
$ kubectl get petset -n demo kafka-prod-controller -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2041405440"
$ kubectl get pv -n demo
NAME                                       CAPACITY     ACCESS MODES   RECLAIM POLICY   STATUS     CLAIM                                                      STORAGECLASS          REASON     AGE
pvc-128d9138-64da-4021-8a7c-7ca80823e842   1948Mi       RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-1               longhorn              <unset>    33s
pvc-27fe9102-2e7d-41e0-b77d-729a82c64e21   1948Mi       RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-0                   longhorn              <unset>    51s
pvc-3bb98ba1-9cea-46ad-857f-fc843c265d57   1948Mi       RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-controller-0               longhorn              <unset>    50s
pvc-68f86aac-33d1-423a-bc56-8a905b546db2   1948Mi       RWO            Delete           Bound      demo/kafka-prod-data-kafka-prod-broker-1                   longhorn              <unset>    32s
```

The above output verifies that we have successfully autoscaled the volume of the Kafka topology cluster for both broker and controller.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequests -n demo kfops-kafka-prod-7qwpbn kfops-kafka-prod-sa4thn
kubectl delete kafkautoscaler -n demo kf-storage-autoscaler-topology
kubectl delete kf -n demo kafka-prod
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
