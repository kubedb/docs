---
title: Weaviate Storage Autoscaler
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-storage-description
    name: Autoscale Storage
    parent: weaviate-autoscaler-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Weaviate Database

This guide will show you how to use `KubeDB` to auto-scale the storage of a Weaviate database when the volumes start filling up.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-Manager, and Autoscaler operators in your cluster following the steps [here](/docs/setup/README.md).

- You must have a `StorageClass` that supports **volume expansion** (`allowVolumeExpansion: true`). Storage autoscaling expands the existing PVCs in place.

- Storage autoscaling reacts to the volume-usage metric (`volume_used_percentage`) exposed through KubeDB's metrics API. Make sure the KubeDB metrics stack is installed and serving this metric in your cluster.

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Storage Autoscaling Overview](/docs/guides/weaviate/autoscaler/storage/overview.md)
  - [Volume Expansion](/docs/guides/weaviate/volume-expansion/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

## Storage Autoscaling of Database

Here, we are going to deploy a `Weaviate` database and then set up storage autoscaling with a `WeaviateAutoscaler`.

### Deploy Weaviate Database

In this section, we are going to deploy a Weaviate database with `1Gi` of storage on a volume-expansion-capable StorageClass (`longhorn`):

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

Let's create the `Weaviate` CR and wait for it to become `Ready`. Then check the current storage:

```bash
kubectl get pvc -n demo -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage
```
NAME                     SIZE
data-weaviate-sample-0   1Gi
data-weaviate-sample-1   1Gi
data-weaviate-sample-2   1Gi

### Create WeaviateAutoscaler

Now, we are going to set up storage autoscaling using a `WeaviateAutoscaler` object. Note the storage knob is under `spec.storage.weaviate`:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: WeaviateAutoscaler
metadata:
  name: weaviate-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: weaviate-sample
  storage:
    weaviate:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 50
      expansionMode: "Online"
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
```

Here,

- `spec.databaseRef.name` specifies that we are performing storage autoscaling on the `weaviate-sample` database.
- `spec.storage.weaviate.trigger` enables storage autoscaling for the Weaviate nodes.
- `spec.storage.weaviate.usageThreshold` specifies the used-space percentage (here, `20%`) that triggers an expansion.
- `spec.storage.weaviate.scalingThreshold` specifies the percentage by which the volume is expanded each time (here, `50%`).
- `spec.storage.weaviate.expansionMode` specifies whether the expansion is `Online` or `Offline`.
- `spec.opsRequestOptions` controls how the generated ops request is applied (`apply: IfReady`) and its timeout.

Let's create the `WeaviateAutoscaler`:

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/autoscaler/storage/weaviate-storage-autoscaler.yaml
```
weaviateautoscaler.autoscaling.kubedb.com/weaviate-storage-autoscaler created

### Verify Autoscaler is Set Up

Let's describe the `WeaviateAutoscaler` to confirm it is configured and watching the volumes:

```bash
kubectl describe weaviateautoscaler -n demo weaviate-storage-autoscaler
```
Name:         weaviate-storage-autoscaler
Namespace:    demo
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         WeaviateAutoscaler
Metadata:
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Controller:            true
    Kind:                  Weaviate
    Name:                  weaviate-sample
Spec:
  Database Ref:
    Name:  weaviate-sample
  Ops Request Options:
    Apply:        IfReady
    Max Retries:  1
    Timeout:      10m0s
  Storage:
    Weaviate:
      Expansion Mode:  Online
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>

The autoscaler is now watching the PVC usage of the Weaviate pods.

### Trigger an Expansion

When a volume's used space crosses the `usageThreshold` (20%), the autoscaler operator creates a `WeaviateOpsRequest` of type `VolumeExpansion` that grows the volume by `scalingThreshold` (50%). For example, after writing enough data to fill more than 20% of a `1Gi` volume:

# usage on each node's data volume crosses 20%
```bash
kubectl exec -n demo weaviate-sample-0 -c weaviate -- df -h /var/lib/weaviate
```
Filesystem                Size      Used Available Use% Mounted on
/dev/longhorn/pvc-...   973.4M    401.7M    555.7M  42% /var/lib/weaviate

the autoscaler creates a `VolumeExpansion` ops request:

```bash
kubectl get weaviateopsrequest -n demo
```
NAME                                TYPE              STATUS       AGE
wvops-weaviate-sample-xxxxxx        VolumeExpansion   Successful   3m

and the PVCs are expanded (here, from `1Gi` to `1.5Gi` — a 50% increase):

```bash
kubectl get pvc -n demo -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage
```
NAME                     SIZE
data-weaviate-sample-0   1531584Ki
data-weaviate-sample-1   1531584Ki
data-weaviate-sample-2   1531584Ki

> **Note:** The auto-trigger relies on the `volume_used_percentage` metric being available through KubeDB's metrics API. If that metric is not exposed in your cluster, the autoscaler will stay configured and watching but will not generate an ops request. The underlying `VolumeExpansion` mechanism it uses is the same one demonstrated step-by-step in the [Volume Expansion](/docs/guides/weaviate/volume-expansion/volume-expansion.md) guide.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviateautoscaler -n demo weaviate-storage-autoscaler
```

```bash
kubectl delete weaviate -n demo weaviate-sample
```

```bash
kubectl delete ns demo
```
