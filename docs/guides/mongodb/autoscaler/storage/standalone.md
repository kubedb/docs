---
title: MongoDB Standalone Autoscaling
menu:
  docs_{{ .version }}:
    identifier: mg-storage-auto-scaling-standalone
    name: Standalone
    parent: mg-storage-auto-scaling
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a MongoDB Standalone Database

This guide will show you how to use `KubeDB` to autoscale the storage of a MongoDB standalone database.

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

## Storage Autoscaling of Standalone Database

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  9h
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   9h
```

We can see from the output the `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it. You can install topolvm from [here](https://github.com/topolvm/topolvm)

Now, we are going to deploy a `MongoDB` standalone using a supported version by `KubeDB` operator. Then we are going to apply `MongoDBAutoscaler` to set up autoscaling.

#### Deploy MongoDB standalone

In this section, we are going to deploy a MongoDB standalone database with version `4.4.26`.  Then, in the next section we will set up autoscaling for this database using `MongoDBAutoscaler` CRD. Below is the YAML of the `MongoDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-standalone
  namespace: demo
spec:
  version: "4.4.26"
  storageType: Durable
  storage:
    storageClassName: topolvm-provisioner
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MongoDB` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/storage/mg-standalone.yaml
mongodb.kubedb.com/mg-standalone created
```

Now, wait until `mg-standalone` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo
NAME            VERSION    STATUS    AGE
mg-standalone   4.4.26      Ready     2m53s
```

Let's check volume size from petset, and from the persistent volume,

```bash
$ kubectl get sts -n demo mg-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS          REASON   AGE
pvc-cf469ed8-a89a-49ca-bf7c-8c76b7889428   1Gi        RWO            Delete           Bound    demo/datadir-mg-standalone-0   topolvm-provisioner            7m41s
```

You can see the petset has 1GB storage, and the capacity of the persistent volume is also 1GB.

We are now ready to apply the `MongoDBAutoscaler` CRO to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a MongoDBAutoscaler Object.

#### Create MongoDBAutoscaler Object

In order to set up vertical autoscaling for this standalone database, we have to create a `MongoDBAutoscaler` CRO with our desired configuration. Below is the YAML of the `MongoDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: MongoDBAutoscaler
metadata:
  name: mg-as
  namespace: demo
spec:
  databaseRef:
    name: mg-standalone
  storage:
    standalone:
      expansionMode: "Online"
      trigger: "On"
      usageThreshold: 60
      scalingThreshold: 50
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `mg-standalone` database.
- `spec.storage.standalone.trigger` specifies that storage autoscaling is enabled for this database.
- `spec.storage.standalone.usageThreshold` specifies storage usage threshold, if storage usage exceeds `60%` then storage autoscaling will be triggered.
- `spec.storage.standalone.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- It has another field `spec.storage.replicaSet.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `MongoDBAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/autoscaling/storage/mg-as-standalone.yaml
mongodbautoscaler.autoscaling.kubedb.com/mg-as created
```

#### Storage Autoscaling is set up successfully

Let's check that the `mongodbautoscaler` resource is created successfully,

```bash
$ kubectl get mongodbautoscaler -n demo
NAME    AGE
mg-as   102s

$ kubectl describe mongodbautoscaler mg-as -n demo
Name:         mg-as
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         MongoDBAutoscaler
Metadata:
  Creation Timestamp:  2021-03-08T12:58:01Z
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
          f:standalone:
            .:
            f:scalingThreshold:
            f:trigger:
            f:usageThreshold:
    Manager:         kubectl-client-side-apply
    Operation:       Update
    Time:            2021-03-08T12:58:01Z
  Resource Version:  134423
  Self Link:         /apis/autoscaling.kubedb.com/v1alpha1/namespaces/demo/mongodbautoscalers/mg-as
  UID:               999a2dc9-7eb7-4ed2-9e90-d3f8b21c091a
Spec:
  Database Ref:
    Name:  mg-standalone
  Storage:
    Standalone:
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    60
Events:                   <none>
```
So, the `mongodbautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the database pod and fill the database volume using the following commands:

```bash
$ kubectl exec -it -n demo mg-standalone-0 -- bash
root@mg-standalone-0:/# df -h /data/db
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/1df4ee9e-b900-4c0f-9d2c-8493fb30bdc0 1014M  334M  681M  33% /data/db
root@mg-standalone-0:/# dd if=/dev/zero of=/data/db/file.img bs=500M count=1
1+0 records in
1+0 records out
524288000 bytes (524 MB, 500 MiB) copied, 0.359202 s, 1.5 GB/s
root@mg-standalone-0:/# df -h /data/db
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/1df4ee9e-b900-4c0f-9d2c-8493fb30bdc0 1014M  835M  180M  83% /data/db
```

So, from the above output we can see that the storage usage is 84%, which exceeded the `usageThreshold` 60%.

Let's watch the `mongodbopsrequest` in the demo namespace to see if any `mongodbopsrequest` object is created. After some time you'll see that a `mongodbopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                        TYPE              STATUS        AGE
mops-mg-standalone-p27c11   VolumeExpansion   Progressing   26s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get mongodbopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                        TYPE              STATUS        AGE
mops-mg-standalone-p27c11   VolumeExpansion   Successful    73s
```

We can see from the above output that the `MongoDBOpsRequest` has succeeded. If we describe the `MongoDBOpsRequest` we will get an overview of the steps that were followed to expand the volume of the database.

```bash
$ kubectl describe mongodbopsrequest -n demo mops-mg-standalone-p27c11
Name:         mops-mg-standalone-p27c11
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=mg-standalone
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=mongodbs.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MongoDBOpsRequest
Metadata:
  Creation Timestamp:  2021-03-08T13:19:51Z
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
          f:standalone:
    Manager:      kubedb-autoscaler
    Operation:    Update
    Time:         2021-03-08T13:19:51Z
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
    Time:       2021-03-08T13:19:52Z
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  MongoDBAutoscaler
    Name:                  mg-as
    UID:                   999a2dc9-7eb7-4ed2-9e90-d3f8b21c091a
  Resource Version:        139871
  Self Link:               /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mongodbopsrequests/mops-mg-standalone-p27c11
  UID:                     9606485d-9dd8-4787-9c7c-61fc874c555e
Spec:
  Database Ref:
    Name:  mg-standalone
  Type:    VolumeExpansion
  Volume Expansion:
    Standalone:  1594884096
Status:
  Conditions:
    Last Transition Time:  2021-03-08T13:19:52Z
    Message:               MongoDB ops request is expanding volume of database
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2021-03-08T13:20:47Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                StandaloneVolumeExpansion
    Status:                True
    Type:                  StandaloneVolumeExpansion
    Last Transition Time:  2021-03-08T13:20:52Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                
    Status:                True
    Type:                  
    Last Transition Time:  2021-03-08T13:20:57Z
    Message:               PetSet is recreated
    Observed Generation:   1
    Reason:                ReadyPetSets
    Status:                True
    Type:                  ReadyPetSets
    Last Transition Time:  2021-03-08T13:20:57Z
    Message:               Successfully Expanded Volume
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                     Age   From                        Message
  ----    ------                     ----  ----                        -------
  Normal  PauseDatabase              110s  KubeDB Ops-manager operator  Pausing MongoDB demo/mg-standalone
  Normal  PauseDatabase              110s  KubeDB Ops-manager operator  Successfully paused MongoDB demo/mg-standalone
  Normal  StandaloneVolumeExpansion  55s   KubeDB Ops-manager operator  Successfully Expanded Volume
  Normal                             50s   KubeDB Ops-manager operator  Successfully Expanded Volume
  Normal  ResumeDatabase             50s   KubeDB Ops-manager operator  Resuming MongoDB demo/mg-standalone
  Normal  ResumeDatabase             50s   KubeDB Ops-manager operator  Successfully resumed MongoDB demo/mg-standalone
  Normal  ReadyPetSets          45s   KubeDB Ops-manager operator  PetSet is recreated
  Normal  Successful                 45s   KubeDB Ops-manager operator  Successfully Expanded Volume
```

Now, we are going to verify from the `Petset`, and the `Persistent Volume` whether the volume of the standalone database has expanded to meet the desired state, Let's check,

```bash
$ kubectl get sts -n demo mg-standalone -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1594884096"
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                          STORAGECLASS          REASON   AGE
pvc-cf469ed8-a89a-49ca-bf7c-8c76b7889428   2Gi        RWO            Delete           Bound    demo/datadir-mg-standalone-0   topolvm-provisioner            26m
```

The above output verifies that we have successfully autoscaled the volume of the MongoDB standalone database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-standalone
kubectl delete mongodbautoscaler -n demo mg-as
```
