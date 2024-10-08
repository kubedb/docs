---
title: SingleStore Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-volume-expansion-volume-expansion
    name: SingleStore Volume Expansion
    parent: guides-sdb-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a SingleStore.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/singlestore/volume-expansion/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Expand Volume of SingleStore

Here, we are going to deploy a  `SingleStore` cluster using a supported version by `KubeDB` operator. Then we are going to apply `SingleStoreOpsRequest` to expand its volume. The process of expanding SingleStore `standalone` is same as SingleStore cluster.

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

### Prepare SingleStore Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageClass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  6d2h
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   3d21h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   42m
```

Here, we will use `longhorn` storageClass for this tuitorial.

Now, we are going to deploy a `SingleStore` database of 3 replicas with version `8.7.10`.

### Deploy SingleStore

In this section, we are going to deploy a SingleStore Cluster with 1GB volume for `aggregator` nodes and 10GB volume for `leaf` nodes. Then, in the next section we will expand its volume to 2GB using `SingleStoreOpsRequest` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sample-sdb
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        storageClassName: "longhorn"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/volume-expansion/volume-expansion/example/sample-sdb.yaml
singlestore.kubedb.com/sample-sdb created
```

Now, wait until `sample-sdb` has status `Ready`. i.e,

```bash
$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
sample-sdb   kubedb.com/v1alpha2   8.7.10    Ready    4m25s

```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get petset -n demo sample-sdb-aggregator -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get petset -n demo sample-sdb-leaf -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"10Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-41cb892c-99fc-4211-a8c2-4e6f8a16c661   10Gi       RWO            Delete           Bound    demo/data-sample-sdb-leaf-0         longhorn       <unset>                          90s
pvc-6e241724-6577-408e-b8de-9569d7d785c4   10Gi       RWO            Delete           Bound    demo/data-sample-sdb-leaf-1         longhorn       <unset>                          75s
pvc-95ecc525-540b-4496-bf14-bfac901d73c4   1Gi        RWO            Delete           Bound    demo/data-sample-sdb-aggregator-0   longhorn       <unset>                          94s


```

You can see the `aggregator` petset has 1GB storage, and the capacity of all the `aggregator` persistent volumes are also 1GB.

You can see the `leaf` petset has 10GB storage, and the capacity of all the `leaf` persistent volumes are also 10GB.

We are now ready to apply the `SingleStoreOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the SingleStore cluster.

#### Create SingleStoreOpsRequest

In order to expand the volume of the database, we have to create a `SingleStoreOpsRequest` CR with our desired volume size. Below is the YAML of the `SingleStoreOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-offline-vol-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: sample-sdb
  volumeExpansion:
    mode: "Offline"
    aggregator: 2Gi
    leaf: 11Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.aggregator` and `spec.volumeExpansion.leaf` specifies the desired volume size for `aggregator` and `leaf` nodes.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`). Storageclass `longhorn` supports `Offline` volume expansion.

> **Note:** If the Storageclass you are using doesn't support `Online` Volume Expansion, Try offline volume expansion by using `spec.volumeExpansion.mode:"Offline"`.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/volume-expansion/volume-expansion/example/sdb-offline-volume-expansion.yaml
singlestoreopsrequest.ops.kubedb.com/sdb-offline-vol-expansion created
```

#### Verify SingleStore volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of `SingleStore` object and related `PetSets` and `Persistent Volumes`.

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CR,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                        TYPE              STATUS       AGE
sdb-offline-vol-expansion   VolumeExpansion   Successful   13m
```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded. If we describe the `SingleStoreOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe sdbops -n demo sdb-offline-vol-expansion 
Name:         sdb-offline-vol-expansion
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SinglestoreOpsRequest
Metadata:
  Creation Timestamp:  2024-10-15T08:49:11Z
  Generation:          1
  Resource Version:    12476
  UID:                 a0e2f1c3-a6b7-4993-a012-2823c3a2675b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  sample-sdb
  Type:    VolumeExpansion
  Volume Expansion:
    Aggregator:  2Gi
    Leaf:        11Gi
    Mode:        Offline
Status:
  Conditions:
    Last Transition Time:  2024-10-15T08:49:11Z
    Message:               Singlestore ops-request has started to expand volume of singlestore nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2024-10-15T08:49:17Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-15T08:49:42Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2024-10-15T08:49:22Z
    Message:               get pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetSet
    Last Transition Time:  2024-10-15T08:49:22Z
    Message:               delete pet set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetSet
    Last Transition Time:  2024-10-15T08:51:07Z
    Message:               successfully updated Aggregator node PVC sizes
    Observed Generation:   1
    Reason:                UpdateAggregatorNodePVCs
    Status:                True
    Type:                  UpdateAggregatorNodePVCs
    Last Transition Time:  2024-10-15T08:53:32Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2024-10-15T08:49:47Z
    Message:               is ops req patch; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsOpsReqPatch
    Last Transition Time:  2024-10-15T08:49:47Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2024-10-15T08:50:22Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2024-10-15T08:50:22Z
    Message:               is pvc patch; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsPvcPatch
    Last Transition Time:  2024-10-15T08:53:52Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2024-10-15T08:50:42Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2024-10-15T08:50:47Z
    Message:               is running single store; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsRunningSingleStore
    Last Transition Time:  2024-10-15T08:54:32Z
    Message:               successfully updated Leaf node PVC sizes
    Observed Generation:   1
    Reason:                UpdateLeafNodePVCs
    Status:                True
    Type:                  UpdateLeafNodePVCs
    Last Transition Time:  2024-10-15T08:54:43Z
    Message:               successfully reconciled the Singlestore resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-15T08:54:48Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2024-10-15T08:54:48Z
    Message:               Successfully completed volumeExpansion for Singlestore
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                          Age    From                         Message
  ----     ------                                          ----   ----                         -------
  Normal   Starting                                        14m    KubeDB Ops-manager Operator  Start processing for SinglestoreOpsRequest: demo/sdb-offline-vol-expansion
  Normal   Starting                                        14m    KubeDB Ops-manager Operator  Pausing Singlestore database: demo/sample-sdb
  Normal   Successful                                      14m    KubeDB Ops-manager Operator  Successfully paused Singlestore database: demo/sample-sdb for SinglestoreOpsRequest: sdb-offline-vol-expansion
  Warning  get pet set; ConditionStatus:True               14m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  delete pet set; ConditionStatus:True            14m    KubeDB Ops-manager Operator  delete pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True               14m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True               14m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  delete pet set; ConditionStatus:True            14m    KubeDB Ops-manager Operator  delete pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True               14m    KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   OrphanPetSetPods                                13m    KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  get pod; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is ops req patch; ConditionStatus:True          13m    KubeDB Ops-manager Operator  is ops req patch; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True                13m    KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False                  13m    KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patch; ConditionStatus:True              13m    KubeDB Ops-manager Operator  is pvc patch; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False          13m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   13m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   12m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   12m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True           12m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True                12m    KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  is ops req patch; ConditionStatus:True          12m    KubeDB Ops-manager Operator  is ops req patch; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   12m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is running single store; ConditionStatus:False  12m    KubeDB Ops-manager Operator  is running single store; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   12m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   12m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   12m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateAggregatorNodePVCs                        12m    KubeDB Ops-manager Operator  successfully updated Aggregator node PVC sizes
  Warning  get pod; ConditionStatus:True                   12m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is ops req patch; ConditionStatus:True          12m    KubeDB Ops-manager Operator  is ops req patch; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True                12m    KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False                  12m    KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patch; ConditionStatus:True              11m    KubeDB Ops-manager Operator  is pvc patch; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False          11m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True           11m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True                11m    KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  is ops req patch; ConditionStatus:True          11m    KubeDB Ops-manager Operator  is ops req patch; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is running single store; ConditionStatus:False  11m    KubeDB Ops-manager Operator  is running single store; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   11m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  is ops req patch; ConditionStatus:True          10m    KubeDB Ops-manager Operator  is ops req patch; ConditionStatus:True
  Warning  delete pod; ConditionStatus:True                10m    KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:False                  10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  is pvc patch; ConditionStatus:True              10m    KubeDB Ops-manager Operator  is pvc patch; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False          10m    KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pod; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   10m    KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   9m55s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   9m50s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m45s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True                   9m45s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True           9m45s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  create pod; ConditionStatus:True                9m45s  KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Warning  is ops req patch; ConditionStatus:True          9m45s  KubeDB Ops-manager Operator  is ops req patch; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m40s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m35s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m30s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m20s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Warning  get pod; ConditionStatus:True                   9m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True
  Normal   UpdateLeafNodePVCs                              9m5s   KubeDB Ops-manager Operator  successfully updated Leaf node PVC sizes
  Normal   UpdatePetSets                                   8m54s  KubeDB Ops-manager Operator  successfully reconciled the Singlestore resources
  Warning  get pet set; ConditionStatus:True               8m49s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Warning  get pet set; ConditionStatus:True               8m49s  KubeDB Ops-manager Operator  get pet set; ConditionStatus:True
  Normal   ReadyPetSets                                    8m49s  KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                                        8m49s  KubeDB Ops-manager Operator  Resuming Singlestore database: demo/sample-sdb
  Normal   Successful                                      8m49s  KubeDB Ops-manager Operator  Successfully resumed Singlestore database: demo/sample-sdb for SinglestoreOpsRequest: sdb-offline-vol-expansion

  
```

Now, we are going to verify from the `Petset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get petset -n demo sample-sdb-aggregator -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"
$ kubectl get petset -n demo sample-sdb-leaf -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"11Gi"


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-0a4b35e6-988e-4088-ae41-852ad82c5800   2Gi        RWO            Delete           Bound    demo/data-sample-sdb-aggregator-0   longhorn       <unset>                          22m
pvc-f6df5743-2bb1-4705-a2f7-be6cf7cdd7f1   11Gi       RWO            Delete           Bound    demo/data-sample-sdb-leaf-0         longhorn       <unset>                          22m
pvc-f8fee59d-74dc-46ac-9973-ff1701a6837b   11Gi       RWO            Delete           Bound    demo/data-sample-sdb-leaf-1         longhorn       <unset>                          19m
```

The above output verifies that we have successfully expanded the volume of the SingleStore database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete sdb -n demo sample-sdb
$ kubectl delete singlestoreopsrequest -n demo sdb-offline-volume-expansion
```
