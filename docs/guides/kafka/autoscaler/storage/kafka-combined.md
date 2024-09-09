---
title: Kafka Combined Autoscaling
menu:
  docs_{{ .version }}:
    identifier: kf-storage-auto-scaling-combined
    name: Combined Cluster
    parent: kf-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Kafka Combined Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Kafka Combined cluster.

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

## Storage Autoscaling of Combined Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Kafka` combined using a supported version by `KubeDB` operator. Then we are going to apply `KafkaAutoscaler` to set up autoscaling.

#### Deploy Kafka combined

In this section, we are going to deploy a Kafka combined cluster with version `4.4.26`.  Then, in the next section we will set up autoscaling for this cluster using `KafkaAutoscaler` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-dev
  namespace: demo
spec:
  replicas: 2
  version: 3.6.1
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
kafka-dev    kubedb.com/v1   3.6.1     Provisioning   0s
kafka-dev    kubedb.com/v1   3.6.1     Provisioning   24s
.
.
kafka-dev    kubedb.com/v1   3.6.1     Ready          92s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo kafka-dev -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                       STORAGECLASS          REASON     AGE
pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51   1Gi        RWO            Delete           Bound    demo/kafka-dev-data-kafka-dev-0             longhorn              <unset>    40s
pvc-f068d245-718b-4561-b452-f3130bb260f6   1Gi        RWO            Delete           Bound    demo/kafka-dev-data-kafka-dev-1             longhorn              <unset>    35s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `KafkaAutoscaler` CRO to set up storage autoscaling for this cluster.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a KafkaAutoscaler Object.

#### Create KafkaAutoscaler Object

In order to set up vertical autoscaling for this combined cluster, we have to create a `KafkaAutoscaler` CRO with our desired configuration. Below is the YAML of the `KafkaAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: KafkaAutoscaler
metadata:
  name: kf-storage-autoscaler-combined
  namespace: demo
spec:
  databaseRef:
    name: kafka-dev
  storage:
    node:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here,

- `spec.clusterRef.name` specifies that we are performing vertical scaling operation on `kafka-dev` cluster.
- `spec.storage.node.trigger` specifies that storage autoscaling is enabled for this cluster.
- `spec.storage.node.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.node.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- It has another field `spec.storage.node.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `KafkaAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/autoscaling/storage/kafka-storage-autoscaler-combined.yaml
kafkaautoscaler.autoscaling.kubedb.com/kf-storage-autoscaler-combined created
```

#### Storage Autoscaling is set up successfully

Let's check that the `kafkaautoscaler` resource is created successfully,

```bash
NAME                             AGE
kf-storage-autoscaler-combined   8s


$ kubectl describe kafkaautoscaler -n demo kf-storage-autoscaler-combined 
Name:         kf-storage-autoscaler-combined
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         KafkaAutoscaler
Metadata:
  Creation Timestamp:  2024-08-27T06:56:57Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Kafka
    Name:                  kafka-dev
    UID:                   a1d1b2f9-ef72-4ef6-8652-f39ee548c744
  Resource Version:        1123501
  UID:                     83c7a7b6-aaf2-4776-8337-114bd1800d7c
Spec:
  Database Ref:
    Name:  kafka-dev
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Node:
      Expansion Mode:  Online
      Scaling Rules:
        Applies Upto:     
        Threshold:        50pc
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```
So, the `kafkaautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the cluster pod and fill the cluster volume using the following commands:

```bash
 $ kubectl exec -it -n demo kafka-dev-0 -- bash
kafka@kafka-dev-0:~$ df -h /var/log/kafka
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51  974M  168K  958M   1% /var/log/kafka
kafka@kafka-dev-0:~$ dd if=/dev/zero of=/var/log/kafka/file.img bs=600M count=1
1+0 records in
1+0 records out
629145600 bytes (629 MB, 600 MiB) copied, 7.44144 s, 84.5 MB/s
kafka@kafka-dev-0:~$ df -h /var/log/kafka
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51  974M  601M  358M  63% /var/log/kafka
```

So, from the above output we can see that the storage usage is 83%, which exceeded the `usageThreshold` 60%.

Let's watch the `kafkaopsrequest` in the demo namespace to see if any `kafkaopsrequest` object is created. After some time you'll see that a `kafkaopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get kafkaopsrequest -n demo
Every 2.0s: kubectl get kafkaopsrequest -n demo
NAME                     TYPE              STATUS        AGE
kfops-kafka-dev-sa4thn   VolumeExpansion   Progressing   10s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get kafkaopsrequest -n demo 
NAME                    TYPE              STATUS        AGE
kfops-kafka-dev-sa4thn  VolumeExpansion   Successful    97s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed to expand the volume of the cluster.

```bash
$ kubectl describe kafkaopsrequests -n demo kfops-kafka-dev-sa4thn 
Name:         kfops-kafka-dev-sa4thn
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=kafka-dev
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=kafkas.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-08-27T08:12:33Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  KafkaAutoscaler
    Name:                  kf-storage-autoscaler-combined
    UID:                   a0ce73df-0d42-483a-9c47-ca58e57ea614
  Resource Version:        1135462
  UID:                     78b52373-75f9-40a1-8528-3d0cd9beb4c5
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-dev
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Online
    Node:  1531054080
Status:
  Conditions:
    Last Transition Time:  2024-08-27T08:12:33Z
    Message:               Kafka ops-request has started to expand volume of kafka nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-08-27T08:12:41Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-08-27T08:12:41Z
    Message:               is petset deleted; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPetsetDeleted
    Last Transition Time:  2024-08-27T08:12:51Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-08-27T08:12:56Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-08-27T08:12:56Z
    Message:               is pvc patched; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatched
    Last Transition Time:  2024-08-27T08:18:16Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-08-27T08:18:21Z
    Message:               successfully updated combined node PVC sizes
    Observed Generation:   1
    Reason:                UpdateCombinedNodePVCs
    Status:                True
    Type:                  UpdateCombinedNodePVCs
    Last Transition Time:  2024-08-27T08:18:27Z
    Message:               successfully reconciled the Kafka resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-08-27T08:18:32Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-08-27T08:18:32Z
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
  Normal   Starting                                 6m19s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-kafka-dev-sa4thn
  Normal   Starting                                 6m19s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-dev
  Normal   Successful                               6m19s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-kafka-dev-sa4thn
  Warning  get pet set; ConditionStatus:True        6m11s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  is petset deleted; ConditionStatus:True  6m11s  KubeDB Ops-manager Operator  is petset deleted; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True        6m6s   KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                         6m1s   KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True            5m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     5m56s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   5m51s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True            5m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            5m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    3m21s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patched; ConditionStatus:True     3m16s  KubeDB Ops-manager Operator  is pvc patched; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   3m11s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True            3m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            116s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            111s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            106s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            101s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            96s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            91s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            86s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            81s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            76s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            71s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            66s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            61s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            56s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            51s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            46s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            41s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            36s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    36s    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   UpdateCombinedNodePVCs                   31s    KubeDB Ops-manager Operator  successfully updated combined node PVC sizes
  Normal   UpdatePetSets                            25s    KubeDB Ops-manager Operator  successfully reconciled the Kafka resources
  Warning  get pet set; ConditionStatus:True        20s    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             20s    KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 20s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-dev
  Normal   Successful                               20s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-dev for KafkaOpsRequest: kfops-kafka-dev-sa4thn
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the combined cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo kafka-dev -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1531054080"
$ kubectl get pv -n demo
NAME                                       CAPACITY      ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                       STORAGECLASS          REASON     AGE
pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51   1462Mi        RWO            Delete           Bound    demo/kafka-dev-data-kafka-dev-0             longhorn              <unset>    30m5s
pvc-f068d245-718b-4561-b452-f3130bb260f6   1462Mi        RWO            Delete           Bound    demo/kafka-dev-data-kafka-dev-1             longhorn              <unset>    30m1s
```

The above output verifies that we have successfully autoscaled the volume of the Kafka combined cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequests -n demo kfops-kafka-dev-sa4thn
kubectl delete kafkautoscaler -n demo kf-storage-autoscaler-combined
kubectl delete kf -n demo kafka-dev
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
