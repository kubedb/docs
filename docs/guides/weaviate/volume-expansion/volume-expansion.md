---
title: Expand Weaviate Volume
menu:
  docs_{{ .version }}:
    identifier: weaviate-volume-expansion-cluster
    name: Volume Expansion
    parent: weaviate-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand Weaviate Volume

This guide will show you how to use the `KubeDB` Ops Manager to expand the volume of a Weaviate cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You need a `StorageClass` that supports volume expansion. Verify with:

  ```bash
  $ kubectl get storageclass
  NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  38h
  longhorn               driver.longhorn.io      Delete          Immediate              true                   30m
  ```

  Here, the `longhorn` StorageClass has `ALLOWVOLUMEEXPANSION` set to `true`.

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Volume Expansion Overview](/docs/guides/weaviate/volume-expansion/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/volume-expansion](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/volume-expansion) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate

In this section, we are going to deploy a Weaviate cluster with `1Gi` storage using the `longhorn` StorageClass.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR and wait for it to become `Ready`. Then check the PVCs:

```bash
$ kubectl get pvc -n demo
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-weaviate-sample-0   Bound    pvc-b8b6d9e6-634f-4ead-b1fd-1bfe549976e4   1Gi        RWO            longhorn       <unset>                 5m
data-weaviate-sample-1   Bound    pvc-4e9329c0-8a3d-4402-919c-afa4fe2144c9   1Gi        RWO            longhorn       <unset>                 5m
data-weaviate-sample-2   Bound    pvc-a846947c-212f-4aea-92a7-c8f88ae7f463   1Gi        RWO            longhorn       <unset>                 5m
```

Each PVC has `1Gi` of storage.

## Expand Volume of the Weaviate Cluster

Here, we are going to expand the volume of the cluster from `1Gi` to `3Gi`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wv-volume-expansion-offline
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: weaviate-sample
  volumeExpansion:
    node: 3Gi
    mode: Offline
```

- `spec.type` specifies that this is a `VolumeExpansion` operation.
- `spec.volumeExpansion.node` specifies the desired size of the volume after expansion.
- `spec.volumeExpansion.mode` specifies the expansion mode, either `Online` or `Offline`.

Let's create the `WeaviateOpsRequest` CR:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/volume-expansion/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/wv-volume-expansion-offline created
```

The Ops Manager will expand the PVCs and reconcile the cluster.

```bash
$ kubectl get weaviateopsrequest -n demo wv-volume-expansion-offline
NAME                          TYPE              STATUS       AGE
wv-volume-expansion-offline   VolumeExpansion   Successful   2m
```

Let's check the `status.conditions` of the `WeaviateOpsRequest`:

```bash
$ kubectl get weaviateopsrequest -n demo wv-volume-expansion-offline -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wv-volume-expansion-offline
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: weaviate-sample
  maxRetries: 1
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    node: 3Gi
status:
  conditions:
  - message: Weaviate ops-request has started to expand volume of Weaviate nodes
    reason: VolumeExpansion
    status: "True"
    type: VolumeExpansion
  - message: get petset; ConditionStatus:True
    status: "True"
    type: GetPetset
  - message: delete petset; ConditionStatus:True
    status: "True"
    type: DeletePetset
  - message: successfully deleted the petSets with orphan propagation policy
    reason: OrphanPetSetPods
    status: "True"
    type: OrphanPetSetPods
  - message: get pvc; ConditionStatus:True
    status: "True"
    type: GetPvc
  - message: patch pvc; ConditionStatus:True
    status: "True"
    type: PatchPvc
  - message: compare storage; ConditionStatus:True
    status: "True"
    type: CompareStorage
  - message: successfully updated node PVC sizes
    reason: UpdateNodePVCs
    status: "True"
    type: UpdateNodePVCs
  - message: successfully reconciled the Weaviate resources
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: PetSet is recreated
    reason: ReadyPetSets
    status: "True"
    type: ReadyPetSets
  - message: Successfully completed volumeExpansion for Weaviate
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

Now, let's verify that the PVCs have been expanded to `3Gi`:

```bash
$ kubectl get pvc -n demo
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-weaviate-sample-0   Bound    pvc-b8b6d9e6-634f-4ead-b1fd-1bfe549976e4   3Gi        RWO            longhorn       <unset>                 14m
data-weaviate-sample-1   Bound    pvc-4e9329c0-8a3d-4402-919c-afa4fe2144c9   3Gi        RWO            longhorn       <unset>                 10m
data-weaviate-sample-2   Bound    pvc-a846947c-212f-4aea-92a7-c8f88ae7f463   3Gi        RWO            longhorn       <unset>                 10m

$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.storage.resources.requests.storage}'
3Gi
```

The volume has been expanded successfully.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- Automatically scale the storage with the [Storage Autoscaler](/docs/guides/weaviate/autoscaler/storage/storage-autoscale.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete weaviateopsrequest -n demo wv-volume-expansion-offline
$ kubectl delete weaviate -n demo weaviate-sample
$ kubectl delete ns demo
```
