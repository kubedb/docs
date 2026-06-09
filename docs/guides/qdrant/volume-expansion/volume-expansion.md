---
title: Expand Qdrant Volume
menu:
  docs_{{ .version }}:
    identifier: qdrant-volume-expansion-cluster
    name: Volume Expansion
    parent: qdrant-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand Qdrant Volume

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of a Qdrant database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [Volume Expansion Overview](/docs/guides/qdrant/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/volume-expansion](/docs/examples/qdrant/volume-expansion) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Expand Volume of Qdrant Database

Here, we are going to deploy a `Qdrant` cluster using a supported version by `KubeDB` operator. Then we are going to apply `QdrantOpsRequest` to expand its volume.

### Prepare Qdrant Database

At first verify that your cluster has a storage class that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  2d
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   3m25s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   3m19s
```

We can see from the output that `longhorn (default)` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We will use this storage class.

Now, we are going to deploy a `Qdrant` cluster with version `1.17.0`.

#### Deploy Qdrant

In this section, we are going to deploy a Qdrant cluster with 1Gi volume. Then, in the next section we will expand its volume to 3Gi using `QdrantOpsRequest` CRD. Below is the YAML of the `Qdrant` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/volume-expansion/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

Now, wait until `qdrant-sample` has status `Ready`:

```bash
$ kubectl get qdrant -n demo
NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    3m47s
```

Let's check volume size from the PetSet and from the persistent volumes:

```bash
$ kubectl get petset -n demo qdrant-sample -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
pvc-0e300ccf-49f1-4e11-b630-bc3756baeaa0   1Gi        RWO            Delete           Bound    demo/data-qdrant-sample-0      longhorn       <unset>                 4m
pvc-20ab1d50-23d7-409a-ba2e-759250f9f758   1Gi        RWO            Delete           Bound    demo/data-qdrant-sample-2      longhorn       <unset>                 4m
pvc-ccee01bf-9551-4efc-8945-5a3d25c60c7b   1Gi        RWO            Delete           Bound    demo/data-qdrant-sample-1      longhorn       <unset>                 4m
```

You can see the PetSet has 1Gi storage, and the capacity of all the persistent volumes are also 1Gi.

We are now ready to apply the `QdrantOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the Qdrant cluster.

#### Create QdrantOpsRequest

In order to expand the volume of the database, we have to create a `QdrantOpsRequest` CR with our desired volume size. Below is the YAML of the `QdrantOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-vol-exp
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: qdrant-sample
  type: VolumeExpansion
  volumeExpansion:
    mode: "Offline"
    node: 3Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `qdrant-sample` Qdrant database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.node` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`). Storageclass `longhorn` supports `Offline` volume expansion.

> **Note:** If the Storageclass you are using support `Online` Volume Expansion, Try Online volume expansion by using `spec.volumeExpansion.mode:"Online"`.

During `Online` VolumeExpansion KubeDB expands volume without deleting the pods, it directly updates the underlying PVC. And for Offline volume expansion, the database is paused. The Pods are deleted and PVC is updated. Then the database Pods are recreated with updated PVC.

Let's create the `QdrantOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/volume-expansion/ops-request.yaml
qdrantopsrequest.ops.kubedb.com/qdops-vol-exp created
```

#### Verify Qdrant volume expanded successfully

If everything goes well, `KubeDB` Ops-manager operator will update the volume size of the `Qdrant` object and related `PetSet` and `Persistent Volumes`.

Let's wait for `QdrantOpsRequest` to be `Successful`. Run the following command to watch `QdrantOpsRequest` CR,

```bash
$ kubectl get qdrantopsrequest -n demo
NAME             TYPE              STATUS       AGE
qdops-vol-exp    VolumeExpansion   Successful   10m
```

We can see from the above output that the `QdrantOpsRequest` has succeeded. If we describe the `QdrantOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe qdrantopsrequest qdops-vol-exp -n demo
Name:         qdops-vol-exp
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         QdrantOpsRequest
Spec:
  Apply:  IfReady
  Database Ref:
    Name:       qdrant-sample
  Max Retries:  1
  Type:         VolumeExpansion
  Volume Expansion:
    Mode:  Offline
    Node:  3Gi
Status:
  Conditions:
    Last Transition Time:  2026-05-15T05:20:42Z
    Message:               Qdrant ops-request has started to expand volume of Qdrant nodes
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2026-05-15T05:21:04Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2026-05-15T05:20:54Z
    Message:               get petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPetset
    Last Transition Time:  2026-05-15T05:20:54Z
    Message:               delete petset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePetset
    Last Transition Time:  2026-05-15T05:23:15Z
    Message:               successfully updated node PVC sizes
    Observed Generation:   1
    Reason:                UpdateNodePVCs
    Status:                True
    Type:                  UpdateNodePVCs
    Last Transition Time:  2026-05-15T05:23:10Z
    Message:               get pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPod
    Last Transition Time:  2026-05-15T05:21:15Z
    Message:               patch ops request; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchOpsRequest
    Last Transition Time:  2026-05-15T05:21:15Z
    Message:               delete pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeletePod
    Last Transition Time:  2026-05-15T05:21:25Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2026-05-15T05:21:26Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2026-05-15T05:23:05Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2026-05-15T05:21:45Z
    Message:               create pod; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreatePod
    Last Transition Time:  2026-05-15T05:23:25Z
    Message:               successfully reconciled the Qdrant resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-05-15T05:23:36Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2026-05-15T05:23:36Z
    Message:               Successfully completed volumeExpansion for Qdrant
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                    Age   From                         Message
  ----     ------                    ----  ----                         -------
  Normal   Starting                  3m    KubeDB Ops-manager Operator  Pausing Qdrant database demo/qdrant-sample
  Normal   Successful                3m    KubeDB Ops-manager Operator  Successfully paused Qdrant database: demo/qdrant-sample for QdrantOpsRequest: qdops-vol-exp
  Warning  delete petset             3m    KubeDB Ops-manager Operator  delete petset; ConditionStatus:True
  Normal   OrphanPetSetPods          3m    KubeDB Ops-manager Operator  successfully deleted the petSets with orphan propagation policy
  Warning  delete pod                2m    KubeDB Ops-manager Operator  delete pod; ConditionStatus:True
  Warning  patch pvc                 2m    KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  create pod                1m    KubeDB Ops-manager Operator  create pod; ConditionStatus:True
  Normal   UpdateNodePVCs           1m    KubeDB Ops-manager Operator  successfully updated node PVC sizes
  Normal   UpdatePetSets            30s    KubeDB Ops-manager Operator  successfully reconciled the Qdrant resources
  Normal   ReadyPetSets             10s    KubeDB Ops-manager Operator  PetSet is recreated
  Normal   Starting                 10s    KubeDB Ops-manager Operator  Resuming Qdrant database: demo/qdrant-sample
  Normal   Successful               10s    KubeDB Ops-manager Operator  Successfully resumed Qdrant database: demo/qdrant-sample for QdrantOpsRequest: qdops-vol-exp
```

Now, we are going to verify from the `PetSet` and `Persistent Volumes` whether the volume of the Qdrant database has expanded to meet the desired state:

```bash
$ kubectl get petset -n demo qdrant-sample -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"3Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
pvc-0e300ccf-49f1-4e11-b630-bc3756baeaa0   3Gi        RWO            Delete           Bound    demo/data-qdrant-sample-0      longhorn       <unset>                 5m
pvc-20ab1d50-23d7-409a-ba2e-759250f9f758   3Gi        RWO            Delete           Bound    demo/data-qdrant-sample-2      longhorn       <unset>                 5m
pvc-ccee01bf-9551-4efc-8945-5a3d25c60c7b   3Gi        RWO            Delete           Bound    demo/data-qdrant-sample-1      longhorn       <unset>                 5m
```

The above output verifies that we have successfully expanded the volume of the Qdrant database.

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted

$ kubectl delete qdrantopsrequest -n demo qdops-vol-exp
qdrantopsrequest.ops.kubedb.com "qdops-vol-exp" deleted
```