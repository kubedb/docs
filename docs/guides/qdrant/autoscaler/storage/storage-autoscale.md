---
title: Qdrant Storage Autoscaler
menu:
  docs_{{ .version }}:
    identifier: qdrant-autoscaler-storage-description
    name: Autoscale Storage
    parent: qdrant-autoscaler-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Qdrant Database

This guide will show you how to use `KubeDB` to autoscale the storage of a Qdrant database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation).

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/)
  - [QdrantAutoscaler](/docs/guides/qdrant/concepts/autoscaler.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [Storage Autoscaling Overview](/docs/guides/qdrant/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

## Storage Autoscaling of Database

At first, verify that your cluster has a storage class that supports volume expansion:

```bash
kubectl get storageclass
```
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  28d
longhorn (default)     driver.longhorn.io      Delete          Immediate              true                   25d
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   28d

We can see from the output that `longhorn` storage class has `ALLOWVOLUMEEXPANSION` set to `true`. We will use it for this tutorial.

Now, we are going to deploy a `Qdrant` database using a supported version by `KubeDB` operator. Then we are going to apply `QdrantAutoscaler` to set up autoscaling.

### Deploy Qdrant Database

In this section, we are going to deploy a Qdrant database with version `1.17.0`. Then, in the next section we will set up autoscaling for this database using `QdrantAutoscaler` CRD. Below is the YAML of the `Qdrant` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` CR we have shown above:

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/autoscaler/storage/qdrant.yaml
```
qdrant.kubedb.com/qdrant-sample created

Now, wait until `qdrant-sample` has status `Ready`:

```bash
kubectl get qdrant -n demo
```
NAME            VERSION   STATUS   AGE
qdrant-sample   1.17.0    Ready    101s

Let's check the volume size from the Petset and from the persistent volumes:

```bash
kubectl get petset -n demo qdrant-sample -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
```
"1Gi"

```bash
kubectl get pv -o custom-columns=NAME:.metadata.name,CAPACITY:.spec.capacity.storage,STORAGECLASS:.spec.storageClassName,CLAIM:.spec.claimRef.name | grep qdrant-sample
```
pvc-31485d1d-5048-4dc2-a2c3-18910b27b661   1Gi        longhorn       data-qdrant-sample-0
pvc-683755b9-023d-4d36-8318-8a22d5b79acb   1Gi        longhorn       data-qdrant-sample-1
pvc-d494f0aa-41b8-458d-ab56-39946c9b9bfe   1Gi        longhorn       data-qdrant-sample-2

You can see the Petset has 1GB storage and the capacity of all the persistent volumes is also 1GB.

We are now ready to apply the `QdrantAutoscaler` CRD to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a `QdrantAutoscaler` Object.

#### Create QdrantAutoscaler Object

In order to set up storage autoscaling for this database, we have to create a `QdrantAutoscaler` CR with our desired configuration. Below is the YAML of the `QdrantAutoscaler` object that we are going to create:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: QdrantAutoscaler
metadata:
  name: qdrant-as-storage
  namespace: demo
spec:
  databaseRef:
    name: qdrant-sample
  storage:
    node:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

Here,

- `spec.databaseRef.name` specifies that we are performing storage autoscaling on `qdrant-sample` database.
- `spec.storage.node.trigger` specifies that storage autoscaling is enabled for the Qdrant nodes.
- `spec.storage.node.usageThreshold` specifies the storage usage threshold — if storage usage exceeds `20%`, storage autoscaling will be triggered.
- `spec.storage.node.scalingThreshold` specifies the scaling threshold — storage will be scaled to `20%` of the current amount.
- `spec.storage.node.expansionMode` specifies the expansion mode of the volume expansion `QdrantOpsRequest` created by `QdrantAutoscaler`. longhorn supports online volume expansion so `expansionMode` is set to `Online`.

Let's create the `QdrantAutoscaler` CR we have shown above:

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/autoscaler/storage/qdrant-as-storage.yaml
```
qdrantautoscaler.autoscaling.kubedb.com/qdrant-as-storage created

#### Verify Autoscaler is set up successfully

Let's check that the `QdrantAutoscaler` resource is created successfully:

```bash
kubectl get qdrantautoscaler -n demo
```
NAME                AGE
qdrant-as-storage   33s

```bash
kubectl describe qdrantautoscaler qdrant-as-storage -n demo
```
Name:         qdrant-as-storage
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         QdrantAutoscaler
Spec:
  Database Ref:
    Name:  qdrant-sample
  Storage:
    Node:
      Expansion Mode:   Online
      Scaling Threshold:  20
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>

So, the `QdrantAutoscaler` resource is created successfully. The operator will now continuously watch the storage usage of the Qdrant pods. When the usage crosses the `usageThreshold`, it will create a `QdrantOpsRequest` to expand the storage.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using the `dd` command to see if storage autoscaling is working:

```bash
kubectl exec -n demo qdrant-sample-0 -- df -h /qdrant/storage
```
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-9d79a391-6777-4be5-8f9e-0139d178aada  974M  296K  958M   1% /qdrant/storage

```bash
kubectl exec -n demo qdrant-sample-0 -- bash -c "dd if=/dev/zero of=/qdrant/storage/file.img bs=250M count=1 && df -h /qdrant/storage"
```
1+0 records in
1+0 records out
262144000 bytes (262 MB, 250 MiB) copied, 2.01673 s, 130 MB/s
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-9d79a391-6777-4be5-8f9e-0139d178aada  974M  251M  708M  27% /qdrant/storage

Now let's watch the `QdrantOpsRequest` in the demo namespace:

```bash
kubectl get qdrantopsrequest -n demo -w
```
NAME                              TYPE              STATUS        AGE
qdops-qdrant-sample-ka4wgv        VolumeExpansion   Progressing   2s
qdops-qdrant-sample-ka4wgv        VolumeExpansion   Successful    8m

After the `QdrantOpsRequest` completes successfully, let's check the updated storage:

```bash
kubectl get pv -o custom-columns=NAME:.metadata.name,CAPACITY:.spec.capacity.storage,STORAGECLASS:.spec.storageClassName,CLAIM:.spec.claimRef.name | grep qdrant-sample
```
pvc-31485d1d-5048-4dc2-a2c3-18910b27b661   1168Mi    longhorn       data-qdrant-sample-0
pvc-683755b9-023d-4d36-8318-8a22d5b79acb   1168Mi    longhorn       data-qdrant-sample-1
pvc-d494f0aa-41b8-458d-ab56-39946c9b9bfe   1168Mi    longhorn       data-qdrant-sample-2

The storage has been automatically scaled from 1Gi to ~1168Mi as we specified a `scalingThreshold` of 20%.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete qdrantautoscaler -n demo qdrant-as-storage
kubectl delete ns demo
```
