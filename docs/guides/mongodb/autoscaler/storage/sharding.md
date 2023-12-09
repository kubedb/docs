---
title: MongoDB Shard Autoscaling
menu:
  docs_{{ .version }}:
    identifier: mg-storage-auto-scaling-shard
    name: Sharding
    parent: mg-storage-auto-scaling
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Storage Autoscaling of a MongoDB Sharded Database

This guide will show you how to use `KubeDB` to autoscale the storage of a MongoDB Sharded database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
    - [MongoDBAutoscaler](/docs/guides/mongodb/concepts/autoscaler.md)
    - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
    - [Storage Autoscaling Overview](/docs/guides/mongodb/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mongodb](/docs/examples/mongodb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of Sharded Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  9h
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   9h
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `MongoDB` sharded database using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBAutoscaler` to set up autoscaling.

#### Deploy MongoDB Sharded Database

In this section, we are going to deploy a MongoDB sharded database with version `4.4.26`.  Then, in the next section we will set up autoscaling for this database using `MongoDBAutoscaler` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-sh
  namespace: demo
spec:
  version: "4.4.26"
  storageType: Durable
  shardTopology:
    configServer:
      storage:
        storageClassName: topolvm-provisioner
        resources:
          requests:
            storage: 1Gi
      replicas: 3
    mongos:
      replicas: 2
    shard:
      storage:
        storageClassName: topolvm-provisioner
        resources:
          requests:
            storage: 1Gi
      replicas: 3
      shards: 2
  terminationPolicy: WipeOut
```

Let's create the `MongoDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/storage/mg-sh.yaml
mongodb.kubedb.com/mg-sh created
```

Now, wait until `mg-sh` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME      VERSION    STATUS    AGE
mg-sh     4.4.26      Ready     3m51s
```

Let's check volume size from one of the shard statefulset, and from the persistent volume,

```bash
$ kubectl get sts -n demo mg-sh-shard0 -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                            STORAGECLASS          REASON   AGE
pvc-031836c6-95ae-4015-938c-da183c205828   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-configsvr-0   topolvm-provisioner            5m1s
pvc-2515233f-0f7d-4d0d-8b45-97a3cb9d4488   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard0-2      topolvm-provisioner            3m44s
pvc-35f73708-3c11-4ead-a60b-e1679a294b81   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard0-0      topolvm-provisioner            5m
pvc-4b329feb-8c92-4605-a37e-c02b3499e311   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-configsvr-2   topolvm-provisioner            3m55s
pvc-52490270-1355-4045-b2a1-872a671ab006   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-configsvr-1   topolvm-provisioner            4m28s
pvc-80dc91d3-f56f-4037-b6e1-f69e13fb434c   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard1-1      topolvm-provisioner            4m26s
pvc-c1965a32-7471-4885-ac52-f9eab056d48e   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard1-2      topolvm-provisioner            3m57s
pvc-c838a27d-c75d-4caa-9c1d-456af3bfaba0   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard1-0      topolvm-provisioner            4m59s
pvc-d47f19be-f206-41c5-a0b1-5022776fea2f   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard0-1      topolvm-provisioner            4m25s
```

You can see the statefulset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `MongoDBAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a MongoDBAutoscaler Object.

#### Create MongoDBAutoscaler Object

In order to set up vertical autoscaling for this sharded database, we have to create a `MongoDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MongoDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: mg-as-sh
  namespace: demo
spec:
  databaseRef:
    name: mg-sh
  storage:
    shard:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mg-sh` database.
- `spec.storage.shard.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.shard.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.shard.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- It has another field `spec.storage.replicaSet.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

> Note: In this demo we are only setting up the storage autoscaling for the shard pods, that's why we only specified the shard section of the autoscaler. You can enable autoscaling for configServer pods in the same yaml, by specifying the `spec.configServer` section, similar to the `spec.shard` section we have configured in this demo.


Let's create the `MongoDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/storage/mg-as-sh.yaml
mongodbautoscaler.autoscaling.kubedb.com/mg-as-sh created
```

#### Storage Autoscaling is set up successfully

Let's check that the `mongodbautoscaler` resource is created successfully,

```bash
$ kubectl get mongodbautoscaler -n demo
NAME       AGE
mg-as-sh   20s

$ kubectl describe mongodbautoscaler mg-as-sh -n demo
Name:         mg-as-sh
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MongoDBAutoscaler
Metadata:
  Creation Timestamp:  2021-03-08T14:26:06Z
  Generation:          1
  Managed Fields:
    API Version:  autoscaling.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:storage:
          .:
          f:shard:
            .:
            f:scalingThreshold:
            f:trigger:
            f:usageThreshold:
    Manager:         kubectl-client-side-apply
    Operation:       Update
    Time:            2021-03-08T14:26:06Z
  Resource Version:  156292
  Self Link:         /apis/autoscaling.kubedb.com/v1alpha1/namespaces/demo/mongodbautoscalers/mg-as-sh
  UID:               203e332f-bdfe-470f-a429-a7b60c7be2ee
Spec:
  Database Ref:
    Name:  mg-sh
  Storage:
    Shard:
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```
So, the `mongodbautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up one of the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume using the following commands:

```bash
$ kubectl exec -it -n demo mg-sh-shard0-0 -- bash
root@mg-sh-shard0-0:/# df -h /data/db
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/ad11042f-f4cc-4dfc-9680-2afbbb199d48 1014M  335M  680M  34% /data/db
root@mg-sh-shard0-0:/# dd if=/dev/zero of=/data/db/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 0.595358 s, 881 MB/s
root@mg-sh-shard0-0:/# df -h /data/db
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/ad11042f-f4cc-4dfc-9680-2afbbb199d48 1014M  837M  178M  83% /data/db
```

So, from the above output we can see that the storage usage is 83%, which exceeded the `usageThreshold` 60%.

Let's watch the `mongodbopsrequest` in the demo namespace to see if any `mongodbopsrequest` object is created. After some time you'll see that a `mongodbopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS        AGE
mops-mg-sh-ba5ikn   VolumeExpansion   Progressing   41s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS        AGE
mops-mg-sh-ba5ikn   VolumeExpansion   Successful    2m54s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-mg-sh-ba5ikn
Name:         mops-mg-sh-ba5ikn
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mg-sh
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mongodbs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-08T14:31:52Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:labels:
          .:
          f:app.kubernetes.io/component:
          f:app.kubernetes.io/instance:
          f:app.kubernetes.io/managed-by:
          f:app.kubernetes.io/name:
        f:ownerReferences:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:type:
        f:volumeExpansion:
          .:
          f:shard:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2021-03-08T14:31:52Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:    kubedb-enterprise
    Operation:  Update
    Time:       2021-03-08T14:31:52Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MongoDBAutoscaler
    Name:                  mg-as-sh
    UID:                   203e332f-bdfe-470f-a429-a7b60c7be2ee
  Resource Version:        158488
  Self Link:               /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-mg-sh-ba5ikn
  UID:                     c56236c2-5b64-4775-ba5a-35727b96a414
Spec:
  Database Ref:
    Name:  mg-sh
  Type:    VolumeExpansion
  Volume Expansion:
    Shard:  1594884096
Status:
  Conditions:
    Last Transition Time:  2021-03-08T14:31:52Z
    Message:               MongoDB ops request is expanding volume of database
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2021-03-08T14:34:32Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                ShardVolumeExpansion
    Status:                True
    Type:                  ShardVolumeExpansion
    Last Transition Time:  2021-03-08T14:34:37Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                
    Status:                True
    Type:                  
    Last Transition Time:  2021-03-08T14:34:42Z
    Message:               StatefulSet is recreated
    Observed Generation:   1
    Reason:                ReadyStatefulSets
    Status:                True
    Type:                  ReadyStatefulSets
    Last Transition Time:  2021-03-08T14:34:42Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                Age    From                        Message
  ----    ------                ----   ----                        -------
  Normal  PauseDatabase         3m21s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-sh
  Normal  PauseDatabase         3m21s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-sh
  Normal  ShardVolumeExpansion  41s    KubeDB Ops-manager operator  Successfully Expanded Volume
  Normal                        36s    KubeDB Ops-manager operator  Successfully Expanded Volume
  Normal  ResumeDatabase        36s    KubeDB Ops-manager operator  Resuming MongoDB demo/mg-sh
  Normal  ResumeDatabase        36s    KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-sh
  Normal  ReadyStatefulSets     31s    KubeDB Ops-manager operator  StatefulSet is recreated
  Normal  Successful            31s    KubeDB Ops-manager operator  Successfully Expanded Volume
```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volume` whether the volume of the shard nodes of the database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo mg-sh-shard0 -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                            STORAGECLASS          REASON   AGE
pvc-031836c6-95ae-4015-938c-da183c205828   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-configsvr-0   topolvm-provisioner            13m
pvc-2515233f-0f7d-4d0d-8b45-97a3cb9d4488   2Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard0-2      topolvm-provisioner            11m
pvc-35f73708-3c11-4ead-a60b-e1679a294b81   2Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard0-0      topolvm-provisioner            13m
pvc-4b329feb-8c92-4605-a37e-c02b3499e311   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-configsvr-2   topolvm-provisioner            11m
pvc-52490270-1355-4045-b2a1-872a671ab006   1Gi        RWO            Delete           Bound    demo/datadir-mg-sh-configsvr-1   topolvm-provisioner            12m
pvc-80dc91d3-f56f-4037-b6e1-f69e13fb434c   2Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard1-1      topolvm-provisioner            12m
pvc-c1965a32-7471-4885-ac52-f9eab056d48e   2Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard1-2      topolvm-provisioner            11m
pvc-c838a27d-c75d-4caa-9c1d-456af3bfaba0   2Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard1-0      topolvm-provisioner            12m
pvc-d47f19be-f206-41c5-a0b1-5022776fea2f   2Gi        RWO            Delete           Bound    demo/datadir-mg-sh-shard0-1      topolvm-provisioner            12m
```

The above output verifies that we have successfully autoscaled the volume of the shard nodes of this MongoDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-sh
kubectl delete mongodbautoscaler -n demo mg-as-sh
```
