---
title: PerconaXtraDB Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-volume-expansion-volume-expansion
    name: PerconaXtraDB Volume Expansion
    parent: guides-perconaxtradb-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# PerconaXtraDB Volume Expansion

This guide will show you how to use `KubeDB` Enterprise operator to expand the volume of a PerconaXtraDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have a `StorageClass` that supports volume expansion.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/perconaxtradb/concepts/perconaxtradb)
  - [PerconaXtraDBOpsRequest](/docs/guides/perconaxtradb/concepts/opsrequest)
  - [Volume Expansion Overview](/docs/guides/perconaxtradb/volume-expansion/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Expand Volume of PerconaXtraDB

Here, we are going to deploy a  `PerconaXtraDB` cluster using a supported version by `KubeDB` operator. Then we are going to apply `PerconaXtraDBOpsRequest` to expand its volume. The process of expanding PerconaXtraDB `standalone` is same as PerconaXtraDB cluster.

### Prepare PerconaXtraDB Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  69s
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   37s

```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We will use this storage class. You can install topolvm from [here](https://github.com/topolvm/topolvm).

Now, we are going to deploy a `PerconaXtraDB` database of 3 replicas with version `8.0.26`.

### Deploy PerconaXtraDB

In this section, we are going to deploy a PerconaXtraDB Cluster with 1GB volume. Then, in the next section we will expand its volume to 2GB using `PerconaXtraDBOpsRequest` CRD. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "topolvm-provisioner"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut

```

Let's create the `PerconaXtraDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/volume-expansion/volume-expansion/example/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo
NAME             VERSION   STATUS   AGE
sample-pxc   8.0.26    Ready    5m4s
```

Let's check volume size from statefulset, and from the persistent volume,

```bash
$ kubectl get sts -n demo sample-pxc -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-331335d1-c8e0-4b73-9dab-dae57920e997   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-0   topolvm-provisioner            63s
pvc-b90179f8-c40a-4273-ad77-74ca8470b782   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-1   topolvm-provisioner            62s
pvc-f72411a4-80d5-4d32-b713-cb30ec662180   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-2   topolvm-provisioner            62s
```

You can see the statefulset has 1GB storage, and the capacity of all the persistent volumes are also 1GB.

We are now ready to apply the `PerconaXtraDBOpsRequest` CR to expand the volume of this database.

### Volume Expansion

Here, we are going to expand the volume of the PerconaXtraDB cluster.

#### Create PerconaXtraDBOpsRequest

In order to expand the volume of the database, we have to create a `PerconaXtraDBOpsRequest` CR with our desired volume size. Below is the YAML of the `PerconaXtraDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: md-online-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: sample-pxc
  volumeExpansion:   
    mode: "Online"
    perconaxtradb: 2Gi
```

Here,

- `spec.databaseRef.name` specifies that we are performing volume expansion operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `VolumeExpansion` on our database.
- `spec.volumeExpansion.perconaxtradb` specifies the desired volume size.
- `spec.volumeExpansion.mode` specifies the desired volume expansion mode (`Online` or `Offline`). Storageclass `topolvm-provisioner` supports `Online` volume expansion.

> **Note:** If the Storageclass you are using doesn't support `Online` Volume Expansion, Try offline volume expansion by using `spec.volumeExpansion.mode:"Offline"`.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/volume-expansion/volume-expansion/example/online-volume-expansion.yaml
perconaxtradbopsrequest.ops.kubedb.com/md-online-volume-expansion created
```

#### Verify PerconaXtraDB volume expanded successfully

If everything goes well, `KubeDB` Enterprise operator will update the volume size of `PerconaXtraDB` object and related `StatefulSets` and `Persistent Volumes`.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ kubectl get perconaxtradbopsrequest -n demo
NAME                         TYPE              STATUS       AGE
md-online-volume-expansion   VolumeExpansion   Successful   96s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe perconaxtradbopsrequest -n demo md-online-volume-expansion
Name:         md-online-volume-expansion
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  ops.kubedb.com/v1alpha1
Kind:         PerconaXtraDBOpsRequest
Metadata:
  UID:               09a119aa-4f2a-4cb4-b620-2aa3a514df11
Spec:
  Database Ref:
    Name:  sample-pxc
  Type:    VolumeExpansion
  Volume Expansion:
    Mariadb:  2Gi
    Mode:     Online
Status:
  Conditions:
    Last Transition Time:  2022-01-07T06:38:29Z
    Message:               Controller has started to Progress the PerconaXtraDBOpsRequest: demo/md-online-volume-expansion
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-07T06:39:49Z
    Message:               Online Volume Expansion performed successfully in PerconaXtraDB pod for PerconaXtraDBOpsRequest: demo/md-online-volume-expansion
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2022-01-07T06:39:49Z
    Message:               Controller has successfully expand the volume of PerconaXtraDB demo/md-online-volume-expansion
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    2m1s  KubeDB Enterprise Operator  Start processing for PerconaXtraDBOpsRequest: demo/md-online-volume-expansion
  Normal  Starting    2m1s  KubeDB Enterprise Operator  Pausing PerconaXtraDB databse: demo/sample-pxc
  Normal  Successful  2m1s  KubeDB Enterprise Operator  Successfully paused PerconaXtraDB database: demo/sample-pxc for PerconaXtraDBOpsRequest: md-online-volume-expansion
  Normal  Successful  41s   KubeDB Enterprise Operator  Online Volume Expansion performed successfully in PerconaXtraDB pod for PerconaXtraDBOpsRequest: demo/md-online-volume-expansion
  Normal  Starting    41s   KubeDB Enterprise Operator  Updating PerconaXtraDB storage
  Normal  Successful  41s   KubeDB Enterprise Operator  Successfully Updated PerconaXtraDB storage
  Normal  Starting    41s   KubeDB Enterprise Operator  Resuming PerconaXtraDB database: demo/sample-pxc
  Normal  Successful  41s   KubeDB Enterprise Operator  Successfully resumed PerconaXtraDB database: demo/sample-pxc
  Normal  Successful  41s   KubeDB Enterprise Operator  Controller has Successfully expand the volume of PerconaXtraDB: demo/sample-pxc
  
```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volumes` whether the volume of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo sample-pxc -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"2Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-331335d1-c8e0-4b73-9dab-dae57920e997   2Gi        RWO            Delete           Bound    demo/data-sample-pxc-0   topolvm-provisioner            12m
pvc-b90179f8-c40a-4273-ad77-74ca8470b782   2Gi        RWO            Delete           Bound    demo/data-sample-pxc-1   topolvm-provisioner            12m
pvc-f72411a4-80d5-4d32-b713-cb30ec662180   2Gi        RWO            Delete           Bound    demo/data-sample-pxc-2   topolvm-provisioner            12m
```

The above output verifies that we have successfully expanded the volume of the PerconaXtraDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
$ kubectl delete perconaxtradbopsrequest -n demo md-online-volume-expansion
```
