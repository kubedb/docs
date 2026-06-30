---
title: DocumentDB Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: dc-auto-storage
    name: Storage Autoscaling
    parent: dc-auto-scaling
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a DocumentDB Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a `DocumentDB` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation), and the **custom metrics API** (`custom.metrics.k8s.io`) backed by the KubeDB storage-metrics apiserver. The storage autoscaler reads PVC usage from this API — `metrics-server` alone is **not** enough.

- You must have a `StorageClass` that supports volume expansion (`allowVolumeExpansion: true`).

- You should be familiar with the following `KubeDB` concepts:

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> A DocumentDB exposes the MongoDB wire protocol (port `10260`, TLS) backed by an internal PostgreSQL engine. Each pod runs the `documentdb` and `documentdb-coordinator` containers, and the data directory (`/var/pv`) lives on the per-pod PVC `data-dcdb-<ordinal>`.

## How Storage Autoscaling Works

The `DocumentDBAutoscaler` storage loop is **PVC-usage-driven**:

1. Every reconcile, the Autoscaler operator reads the `volume_used_percentage` metric for each of the DB's PVCs from `custom.metrics.k8s.io`.
2. When a PVC's usage exceeds `usageThreshold`, the operator computes a new size from `scalingRules` and creates a `VolumeExpansion` `DocumentDBOpsRequest` (capped at `upperBound`) using the configured `expansionMode`.
3. The Ops-Manager operator performs the expansion. With `expansionMode: Online` and an online-resize-capable CSI (longhorn here), the PVCs grow without taking the database offline.

> **IMPORTANT — the new size comes from `scalingRules[].threshold`, not `scalingThreshold`.** The DocumentDB storage autoscaler computes the scaled size only from `scalingRules`. The simpler top-level `scalingThreshold` field is **not** read by this controller path, so you must provide `scalingRules` or no ops request is ever created. A single rule with an empty `appliesUpto` applies to all current sizes; `threshold: 50%` grows capacity by 50%.

> **IMPORTANT — RBAC for the custom metrics API.** The autoscaler ServiceAccount must be allowed to `get`/`list` on `custom.metrics.k8s.io`. If this permission is missing, the operator logs `custom metrics API returned 403 Forbidden` and silently never creates an ops request. Add the rule to the autoscaler's ClusterRole (e.g. `kubedb-kubedb-autoscaler`):
>
> ```yaml
> - apiGroups: ["custom.metrics.k8s.io"]
>   resources: ["*"]
>   verbs: ["get", "list"]
> ```

## Storage Autoscaling of Cluster Database

At first, verify that your cluster has a storage class that supports volume expansion.

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  22d
longhorn               driver.longhorn.io      Delete          Immediate              true                   18d
```

We can see the `longhorn` storage class has `ALLOWVOLUMEEXPANSION` set to `true`, and it supports online volume expansion, so we will use it. You can install longhorn from [here](https://longhorn.io/docs/).

#### Deploy DocumentDB Cluster

In this section, we are going to deploy a `DocumentDB` cluster with 3 replicas and a small `2Gi` volume on `longhorn`. Below is the YAML of the `DocumentDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: dcdb
  namespace: demo
spec:
  version: 'pg17-0.109.0'
  storageType: Durable
  deletionPolicy: Delete
  replicas: 3
  podTemplate:
    spec:
      containers:
        - name: documentdb
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 500m
              memory: 1Gi
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

Let's create the `DocumentDB` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/autoscaler/storage/autoscaling-storage-object.yaml
documentdb.kubedb.com/dcdb created
```

Now, wait until `dcdb` has status `Ready`. i.e,

```bash
$ kubectl get docdb -n demo
NAME   NAMESPACE   VERSION        STATUS   AGE
dcdb   demo        pg17-0.109.0   Ready    2m56s
```

Let's check the PVC sizes of the cluster,

```bash
$ kubectl get pvc -n demo | grep dcdb
data-dcdb-0   Bound   pvc-de4bfaa2-ea8e-4db5-b352-72abe3ab5b67   2Gi   RWO   longhorn   <unset>   2m47s
data-dcdb-1   Bound   pvc-ad3b996c-3ffe-460c-8da3-ea14d534d217   2Gi   RWO   longhorn   <unset>   2m
data-dcdb-2   Bound   pvc-e36556ef-80aa-49ff-91ac-ad07f237e203   2Gi   RWO   longhorn   <unset>   93s
```

You can see all three PVCs have `2Gi` of storage. We are now ready to apply the `DocumentDBAutoscaler` CR to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a `DocumentDBAutoscaler` Object.

#### Create DocumentDBAutoscaler Object

In order to set up storage autoscaling for this cluster database, we have to create a `DocumentDBAutoscaler` CR with our desired configuration. Below is the YAML of the `DocumentDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DocumentDBAutoscaler
metadata:
  name: dcdb-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: dcdb
  storage:
    documentdb:
      trigger: "On"
      usageThreshold: 60
      scalingRules:
        - appliesUpto: ""
          threshold: 50%
      expansionMode: "Online"
      upperBound: 10Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing storage autoscaling on the `dcdb` database.
- `spec.storage.documentdb.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.documentdb.usageThreshold` specifies the storage usage threshold — when a PVC's usage exceeds `60%`, storage autoscaling is triggered.
- `spec.storage.documentdb.scalingRules` drives the **new size**. A rule with an empty `appliesUpto` applies to every current size, and `threshold: 50%` grows the capacity by 50%.
- `spec.storage.documentdb.expansionMode` specifies the expansion mode of the `VolumeExpansion` `DocumentDBOpsRequest`. longhorn supports online volume expansion, so it is set to `Online`.
- `spec.storage.documentdb.upperBound` caps how large the volume may ever grow (`10Gi`).

Let's create the `DocumentDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/autoscaler/storage/autoscaling-storage.yaml
documentdbautoscaler.autoscaling.kubedb.com/dcdb-storage-autoscaler created
```

#### Storage Autoscaling is set up successfully

Let's check that the `documentdbautoscaler` resource is created successfully,

```bash
$ kubectl get documentdbautoscaler -n demo
NAME                      AGE
dcdb-storage-autoscaler   8s

$ kubectl describe documentdbautoscaler dcdb-storage-autoscaler -n demo
Name:         dcdb-storage-autoscaler
Namespace:    demo
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         DocumentDBAutoscaler
Metadata:
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  DocumentDB
    Name:                  dcdb
Spec:
  Database Ref:
    Name:  dcdb
  Storage:
    Documentdb:
      Expansion Mode:  Online
      Scaling Rules:
        Applies Upto:
        Threshold:      50%
      Trigger:          On
      Upper Bound:      10Gi
      Usage Threshold:  60
Events:                 <none>
```

So, the `documentdbautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volumes to exceed the `usageThreshold` using the `dd` command. The DocumentDB data directory is mounted at `/var/pv` (PVC `data-dcdb-<ordinal>`). The autoscaler evaluates usage per PVC, so we fill all three replicas.

```bash
$ for p in dcdb-0 dcdb-1 dcdb-2; do
    kubectl exec -n demo $p -c documentdb -- sh -c 'dd if=/dev/zero of=/var/pv/_fill bs=1M count=1500; sync; df -h /var/pv'
  done
...
/dev/longhorn/pvc-de4bfaa2-ea8e-4db5-b352-72abe3ab5b67  2.0G  1.8G  180M  91% /var/pv
/dev/longhorn/pvc-ad3b996c-3ffe-460c-8da3-ea14d534d217  2.0G  1.7G  212M  90% /var/pv
/dev/longhorn/pvc-e36556ef-80aa-49ff-91ac-ad07f237e203  2.0G  1.8G  180M  90% /var/pv
```

So, from the above output the storage usage of each PVC is around `90%`, which exceeds the `usageThreshold` of `60%`.

On its next reconcile, the autoscaler reads the per-PVC usage from the custom metrics API (visible in the operator logs) and creates the ops request:

```
storage_autoscaler.go:77] Running storage Autoscaler for demo/dcdb-storage-autoscaler, referred database = dcdb
storage_metrics.go:105] LENGTH OF PVCS 3
storage_metrics.go:119] USED SPACE 89.98
storage_metrics.go:119] USED SPACE 74.039
storage_metrics.go:119] USED SPACE 88.326
client.go:88] Creating ops.kubedb.com/v1alpha1, Kind=DocumentDBOpsRequest demo/dcops-dcdb-w5q6tl.
```

Let's watch the `documentdbopsrequest` in the demo namespace to see if any `documentdbopsrequest` object is created. After some time you'll see that a `documentdbopsrequest` of type `VolumeExpansion` is created based on the `scalingRules`.

```bash
$ kubectl get documentdbopsrequest -n demo
NAME                TYPE              STATUS        AGE
dcops-dcdb-w5q6tl   VolumeExpansion   Progressing   13s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get documentdbopsrequest -n demo
NAME                TYPE              STATUS       AGE
dcops-dcdb-w5q6tl   VolumeExpansion   Successful   3m43s
```

We can see from the above output that the `DocumentDBOpsRequest` has succeeded. If we print its YAML we get an overview of the steps that were followed to expand the volume.

```bash
$ kubectl get documentdbopsrequest -n demo dcops-dcdb-w5q6tl -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: dcops-dcdb-w5q6tl
  namespace: demo
  ownerReferences:
  - apiVersion: autoscaling.kubedb.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: DocumentDBAutoscaler
    name: dcdb-storage-autoscaler
spec:
  apply: IfReady
  databaseRef:
    name: dcdb
  maxRetries: 1
  type: VolumeExpansion
  volumeExpansion:
    documentdb: "3060559872"
    mode: Online
status:
  conditions:
  - message: Volume Expansion is in progress
    reason: Running
    status: "True"
    type: Running
  - message: Successfully Set Raft Key OpsRequestProgressing
    type: SetRaftKeyOpsRequestProgressing
  - message: list pvc; ConditionStatus:True
    type: ListPvc
  - message: is pvc data-dcdb-0 updated; ConditionStatus:True
    type: IsPvcData-dcdb-0Updated
  - message: is pvc data-dcdb-1 updated; ConditionStatus:True
    type: IsPvcData-dcdb-1Updated
  - message: is pvc data-dcdb-2 updated; ConditionStatus:True
    type: IsPvcData-dcdb-2Updated
  - message: 'Online Volume Expansion performed successfully in DocumentDB pods for
      DocumentDBOpsRequest: demo/dcops-dcdb-w5q6tl'
    type: VolumeExpansion
  - message: is petset ready; ConditionStatus:True
    type: IsPetsetReady
  - message: PetSet is recreated
    type: ReadyPetSets
  - message: Successfully Expanded Volume.
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

Notice that the ops request body carries the computed size `3060559872` bytes (≈ `2.85Gi`) — the result of growing the `2Gi` volume by the `50%` `scalingRules` threshold — and `mode: Online`, so the expansion happens while the cluster stays available.

Now, let's verify from the PVCs that the volume of the cluster database has expanded.

```bash
$ kubectl get pvc -n demo | grep dcdb
data-dcdb-0   Bound   pvc-de4bfaa2-ea8e-4db5-b352-72abe3ab5b67   2920Mi   RWO   longhorn   <unset>   27m
data-dcdb-1   Bound   pvc-ad3b996c-3ffe-460c-8da3-ea14d534d217   2920Mi   RWO   longhorn   <unset>   26m
data-dcdb-2   Bound   pvc-e36556ef-80aa-49ff-91ac-ad07f237e203   2920Mi   RWO   longhorn   <unset>   26m

$ kubectl exec -n demo dcdb-0 -c documentdb -- df -h /var/pv
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/longhorn/pvc-de4bfaa2-ea8e-4db5-b352-72abe3ab5b67  2.8G  1.8G  1.1G  63% /var/pv
```

The above output verifies that we have successfully autoscaled the volume of the DocumentDB cluster database from `2Gi` to `2920Mi` (≈ `2.85Gi`). With the larger volume the same data now sits at `63%` usage, below the threshold, so no further expansion is triggered.

Finally, let's confirm the database is healthy over the MongoDB wire protocol:

```bash
$ PASS=$(kubectl get secret -n demo dcdb-auth -o jsonpath='{.data.password}' | base64 -d)
$ kubectl exec -n demo dcdb-0 -c documentdb -- mongosh \
    "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
{ ok: 1 }
```

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete documentdbautoscaler -n demo dcdb-storage-autoscaler
kubectl delete documentdb -n demo dcdb
kubectl delete ns demo
```

## Next Steps

- Learn how to autoscale the compute resources of a DocumentDB cluster in the [Compute Autoscaling](/docs/guides/documentdb/autoscaler/compute/index.md) guide.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
