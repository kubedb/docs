---
title: Etcd Storage Autoscaling
menu:
  docs_{{ .version }}:
    identifier: etcd-autoscaling-storage-description
    name: Autoscale Storage
    parent: etcd-autoscaling-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of an Etcd Cluster

This guide shows you how to use `KubeDB` to autoscale the storage of an etcd cluster.

## Before You Begin

- At first, you need a Kubernetes cluster, and the `kubectl` command-line tool must be configured
  to communicate with your cluster.

- Install the `KubeDB` Provisioner, Ops-manager and Autoscaler operators in your cluster
  following the steps [here](/docs/setup/README.md).

- etcd support is behind an alpha feature gate. Make sure the operators are installed with
  `Etcd=true` enabled (Helm value `featureGates.Etcd=true`).

- During KubeDB installation, enable the KubeDB storage metrics server by passing the following Helm flag:

  ```bash
  --set kubedb-autoscaler.storage-metrics-server.enabled=true
  ```

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdAutoscaler](/docs/guides/etcd/concepts/etcdautoscaler.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Storage Autoscaling Overview](/docs/guides/etcd/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout
this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd)
> folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Verify a volume-expandable StorageClass

First verify that your cluster has a storage class that supports volume expansion.

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  32m
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   31m
```

The `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` set to `true`, so we can use
it. You can install topolvm from [here](https://github.com/topolvm/topolvm).

## Deploy Etcd

In this section we deploy a three member etcd cluster with version `3.6.4` on that storage class.
Below is the YAML of the `Etcd` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-autoscale
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "topolvm-provisioner"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Etcd` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/autoscaling/storage/etcd.yaml
etcd.kubedb.com/etcd-autoscale created
```

Now, wait until `etcd-autoscale` has the status `Ready`.

```bash
$ kubectl get etcd -n demo
NAME             VERSION   STATUS   AGE
etcd-autoscale   3.6.4     Ready    3m
```

Let's check the volume size from the PetSet and from the persistent volumes,

```bash
$ kubectl get petset -n demo etcd-autoscale -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS          AGE
pvc-0a2f6b3a-6a3b-4a7e-8b58-1b0d6e0a1c11   1Gi        RWO            Delete           Bound    demo/data-etcd-autoscale-0     topolvm-provisioner   3m
pvc-2c7d1b90-2f11-4d0c-b7a1-9d1f2b3c4d55   1Gi        RWO            Delete           Bound    demo/data-etcd-autoscale-1     topolvm-provisioner   3m
pvc-5e91a4c2-77c9-41bb-9f3e-6a2c7d8e9f00   1Gi        RWO            Delete           Bound    demo/data-etcd-autoscale-2     topolvm-provisioner   3m
```

The PetSet requests 1Gi, and all three persistent volumes are 1Gi. We are now ready to apply the
`EtcdAutoscaler` CRO.

### Storage Autoscaling

#### Create EtcdAutoscaler Object

To set up storage autoscaling for this etcd cluster, we create an `EtcdAutoscaler` CRO with the
desired configuration. Below is the YAML of the `EtcdAutoscaler` object that we are going to
create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: EtcdAutoscaler
metadata:
  name: etcd-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: etcd-autoscale
  storage:
    etcd:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
      expansionMode: "Online"
      upperBound: 100Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing storage autoscaling on the
  `etcd-autoscale` database.
- `spec.storage.etcd.trigger` specifies that storage autoscaling is enabled for this database.
  The default is `Off`.
- `spec.storage.etcd.usageThreshold` specifies the PVC usage threshold in percent. If no member's
  volume is more than `60%` full, nothing happens.
- `spec.storage.etcd.scalingThreshold` specifies by how much, in percent, the volume is grown once
  the usage threshold is crossed — here `50%`, so a 1Gi volume becomes roughly 1.5Gi.
- `spec.storage.etcd.expansionMode` selects the mode of the generated volume expansion
  `EtcdOpsRequest`: `Online` or `Offline`. Use `Online` when your CSI driver supports expanding a
  mounted volume, as topolvm does.
- `spec.storage.etcd.upperBound` caps the volume growth. If the newly computed size would exceed
  this, no ops request is created at all — so pick it deliberately.

> **Note:** unlike some other KubeDB autoscalers, `EtcdAutoscaler` does not currently ship a
> defaulting webhook, so the documented defaults for these fields are **not** filled in for you.
> Set `trigger`, `usageThreshold`, `expansionMode` and either `scalingThreshold` or
> `scalingRules` explicitly.

##### Using scalingRules instead of a flat percentage

A flat percentage is a poor fit once volumes get large — growing a 500Gi volume by 50% is a 250Gi
jump. `scalingRules` lets the growth step depend on how much is currently used. Rules are matched
in order of their `appliesUpto` size, and an entry with an empty `appliesUpto` acts as the
catch-all. A threshold ending in `pc` or `%` is relative; anything else is parsed as an absolute
quantity that is added to the current capacity.

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: EtcdAutoscaler
metadata:
  name: etcd-storage-autoscaler-rules
  namespace: demo
spec:
  databaseRef:
    name: etcd-autoscale
  storage:
    etcd:
      trigger: "On"
      usageThreshold: 60
      expansionMode: "Online"
      upperBound: 500Gi
      scalingRules:
        - appliesUpto: "50Gi"
          threshold: "50pc"
        - appliesUpto: "200Gi"
          threshold: "25pc"
        - appliesUpto: ""
          threshold: "50Gi"
```

That reads as: while less than 50Gi is used, grow by 50%; between 50Gi and 200Gi, grow by 25%;
beyond that, add a flat 50Gi at a time — never going past the 500Gi `upperBound`.

Let's create the `EtcdAutoscaler` CR we showed first,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/autoscaling/storage/etcd-storage-autoscaler.yaml
etcdautoscaler.autoscaling.kubedb.com/etcd-storage-autoscaler created
```

#### Verify autoscaling is set up successfully

Let's check that the `etcdautoscaler` resource was created successfully,

```bash
$ kubectl get etcdautoscaler -n demo
NAME                      AGE
etcd-storage-autoscaler   40s

$ kubectl describe etcdautoscaler etcd-storage-autoscaler -n demo
Name:         etcd-storage-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         EtcdAutoscaler
Metadata:
  Creation Timestamp:  2026-02-11T10:02:11Z
  Generation:          1
  Resource Version:    52140
  UID:                 8a2b6c1e-4f7a-4d20-9b18-70c4f2d5a3b9
Spec:
  Database Ref:
    Name:  etcd-autoscale
  Storage:
    Etcd:
      Expansion Mode:      Online
      Scaling Threshold:   50
      Trigger:             On
      Upper Bound:         100Gi
      Usage Threshold:     60
Events:                    <none>
```

So the `EtcdAutoscaler` resource was created successfully.

#### Watch the generated EtcdOpsRequest

Now write enough data into the cluster that at least one member's volume crosses the 60% usage
threshold. Once the Autoscaler operator observes that, it creates an `EtcdOpsRequest` of type
`VolumeExpansion`. Note the `etcdops-` name prefix, which is how the Autoscaler operator names
the requests it generates.

```bash
$ kubectl get etcdopsrequest -n demo
NAME                            TYPE              STATUS        AGE
etcdops-etcd-autoscale-w7q2rk   VolumeExpansion   Progressing   20s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get etcdopsrequest -n demo
NAME                            TYPE              STATUS       AGE
etcdops-etcd-autoscale-w7q2rk   VolumeExpansion   Successful   2m
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed to expand
the volumes. The owner reference back to the `EtcdAutoscaler` is what tells you this request was
generated rather than hand-written.

```bash
$ kubectl describe etcdopsrequest -n demo etcdops-etcd-autoscale-w7q2rk
Name:         etcdops-etcd-autoscale-w7q2rk
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=etcd-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=etcds.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-02-11T10:14:52Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  EtcdAutoscaler
    Name:                  etcd-storage-autoscaler
    UID:                   8a2b6c1e-4f7a-4d20-9b18-70c4f2d5a3b9
  Resource Version:        53310
  UID:                     b1e7d2a4-9c8f-4b12-a0d7-53f9c1e2b6aa
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  etcd-autoscale
  Type:   VolumeExpansion
  Volume Expansion:
    Etcd:  1610612736
    Mode:  Online
Status:
  Conditions:
    Last Transition Time:  2026-02-11T10:14:52Z
    Message:               Volume Expansion is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-02-11T10:14:55Z
    Message:               3
    Observed Generation:   1
    Reason:                PetSetReplicasBeforeExpansion
    Status:                True
    Type:                  PetSetReplicasBeforeExpansion
    Last Transition Time:  2026-02-11T10:16:38Z
    Message:               Successfully expanded the etcd data volumes
    Observed Generation:   1
    Reason:                UpdateEtcdNodePVCs
    Status:                True
    Type:                  UpdateEtcdNodePVCs
    Last Transition Time:  2026-02-11T10:16:39Z
    Message:               Successfully updated the Etcd storage request
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-02-11T10:16:44Z
    Message:               PetSet has been recreated with the expanded volume claim template
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2026-02-11T10:16:45Z
    Message:               Successfully expanded the etcd data volumes
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

> The size in `spec.volumeExpansion.etcd` is written as a plain byte count because the autoscaler
> computes it arithmetically from the observed capacity; `1610612736` is 1.5Gi.

#### Verify the new volume size

Now let's verify from the `PetSet` and the `PersistentVolume`s that the volumes were expanded.

```bash
$ kubectl get petset -n demo etcd-autoscale -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1610612736"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS          AGE
pvc-0a2f6b3a-6a3b-4a7e-8b58-1b0d6e0a1c11   2Gi        RWO            Delete           Bound    demo/data-etcd-autoscale-0     topolvm-provisioner   18m
pvc-2c7d1b90-2f11-4d0c-b7a1-9d1f2b3c4d55   2Gi        RWO            Delete           Bound    demo/data-etcd-autoscale-1     topolvm-provisioner   18m
pvc-5e91a4c2-77c9-41bb-9f3e-6a2c7d8e9f00   2Gi        RWO            Delete           Bound    demo/data-etcd-autoscale-2     topolvm-provisioner   18m
```

The output above confirms that the storage of the etcd cluster was autoscaled. (The reported PV
capacity is rounded up to whatever unit the CSI driver allocates in, so it may be a little larger
than the requested size.)

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdautoscaler -n demo etcd-storage-autoscaler
kubectl delete etcd -n demo etcd-autoscale
kubectl delete ns demo
```

## Next Steps

- [Compute Resource Autoscaling of an Etcd cluster](/docs/guides/etcd/autoscaler/compute/compute-autoscale.md).
- [Volume Expansion of an Etcd cluster](/docs/guides/etcd/volume-expansion/volume-expansion.md) — the same operation, applied by hand.
- Reclaim space inside the etcd backend with [Compact](/docs/guides/etcd/maintenance/compact.md)
  and [Defragment](/docs/guides/etcd/maintenance/defragment.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
