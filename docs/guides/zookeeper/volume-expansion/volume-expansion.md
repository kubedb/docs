---
title: ZooKeeper Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: zk-volume-expansion-describe
    name: Expand Storage Volume
    parent: zk-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Volume Expansion of ZooKeeper Ensemble

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a ZooKeeper database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md)
    - [ZooKeeperOpsRequest](/docs/guides/zookeeper/concepts/opsrequest.md)
    - [Volume Expansion Overview](/docs/guides/zookeeper/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/ZooKeeper](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/zookeeper) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Expand Volume of ZooKeeper Ensemble

Here, we are going to deploy a `ZooKeeper` standalone using a supported version by `KubeDB` operator. Then we are going to apply `ZooKeeperOpsRequest` to expand its volume.

### Prepare ZooKeeper Ensemble

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   93s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   90s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `ZooKeeper` standalone database with version `3.8.3`.

#### Deploy ZooKeeper Ensemble

In this section, we are going to deploy a ZooKeeper standalone database with 1GB volume. Then, in the next section we will expand its volume to 2GB using `ZooKeeperOpsRequest` CRD. Below is the YAML of the `ZooKeeper` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-quickstart
  namespace: demo
spec:
  version: "3.8.3"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

Let's create the `ZooKeeper` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/volume-expansion/zookeeper.yaml
zookeeper.kubedb.com/zk-quickstart created
```

Now, wait until `zk-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get zk -n demo
NAME            VERSION    STATUS    AGE
zk-quickstart   3.8.3      Ready     5m56s
```

Let's check volume size from PetSet, and from the persistent volume,

```bash
$ kubectl get petset -n demo zk-quickstart -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-3551d7c0-0df6-4f94-b1e0-21834319ecab   1Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-0   longhorn       <unset>                          92s
pvc-b5882e9e-3c61-4609-b5ba-0eb9f32edbbc   1Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-2   longhorn       <unset>                          58s
pvc-dccf2b12-d695-4792-8e4b-de4342e7fed4   1Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-1   longhorn       <unset>                          74s
```

You can see the PetSet has 1GB storage, and the capacity of the persistent volume is also 1GB.

We are now ready to apply the `ZooKeeperOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the standalone database.

#### Create ZooKeeperOpsRequest

In order to expand the volume of the database, we have to create a `ZooKeeperOpsRequest` CR with our desired volume size. Below is the YAML of the `ZooKeeperOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zk-offline-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: zk-quickstart
  volumeExpansion:
    mode: "Offline"
    node: 2Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `zk-quickstart` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.node` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode(`Online` or `Offline`).

During `Online` VolumeExpansion KubeDB expands volume without pausing database object, it directly updates the underlying PVC. And for `Offline` volume expansion, the database is paused. The Pods are deleted and PVC is updated. Then the database Pods are recreated with updated PVC.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/volume-expansion/zkops-volume-exp-offline.yaml
zookeeperopsrequest.ops.kubedb.com/zk-offline-volume-expansion created
```

#### Verify ZooKeeper Standalone volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `ZooKeeper` object and related `Petsets` and `Persistent Volume`.

Let's wait for `ZooKeeperOpsRequest` to be `Successful`. Run the following command to watch `ZooKeeperOpsRequest` CR,

```bash
$ kubectl get zookeeperopsrequest -n demo
NAME                          TYPE              STATUS       AGE
zk-offline-volume-expansion   VolumeExpansion   Successful   75s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe zookeeperopsrequest -n demo zk-offline-volume-expansion
Name:         zk-offline-volume-expansion
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-10-28T11:12:02Z
  Generation:          1
  Resource Version:    1321277
  UID:                 13851249-f148-4745-a565-0aaea704f830
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  zk-quickstart
  Type:    VolumeExpansion
  Volume Expansion:
    Mode:  Offline
    Node:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-28T11:12:02Z
    Message:               ZooKeeper ops-request has started to expand volume of zookeeper nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-10-28T11:12:20Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-10-28T11:12:10Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2024-10-28T11:12:10Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2024-10-28T11:15:55Z
    Message:               successfully updated node PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2024-10-28T11:15:05Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-10-28T11:12:25Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2024-10-28T11:12:25Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-10-28T11:13:00Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-28T11:13:00Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2024-10-28T11:15:45Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-10-28T11:13:15Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-10-28T11:16:00Z
    Message:               successfully reconciled the ZooKeeper resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-28T11:16:05Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-10-28T11:16:05Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-10-28T11:16:05Z
    Message:               Successfully completed volumeExpansion for ZooKeeper
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                   Age    From                         Message
  ----     ------                                   ----   ----                         -------
  Normal   Starting                                 5m19s  KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/zk-offline-volume-expansion
  Normal   Starting                                 5m19s  KubeDB Ops-manager Operator  Pausing ZooKeeper database: demo/zk-quickstart
  Normal   Successful                               5m19s  KubeDB Ops-manager Operator  Successfully paused ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: zk-offline-volume-expansion
  Warning  get petset; ConditionStatus:True         5m11s  KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Warning  delete petset; ConditionStatus:True      5m11s  KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Warning  get petset; ConditionStatus:True         5m6s   KubeDB Ops-manager Operator  get petset; ConditionStatus:True
  Normal   OrphanPetSetPods                         5m1s   KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True            4m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  4m56s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         4m56s  KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False           4m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            4m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          4m21s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   4m21s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            4m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            4m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            4m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            4m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    4m6s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         4m6s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  4m6s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            4m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            3m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  3m56s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         3m56s  KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False           3m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            3m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m21s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          3m21s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   3m21s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            3m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            3m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            3m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            3m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            3m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    3m1s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         3m1s   KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  3m1s   KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m56s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m51s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  2m51s  KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True         2m51s  KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False           2m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            2m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True          2m16s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False   2m16s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True            2m11s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m11s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            2m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            2m1s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            116s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            116s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            111s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            111s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            106s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            106s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            101s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            101s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            96s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True            96s    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True    96s    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True         96s    KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  patch ops request; ConditionStatus:True  96s    KubeDB Ops-manager Operator  patch ops request; ConditionStatus:True
  Warning  get pod; ConditionStatus:True            91s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateNodePVCs                           86s    KubeDB Ops-manager Operator  successfully updated node PVC sizes
  Normal   UpdatePetSets                            81s    KubeDB Ops-manager Operator  successfully reconciled the ZooKeeper resources
  Warning  get pet set; ConditionStatus:True        76s    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                             76s    KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                 76s    KubeDB Ops-manager Operator  Resuming ZooKeeper database: demo/zk-quickstart
  Normal   Successful                               76s    KubeDB Ops-manager Operator  Successfully resumed ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: zk-offline-volume-expansion
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the standalone database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo zk-quickstart -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                     STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-1b112414-6162-4e75-99c9-3e62cb4efb4a   2Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-1   longhorn       <unset>                          16m
pvc-3159b881-1954-4008-8594-599bee9fd11e   2Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-0   longhorn       <unset>                          17m
pvc-43ba80bd-9029-413e-b89c-1f373fd0cd3d   2Gi        RWO            Delete           Bound    demo/zk-quickstart-data-zk-quickstart-2   longhorn       <unset>                          16m
```

The above output verifies that we have successfully expanded the volume of the ZooKeeper standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete zk -n demo zk-quickstart
kubectl delete zookeeperopsrequest -n demo zk-offline-volume-expansion
```
