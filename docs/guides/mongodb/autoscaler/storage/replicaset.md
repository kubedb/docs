---
title: MongoDB Replicaset Autoscaling
menu:
  docs_{{ .version }}:
    identifier: mg-storage-auto-scaling-replicaset
    name: ReplicaSet
    parent: mg-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a MongoDB Replicaset Database

This guide will show you how to use `KubeDB` to autoscale the storage of a MongoDB Replicaset database.

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

## Storage Autoscaling of ReplicaSet Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  9h
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   9h
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `MongoDB` replicaset using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBAutoscaler` to set up autoscaling.

#### Deploy MongoDB replicaset

In this section, we are going to deploy a MongoDB replicaset database with version `4.4.26`.  Then, in the next section we will set up autoscaling for this database using `MongoDBAutoscaler` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "replicaset"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: topolvm-provisioner
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `MongoDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/storage/mg-rs.yaml
mongodb.kubedb.com/mg-rs created
```

Now, wait until `mg-rs` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME      VERSION    STATUS    AGE
mg-rs     4.4.26      Ready     2m53s
```

Let's check volume size from statefulset, and from the persistent volume,

```bash
$ kubectl get sts -n demo mg-rs -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS          REASON   AGE
pvc-b16daa50-83fc-4d25-b553-4a25f13166d5   1Gi        RWO            Delete           Bound    demo/datadir-mg-rs-0   topolvm-provisioner            2m12s
pvc-d4616bef-359d-4b73-ab9f-38c24aaaec8c   1Gi        RWO            Delete           Bound    demo/datadir-mg-rs-1   topolvm-provisioner            61s
pvc-ead21204-3dc7-453c-8121-d2fe48b1c3e2   1Gi        RWO            Delete           Bound    demo/datadir-mg-rs-2   topolvm-provisioner            18s
```

You can see the statefulset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `MongoDBAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a MongoDBAutoscaler Object.

#### Create MongoDBAutoscaler Object

In order to set up vertical autoscaling for this replicaset database, we have to create a `MongoDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MongoDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: mg-as-rs
  namespace: demo
spec:
  databaseRef:
    name: mg-rs
  storage:
    replicaSet:
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mg-rs` database.
- `spec.storage.replicaSet.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.replicaSet.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.replicaSet.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- It has another field `spec.storage.replicaSet.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `MongoDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/storage/mg-as-rs.yaml
mongodbautoscaler.autoscaling.kubedb.com/mg-as-rs created
```

#### Storage Autoscaling is set up successfully

Let's check that the `mongodbautoscaler` resource is created successfully,

```bash
$ kubectl get mongodbautoscaler -n demo
NAME       AGE
mg-as-rs   20s

$ kubectl describe mongodbautoscaler mg-as-rs -n demo
Name:         mg-as-rs
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MongoDBAutoscaler
Metadata:
  Creation Timestamp:  2021-03-08T14:11:46Z
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
          f:replicaSet:
            .:
            f:scalingThreshold:
            f:trigger:
            f:usageThreshold:
    Manager:         kubectl-client-side-apply
    Operation:       Update
    Time:            2021-03-08T14:11:46Z
  Resource Version:  152149
  Self Link:         /apis/autoscaling.kubedb.com/v1alpha1/namespaces/demo/mongodbautoscalers/mg-as-rs
  UID:               a0dab64d-e7c4-4819-8ffe-360c70231577
Spec:
  Database Ref:
    Name:  mg-rs
  Storage:
    Replica Set:
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```
So, the `mongodbautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume using the following commands:

```bash
$ kubectl exec -it -n demo mg-rs-0 -- bash
root@mg-rs-0:/# df -h /data/db
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/760cb655-91fe-4497-ab4a-a771aa53ece4 1014M  335M  680M  33% /data/db
root@mg-rs-0:/# dd if=/dev/zero of=/data/db/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 0.482378 s, 1.1 GB/s
root@mg-rs-0:/# df -h /data/db
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/760cb655-91fe-4497-ab4a-a771aa53ece4 1014M  835M  180M  83% /data/db
```

So, from the above output we can see that the storage usage is 83%, which exceeded the `usageThreshold` 60%.

Let's watch the `mongodbopsrequest` in the demo namespace to see if any `mongodbopsrequest` object is created. After some time you'll see that a `mongodbopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS        AGE
mops-mg-rs-mft11m   VolumeExpansion   Progressing   10s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                TYPE              STATUS        AGE
mops-mg-rs-mft11m   VolumeExpansion   Successful    97s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-mg-rs-mft11m
Name:         mops-mg-rs-mft11m
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mg-rs
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mongodbs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-08T14:15:52Z
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
          f:replicaSet:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2021-03-08T14:15:52Z
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
    Time:       2021-03-08T14:15:52Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MongoDBAutoscaler
    Name:                  mg-as-rs
    UID:                   a0dab64d-e7c4-4819-8ffe-360c70231577
  Resource Version:        153496
  Self Link:               /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-mg-rs-mft11m
  UID:                     84567b84-6de4-4658-b0d2-2c374e03e63d
Spec:
  Database Ref:
    Name:  mg-rs
  Type:    VolumeExpansion
  Volume Expansion:
    Replica Set:  1594884096
Status:
  Conditions:
    Last Transition Time:  2021-03-08T14:15:52Z
    Message:               MongoDB ops request is expanding volume of database
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2021-03-08T14:17:02Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                ReplicasetVolumeExpansion
    Status:                True
    Type:                  ReplicasetVolumeExpansion
    Last Transition Time:  2021-03-08T14:17:07Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                
    Status:                True
    Type:                  
    Last Transition Time:  2021-03-08T14:17:12Z
    Message:               StatefulSet is recreated
    Observed Generation:   1
    Reason:                ReadyStatefulSets
    Status:                True
    Type:                  ReadyStatefulSets
    Last Transition Time:  2021-03-08T14:17:12Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                     Age    From                        Message
  ----    ------                     ----   ----                        -------
  Normal  PauseDatabase              2m36s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-rs
  Normal  PauseDatabase              2m36s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-rs
  Normal  ReplicasetVolumeExpansion  86s    KubeDB Ops-manager operator  Successfully Expanded Volume
  Normal                             81s    KubeDB Ops-manager operator  Successfully Expanded Volume
  Normal  ResumeDatabase             81s    KubeDB Ops-manager operator  Resuming MongoDB demo/mg-rs
  Normal  ResumeDatabase             81s    KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-rs
  Normal  ReadyStatefulSets          76s    KubeDB Ops-manager operator  StatefulSet is recreated
  Normal  Successful                 76s    KubeDB Ops-manager operator  Successfully Expanded Volume
```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volume` whether the volume of the replicaset database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo mg-rs -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS          REASON   AGE
pvc-b16daa50-83fc-4d25-b553-4a25f13166d5   2Gi        RWO            Delete           Bound    demo/datadir-mg-rs-0   topolvm-provisioner            11m
pvc-d4616bef-359d-4b73-ab9f-38c24aaaec8c   2Gi        RWO            Delete           Bound    demo/datadir-mg-rs-1   topolvm-provisioner            10m
pvc-ead21204-3dc7-453c-8121-d2fe48b1c3e2   2Gi        RWO            Delete           Bound    demo/datadir-mg-rs-2   topolvm-provisioner            9m52s
```

The above output verifies that we have successfully autoscaled the volume of the MongoDB replicaset database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-rs
kubectl delete mongodbautoscaler -n demo mg-as-rs
```
