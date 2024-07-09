---
title: PerconaXtraDB Cluster Autoscaling
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-autoscaling-storage-cluster
    name: Cluster
    parent: guides-perconaxtradb-autoscaling-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a PerconaXtraDB Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a PerconaXtraDB Replicaset database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb)
  - [PerconaXtraDBAutoscaler](/docs/guides/percona-xtradb/concepts/autoscaler)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest)
  - [Storage Autoscaling Overview](/docs/guides/percona-xtradb/autoscaler/storage/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Storage Autoscaling of Cluster Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  79m
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   78m
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `PerconaXtraDB` replicaset using a supported version by `KubeDB` operator. Then we are going to apply `PerconaXtraDBAutoscaler` to set up autoscaling.

#### Deploy PerconaXtraDB Cluster

In this section, we are going to deploy a PerconaXtraDB replicaset database with version `8.0.26`.  Then, in the next section we will set up autoscaling for this database using `PerconaXtraDBAutoscaler` CRD. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,

> If you want to autoscale PerconaXtraDB `Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

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
  deletionPolicy: WipeOut
```

Let's create the `PerconaXtraDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/autoscaler/storage/cluster/examples/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo
NAME             VERSION   STATUS   AGE
sample-pxc      8.0.26    Ready    3m46s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo sample-pxc -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-2   topolvm-provisioner            57s
pvc-4a509b05-774b-42d9-b36d-599c9056af37   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-0   topolvm-provisioner            58s
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   1Gi        RWO            Delete           Bound    demo/data-sample-pxc-1   topolvm-provisioner            57s
```

You can see the petset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `PerconaXtraDBAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a PerconaXtraDBAutoscaler Object.

#### Create PerconaXtraDBAutoscaler Object

In order to set up vertical autoscaling for this replicaset database, we have to create a `PerconaXtraDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `PerconaXtraDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: PerconaXtraDBAutoscaler
metadata:
  name: px-as-st
  namespace: demo
spec:
  databaseRef:
    name: sample-pxc
  storage:
    perconaxtradb:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `sample-pxc` database.
- `spec.storage.perconaxtradb.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.perconaxtradb.usageThreshold` specifies storage usage threshold, if storage usage exceeds `20%` then storage autoscaling will be triggered.
- `spec.storage.perconaxtradb.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `20%` of the current amount.
- `spec.storage.perconaxtradb.expansionMode` specifies the expansion mode of volume expansion `PerconaXtraDBOpsRequest` created by `PerconaXtraDBAutoscaler`. topolvm-provisioner supports online volume expansion so here `expansionMode` is set as "Online".

Let's create the `PerconaXtraDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/autoscaler/storage/cluster/examples/pxas-storage.yaml
perconaxtradbautoscaler.autoscaling.kubedb.com/px-as-st created
```

#### Storage Autoscaling is set up successfully

Let's check that the `perconaxtradbautoscaler` resource is created successfully,

```bash
$ kubectl get perconaxtradbautoscaler -n demo
NAME           AGE
px-as-st   33s

$ kubectl describe perconaxtradbautoscaler px-as-st -n demo
Name:         px-as-st
Namespace:    demo
Labels:       <none>
Annotations:  API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         PerconaXtraDBAutoscaler
Metadata:
  Creation Timestamp:  2022-01-14T06:08:02Z
  Generation:          1
  Managed Fields:
    ...
  Resource Version:  24009
  UID:               4f45a3b3-fc72-4d04-b52c-a770944311f6
Spec:
  Database Ref:
    Name:  sample-pxc
  Storage:
    Mariadb:
      Scaling Threshold:  20
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>
```

So, the `perconaxtradbautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume(`var/lib/mysql`) using the following commands:

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
root@sample-pxc-0:/ df -h /var/lib/mysql
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M  357M  658M  36% /var/lib/mysql
root@sample-pxc-0:/ dd if=/dev/zero of=/var/lib/mysql/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 0.340877 s, 1.5 GB/s
root@sample-pxc-0:/ df -h /var/lib/mysql
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M  857M  158M  85% /var/lib/mysql
```

So, from the above output we can see that the storage usage is 83%, which exceeded the `usageThreshold` 20%.

Let's watch the `perconaxtradbopsrequest` in the demo namespace to see if any `perconaxtradbopsrequest` object is created. After some time you'll see that a `perconaxtradbopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ kubectl get perconaxtradbopsrequest -n demo
NAME                         TYPE              STATUS        AGE
mops-sample-pxc-xojkua   VolumeExpansion   Progressing   15s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get perconaxtradbopsrequest -n demo
NAME                         TYPE              STATUS       AGE
mops-sample-pxc-xojkua   VolumeExpansion   Successful   97s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe perconaxtradbopsrequest -n demo mops-sample-pxc-xojkua
Name:         mops-sample-pxc-xojkua
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sample-pxc
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=perconaxtradbs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PerconaXtraDBOpsRequest
Metadata:
  Creation Timestamp:  2022-01-14T06:13:10Z
  Generation:          1
  Managed Fields: ...
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  PerconaXtraDBAutoscaler
    Name:                  px-as-st
    UID:                   4f45a3b3-fc72-4d04-b52c-a770944311f6
  Resource Version:        25557
  UID:                     90763a49-a03f-407c-a233-fb20c4ab57d7
Spec:
  Database Ref:
    Name:  sample-pxc
  Type:    VolumeExpansion
  Volume Expansion:
    Mariadb:  1594884096
Status:
  Conditions:
    Last Transition Time:  2022-01-14T06:13:10Z
    Message:               Controller has started to Progress the PerconaXtraDBOpsRequest: demo/mops-sample-pxc-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Volume Expansion performed successfully in PerconaXtraDB pod for PerconaXtraDBOpsRequest: demo/mops-sample-pxc-xojkua
    Observed Generation:   1
    Reason:                SuccessfullyVolumeExpanded
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2022-01-14T06:14:25Z
    Message:               Controller has successfully expand the volume of PerconaXtraDB demo/mops-sample-pxc-xojkua
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     3
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Start processing for PerconaXtraDBOpsRequest: demo/mops-sample-pxc-xojkua
  Normal  Starting    2m58s  KubeDB Enterprise Operator  Pausing PerconaXtraDB databse: demo/sample-pxc
  Normal  Successful  2m58s  KubeDB Enterprise Operator  Successfully paused PerconaXtraDB database: demo/sample-pxc for PerconaXtraDBOpsRequest: mops-sample-pxc-xojkua
  Normal  Successful  103s   KubeDB Enterprise Operator  Volume Expansion performed successfully in PerconaXtraDB pod for PerconaXtraDBOpsRequest: demo/mops-sample-pxc-xojkua
  Normal  Starting    103s   KubeDB Enterprise Operator  Updating PerconaXtraDB storage
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully Updated PerconaXtraDB storage
  Normal  Starting    103s   KubeDB Enterprise Operator  Resuming PerconaXtraDB database: demo/sample-pxc
  Normal  Successful  103s   KubeDB Enterprise Operator  Successfully resumed PerconaXtraDB database: demo/sample-pxc
  Normal  Successful  103s   KubeDB Enterprise Operator  Controller has Successfully expand the volume of PerconaXtraDB: demo/sample-pxc
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the replicaset database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo sample-pxc -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   2Gi        RWO            Delete           Bound    demo/data-sample-pxc-2   topolvm-provisioner            23m
pvc-4a509b05-774b-42d9-b36d-599c9056af37   2Gi        RWO            Delete           Bound    demo/data-sample-pxc-0   topolvm-provisioner            24m
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   2Gi        RWO            Delete           Bound    demo/data-sample-pxc-1   topolvm-provisioner            23m
```

The above output verifies that we have successfully autoscaled the volume of the PerconaXtraDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete perconaxtradb -n demo sample-pxc
kubectl delete perconaxtradbautoscaler -n demo px-as-st
kubectl delete ns demo
```
