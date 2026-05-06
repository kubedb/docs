---
title: Cassandra Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: cas-auto-scaling-storage-autoscale
    name: Topology Cluster
    parent: cas-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Cassandra Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Cassandra cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
  - [CassandraAutoscaler](/docs/guides/cassandra/concepts/cassandraautoscaler.md)
  - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)
  - [Storage Autoscaling Overview](/docs/guides/cassandra/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Storage Autoscaling of Cluster Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  6h2m
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   9m41s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   9m24s
```

We can see from the output the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Cassandra` cluster using a supported version by `KubeDB` operator. Then we are going to apply `CassandraAutoscaler` to set up autoscaling.

#### Deploy Cassandra Cluster

In this section, we are going to deploy a Cassandra cluster with version `3.13.2`.  Then, in the next section we will set up autoscaling for this database using `CassandraAutoscaler` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

> If you want to autoscale Cassandra `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:

  name: cassandra-autoscale
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          storageClassName:
            longhorn
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 600Mi
        storageType: Durable
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 3Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
  deletionPolicy: WipeOut

```

Let's create the `Cassandra` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/examples/cassandra/autoscaling/storage/cassandra-autoscale.yaml
cassandra.kubedb.com/cassandra-autoscale created
```

Now, wait until `cassandra-autoscale` has status `Ready`. i.e,

```bash
$ kubectl get cassandra -n demo
NAME                  TYPE                  VERSION   STATUS   AGE
cassandra-autoscale   kubedb.com/v1alpha2   5.0.3     Ready    16m
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo cassandra-autoscale-rack-r0 -o json | jq '.spec.volumeClaimTemplates[0].spec.resources.requests.storage'
"600Mi"


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                                       STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-394fefad-d4ad-4dfa-ba11-df96e015da30   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-main-config-volume-cassandra-autoscale-rack-r0-1   longhorn       <unset>                          21m
pvc-86ece3c8-520a-4d41-834e-66108867ca36   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-data-cassandra-autoscale-rack-r0-1                 longhorn       <unset>                          21m
pvc-c35bb138-9f13-4098-b2b0-cc151f013f6d   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-main-config-volume-cassandra-autoscale-rack-r0-0   longhorn       <unset>                          21m
pvc-cc932132-de53-425f-bd31-91af255a47e8   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-data-cassandra-autoscale-rack-r0-0                 longhorn       <unset>                          21m
pvc-cd57fb5f-b2f3-48de-b9d2-03059b05113f   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-nodetool-cassandra-autoscale-rack-r0-1             longhorn       <unset>                          21m
pvc-e550c573-60c7-4ec0-9e01-cf22683c502c   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-nodetool-cassandra-autoscale-rack-r0-0             longhorn       <unset>                          21m
```

You can see the petset has 600Mi storage, and the capacity of all the persistent volume is also 600Mi.

We are now ready to apply the `CassandraAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a CassandraAutoscaler Object.

#### Create CassandraAutoscaler Object

In order to set up vertical autoscaling for this replicaset database, we have to create a `CassandraAutoscaler` CRO with our desired configuration. Below is the YAML of the `CassandraAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: CassandraAutoscaler
metadata:
  name: cassandra-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: cassandra-autoscale
  storage:
    cassandra:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 100
      expansionMode: "Offline"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `cassandra-autoscale` database.
- `spec.storage.cassandra.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.cassandra.usageThreshold` specifies storage usage threshold, if storage usage exceeds `20%` then storage autoscaling will be triggered.
- `spec.storage.cassandra.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `100%` of the current amount.
- `spec.storage.cassandra.expansionMode` specifies the expansion mode of volume expansion `cassandraOpsRequest` created by `cassandraAutoscaler`. topolvm-provisioner supports online volume expansion so here `expansionMode` is set as "Online".

Let's create the `cassandraAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/autoscaling/storage/cassandra-autoscaler-ops.yaml
cassandraautoscaler.autoscaling.kubedb.com/cassandra-storage-autosclaer created
```

#### Storage Autoscaling is set up successfully

Let's check that the `cassandraautoscaler` resource is created successfully,

```bash
$ kubectl get cassandraautoscaler -n demo
NAME                           AGE
cassandra-storage-autoscaler   1m25s

$ kubectl describe cassandraautoscaler cassandra-storage-autoscaler -n demo
Name:         cassandra-storage-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         CassandraAutoscaler
Metadata:
  Creation Timestamp:  2025-07-15T04:48:26Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Cassandra
    Name:                  cassandra-autoscale
    UID:                   ca165d22-0cdd-44c9-8fe6-6691ceccc05b
  Resource Version:        82339
  UID:                     d49df408-a2b8-4f30-a9d4-320ca4351885
Spec:
  Database Ref:
    Name:  cassandra-autoscale
  Ops Request Options:
    Apply:  IfReady
  Storage:
    Cassandra:
      Expansion Mode:  Offline
      Scaling Rules:
        Applies Upto:     
        Threshold:        100pc
      Scaling Threshold:  100
      Trigger:            On
      Usage Threshold:    2
Events:                   <none>
```
So, the `cassandraautoscaler` resource is created successfully.

Let's watch the `cassandraopsrequest` in the demo namespace to see if any `cassandraopsrequest` object is created. After some time you'll see that a `cassandraopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                              TYPE              STATUS        AGE
casops-cassandra-autoscale-xojkua   VolumeExpansion   Progressing   15s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                                TYPE              STATUS       AGE
casops-cassandra-autoscale-9ah2rp   VolumeExpansion   Successful   10m
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$kubectl describe cassandraopsrequest -n demo casops-cassandra-autoscale-9ah2rp 
Name:         casops-cassandra-autoscale-9ah2rp
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=cassandra-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=cassandras.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-15T04:54:52Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  CassandraAutoscaler
    Name:                  cassandra-storage-autoscaler
    UID:                   d49df408-a2b8-4f30-a9d4-320ca4351885
  Resource Version:        85451
  UID:                     6f6d2997-12ea-47d0-b07f-f6820827373e
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-autoscale
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Offline
    Node:  1203126272
Status:
  Conditions:
    Last Transition Time:  2025-07-15T04:54:52Z
    Message:               Cassandra ops-request has started to expand volume of cassandra nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-07-15T04:55:10Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2025-07-15T04:55:00Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2025-07-15T04:55:00Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2025-07-15T05:04:40Z
    Message:               successfully updated node PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2025-07-15T05:03:50Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2025-07-15T04:55:15Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2025-07-15T04:55:15Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2025-07-15T04:55:50Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-07-15T04:55:50Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-07-15T05:01:05Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-07-15T04:57:50Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2025-07-15T04:57:55Z
    Message:               running cassandra; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningCassandra
    Last Transition Time:  2025-07-15T05:04:48Z
    Message:               successfully reconciled the Cassandra resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-15T05:04:53Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2025-07-15T05:04:53Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2025-07-15T05:04:53Z
    Message:               Successfully completed volumeExpansion for Cassandra
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age    From                         Message
  ----     ------                                    ----   ----                         -------
  Normal   Starting                                  11m    KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-cassandra-autoscale-9ah2rp
  Normal   Starting                                  11m    KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-autoscale
  Normal   Successful                                11m    KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-autoscale for CassandraOpsRequest: casops-cassandra-autoscale-9ah2rp
  Warning  get petset; ConditionStatus:True          10m    KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True       10m    KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True          10m    KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Normal   OrphanPetSetPods                          10m    KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True             10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   10m    KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True          10m    KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False            10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             10m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True           10m    KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    10m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             10m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             9m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             9m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             8m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     8m6s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          8m6s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   8m6s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             8m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  running cassandra; ConditionStatus:False  8m1s   KubeDB Ops-manager Operator  running cassandra; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             7m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             7m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   7m21s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True          7m21s  KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False            7m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             6m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True           6m46s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    6m46s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             6m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             6m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             6m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m41s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             5m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m51s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     4m51s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          4m51s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   4m51s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             4m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   4m6s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True          4m6s   KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False            4m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             3m31s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m31s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     3m31s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          3m31s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   3m31s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             3m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             2m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             2m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             2m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             2m41s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   2m41s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True          2m41s  KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False            2m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True             2m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             2m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     2m6s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True          2m6s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True   2m6s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             2m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             102s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             98s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             95s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             91s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             86s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True             81s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateNodePVCs                            76s    KubeDB Ops-manager Operator  successfully updated node PVC sizes
  Normal   UpdatePetSets                             68s    KubeDB Ops-manager Operator  successfully reconciled the Cassandra resources
  Warning  get pet set; ConditionStatus:True         63s    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                              63s    KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                  63s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-autoscale
  Normal   Successful                                63s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-autoscale for CassandraOpsRequest: casops-cassandra-autoscale-9ah2rp
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the replicaset database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo cassandra-autoscale-rack-r0 -o json | jq '.spec.volumeClaimTemplates[0].spec.resources.requests.storage'
"1203126272"

$  kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                                       STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-394fefad-d4ad-4dfa-ba11-df96e015da30   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-main-config-volume-cassandra-autoscale-rack-r0-1   longhorn       <unset>                          45m
pvc-86ece3c8-520a-4d41-834e-66108867ca36   1148Mi     RWO            Delete           Bound    demo/cassandra-autoscale-data-cassandra-autoscale-rack-r0-1                 longhorn       <unset>                          45m
pvc-c35bb138-9f13-4098-b2b0-cc151f013f6d   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-main-config-volume-cassandra-autoscale-rack-r0-0   longhorn       <unset>                          45m
pvc-cc932132-de53-425f-bd31-91af255a47e8   1148Mi     RWO            Delete           Bound    demo/cassandra-autoscale-data-cassandra-autoscale-rack-r0-0                 longhorn       <unset>                          45m
pvc-cd57fb5f-b2f3-48de-b9d2-03059b05113f   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-nodetool-cassandra-autoscale-rack-r0-1             longhorn       <unset>                          45m
pvc-e550c573-60c7-4ec0-9e01-cf22683c502c   600Mi      RWO            Delete           Bound    demo/cassandra-autoscale-nodetool-cassandra-autoscale-rack-r0-0             longhorn       <unset>                          45m

```

The above output verifies that we have successfully autoscaled the volume related to data of the cassandra cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cassandra -n demo cassandra-autoscale
kubectl delete cassandraautoscaler -n demo casops-cassandra-autoscale-xojkua
kubectl delete ns demo
```
